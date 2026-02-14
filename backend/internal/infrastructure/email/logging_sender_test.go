package email

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"
)

func TestLoggingSender_Send(t *testing.T) {
	var buf bytes.Buffer
	sender := NewLoggingSender(log.New(&buf, "", 0), "noreply@example.com")

	err := sender.Send(context.Background(), "user@example.com", "Magic code", "123456")
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	logged := buf.String()
	for _, want := range []string{"email test mode", "noreply@example.com", "user@example.com", "Magic code", "123456"} {
		if !strings.Contains(logged, want) {
			t.Fatalf("log output %q does not contain %q", logged, want)
		}
	}
}
