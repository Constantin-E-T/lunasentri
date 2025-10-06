# LunaSentri

Lightweight server monitoring dashboard for solo developers — real-time health, app statuses, and AI-powered optimization suggestions.

## 🎯 Project Goal

LunaSentri provides a simple, elegant way for solo developers to monitor their servers and applications. With real-time health metrics, application status tracking, and AI-powered optimization suggestions, it helps you keep your infrastructure running smoothly without the complexity of enterprise monitoring tools.

## 🛠️ Tech Stack

### Backend (`apps/api-go`)
- **Go** - High-performance backend API
- RESTful API for monitoring data collection and retrieval

### Frontend (`apps/web-next`)
- **Next.js 15** - React framework for production
- **Tailwind CSS** - Utility-first CSS framework
- Modern, responsive dashboard interface

## 📁 Project Structure

```
lunasentri/
├── apps/
│   ├── api-go/        # Go backend API
│   └── web-next/      # Next.js frontend
├── LICENSE            # MIT License
└── README.md
```

## 🚀 Getting Started

See [docs/LOCAL_DEV.md](docs/LOCAL_DEV.md) for complete development setup instructions.

### Quick Start

1. **Clone and install dependencies**:
   ```bash
   git clone <repository-url>
   cd lunasentri
   pnpm install
   ```

2. **Configure environment** - Create `apps/web-next/.env.local`:
   ```bash
   NEXT_PUBLIC_API_URL=http://localhost:8080
   ```

3. **Start the backend** (Terminal 1):
   ```bash
   cd apps/api-go
   go run main.go
   ```

4. **Start the frontend** (Terminal 2):
   ```bash
   pnpm dev:web
   ```

5. **Open** `http://localhost:3000` to see the live metrics dashboard

For detailed setup, troubleshooting, and environment variable configuration, see [docs/LOCAL_DEV.md](docs/LOCAL_DEV.md).

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.
