package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type PlayerStore interface {
	GetPlayerScore(name string) (int, error)
	RecordWin(name string) error
}

type PlayerServer struct {
	store PlayerStore
}

func (p *PlayerServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
