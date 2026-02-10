package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/katalabut/pocket-ledger/backend/internal/application/auth"
	"github.com/katalabut/pocket-ledger/backend/internal/interfaces/httpauth"
)

type API struct {
	authSvc *auth.Service
	secret  string
	issuer  string
}

type Config struct {
	JWTSecret string
	Issuer    string
}

func New(authSvc *auth.Service, cfg Config) *API {
	return &API{authSvc: authSvc, secret: cfg.JWTSecret, issuer: cfg.Issuer}
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

	mux.Handle("/api/", httpauth.Middleware(a.secret, a.issuer, protected))
	return mux
}

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
