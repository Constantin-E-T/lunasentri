package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	// CookieName is the name of the session cookie
	CookieName = "lunasentri_session"
)

// JWTClaims represents the claims in a JWT token
type JWTClaims struct {
	UserID int   `json:"uid"`
	Exp    int64 `json:"exp"`
	Iat    int64 `json:"iat"`
}

// CreateJWT creates a new JWT token for the given user ID
func CreateJWT(userID int, secret []byte, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID: userID,
		Iat:    now.Unix(),
		Exp:    now.Add(ttl).Unix(),
	}

	// Create header
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	// Encode header
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %w", err)
	}
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Encode claims
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Create signature
	message := headerEncoded + "." + claimsEncoded
	signature := createSignature(message, secret)

	// Combine parts
	token := message + "." + signature

	return token, nil
}

// ValidateJWT validates a JWT token and returns the user ID
func ValidateJWT(token string, secret []byte) (int, error) {
	// Split token into parts
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid token format")
	}

	headerEncoded := parts[0]
	claimsEncoded := parts[1]
	signature := parts[2]

	// Verify signature
	message := headerEncoded + "." + claimsEncoded
	expectedSignature := createSignature(message, secret)
	if signature != expectedSignature {
		return 0, fmt.Errorf("invalid signature")
	}

	// Decode claims
	claimsJSON, err := base64.RawURLEncoding.DecodeString(claimsEncoded)
	if err != nil {
		return 0, fmt.Errorf("failed to decode claims: %w", err)
	}

	var claims JWTClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return 0, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	// Check expiration
	if time.Now().Unix() > claims.Exp {
		return 0, fmt.Errorf("token expired")
	}

	return claims.UserID, nil
}

// createSignature creates an HMAC-SHA256 signature for the given message
func createSignature(message string, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// SetSessionCookie sets the session cookie on the response
func SetSessionCookie(w http.ResponseWriter, token string, maxAge int, secure bool) {
	cookie := &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

// ClearSessionCookie clears the session cookie
func ClearSessionCookie(w http.ResponseWriter, secure bool) {
	cookie := &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

// GetSessionCookie retrieves the session cookie from the request
func GetSessionCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return "", fmt.Errorf("session cookie not found")
	}
	return cookie.Value, nil
}
