package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/codecrafters-redis-go/internal/server"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Create and start the server
	srv := server.New("0.0.0.0:6379")

	if err := srv.Start(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down server...")
		srv.Stop()
	}()

	// Wait for server to shut down
	srv.Wait()
}
