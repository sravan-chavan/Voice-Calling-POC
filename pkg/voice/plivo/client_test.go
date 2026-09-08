package plivo

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	plivosdk "github.com/plivo/plivo-go/v7"

	"github.com/sravan-chavan/Voice-Calling-POC/pkg/voice"
)

type fakeCalls struct {
	resp *plivosdk.CallCreateResponse
	err  error
	last plivosdk.CallCreateParams
}

func (f *fakeCalls) Create(params plivosdk.CallCreateParams) (*plivosdk.CallCreateResponse, error) {
	f.last = params
	return f.resp, f.err
}

func TestInitiateCall_Success(t *testing.T) {
	fake := &fakeCalls{
		resp: &plivosdk.CallCreateResponse{
			ApiID:       "api-123",
			Message:     "call queued",
			RequestUUID: "req-uuid-1",
		},
	}
	client := NewClientWithCaller("+14151234567", "https://example.com/answer.xml", "GET", fake, slog.New(slog.NewTextHandler(io.Discard, nil)))

	resp, err := client.InitiateCall(context.Background(), voice.InitiateCallRequest{To: "+14155552671"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.RequestUUID != "req-uuid-1" {
		t.Fatalf("unexpected request uuid: %q", resp.RequestUUID)
	}
	if fake.last.To != "+14155552671" || fake.last.From != "+14151234567" {
		t.Fatalf("unexpected create params: %+v", fake.last)
	}
}

func TestInitiateCall_InvalidNumber(t *testing.T) {
	fake := &fakeCalls{}
	client := NewClientWithCaller("+14151234567", "https://example.com/answer.xml", "GET", fake, slog.New(slog.NewTextHandler(io.Discard, nil)))

	_, err := client.InitiateCall(context.Background(), voice.InitiateCallRequest{To: "not-a-number"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !errors.Is(err, voice.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestInitiateCall_ContextCancelled(t *testing.T) {
	fake := &fakeCalls{}
	client := NewClientWithCaller("+14151234567", "https://example.com/answer.xml", "GET", fake, slog.New(slog.NewTextHandler(io.Discard, nil)))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.InitiateCall(ctx, voice.InitiateCallRequest{To: "+14155552671"})
	if err == nil {
		t.Fatal("expected context error")
	}
}

func TestMapSDKError_Auth(t *testing.T) {
	err := mapSDKError(errors.New("401 unauthorized"))
	if !errors.Is(err, voice.ErrAuthentication) {
		t.Fatalf("expected authentication error, got %v", err)
	}
}
