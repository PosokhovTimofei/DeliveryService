package utils

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
)

var tmpl *template.Template

func init() {
	var err error
	tmpl, err = template.ParseFiles("./gateway/internal/templates/template.html")
	if err != nil {
		panic("failed to parse HTML template: " + err.Error())
	}
}

func RespondJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	w.WriteHeader(code)

	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "text/html") {
		jsonData, _ := json.MarshalIndent(payload, "", "  ")
		tmpl.Execute(w, string(jsonData))
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}
}

func RespondError(w http.ResponseWriter, r *http.Request, code int, message string) {
	RespondJSON(w, r, code, map[string]string{"error": message})
}
