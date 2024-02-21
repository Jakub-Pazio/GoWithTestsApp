package main

import (
	"log"
	"net/http"
	psqlstore "poker/psqlStore"
)

func main() {
	store, err := psqlstore.New()
	if err != nil {
		log.Fatalf("cant connect to db: %s", err.Error())
	}

	server := NewPlayerServer(store)
	log.Fatal(http.ListenAndServe(":5050", server))
}
