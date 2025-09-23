package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/responses"
	"github.com/openai/openai-go/shared"
)

var SystemPrompt string = new(PromptBuilder).
	WithRole("You are a participant in a simulation of government, like a Model UN.").
	WithSystemMessage("You are to read messages from other participants and respond accordingly.").
	Build()

type Agent struct {
	ID       string
	logger   *slog.Logger
	auditLog chan Message

	client openai.Client
	model  string

	cache        *redis.Client
	channel      string
	subscription *redis.PubSub
}

func NewAgent(ctx context.Context, cache *redis.Client, channel string, auditLog chan Message) *Agent {
	agentId := uuid.NewString()
	subscription := cache.Subscribe(ctx, channel)
	logger := slog.With("agentId", agentId)

	return &Agent{
		ID:           agentId,
		logger:       logger,
		auditLog:     auditLog,
		client:       openai.NewClient(),
		model:        "gpt-5",
		cache:        cache,
		channel:      channel,
		subscription: subscription,
	}
}

func (a *Agent) Stop() {
	a.subscription.Close()
}

func (a *Agent) Publish(ctx context.Context, msg string) {
	message := NewMessage(a.ID, msg)
	a.auditLog <- message
	a.cache.Publish(ctx, a.channel, message.Bytes())
}

func (a *Agent) readMessage(ctx context.Context, msg Message) (string, error) {
	// TODO: We need to keep track of message history pretty much immediately
	prompt := new(PromptBuilder).
		WithTask("You have received a new message. Read it and generate a reply.").
		WithItems(
			fmt.Sprintf("You are agent %s", a.ID),
			fmt.Sprintf("The sender agent was %s", msg.Sender),
			fmt.Sprintf("The message was sent at %s", msg.Metadata.SentAt),
			fmt.Sprintf("The current time is %s", time.Now().Format(time.RFC3339)),
		).
		WithIntroducer("Here is the message").
		WithUserMessage(msg.Contents).
		Build()

	params := responses.ResponseNewParams{
		Model:        a.model,
		Instructions: openai.String(SystemPrompt),
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(prompt),
		},
		Reasoning: shared.ReasoningParam{Effort: shared.ReasoningEffortMedium},
	}

	response, err := a.client.Responses.New(ctx, params)
	if err != nil {
		return "", err
	}

	return response.OutputText(), nil
}

func (a *Agent) Run() {
	slog.Info("starting agent", "id", a.ID)

	for m := range a.subscription.Channel() {
		var msg Message
		err := json.Unmarshal([]byte(m.Payload), &msg)
		if err != nil {
			a.logger.Error("failed to parse pubsub message", "error", err)
			continue
		}

		if msg.Sender == a.ID {
			// Skip reading messages sent by the current agent
			continue
		}

		ctx := context.Background()
		reply, err := a.readMessage(ctx, msg)
		if err != nil {
			a.logger.Error("failed to generate a reply to message", "error", err)
			continue
		}

		a.Publish(ctx, reply)

		// 3. Take Action?
		// 4. Reply?
	}
}
