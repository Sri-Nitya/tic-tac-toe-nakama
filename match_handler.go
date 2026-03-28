package main

import (
	"context"
	"database/sql"

	"github.com/heroiclabs/nakama-common/runtime"
)

const matchLabel = "tic-tac-toe"

type MatchHandler struct{}

type MatchState struct {
	players map[string]runtime.Presence
	marks  map[string]string
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

	// logger.Info("JOIN ATTEMPT: %s", presence.GetUsername())

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
		} else if len(s.players) == 2 {
			s.marks[p.GetUserId()] = "O"
		}

		logger.Info("Player joined: %s, assigned mark: %s", p.GetUsername(), s.marks[p.GetUserId()])
		// log number of players in the match
		logger.Info("Number of players in match: %d", len(s.players))
	}

	return s
}

func (m *MatchHandler) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {

	s := state.(*MatchState)

	for _, p := range presences {
		delete(s.players, p.GetUserId())
		logger.Info("Player left: %s", p.GetUsername())
	}

	return s
}

func (m *MatchHandler) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {

	return state
}

func (m *MatchHandler) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {

	logger.Info("MATCH TERMINATE")

	return state
}

func (m *MatchHandler) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {

	logger.Info("MATCH SIGNAL: %s", data)

	return state, ""
}
