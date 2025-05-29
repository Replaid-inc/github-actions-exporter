package main

import (
	"os"

	"gh-actions-exporter/internal/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080" // Default port if not specified
	}

	webhookSecret := os.Getenv("GITHUB_WEBHOOK_SECRET")

	server.StartServer(port, webhookSecret)
}
