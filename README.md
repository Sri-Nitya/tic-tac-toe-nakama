# Tic Tac Toe Backend (Nakama + Go)

Live Game: https://tic-tac-toe-frontend-lovat.vercel.app/

## ⚠️ Note on Initial Load

The backend is hosted on Render free tier, which may cause a short delay (cold start) when the service is inactive.

- On first load, the connection may fail initially
- Please wait a few seconds and refresh the page once or twice

After the backend wakes up, the application works normally without issues.

Deployed Nakama server: https://tic-tac-toe-nakama-itj6.onrender.com

This repository contains the Nakama backend for a real-time multiplayer Tic Tac Toe Game. 
It implements server-authoritative game logic, matchmaking, concurrent game support and Leaderboard system.

Frontend Repository: https://github.com/Sri-Nitya/tic-tac-toe-frontend

## Setup and Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/Sri-Nitya/tic-tac-toe-nakama.git
   cd tic-tac-toe-nakama
   ```
2. Start the server
   ```bash
   docker-compose up --build
   ```
3. The Nakama server will start on http://localhost:7350
4. Refer [Frontend](https://github.com/Sri-Nitya/tic-tac-toe-frontend) for further setup

## Architecture and design decisions

### Server-Authoritative Design
All game logic is handled on the server to prevent cheating and ensure consistency.

The server is responsible for:
- Validating moves
- Maintaining board state
- Enforcing turn order
- Determining win/draw conditions

### Match Lifecycle
Each game runs as a Nakama match instance:
- `MatchInit` → initializes state
- `MatchJoin` → assigns player roles (X/O)
- `MatchLoop` → processes moves
- `MatchLeave` → handles disconnects
- `MatchTerminate` → cleanup

### Matchmaking
- Players can create a new match or join an existing one.
- Players can find open rooms and join

## Deployment

### Local Development
For local development, Nakama and CockroachDB are run using `docker-compose`.  
CockroachDB runs as a local container, and Nakama connects to it using the internal Docker service address.

### Production Deployment
The backend is deployed on Render as a Docker-based service.
- Database is hosted on [CockroachDB Cloud](https://cockroachlabs.cloud/)
- The connection string is configured in Render
- Nakama connects to the database using the provided database URL

## API / Server Configuration
### RPC Endpoints
- create_match → creates a new match and returns match ID
- quick_match → joins an available match or creates a new one
- get_leaderboard → returns player statistics

## Multiplayer Testing

### To test multiplayer functionality:

Open the [application](https://tic-tac-toe-frontend-gfvvs7v4a-sri-nityas-projects.vercel.app/) on two devices or clone the frontend repository and run it locally:

Player 1:
Create a match

Player 2:
Join via quick match or match ID or from the open rooms.

Verify:
- Turn-based gameplay
- Win/draw/lose logic
- Real-time updates
Test disconnect:
- Close one player
- Remaining player should win

⚠️ Note
Opening multiple tabs in the same browser shares session storage and may behave as a single player. Use separate browser sessions for accurate multiplayer testing.

## Leaderboard system
- Tracks wins, loses and maintains win streak
- Ranking system
- Shows top players on leaderboard

## Additional Notes
- Supports multiple concurrent matches
- Each match runs independently with isolated state
- Server ensures consistent and fair gameplay across all sessions
