package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

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

	// Create the table if it doesn't exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS submissions (
		id SERIAL PRIMARY KEY,
		name TEXT,
		message TEXT
	);`
	if _, err := db.Exec(createTableSQL); err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

	// Handle form submissions
	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
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

		w.Write([]byte("Submission successful!"))
	})

	// Start the backend server on port 9191
	fmt.Println("Backend running at http://localhost:9191")
	log.Fatal(http.ListenAndServe(":9191", nil))
}
