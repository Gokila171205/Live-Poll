# LivePoll

LivePoll is a real-time polling application designed with a production-grade architecture featuring a clean separation of concerns between frontend and backend.

## Tech Stack
- **Frontend**: React (Vite)
- **Backend**: Go with Gin
- **Database**: MongoDB
- **Realtime / Cache**: Redis

## Project Structure
```text
live-poll/
├── frontend/             # React single-page application
│   ├── src/              # React components, state, and styles
│   ├── .env.example      # Frontend environment template
│   └── package.json      # Dependencies and scripts
├── backend/              # Go REST & WebSocket backend service
│   ├── cmd/server/       # Application entry point (main.go)
│   ├── internal/         # Private application packages
│   │   ├── config/       # Environment variable loader
│   │   ├── database/     # MongoDB connection management
│   │   ├── handlers/     # HTTP handlers / controllers
│   │   ├── middleware/   # Auth, logging, recovery, CORS
│   │   ├── models/       # Data structures & domain entities
│   │   ├── redis/        # Redis client & connection management
│   │   ├── repository/   # Data access logic & database queries
│   │   ├── services/     # Business logic & health checks
│   │   └── websocket/    # Real-time WebSocket hubs
│   ├── .env.example      # Backend environment template
│   └── go.mod            # Go module definitions
├── README.md             # Project documentation
└── .gitignore            # Git ignore rules for React, Go, and secrets
```
