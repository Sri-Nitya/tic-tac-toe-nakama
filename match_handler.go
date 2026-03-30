package main

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/heroiclabs/nakama-common/runtime"
)

const matchLabel = "tic-tac-toe"

var winConditions = [8][3][2]int{
	{{0, 0}, {0, 1}, {0, 2}}, // Row 1
	{{1, 0}, {1, 1}, {1, 2}}, // Row 2
	{{2, 0}, {2, 1}, {2, 2}}, // Row 3
	{{0, 0}, {1, 0}, {2, 0}}, // Column 1
	{{0, 1}, {1, 1}, {2, 1}}, // Column 2
	{{0, 2}, {1, 2}, {2, 2}}, // Column 3
	{{0, 0}, {1, 1}, {2, 2}}, // Diagonal \
	{{0, 2}, {1, 1}, {2, 0}}, // Diagonal /
}	

type MatchHandler struct{}

type MatchState struct {
	players map[string]runtime.Presence
	marks  map[string]string
	board  [3][3]string
	currentTurn string
	gameOver    bool
	winner      string
}

type Move struct {
	Row int `json:"row"`
	Col int `json:"col"`	
}

func broadcastState(dispatcher runtime.MatchDispatcher, s *MatchState) {
	response := map[string]interface{}{
		"board":    s.board,
		"turn":     s.marks[s.currentTurn],
		"players":  s.marks,
		"gameOver": s.gameOver,
		"winner":   s.winner,
	}

	responseBytes, _ := json.Marshal(response)
	_ = dispatcher.BroadcastMessage(1, responseBytes, nil, nil, true)
}

func broadcastResult(dispatcher runtime.MatchDispatcher, payload map[string]interface{}) {
	resultBytes, _ := json.Marshal(payload)
	_ = dispatcher.BroadcastMessage(2, resultBytes, nil, nil, true)
}

func (m *MatchHandler) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {

	logger.Info("MATCH INIT")

	state := &MatchState{
		players: make(map[string]runtime.Presence),
		marks:   make(map[string]string),
	}

	return state, 1, matchLabel
}

func (m *MatchHandler) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {

	if len(state.(*MatchState).players) >= 2 {
		logger.Info("JOIN ATTEMPT REJECTED: MATCH FULL")
		return state, false, "Match is full"
	}

	return state, true, ""
}

func (m *MatchHandler) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {

	s := state.(*MatchState)

	for _, p := range presences {
		s.players[p.GetUserId()] = p

		if len(s.players) == 1 {
			s.marks[p.GetUserId()] = "X"
			s.currentTurn = p.GetUserId()
			logger.Info("First player joined: %s as X", p.GetUsername())
		} else if len(s.players) == 2 {
			s.marks[p.GetUserId()] = "O"
			logger.Info("Second player joined: %s as O", p.GetUsername())
		}
	}

	broadcastState(dispatcher, s)
	return s
}

func (m *MatchHandler) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {

	s := state.(*MatchState)

	if s.gameOver {
		for _, p := range presences {
			delete(s.players, p.GetUserId())
		}
		return nil
	}

	for _, p := range presences {
		leftUserID := p.GetUserId()
		leftUsername := p.GetUsername()

		delete(s.players, leftUserID)

		logger.Info("Player left: %s", leftUsername)

		if len(s.players) == 1 {
			for remainingUserID, remainingPresence := range s.players {
				s.gameOver = true
				s.winner = remainingPresence.GetUserId()

				broadcastResult(dispatcher, map[string]interface{}{
					"type":   "disconnect",
					"winner": s.winner,
					"message": "Opponent disconnected",
				})

				logger.Info("Opponent disconnected. Winner: %s (%s)", s.winner, remainingUserID)
				break
			}
		}
	}

	return s
}

func (m *MatchHandler) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {

	s := state.(*MatchState)

	for _, msg := range messages {
		if len(s.players) < 2 || s.gameOver {
			logger.Info("Ignoring move")
			continue
		}

		var move Move
		err := json.Unmarshal(msg.GetData(), &move)
		if err != nil {
			logger.Error("Invalid move data: %v", err)
			continue
		}
		
		userId := msg.GetUserId()
		logger.Info("Received move from player %s: row %d, col %d", userId, move.Row, move.Col)

		if move.Row < 0 || move.Row > 2 || move.Col < 0 || move.Col > 2 {
			logger.Info("Invalid move: row and col must be between 0 and 2")
			continue
		}

		// Check if it's the player's turn
		if s.currentTurn != "" && s.currentTurn != userId {
			logger.Info("It's not player %s's turn", userId)
			continue
		}

		// Check if cell is empty
		if s.board[move.Row][move.Col] != "" {
			logger.Info("Cell (%d, %d) is already occupied", move.Row, move.Col)
			continue
		}

		mark := s.marks[userId]
		if mark == "" {
			logger.Info("Player %s has no assigned mark", userId)
			continue
		}

		s.board[move.Row][move.Col] = mark
		logger.Info("Player %s placed %s at (%d, %d)",userId, mark, move.Row, move.Col)

		win := false
		for _, condition := range winConditions {
			if s.board[condition[0][0]][condition[0][1]] == mark &&
				s.board[condition[1][0]][condition[1][1]] == mark &&
				s.board[condition[2][0]][condition[2][1]] == mark {
				win = true
				break
			}
		}

		if win {
			s.gameOver = true
			s.winner = userId

			broadcastState(dispatcher, s)
			broadcastResult(dispatcher, map[string]interface{}{
				"type":   "win",
				"winner": s.winner,
			})

			logger.Info("Player %s wins!", s.winner)
			return nil
		}

		isBoardFull := true
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				if s.board[i][j] == "" {
					isBoardFull = false
					break
				}
			}
		}

		if isBoardFull {
			s.gameOver = true

			broadcastState(dispatcher, s)
			broadcastResult(dispatcher, map[string]interface{}{
				"type":    "draw",
				"message": "It's a draw",
			})

			logger.Info("Board is full, it's a draw")
			return nil
		}

		for player := range s.players {
			if player != userId {
				s.currentTurn = player
				break
			}
		}

		broadcastState(dispatcher, s)
	}

	return s
}

func (m *MatchHandler) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {

	logger.Info("MATCH TERMINATE")

	return state
}

func (m *MatchHandler) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {

	logger.Info("MATCH SIGNAL: %s", data)

	return state, ""
}
