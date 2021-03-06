package main

import (
    "fmt"
    "net/http"
	"encoding/json"
	"flag"
	"log"
)

type ClientState struct {
	UserID string
	GameID int
	Progress int // length of correct input to show comparison to other players
	UserInput string // TODO: check input on client or server side
	Complete bool // indicates client has finished the input
	ResponseWriter http.ResponseWriter
	//isCreate bool // indicates that the user is the game creator - for asking if they want to start another
}

type GameState struct {
	//Message string
	ID int
	Over bool
	Clients map[string]ClientState // take length to verify max of 4 participants
	String string // the string/paragraph to type
	//clients []*ClientState // take length to verify max of 4 participants
	// also use the progress attribute to check against other players
}

var GAMES = make(map[int]GameState) // keep track of games in map TODO allow for multiple games at a time

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

func typeracer(w http.ResponseWriter, req *http.Request) {

	log.Printf("%s %s /typeracer from %s", req.Method, req.Proto, req.RemoteAddr)

	// read incoming json request
	var cs ClientState
	err := json.NewDecoder(req.Body).Decode(&cs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the specified gameID is in map of Games
	var gs GameState
	if gs, isset := GAMES[cs.GameID]; isset == true {
		// If Game is complete, send end state to all clients in the Game
		if cs.Complete {
			gs.Over = true
			for _, client := range gs.Clients {
				json.NewEncoder(client.ResponseWriter).Encode(gs)
			}
		}
		cs.ResponseWriter = w
		gs.Clients[cs.UserID] = cs
		fmt.Println("CLIENTS:" , gs.Clients)
	} else { // new game
		cs.ResponseWriter = w
		s := "test string."
		gs := GameState{cs.GameID, false, make(map[string]ClientState), s}
		gs.Clients[cs.UserID] = cs
		GAMES[cs.GameID] = gs
	}


	log.Printf("Client Request JSON: %v", cs)
	if cs.Complete == true {
		gs.Over = true
	}
	if cs.GameID == -1 {
		gs.ID = 10 // in place before multi game logic
	}

	log.Printf("Server Response JSON: %v", gs)

	//gameEndPt := fmt.Sprintf("/typeracer/%v", gs.ID)
    //http.HandleFunc(gameEndPt, game)

	// write back to client
	fmt.Println(gs)
	json.NewEncoder(w).Encode(gs)
}

func headers(w http.ResponseWriter, req *http.Request) {
    for name, headers := range req.Header {
        for _, h := range headers {
            fmt.Fprintf(w, "%v: %v\n", name, h)
        }
    }
}

func main() {

	/* Parse Args */
	port := flag.Int("p", 8080, "Host port")

	flag.Parse()

    http.HandleFunc("/typeracer", typeracer)
    http.HandleFunc("/headers", headers)

	p := fmt.Sprintf(":%v", *port)
	log.Printf("Listening on port %v", *port)
    http.ListenAndServe(p, nil)
}
