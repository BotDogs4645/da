// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

// Go version 1.20 or newer is required due to how it initializes the PRNG.
//go:build go1.20

package main

import (
	"log"

	"github.com/BotDogs4645/da/field"
	"github.com/BotDogs4645/da/web"
)

const eventDbPath = "./event.db"
const httpPort = 8080

// Main entry point for the application.
func main() {
	arena, err := field.NewArena(eventDbPath)
	if err != nil {
		log.Fatalln("Error during startup: ", err)
	}

	// Start the web server in a separate goroutine.
	web := web.NewWeb(arena)
	go web.ServeWebInterface(httpPort)

	// Run the arena state machine in the main thread.
	arena.Run()
}
