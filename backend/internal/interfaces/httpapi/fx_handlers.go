package httpapi

import (
	"net/http"
)

func (a *API) handleFXRates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	date := r.URL.Query().Get("date")
	if date == "" {
		date = a.clock().Format("2006-01-02")
	}
	rates, err := a.fxSvc.GetRatesForDate(r.Context(), date)
	if err != nil {
		writeErr(w, err)
		return
	}
	if rates == nil {
		rates = map[string]float64{}
	}
	writeJSON(w, http.StatusOK, rates)
}

func (a *API) handleFXSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	count, err := a.fxSvc.SyncRates(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"synced": count})
}

func (a *API) handleReportSpending(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	rows, err := a.reportsSvc.SpendingByCategory(r.Context(), from, to)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func (a *API) handleReportBalances(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	rows, err := a.reportsSvc.AccountBalances(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}
