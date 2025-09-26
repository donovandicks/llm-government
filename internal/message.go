package internal

import (
	"context"
	"encoding/json"
	"sync"
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

type MessageBus interface {
	// TODO: Consider merging AuditLog with the MessageBus
	PublishAudit(context.Context, Message) error
	Publish(context.Context, string, string) error
	Subscribe(ctx context.Context, subscriber string)
	Drain(ctx context.Context, subscriber string) []Message
}

type InMemoryBus struct {
	sync.Mutex

	auditLog chan<- Message       // Audit log records all messages published on the bus
	buffers  map[string][]Message // Message buffers registered per subscriber
}

func NewInMemoryMessageBus(auditLog chan<- Message) *InMemoryBus {
	return &InMemoryBus{
		auditLog: auditLog,
		buffers:  make(map[string][]Message),
	}
}

func (b *InMemoryBus) PublishAudit(ctx context.Context, msg Message) error {
	b.auditLog <- msg
	return nil
}

func (b *InMemoryBus) Publish(ctx context.Context, from, msg string) error {
	b.Lock()
	defer b.Unlock()

	message := NewMessage(from, msg)

	b.PublishAudit(ctx, message)

	// NOTE: Consider supporting direct messages between agents.
	// Consider making this configurable, e.g. DM_ALLOWED, to allow
	// agents to communicate "in private"

	for subscriber := range b.buffers {
		if subscriber == from {
			continue
		}

		b.buffers[subscriber] = append(b.buffers[subscriber], message)
	}

	return nil
}

func (b *InMemoryBus) Subscribe(ctx context.Context, subscriber string) {
	b.Lock()
	defer b.Unlock()

	b.buffers[subscriber] = []Message{}

	// ch := make(chan Message)
	//
	// // TODO: Implement function to grab messages from the buffer and drop into the channel
	// go func() {
	// 	ticker := time.NewTicker(10 * time.Millisecond)
	// 	defer ticker.Stop()
	// 	for range ticker.C {
	// 		msgs := b.Drain(ctx, subscriber)
	// 		for _, m := range msgs {
	// 			ch <- m
	// 		}
	// 	}
	// }()
	//
	// return ch, nil
}

func (b *InMemoryBus) Drain(ctx context.Context, subscriber string) []Message {
	b.Lock()
	defer b.Unlock()

	msgs := b.buffers[subscriber]
	b.buffers[subscriber] = nil

	return msgs
}
