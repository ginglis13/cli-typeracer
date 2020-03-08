package main

import (
	"fmt"
	"github.com/eiannone/keyboard"
	"github.com/fatih/color"
	"net/http"
	"encoding/json"
	"log"
	"bytes"
	"strings"
//	"flag"
)

type GameState struct {
	//Message string
	//Completed bool
	client *ClientState // take length to verify max of 4 participants
	//clients []*ClientState // take length to verify max of 4 participants
	// also use the progress attribute to check against other players
}

type ClientState struct {
	UserID string
	//gameID int
	Progress int // length of correct input to show comparison to other players
	UserInput string // TODO: check input on client or server side
	Complete bool // indicates client has finished the input
	//isCreate bool // indicates that the user is the game creator - for asking if they want to start another
}

func sendState(c *ClientState) {

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

	resp, err := http.Post("http://localhost:8080/typeracer", "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Fatalln(err)

	}

	var result map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&result)
	log.Println(result)

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

func main() {

	/* Parse Args */
//	gameID  := flag.Int("g", 0, "Join game by game id")
//	nick   := flag.String("n", "", "Set nickname")
//	host   := flag.String("host", "", "Host address/domain of game")
//	port   := flag.Int("p", 443, "Host port")

//	flag.Parse()

	c := ClientState{"ginglis", 0, "", false}

	/* Open Keyboard */
	err := keyboard.Open()
	if err != nil {
		panic(err)
	}
	defer keyboard.Close()

	chars := make([]int32, 0)

	_s := "test string."

	ss := strings.Split(_s, " ")
	fmt.Println(ss)

	s := []byte(_s)

	fmt.Println("Press ESC to quit")
	for {
		fmt.Println("Quote:", _s)

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

		if res := checkInput(chars, s); res && len(chars) == len(s){
			fmt.Println("Game over.")
			c.Complete = true
			//break
		}

		c.UserInput = string(chars)
		sendState(&c)
	}
}
