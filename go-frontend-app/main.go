package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
)

func main() {
	// Serve static assets (like CSS)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	backendURL := os.Getenv("BACKEND_URL") // e.g., http://backend-service:9191
	if backendURL == "" {
		backendURL = "http://localhost:9191" // default value
	}

	// Serve the HTML page with a form
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.New("index").Parse(`
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<title>Go & Docker with Kubernetes</title>
				<link href="/static/style.css" rel="stylesheet">
			</head>
			<body class="bg-gray-100 flex items-center justify-center min-h-screen">
				<div class="bg-white p-8 rounded-lg shadow-lg text-center w-96">
					<h1 class="text-3xl font-bold text-blue-600">Welcome to Go + Docker + Kubernetes!</h1>
					<p class="text-gray-700 mt-4">Enter your details below:</p>
					<form action="{{.BackendURL}}/submit" method="POST" class="mt-4">
						<input type="text" name="name" placeholder="Your Name" class="border p-2 rounded w-full mb-2">
						<textarea name="message" placeholder="Your Message" class="border p-2 rounded w-full mb-2"></textarea>
						<button type="submit" class="bg-blue-600 text-white p-2 rounded w-full">Submit</button>
					</form>
				</div>
			</body>
			</html>
		`)
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

	// Start the server
	fmt.Println("Frontend running at http://localhost:9090")
	http.ListenAndServe(":9090", nil)
}
