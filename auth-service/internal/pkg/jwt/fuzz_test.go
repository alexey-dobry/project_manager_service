package jwt

import (
	"strings"
	"testing"
)

// FuzzParseAccessToken проверяет, что ParseAccess не паникует на произвольном
// входе и не возвращает непустые (id, role) одновременно с ошибкой.
func FuzzParseAccessToken(f *testing.F) {
	p, err := New(Config{
		Secret: "fuzz-secret-min-16-chars-OK",
		Issuer: "fuzz",
	})
	if err != nil {
		f.Fatalf("setup: %v", err)
	}

	seeds := []string{
		"",
		".",
		"..",
		"a.b.c",
		"eyJhbGciOiJub25lIn0.e30.",
		"eyJhbGciOiJIUzI1NiJ9.e30.signature",
		"Bearer eyJhbGciOiJIUzI1NiJ9.e30.x",
		strings.Repeat("A", 8192),
		"\x00\x01\x02",
		"\xff\xfe\xfd",
		"тест.токен.подпись",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, token string) {
		id, role, err := p.ParseAccess(token)
		if err != nil {
			if id.String() != "00000000-0000-0000-0000-000000000000" {
				t.Fatalf("err set, but id != nil: %q", id)
			}
			if role != "" {
				t.Fatalf("err set, but role != empty: %q", role)
			}
			return
		}
		if !role.IsValid() {
			t.Fatalf("no err but role invalid: %q (token=%q)", role, token)
		}
	})
}
