package message

// Message is just the structure of the data stored in the db
type Message struct {
	Channel string `json:"channel"`
	ID      string `json:"id"`
	Created int    `json:"created"`
	Author  string `json:"author"`
	Text    string `json:"text"`
}
