// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"flag"
	"log"
	"net/http"
    "time"

    "github.com/ginglis13/cli-typeracer/models"
	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8880", "http service address")


var upgrader = websocket.Upgrader{} // use default options

var gameState = models.GameState{0, false, make(map[string]*models.ClientState), []byte("test string."), []byte(""), time.Now() /*TODO*/, 2}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
        var clientState models.ClientState
		err := c.ReadJSON(&clientState)
        if err != nil {
            log.Println("[ERROR] Unable to read JSON from client.")
        }

		log.Printf("[CLIENT STATE]: %s", clientState)

        // Add client to list of clients and set as leader if they are first client
        if len(gameState.Clients) == 0 {
            clientState.IsLeader = true
        }

        gameState.Clients[clientState.UserID] = &clientState

        if clientState.Complete {
            elapsedTime := time.Now().Sub(gameState.StartTime).Seconds()
            log.Printf("****** ELAPSED TIME %v *******\n", elapsedTime)
            log.Printf("****** WORD COUNT %v *******\n", gameState.StrLen)
            log.Printf("****** WPM %.3f *******\n", float64(gameState.StrLen)/elapsedTime*60.0)

            clientState.WPM = float64(gameState.StrLen)/elapsedTime*60.0


            if len(gameState.Winner) == 0 { // Don't reset winner if someone else finishes
                gameState.Winner = []byte(clientState.UserID)
            }

            gameState.Over = true
            // Write game over to all clients in game (but the incoming client so as not to close connection)
            for _, client := range gameState.Clients {
                client.WPM = float64(gameState.StrLen)/elapsedTime*60.0
                if client.UserID == clientState.UserID {
                    continue
                }
                client.Complete = true
                log.Println("GAME OVER SENT TO ", client.UserID)
            }
        }

		log.Printf("[GAME STATE]: %s", gameState)
	    err = c.WriteJSON(&gameState)
        if err != nil {
            log.Println("[ERROR] Unable to write JSON back to client.")
        }

    }
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo", echo)
	log.Fatal(http.ListenAndServe(*addr, nil))
}