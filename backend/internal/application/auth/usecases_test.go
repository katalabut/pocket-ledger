package auth

import (
	"context"
	"testing"
	"time"
)

type memCodes struct{ ok bool }

func (m *memCodes) CreateCode(ctx context.Context, email, codeSha256 string, expiresAt time.Time) error { return nil }
func (m *memCodes) ConsumeValidCode(ctx context.Context, email, codeSha256 string, now time.Time) (bool, error) {
	return m.ok, nil
}

type memUsers struct{}

func (m *memUsers) EnsureUser(ctx context.Context, email string) (string, error) { return "u1", nil }

type memEmail struct{ sent bool }

func (m *memEmail) Send(ctx context.Context, to, subject, body string) error { m.sent = true; return nil }

type memIssuer struct{}

func (m *memIssuer) Issue(userID, email string, now time.Time) (string, error) { return "tok", nil }

func TestConfirmCode_Invalid(t *testing.T) {
	svc := NewService(&memCodes{ok: false}, &memUsers{}, &memEmail{}, &memIssuer{}, 10*time.Minute, 6, func() time.Time { return time.Unix(0, 0) })
	_, err := svc.ConfirmCode(context.Background(), "a@b.com", "000000")
	if err == nil {
		t.Fatalf("expected error")
	}
}
