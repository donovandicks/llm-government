package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

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
	simulation Simulation

	client       openai.Client
	model        string
	systemPrompt string

	bus MessageBus
}

func NewAgent(ctx context.Context, sim Simulation, bus MessageBus) *Agent {
	agentId := uuid.NewString()
	logger := slog.With("agentId", agentId)

	bus.Subscribe(ctx, agentId)

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
		simulation:   sim,
		client:       openai.NewClient(),
		model:        "gpt-5",
		systemPrompt: systemPrompt,
		bus:          bus,
	}
}

func (a *Agent) readInbox(ctx context.Context, inbox []Message, obs *Observation) (string, error) {
	// TODO: Keep track of conversation history in agent context

	taskPrompt := new(PromptBuilder).
		WithTask(
			"You have received at least one new message. Read it/them and generate a reply.",
			WithItems(
				fmt.Sprintf("You are agent %s", a.ID),
				fmt.Sprintf("The current time is %s", time.Now().Format(time.RFC3339)),
			),
		).
		Build()

	messages := []string{}
	for _, msg := range inbox {
		prompt := new(PromptBuilder).
			WithItems(
				fmt.Sprintf("This is message %s", msg.Metadata.ID),
				fmt.Sprintf("This message is from %s", msg.Metadata.Sender),
				fmt.Sprintf("This message was sent at %s", msg.Metadata.SentAt),
			).
			Build()
		messages = append(messages, prompt)
	}

	inboxPrompt := new(PromptBuilder).
		WithIntroducer("Here is your inbox:").
		WithItems(messages...).
		Build()

	bs, _ := json.Marshal(obs)
	worldPrompt := new(PromptBuilder).
		WithIntroducer("Here is the world state:").
		WithCode(string(bs), "json").
		Build()

	params := responses.ResponseNewParams{
		Model:        a.model,
		Instructions: openai.String(a.systemPrompt),
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: responses.ResponseInputParam{
				{OfMessage: &responses.EasyInputMessageParam{
					Content: responses.EasyInputMessageContentUnionParam{OfString: openai.String(taskPrompt)},
					Role:    "system",
				}},
				{OfMessage: &responses.EasyInputMessageParam{
					Content: responses.EasyInputMessageContentUnionParam{OfString: openai.String(worldPrompt)},
					Role:    "system",
				}},
				{OfMessage: &responses.EasyInputMessageParam{
					Content: responses.EasyInputMessageContentUnionParam{OfString: openai.String(inboxPrompt)},
					Role:    "assistant",
				}},
			},
		},
		Reasoning: shared.ReasoningParam{Effort: shared.ReasoningEffortMedium},
	}

	response, err := a.client.Responses.New(ctx, params)
	if err != nil {
		return "", err
	}

	text := response.OutputText()
	trace.SpanFromContext(ctx).SetAttributes(attribute.String("response", text))

	return text, nil
}

func (a *Agent) Run(ctx context.Context, obs *Observation) {
	ctx, span := Tracer.Start(ctx, "run agent", trace.WithAttributes(
		attribute.String("simulation", a.simulation.ID()),
		attribute.String("model", a.model),
		attribute.String("agent", a.ID),
		attribute.String("observation", obs.ToJSON()),
	))
	defer span.End()

	msgs := a.bus.Drain(ctx, a.ID)
	span.SetAttributes(attribute.Int("inboxSize", len(msgs)))
	if len(msgs) == 0 {
		return
	}

	reply, err := a.readInbox(ctx, msgs, obs)
	if err != nil {
		a.logger.Error("failed to generate a reply to message", "error", err)
		return
	}

	a.bus.Publish(ctx, a.ID, reply)

	// Take Action?
}
