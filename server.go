package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"poker/player"
	"strings"
	"sync"
)

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

const jsonContentType = "application/json"

type PlayerStore interface {
	GetPlayerScore(name string) (int, error)
	RecordWin(name string) error
	GetLeague() ([]player.Player, error)
}

type PlayerServer struct {
	store PlayerStore
	// it is an embeded (field?) our type now implement this interface, but we need to assign object/function
	// with required methods of interface. In this case ServeHTTP and http.NewServerMux does that
	http.Handler
}

func NewPlayerServer(store PlayerStore) *PlayerServer {
	server := new(PlayerServer)

	server.store = store
	router := http.NewServeMux()

	router.Handle("/league", http.HandlerFunc(server.leagueHandler))
	router.Handle("/players/", http.HandlerFunc(server.playersHandler))

	server.Handler = router

	return server
}

func (p *PlayerServer) leagueHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	players, err := p.store.GetLeague()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error: %s", err)
		return
	}

	w.Header().Set("content-type", jsonContentType)
	json.NewEncoder(w).Encode(players)
}

func (p *PlayerServer) playersHandler(w http.ResponseWriter, r *http.Request) {
	player := strings.TrimPrefix(r.URL.String(), "/players/")

	switch r.Method {
	case http.MethodGet:
		p.showScore(w, r, player)
	case http.MethodPost:
		p.processWin(w, r, player)
	}
}

func (p *PlayerServer) showScore(w http.ResponseWriter, r *http.Request, player string) {
	score, err := p.store.GetPlayerScore(player)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "%d", score)
}

func (p *PlayerServer) processWin(w http.ResponseWriter, r *http.Request, player string) {
	p.store.RecordWin(player)

	w.WriteHeader(http.StatusAccepted)
}

type StubPlayerStore struct {
	scores   map[string]int
	winCalls []string

	mu sync.Mutex
}

func (s *StubPlayerStore) GetPlayerScore(name string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	res, ok := s.scores[name]
	if !ok {
		return 0, errors.New("no such user")
	}
	return res, nil
}

func (s *StubPlayerStore) RecordWin(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scores[name]++
	s.winCalls = append(s.winCalls, name)
	return nil
}

func (s *StubPlayerStore) GetLeague() ([]player.Player, error) {
	return []player.Player{
		{Name: "Jan", Wins: 2137},
	}, nil
}
