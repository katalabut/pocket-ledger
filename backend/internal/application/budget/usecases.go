package budget

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/katalabut/pocket-ledger/backend/internal/application/fx"
	"github.com/katalabut/pocket-ledger/backend/internal/application/ledger"
	"github.com/katalabut/pocket-ledger/backend/internal/domain"
)

type Service struct {
	budgets  BudgetRepository
	ledger   *ledger.Service
	fx       *fx.Service
	clock    func() time.Time
}

func NewService(budgets BudgetRepository, ledger *ledger.Service, fx *fx.Service, clock func() time.Time) *Service {
	if clock == nil {
		clock = func() time.Time { return time.Now().UTC() }
	}
	return &Service{budgets: budgets, ledger: ledger, fx: fx, clock: clock}
}

type UpsertInput struct {
	Month       string `json:"month"`       // YYYY-MM
	CategoryID  string `json:"category_id"`
	PlannedMinor int64 `json:"planned_minor"` // in base currency minor units
}

func (s *Service) Upsert(ctx context.Context, in UpsertInput) (*domain.Budget, error) {
	if strings.TrimSpace(in.Month) == "" || len(in.Month) != 7 {
		return nil, fmt.Errorf("%w: month must be YYYY-MM", domain.ErrValidation)
	}
	if in.CategoryID == "" {
		return nil, fmt.Errorf("%w: category_id is required", domain.ErrValidation)
	}
	now := s.clock()
	b := &domain.Budget{
		ID:                     uuid.NewString(),
		Month:                  in.Month,
		CategoryID:             in.CategoryID,
		PlannedAmountMinorBase: in.PlannedMinor,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
	if err := s.budgets.Upsert(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

type BudgetReport struct {
	ID             string `json:"ID"`
	Month          string `json:"Month"`
	CategoryID     string `json:"CategoryID"`
	CategoryName   string `json:"CategoryName,omitempty"`
	PlannedMinor   int64  `json:"PlannedMinor"`
	SpentMinor     int64  `json:"SpentMinor"`
	RemainingMinor int64  `json:"RemainingMinor"`
}

func (s *Service) GetReport(ctx context.Context, month string) ([]BudgetReport, error) {
	budgets, err := s.budgets.ListByMonth(ctx, month)
	if err != nil {
		return nil, err
	}

	// Category names
	cats, _ := s.ledger.ListCategories(ctx)
	catMap := map[string]string{}
	for _, c := range cats {
		catMap[c.ID] = c.Name
	}

	// Compute date range for the month
	from := month + "-01T00:00:00Z"
	// Next month
	t, _ := time.Parse("2006-01", month)
	nextMonth := t.AddDate(0, 1, 0)
	to := nextMonth.Format("2006-01") + "-01T00:00:00Z"

	// Get all transactions in the month
	f := ledger.TransactionFilter{
		From:  &from,
		To:    &to,
		Limit: 100000,
	}
	txs, _, _ := s.ledger.ListTransactions(ctx, f)

	// Compute spent per category (with splits + FX conversion)
	spent := map[string]int64{}
	for _, tx := range txs {
		date := tx.OccurredAt.Format("2006-01-02")
		splits, _ := s.ledger.GetSplits(ctx, tx.ID)
		if len(splits) > 0 {
			for _, sp := range splits {
				amountBase, err := s.fx.Convert(ctx, sp.AmountMinor, tx.Currency, date)
				if err != nil {
					amountBase = sp.AmountMinor
				}
				spent[sp.CategoryID] += amountBase
			}
		} else if tx.CategoryID != nil && *tx.CategoryID != "" {
			amountBase, err := s.fx.Convert(ctx, tx.AmountMinor, tx.Currency, date)
			if err != nil {
				amountBase = tx.AmountMinor
			}
			spent[*tx.CategoryID] += amountBase
		}
	}

	var result []BudgetReport
	for _, b := range budgets {
		s := spent[b.CategoryID]
		// Spent is typically negative (expenses), budget is positive
		// remaining = planned + spent (since spent is negative)
		remaining := b.PlannedAmountMinorBase + s
		result = append(result, BudgetReport{
			ID:             b.ID,
			Month:          b.Month,
			CategoryID:     b.CategoryID,
			CategoryName:   catMap[b.CategoryID],
			PlannedMinor:   b.PlannedAmountMinorBase,
			SpentMinor:     s,
			RemainingMinor: remaining,
		})
	}
	return result, nil
}
