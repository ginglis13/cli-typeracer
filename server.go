package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	//"math/rand"
	"time"
)

// type ClientState struct {
// 	UserID string
// 	GameID int
// 	Progress int // length of correct input to show comparison to other players
// 	UserInput string // TODO: check input on client or server side
// 	Complete bool // indicates client has finished the input
// 	//ResponseWriter http.ResponseWriter
// 	//isCreate bool // indicates that the user is the game creator - for asking if they want to start another
// }
type ClientState struct {
	UserID         string `json:"UserID"`
	GameID         int    `json:"GameID"`
	Progress       int    `json:"Progress"` // length of correct input to show comparison to other players
	UserInput      string `json:"UserInput"`
	Complete       bool   `json:"Complete"` // indicates client has finished the input
	// Used for beginning the game - leader will indicate all players have joined, will be asked to re run game
	IsLeader	   bool   `json:"IsLeader"` 
	WPM 		   float64    `json:"WPM"`
	ResponseWriter http.ResponseWriter
}

type GameState struct {
	ID      int `json:"ID"`
	Over    bool `json:"Over"`
	Clients map[string]*ClientState `json:"Clients"` // take length to verify max of 4 participants
	String  []byte `json:"String"`                // the string/paragraph to type
	Winner []byte  `json:"Winner"`
	StartTime	time.Time `json:"StartTime"`
	StrLen int `json:"StrLen"`
	// also use the progress attribute to check against other players
}

type GameMap map[int]*GameState

var GAMES = make(GameMap) // keep track of games in map TODO allow for multiple games at a time

/*
func joinGame(games map[int]GameState, gameID int){
	exists, gs = games[gameID]
	if exists {
		if len(gs) < 4{
			// ... connect client to game
		}
	}
}
*/
func game(w http.ResponseWriter, req *http.Request, gs *GameState) {
	// need to write an interface for HandleFunc to pass vars

}

/* UNCOMMENT FOR REAL DEAL
func getQuote() []byte {
	quoteNum := rand.Intn(10)
	quotePath := fmt.Sprintf("quotes/%v.txt", quoteNum)
	data, err := ioutil.ReadFile(quotePath)
	if err != nil {
		panic(err)
	}

	return data
}
*/

func continueMenu() {

}

func typeracer(w http.ResponseWriter, req *http.Request) {
	_s := "test string."
	s := []byte(_s)
	quoteLen := len(strings.Fields(_s)) // Number of Words in Quote, used for calcs

	log.Printf("%s %s /typeracer from %s", req.Method, req.Proto, req.RemoteAddr)

	// read incoming json request
	//var cs ClientState
	// err := json.NewDecoder(req.Body).Decode(&cs)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return
	// }

	body, readErr := ioutil.ReadAll(req.Body)
	if readErr != nil {
		log.Fatalln(readErr)
	}

	cs := ClientState{}

	jErr := json.Unmarshal(body, &cs)
	if jErr != nil {
		log.Fatalln(jErr)
	}

	cs.ResponseWriter = w
	fmt.Println(cs)

	// Check if the specified gameID is in map of Games
	var gs *GameState
	var isset bool
	if gs, isset = GAMES[cs.GameID]; isset == true {
		// If Game is complete, send end state to all clients in the Game
		if cs.Complete {

			elapsedTime := time.Now().Sub(GAMES[cs.GameID].StartTime).Seconds()
			fmt.Printf("****** ELAPSED TIME %v *******\n", elapsedTime)
			fmt.Printf("****** WORD COUNT %v *******\n", GAMES[cs.GameID].StrLen)
			fmt.Printf("****** WPM %.3f *******\n", float64(GAMES[cs.GameID].StrLen)/elapsedTime*60.0)


			if len(GAMES[cs.GameID].Winner) == 0 { // Don't reset winner if someone else finishes
				GAMES[cs.GameID].Winner = []byte(cs.UserID)
			}

			GAMES[cs.GameID].Over = true
			gs.Over = true
			// Write game over to all clients in game (but the incoming client so as not to close connection)
			for _, client := range GAMES[cs.GameID].Clients {
				client.WPM = float64(GAMES[cs.GameID].StrLen)/elapsedTime*60.0
				if client.UserID == cs.UserID {
					continue
				}
				client.Complete = true
				json.NewEncoder(client.ResponseWriter).Encode(GAMES[cs.GameID])
				fmt.Println("GAME OVER SENT TO ", client.UserID)

			}
			// Write to this incoming clientk
			json.NewEncoder(w).Encode(GAMES[cs.GameID])
			//delete(GAMES, cs.GameID)
			return
		}
		//cs.ResponseWriter = w
		gs.Clients[cs.UserID] = &cs
		fmt.Println("CLIENTS:", gs.Clients)
	} else { // new game, TODO search for an open game (10 max)
		cs.ResponseWriter = w
		cs.GameID = 10
		gs = &GameState{cs.GameID, false, make(map[string]*ClientState),s /*getQuote() TODO UNCOMMENT AFTER TESTING*/, nil, time.Now(), quoteLen} 
		fmt.Println("gs", gs)
		gs.Clients[cs.UserID] = &cs
		GAMES[cs.GameID] = gs
		fmt.Println("CLIENTS:", gs.Clients)
	}

	log.Printf("Client Request JSON: %v", cs)

	log.Printf("Server Response JSON: %v", gs)
	log.Printf("GAME: %v", *GAMES[cs.GameID])

	// write back to client
	json.NewEncoder(w).Encode(*GAMES[cs.GameID])
	
}

func headers(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func main() {
	// TODO UNCOMMENT AFTER TESTING
	//rand.Seed(time.Now().UnixNano())

	/* Parse Args */
	port := flag.Int("p", 8080, "Host port")
	/*
	   f, _ := os.Create("/var/log/golang/golang-server.log")
	   defer f.Close()
	   log.SetOutput(f)
	*/

	flag.Parse()

	http.HandleFunc("/typeracer", typeracer)
	http.HandleFunc("/headers", headers)

	p := fmt.Sprintf(":%v", *port)
	log.Printf("Listening on port %v", *port)
	http.ListenAndServe(p, nil)
}
