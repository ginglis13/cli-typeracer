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

var addr = flag.String("addr", "localhost:8888", "http service address")

type ClientCliArgs struct {
	nick string
	host string
	port int
	gameID int
}

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

func checkInput(chars []int32, s []byte) bool {
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
	} else {
		color.Red(cs)
	}
	return res
}

func keyboardInput(clientState *models.ClientState, inp chan struct{}) {
	defer close(inp) 
	chars := make([]int32, 0)
	for {
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
	//done := make(chan struct{})
	//defer close(done)
	for {
		var gs models.GameState
		err := conn.ReadJSON(&gs)
		if err != nil {
			log.Printf("Error reading receiver json")
		}
		s := gs.String
		log.Printf("recv: %s", gs)
		res := checkInput([]int32(clientState.UserInput), s)
		if res && len(clientState.UserInput) == len(s) {
			clientState.Complete = true
		}

		// Check if server registered them as leader
		if gs.Clients[clientState.UserID].IsLeader {
			clientState.IsLeader = true
		}

		log.Printf("GAME OVER STATUS: %v", gs.Over)

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

func main() {
	/* Parse Args */
	args := parseArgs()

	/* Initial Client State */
	clientState := models.ClientState{args.nick, args.gameID, 0, "", false, false, 0}

	log.Println("Sending state", clientState)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	host := fmt.Sprintf("%s:%v", args.host, args.port)

	u := url.URL{Scheme: "ws", Host: host, Path: "/echo"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	log.Printf("TYPE OF c: %T", c)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	//done := make(chan struct{})
	done := make(chan bool, 1)
	defer close(done)

	/* For receiving game state data from server */
	go recvGameState(c, &clientState, done)

	/* Open Keyboard and receive input in background goroutine */
	keyErr := keyboard.Open()
	if keyErr != nil {
		panic(keyErr)
	}
	defer keyboard.Close()

	inp := make(chan struct{})
	go keyboardInput(&clientState, inp)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()


	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
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
		/* TODO : UNCOMMENT FOR REAL TIME w/o 1s delay
		default:
			err := c.WriteJSON(&clientState)
			if err != nil {
				log.Println("write:", err)
				log.Println("t:", t)
				return
			}
			log.Println(&clientState)
		*/
		}
	}
}
