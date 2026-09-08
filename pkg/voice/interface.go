package voice

import "context"

// Service initiates outbound voice calls without exposing provider details.
// Consuming backends (e.g. PhishSheriff) should depend on this interface only.
type Service interface {
	InitiateCall(ctx context.Context, req InitiateCallRequest) (*InitiateCallResponse, error)
}
