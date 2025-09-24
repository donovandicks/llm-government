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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Agent struct {
	ID         string
	logger     *slog.Logger
	auditLog   chan Message
	simulation Simulation

	client       openai.Client
	model        string
	systemPrompt string

	cache        *redis.Client
	channel      string
	subscription *redis.PubSub
}

func NewAgent(ctx context.Context, sim Simulation, cache *redis.Client, channel string, auditLog chan Message) *Agent {
	agentId := uuid.NewString()
	subscription := cache.Subscribe(ctx, channel)
	logger := slog.With("agentId", agentId)

	systemPrompt := new(PromptBuilder).
		WithRole("You are acting in a simulation with other agents.").
		WithIntroducer("Here is the scenario.").
		WithParagraph(sim.Scenario).
		// TODO: Incorporate other simulation details
		WithSystemMessage("Read messages from other participants and respond accordingly.").
		Build()

	return &Agent{
		ID:           agentId,
		logger:       logger,
		auditLog:     auditLog,
		simulation:   sim,
		client:       openai.NewClient(),
		model:        "gpt-5",
		systemPrompt: systemPrompt,
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
	ctx, span := Tracer.Start(ctx, "read message", trace.WithAttributes(
		attribute.String("simulation", a.simulation.ID()),
		attribute.String("model", a.model),
		attribute.String("agent", a.ID),
		attribute.String("systemPrompt", a.systemPrompt),
	))
	defer span.End()

	// TODO: We need to keep track of message history pretty much immediately
	prompt := new(PromptBuilder).
		WithTask(
			"You have received a new message. Read it and generate a reply.",
			WithItems(
				fmt.Sprintf("You are agent %s", a.ID),
				fmt.Sprintf("The sender agent was %s", msg.Metadata.Sender),
				fmt.Sprintf("The message was sent at %s", msg.Metadata.SentAt),
				fmt.Sprintf("The current time is %s", time.Now().Format(time.RFC3339)),
			),
		).
		WithIntroducer("Here is the message").
		WithUserMessage(msg.Contents).
		Build()

	params := responses.ResponseNewParams{
		Model:        a.model,
		Instructions: openai.String(a.systemPrompt),
		Input:        responses.ResponseNewParamsInputUnion{OfString: openai.String(prompt)},
		Reasoning:    shared.ReasoningParam{Effort: shared.ReasoningEffortMedium},
	}

	response, err := a.client.Responses.New(ctx, params)
	if err != nil {
		return "", err
	}

	text := response.OutputText()
	span.SetAttributes(attribute.String("response", text))

	return text, nil
}

func (a *Agent) Observe(ctx context.Context, w *World) Observation {
	w.RLock()
	defer w.RUnlock()

	inputs := make(map[string]any)
	for name, in := range w.inputs {
		inputs[name] = in.Get()
	}

	return Observation{
		Tick:   w.tick,
		Inputs: inputs,
	}
}

func (a *Agent) Decide(ctx context.Context, obs Observation) []Action {
	// TODO: Implement
	panic("unimplemented")
}

func (a *Agent) Run() {
	a.logger.Info("starting agent")

	for m := range a.subscription.Channel() {
		var msg Message
		err := json.Unmarshal([]byte(m.Payload), &msg)
		if err != nil {
			a.logger.Error("failed to parse pubsub message", "error", err)
			continue
		}

		// Skip reading messages sent by the current agent
		if msg.Metadata.Sender == a.ID {
			continue
		}

		ctx := context.Background()
		reply, err := a.readMessage(ctx, msg)
		if err != nil {
			a.logger.Error("failed to generate a reply to message", "error", err)
			continue
		}

		a.Publish(ctx, reply)

		// Take Action?
	}
}
