package internal

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Metadata struct {
	ID     string `json:"id"`     // Unique identifier of the specific message
	Sender string `json:"sender"` // The ID of the agent that sent the message
	SentAt string `json:"sentAt"` // RFC3339 When the message was sent by the agent
}

type Message struct {
	Contents string   `json:"contents"` // The actual contents of the message
	Metadata Metadata `json:"metadata"`
}

func NewMessage(sender, contents string) Message {
	return Message{
		Contents: contents,
		Metadata: Metadata{
			ID:     uuid.NewString(),
			Sender: sender,
			SentAt: time.Now().Format(time.RFC3339),
		},
	}
}

func (m Message) Bytes() []byte {
	bs, _ := json.Marshal(m)
	return bs
}
