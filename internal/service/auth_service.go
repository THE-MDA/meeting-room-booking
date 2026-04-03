package service

import (
    "context"
    "errors"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
    "meeting-room-booking/internal/domain"
)

type AuthService struct {
    userRepo      domain.UserRepository
    jwtSecret     []byte
    tokenDuration time.Duration
}

type Claims struct {
    UserID uuid.UUID `json:"user_id"`
    Role   string    `json:"role"`
    jwt.RegisteredClaims
}

func NewAuthService(userRepo domain.UserRepository, jwtSecret string, tokenDuration time.Duration) *AuthService {
    return &AuthService{
        userRepo:      userRepo,
        jwtSecret:     []byte(jwtSecret),
        tokenDuration: tokenDuration,
    }
}

func (s *AuthService) GenerateToken(userID uuid.UUID, role string) (string, error) {
    claims := &Claims{
        UserID: userID,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenDuration)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(s.jwtSecret)
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return s.jwtSecret, nil
    })

    if err != nil {
        return nil, errors.New("invalid token")
    }

    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, errors.New("invalid token claims")
    }

    return claims, nil
}

func (s *AuthService) DummyLogin(ctx context.Context, role string) (string, error) {
    var userID uuid.UUID
    var email string
    
    switch role {
    case "admin":
        userID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
        email = "admin@example.com"
    case "user":
        userID = uuid.MustParse("00000000-0000-0000-0000-000000000002")
        email = "user@example.com"
    default:
        return "", errors.New("invalid role")
    }
    
    user, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        return "", err
    }
    
    if user == nil {
        user, err = domain.NewUser(email, domain.UserRole(role))
        if err != nil {
            return "", err
        }
        if err := s.userRepo.Create(ctx, user); err != nil {
            return "", err
        }
    }
    
    return s.GenerateToken(user.ID, string(user.Role))
}

func (s *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
    return s.userRepo.GetByID(ctx, userID)
}