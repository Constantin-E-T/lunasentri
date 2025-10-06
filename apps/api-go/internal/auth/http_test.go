package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLoginHandler_Success(t *testing.T) {
	store := newMockStore()
	secret := "test-secret-key"
	ttl := 15 * time.Minute

	service, err := NewService(store, secret, ttl)
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}

	// Create a test user
	email := "test@example.com"
	password := "testpassword123"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	ctx := context.Background()
	_, err = store.CreateUser(ctx, email, hashedPassword)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Prepare login request
	loginReq := map[string]string{
		"email":    email,
		"password": password,
	}
	body, _ := json.Marshal(loginReq)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Create a test handler that mimics the main.go handleLogin
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		user, err := service.Authenticate(r.Context(), req.Email, req.Password)
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		token, err := service.CreateSession(user.ID)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		SetSessionCookie(w, token, int(ttl.Seconds()), false) // false for testing

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
		})
	})

	handler.ServeHTTP(rec, req)

	// Check response
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	// Check that cookie was set
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("Expected session cookie to be set")
	}

	found := false
	for _, cookie := range cookies {
		if cookie.Name == CookieName {
			found = true
			if cookie.HttpOnly != true {
				t.Error("Cookie should be HttpOnly")
			}
			if cookie.Secure != false {
				t.Error("Cookie should not be Secure in test mode (we passed false)")
			}
			if cookie.SameSite != http.SameSiteLaxMode {
				t.Error("Cookie should have SameSite=Lax")
			}
		}
	}

	if !found {
		t.Fatalf("Expected cookie named %s", CookieName)
	}
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	store := newMockStore()
	secret := "test-secret-key"
	ttl := 15 * time.Minute

	service, err := NewService(store, secret, ttl)
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}

	// Prepare login request with invalid credentials
	loginReq := map[string]string{
		"email":    "invalid@example.com",
		"password": "wrongpassword",
	}
	body, _ := json.Marshal(loginReq)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		_, err := service.Authenticate(r.Context(), req.Email, req.Password)
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
	})

	handler.ServeHTTP(rec, req)

	// Check response
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Expected status 401, got %d", rec.Code)
	}

	// Check that no cookie was set
	cookies := rec.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == CookieName {
			t.Fatal("Session cookie should not be set for invalid credentials")
		}
	}
}

func TestLogoutHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ClearSessionCookie(w, false) // false for testing
		w.WriteHeader(http.StatusNoContent)
	})

	handler.ServeHTTP(rec, req)

	// Check response
	if rec.Code != http.StatusNoContent {
		t.Fatalf("Expected status 204, got %d", rec.Code)
	}

	// Check that cookie was cleared
	cookies := rec.Result().Cookies()
	found := false
	for _, cookie := range cookies {
		if cookie.Name == CookieName {
			found = true
			if cookie.MaxAge != -1 {
				t.Error("Cookie MaxAge should be -1 to clear it")
			}
		}
	}

	if !found {
		t.Fatal("Expected cookie to be cleared")
	}
}

func TestRequireAuthMiddleware_Authenticated(t *testing.T) {
	store := newMockStore()
	secret := "test-secret-key"
	ttl := 15 * time.Minute

	service, err := NewService(store, secret, ttl)
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}

	// Create a test user
	email := "test@example.com"
	hashedPassword, _ := HashPassword("password")
	ctx := context.Background()
	user, err := store.CreateUser(ctx, email, hashedPassword)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Create a valid session token
	token, err := service.CreateSession(user.ID)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Create request with valid session cookie
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{
		Name:  CookieName,
		Value: token,
	})
	rec := httptest.NewRecorder()

	// Protected handler
	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "User not in context", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(user.Email))
	})

	// Wrap with auth middleware
	handler := service.RequireAuth(protectedHandler)
	handler.ServeHTTP(rec, req)

	// Check response
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	if rec.Body.String() != email {
		t.Fatalf("Expected body %s, got %s", email, rec.Body.String())
	}
}

func TestRequireAuthMiddleware_Unauthenticated(t *testing.T) {
	store := newMockStore()
	secret := "test-secret-key"
	ttl := 15 * time.Minute

	service, err := NewService(store, secret, ttl)
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}

	// Create request without session cookie
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()

	// Protected handler
	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with auth middleware
	handler := service.RequireAuth(protectedHandler)
	handler.ServeHTTP(rec, req)

	// Check response
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Expected status 401, got %d", rec.Code)
	}
}

func TestRequireAuthMiddleware_InvalidToken(t *testing.T) {
	store := newMockStore()
	secret := "test-secret-key"
	ttl := 15 * time.Minute

	service, err := NewService(store, secret, ttl)
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}

	// Create request with invalid session cookie
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{
		Name:  CookieName,
		Value: "invalid-token",
	})
	rec := httptest.NewRecorder()

	// Protected handler
	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with auth middleware
	handler := service.RequireAuth(protectedHandler)
	handler.ServeHTTP(rec, req)

	// Check response
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Expected status 401, got %d", rec.Code)
	}
}
