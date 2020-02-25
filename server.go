package main

import (
    "fmt"
    "net/http"
	"encoding/json"
	"flag"
	"log"
)

type GameState struct {
	Message string
	Completed bool
	//	clients []*ClientState // take length to verify max of 4 participants
	// also use the progress attribute to check against other players
}

/*
game := make(map[int]GameState) // keep track of games in map TODO allow for multiple games at a time

func joinGame(games map[int]GameState, gameID int){
	exists, gs = games[gameID]
	if exists {
		if len(gs) < 4{
			// ... connect client to game
		}
	}
}
*/

func hello(w http.ResponseWriter, req *http.Request) {

	log.Printf("%s %s /typeracer from %s", req.Method, req.Proto, req.RemoteAddr)

	// read incoming json request
	var gs GameState
	err := json.NewDecoder(req.Body).Decode(&gs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Client Request JSON: %v", gs)
	gs.Completed = true
	log.Printf("Server Response JSON: %v", gs)

	// write back to client
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

    http.HandleFunc("/typeracer", hello)
    http.HandleFunc("/headers", headers)

	p := fmt.Sprintf(":%v", *port)
	log.Printf("Listening on port %v", *port)
    http.ListenAndServe(p, nil)
}
