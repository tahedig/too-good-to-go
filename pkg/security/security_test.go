package security

import (
	"fmt"
	"regexp"
	"testing"
)

func TestGenerateNumericCode(t *testing.T) {
	for _, length := range []int{4, 6, 8} {
		re := regexp.MustCompile(fmt.Sprintf(`^\d{%d}$`, length))
		seen := make(map[string]bool)
		for i := 0; i < 200; i++ {
			code, err := GenerateNumericCode(length)
			if err != nil || !re.MatchString(code) {
				t.Fatalf("GenerateNumericCode(%d) = %q, %v", length, code, err)
			}
			seen[code] = true
		}
		if len(seen) < 150 {
			t.Fatalf("length %d: only %d distinct codes out of 200, generator looks broken", length, len(seen))
		}
	}
}

func TestHashAndMatch(t *testing.T) {
	salt, err := NewSalt()
	if err != nil {
		t.Fatal(err)
	}
	other, _ := NewSalt()
	if salt == other {
		t.Fatal("salts must be random")
	}

	hash := HashSecret(salt, "123456")
	if hash == "123456" || hash == HashSecret(other, "123456") {
		t.Fatal("hash must depend on the salt and never equal the secret")
	}
	if !SecretMatches(salt, "123456", hash) {
		t.Fatal("correct secret should match")
	}
	if SecretMatches(salt, "123457", hash) || SecretMatches(other, "123456", hash) {
		t.Fatal("wrong secret or salt must not match")
	}
}

func TestNewID(t *testing.T) {
	uuidV4 := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	a, err := NewID()
	if err != nil || !uuidV4.MatchString(a) {
		t.Fatalf("NewID() = %q, %v; want a UUID v4", a, err)
	}
	if b, _ := NewID(); a == b {
		t.Fatal("ids must be unique")
	}
}
