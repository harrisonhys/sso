package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Database connection
	dsn := "root:root@tcp(localhost:3306)/sso-main?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	// Hash password
	password := "Test123!"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	// Create user
	userID := uuid.New().String()
	email := "test@example.com"
	name := "Test User"
	now := time.Now()

	query := `
		INSERT INTO users (id, email, name, password_hash, is_active, email_verified, password_changed_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE password_hash = VALUES(password_hash), password_changed_at = VALUES(password_changed_at)
	`

	_, err = db.Exec(query, userID, email, name, string(hashedPassword), true, true, now, now, now)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("User created successfully!\n")
	fmt.Printf("Email: %s\n", email)
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("User ID: %s\n", userID)
}
