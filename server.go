jackage main

import (
    "fmt"
    "net/http"
)

type GameState struct {
	clients []*ClientState // take length to verify max of 4 participants
	// also use the progress attribute to check against other players
}

game := make(map[int]GameState) // keep track of games in map TODO allow for multiple games at a time

func joinGame(games map[int]GameState, gameID int){
	exists, gs = games[gameID]
	if exists {
		if len(gs) < 4{
			// ... connect client to game
		}
	}
}

func hello(w http.ResponseWriter, req *http.Request) {
    fmt.Fprintf(w, "hello\n")
}

func headers(w http.ResponseWriter, req *http.Request) {
    for name, headers := range req.Header {
        for _, h := range headers {
            fmt.Fprintf(w, "%v: %v\n", name, h)
        }
    }
}

func main() {

    http.HandleFunc("/typeracer", hello)
    http.HandleFunc("/headers", headers)

    http.ListenAndServe(":8090", nil)
}
