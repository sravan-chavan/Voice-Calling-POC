package voice

import (
	"testing"
)

func TestValidateInitiateCallRequest(t *testing.T) {
	tests := []struct {
		name    string
		to      string
		wantErr bool
	}{
		{name: "valid us", to: "+14155552671", wantErr: false},
		{name: "valid india", to: "+919876543210", wantErr: false},
		{name: "empty", to: "", wantErr: true},
		{name: "whitespace", to: "   ", wantErr: true},
		{name: "missing plus", to: "14155552671", wantErr: true},
		{name: "too short", to: "+1234", wantErr: true},
		{name: "letters", to: "+14abc5552671", wantErr: true},
		{name: "leading zero country", to: "+014155552671", wantErr: true},
	}

	for _, tt := range tests {	
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInitiateCallRequest(InitiateCallRequest{To: tt.to})
			if tt.wantErr && err == nil {
				t.Fatalf("expected error for %q", tt.to)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.to, err)
			}
		})
	}
}
