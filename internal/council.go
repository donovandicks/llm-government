package internal

import (
	"context"
	"fmt"
)

type CouncilOptions struct {
	MaxRounds int
}

type Council struct {
	agents map[string]*Agent
	bus    MessageBus
	world  *World

	opts CouncilOptions
}

func NewCouncil(bus MessageBus, w *World, opts CouncilOptions) *Council {
	return &Council{
		agents: make(map[string]*Agent),
		bus:    bus,
		world:  w,
		opts:   opts,
	}
}

func (c *Council) RegisterAgents(agents ...*Agent) *Council {
	for _, a := range agents {
		c.agents[a.ID] = a
	}
	return c
}

func (c *Council) AgentCount() int { return len(c.agents) }

func (c *Council) initMessage() string {
	return new(PromptBuilder).
		WithTask(
			"Begin the simulation.",
			WithItems(
				fmt.Sprintf("There are %d total agents on the council.", c.AgentCount()),
				fmt.Sprintf("You have at most %d rounds of discussion before the next world state observation.", c.opts.MaxRounds),
				"The world does not stop during council deliberations.",
			),
		).
		Build()
}

func (c *Council) Start(ctx context.Context) {
	c.bus.Publish(ctx, "<system>", c.initMessage())

	for {
		// Observe the world
		obs := c.world.Observe(ctx)

		// Agent discussion
		for range c.opts.MaxRounds {
			for _, a := range c.agents {
				a.Run(ctx, &obs)
			}
		}
	}
}
