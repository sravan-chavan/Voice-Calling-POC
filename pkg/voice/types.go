package voice

// InitiateCallRequest is the public request used by consuming backends.
type InitiateCallRequest struct {
	// To is the destination phone number in E.164 format (e.g. +14155552671).
	To string `json:"to"`
}

// InitiateCallResponse is a provider-agnostic result for an outbound call.
type InitiateCallResponse struct {
	// RequestUUID uniquely identifies the call request at the provider.
	RequestUUID string `json:"request_uuid"`
	// Message is a human-readable status from the provider.
	Message string `json:"message"`
	// APIID is the provider API request identifier when available.
	APIID string `json:"api_id,omitempty"`
	// From is the caller ID used for the call.
	From string `json:"from,omitempty"`
	// To is the destination number that was dialed.
	To string `json:"to,omitempty"`
}
