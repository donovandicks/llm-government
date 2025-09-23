package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/donovandicks/llm-government/internal"

	"github.com/go-redis/redis/v8"
)

func init() {
	// TODO: Parameterize
	slog.SetLogLoggerLevel(slog.LevelDebug)
}

func main() {
	ctx := context.Background()

	channel := "llm-messages"

	cache := redis.NewClient(&redis.Options{
		Addr: os.Getenv("CACHE_ADDR"),
	})

	auditor := internal.NewAuditor()
	defer auditor.Stop()

	agent1 := internal.NewAgent(ctx, cache, channel, auditor.AuditLog)
	defer agent1.Stop()

	agent2 := internal.NewAgent(ctx, cache, channel, auditor.AuditLog)
	defer agent2.Stop()

	go auditor.Run()
	go agent1.Run()
	go agent2.Run()

	agent1.Publish(ctx, fmt.Sprintf("Hello everyone, I am agent %s. Nice to meet you.", agent1.ID))
	for {
	}
}
