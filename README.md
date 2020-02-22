# cli-typeracer

### Q's/Strat
- should client verify input or server?
  - currently client. leaving it as that. will put on server eventually if needed
  - if this is the case, server will more or less just be there to connect users.
- connecting multiple clients to same game
  - generate game id or allow client to use 'password' to all join same game?
  - real typeracer generates unique url to share w/ peeps
  - use of goroutines
- ensure that first finished client wins
  - todo-ish - maybe it will just work ?
- client - show status bar of how far they are in comparison to others
  - use #'s * length of correct input entered, mod 20 ?

### client

`main()`
TODO: cli parsing
- host
- port
- maybe -nick [nick] -join [gameid]
`join_game()`
- input a game id to join w others
- prompt for a nick if the user didn't specify in the cl args
- send client state to server
`create_game()`
- return a game id to initialize a game for others to join
- prompt for a nick if the user didn't specify in the cl args
- send client state to server
- **maybe juts call join_game() w/ the newly created game id**
```go
type ClientState struct {
	userID string
	gameID int
	progress int // length of correct input to show comparison to other players
	userInput string // TODO: check input on client or server side
	complete bool // indicates client has finished the input
}
```


### server

- **TCP vs. HTTP**
`init_game()`
- create game id, return to client
- should probably start a new goroutine, esp w multiple games occurring simultaneously on server
`end_game()`
- check that one client with correct game id has sent a game state that indicates they finished
- should stop game on all other clients
- free the game id from the map
  - maybe use a free list of game ids... or not, this app doesn't yet have to be that complicated
`choose_quote()`
- pick from list of quotes/phrases wtvr
- prolly just a text file
```go
type GameState struct {
	clients []*ClientState // take length to verify max of 4 participants
	// also use the progress attribute to check against other players
}
```
