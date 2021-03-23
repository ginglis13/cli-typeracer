type ClientState struct {
	UserID         string `json:"UserID"`
	GameID         int    `json:"GameID"`
	Progress       int    `json:"Progress"` // length of correct input to show comparison to other players
	UserInput      string `json:"UserInput"`
	Complete       bool   `json:"Complete"` // indicates client has finished the input
	IsLeader	   bool   `json:"IsLeader"` 
	WPM 		   float64    `json:"WPM"`
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