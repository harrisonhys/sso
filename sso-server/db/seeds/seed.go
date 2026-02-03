package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Database connection
	db, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/sso-main?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println("Starting database seeding...")

	// 1. Seed Roles
	seedRoles(db)

	// 2. Seed Permissions
	seedPermissions(db)

	// 3. Seed Role-Permission Mappings
	seedRolePermissions(db)

	// 4. Seed Default Admin User
	seedDefaultAdmin(db)

	// 5. Seed OAuth Clients
	seedOAuthClients(db)

	// 6. Seed System Config
	seedSystemConfig(db)

	log.Println("Database seeding completed successfully!")
}

func seedRoles(db *sql.DB) {
	log.Println("Seeding roles...")

	roles := []struct {
		name        string
		description string
	}{
		{"super_admin", "Full system access including system configuration"},
		{"admin", "Management dashboard access for user and role management"},
		{"user", "Basic user access to client applications"},
	}

	for _, role := range roles {
		id := uuid.New().String()
		_, err := db.Exec(`
			INSERT INTO roles (id, name, description)
			VALUES (?, ?, ?)
			ON DUPLICATE KEY UPDATE description = VALUES(description)
		`, id, role.name, role.description)

		if err != nil {
			log.Printf("Error seeding role %s: %v", role.name, err)
		} else {
			log.Printf("✓ Seeded role: %s", role.name)
		}
	}
}

func seedPermissions(db *sql.DB) {
	log.Println("Seeding permissions...")

	permissions := []struct {
		name        string
		description string
		resource    string
		action      string
	}{
		{"users:read", "Read user data", "users", "read"},
		{"users:write", "Create and update users", "users", "write"},
		{"users:delete", "Delete users", "users", "delete"},
		{"roles:read", "Read roles", "roles", "read"},
		{"roles:write", "Create and update roles", "roles", "write"},
		{"roles:delete", "Delete roles", "roles", "delete"},
		{"permissions:read", "Read permissions", "permissions", "read"},
		{"permissions:write", "Manage permissions", "permissions", "write"},
		{"config:read", "Read system configuration", "config", "read"},
		{"config:write", "Update system configuration", "config", "write"},
		{"audit:read", "Read audit logs", "audit", "read"},
		{"clients:read", "Read OAuth clients", "clients", "read"},
		{"clients:write", "Manage OAuth clients", "clients", "write"},
	}

	for _, perm := range permissions {
		id := uuid.New().String()
		_, err := db.Exec(`
			INSERT INTO permissions (id, name, description, resource, action)
			VALUES (?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE description = VALUES(description)
		`, id, perm.name, perm.description, perm.resource, perm.action)

		if err != nil {
			log.Printf("Error seeding permission %s: %v", perm.name, err)
		} else {
			log.Printf("✓ Seeded permission: %s", perm.name)
		}
	}
}

func seedRolePermissions(db *sql.DB) {
	log.Println("Seeding role-permission mappings...")

	// Get role IDs
	var superAdminID, adminID, userID string
	db.QueryRow("SELECT id FROM roles WHERE name = 'super_admin'").Scan(&superAdminID)
	db.QueryRow("SELECT id FROM roles WHERE name = 'admin'").Scan(&adminID)
	db.QueryRow("SELECT id FROM roles WHERE name = 'user'").Scan(&userID)

	// Get all permission IDs
	var allPermissions []string
	rows, _ := db.Query("SELECT id FROM permissions")
	defer rows.Close()
	for rows.Next() {
		var id string
		rows.Scan(&id)
		allPermissions = append(allPermissions, id)
	}

	// Super Admin gets all permissions
	for _, permID := range allPermissions {
		db.Exec(`
			INSERT IGNORE INTO role_permissions (role_id, permission_id)
			VALUES (?, ?)
		`, superAdminID, permID)
	}
	log.Printf("✓ Assigned all permissions to super_admin")

	// Admin gets limited permissions
	adminPerms := []string{"users:read", "users:write", "roles:read", "audit:read", "clients:read"}
	for _, permName := range adminPerms {
		var permID string
		db.QueryRow("SELECT id FROM permissions WHERE name = ?", permName).Scan(&permID)
		db.Exec(`
			INSERT IGNORE INTO role_permissions (role_id, permission_id)
			VALUES (?, ?)
		`, adminID, permID)
	}
	log.Printf("✓ Assigned limited permissions to admin")

	// User gets basic permissions
	userPerms := []string{"users:read"}
	for _, permName := range userPerms {
		var permID string
		db.QueryRow("SELECT id FROM permissions WHERE name = ?", permName).Scan(&permID)
		db.Exec(`
			INSERT IGNORE INTO role_permissions (role_id, permission_id)
			VALUES (?, ?)
		`, userID, permID)
	}
	log.Printf("✓ Assigned basic permissions to user")
}

func seedDefaultAdmin(db *sql.DB) {
	log.Println("Seeding default admin user...")

	// Hash the password properly
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("Admin@123"), 12)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return
	}

	userID := uuid.New().String()
	_, err = db.Exec(`
		INSERT INTO users (id, email, password_hash, name, email_verified, is_active, password_changed_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE password_hash = VALUES(password_hash), password_changed_at = NOW()
	`, userID, "admin@sso.local", string(passwordHash), "System Administrator", true, true)

	if err != nil {
		log.Printf("Error seeding admin user: %v", err)
		return
	}

	// Assign super_admin role
	var superAdminID string
	db.QueryRow("SELECT id FROM roles WHERE name = 'super_admin'").Scan(&superAdminID)

	db.Exec(`
		INSERT IGNORE INTO user_roles (user_id, role_id)
		VALUES (?, ?)
	`, userID, superAdminID)

	log.Printf("✓ Created admin user: admin@sso.local / Admin@123")
}

func seedOAuthClients(db *sql.DB) {
	log.Println("Seeding OAuth clients...")

	clients := []struct {
		clientID      string
		clientSecret  string
		name          string
		redirectURIs  string
		allowedScopes string
	}{
		{
			"client-app-1",
			"$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5kuyvv5ABwNNm", // hashed "secret-1"
			"Demo Client Application",
			"http://localhost:3001/callback,http://localhost:3001/auth/callback",
			"openid,profile,email",
		},
		{
			"management-ui",
			"$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5kuyvv5ABwNNm", // hashed "management-secret"
			"SSO Management Dashboard",
			"http://localhost:3002/callback,http://localhost:3002/auth/callback",
			"openid,profile,email",
		},
	}

	for _, client := range clients {
		id := uuid.New().String()
		_, err := db.Exec(`
			INSERT INTO oauth_clients (id, client_id, client_secret, name, redirect_uris, allowed_scopes, is_active)
			VALUES (?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE name = VALUES(name)
		`, id, client.clientID, client.clientSecret, client.name, client.redirectURIs, client.allowedScopes, true)

		if err != nil {
			log.Printf("Error seeding OAuth client %s: %v", client.clientID, err)
		} else {
			log.Printf("✓ Seeded OAuth client: %s", client.name)
		}
	}
}

func seedSystemConfig(db *sql.DB) {
	log.Println("Seeding system configuration...")

	configs := []struct {
		key   string
		value string
	}{
		{"password_policy", `{"min_length":8,"require_uppercase":true,"require_lowercase":true,"require_number":true,"require_special":true,"prevent_reuse_count":5,"expiry_days":90,"warning_days":7}`},
		{"session", `{"timeout_minutes":30,"max_concurrent_sessions":5}`},
		{"security", `{"max_login_attempts":5,"lockout_duration_minutes":30,"enforce_2fa":false,"password_reset_token_expiry_hours":1}`},
		{"rate_limit", `{"login_per_minute":5,"api_per_minute":100}`},
	}

	for _, cfg := range configs {
		id := uuid.New().String()
		_, err := db.Exec(`
			INSERT INTO system_config (id, config_key, config_value)
			VALUES (?, ?, ?)
			ON DUPLICATE KEY UPDATE config_value = VALUES(config_value)
		`, id, cfg.key, cfg.value)

		if err != nil {
			log.Printf("Error seeding config %s: %v", cfg.key, err)
		} else {
			log.Printf("✓ Seeded config: %s", cfg.key)
		}
	}
}
