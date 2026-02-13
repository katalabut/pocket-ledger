package importer

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/katalabut/pocket-ledger/backend/internal/domain"
)

var fixedTime = time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)

func clock() time.Time { return fixedTime }

func newTestService() (*Service, *memTxRepo) {
	txRepo := &memTxRepo{items: map[string]*domain.Transaction{}, dedupes: map[string]bool{}}
	profiles := &memProfileRepo{items: map[string]*domain.ImportProfile{}}
	imports := &memImportRepo{imports: map[string]*domain.Import{}, rows: map[string][]domain.ImportRow{}}
	svc := NewService(profiles, imports, txRepo, clock)
	return svc, txRepo
}

// --- in-memory repos ---

type memProfileRepo struct{ items map[string]*domain.ImportProfile }

func (r *memProfileRepo) Create(_ context.Context, p *domain.ImportProfile) error {
	r.items[p.ID] = p; return nil
}
func (r *memProfileRepo) GetByID(_ context.Context, id string) (*domain.ImportProfile, error) {
	p, ok := r.items[id]; if !ok { return nil, domain.ErrNotFound }; return p, nil
}
func (r *memProfileRepo) List(_ context.Context) ([]domain.ImportProfile, error) {
	out := make([]domain.ImportProfile, 0, len(r.items))
	for _, p := range r.items { out = append(out, *p) }; return out, nil
}

type memImportRepo struct {
	imports map[string]*domain.Import
	rows    map[string][]domain.ImportRow
}

func (r *memImportRepo) CreateImport(_ context.Context, imp *domain.Import) error {
	r.imports[imp.ID] = imp; return nil
}
func (r *memImportRepo) GetImport(_ context.Context, id string) (*domain.Import, error) {
	imp, ok := r.imports[id]; if !ok { return nil, domain.ErrNotFound }; return imp, nil
}
func (r *memImportRepo) UpdateImport(_ context.Context, imp *domain.Import) error {
	r.imports[imp.ID] = imp; return nil
}
func (r *memImportRepo) CreateRows(_ context.Context, rows []domain.ImportRow) error {
	if len(rows) == 0 { return nil }
	r.rows[rows[0].ImportID] = rows; return nil
}
func (r *memImportRepo) GetRows(_ context.Context, importID string) ([]domain.ImportRow, error) {
	return r.rows[importID], nil
}
func (r *memImportRepo) UpdateRow(_ context.Context, row *domain.ImportRow) error {
	rows := r.rows[row.ImportID]
	for i, old := range rows {
		if old.ID == row.ID { rows[i] = *row; break }
	}
	return nil
}

type memTxRepo struct {
	items   map[string]*domain.Transaction
	dedupes map[string]bool
}

func (r *memTxRepo) Create(_ context.Context, t *domain.Transaction) error {
	r.items[t.ID] = t
	if t.DedupeKey != nil { r.dedupes[t.AccountID+"|"+*t.DedupeKey] = true }
	return nil
}
func (r *memTxRepo) ExistsByDedupeKey(_ context.Context, accountID, dedupeKey string) (bool, error) {
	return r.dedupes[accountID+"|"+dedupeKey], nil
}

// --- tests ---

func TestImportEndToEnd(t *testing.T) {
	svc, txRepo := newTestService()
	ctx := context.Background()

	// Create profile
	profile, err := svc.CreateProfile(ctx, CreateProfileInput{
		Name:           "Test CSV",
		AccountID:      "acc-1",
		Separator:      ",",
		DateFormat:     "2006-01-02",
		ColumnMapping:  map[string]int{"date": 0, "amount": 1, "currency": 2, "description": 3},
		SkipHeaderRows: 1,
	})
	if err != nil {
		t.Fatalf("create profile: %v", err)
	}

	csvData := "date,amount,currency,description\n2025-01-15,-12.50,EUR,Grocery store\n2025-01-16,-5.00,EUR,Coffee\n"

	// Upload
	result, err := svc.Upload(ctx, profile.ID, "test.csv", strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if result.Import.TotalRows != 2 {
		t.Fatalf("expected 2 rows, got %d", result.Import.TotalRows)
	}
	if result.Import.Status != "previewed" {
		t.Fatalf("expected previewed, got %s", result.Import.Status)
	}

	// Preview
	preview, err := svc.Preview(ctx, result.Import.ID)
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	if len(preview.Rows) != 2 {
		t.Fatalf("expected 2 preview rows, got %d", len(preview.Rows))
	}

	// Commit
	imp, err := svc.Commit(ctx, result.Import.ID)
	if err != nil {
		t.Fatalf("commit: %v", err)
	}
	if imp.ImportedRows != 2 {
		t.Fatalf("expected 2 imported, got %d", imp.ImportedRows)
	}
	if len(txRepo.items) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(txRepo.items))
	}

	// Verify dedup: re-upload same file
	result2, err := svc.Upload(ctx, profile.ID, "test.csv", strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("re-upload: %v", err)
	}
	imp2, err := svc.Commit(ctx, result2.Import.ID)
	if err != nil {
		t.Fatalf("re-commit: %v", err)
	}
	if imp2.ImportedRows != 0 {
		t.Fatalf("expected 0 imported on re-import, got %d", imp2.ImportedRows)
	}
	if imp2.SkippedRows != 2 {
		t.Fatalf("expected 2 skipped on re-import, got %d", imp2.SkippedRows)
	}
}

func TestImportWithErrors(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	profile, _ := svc.CreateProfile(ctx, CreateProfileInput{
		Name: "Err test", AccountID: "acc-1", Separator: ",",
		DateFormat: "2006-01-02",
		ColumnMapping: map[string]int{"date": 0, "amount": 1, "currency": 2, "description": 3},
		SkipHeaderRows: 1,
	})

	csvData := "date,amount,currency,description\nbad-date,-12.50,EUR,Grocery\n2025-01-16,-5.00,EUR,Coffee\n"
	result, err := svc.Upload(ctx, profile.ID, "errors.csv", strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if result.Import.ErrorRows != 1 {
		t.Fatalf("expected 1 error row, got %d", result.Import.ErrorRows)
	}

	imp, err := svc.Commit(ctx, result.Import.ID)
	if err != nil {
		t.Fatalf("commit: %v", err)
	}
	if imp.ImportedRows != 1 {
		t.Fatalf("expected 1 imported, got %d", imp.ImportedRows)
	}
	if imp.ErrorRows != 1 {
		t.Fatalf("expected 1 error, got %d", imp.ErrorRows)
	}
}

func TestDedupeKeyDeterministic(t *testing.T) {
	p := &parsedRow{
		date: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		amountMinor: -1250, currency: "EUR", description: "Grocery store",
	}
	k1 := computeDedupeKey("acc-1", p)
	k2 := computeDedupeKey("acc-1", p)
	if k1 != k2 {
		t.Fatalf("dedupe keys not deterministic: %s != %s", k1, k2)
	}
	// Different account produces different key
	k3 := computeDedupeKey("acc-2", p)
	if k1 == k3 {
		t.Fatal("different account should produce different key")
	}
}
