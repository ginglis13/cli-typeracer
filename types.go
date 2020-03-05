package types

type ClientState struct {
	userID string
	gameID int
	progress int // length of correct input to show comparison to other players
	userInput string // TODO: check input on client or server side
	complete bool // indicates client has finished the input
	isCreate bool // indicates that the user is the game creator - for asking if they want to start another
}

type GameState struct {
	Message string
	Completed bool
	//	clients []*ClientState // take length to verify max of 4 participants
	// also use the progress attribute to check against other players
}
