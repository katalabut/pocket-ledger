package auth

import (
	"context"
	"time"
)

type EmailSender interface {
	Send(ctx context.Context, to, subject, body string) error
}

type CodeRepository interface {
	CreateCode(ctx context.Context, email, codeSha256 string, expiresAt time.Time) error
	ConsumeValidCode(ctx context.Context, email, codeSha256 string, now time.Time) (bool, error)
}

type UserRepository interface {
	EnsureUser(ctx context.Context, email string) (userID string, err error)
}
