package auth

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

// BootstrapAdmin creates or updates an admin user if ADMIN_EMAIL and ADMIN_PASSWORD env vars are set
func BootstrapAdmin(ctx context.Context, store storage.Store) error {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")

	// Skip if either environment variable is not set
	if adminEmail == "" || adminPassword == "" {
		log.Println("Admin bootstrap skipped: ADMIN_EMAIL or ADMIN_PASSWORD not set")
		return nil
	}

	// Validate email is not empty after trimming
	if adminEmail == "" {
		return fmt.Errorf("ADMIN_EMAIL cannot be empty")
	}

	// Validate password is not empty
	if adminPassword == "" {
		return fmt.Errorf("ADMIN_PASSWORD cannot be empty")
	}

	// Hash the password
	hashedPassword, err := HashPassword(adminPassword)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	// Upsert the admin user
	user, err := store.UpsertAdmin(ctx, adminEmail, hashedPassword)
	if err != nil {
		return fmt.Errorf("failed to upsert admin user: %w", err)
	}

	log.Printf("Admin user initialized: email=%s, id=%d", user.Email, user.ID)
	return nil
}