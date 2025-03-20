package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

// Submission represents a user submission.
type Submission struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Message string `json:"message"`
}

// enableCORS adds the necessary CORS headers.
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins; restrict as needed.
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
}

func main() {
	// Read database configuration from environment variables
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")

	// Set default values if necessary
	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbPort == "" {
		dbPort = "5432"
	}
	if dbUser == "" {
		dbUser = "postgres"
	}
	if dbPass == "" {
		dbPass = "password"
	}
	if dbName == "" {
		dbName = "mydb"
	}

	// Construct the Postgres connection string
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	// Connect to the database
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Ensure the connection is valid
	if err = db.Ping(); err != nil {
		log.Fatalf("Unable to reach the database: %v", err)
	}

	// Create the submissions table if it doesn't exist.
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS submissions (
		id SERIAL PRIMARY KEY,
		name TEXT,
		message TEXT
	);`
	if _, err := db.Exec(createTableSQL); err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

	// Handle preflight requests for CORS
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		// This is just a dummy endpoint to confirm the server is running.
		w.Write([]byte("Backend is running"))
	})

	// Endpoint to handle new submissions
	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		// Handle OPTIONS method for preflight requests
		if r.Method == http.MethodOptions {
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}
		// Parse form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form data", http.StatusBadRequest)
			return
		}
		name := r.FormValue("name")
		message := r.FormValue("message")
		if name == "" || message == "" {
			http.Error(w, "Missing form fields", http.StatusBadRequest)
			return
		}

		// Insert the data into the database
		insertSQL := "INSERT INTO submissions (name, message) VALUES ($1, $2)"
		_, err := db.Exec(insertSQL, name, message)
		if err != nil {
			http.Error(w, "Error inserting data", http.StatusInternalServerError)
			return
		}

		// Return a JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "Submission successful!"})
	})

	// Endpoint to list all submissions
	http.HandleFunc("/submissions", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		// Handle OPTIONS method for preflight requests
		if r.Method == http.MethodOptions {
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}
		rows, err := db.Query("SELECT id, name, message FROM submissions ORDER BY id DESC")
		if err != nil {
			http.Error(w, "Error fetching data", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var submissions []Submission
		for rows.Next() {
			var s Submission
			if err := rows.Scan(&s.ID, &s.Name, &s.Message); err != nil {
				http.Error(w, "Error scanning data", http.StatusInternalServerError)
				return
			}
			submissions = append(submissions, s)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(submissions)
	})

	fmt.Println("Backend running at http://localhost:9191")
	log.Fatal(http.ListenAndServe(":9191", nil))
}
