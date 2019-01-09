package message

// Message is just the structure of the data stored in the db
type Message struct {
	Channel string `json:"channel"`
	Author  string `json:"author"`
	ID      string `json:"id"`
	Text    string `json:"text"`
	File    string `json:"file"`
}
