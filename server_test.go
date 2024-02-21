package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	psqlstore "poker/psqlStore"
	"slices"
	"sync"
	"testing"
)

func TestGETPlayers(t *testing.T) {
	stubStore := &StubPlayerStore{scores: map[string]int{
		"Alice": 20,
		"Bob":   10,
	}}
	server := &PlayerServer{store: stubStore}

	t.Run("returning score of player A", func(t *testing.T) {
		request := newGetScoreRequest("Alice")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := response.Body.String()
		want := "20"

		assertScore(t, got, want)
	})

	t.Run("returning score of player B", func(t *testing.T) {
		request := newGetScoreRequest("Bob")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := response.Body.String()
		want := "10"

		assertScore(t, got, want)
	})

	t.Run("return 404 for non-existing player", func(t *testing.T) {
		request := newGetScoreRequest("Oscar")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := response.Code
		want := http.StatusNotFound

		assertCode(t, got, want)
	})
}

func TestStoreWins(t *testing.T) {
	t.Run("single set winner test", func(t *testing.T) {
		store := StubPlayerStore{scores: map[string]int{}}
		server := &PlayerServer{store: &store}
		winnerName := "Alice"
		request := newPostScoreRequest(winnerName)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := response.Code
		want := http.StatusAccepted

		assertCode(t, got, want)
		lenght := 1
		assertWinnersLen(t, &store, lenght)

		winnersArray := []string{winnerName}
		assertWinnersContent(t, &store, winnersArray)
	})

	t.Run("test concurrent writes", func(t *testing.T) {
		threadNr := 1000

		store := StubPlayerStore{scores: map[string]int{}}
		server := &PlayerServer{store: &store}
		var wg sync.WaitGroup
		for i := range threadNr {
			wg.Add(1)
			go func(i int) {
				winnerName := fmt.Sprintf("player%d", i)
				request := newPostScoreRequest(winnerName)
				response := httptest.NewRecorder()
				server.ServeHTTP(response, request)
				wg.Done()
			}(i)
		}
		wg.Wait()

		lenght := threadNr
		assertWinnersLen(t, &store, lenght)
	})
}

// TODO: add crearing database after tests.
// Now I need to remove data after each test manually
func TestPostgreSQLStore(t *testing.T) {
	t.Run("integration test for postgres database", func(t *testing.T) {
		threadNr := 1000

		store, err := psqlstore.New()
		if err != nil {
			t.Fatalf("cant create conn to db: %s", err)
		}
		server := &PlayerServer{store: store}
		userName := "Pudzian"

		var wg sync.WaitGroup
		for range threadNr {
			wg.Add(1)
			go func() {
				request := newPostScoreRequest(userName)
				response := httptest.NewRecorder()
				server.ServeHTTP(response, request)
				wg.Done()
			}()
		}
		wg.Wait()

		assertUserWins(t, store, userName, threadNr)
	})
}

func newGetScoreRequest(name string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/players/%s", name), nil)
	return req
}

func newPostScoreRequest(name string) *http.Request {
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/players/%s", name), nil)
	return req
}

func assertScore(t testing.TB, got, want string) {
	t.Helper()
	if want != got {
		t.Errorf("exptected %s but got %s", want, got)
	}
}

func assertCode(t testing.TB, got, want int) {
	t.Helper()
	if want != got {
		t.Errorf("exprected code %d but got code %d", want, got)
	}
}

func assertWinnersLen(t testing.TB, store *StubPlayerStore, want int) {
	t.Helper()
	got := len(store.winCalls)
	if got != want {
		t.Errorf("exprected %d winners but got %d winners", want, got)
	}
}

func assertWinnersContent(t testing.TB, store *StubPlayerStore, want []string) {
	t.Helper()
	got := store.winCalls
	if !slices.Equal(want, got) {
		t.Errorf("exprected %v winners but got %v winners", want, got)
	}
}

func assertResponseBody(t testing.TB, got, want string) {
	t.Helper()
	if want != got {
		t.Errorf("exprected %s but got %s", want, got)
	}
}

func assertUserWins(t testing.TB, store PlayerStore, user string, want int) {
	t.Helper()
	got, _ := store.GetPlayerScore(user)
	if want != got {
		t.Errorf("exprected %d but got %d", want, got)
	}
}
