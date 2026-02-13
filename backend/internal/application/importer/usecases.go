package importer

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/katalabut/pocket-ledger/backend/internal/domain"
	"github.com/katalabut/pocket-ledger/backend/internal/platform/cryptoutil"
)

type Service struct {
	profiles ImportProfileRepository
	imports  ImportRepository
	txRepo   TransactionCreator
	clock    func() time.Time
}

func NewService(
	profiles ImportProfileRepository,
	imports ImportRepository,
	txRepo TransactionCreator,
	clock func() time.Time,
) *Service {
	if clock == nil {
		clock = func() time.Time { return time.Now().UTC() }
	}
	return &Service{profiles: profiles, imports: imports, txRepo: txRepo, clock: clock}
}

// --- Profiles ---

type CreateProfileInput struct {
	Name           string         `json:"name"`
	AccountID      string         `json:"account_id"`
	Separator      string         `json:"separator"`
	DateFormat     string         `json:"date_format"`
	ColumnMapping  map[string]int `json:"column_mapping"`
	AmountSignFlip bool           `json:"amount_sign_flip"`
	SkipHeaderRows int            `json:"skip_header_rows"`
}

func (s *Service) CreateProfile(ctx context.Context, in CreateProfileInput) (*domain.ImportProfile, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, fmt.Errorf("%w: name is required", domain.ErrValidation)
	}
	if in.AccountID == "" {
		return nil, fmt.Errorf("%w: account_id is required", domain.ErrValidation)
	}
	if in.Separator == "" {
		in.Separator = ","
	}
	if in.DateFormat == "" {
		in.DateFormat = "2006-01-02"
	}
	now := s.clock()
	p := &domain.ImportProfile{
		ID:             uuid.NewString(),
		Name:           strings.TrimSpace(in.Name),
		AccountID:      in.AccountID,
		Separator:       in.Separator,
		DateFormat:     in.DateFormat,
		ColumnMapping:  in.ColumnMapping,
		AmountSignFlip: in.AmountSignFlip,
		SkipHeaderRows: in.SkipHeaderRows,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.profiles.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) ListProfiles(ctx context.Context) ([]domain.ImportProfile, error) {
	return s.profiles.List(ctx)
}

// --- Import sessions ---

type UploadResult struct {
	Import *domain.Import
}

func (s *Service) Upload(ctx context.Context, profileID string, filename string, csvReader io.Reader) (*UploadResult, error) {
	profile, err := s.profiles.GetByID(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("profile: %w", err)
	}

	now := s.clock()
	imp := &domain.Import{
		ID:        uuid.NewString(),
		ProfileID: profileID,
		AccountID: profile.AccountID,
		Filename:  filename,
		Status:    "pending",
		CreatedAt: now,
	}
	if err := s.imports.CreateImport(ctx, imp); err != nil {
		return nil, err
	}

	// Parse CSV
	reader := csv.NewReader(csvReader)
	if len(profile.Separator) > 0 {
		reader.Comma = rune(profile.Separator[0])
	}
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	var rows []domain.ImportRow
	rowNum := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		rowNum++
		if rowNum <= profile.SkipHeaderRows {
			continue
		}
		rawData := strings.Join(record, string(reader.Comma))
		row := domain.ImportRow{
			ID:        uuid.NewString(),
			ImportID:  imp.ID,
			RowNumber: rowNum,
			RawData:   rawData,
			Status:    "pending",
			CreatedAt: now,
		}

		// Parse fields and generate dedupe key
		parsed, parseErr := parseRow(record, profile)
		if parseErr != nil {
			errMsg := parseErr.Error()
			row.ErrorMessage = &errMsg
			row.Status = "error"
		} else {
			dk := computeDedupeKey(profile.AccountID, parsed)
			row.DedupeKey = &dk
		}
		rows = append(rows, row)

		if err != nil {
			errMsg := err.Error()
			row.ErrorMessage = &errMsg
			row.Status = "error"
		}
	}

	imp.TotalRows = len(rows)
	imp.ErrorRows = countStatus(rows, "error")
	imp.Status = "previewed"
	if err := s.imports.CreateRows(ctx, rows); err != nil {
		return nil, err
	}
	if err := s.imports.UpdateImport(ctx, imp); err != nil {
		return nil, err
	}

	return &UploadResult{Import: imp}, nil
}

type PreviewResult struct {
	Import *domain.Import      `json:"import"`
	Rows   []domain.ImportRow  `json:"rows"`
}

func (s *Service) Preview(ctx context.Context, importID string) (*PreviewResult, error) {
	imp, err := s.imports.GetImport(ctx, importID)
	if err != nil {
		return nil, err
	}
	rows, err := s.imports.GetRows(ctx, importID)
	if err != nil {
		return nil, err
	}
	return &PreviewResult{Import: imp, Rows: rows}, nil
}

func (s *Service) Commit(ctx context.Context, importID string) (*domain.Import, error) {
	imp, err := s.imports.GetImport(ctx, importID)
	if err != nil {
		return nil, err
	}
	if imp.Status == "committed" {
		return imp, nil
	}
	if imp.Status != "previewed" {
		return nil, fmt.Errorf("%w: import must be previewed before commit", domain.ErrValidation)
	}

	profile, err := s.profiles.GetByID(ctx, imp.ProfileID)
	if err != nil {
		return nil, err
	}

	rows, err := s.imports.GetRows(ctx, importID)
	if err != nil {
		return nil, err
	}

	now := s.clock()
	imported, skipped, errors := 0, 0, 0

	for i, row := range rows {
		if row.Status == "error" {
			errors++
			continue
		}

		// Check dedupe
		if row.DedupeKey != nil {
			exists, err := s.txRepo.ExistsByDedupeKey(ctx, imp.AccountID, *row.DedupeKey)
			if err != nil {
				errMsg := err.Error()
				rows[i].ErrorMessage = &errMsg
				rows[i].Status = "error"
				errors++
				_ = s.imports.UpdateRow(ctx, &rows[i])
				continue
			}
			if exists {
				rows[i].Status = "skipped"
				skipped++
				_ = s.imports.UpdateRow(ctx, &rows[i])
				continue
			}
		}

		// Parse and create transaction
		record := strings.Split(row.RawData, string(rune(profile.Separator[0])))
		parsed, parseErr := parseRow(record, profile)
		if parseErr != nil {
			errMsg := parseErr.Error()
			rows[i].ErrorMessage = &errMsg
			rows[i].Status = "error"
			errors++
			_ = s.imports.UpdateRow(ctx, &rows[i])
			continue
		}

		tx := &domain.Transaction{
			ID:          uuid.NewString(),
			AccountID:   imp.AccountID,
			OccurredAt:  parsed.date,
			AmountMinor: parsed.amountMinor,
			Currency:    parsed.currency,
			Description: parsed.description,
			DedupeKey:   row.DedupeKey,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := s.txRepo.Create(ctx, tx); err != nil {
			errMsg := err.Error()
			rows[i].ErrorMessage = &errMsg
			rows[i].Status = "error"
			errors++
			_ = s.imports.UpdateRow(ctx, &rows[i])
			continue
		}

		rows[i].Status = "imported"
		rows[i].TransactionID = &tx.ID
		imported++
		_ = s.imports.UpdateRow(ctx, &rows[i])
	}

	committedAt := now
	imp.Status = "committed"
	imp.ImportedRows = imported
	imp.SkippedRows = skipped
	imp.ErrorRows = errors
	imp.CommittedAt = &committedAt
	if err := s.imports.UpdateImport(ctx, imp); err != nil {
		return nil, err
	}

	return imp, nil
}

// --- parsing ---

type parsedRow struct {
	date        time.Time
	amountMinor int64
	currency    string
	description string
	externalID  string
}

func parseRow(record []string, profile *domain.ImportProfile) (*parsedRow, error) {
	p := &parsedRow{}
	get := func(field string) string {
		idx, ok := profile.ColumnMapping[field]
		if !ok || idx < 0 || idx >= len(record) {
			return ""
		}
		return strings.TrimSpace(record[idx])
	}

	// Date
	dateStr := get("date")
	if dateStr == "" {
		return nil, fmt.Errorf("missing date")
	}
	d, err := time.Parse(profile.DateFormat, dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date %q: %w", dateStr, err)
	}
	p.date = d.UTC()

	// Amount
	amountStr := get("amount")
	if amountStr == "" {
		return nil, fmt.Errorf("missing amount")
	}
	// Clean up amount string: handle commas as decimals, remove thousands sep
	amountStr = strings.ReplaceAll(amountStr, " ", "")
	if strings.Contains(amountStr, ",") && strings.Contains(amountStr, ".") {
		amountStr = strings.ReplaceAll(amountStr, ",", "") // 1,234.56 → 1234.56
	} else if strings.Contains(amountStr, ",") {
		amountStr = strings.ReplaceAll(amountStr, ",", ".") // 1234,56 → 1234.56
	}
	amountFloat, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid amount %q: %w", amountStr, err)
	}
	if profile.AmountSignFlip {
		amountFloat = -amountFloat
	}
	p.amountMinor = int64(math.Round(amountFloat * 100))

	// Currency
	p.currency = strings.ToUpper(get("currency"))
	if p.currency == "" {
		p.currency = "EUR" // default
	}

	// Description
	p.description = get("description")

	// External ID
	p.externalID = get("external_id")

	return p, nil
}

func computeDedupeKey(accountID string, p *parsedRow) string {
	raw := fmt.Sprintf("%s|%s|%d|%s|%s|%s",
		accountID, p.date.Format("2006-01-02"), p.amountMinor, p.currency, p.description, p.externalID)
	return cryptoutil.SHA256Hex(raw)
}

func countStatus(rows []domain.ImportRow, status string) int {
	n := 0
	for _, r := range rows {
		if r.Status == status {
			n++
		}
	}
	return n
}
