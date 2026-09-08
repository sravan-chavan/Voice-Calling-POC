package plivo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"

	plivosdk "github.com/plivo/plivo-go/v7"

	"github.com/sravan-chavan/Voice-Calling-POC/pkg/voice"
)

// callCreator abstracts the Plivo Calls API for unit testing.
type callCreator interface {
	Create(params plivosdk.CallCreateParams) (*plivosdk.CallCreateResponse, error)
}

// Client implements voice.Service using the official Plivo Go SDK.
type Client struct {
	from         string
	answerURL    string
	answerMethod string
	calls        callCreator
	logger       *slog.Logger
}

// Config holds Plivo-specific settings needed to construct a Client.
type Config struct {
	AuthID       string
	AuthToken    string
	FromNumber   string
	AnswerURL    string
	AnswerMethod string
	Logger       *slog.Logger
}

// NewClient creates a Plivo-backed voice.Service.
func NewClient(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.AuthID) == "" || strings.TrimSpace(cfg.AuthToken) == "" {
		return nil, voice.NewError(voice.ErrConfiguration, "plivo auth id and auth token are required", nil)
	}
	if strings.TrimSpace(cfg.FromNumber) == "" {
		return nil, voice.NewError(voice.ErrConfiguration, "plivo from number is required", nil)
	}
	if strings.TrimSpace(cfg.AnswerURL) == "" {
		return nil, voice.NewError(voice.ErrConfiguration, "plivo answer url is required", nil)
	}

	method := strings.ToUpper(strings.TrimSpace(cfg.AnswerMethod))
	if method == "" {
		method = "GET"
	}

	sdkClient, err := plivosdk.NewClient(cfg.AuthID, cfg.AuthToken, &plivosdk.ClientOptions{})
	if err != nil {
		return nil, mapSDKError(err)
	}

	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return &Client{
		from:         strings.TrimSpace(cfg.FromNumber),
		answerURL:    strings.TrimSpace(cfg.AnswerURL),
		answerMethod: method,
		calls:        sdkClient.Calls,
		logger:       logger,
	}, nil
}

// NewClientWithCaller constructs a Client with an injected call creator (tests).
func NewClientWithCaller(from, answerURL, answerMethod string, calls callCreator, logger *slog.Logger) *Client {
	if logger == nil {
		logger = slog.Default()
	}
	if answerMethod == "" {
		answerMethod = "GET"
	}
	return &Client{
		from:         from,
		answerURL:    answerURL,
		answerMethod: answerMethod,
		calls:        calls,
		logger:       logger,
	}
}

// InitiateCall places an outbound call via Plivo.
func (c *Client) InitiateCall(ctx context.Context, req voice.InitiateCallRequest) (*voice.InitiateCallResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, voice.NewError(voice.ErrUnexpected, "context cancelled before call initiation", err)
	}

	req.To = voice.NormalizePhone(req.To)
	if err := voice.ValidateInitiateCallRequest(req); err != nil {
		return nil, err
	}

	c.logger.Info("call request received", "to", req.To, "from", c.from)

	resp, err := c.calls.Create(plivosdk.CallCreateParams{
		From:         c.from,
		To:           req.To,
		AnswerURL:    c.answerURL,
		AnswerMethod: c.answerMethod,
	})
	if err != nil {
		c.logger.Error("call failed", "to", req.To, "error", sanitizeErr(err))
		return nil, mapSDKError(err)
	}

	out := &voice.InitiateCallResponse{
		Message:     resp.Message,
		APIID:       resp.ApiID,
		RequestUUID: firstRequestUUID(resp.RequestUUID),
		From:        c.from,
		To:          req.To,
	}

	c.logger.Info("call initiated successfully",
		"to", req.To,
		"from", c.from,
		"request_uuid", out.RequestUUID,
		"api_id", out.APIID,
	)

	return out, nil
}

// firstRequestUUID normalizes Plivo's request_uuid which may be a string or []string.
func firstRequestUUID(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case []string:
		if len(v) > 0 {
			return v[0]
		}
	case []interface{}:
		if len(v) > 0 {
			if s, ok := v[0].(string); ok {
				return s
			}
		}
	}
	return ""
}

func mapSDKError(err error) error {
	if err == nil {
		return nil
	}

	var netErr net.Error
	if errors.As(err, &netErr) || isNetworkError(err) {
		return voice.NewError(voice.ErrNetwork, "failed to reach Plivo API", err)
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "401"), strings.Contains(msg, "unauthorized"), strings.Contains(msg, "authentication"):
		return voice.NewError(voice.ErrAuthentication, "plivo authentication failed; check PLIVO_AUTH_ID and PLIVO_AUTH_TOKEN", err)
	case strings.Contains(msg, "403"), strings.Contains(msg, "forbidden"):
		return voice.NewError(voice.ErrAuthentication, "plivo rejected credentials or caller id authorization", err)
	default:
		return voice.NewError(voice.ErrProviderAPI, fmt.Sprintf("plivo api error: %s", sanitizeErr(err)), err)
	}
}

func isNetworkError(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "no such host") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "temporary failure") ||
		strings.Contains(msg, "i/o timeout")
}

// sanitizeErr returns an error string that should never include auth tokens.
func sanitizeErr(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
