package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"pursuit-go/internal/server"
	"syscall"
)

func main() {
	fmt.Println("Starting Pursuit Go engine...")

	srv := server.New(":8000", 5)

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		fmt.Println("Shutting down server...")
		srv.Stop()
	}()

	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}

}
