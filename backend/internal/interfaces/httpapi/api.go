package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/katalabut/pocket-ledger/backend/internal/application/auth"
	"github.com/katalabut/pocket-ledger/backend/internal/application/budget"
	"github.com/katalabut/pocket-ledger/backend/internal/application/fx"
	"github.com/katalabut/pocket-ledger/backend/internal/application/importer"
	"github.com/katalabut/pocket-ledger/backend/internal/application/ledger"
	"github.com/katalabut/pocket-ledger/backend/internal/application/reports"
	"github.com/katalabut/pocket-ledger/backend/internal/domain"
	"github.com/katalabut/pocket-ledger/backend/internal/interfaces/httpauth"
)

type API struct {
	authSvc    *auth.Service
	ledgerSvc  *ledger.Service
	importSvc  *importer.Service
	fxSvc      *fx.Service
	reportsSvc *reports.Service
	budgetSvc  *budget.Service
	secret     string
	issuer     string
	clock      func() time.Time
	logger     *log.Logger
}

type Config struct {
	JWTSecret string
	Issuer    string
}

func New(authSvc *auth.Service, ledgerSvc *ledger.Service, importSvc *importer.Service, fxSvc *fx.Service, reportsSvc *reports.Service, budgetSvc *budget.Service, cfg Config) *API {
	return &API{
		authSvc: authSvc, ledgerSvc: ledgerSvc, importSvc: importSvc,
		fxSvc: fxSvc, reportsSvc: reportsSvc, budgetSvc: budgetSvc,
		secret: cfg.JWTSecret, issuer: cfg.Issuer,
		clock:  func() time.Time { return time.Now().UTC() },
		logger: log.Default(),
	}
}

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})

	mux.HandleFunc("/auth/request_code", a.handleRequestCode)
	mux.HandleFunc("/auth/confirm_code", a.handleConfirmCode)

	protected := http.NewServeMux()
	protected.HandleFunc("/api/me", func(w http.ResponseWriter, r *http.Request) {
		cl, _ := httpauth.ClaimsFromContext(r.Context())
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"user_id": cl.UserID, "email": cl.Email})
	})

	// Accounts
	protected.HandleFunc("/api/accounts", a.handleAccounts)
	protected.HandleFunc("/api/accounts/", a.handleAccountByID)

	// Categories
	protected.HandleFunc("/api/categories", a.handleCategories)
	protected.HandleFunc("/api/categories/", a.handleCategoryByID)

	// Transactions
	protected.HandleFunc("/api/transactions", a.handleTransactions)
	protected.HandleFunc("/api/transactions/", a.handleTransactionByID)

	// Import profiles + sessions
	protected.HandleFunc("/api/import-profiles", a.handleImportProfiles)
	protected.HandleFunc("/api/imports/upload", a.handleImportUpload)
	protected.HandleFunc("/api/imports/", a.handleImportByID)

	// FX
	protected.HandleFunc("/api/fx/rates", a.handleFXRates)
	protected.HandleFunc("/api/fx/sync", a.handleFXSync)

	// Reports
	protected.HandleFunc("/api/reports/spending", a.handleReportSpending)
	protected.HandleFunc("/api/reports/balances", a.handleReportBalances)

	// Budgets
	protected.HandleFunc("/api/budgets", a.handleBudgets)

	mux.Handle("/api/", httpauth.Middleware(a.secret, a.issuer, protected))
	return a.requestLoggingMiddleware(corsMiddleware(mux))
}

// --- Auth handlers ---

type requestCodeReq struct {
	Email string `json:"email"`
}

type confirmCodeReq struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func (a *API) handleRequestCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req requestCodeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	out, err := a.authSvc.RequestCode(r.Context(), req.Email)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (a *API) handleConfirmCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req confirmCodeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	out, err := a.authSvc.ConfirmCode(r.Context(), req.Email, req.Code)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(map[string]any{"token": out.Token})
}

// --- Accounts ---

func (a *API) handleAccounts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		list, err := a.ledgerSvc.ListAccounts(r.Context())
		if err != nil {
			a.writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, list)
	case http.MethodPost:
		var in ledger.CreateAccountInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeBadRequest(w, "invalid json")
			return
		}
		acc, err := a.ledgerSvc.CreateAccount(r.Context(), in)
		if err != nil {
			a.writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, acc)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *API) handleAccountByID(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/accounts/")
	if id == "" {
		writeBadRequest(w, "missing id")
		return
	}
	switch r.Method {
	case http.MethodGet:
		acc, err := a.ledgerSvc.GetAccount(r.Context(), id)
		if err != nil {
			a.writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, acc)
	case http.MethodPatch:
		var in ledger.UpdateAccountInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeBadRequest(w, "invalid json")
			return
		}
		acc, err := a.ledgerSvc.UpdateAccount(r.Context(), id, in)
		if err != nil {
			a.writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, acc)
	case http.MethodDelete:
		if err := a.ledgerSvc.DeleteAccount(r.Context(), id); err != nil {
			a.writeErr(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// --- Categories ---

func (a *API) handleCategories(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		list, err := a.ledgerSvc.ListCategories(r.Context())
		if err != nil {
			a.writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, list)
	case http.MethodPost:
		var in ledger.CreateCategoryInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeBadRequest(w, "invalid json")
			return
		}
		cat, err := a.ledgerSvc.CreateCategory(r.Context(), in)
		if err != nil {
			a.writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, cat)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *API) handleCategoryByID(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/categories/")
	if id == "" {
		writeBadRequest(w, "missing id")
		return
	}
	switch r.Method {
	case http.MethodGet:
		cat, err := a.ledgerSvc.GetCategory(r.Context(), id)
		if err != nil {
			a.writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, cat)
	case http.MethodPatch:
		var in ledger.UpdateCategoryInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeBadRequest(w, "invalid json")
			return
		}
		cat, err := a.ledgerSvc.UpdateCategory(r.Context(), id, in)
		if err != nil {
			a.writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, cat)
	case http.MethodDelete:
		if err := a.ledgerSvc.DeleteCategory(r.Context(), id); err != nil {
			a.writeErr(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// --- Transactions ---

func (a *API) handleTransactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		q := r.URL.Query()
		f := ledger.TransactionFilter{
			AccountID:  ptrIfSet(q.Get("account_id")),
			CategoryID: ptrIfSet(q.Get("category_id")),
			From:       ptrIfSet(q.Get("from")),
			To:         ptrIfSet(q.Get("to")),
			Query:      ptrIfSet(q.Get("q")),
		}
		if v := q.Get("limit"); v != "" {
			f.Limit, _ = strconv.Atoi(v)
		}
		if v := q.Get("offset"); v != "" {
			f.Offset, _ = strconv.Atoi(v)
		}
		items, total, err := a.ledgerSvc.ListTransactions(r.Context(), f)
		if err != nil {
			a.writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
	case http.MethodPost:
		var in ledger.CreateTransactionInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeBadRequest(w, "invalid json")
			return
		}
		tx, err := a.ledgerSvc.CreateTransaction(r.Context(), in)
		if err != nil {
			a.writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, tx)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *API) handleTransactionByID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	id := extractID(path, "/api/transactions/")

	// Check for /api/transactions/{id}/splits
	if strings.Contains(id, "/splits") {
		txID := strings.TrimSuffix(id, "/splits")
		a.handleSplits(w, r, txID)
		return
	}

	if id == "" {
		writeBadRequest(w, "missing id")
		return
	}
	switch r.Method {
	case http.MethodGet:
		tx, err := a.ledgerSvc.GetTransaction(r.Context(), id)
		if err != nil {
			a.writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, tx)
	case http.MethodPatch:
		var in ledger.UpdateTransactionInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeBadRequest(w, "invalid json")
			return
		}
		tx, err := a.ledgerSvc.UpdateTransaction(r.Context(), id, in)
		if err != nil {
			a.writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, tx)
	case http.MethodDelete:
		if err := a.ledgerSvc.DeleteTransaction(r.Context(), id); err != nil {
			a.writeErr(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// --- Splits ---

func (a *API) handleSplits(w http.ResponseWriter, r *http.Request, txID string) {
	switch r.Method {
	case http.MethodGet:
		splits, err := a.ledgerSvc.GetSplits(r.Context(), txID)
		if err != nil {
			a.writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, splits)
	case http.MethodPost:
		var inputs []ledger.SplitInput
		if err := json.NewDecoder(r.Body).Decode(&inputs); err != nil {
			writeBadRequest(w, "invalid json")
			return
		}
		splits, err := a.ledgerSvc.ReplaceSplits(r.Context(), txID, inputs)
		if err != nil {
			a.writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, splits)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// --- helpers ---

func extractID(path, prefix string) string {
	s := strings.TrimPrefix(path, prefix)
	s = strings.TrimRight(s, "/")
	return s
}

func ptrIfSet(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeBadRequest(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func (a *API) writeErr(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	if errors.Is(err, domain.ErrNotFound) {
		code = http.StatusNotFound
	} else if errors.Is(err, domain.ErrValidation) {
		code = http.StatusBadRequest
	} else if errors.Is(err, domain.ErrConflict) {
		code = http.StatusConflict
	} else if errors.Is(err, domain.ErrSplitMismatch) {
		code = http.StatusBadRequest
	} else if errors.Is(err, domain.ErrForeignKey) {
		code = http.StatusBadRequest
	}

	if a.logger != nil {
		a.logger.Printf("http error: status=%d err=%v", code, err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *API) requestLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		rw := &statusResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		if a.logger != nil {
			a.logger.Printf("http request: method=%s path=%s status=%d duration=%s remote=%s", r.Method, r.URL.Path, rw.status, time.Since(startedAt).Round(time.Millisecond), r.RemoteAddr)
		}
	})
}

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
