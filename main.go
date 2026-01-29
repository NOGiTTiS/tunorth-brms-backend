package main

import (
	"log"
	"os"
	"tunorth-brms-backend/internal/bootstrap"
)

func main() {
	app := bootstrap.CreateApp()

	// Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(app.Listen(":" + port))
}
