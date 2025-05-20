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

type templateData struct {
	Content interface{}
}

func RespondJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	respond(w, r, code, payload)
}

func RespondError(w http.ResponseWriter, r *http.Request, code int, message string) {
	payload := map[string]string{"error": message}
	respond(w, r, code, payload)
}

func respond(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	w.WriteHeader(code)

	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "text/html") {
		var content string
		switch v := payload.(type) {
		case string:
			content = v
		default:
			jsonData, _ := json.MarshalIndent(payload, "", "  ")
			content = string(jsonData)
		}
		tmpl.Execute(w, templateData{Content: content})
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}
}
