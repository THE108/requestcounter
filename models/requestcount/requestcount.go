package requestcount

import (
	"golang.org/x/net/context"
)

type RequestCount struct {
	Count uint64 `json:"count"`
}

type IRequestCounter interface {
	Get(ctx context.Context) *RequestCount
	Run() error
	Close() error
}
