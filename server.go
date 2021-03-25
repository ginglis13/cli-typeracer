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
    "math/rand"
    "fmt"
    "io/ioutil"
    "sync"
    "strings"

    "github.com/ginglis13/cli-typeracer/models"
	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8880", "http service address")

var upgrader = websocket.Upgrader{} // use default options

var mutex = &sync.Mutex{}

var gameState = models.GameState{0, false, make(map[string]*models.ClientState), /*[]byte("test string.")*/getQuote(), []byte(""), time.Now() /*TODO*/, 2, false}

func getQuote() []byte {
	quoteNum := rand.Intn(10)
	quotePath := fmt.Sprintf("quotes/%v.txt", quoteNum)
	data, err := ioutil.ReadFile(quotePath)
	if err != nil {
		panic(err)
	}

	return data
}

func startGame(w http.ResponseWriter, r *http.Request) {
    log.Println("IN START GAME")
    gameState.StrLen = len(strings.Fields(string(gameState.String)))

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer conn.Close()
    var clientState models.ClientState
	for {
		err := conn.ReadJSON(&clientState)
        if err != nil {
            log.Println("[ERROR] Unable to read JSON from client for START GAME: %v", clientState)
            break
        }
        // set leader on clientstate
        if len(gameState.Clients) == 0 {
            clientState.IsLeader = true
        }


        mutex.Lock()
        gameState.Clients[clientState.UserID] = &clientState
        mutex.Unlock()

        if clientState.IsLeader && clientState.StartGame {
            gameState.Started = true
	        err = conn.WriteJSON(&gameState)
            gameState.StartTime = time.Now()
            break
        }
	    err = conn.WriteJSON(&gameState)
        if err != nil {
            log.Println("[ERROR] Unable to write JSON back to client for START GAME %v", clientState)
            break
        }
    }
}

func typeracer(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

    var clientState models.ClientState

	for {
		err := c.ReadJSON(&clientState)
        if err != nil {
            log.Println("[ERROR] Unable to read JSON from client.")
            break
        }

		log.Printf("[CLIENT STATE]: %s", clientState)

        // Add client to list of clients and set as leader if they are first client
        if len(gameState.Clients) == 0 {
            clientState.IsLeader = true
        }

        mutex.Lock()
        gameState.Clients[clientState.UserID] = &clientState
        mutex.Unlock()

        if clientState.Complete || gameState.Over {
            elapsedTime := time.Now().Sub(gameState.StartTime).Seconds()
            log.Printf("****** ELAPSED TIME %v *******\n", elapsedTime)
            log.Printf("****** WORD COUNT %v *******\n", gameState.StrLen)
            log.Printf("****** WPM %.3f *******\n", float64(gameState.StrLen)/elapsedTime*60.0)

            clientState.WPM = float64(gameState.StrLen)/elapsedTime*60.0

            if len(gameState.Winner) == 0 { // Don't reset winner if someone else finishes
                gameState.Winner = []byte(clientState.UserID)
            }

            gameState.Over = true
            // Write game over to all clients in game 
            mutex.Lock()
            for _, client := range gameState.Clients {
                strLen := len(strings.Fields(string(client.UserInput)))
                client.WPM = float64(strLen)/elapsedTime*60.0
                client.Complete = true
                log.Println("GAME OVER SENT TO ", client.UserID)
            }
            mutex.Unlock()
        }

        // TODO: mutex is necessary for reading gamestate here - do I need this logging ?
        mutex.Lock()
		log.Printf("[GAME STATE]: %s", gameState)
        mutex.Unlock()

        mutex.Lock()
        err = c.WriteJSON(&gameState)
        log.Printf("[NOT OVER GAME STATE]: %s", gameState)
        if err != nil {
            log.Println("[ERROR] Unable to write JSON back to client.")
            break
        }
        mutex.Unlock()

        if gameState.Over {
            time.Sleep(time.Millisecond * 100) // delay to reset server, must be more than the delay for sending message from client
            gameState  = models.GameState{0, false, make(map[string]*models.ClientState), getQuote(), []byte(""), time.Now() /*TODO*/, 0, false}
        }

        // Receive closed websocket message to gracefully end connection.
        _, message, err := c.ReadMessage()
		if err != nil {
			log.Println("recv websock close:", err)
            break
		}
		log.Printf("recv: %s", message)
    }
}

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/typeracer", typeracer)
	http.HandleFunc("/startgame", startGame)
	log.Fatal(http.ListenAndServe(*addr, nil))
}