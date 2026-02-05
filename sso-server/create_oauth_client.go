package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
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

	// Create OAuth2 client
	clientID := "demo-client"
	clientName := "Demo Client Application"
	clientDescription := "Demo client for testing SSO"

	// Note: For public clients (PKCE), we don't need a client_secret
	// Set client_secret_hash to NULL for public clients

	redirectURIs, _ := json.Marshal([]string{"http://localhost:3001/callback"})
	allowedScopes, _ := json.Marshal([]string{"openid", "profile", "email"})
	grantTypes, _ := json.Marshal([]string{"authorization_code", "refresh_token"})

	id := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO oauth2_clients (
			id, client_id, client_secret_hash, name, description, 
			redirect_uris, allowed_scopes, grant_types, 
			is_public, is_active, created_at, updated_at
		)
		VALUES (?, ?, NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE 
			redirect_uris = VALUES(redirect_uris),
			allowed_scopes = VALUES(allowed_scopes),
			grant_types = VALUES(grant_types)
	`

	_, err = db.Exec(query,
		id, clientID, clientName, clientDescription,
		string(redirectURIs), string(allowedScopes), string(grantTypes),
		true, // is_public (PKCE client)
		true, // is_active
		now, now,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("OAuth2 Client created successfully!\n")
	fmt.Printf("Client ID: %s\n", clientID)
	fmt.Printf("Client Name: %s\n", clientName)
	fmt.Printf("Redirect URIs: %s\n", string(redirectURIs))
	fmt.Printf("Is Public (PKCE): true\n")
}
