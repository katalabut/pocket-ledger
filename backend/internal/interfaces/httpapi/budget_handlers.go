package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/katalabut/pocket-ledger/backend/internal/application/budget"
)

func (a *API) handleBudgets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		month := r.URL.Query().Get("month")
		if month == "" {
			writeBadRequest(w, "month query parameter required (YYYY-MM)")
			return
		}
		report, err := a.budgetSvc.GetReport(r.Context(), month)
		if err != nil {
			a.writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, report)
	case http.MethodPost:
		var in budget.UpsertInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeBadRequest(w, "invalid json")
			return
		}
		b, err := a.budgetSvc.Upsert(r.Context(), in)
		if err != nil {
			a.writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, b)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
