package core

import "context"

// CollectorService handles incoming messages from sources.
type CollectorService interface {
	// Handle processes an incoming message.
	Handle(ctx context.Context, msg Message) error
}
