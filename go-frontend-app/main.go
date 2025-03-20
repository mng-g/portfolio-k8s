package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	// Serve static assets (like CSS)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Read backend URL from environment variable, with a default value if not set
	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://localhost:9191"
	}

	// Define handler to serve the index.html template
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Determine the absolute path to the template file
		templatePath, err := filepath.Abs("templates/index.html")
		if err != nil {
			http.Error(w, "Template path error", http.StatusInternalServerError)
			return
		}
		tmpl, err := template.ParseFiles(templatePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := struct {
			BackendURL string
		}{
			BackendURL: backendURL,
		}
		tmpl.Execute(w, data)
	})

	fmt.Println("Frontend running at http://localhost:9090")
	http.ListenAndServe(":9090", nil)
}
