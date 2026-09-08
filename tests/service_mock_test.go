package voice_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sravan-chavan/Voice-Calling-POC/pkg/voice"
)

// mockService demonstrates how consuming backends can mock voice.Service in tests.
type mockService struct {
	resp *voice.InitiateCallResponse
	err  error
}

func (m *mockService) InitiateCall(_ context.Context, _ voice.InitiateCallRequest) (*voice.InitiateCallResponse, error) {
	return m.resp, m.err
}

func TestVoiceServiceInterface_Mockable(t *testing.T) {
	var svc voice.Service = &mockService{
		resp: &voice.InitiateCallResponse{RequestUUID: "mock-uuid", Message: "ok"},
	}

	resp, err := svc.InitiateCall(context.Background(), voice.InitiateCallRequest{To: "+14155552671"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.RequestUUID != "mock-uuid" {
		t.Fatalf("unexpected uuid: %s", resp.RequestUUID)
	}
}

func TestVoiceServiceInterface_ErrorPropagation(t *testing.T) {
	var svc voice.Service = &mockService{
		err: voice.NewError(voice.ErrProviderAPI, "simulated failure", errors.New("upstream")),
	}

	_, err := svc.InitiateCall(context.Background(), voice.InitiateCallRequest{To: "+14155552671"})
	if !errors.Is(err, voice.ErrProviderAPI) {
		t.Fatalf("expected ErrProviderAPI, got %v", err)
	}
}
