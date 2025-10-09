# LunaSentri Documentation

Complete documentation for the LunaSentri server monitoring system.

---

## Essential Documentation

### Quick Start

- **[README.md](../README.md)** - Main project documentation and overview
- **[QUICK_START.md](../QUICK_START.md)** - Get up and running in 5 minutes
- **[CLAUDE.md](../CLAUDE.md)** - AI assistant development context

### Deployment

- **[DEPLOYMENT.md](deployment/DEPLOYMENT.md)** - Production deployment guide (CapRover, Docker, environment variables)

---

## Feature Documentation

### Core Features

- **[Authentication & Users](features/auth-users.md)** - User management, JWT authentication, password reset, admin setup
- **[Alert System](features/alerts.md)** - Alert rules, event tracking, thresholds, acknowledgments
- **[Notifications](features/notifications.md)** - Webhook and Telegram notification overview

### Implementation Reports

- **[Telegram Notifications](features/implementation/telegram-notifications.md)** - Complete Telegram bot implementation (backend + frontend)
- **[Webhook Notifications](features/implementation/webhook-notifications.md)** - Webhook system with HMAC signatures and rate limiting
- **[Email Removal](features/implementation/email-removal.md)** - Email notification system removal report

---

## Development Documentation

### For Developers

- **[Workflow Improvements](development/workflow-improvements.md)** - Database persistence, port management, error handling fixes
- **[Agent Reports](development/agent-reports.md)** - AI-assisted development documentation and guidelines

### Architecture

- **Backend**: Go 1.24 with standard library (no frameworks)
- **Frontend**: Next.js 15 with App Router and Tailwind CSS v4
- **Database**: SQLite with automatic migrations
- **Authentication**: JWT with HTTPOnly cookies
- **Real-time**: WebSocket for metrics streaming

---

## Quick Links

### Getting Started

1. [Install dependencies](../README.md#quick-start)
2. [Start backend](../README.md#local-development)
3. [Start frontend](../README.md#local-development)
4. [Access dashboard](../README.md#local-development)

### Configuration

- [Environment Variables](deployment/DEPLOYMENT.md#environment-variables)
- [Admin Setup](features/auth-users.md#admin-users)
- [Notification Setup](features/notifications.md)

### Development

- [Development Commands](development/agent-reports.md#development-commands)
- [Dev Workflow](development/workflow-improvements.md)
- [API Endpoints](../README.md#api-endpoints)

---

## Documentation Structure

```
docs/
├── README.md                          # This file - documentation index
├── deployment/
│   └── DEPLOYMENT.md                  # Production deployment guide
├── features/
│   ├── auth-users.md                  # Authentication & user management
│   ├── alerts.md                      # Alert system guide
│   ├── notifications.md               # Notification system overview
│   └── implementation/
│       ├── telegram-notifications.md  # Telegram implementation details
│       ├── webhook-notifications.md   # Webhook implementation details
│       └── email-removal.md          # Email removal report
└── development/
    ├── workflow-improvements.md       # Dev workflow fixes
    └── agent-reports.md              # AI agent development reports
```

---

## Contributing

When adding documentation:

1. Place feature docs in `docs/features/`
2. Place implementation reports in `docs/features/implementation/`
3. Place development guides in `docs/development/`
4. Place deployment guides in `docs/deployment/`
5. Update this README.md with links
6. Use relative paths for all links

---

## Need Help?

- **Issues**: [GitHub Issues](https://github.com/Constantin-E-T/lunasentri/issues)
- **Security**: Report vulnerabilities privately
- **Development**: Check [agent-reports.md](development/agent-reports.md) for AI development guidelines

---

**Made with 🌙 by the LunaSentri Team**
