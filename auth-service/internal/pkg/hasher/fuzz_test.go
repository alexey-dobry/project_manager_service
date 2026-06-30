package hasher

import (
	"strings"
	"testing"
)

// FuzzBcryptCompare — два инварианта:
//  1. Compare(Hash(p), p) ≡ nil для любого p длиной ≤ 72 байт (предел bcrypt).
//  2. Compare(Hash(p), q) ≡ ErrInvalidCredentials для q ≠ p.
//
// Дополнительно — Hash/Compare не должны паниковать ни на каких байтах,
// включая нулевые, не-UTF-8 и многобайтные.
//
// Запуск:
//   go test ./internal/pkg/hasher -run=- -fuzz=FuzzBcryptCompare -fuzztime=20s
func FuzzBcryptCompare(f *testing.F) {
	h := New(4) // bcrypt MinCost — fuzz нужно гонять быстро

	seeds := []string{
		"",
		"a",
		"password",
		"пароль",
		"\x00\x00\x00",
		"\xff\xfe",
		strings.Repeat("x", 72), // ровно граница bcrypt
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, password string) {
		// bcrypt не поддерживает пароли длиннее 72 байт — режем.
		// Это не часть проверяемой логики, а ограничение библиотеки.
		if len(password) > 72 {
			password = password[:72]
		}
		hash, err := h.Hash(password)
		if err != nil {
			// Любая ошибка хеширования с укороченным input — баг.
			t.Fatalf("Hash failed for %q: %v", password, err)
		}
		// 1. Hash → Compare с тем же паролем — успех.
		if err := h.Compare(hash, password); err != nil {
			t.Fatalf("self-compare failed: %v (pass=%q)", err, password)
		}
		// 2. Compare с другим паролем (если пустой — добавим что-то).
		other := password + "x"
		if len(other) > 72 {
			other = other[:72]
		}
		if other == password {
			return // не на чем проверять — пропускаем
		}
		if err := h.Compare(hash, other); err == nil {
			t.Fatalf("compare with wrong password returned nil err (pass=%q, other=%q)",
				password, other)
		}
	})
}
