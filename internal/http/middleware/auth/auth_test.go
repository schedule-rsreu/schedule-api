package auth_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	authmw "github.com/schedule-rsreu/schedule-api/internal/http/middleware/auth"
)

func TestMiddlewareAcceptsSupportedCredentials(t *testing.T) {
	const secret = "123456:test-secret"
	accessToken := signedToken(t, secret, "access")

	tests := []struct {
		name   string
		header string
		cookie *http.Cookie
	}{
		{"bearer", "Bearer " + accessToken, nil},
		{"case-insensitive bearer", "bearer " + accessToken, nil},
		{"access token cookie", "", &http.Cookie{Name: "access_token", Value: accessToken}},
		{"telegram mini app", "TMA " + signedTMA(secret), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := request(t, secret, tt.header, tt.cookie)
			if rec.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want %d; body = %q", rec.Code, http.StatusNoContent, rec.Body.String())
			}
		})
	}
}

func TestMiddlewareRejectsInvalidCredentials(t *testing.T) {
	const secret = "123456:test-secret"
	withoutExpiration := signedClaims(t, secret, jwt.SigningMethodHS256, jwt.MapClaims{"type": "access"})
	wrongAlgorithm := signedClaims(t, secret, jwt.SigningMethodHS512, jwt.MapClaims{
		"type": "access",
		"exp":  time.Now().Add(time.Hour).Unix(),
	})
	tests := []struct {
		name   string
		header string
		cookie *http.Cookie
	}{
		{"missing credentials", "", nil},
		{"refresh token", "Bearer " + signedToken(t, secret, "refresh"), nil},
		{"invalid token", "Bearer invalid", nil},
		{"token without expiration", "Bearer " + withoutExpiration, nil},
		{"token with another HMAC algorithm", "Bearer " + wrongAlgorithm, nil},
		{"malformed header does not fall back to cookie", "Bearer", &http.Cookie{Name: "access_token", Value: signedToken(t, secret, "access")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := request(t, secret, tt.header, tt.cookie)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
			}
		})
	}
}

func request(t *testing.T, secret, header string, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	e.Use(authmw.New(secret))
	e.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusNoContent) })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if header != "" {
		req.Header.Set(echo.HeaderAuthorization, header)
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func signedToken(t *testing.T, secret, tokenType string) string {
	t.Helper()
	return signedClaims(t, secret, jwt.SigningMethodHS256, jwt.MapClaims{
		"type": tokenType,
		"exp":  time.Now().Add(time.Hour).Unix(),
	})
}

func signedClaims(t *testing.T, secret string, method jwt.SigningMethod, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(method, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

func signedTMA(botToken string) string {
	values := url.Values{
		"auth_date": {strconv.FormatInt(time.Now().Unix(), 10)},
		"query_id":  {"test-query"},
		"user":      {`{"id":1,"first_name":"Test"}`},
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+values.Get(key))
	}
	secretHash := hmac.New(sha256.New, []byte("WebAppData"))
	secretHash.Write([]byte(botToken))
	signature := hmac.New(sha256.New, secretHash.Sum(nil))
	signature.Write([]byte(strings.Join(parts, "\n")))
	values.Set("hash", hex.EncodeToString(signature.Sum(nil)))
	return values.Encode()
}
