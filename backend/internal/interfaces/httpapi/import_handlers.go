package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/katalabut/pocket-ledger/backend/internal/application/importer"
)

func (a *API) handleImportProfiles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		list, err := a.importSvc.ListProfiles(r.Context())
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, list)
	case http.MethodPost:
		var in importer.CreateProfileInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeBadRequest(w, "invalid json")
			return
		}
		p, err := a.importSvc.CreateProfile(r.Context(), in)
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, p)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *API) handleImportUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		writeBadRequest(w, "invalid form data")
		return
	}
	profileID := r.FormValue("profile_id")
	if profileID == "" {
		writeBadRequest(w, "profile_id is required")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeBadRequest(w, "file is required")
		return
	}
	defer file.Close()

	result, err := a.importSvc.Upload(r.Context(), profileID, header.Filename, file)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result.Import)
}

func (a *API) handleImportByID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	id := extractID(path, "/api/imports/")

	if strings.HasSuffix(id, "/preview") {
		importID := strings.TrimSuffix(id, "/preview")
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		preview, err := a.importSvc.Preview(r.Context(), importID)
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, preview)
		return
	}

	if strings.HasSuffix(id, "/commit") {
		importID := strings.TrimSuffix(id, "/commit")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		imp, err := a.importSvc.Commit(r.Context(), importID)
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, imp)
		return
	}
}
