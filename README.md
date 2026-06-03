# Random City Picker

A full-stack application that picks a random city from a chosen country (starting with France). Cities are tracked so they aren't shown again until all have been exhausted. Built with React, Go, and SQLite — containerized with Docker for easy development and deployment.

## Features

- Pick random cities from a country with configurable min/max population filters
- Cities are exhausted before being re-shown (tracked in SQLite)
- When all cities have been picked, the app cycles back to least-picked cities
- Full pick history with timestamps and pick counts
- Docker & Docker Compose ready for development and production

## Tech Stack

- **Frontend**: React + TypeScript + Vite
- **Backend**: Go 1.22+ with standard `net/http`
- **Database**: SQLite (pure Go, no CGO)
- **Containers**: Docker + Docker Compose

## Quick Start

### Local Development (without Docker)

```bash
# Terminal 1 — Backend
cd backend
go mod tidy
go run .
# API runs on http://localhost:8080

# Terminal 2 — Frontend
cd frontend
npm install
npm run dev
# App runs on http://localhost:5173
```

### Development with Docker Compose

```bash
docker compose up
```

- Frontend: http://localhost:5173
- Backend API: http://localhost:8080

### Production with Docker Compose

```bash
docker compose -f docker-compose.prod.yml up --build
```

- App served on http://localhost (nginx reverse-proxies `/api` to the backend)

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/cities/random?country=FR&minPop=0&maxPop=9999999` | Pick a random city |
| GET | `/api/cities/picked` | List all picked cities with counts |
| POST | `/api/cities/reset` | Reset all pick history |

## Adding More Countries

1. Create a new JSON file in `backend/db/data/cities_XX.json` with the format:
   ```json
   [
     {"name": "City Name", "population": 123456, "latitude": 12.34, "longitude": 56.78}
   ]
   ```
2. Update `backend/db/seed.go` to also embed and seed the new file.
3. Add the new country code to the frontend dropdown.

## Project Structure

```
.
├── backend/
│   ├── main.go
│   ├── db/
│   │   ├── db.go
│   │   ├── seed.go
│   │   ├── schema.sql
│   │   └── data/cities_fr.json
│   ├── handlers/
│   │   └── cities.go
│   ├── models/
│   │   └── city.go
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── App.tsx
│   │   ├── api.ts
│   │   └── components/
│   │       ├── CityPicker.tsx
│   │       └── PickHistory.tsx
│   ├── index.html
│   ├── package.json
│   ├── nginx.conf
│   └── Dockerfile
├── docker-compose.yml
├── docker-compose.prod.yml
└── README.md
```
