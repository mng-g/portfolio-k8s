package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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

// getEnv fetches an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func main() {
	// Read database configuration from environment variables
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASS", "password")
	dbName := getEnv("DB_NAME", "mydb")

	// Construct the Postgres connection string
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	// Connect to the database with retries
	var db *sql.DB
	var err error
	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", psqlInfo)
		if err == nil {
			break
		}
		log.Printf("Database connection failed (attempt %d): %v", i+1, err)
		time.Sleep(2 * time.Second) // Wait before retrying
	}
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Ensure the connection is valid
	if err = db.Ping(); err != nil {
		log.Fatalf("Unable to reach the database: %v", err)
	}

	log.Println("Connected to the database successfully.")

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
	log.Println("Table 'submissions' ensured to exist.")

	// Handle preflight requests for CORS
	http.HandleFunc("/api/ready", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		log.Printf("Received request: %s %s", r.Method, r.URL.Path)
		w.Write([]byte("Backend is running"))
	})

	// Health Check Endpoint
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		log.Println("Health check requested.")
		if err := db.Ping(); err != nil {
			http.Error(w, "Database connection error", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Endpoint to handle new submissions
	http.HandleFunc("/api/submit", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		log.Printf("Received submission request: %s", r.Method)

		// Handle OPTIONS method for CORS preflight
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
			log.Printf("Error inserting data: %v", err)
			http.Error(w, "Error inserting data", http.StatusInternalServerError)
			return
		}

		log.Printf("New submission: Name=%s, Message=%s", name, message)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "Submission successful!"})
	})

	// Endpoint to list all submissions
	http.HandleFunc("/api/submissions", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		log.Printf("Received request: %s %s", r.Method, r.URL.Path)

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
			log.Printf("Error fetching submissions: %v", err)
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

		log.Printf("Fetched %d submissions", len(submissions))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(submissions)
	})

	log.Println("Backend running at :9191")
	log.Fatal(http.ListenAndServe(":9191", nil))
}
