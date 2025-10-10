# 🌙 LunaSentri

**Production-ready server monitoring for solo developers and small teams**

LunaSentri is a lightweight, self-hosted monitoring dashboard that provides real-time server metrics, intelligent alerting, and multi-channel notifications. Built with Go and Next.js for performance and simplicity.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## ✨ Features

### Core Monitoring

- **Multi-Machine Support**: Monitor multiple servers from a single dashboard (NEW!)
- **Lightweight Agent**: Install on any Linux server to stream metrics
- **Real-time Metrics**: CPU, Memory, Disk, and Network usage with WebSocket streaming
- **System Information**: OS, kernel, architecture, and runtime details
- **Live Dashboard**: Beautiful dark-themed interface with real-time charts
- **Persistent Storage**: SQLite database for historical data and configurations

### Intelligent Alerting

- **Flexible Alert Rules**: Define custom thresholds for any metric
- **Alert Management**: Create, update, and delete alert rules via UI
- **Event Tracking**: View alert history and acknowledge events
- **Configurable Triggers**: Set trigger conditions (above/below thresholds)

### Multi-Channel Notifications

- **Webhook Integration**: Send alerts to any HTTP endpoint
- **Telegram Bot**: Native Telegram notifications with instant delivery
- **Rate Limiting**: Prevent notification spam with configurable cooldowns
- **Test Functionality**: Verify notifications before going live

### Security & Authentication

- **JWT-based Auth**: Secure session management with configurable TTL
- **Admin System**: Role-based access control with admin privileges
- **First-User Admin**: Automatic admin promotion for first registered user
- **Password Management**: Secure password hashing with bcrypt
- **Password Reset**: Token-based password recovery flow

## 🛠️ Tech Stack

### Backend (`apps/api-go`)

- **Go 1.24** - High-performance, minimal HTTP server
- **Standard Library** - No external frameworks (net/http only)
- **SQLite** - Embedded database with zero configuration
- **WebSocket** - Real-time metric streaming
- **JWT** - Stateless authentication

### Frontend (`apps/web-next`)

- **Next.js 15** - React framework with App Router
- **React 19** - Latest React with Server Components
- **Tailwind CSS v4** - Modern utility-first styling
- **TypeScript** - Type-safe development
- **Recharts** - Beautiful, responsive charts
- **pnpm** - Fast, disk-space efficient package manager

### Deployment

- **Docker** - Multi-stage builds for both apps
- **CapRover** - PaaS deployment with zero-downtime updates
- **Alpine Linux** - Minimal container images
- **Health Checks** - Container health monitoring

## 📁 Project Structure

```
lunasentri/
├── apps/
│   ├── agent/                     # Monitoring agent (NEW!)
│   │   ├── main.go               # Agent entry point
│   │   ├── internal/
│   │   │   ├── config/           # Configuration loading
│   │   │   ├── collector/        # Metrics collection
│   │   │   └── transport/        # API communication
│   │   ├── scripts/
│   │   │   └── install.sh        # Linux installation script
│   │   └── Dockerfile            # Docker image
│   │
│   ├── api-go/                    # Go backend
│   │   ├── main.go               # Server entry point
│   │   ├── internal/
│   │   │   ├── auth/             # Authentication & authorization
│   │   │   ├── alerts/           # Alert rule engine
│   │   │   ├── machines/         # Machine registry & metrics
│   │   │   ├── metrics/          # System metrics collection
│   │   │   ├── notifications/    # Webhook & Telegram notifiers
│   │   │   ├── storage/          # SQLite database layer
│   │   │   └── system/           # System information
│   │   └── Dockerfile            # Production container
│   │
│   └── web-next/                  # Next.js frontend
│       ├── app/                   # App Router pages
│       ├── components/            # React components
│       ├── lib/                   # Utilities & API client
│       └── Dockerfile             # Production container
│
├── deploy/                        # Deployment configurations
│   └── caprover/                  # CapRover deployment files
│       ├── backend/
│       └── frontend/
│
├── docs/                          # Documentation (see below)
│   ├── agent/                     # Agent documentation
│   │   └── INSTALLATION.md       # Installation guide
├── deploy.sh                      # Deployment script
├── CLAUDE.md                      # AI assistant context
└── README.md                      # This file
```

## 🚀 Quick Start

### Prerequisites

- **Go 1.24+** for backend development
- **Node.js 20+** and **pnpm** for frontend
- **Docker** (optional, for containerized deployment)

### Local Development

#### Option 1: Using Development Scripts (Recommended)

The easiest way to run the full stack with proper environment loading:

1. **Clone the repository**

   ```bash
   git clone https://github.com/Constantin-E-T/lunasentri.git
   cd lunasentri
   ```

2. **Create `.env` file** (optional, for Telegram notifications)

   ```bash
   # .env
   TELEGRAM_BOT_TOKEN=your_bot_token_here
   ```

3. **Install frontend dependencies**

   ```bash
   pnpm install
   ```

4. **Start everything**

   ```bash
   ./scripts/dev/start.sh
   ```

   This will:
   - Load environment variables from `.env`
   - Start the Go API server on port 80
   - Start the Next.js frontend on port 3002
   - Enable Telegram notifications if token is set

5. **Access the dashboard**
   - Open `http://localhost:3002`
   - Register a new account (first user becomes admin)

#### Option 2: Manual Setup

1. **Clone the repository**

   ```bash
   git clone https://github.com/Constantin-E-T/lunasentri.git
   cd lunasentri
   ```

2. **Install frontend dependencies**

   ```bash
   pnpm install
   ```

3. **Start the backend** (Terminal 1)

   ```bash
   cd apps/api-go
   AUTH_JWT_SECRET="test-secret-key-for-development-only-32chars" \
   SECURE_COOKIE=false \
   ADMIN_EMAIL="admin@test.com" \
   ADMIN_PASSWORD="admin123" \
   go run main.go
   ```

   Backend runs on `http://localhost:80`

4. **Start the frontend** (Terminal 2)

   ```bash
   cd apps/web-next
   pnpm dev
   ```

   Frontend runs on `http://localhost:3000`

5. **Access the dashboard**
   - Open `http://localhost:3000`
   - Login with admin credentials: `admin@test.com` / `admin123`

> **Note**: For Telegram notifications, set `TELEGRAM_BOT_TOKEN` environment variable before running the API server. See [scripts/dev/README.md](scripts/dev/README.md) for details.

### Production Deployment

See [docs/deployment/DEPLOYMENT.md](docs/deployment/DEPLOYMENT.md) for complete deployment instructions including:

- CapRover setup and configuration
- Docker build and deployment
- Environment variable management
- Database persistence
- HTTPS and custom domains

### Installing the Monitoring Agent

To monitor remote servers, install the LunaSentri agent:

1. **Register a machine** in the web dashboard to get an API key
2. **Install on your server**:

   ```bash
   curl -fsSL https://raw.githubusercontent.com/Constantin-E-T/lunasentri/main/apps/agent/scripts/install.sh | sudo bash
   ```

3. **Verify it's running**:

   ```bash
   sudo systemctl status lunasentri-agent
   ```

See [docs/agent/INSTALLATION.md](docs/agent/INSTALLATION.md) for detailed installation instructions, Docker usage, and configuration options.

## 📚 Documentation

### Essential Docs

- **[QUICK_START.md](QUICK_START.md)** - Get running in 5 minutes
- **[DEPLOYMENT.md](docs/deployment/DEPLOYMENT.md)** - Production deployment guide
- **[CLAUDE.md](CLAUDE.md)** - AI assistant development context

### Feature Documentation

- **[Authentication & Users](docs/features/auth-users.md)** - User management and security
- **[Alert System](docs/features/alerts.md)** - Alert rules and event tracking
- **[Notifications](docs/features/notifications.md)** - Webhook and Telegram setup

### Development Reports

See [docs/development/](docs/development/) for detailed implementation reports:

- Telegram notification implementation
- Webhook notification system
- Dev workflow improvements
- Email notifications removal

## 🔐 Security

### Authentication Flow

1. **Bootstrap Admin**: First user or environment variable creates admin
2. **JWT Tokens**: Secure session tokens with configurable expiry
3. **Secure Cookies**: HTTPOnly, Secure flags for production
4. **Password Reset**: Token-based recovery with TTL

### Environment Variables

#### Backend (Required)

```bash
AUTH_JWT_SECRET=your-secret-key-min-32-chars    # JWT signing key
ADMIN_EMAIL=admin@example.com                   # Bootstrap admin email
ADMIN_PASSWORD=securepassword                   # Bootstrap admin password
```

#### Backend (Optional)

```bash
PORT=80                                         # Server port (default: 80)
DB_PATH=./data/lunasentri.db                   # Database location
CORS_ALLOWED_ORIGIN=https://your-domain.com    # CORS origin
SECURE_COOKIE=true                             # Secure cookie flag
ACCESS_TOKEN_TTL=15m                           # JWT expiry
PASSWORD_RESET_TTL=1h                          # Reset token expiry
```

#### Telegram Notifications (Optional)

```bash
TELEGRAM_BOT_TOKEN=your-bot-token              # From @BotFather
```

#### Frontend (Required)

```bash
NEXT_PUBLIC_API_URL=https://api.example.com    # Backend URL
```

## 🎯 Admin User Setup

LunaSentri provides three ways to create admin users:

1. **Environment Variables** (Recommended for production)
   - Set `ADMIN_EMAIL` and `ADMIN_PASSWORD` before first run
   - User is created/updated on server start

2. **First Registration** (Development)
   - First user to register automatically becomes admin
   - Navigate to `/register` and create account

3. **Database Access** (Manual)
   - SSH into container: `docker exec -it <container> /bin/sh`
   - Update user: `sqlite3 /app/data/lunasentri.db`
   - Run: `UPDATE users SET is_admin = 1 WHERE email = 'user@example.com';`

## 🐳 Docker Deployment

### Backend

```bash
cd apps/api-go
docker build -t lunasentri-api .
docker run -p 80:80 \
  -e AUTH_JWT_SECRET="your-secret" \
  -e ADMIN_EMAIL="admin@test.com" \
  -e ADMIN_PASSWORD="admin123" \
  -e SECURE_COOKIE=true \
  -v $(pwd)/data:/app/data \
  lunasentri-api
```

### Frontend

```bash
cd apps/web-next
docker build -t lunasentri-web \
  --build-arg NEXT_PUBLIC_API_URL=https://api.example.com .
docker run -p 80:80 lunasentri-web
```

## 🔧 Development Commands

### Backend

```bash
# Run with environment variables
go run main.go

# Run tests
go test ./...

# Build binary
go build -o lunasentri-api

# Format code
go fmt ./...

# Vet code
go vet ./...
```

### Frontend

```bash
# Development server
pnpm dev

# Production build
pnpm build

# Start production server
pnpm start

# Run tests
pnpm test

# Type checking
tsc --noEmit
```

## 📊 API Endpoints

### Public Endpoints

- `POST /auth/register` - Register new user
- `POST /auth/login` - Login and get session
- `POST /auth/logout` - Logout and clear session
- `POST /auth/forgot-password` - Request password reset
- `POST /auth/reset-password` - Reset password with token

### Protected Endpoints (Requires Auth)

- `GET /auth/me` - Get current user profile
- `POST /auth/change-password` - Change password
- `GET /metrics` - Get current system metrics
- `GET /ws` - WebSocket for real-time metrics
- `GET /system/info` - Get system information

### Admin Endpoints (Requires Admin Role)

- `GET /auth/users` - List all users
- `POST /auth/users` - Create new user
- `DELETE /auth/users/:id` - Delete user
- `GET /alerts/rules` - List alert rules
- `POST /alerts/rules` - Create alert rule
- `PUT /alerts/rules/:id` - Update alert rule
- `DELETE /alerts/rules/:id` - Delete alert rule
- `GET /alerts/events` - List alert events
- `POST /alerts/events/:id/ack` - Acknowledge alert
- `GET /notifications/webhooks` - List webhooks
- `POST /notifications/webhooks` - Create webhook
- `PUT /notifications/webhooks/:id` - Update webhook
- `DELETE /notifications/webhooks/:id` - Delete webhook
- `POST /notifications/webhooks/:id/test` - Test webhook
- `GET /notifications/telegram` - List Telegram recipients
- `POST /notifications/telegram` - Create Telegram recipient
- `PUT /notifications/telegram/:id` - Update Telegram recipient
- `DELETE /notifications/telegram/:id` - Delete Telegram recipient
- `POST /notifications/telegram/:id/test` - Test Telegram

## 🤝 Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with ❤️ for solo developers and small teams
- Inspired by the need for simple, effective monitoring
- Powered by Go's performance and Next.js's elegance

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/Constantin-E-T/lunasentri/issues)
- **Documentation**: See `docs/` directory
- **Security**: Report vulnerabilities privately to <security@example.com>

---

**Made with 🌙 by the LunaSentri Team**
