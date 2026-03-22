package middleware_test

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mcopland/spotifind/internal/middleware"
)

const testSecret = "super-secret"

func mintToken(t *testing.T, claims jwt.MapClaims, method jwt.SigningMethod, key any) string {
	t.Helper()
	tok := jwt.NewWithClaims(method, claims)
	signed, err := tok.SignedString(key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func requestWithCookie(token string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	return req
}

// nextHandler records whether it was called and what user ID it saw.
func nextHandler(called *bool, gotUserID *int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*called = true
		id, _ := middleware.GetUserID(r.Context())
		*gotUserID = id
		w.WriteHeader(http.StatusOK)
	})
}

func TestAuth_NoCookie(t *testing.T) {
	var called bool
	var gotID int64
	handler := middleware.Auth(testSecret)(nextHandler(&called, &gotID))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
	if called {
		t.Error("next handler must not be called when cookie is absent")
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	var called bool
	var gotID int64
	handler := middleware.Auth(testSecret)(nextHandler(&called, &gotID))

	req := requestWithCookie("this.is.not.a.jwt")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
	if called {
		t.Error("next handler must not be called for invalid token")
	}
}

func TestAuth_ExpiredToken(t *testing.T) {
	var called bool
	var gotID int64
	h := middleware.Auth(testSecret)(nextHandler(&called, &gotID))

	signed := mintToken(t, jwt.MapClaims{
		"sub":     "user123",
		"user_id": float64(1),
		"exp":     time.Now().Add(-time.Hour).Unix(),
	}, jwt.SigningMethodHS256, []byte(testSecret))

	req := requestWithCookie(signed)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
	if called {
		t.Error("next handler must not be called for expired token")
	}
}

func TestAuth_WrongSigningMethod(t *testing.T) {
	var called bool
	var gotID int64
	h := middleware.Auth(testSecret)(nextHandler(&called, &gotID))

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	signed := mintToken(t, jwt.MapClaims{
		"sub":     "user123",
		"user_id": float64(1),
		"exp":     time.Now().Add(time.Hour).Unix(),
	}, jwt.SigningMethodRS256, privKey)

	req := requestWithCookie(signed)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
	if called {
		t.Error("next handler must not be called for wrong signing method")
	}
}

func TestAuth_ValidToken(t *testing.T) {
	var called bool
	var gotID int64
	h := middleware.Auth(testSecret)(nextHandler(&called, &gotID))

	signed := mintToken(t, jwt.MapClaims{
		"sub":     "user123",
		"user_id": float64(42),
		"exp":     time.Now().Add(time.Hour).Unix(),
	}, jwt.SigningMethodHS256, []byte(testSecret))

	req := requestWithCookie(signed)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if !called {
		t.Error("next handler must be called for valid token")
	}
	if gotID != 42 {
		t.Errorf("expected user_id 42 in context, got %d", gotID)
	}
}

func TestAuth_MissingUserID(t *testing.T) {
	var called bool
	var gotID int64
	h := middleware.Auth(testSecret)(nextHandler(&called, &gotID))

	// Token has sub but no user_id claim.
	signed := mintToken(t, jwt.MapClaims{
		"sub": "user123",
		"exp": time.Now().Add(time.Hour).Unix(),
	}, jwt.SigningMethodHS256, []byte(testSecret))

	req := requestWithCookie(signed)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
	if called {
		t.Error("next handler must not be called when user_id is missing")
	}
}

func TestAuth_InvalidSubjectType(t *testing.T) {
	var called bool
	var gotID int64
	h := middleware.Auth(testSecret)(nextHandler(&called, &gotID))

	// JWT with numeric sub causes GetSubject() to return an error.
	signed := mintToken(t, jwt.MapClaims{
		"sub":     float64(12345),
		"user_id": float64(42),
		"exp":     time.Now().Add(time.Hour).Unix(),
	}, jwt.SigningMethodHS256, []byte(testSecret))

	req := requestWithCookie(signed)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
	if called {
		t.Error("next handler must not be called when subject type is invalid")
	}
}
