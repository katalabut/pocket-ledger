package reports

import (
	"context"
	"time"

	"github.com/katalabut/pocket-ledger/backend/internal/application/fx"
	"github.com/katalabut/pocket-ledger/backend/internal/application/ledger"
	"github.com/katalabut/pocket-ledger/backend/internal/domain"
)

type Service struct {
	ledger *ledger.Service
	fx     *fx.Service
}

func NewService(ledger *ledger.Service, fx *fx.Service) *Service {
	return &Service{ledger: ledger, fx: fx}
}

type SpendingRow struct {
	CategoryID   string `json:"CategoryID"`
	CategoryName string `json:"CategoryName"`
	TotalMinor   int64  `json:"TotalMinor"`
}

func (s *Service) SpendingByCategory(ctx context.Context, from, to string) ([]SpendingRow, error) {
	// Get all transactions in range
	f := ledger.TransactionFilter{
		From:  ptrStr(from),
		To:    ptrStr(to),
		Limit: 10000,
	}
	txs, _, err := s.ledger.ListTransactions(ctx, f)
	if err != nil {
		return nil, err
	}

	// Get categories for names
	cats, err := s.ledger.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	catMap := map[string]string{}
	for _, c := range cats {
		catMap[c.ID] = c.Name
	}

	// Collect transaction IDs to batch-fetch splits
	var txIDs []string
	txMap := map[string]*domain.Transaction{}
	for i := range txs {
		txIDs = append(txIDs, txs[i].ID)
		txMap[txs[i].ID] = &txs[i]
	}

	// Aggregate spending per category
	totals := map[string]int64{}

	for _, tx := range txs {
		date := tx.OccurredAt.Format("2006-01-02")

		// Check if transaction has splits
		splits, _ := s.ledger.GetSplits(ctx, tx.ID)
		if len(splits) > 0 {
			for _, sp := range splits {
				amountBase, err := s.fx.Convert(ctx, sp.AmountMinor, tx.Currency, date)
				if err != nil {
					amountBase = sp.AmountMinor // fallback to original
				}
				totals[sp.CategoryID] += amountBase
			}
		} else if tx.CategoryID != nil && *tx.CategoryID != "" {
			amountBase, err := s.fx.Convert(ctx, tx.AmountMinor, tx.Currency, date)
			if err != nil {
				amountBase = tx.AmountMinor
			}
			totals[*tx.CategoryID] += amountBase
		} else {
			amountBase, err := s.fx.Convert(ctx, tx.AmountMinor, tx.Currency, date)
			if err != nil {
				amountBase = tx.AmountMinor
			}
			totals["uncategorized"] += amountBase
		}
	}

	var result []SpendingRow
	for catID, total := range totals {
		name := catMap[catID]
		if name == "" {
			name = "Uncategorized"
		}
		result = append(result, SpendingRow{CategoryID: catID, CategoryName: name, TotalMinor: total})
	}
	return result, nil
}

type AccountBalanceRow struct {
	AccountID       string `json:"AccountID"`
	AccountName     string `json:"AccountName"`
	Currency        string `json:"Currency"`
	BalanceMinor    int64  `json:"BalanceMinor"`
	BalanceBaseMinor int64 `json:"BalanceBaseMinor"`
}

func (s *Service) AccountBalances(ctx context.Context) ([]AccountBalanceRow, error) {
	accounts, err := s.ledger.ListAccounts(ctx)
	if err != nil {
		return nil, err
	}

	today := time.Now().UTC().Format("2006-01-02")
	var result []AccountBalanceRow
	for _, acc := range accounts {
		f := ledger.TransactionFilter{
			AccountID: &acc.ID,
			Limit:     100000,
		}
		txs, _, err := s.ledger.ListTransactions(ctx, f)
		if err != nil {
			return nil, err
		}
		balance := acc.InitialBalanceMinor
		for _, tx := range txs {
			balance += tx.AmountMinor
		}

		balanceBase, err := s.fx.Convert(ctx, balance, acc.Currency, today)
		if err != nil {
			balanceBase = balance // fallback for base currency
		}

		result = append(result, AccountBalanceRow{
			AccountID:       acc.ID,
			AccountName:     acc.Name,
			Currency:        acc.Currency,
			BalanceMinor:    balance,
			BalanceBaseMinor: balanceBase,
		})
	}
	return result, nil
}

func ptrStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
