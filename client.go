package main

import (
	"fmt"
	"github.com/eiannone/keyboard"
	"github.com/fatih/color"
	"net/http"
	"encoding/json"
	"log"
	"bytes"
	"io/ioutil"
	"strings"
	"flag"
)

type GameState struct {
	ID int `json:"ID"`
	Over bool `json:"Over"`
	Clients map[string]ClientState `json:"Clients"` // take length to verify max of 4 participants
	String []byte `json:"String"` // the string/paragraph to type
	Winner []byte `json:"Winner"`
	StrLen int `json:"StrLen"`
	// also use the progress attribute to check against other players
}

type ClientState struct {
	UserID string `json:"UserID"`
	GameID int  `json:"GameID"`
	Progress int `json:"Progress"` // length of correct input to show comparison to other players
	UserInput string `json:"UserInput"`
	Complete bool `json:"Complete"` // indicates client has finished the input
	IsLeader bool `json:"IsLeader"`
	WPM float64    `json:"WPM"`
}

// Display Game Over Message, Winner, WPM, menu options to start over
func gameOver(gs *GameState) {
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

}

func sendState(c *ClientState, host string, port int) *GameState {

	bytesRepresentation, err := json.Marshal(&c)
	if err != nil {
		log.Fatalln(err)
	}

	server := fmt.Sprintf("http://%s:%v/typeracer", host, port)
	resp, err := http.Post(server, "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Fatalln(err)

	}

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatalln(readErr)
	}

	gs := GameState{c.GameID, false, make(map[string]ClientState), []byte(""), []byte(""), 0}

	jErr := json.Unmarshal(body, &gs)
	if jErr != nil {
		log.Fatalln(jErr)
	}

	fmt.Println("GS: ", gs)

	// Set Game ID From Server Response
	*&c.GameID = gs.ID

	return &gs

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

func beginGame(c *ClientState, host string, port int) *GameState {
	beginGame := map[string]interface{}{
		"userID": c.UserID,
		"gameID": c.GameID,
	}

	fmt.Println("BEGINNING GAME WITH ID ", c.GameID)

	bytesRepresentation, err := json.Marshal(beginGame)
	if err != nil {
		log.Fatalln(err)
	}
	server := fmt.Sprintf("http://%s:%v/typeracer", host, port)
	resp, err := http.Post(server, "application/json", bytes.NewBuffer(bytesRepresentation))
	fmt.Println("RESPONSE:", resp)
	if err != nil {
		log.Fatalln(err)
	}

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatalln(readErr)
	}

	gs := GameState{c.GameID, false, make(map[string]ClientState), []byte(""), []byte(""), 0}

	jErr := json.Unmarshal(body, &gs)
	if jErr != nil {
		log.Fatalln(jErr)
	}

	fmt.Println(gs)

	// Set Game ID From Server Response
	*&c.GameID = gs.ID

	return &gs
}


func main() {

	/* Parse Args */
	var nick, host string
	var port, gameID int

	flag.IntVar(&gameID, "g", 0, "Join game by game id")
	flag.StringVar(&nick, "n", "", "Set nickname")
	flag.StringVar(&host, "host", "localhost", "Host address/domain of game")
	flag.IntVar(&port, "p", 8080, "Host port")
	flag.Parse()

	fmt.Println("CLI Typeracer")
	fmt.Println("Press ESC to quit")

	fmt.Printf("Host is set to %v:%v\n", host, port)

	if gameID == 0 {
		fmt.Println("Enter the ID of the game you'd like to join, or enter -1 to start a new game:")
		fmt.Scanf("%d", &gameID)
	}

	if nick == "" {
		fmt.Println("Enter a nickname to join the game with:")
		fmt.Scanf("%s", &nick)
	}

	c := ClientState{nick, gameID, 0, "", false, false, 0}

	fmt.Println("Sending state", c)

	/* return new gameID if -1 specified */
	gs := beginGame(&c, host, port)
	c.GameID = *&gs.ID
	s := *&gs.String

	/* Open Keyboard */
	err := keyboard.Open()
	if err != nil {
		panic(err)
	}
	defer keyboard.Close()

	chars := make([]int32, 0)

	//_s := "test string."

	//s := []byte(_s)

	for {
		/* clear screen, print prompt */
		//fmt.Print("\033[H\033[2J")
		fmt.Println("Quote:", string(s))
		res := checkInput(chars, s)

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

		if res && len(chars) == len(s){
			fmt.Println("Game over.")
			c.Complete = true
			//sendState(&c, host, port)
			//break
		}

		c.UserInput = string(chars)
		fmt.Println("SENDING TO GAME",c.GameID)
		game := sendState(&c, host, port)
		//c.GameID = game.ID.(int)
		c.GameID = *&game.ID
		fmt.Println("GAME:", c.GameID)
		fmt.Println("GAME:", *&game.Over)
		fmt.Println("GAME:", *&game)
		if *&game.Over == true {
			fmt.Println("GAME OVER")
			gameOver(game)
			break
		}
	}
}
