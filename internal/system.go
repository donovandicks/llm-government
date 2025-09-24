package internal

import (
	"context"
	"time"
)

type System interface {
	Name() string
	Update(context.Context, *World, time.Duration)
}
