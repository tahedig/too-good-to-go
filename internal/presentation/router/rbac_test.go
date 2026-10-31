package router_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/tahedig/too-good-to-go/backend/internal/application/dto"
	"github.com/tahedig/too-good-to-go/backend/internal/domain/enum"
	"github.com/tahedig/too-good-to-go/backend/internal/domain/exception"
)

// login runs the full OTP flow and returns the verify response.
func (e *authEnv) login(t *testing.T, phone string) dto.VerifyOTPResponse {
	t.Helper()
	code, w := e.request(t, phone)
	if w.Code != http.StatusOK {
		t.Fatalf("request otp: %d %s", w.Code, w.Body.String())
	}
	w = e.verify(phone, code)
	if w.Code != http.StatusOK {
		t.Fatalf("verify otp: %d %s", w.Code, w.Body.String())
	}
	var resp dto.VerifyOTPResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	return resp
}

func (e *authEnv) get(path, authHeader string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	w := httptest.NewRecorder()
	e.r.ServeHTTP(w, req)
	return w
}

func TestVerifyReturnsTokenAndCreatesOrReusesUser(t *testing.T) {
	e := newAuthEnv(t, time.Minute, 3)

	first := e.login(t, testPhone)
	if first.AccessToken == "" || first.TokenType != "Bearer" || first.ExpiresIn != 3600 {
		t.Fatalf("unexpected token response: %+v", first)
	}
	if first.User.Role != string(enum.RoleCustomer) || first.User.Phone != testPhone || first.User.ID == "" {
		t.Fatalf("new users should be customers: %+v", first.User)
	}

	second := e.login(t, testPhone)
	if second.User.ID != first.User.ID {
		t.Fatalf("same phone must map to the same user: %s vs %s", first.User.ID, second.User.ID)
	}
	if second.AccessToken == "" {
		t.Fatal("second login should also return a token")
	}

	claims, err := e.tokens.Parse(first.AccessToken)
	if err != nil || claims.UserID != first.User.ID || claims.Role != enum.RoleCustomer {
		t.Fatalf("token claims mismatch: %+v err=%v", claims, err)
	}
}

func TestRolesFromConfiguredPhones(t *testing.T) {
	e := newAuthEnv(t, time.Minute, 3)
	if got := e.login(t, vendorPhone).User.Role; got != string(enum.RoleVendor) {
		t.Errorf("vendor phone got role %s", got)
	}
	if got := e.login(t, adminPhone).User.Role; got != string(enum.RoleAdmin) {
		t.Errorf("admin phone got role %s", got)
	}
}

func TestMissingOrInvalidAuthReturns401(t *testing.T) {
	e := newAuthEnv(t, time.Minute, 3)

	expired := signToken(t, testSecret, testIssuer, enum.RoleAdmin, jwt.SigningMethodHS256, time.Now().Add(-time.Hour))
	wrongSecret := signToken(t, "another-secret-another-secret-12345", testIssuer, enum.RoleAdmin, jwt.SigningMethodHS256, time.Now().Add(time.Hour))
	wrongIssuer := signToken(t, testSecret, "someone-else", enum.RoleAdmin, jwt.SigningMethodHS256, time.Now().Add(time.Hour))
	hs512 := signToken(t, testSecret, testIssuer, enum.RoleAdmin, jwt.SigningMethodHS512, time.Now().Add(time.Hour))
	badRole := signToken(t, testSecret, testIssuer, enum.Role("root"), jwt.SigningMethodHS256, time.Now().Add(time.Hour))

	cases := map[string]string{
		"no header":     "",
		"wrong scheme":  "Basic abc",
		"empty bearer":  "Bearer ",
		"garbage":       "Bearer not.a.jwt",
		"expired":       "Bearer " + expired,
		"wrong secret":  "Bearer " + wrongSecret,
		"wrong issuer":  "Bearer " + wrongIssuer,
		"other alg":     "Bearer " + hs512,
		"unknown role":  "Bearer " + badRole,
		"token no type": expired,
	}
	for name, header := range cases {
		for _, path := range []string{"/auth/me", "/test/admin", "/test/customer"} {
			w := e.get(path, header)
			if w.Code != http.StatusUnauthorized || decodeError(t, w).Code != exception.CodeUnauthorized {
				t.Errorf("%s %s: got %d %s, want 401", name, path, w.Code, w.Body.String())
			}
		}
	}
}

func TestRoleAccessMatrix(t *testing.T) {
	e := newAuthEnv(t, time.Minute, 3)
	tokens := map[enum.Role]string{
		enum.RoleCustomer: e.login(t, testPhone).AccessToken,
		enum.RoleVendor:   e.login(t, vendorPhone).AccessToken,
		enum.RoleAdmin:    e.login(t, adminPhone).AccessToken,
	}
	routes := map[string]enum.Role{
		"/test/customer": enum.RoleCustomer,
		"/test/vendor":   enum.RoleVendor,
		"/test/admin":    enum.RoleAdmin,
	}

	for role, tok := range tokens {
		for path, allowed := range routes {
			w := e.get(path, "Bearer "+tok)
			if role == allowed {
				var body dto.MeResponse
				if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &body) != nil || body.Role != string(role) {
					t.Errorf("%s on %s: got %d %s, want 200", role, path, w.Code, w.Body.String())
				}
				continue
			}
			if w.Code != http.StatusForbidden || decodeError(t, w).Code != exception.CodeForbidden {
				t.Errorf("%s on %s: got %d %s, want 403", role, path, w.Code, w.Body.String())
			}
		}
		// every role may read its own identity
		if w := e.get("/auth/me", "Bearer "+tok); w.Code != http.StatusOK {
			t.Errorf("%s on /auth/me: got %d", role, w.Code)
		}
	}
}

func signToken(t *testing.T, secret, issuer string, role enum.Role, method jwt.SigningMethod, exp time.Time) string {
	t.Helper()
	s, err := jwt.NewWithClaims(method, jwt.MapClaims{
		"sub":  "user-1",
		"role": string(role),
		"iss":  issuer,
		"iat":  time.Now().Add(-2 * time.Hour).Unix(),
		"exp":  exp.Unix(),
	}).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return s
}
