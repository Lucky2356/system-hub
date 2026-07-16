package i18n_test

import (
	"testing"

	"github.com/Lucky2356/system-hub/internal/i18n"
)

func TestSetLanguageResolvesCodes(t *testing.T) {
	defer i18n.SetLanguage(i18n.English)

	tests := []struct {
		name string
		code string
		want string
	}{
		{"russian", i18n.Russian, i18n.Russian},
		{"english", i18n.English, i18n.English},
		{"case and spaces are tolerated", "  RU  ", i18n.Russian},
		// An unknown code must not leave the app with no language at all.
		{"unknown falls back to english", "klingon", i18n.English},
		{"empty falls back to english", "", i18n.English},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := i18n.SetLanguage(tc.code); got != tc.want {
				t.Errorf("SetLanguage(%q) = %q, want %q", tc.code, got, tc.want)
			}
			if got := i18n.Language(); got != tc.want {
				t.Errorf("Language() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSetLanguageAutoPicksASupportedLanguage(t *testing.T) {
	defer i18n.SetLanguage(i18n.English)

	// "auto" reads the OS locale, so the value depends on the machine; what must
	// hold everywhere is that it never resolves to "auto" itself or to a
	// language with no catalog.
	got := i18n.SetLanguage(i18n.Auto)

	if got == i18n.Auto {
		t.Fatal(`SetLanguage("auto") must resolve to a concrete language`)
	}

	for _, supported := range i18n.Supported() {
		if got == supported {
			return
		}
	}
	t.Errorf("SetLanguage(auto) = %q, which is not in Supported() = %v", got, i18n.Supported())
}

func TestTranslatesAndFallsBackToTheKey(t *testing.T) {
	defer i18n.SetLanguage(i18n.English)

	i18n.SetLanguage(i18n.Russian)
	if got, want := i18n.T("Refresh"), "Обновить"; got != want {
		t.Errorf("T(Refresh) = %q, want %q", got, want)
	}

	// English needs no catalog: the keys are already English.
	i18n.SetLanguage(i18n.English)
	if got, want := i18n.T("Refresh"), "Refresh"; got != want {
		t.Errorf("T(Refresh) = %q, want %q", got, want)
	}

	// An unknown key renders readable English rather than a placeholder.
	i18n.SetLanguage(i18n.Russian)
	if got, want := i18n.T("Not a real key"), "Not a real key"; got != want {
		t.Errorf("T(unknown) = %q, want %q", got, want)
	}
}

func TestTfAppliesArgumentsToTheTranslation(t *testing.T) {
	defer i18n.SetLanguage(i18n.English)

	i18n.SetLanguage(i18n.Russian)
	if got, want := i18n.Tf("Services: %d", 315), "Сервисов: 315"; got != want {
		t.Errorf("Tf = %q, want %q", got, want)
	}
}

func TestOnChangeFiresOnlyOnRealChanges(t *testing.T) {
	defer i18n.SetLanguage(i18n.English)

	i18n.SetLanguage(i18n.English)

	calls := 0
	i18n.OnChange(func() { calls++ })

	i18n.SetLanguage(i18n.Russian)
	if calls != 1 {
		t.Fatalf("switching en->ru fired %d watcher calls, want 1", calls)
	}

	// Re-selecting the active language must not rebuild the window.
	i18n.SetLanguage(i18n.Russian)
	if calls != 1 {
		t.Errorf("re-selecting the active language fired %d watcher calls, want 1", calls)
	}
}
