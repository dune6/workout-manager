package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
	"workout-manager/internal/database"

	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port int
}

func NewServer(db *database.Database) *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	NewServer := &Server{
		port: port,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(db),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
