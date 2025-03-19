package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func main() {
	// Serve static assets (like CSS)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Serve a simple HTML page with Tailwind CSS
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
				<div class="bg-white p-8 rounded-lg shadow-lg text-center">
					<h1 class="text-3xl font-bold text-blue-600">Welcome to Go + Docker + Kubernetes!</h1>
					<p class="text-gray-700 mt-4">This app is running inside a Docker container and orchestrated with Kubernetes.</p>
				</div>
			</body>
			</html>
		`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	})

	// Start the server
	fmt.Println("Server running at http://localhost:9090")
	http.ListenAndServe(":9090", nil)
}
