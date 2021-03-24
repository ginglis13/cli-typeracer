// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
	"fmt"
	"strings"

	"github.com/eiannone/keyboard"
	"github.com/fatih/color"
    "github.com/ginglis13/cli-typeracer/models"
	"github.com/gorilla/websocket"
)

type ClientCliArgs struct {
	nick string
	host string
	port int
	gameID int
}

var addr = flag.String("addr", "localhost:8888", "http service address")

func parseArgs() ClientCliArgs {
	args := ClientCliArgs{}

	flag.IntVar(&args.gameID, "g", 0, "Join game by game id")
	flag.StringVar(&args.nick, "n", "", "Set nickname")
	flag.StringVar(&args.host, "host", "localhost", "Host address/domain of game")
	flag.IntVar(&args.port, "p", 8880, "Host port")
	flag.Parse()
	log.SetFlags(0)

	fmt.Println("CLI Typeracer")
	fmt.Println("Press ESC to quit")

	fmt.Printf("Host is set to %v:%v\n", args.host, args.port)

	if args.gameID == 0 {
		fmt.Println("Enter the ID of the game you'd like to join, or enter -1 to start a new game:")
		fmt.Scanf("%d", &args.gameID)
	}

	if args.nick == "" {
		fmt.Println("Enter a nickname to join the game with:")
		fmt.Scanf("%s", &args.nick)
	}

	return args
}

func delInput(chars []int32) []int32{
	if len(chars) > 0 {
		return append(chars[:len(chars)-1])
	} else {
		return chars
	}
}

func checkInput(chars []int32, s []byte, clientState *models.ClientState) bool {
	var cs string
	for _, v := range chars{
		/* space check */
		if (v == 0){
			v = ' '
		}
		cs = fmt.Sprintf("%s%c", cs, v)
	}
	res := false
	for i, v := range chars {
		/* check str len */
		if i > len(s) - 1 {
			res = false
			break
		/* found a match */
		} else if v == int32(s[i]) {
			res = true
		/* space check */
		} else if v == 0 && int32(s[i]) == 32{
			res = true
		} else {
			res = false
			break
		}
	}

	if res {
		color.Green(cs)
		clientState.Progress = len(chars)
	} else {
		color.Red(cs)
	}
	return res
}

func keyboardInput(clientState *models.ClientState, inp chan bool) {
	chars := make([]int32, 0)
	/* Open Keyboard and receive input in background goroutine */
	keyErr := keyboard.Open()
	if keyErr != nil {
		panic(keyErr)
	}
	defer keyboard.Close()
	for {
		if clientState.Complete {
			inp <- true
			break
		}
		char, key, err := keyboard.GetKey()
		if err != nil {
			panic(err)
		} else if key == keyboard.KeyEsc {
			break
		/* check for both KeyBackspace and KeyBackspace2 for delete */
		} else if key == keyboard.KeyBackspace2 || key == keyboard.KeyBackspace || key == keyboard.KeyDelete {
			chars = delInput(chars)
		} else {
			chars = append(chars, char)
		}
		clientState.UserInput = string(chars)
	}
}

func recvGameState(conn *websocket.Conn, clientState *models.ClientState, done chan bool) {
	for {
		var gs models.GameState
		err := conn.ReadJSON(&gs)
		if err != nil {
			log.Printf("Error reading receiver json")
		}
		printProgress(&gs)
		s := gs.String
		log.Printf("recv: %s", gs)
		res := checkInput([]int32(clientState.UserInput), s, clientState)
		if res && len(clientState.UserInput) == len(s) {
			clientState.Complete = true
		}

		// Check if server registered them as leader
		if gs.Clients[clientState.UserID].IsLeader {
			clientState.IsLeader = true
		}

		if gs.Over {
			gameOver(&gs)
			if err != nil {
				log.Println("write close:", err)
				return
			}
			done<-true
			break
		}
	}

}

func gameOver(gs *models.GameState) {
	fmt.Print("\033[H\033[2J")
	fmt.Println(strings.Repeat("#", 8), "GAME OVER", strings.Repeat("#", 8))
	fmt.Printf("Winner:\t")
	color.Green(string(*&gs.Winner))
	fmt.Println("Player WPM: ")
	finalGameState := *&gs.Clients
	// Print WPM
	for client, state := range finalGameState {
		if client == "" {
			continue
		}
		fmt.Printf("%10s: %.3v WPM\n", client, state.WPM)
	}


	/* TODO: allow player to play again with same players
	if cs.IsLeader {
		fmt.Println("You are the game leader.")
		fmt.Println("Would you like to play again with the same players? [y/N]")
		var s string
		fmt.Scanf("%s", s)
		fmt.Println(s)
	}
	*/

}

func printProgress(gs *models.GameState) {
	chars := len(gs.String)
	for client, state := range *&gs.Clients {
		if client == "" {
			continue
		}
		// Do percentage based on 100
		percentDone := float64(state.Progress) * 100.0 / float64(chars)
		fmt.Printf("%10s: [%s%s]\n", client, strings.Repeat("#", int(percentDone)), strings.Repeat(" ", 100-int(percentDone)))
	}
}


func startGame(conn *websocket.Conn, clientState *models.ClientState) { // Send your initial state to the game server
	err := conn.WriteJSON(&clientState)
	if err != nil {
		log.Println("write:", err)
		return
	}

	// Receive if you are leader
	var gs models.GameState
	err = conn.ReadJSON(&gs)
	if err != nil {
		log.Printf("[ERROR] Error reading receiver json for start game")
	}
	log.Println("Initial Game State", gs)
	log.Println("Initial Client State", gs.Clients[clientState.UserID])

	clientState.IsLeader = gs.Clients[clientState.UserID].IsLeader

	// If leader, wait until all players joined, respond if so, then exit
	if gs.Clients[clientState.UserID].IsLeader {
		//var result string
		fmt.Println("Press enter once everyone has joined.")
		keyErr := keyboard.Open()
		if keyErr != nil {
			panic(keyErr)
		}
		defer keyboard.Close()

		keyboard.GetKey()


		clientState.StartGame = true
		// Write that the game has been started back to server
		err = conn.WriteJSON(&clientState)
		if err != nil {
			log.Printf("[ERROR] Error sending json for start game")
		}


	// If not leader, wait until gameState.Started, then exit
	} else {
		fmt.Print("\033[H\033[2J") // clear screen
		fmt.Println("Waiting for the game leader to start..")
		for {
			conn.WriteJSON(&clientState)
			conn.ReadJSON(&gs) 
			if gs.Started {
				break
			}
		}
	}


}

func main() {
	/* Parse Args */
	args := parseArgs()

	/* Initial Client State */
	clientState := models.ClientState{args.nick, args.gameID, 0, "", false, false, 0, false}

	log.Println("Sending state", clientState)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	host := fmt.Sprintf("%s:%v", args.host, args.port)



	startU := url.URL{Scheme: "ws", Host: host, Path: "/startgame"}
	conn, _, err := websocket.DefaultDialer.Dial(startU.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	startGame(conn, &clientState)
	conn.Close()

	u := url.URL{Scheme: "ws", Host: host, Path: "/typeracer"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	fmt.Println("Game starting in 3...")
	time.Sleep(1 * time.Second)
	fmt.Println("Game starting in 2...")
	time.Sleep(1 * time.Second)
	fmt.Println("Game starting in 1...")
	time.Sleep(1 * time.Second)

	done := make(chan bool, 1)
	defer close(done)

	/* For receiving game state data from server */
	go recvGameState(c, &clientState, done)


	inp := make(chan bool, 1)
	go keyboardInput(&clientState, inp)

	ticker := time.NewTicker(time.Second)  // TODO: time.Millisecond for feal time
	defer ticker.Stop()


	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			log.Printf("[CLIENT COMPLETE]: %v\n", clientState.Complete)
			err := c.WriteJSON(&clientState)
			if err != nil {
				log.Println("write:", err)
				log.Println("t:", t)
				return
			}
			log.Println(&clientState)
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
