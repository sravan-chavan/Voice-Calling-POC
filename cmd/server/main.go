package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sravan-chavan/Voice-Calling-POC/internal/config"
	"github.com/sravan-chavan/Voice-Calling-POC/pkg/voice"
	plivovoice "github.com/sravan-chavan/Voice-Calling-POC/pkg/voice/plivo"
)

func main() {
	toFlag := flag.String("to", "", "destination phone number in E.164 format (e.g. +14155552671)")
	serve := flag.Bool("serve", false, "run HTTP server mode instead of one-shot call")
	httpAddr := flag.String("http-addr", "", "HTTP listen address (overrides HTTP_ADDR), e.g. :8080")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	logger.Info("application started")

	cfg, err := config.Load()
	if err != nil {
		logger.Error("configuration failed", "error", err.Error())
		os.Exit(1)
	}
	logger.Info("configuration loaded", "config", cfg.Redacted())

	voiceService, err := plivovoice.NewClient(plivovoice.Config{
		AuthID:       cfg.AuthID,
		AuthToken:    cfg.AuthToken,
		FromNumber:   cfg.FromNumber,
		AnswerURL:    cfg.AnswerURL,
		AnswerMethod: cfg.AnswerMethod,
		Logger:       logger,
	})
	if err != nil {
		logger.Error("voice service initialization failed", "error", err.Error())
		os.Exit(1)
	}
	logger.Info("voice service initialized")

	addr := firstNonEmpty(*httpAddr, cfg.HTTPAddr)
	if *serve || addr != "" {
		if addr == "" {
			addr = ":8080"
		}
		if err := runServer(addr, voiceService, logger); err != nil {
			logger.Error("server stopped with error", "error", err.Error())
			os.Exit(1)
		}
		return
	}

	to := firstNonEmpty(*toFlag, cfg.CallTo)
	if to == "" {
		logger.Error("destination number required: pass -to or set CALL_TO")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := voiceService.InitiateCall(ctx, voice.InitiateCallRequest{To: to})
	if err != nil {
		logger.Error("call failed", "error", err.Error(), "kind", errorKind(err))
		os.Exit(1)
	}

	logger.Info("call completed",
		"request_uuid", resp.RequestUUID,
		"api_id", resp.APIID,
		"message", resp.Message,
		"to", resp.To,
		"from", resp.From,
	)
	fmt.Printf("Call initiated: request_uuid=%s message=%s\n", resp.RequestUUID, resp.Message)
}

func runServer(addr string, svc voice.Service, logger *slog.Logger) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("POST /v1/calls", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			To string `json:"to"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		logger.Info("http call request received", "to", body.To)
		resp, err := svc.InitiateCall(r.Context(), voice.InitiateCallRequest{To: body.To})
		if err != nil {
			status := http.StatusInternalServerError
			switch {
			case errors.Is(err, voice.ErrInvalidInput):
				status = http.StatusBadRequest
			case errors.Is(err, voice.ErrAuthentication):
				status = http.StatusUnauthorized
			case errors.Is(err, voice.ErrConfiguration):
				status = http.StatusInternalServerError
			}
			logger.Error("http call failed", "error", err.Error(), "kind", errorKind(err))
			writeErr(w, status, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(resp)
	})

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("http server listening", "addr", addr)
		errCh <- server.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.Info("shutdown signal received", "signal", sig.String())
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
		return nil
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func writeErr(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func errorKind(err error) string {
	switch {
	case errors.Is(err, voice.ErrInvalidInput):
		return "invalid_input"
	case errors.Is(err, voice.ErrConfiguration):
		return "configuration"
	case errors.Is(err, voice.ErrAuthentication):
		return "authentication"
	case errors.Is(err, voice.ErrProviderAPI):
		return "provider_api"
	case errors.Is(err, voice.ErrNetwork):
		return "network"
	default:
		return "unexpected"
	}
}
