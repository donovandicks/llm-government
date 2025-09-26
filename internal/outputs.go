package internal

import "context"

type OutputValue struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Value       any    `json:"value"`
}

type Output interface {
	Name() string
	Description() string
	Compute(context.Context, *World) OutputValue
}

type ApprovalMetric struct{}

func (m *ApprovalMetric) Name() string { return "approval_rating" }

func (m *ApprovalMetric) Description() string {
	return "The current approval rating of the population"
}

func (m *ApprovalMetric) Compute(ctx context.Context, w *World) OutputValue {
	// List all 'people' entities
	// Calculate the overall approval rating
	return OutputValue{Name: m.Name(), Description: m.Description(), Value: 0}
}
