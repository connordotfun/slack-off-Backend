package marshaller

import (
	"encoding/json"

	"slack-off-Backend/message"
)

// ToJSON marshals to JSON
func ToJSON(pair [2]message.Message) string {
	byteMarshal, _ := json.Marshal(pair)
	return string(byteMarshal)
}
