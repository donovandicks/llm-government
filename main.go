package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/donovandicks/llm-government/internal"

	"github.com/go-redis/redis/v8"
)

var (
	sim      internal.Simulation = internal.Simulation{}
	fromFile string              = ""
)

func parseArgs() {
	flag.StringVar(&fromFile, "from-file", "", "Load the scenario settings from a JSON file")
	flag.StringVar(&sim.Scenario, "scenario", "", "Describe the scenario the agents are participating in")
	flag.Parse()

	if fromFile != "" {
		loaded, err := internal.LoadSimulationFromFile(fromFile)
		if err != nil {
			slog.Error("failed to load simulation config", "error", err)
			os.Exit(1)
		}
		sim = *loaded
	}

	if sim.Scenario == "" {
		slog.Error("invalid simulation config: scenario cannot be empty")
		os.Exit(1)
	}
}

func main() {
	ctx := context.Background()
	shutdown, err := internal.SetupOTelSDK(ctx)
	if err != nil {
		slog.Error("failed to initialize otel sdk", "error", err)
		os.Exit(1)
	}
	defer shutdown(ctx)

	parseArgs()
	slog.Debug("parsed simulation", "simulation", sim)

	channel := "llm-messages"

	cache := redis.NewClient(&redis.Options{
		Addr: os.Getenv("CACHE_ADDR"),
	})

	auditor := internal.NewAuditor()
	defer auditor.Stop()

	agent1 := internal.NewAgent(ctx, sim, cache, channel, auditor.AuditLog)
	defer agent1.Stop()

	agent2 := internal.NewAgent(ctx, sim, cache, channel, auditor.AuditLog)
	defer agent2.Stop()

	go auditor.Run()
	go agent1.Run()
	go agent2.Run()

	agent1.Publish(ctx, fmt.Sprintf("Hello everyone, I am agent %s. Nice to meet you.", agent1.ID))
	for {
	}
}
