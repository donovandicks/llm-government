package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/donovandicks/llm-government/internal"
)

var (
	TickDuration time.Duration = 16667 * time.Microsecond

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

	auditor := internal.NewAuditor()
	defer auditor.Stop()
	go auditor.Run()

	world := internal.NewWorld().
		RegisterOutput(new(internal.ApprovalMetric))

	// TODO: Register citizens
	// for range 10 {
	// 	world.NewEntity(
	//
	// 	)
	// }

	bus := internal.NewInMemoryMessageBus(auditor.AuditLog)

	council := internal.NewCouncil(bus, world, internal.CouncilOptions{MaxRounds: 3}).
		RegisterAgents(
			internal.NewAgent(ctx, sim, bus),
			internal.NewAgent(ctx, sim, bus),
		)

	go func() {
		for {
			world.Tick(ctx, TickDuration)
		}
	}()

	council.Start(ctx)
}
