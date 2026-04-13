# Meeting Room Booking Service
Сервис бронирования переговорок с JWT авторизацией и ролевой моделью (admin/user).

# Структура репозитория
```
meeting-room-booking/
├── api.yaml                # OpenAPI спецификация
├── docker-compose.yml      # Docker Compose конфиг
├── Dockerfile              # Docker образ
├── go.mod                  # Go модуль
├── README.md               # Этот файл
├── cmd/server/             # Точка входа
├── internal/               # Внутренний код
│   ├── config/             # Конфигурация
│   ├── domain/             # Бизнес-сущности
│   ├── handler/http/       # HTTP хендлеры
│   ├── logger/             # Логирование
│   ├── migrator/           # Миграции
│   ├── repository/         # Репозитории
│   └── service/            # Сервисы
├── migrations/             # SQL миграции
└── tests/integration/      # E2E тесты
```
# Клонировать репозиторий
git clone https://github.com/THE-MDA/meeting-room-booking.git
cd meeting-room-booking
# Запустить сервер
go run cmd/server/main.go

# API Endpoints
Метод	Эндпоинт	Описание	Доступ
POST	/dummyLogin	Получение JWT токена	Public
GET	/rooms/list	Список переговорок	Admin, User
POST	/rooms/create	Создание переговорки	Admin
POST	/rooms/{roomId}/schedule/create	Создание расписания	Admin
GET	/rooms/{roomId}/slots/list	Доступные слоты	Admin, User
POST	/bookings/create	Создание брони	User
GET	/bookings/list	Все брони (пагинация)	Admin
GET	/bookings/my	Мои брони	User
POST	/bookings/{bookingId}/cancel	Отмена брони	User
GET	/_info	Health check	Public
Полная спецификация API в файле api.yaml.

# Проект построен на принципах Clean Architecture:
```
cmd/server/          # Точка входа
internal/
├── domain/          # Бизнес-сущности и правила
├── repository/      # Работа с БД
├── service/         # Бизнес-логика
├── handler/http/    # HTTP обработчики
├── config/          # Конфигурация
├── logger/          # Логирование
└── migrator/        # Миграции
migrations/          # SQL миграции
tests/integration/   # E2E тесты
```

# PostgreSQL схема:
users - пользователи (id, email, role, created_at)
rooms - переговорки (id, name, description, capacity)
schedules — расписания (id, room_id, day_of_week, start_time, end_time)
slots — слоты (id, room_id, start_time, end_time)
bookings — брони (id, slot_id, user_id, status)

# Использован подход: генерация и хранение слотов в БД при создании расписания.
Почему выбран этот подход:
Стабильные UUID слотов — требуется для бронирования по slotId
Производительность — запрос доступных слотов выполняется за <200 мс (SELECT с JOIN)
Простота реализации — не нужно генерировать слоты на каждый запрос
Соответствие ТЗ — слоты должны иметь стабильные UUID, сохранённые в БД

# JWT токен содержит:
user_id — UUID пользователя
role — admin / user
exp — время истечения (30 минут)

Получение токена:
curl -X POST http://localhost:8080/dummyLogin \
  -H "Content-Type: application/json" \
  -d '{"role":"admin"}'
  
Использование:
curl -X GET http://localhost:8080/rooms/list \
  -H "Authorization: Bearer <token>"

# Тестирование
Все тесты
go test ./... -v

Юнит-тесты
go test ./internal/... -v

Интеграционные тесты (E2E)
go test ./tests/integration/... -v

E2E тест 1: создание переговорки -> создание расписания -> создание брони пользователем
E2E тест 2: создание брони -> отмена брони

# Решения по вопросам, не описанным в ТЗ
Генерация слотов -> При создании расписания на 90 дней вперёд -> Пользователи запрашивают слоты в пределах 7 дней; запас прочности
Формат времени в расписании -> HH:MM:SS -> Удобно для хранения в БД типа TIME
Пагинация -> page и pageSize (max 100) -> Согласно API спецификации
Отмена брони -> POST /bookings/{id}/cancel -> Идемпотентная операция, возвращает 200

# Технологии
Go 1.21 — язык сервиса
PostgreSQL 15 — база данных
Docker Compose — оркестрация
JWT — авторизация
slog — структурированное логирование
golang-migrate — миграции

# Автор
Мильчаков Дмитрий Александрович

Тестовое задание для стажёра Backend
