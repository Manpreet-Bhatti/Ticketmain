# Ticketmain (High-Concurrency Ticket Engine)

## 🚀 The Core Concept

A robust, full-stack simulation of a high-demand ticketing platform (such as Ticketmaster) designed to handle extreme concurrency and prevent double bookings.

**The Challenge:** In a standard web app, two users clicking "Buy" at the same time is fairly rare. In a ticket launch, it's guaranteed to happen. This project implements a **distributed locking mechanism** to prevent race conditions and "double bookings" when thousands of concurrent requests target the same inventory.

## 🛠️ Tech Stack & Architecture
 - **Frontend:** React (Vite + TypeScript) with Optimistic UI updates
 - **Backend:** Go (Fiber) for high-performance HTTP & WebSocket handling
 - **State Management:** Redis (for atomic `SET NX` locking and real-time pub/sub)
 - **Persistence:** PostgreSQL (for ACID-compliant transaction recording)
 - **Infrastructure:** Docker & Docker Compose

## ⚡️ Key Features
 - **Real-Time Seat Map:** WebSocket integration pushes seat status updates (Held/Sold) to all connected users instantly
 - **Atomic Locking:** Uses Redis Lua scripts to ensure hold/release operations are atomic and race-condition free
 - **Ephemeral Identity:** Client-side UUID generation to simulate unique users without complex auth guardrails
 - **Performance:** Utilizes Redis Pipelining to batch fetch thousands of seat states in a single network roundtrip

## 📦 Installation & Setup
Follow these steps to run the system locally.

1. Prerequisites:
- **Docker Desktop** (Running)
- **Go** (v1.21 or higher)
- **Node.js** (v18 or higher)

2. Start Infrastructure

Create the Redis and PostgreSQL containers.

```bash
# From root directory
docker-compose up -d
```

3. Start the Backend (Go)

This runs the API server at `http://localhost:3000`. The application will automatically create the necessary database tables on startup.

```bash
cd server
go mod tidy
go run main.go
```
*You should see: `✅ Connected to Redis` and `✅ Connected to Postgres and ensured table exists`*.

4. Start the Frontend (React)

This runs the UI at `http://localhost:5173`.

```bash
# Open a new terminal
cd client
npm install
npm run dev
```

## 🧪 How to Test Concurrency
1. Open the app in **Chrome** (or any other browser of your choice)
2. Open the app in **Firefox** (an incognito window, or any other browser of your choice)
3. Notice that you'll have a different `User: user_xyz` ID in each window
4. Click a seat in Chrome (first window)
- **Result:** Instant update in Firefox (second window, turns grey)
5. Try to click the *same* seat in Firefox (second window)
- **Result:** "Seat already held by another user" error.
