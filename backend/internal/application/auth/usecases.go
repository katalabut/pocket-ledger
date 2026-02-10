package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/katalabut/pocket-ledger/backend/internal/platform/cryptoutil"
)

type RequestCodeInput struct {
	Email string
}

type RequestCodeOutput struct {
	Sent bool `json:"sent"`
}

type ConfirmCodeInput struct {
	Email string
	Code  string
}

type ConfirmCodeOutput struct {
	Token string
}

type TokenIssuer interface {
	Issue(userID, email string, now time.Time) (string, error)
}

type Service struct {
	codes      CodeRepository
	users      UserRepository
	email      EmailSender
	issuer     TokenIssuer
	codeTTL    time.Duration
	codeLength int
	clock      func() time.Time
}

func NewService(codes CodeRepository, users UserRepository, email EmailSender, issuer TokenIssuer, codeTTL time.Duration, codeLength int, clock func() time.Time) *Service {
	if clock == nil {
		clock = time.Now
	}
	return &Service{codes: codes, users: users, email: email, issuer: issuer, codeTTL: codeTTL, codeLength: codeLength, clock: clock}
}

func (s *Service) RequestCode(ctx context.Context, emailRaw string) (RequestCodeOutput, error) {
	email := normalizeEmail(emailRaw)
	if email == "" {
		return RequestCodeOutput{}, fmt.Errorf("email is required")
	}
	code, err := cryptoutil.RandomDigits(s.codeLength)
	if err != nil {
		return RequestCodeOutput{}, err
	}
	now := s.clock().UTC()
	expires := now.Add(s.codeTTL)
	sha := cryptoutil.SHA256Hex(email + ":" + code)
	if err := s.codes.CreateCode(ctx, email, sha, expires); err != nil {
		return RequestCodeOutput{}, err
	}

	subject := "Your pocket-ledger login code"
	body := fmt.Sprintf("Your login code: %s\n\nIt expires in %d minutes.", code, int(s.codeTTL.Minutes()))
	if err := s.email.Send(ctx, email, subject, body); err != nil {
		return RequestCodeOutput{}, err
	}
	return RequestCodeOutput{Sent: true}, nil
}

func (s *Service) ConfirmCode(ctx context.Context, emailRaw, codeRaw string) (ConfirmCodeOutput, error) {
	email := normalizeEmail(emailRaw)
	code := strings.TrimSpace(codeRaw)
	if email == "" || code == "" {
		return ConfirmCodeOutput{}, fmt.Errorf("email and code are required")
	}
	sha := cryptoutil.SHA256Hex(email + ":" + code)
	now := s.clock().UTC()
	ok, err := s.codes.ConsumeValidCode(ctx, email, sha, now)
	if err != nil {
		return ConfirmCodeOutput{}, err
	}
	if !ok {
		return ConfirmCodeOutput{}, fmt.Errorf("invalid code")
	}
	userID, err := s.users.EnsureUser(ctx, email)
	if err != nil {
		return ConfirmCodeOutput{}, err
	}
	tok, err := s.issuer.Issue(userID, email, now)
	if err != nil {
		return ConfirmCodeOutput{}, err
	}
	return ConfirmCodeOutput{Token: tok}, nil
}

func normalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
