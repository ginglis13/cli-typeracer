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
	//"strings"
	"flag"
)

type GameState struct {
	//Message string
	ID int
	Over bool
	Clients map[string]ClientState // take length to verify max of 4 participants
	// also use the progress attribute to check against other players
}

type ClientState struct {
	UserID string
	GameID int
	Progress int // length of correct input to show comparison to other players
	UserInput string
	Complete bool // indicates client has finished the input
	//isCreate bool // indicates that the user is the game creator - for asking if they want to start another
}

func sendState(c *ClientState, host string, port int) {

	message := map[string]interface{}{
		"userID": c.UserID,
		"userInput": c.UserInput,
		"complete": c.Complete,
	}

	fmt.Println(c)

	bytesRepresentation, err := json.Marshal(message)
	if err != nil {
		log.Fatalln(err)
	}

	server := fmt.Sprintf("http://%s:%v/typeracer/%v", host, port, c.GameID)
	fmt.Println(c.GameID)
	fmt.Println(server)
	resp, err := http.Post(server, "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Fatalln(err)

	}

	var result map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&result)

}


func delInput(chars []int32) []int32{
	if len(chars) > 0 {
		return append(chars[:len(chars)-1])
	} else {
		return chars
	}
}

/* TODO: space is 0 in chars, but ascii 32 in s */
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
		/* found a match */
		} else if v == int32(s[i]) {
			res = true
		/* space check */
		} else if v == 0 && int32(s[i]) == 32{
			res = true
		} else {
			res = false
		}
	}

	if res {
		color.Green(cs)
	} else {
		color.Red(cs)
	}
	return res

}

func beginGame(c *ClientState) int {
	beginGame := map[string]interface{}{
		"userID": c.UserID,
		"gameID": c.GameID,
	}

	bytesRepresentation, err := json.Marshal(beginGame)
	if err != nil {
		log.Fatalln(err)
	}
	resp, err := http.Post("http://localhost:8080/typeracer", "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Fatalln(err)
	}

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatalln(readErr)
	}

	gs := GameState{c.GameID, false, make(map[string]ClientState)}

	jErr := json.Unmarshal(body, &gs)
	if jErr != nil {
		log.Fatalln(jErr)
	}

	fmt.Println(gs)

	return gs.ID
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

	c := ClientState{nick, gameID, 0, "", false}

	/* return new gameID if -1 specified */
	c.GameID = beginGame(&c)

	/* Open Keyboard */
	err := keyboard.Open()
	if err != nil {
		panic(err)
	}
	defer keyboard.Close()

	chars := make([]int32, 0)

	_s := "test string."

	s := []byte(_s)

	for {
		/* clear screen, print prompt */
		//fmt.Print("\033[H\033[2J")
		fmt.Println("Quote:", _s)
		res := checkInput(chars, s)

		char, key, err := keyboard.GetKey()
		if (err != nil) {
			panic(err)
		} else if (key == keyboard.KeyEsc) {
			break
		/* check for both KeyBackspace and KeyBackspace2 for delete */
		} else if (key == keyboard.KeyBackspace2 || key == keyboard.KeyBackspace || key == keyboard.KeyDelete) {
			chars = delInput(chars)
		} else {
			chars = append(chars, char)
		}

		if res && len(chars) == len(s){
			fmt.Println("Game over.")
			c.Complete = true
			//break
		}

		c.UserInput = string(chars)
		sendState(&c, host, port)
	}
}
