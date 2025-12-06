package main

import (
	"net/http"

	"game-sandbox/server/game"
	"game-sandbox/server/handlers"
	"game-sandbox/server/utils"
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
	utils.LogServerStart(":8081")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		utils.Fatal("Server failed to start", map[string]interface{}{
			"error": err.Error(),
			"port":  ":8081",
		})
	}
}
