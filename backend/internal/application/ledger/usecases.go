package ledger

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/katalabut/pocket-ledger/backend/internal/domain"
)

type Service struct {
	accounts     AccountRepository
	categories   CategoryRepository
	transactions TransactionRepository
	splits       SplitRepository
	clock        func() time.Time
}

func NewService(
	accounts AccountRepository,
	categories CategoryRepository,
	transactions TransactionRepository,
	splits SplitRepository,
	clock func() time.Time,
) *Service {
	if clock == nil {
		clock = func() time.Time { return time.Now().UTC() }
	}
	return &Service{
		accounts:     accounts,
		categories:   categories,
		transactions: transactions,
		splits:       splits,
		clock:        clock,
	}
}

// --- Accounts ---

type CreateAccountInput struct {
	Name                string `json:"name"`
	Currency            string `json:"currency"`
	Type                string `json:"type"`
	InitialBalanceMinor int64  `json:"initial_balance_minor"`
}

func (s *Service) CreateAccount(ctx context.Context, in CreateAccountInput) (*domain.Account, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, fmt.Errorf("%w: name is required", domain.ErrValidation)
	}
	if strings.TrimSpace(in.Currency) == "" || len(in.Currency) != 3 {
		return nil, fmt.Errorf("%w: currency must be 3-letter ISO code", domain.ErrValidation)
	}
	if !domain.ValidAccountType(in.Type) {
		return nil, fmt.Errorf("%w: invalid account type", domain.ErrValidation)
	}
	now := s.clock()
	a := &domain.Account{
		ID:                  uuid.NewString(),
		Name:                strings.TrimSpace(in.Name),
		Currency:            strings.ToUpper(strings.TrimSpace(in.Currency)),
		Type:                domain.AccountType(in.Type),
		InitialBalanceMinor: in.InitialBalanceMinor,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	if err := s.accounts.Create(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Service) ListAccounts(ctx context.Context) ([]domain.Account, error) {
	return s.accounts.List(ctx)
}

func (s *Service) GetAccount(ctx context.Context, id string) (*domain.Account, error) {
	return s.accounts.GetByID(ctx, id)
}

type UpdateAccountInput struct {
	Name                *string `json:"name,omitempty"`
	Currency            *string `json:"currency,omitempty"`
	Type                *string `json:"type,omitempty"`
	InitialBalanceMinor *int64  `json:"initial_balance_minor,omitempty"`
}

func (s *Service) UpdateAccount(ctx context.Context, id string, in UpdateAccountInput) (*domain.Account, error) {
	a, err := s.accounts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.Name != nil {
		n := strings.TrimSpace(*in.Name)
		if n == "" {
			return nil, fmt.Errorf("%w: name is required", domain.ErrValidation)
		}
		a.Name = n
	}
	if in.Currency != nil {
		c := strings.ToUpper(strings.TrimSpace(*in.Currency))
		if len(c) != 3 {
			return nil, fmt.Errorf("%w: currency must be 3-letter ISO code", domain.ErrValidation)
		}
		a.Currency = c
	}
	if in.Type != nil {
		if !domain.ValidAccountType(*in.Type) {
			return nil, fmt.Errorf("%w: invalid account type", domain.ErrValidation)
		}
		a.Type = domain.AccountType(*in.Type)
	}
	if in.InitialBalanceMinor != nil {
		a.InitialBalanceMinor = *in.InitialBalanceMinor
	}
	a.UpdatedAt = s.clock()
	if err := s.accounts.Update(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Service) DeleteAccount(ctx context.Context, id string) error {
	return s.accounts.Delete(ctx, id)
}

// --- Categories ---

type CreateCategoryInput struct {
	Name     string  `json:"name"`
	ParentID *string `json:"parent_id,omitempty"`
}

func (s *Service) CreateCategory(ctx context.Context, in CreateCategoryInput) (*domain.Category, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, fmt.Errorf("%w: name is required", domain.ErrValidation)
	}
	if in.ParentID != nil && *in.ParentID != "" {
		if _, err := s.categories.GetByID(ctx, *in.ParentID); err != nil {
			return nil, fmt.Errorf("%w: parent category not found", domain.ErrForeignKey)
		}
	}
	now := s.clock()
	c := &domain.Category{
		ID:        uuid.NewString(),
		Name:      strings.TrimSpace(in.Name),
		ParentID:  in.ParentID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.categories.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return s.categories.List(ctx)
}

func (s *Service) GetCategory(ctx context.Context, id string) (*domain.Category, error) {
	return s.categories.GetByID(ctx, id)
}

type UpdateCategoryInput struct {
	Name     *string `json:"name,omitempty"`
	ParentID *string `json:"parent_id,omitempty"`
}

func (s *Service) UpdateCategory(ctx context.Context, id string, in UpdateCategoryInput) (*domain.Category, error) {
	c, err := s.categories.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.Name != nil {
		n := strings.TrimSpace(*in.Name)
		if n == "" {
			return nil, fmt.Errorf("%w: name is required", domain.ErrValidation)
		}
		c.Name = n
	}
	if in.ParentID != nil {
		c.ParentID = in.ParentID
	}
	c.UpdatedAt = s.clock()
	if err := s.categories.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) DeleteCategory(ctx context.Context, id string) error {
	return s.categories.Delete(ctx, id)
}

// --- Transactions ---

type CreateTransactionInput struct {
	AccountID   string  `json:"account_id"`
	OccurredAt  string  `json:"occurred_at"` // RFC3339
	AmountMinor int64   `json:"amount_minor"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
	CategoryID  *string `json:"category_id,omitempty"`
}

func (s *Service) CreateTransaction(ctx context.Context, in CreateTransactionInput) (*domain.Transaction, error) {
	if in.AccountID == "" {
		return nil, fmt.Errorf("%w: account_id is required", domain.ErrValidation)
	}
	if _, err := s.accounts.GetByID(ctx, in.AccountID); err != nil {
		return nil, fmt.Errorf("%w: account not found", domain.ErrForeignKey)
	}
	occ, err := time.Parse(time.RFC3339, in.OccurredAt)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid occurred_at", domain.ErrValidation)
	}
	if strings.TrimSpace(in.Currency) == "" || len(in.Currency) != 3 {
		return nil, fmt.Errorf("%w: currency must be 3-letter ISO code", domain.ErrValidation)
	}
	if in.CategoryID != nil && *in.CategoryID != "" {
		if _, err := s.categories.GetByID(ctx, *in.CategoryID); err != nil {
			return nil, fmt.Errorf("%w: category not found", domain.ErrForeignKey)
		}
	}
	now := s.clock()
	t := &domain.Transaction{
		ID:          uuid.NewString(),
		AccountID:   in.AccountID,
		OccurredAt:  occ.UTC(),
		AmountMinor: in.AmountMinor,
		Currency:    strings.ToUpper(strings.TrimSpace(in.Currency)),
		Description: in.Description,
		CategoryID:  in.CategoryID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.transactions.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) ListTransactions(ctx context.Context, f TransactionFilter) ([]domain.Transaction, int, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Limit > 500 {
		f.Limit = 500
	}
	return s.transactions.List(ctx, f)
}

func (s *Service) GetTransaction(ctx context.Context, id string) (*domain.Transaction, error) {
	return s.transactions.GetByID(ctx, id)
}

type UpdateTransactionInput struct {
	OccurredAt  *string `json:"occurred_at,omitempty"`
	AmountMinor *int64  `json:"amount_minor,omitempty"`
	Currency    *string `json:"currency,omitempty"`
	Description *string `json:"description,omitempty"`
	CategoryID  *string `json:"category_id,omitempty"`
	AccountID   *string `json:"account_id,omitempty"`
}

func (s *Service) UpdateTransaction(ctx context.Context, id string, in UpdateTransactionInput) (*domain.Transaction, error) {
	t, err := s.transactions.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.OccurredAt != nil {
		occ, err := time.Parse(time.RFC3339, *in.OccurredAt)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid occurred_at", domain.ErrValidation)
		}
		t.OccurredAt = occ.UTC()
	}
	if in.AmountMinor != nil {
		t.AmountMinor = *in.AmountMinor
	}
	if in.Currency != nil {
		c := strings.ToUpper(strings.TrimSpace(*in.Currency))
		if len(c) != 3 {
			return nil, fmt.Errorf("%w: currency must be 3-letter ISO code", domain.ErrValidation)
		}
		t.Currency = c
	}
	if in.Description != nil {
		t.Description = *in.Description
	}
	if in.CategoryID != nil {
		t.CategoryID = in.CategoryID
	}
	if in.AccountID != nil {
		if _, err := s.accounts.GetByID(ctx, *in.AccountID); err != nil {
			return nil, fmt.Errorf("%w: account not found", domain.ErrForeignKey)
		}
		t.AccountID = *in.AccountID
	}
	t.UpdatedAt = s.clock()
	if err := s.transactions.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) DeleteTransaction(ctx context.Context, id string) error {
	return s.transactions.Delete(ctx, id)
}

// --- Splits ---

type SplitInput struct {
	CategoryID  string `json:"category_id"`
	AmountMinor int64  `json:"amount_minor"`
}

func (s *Service) ReplaceSplits(ctx context.Context, txID string, inputs []SplitInput) ([]domain.Split, error) {
	tx, err := s.transactions.GetByID(ctx, txID)
	if err != nil {
		return nil, err
	}
	now := s.clock()
	splits := make([]domain.Split, len(inputs))
	for i, in := range inputs {
		if in.CategoryID == "" {
			return nil, fmt.Errorf("%w: split category_id is required", domain.ErrValidation)
		}
		splits[i] = domain.Split{
			ID:            uuid.NewString(),
			TransactionID: txID,
			CategoryID:    in.CategoryID,
			AmountMinor:   in.AmountMinor,
			CreatedAt:     now,
		}
	}
	if err := domain.ValidateSplits(tx.AmountMinor, splits); err != nil {
		return nil, err
	}
	if err := s.splits.ReplaceSplits(ctx, txID, splits); err != nil {
		return nil, err
	}
	return splits, nil
}

func (s *Service) GetSplits(ctx context.Context, txID string) ([]domain.Split, error) {
	return s.splits.GetByTransactionID(ctx, txID)
}
