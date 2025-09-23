package internal

import (
	"fmt"
	"log/slog"
	"os"
)

type Auditor struct {
	AuditLog chan Message
}

func NewAuditor() *Auditor {
	return &Auditor{
		AuditLog: make(chan Message),
	}
}

func (a *Auditor) Stop() {
	close(a.AuditLog)
}

func (a *Auditor) Run() {
	slog.Info("starting auditor")
	f, err := os.OpenFile("/app/audit.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		slog.Error("failed to open audit file", "error", err)
		panic(err)
	}
	defer f.Close()

	for msg := range a.AuditLog {
		slog.Debug("auditing message", "sender", msg.Metadata.Sender, "contents", msg.Contents)
		fmt.Fprintf(f, "[%s] %s - %s\n", msg.Metadata.SentAt, msg.Metadata.Sender, msg.Contents)
	}
}
