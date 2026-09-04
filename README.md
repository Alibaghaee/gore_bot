# 🤖 Gore Bot - Telegram Connector Service

**A Go-based service for connecting to the Telegram Bot API and routing messages**

An advanced project that receives Telegram requests in real-time and forwards them through RabbitMQ to a Laravel backend. This project is implemented with clean architecture and readable, maintainable code.

## ✨ Key Features

- 🔌 **Webhook Support**: Receive Telegram messages via webhook
- 🐰 **RabbitMQ Integration**: Smart message queuing for worker processors
- 🗄️ **PostgreSQL ORM**: Store bot data using GORM
- 🔐 **Authentication**: Protect endpoints with authentication middleware
- 🚀 **Graceful Shutdown**: Safe and proper server shutdown
- 🐳 **Docker Support**: Ready to run in containerized environment
- 📝 **Clean Code**: Layered architecture and readable code

---

## 🏗️ Project Architecture

```
gore_bot/
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
├── internal/
│   ├── api/
│   │   ├── handler.go              # HTTP handlers for endpoints
│   │   └── middleware.go           # Authentication middleware
│   ├── domain/
│   │   └── bot.go                  # Bot data models
│   ├── queue/
│   │   └── rabbitmq.go             # RabbitMQ configuration
│   └── repository/
│       └── bot_repo.go             # Database queries
├── pkg/
│   └── database/
│       └── postgres.go             # PostgreSQL connection
├── deployments/
│   ├── Dockerfile.dev              # Dockerfile for development
│   └── Dockerfile.prod             # Dockerfile for production
├── docker-compose.yml              # Service composition
├── .env                            # Environment variables
├── go.mod & go.sum                 # Go dependencies
└── README.md                       # This file
```

---

## 📦 Main Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| **Gin** | v1.9.1 | Web framework |
| **GORM** | v1.25.4 | ORM database |
| **PostgreSQL Driver** | v1.5.2 | PostgreSQL connection |
| **RabbitMQ** | v1.1.0 | Message queue client |
| **godotenv** | v1.5.1 | Load .env files |

---

## 🚀 Quick Start

### Prerequisites

- Go 1.23+
- Docker & Docker Compose
- PostgreSQL 15+
- RabbitMQ 3+

### 1. Clone the Repository

```bash
git clone https://github.com/Alibaghaee/gore_bot.git
cd gore_bot
```

### 2. Set Environment Variables

Review and modify the `.env` file:

```env
PORT=8080
DB_DSN=host=db user=postgres password=secret dbname=telebridge port=5432 sslmode=disable
RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
ADMIN_API_KEY=my_secure_token_123
```

### 3. Run with Docker Compose

```bash
docker-compose up --build
```

Services will be running at:
- **App**: http://localhost:8090 (mapped to 8080 in container)
- **RabbitMQ Management**: http://localhost:15672 (guest/guest)
- **PostgreSQL**: localhost:5432

### 4. Run Locally without Docker

```bash
# Install dependencies
go mod download

# Run the application
go run cmd/api/main.go
```

---

## 📡 API Endpoints

### 1️⃣ Register a New Bot

**Request:**
```http
POST /api/v1/bots
Authorization: Bearer my_secure_token_123
Content-Type: application/json

{
  "token": "1234567890:ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefgh",
  "slug": "my_bot",
  "owner_id": 42,
  "plan_type": "premium"
}
```

**Success Response:**
```json
{
  "id": 1,
  "token": "1234567890:ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefgh",
  "slug": "my_bot",
  "is_active": true,
  "owner_id": 42,
  "plan_type": "premium",
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Error Codes:**
- `400` - Invalid input
- `401` - Authentication failed
- `500` - Server error

---

### 2️⃣ Receive Webhook

**Request:**
```http
POST /api/v1/webhook/my_bot
Content-Type: application/json

{
  "update_id": 123456789,
  "message": {
    "message_id": 1,
    "from": {
      "id": 987654321,
      "first_name": "John",
      "username": "john_doe"
    },
    "chat": {
      "id": 987654321,
      "type": "private"
    },
    "date": 1705318200,
    "text": "Hello!"
  }
}
```

**Response:**
```json
{
  "status": "queued"
}
```

**Processing Flow:**
1. Receive message at webhook
2. Find bot by `slug`
3. Create Routing Key: `telegram.message.{owner_id}.{plan_type}`
4. Send to RabbitMQ: `telegram_events` exchange
5. Laravel consumer receives it

---

## ⚙️ Configuration and How It Works

### RabbitMQ Routing Strategy

Each Telegram message is intelligently routed:

```
Exchange: telegram_events
Routing Key Format: telegram.message.{owner_id}.{plan_type}

Examples:
- telegram.message.1.free
- telegram.message.42.premium
- telegram.message.100.enterprise
```

**Payload Sent to Queue:**
```json
{
  "bot_id": 1,
  "owner_id": 42,
  "token": "bot_token_here",
  "plan": "premium",
  "payload": {
    "update_id": 123456789,
    "message": { ... }
  }
}
```

### Database Schema

```sql
CREATE TABLE bots (
    id SERIAL PRIMARY KEY,
    token VARCHAR(255) UNIQUE NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    owner_id INTEGER NOT NULL,
    plan_type VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);
```

---

## 🔐 Security and Authentication

### Authentication Middleware

All bot registration requests require a specific header:

```http
Authorization: Bearer {ADMIN_API_KEY}
```

**Implementation:**
```go
// internal/api/middleware.go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token != "Bearer "+os.Getenv("ADMIN_API_KEY") {
            c.JSON(401, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

---

## 🛠️ Development

### Run with Hot Reload (Air)

For faster development, use Air:

```bash
# Install
go install github.com/cosmtrek/air@latest

# Run
air
```

`.air.toml` is already configured.

### Logging

The server displays detailed logs:

```
🚀 Server running on port 8080
✅ Server exited properly
⚠️ Shutting down server...
❌ RabbitMQ connection error: ...
```

---

## 📊 Workflow Diagram

```mermaid
flowchart LR
    A["Telegram"] -->|webhook| B["Gore Bot<br/>/api/v1/webhook/:slug"]
    B -->|"1. Query DB"| C["PostgreSQL<br/>Get Bot Info"]
    C -->|"2. Bot Data"| B
    B -->|"3. Create Routing Key<br/>telegram.message.{owner_id}.{plan_type}"| D["RabbitMQ"]
    D -->|"4. Forward Message"| E["Laravel Worker<br/>Queue Consumer"]
    E -->|"5. Send Response"| A
```

---

## 🐳 Docker

### Build Image for Production

```bash
docker build -f deployments/Dockerfile.prod -t gore_bot:latest .
```

### Run Container

```bash
docker run -d \
  --name gore_bot \
  -p 8080:8080 \
  --env-file .env \
  gore_bot:latest
```

---

## 🧪 Testing

### Test Endpoints with cURL

```bash
# Register bot
curl -X POST http://localhost:8090/api/v1/bots \
  -H "Authorization: Bearer my_secure_token_123" \
  -H "Content-Type: application/json" \
  -d '{
    "token": "test_token",
    "slug": "test_bot",
    "owner_id": 1,
    "plan_type": "free"
  }'

# Send Webhook
curl -X POST http://localhost:8090/api/v1/webhook/test_bot \
  -H "Content-Type: application/json" \
  -d '{
    "update_id": 1,
    "message": {
      "text": "Hello!"
    }
  }'
```

---

## 📋 Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `DB_DSN` | PostgreSQL connection string | `host=localhost user=postgres password=secret dbname=telebridge` |
| `RABBITMQ_URL` | RabbitMQ connection URL | `amqp://guest:guest@localhost:5672/` |
| `ADMIN_API_KEY` | API key for endpoint authentication | `my_secure_token_123` |

---

## 🐛 Troubleshooting

### Error: "Database connection refused"
- Are RabbitMQ and PostgreSQL running?
- Is `DB_DSN` correct?
```bash
docker-compose ps
```

### Error: "RabbitMQ connection error"
- Check RabbitMQ status:
```bash
docker-compose logs rabbitmq
```

### Server won't start
- Check logs:
```bash
docker-compose logs app
```

---

## 📝 License

This project is released under the MIT License. For more details, see the `LICENSE` file.

---

## 👤 Author

**Alibaghaee**
- GitHub: [@Alibaghaee](https://github.com/Alibaghaee)

---

## 💡 Important Notes

- 🔄 **Goroutines**: Messages are sent to RabbitMQ asynchronously
- 📊 **Stateless**: Each instance can work independently
- 🔒 **Security**: Make sure to change the API Key
- 🚀 **Scalability**: Architecture is designed to be scalable

---

## 📞 Feedback and Bug Reports

If you found an issue or have a suggestion:
- [Issues](https://github.com/Alibaghaee/gore_bot/issues)
- [Pull Requests](https://github.com/Alibaghaee/gore_bot/pulls)

---

**Built with ❤️ in Go**

---

This English README includes:

✅ Complete project description  
✅ Key features  
✅ Detailed architecture  
✅ Quick start guide (Docker & Local)  
✅ Full API Endpoints with examples  
✅ RabbitMQ routing explanation  
✅ Authentication and security  
✅ Development guide  
✅ Testing instructions  
✅ Troubleshooting guide  