# Go Parser System

Система парсинга данных с очередями сообщений и хранением в MongoDB.

## 🚀 Features

- ✅ Парсинг данных из различных источников
- ✅ Очередь сообщений RabbitMQ для обработки
- ✅ Хранение данных в MongoDB
- ✅ Параллельная обработка с Worker Pool
- ✅ Конфигурация через .env файл
- ✅ Логирование операций
- ✅ Грациозный shutdown

## 🛠️ Tech Stack

- **Go 1.21+** - язык программирования
- **RabbitMQ** - очередь сообщений
- **MongoDB** - документоориентированная БД
- **Fiber** - веб-фреймворк (Fast HTTP)
- **Docker Compose** - контейнеризация
- **GORM** - ORM для работы с БД

## 🚀 Quick Start

### 1. Клонирование репозитория
```bash
git clone <repository-url>
cd go-parser-system
```

### 2. Копирование конфигурации
```bash
cp .env.example .env
```

### 3. Запуск через Docker Compose
```bash
docker compose up --build
```

### 4. Альтернативный запуск (если сервисы уже запущены)
```bash
go run main.go
```

## 📁 Project Structure

```
go-parser-system/
├── main.go                 # Точка входа приложения
├── go.mod                 # Go модули
├── .env.example           # Пример конфигурации
├── docker-compose.yml     # Docker Compose конфигурация
├── internal/              # Внутренние пакеты
│   ├── config/           # Конфигурация приложения
│   ├── database/         # Работа с БД
│   ├── domain/           # Доменные модели
│   ├── handler/          # HTTP обработчики
│   ├── parser/           # Парсеры данных
│   │   └── plans/       # Парсеры планов
│   ├── queue/            # Работа с RabbitMQ
│   ├── utils/            # Утилиты
│   └── worker/           # Worker Pool
└── tests/                # Тесты
```

## 🔧 Configuration (.env)

```bash
# MongoDB
MONGO_URI=mongodb://localhost:27017
DB_NAME=parser_db
COLLECTION_NAME=records

# RabbitMQ
RABBITMQ_URI=amqp://guest:guest@localhost:5672/
QUEUE_NAME=parsing_queue

# Приложение
APP_PORT=8080
LOG_LEVEL=info
MAX_WORKERS=3
```

## 📖 Data Models

### Record Model

```go
type Record struct {
	ID        primitive.ObjectID `bson:"_id, omitempty" json:"id"`
	URL       string             `bson:"url" json:"url"`
	Data      interface{}        `bson:"data" json:"data"`
	Status    string             `bson:"status" json:"status"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}
```

## 🚀 API Endpoints

### POST /parse - Запустить парсинг

Запускает парсинг данных по указанному URL.

**Request Body:**
```json
{
  "url": "https://example.com/data",
  "parser_type": "registr"
}
```

**Response:**
- `202 Accepted` - Задача принята в обработку
- `400 Bad Request` - Невалидный запрос
- `500 Internal Server Error` - Ошибка сервера

### GET /records - Получить записи

Возвращает список обработанных записей.

**Query Parameters:**
- `status` - Фильтр по статусу
- `page` - Номер страницы (default: 1)
- `limit` - Количество элементов (default: 20)

**Response:**
- `200 OK` - Список записей

### GET /records/:id - Получить запись

Возвращает информацию о конкретной записи.

**Response:**
- `200 OK` - Запись найдена
- `404 Not Found` - Запись не найдена

## 🐳 Docker Compose

### Структура
```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - mongo
      - rabbitmq
    environment:
      - MONGO_URI=mongodb://mongo:27017
      - RABBITMQ_URI=amqp://guest:guest@rabbitmq:5672/

  mongo:
    image: mongo:6
    ports:
      - "27017:27017"
    volumes:
      - mongo_data:/data/db

  rabbitmq:
    image: rabbitmq:3-management
    ports:
      - "5672:5672"
      - "15672:15672"
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq

volumes:
  mongo_data:
  rabbitmq_data:
```

## 🛠️ Development

### Установка зависимостей
```bash
go mod download
```

### Запуск тестов
```bash
go test ./...
```

### Сборка проекта
```bash
go build -o parser-system .
```

### Запуск в development режиме
```bash
go run main.go
```

## 📊 Monitoring

### RabbitMQ Management UI
```
http://localhost:15672
```

**Credentials:**
- Username: `guest`
- Password: `guest`

### MongoDB
```
mongo localhost:27017/parser_db
```

## 🚨 Error Handling

Система включает:
- ✅ Retry механизмы для обработки ошибок
- ✅ Грациозный shutdown
- ✅ Логирование всех операций
- ✅ Валидация входящих данных
- ✅ Обработка таймаутов

## 💡 Examples

### Запуск парсинга
```bash
curl -X POST http://localhost:8080/parse \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com/data",
    "parser_type": "registr"
  }'
```

### Получение записей
```bash
curl "http://localhost:8080/records?status=processed"
```

## 📄 License

MIT License - см. файл LICENSE для деталей.

## 🤝 Contributing

1. Fork репозитория
2. Создать новую ветку (`git checkout -b feature/amazing-feature`)
3. Commit изменения (`git commit -m 'Add amazing feature'`)
4. Push ветку (`git push origin feature/amazing-feature`)
5. Create Pull Request

---

**Made with ❤️ by Go Parser Team**