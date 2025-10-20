package main

import (
	"log"
	"net/http"

	"game-sandbox/server/game"
	"game-sandbox/server/handlers"
)

func main() {
	// Create game hub
	hub := game.NewHub()

	// Setup routes
	http.HandleFunc("/ws", handlers.WebSocketHandler(hub))

	// Serve static files
	fs := http.FileServer(http.Dir("../client/dist"))
	http.Handle("/", fs)

	// Start server
	log.Println("Organism game server running on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
