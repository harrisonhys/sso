package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// The password from seed script
	password := "Admin@123"

	// Generate a new hash
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		fmt.Println("Error generating hash:", err)
		return
	}

	fmt.Println("Password:", password)
	fmt.Println("Hash:", string(hash))

	// Test comparison
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil {
		fmt.Println("Comparison failed:", err)
	} else {
		fmt.Println("✓ Password matches hash!")
	}

	// Test with seed script hash
	seedHash := "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5kuyvv5ABwNNm"
	err = bcrypt.CompareHashAndPassword([]byte(seedHash), []byte(password))
	if err != nil {
		fmt.Println("✗ Seed hash does NOT match password:", err)
	} else {
		fmt.Println("✓ Seed hash matches!")
	}
}
