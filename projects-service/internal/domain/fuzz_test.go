package domain

import (
	"testing"
)

// FuzzTaskStateMachine — проверяет, что:
//  1. CanTransitionTo не паникует ни на каких входных строках, в т.ч.
//     невалидных статусах.
//  2. Сохраняется идемпотентность: s.CanTransitionTo(s) == true для валидных s.
//  3. Несуществующий статус никогда не может переходить никуда.
//
// Запуск:
//   go test ./internal/domain -run=- -fuzz=FuzzTaskStateMachine -fuzztime=10s
func FuzzTaskStateMachine(f *testing.F) {
	allValid := []TaskStatus{TaskTodo, TaskInProgress, TaskDone, TaskBlocked}

	seeds := [][2]string{
		{"todo", "in_progress"},
		{"in_progress", "done"},
		{"done", "todo"},        // запрещённый
		{"blocked", "done"},     // запрещённый
		{"", ""},                // оба невалидные
		{"todo", ""},            // целевой невалидный
		{"\x00", "todo"},        // мусор как from
		{"DROP TABLE", "x"},     // sql-ish injection
		{"todo", "todo"},        // идемпотентность
	}
	for _, s := range seeds {
		f.Add(s[0], s[1])
	}

	f.Fuzz(func(t *testing.T, from, to string) {
		// 1. Не паникует.
		fromS := TaskStatus(from)
		toS := TaskStatus(to)
		got := fromS.CanTransitionTo(toS)

		// 2. Если from невалидный — никаких разрешённых переходов
		//    (кроме случая from == to с тем же невалидным значением,
		//    что мы тоже считаем "невалидным переходом").
		if !fromS.IsValid() {
			if got && fromS != toS {
				t.Fatalf("invalid from=%q allowed transition to %q", from, to)
			}
		}
		// 3. Если to невалидный (и не равен from) — не должен разрешаться.
		if !toS.IsValid() && fromS != toS && got {
			t.Fatalf("transition %q -> %q to invalid target allowed", from, to)
		}

		// 4. Для всех валидных from идемпотентность должна выполняться всегда.
		for _, v := range allValid {
			if !v.CanTransitionTo(v) {
				t.Fatalf("idempotency broken for %q", v)
			}
		}
	})
}
