# Auth Service

gRPC-сервис аутентификации для CRM-платформы. Обрабатывает регистрацию, вход, обновление токенов и выход. Данные хранятся в PostgreSQL. Сервис используется API Gateway (`github.com/JanMird/gateway`) как бэкенд для эндпоинтов `/auth/*`.

## О проекте

Сервис реализует gRPC-контракт `AuthService` (пакет `github.com/JanMird/auth/gen/auth`) и работает поверх PostgreSQL. На каждый входящий вызов выполняется бизнес-логика в слое `service`, которая обращается к БД через слой `storage`.

**Ключевые задачи:**

- Регистрация пользователя (`SignUp`)
- Аутентификация и выдача токенов (`Authenticate`)
- Обновление access-токена (`RefreshToken`)
- Выход / инвалидация токена (`SignOut`)
- Проверка живости (`HealthCheck`)

## Технологический стек

| Компонент | Технология |
|---|---|
| Язык | Go |
| Транспорт | gRPC (`google.golang.org/grpc`) |
| БД | PostgreSQL (`github.com/lib/pq`) |
| Миграции | [goose](https://github.com/pressly/goose) |
| Конфигурация | `.env` через `godotenv` |
| Логирование | `sirupsen/logrus` |
| Рефлексия gRPC | `google.golang.org/grpc/reflection` |

## Быстрый старт

### Требования

- Go 1.21+
- PostgreSQL 14+
- [goose](https://github.com/pressly/goose) (для миграций)

### Клонирование и сборка

```bash
git clone https://github.com/JanMird/auth.git
cd auth

go mod download
go build -o auth .
```

### Переменные окружения

Создайте `.env` в корне проекта:

```env
# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=auth
```

Строка подключения формируется автоматически в формате:

```
host=<DB_HOST> port=<DB_PORT> user=<DB_USER> password=<DB_PASSWORD> dbname=<DB_NAME> sslmode=disable
```

### Миграции

Схема БД создаётся через `goose`. Миграции лежат в каталоге проекта.

```bash
# Установка goose (если ещё нет)
go install github.com/pressly/goose/v3/cmd/goose@latest

# Применение всех миграций
goose -dir ./migrations postgres "host=localhost port=5432 user=postgres password=postgres dbname=auth sslmode=disable" up

# Откат последней миграции
goose -dir ./migrations postgres "host=localhost port=5432 user=postgres password=postgres dbname=auth sslmode=disable" down
```

Создаваемые таблицы:

| Таблица | Назначение |
|---|---|
| `users` | Пользователи: `id`, `email`, `password_hash`, `phone`, `role`, `created_at`, `updated_at` |
| `ref_token` | Refresh-токены (`ref_token`, уникальный) |

### Запуск

```bash
go run .
# или
./auth
```

Сервер слушает `:50052`.

При успешном старте выводится `ready to serve`.

## gRPC API

Сервис регистрирует `AuthServiceServer` из пакета `gen/auth`. Точные сигнатуры методов смотрите в `.proto`-файле и сгенерированном коде.

| Метод | Назначение |
|---|---|
| `SignUp` | Регистрация нового пользователя |
| `Authenticate` | Проверка учётных данных, выдача токенов |
| `RefreshToken` | Обновление access-токена по refresh-токену |
| `SignOut` | Завершение сессии / инвалидация токена |
| `HealthCheck` | Проверка доступности сервиса |

На сервере включена **gRPC reflection**, поэтому сервис можно исследовать через `grpcurl`:

```bash
grpcurl -plaintext localhost:50052 list
grpcurl -plaintext localhost:50052 list auth.AuthService
grpcurl -plaintext localhost:50052 auth.AuthService/HealthCheck
```

## Архитектура

```
gRPC request → handler → service → storage → PostgreSQL
```

Основные слои:

- `internal/handler` — реализация gRPC-методов `AuthServiceServer`; принимает `authService` и `logger`
- `internal/service` — бизнес-логика аутентификации, работа с токенами
- `internal/storage` — слой доступа к данным поверх `*sql.DB` (PostgreSQL)
- `gen/auth` — сгенерированный из `.proto` код (сервер и клиент)

## Структура проекта

```
auth/
├── cmd/
|   └── main.go             # точка входа: БД, gRPC-сервер, reflection
├── proto/                  # proto-файлы
├── internal/
│   ├── config/             # конфиги
│   ├── domain/             # доменные типы
│   ├── handler/            # gRPC-хендлеры
│   ├── service/            # бизнес-логика
│   └── storage/            # работа с PostgreSQL
├── migrations/             # goose-миграции
├── pkg/error               № ошибки
├── .env
└── go.mod
```

## Тестирование

```bash
go test ./...

# Ручная проверка через grpcurl
grpcurl -plaintext -d '{}' localhost:50052 auth.AuthService/HealthCheck

grpcurl -plaintext -d '{"email":"user@example.com","password":"secret"}' \
  localhost:50052 auth.AuthService/SignUp
```
