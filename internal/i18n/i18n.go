// Package i18n provides the application's translations.
//
// Keys are the English source strings, so untranslated code still renders
// readable English and a missing entry degrades to the key rather than to a
// placeholder.
//
// Fyne ships fyne.io/fyne/v2/lang, but its localizer is driven exclusively by
// the OS locale: setupLang is unexported and updateLocalizer always re-reads
// locale.GetLocales(). There is no supported way to force a language, which the
// in-app switcher requires — so the lookup below is our own. Fyne's lang is
// still used for one thing it does well: detecting the system locale.
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"fyne.io/fyne/v2/lang"
)

// Language codes accepted by SetLanguage and stored in the config.
const (
	Auto    = "auto"
	Russian = "ru"
	English = "en"
)

//go:embed translation/*.json
var translationFS embed.FS

var (
	mu       sync.RWMutex
	catalog  = map[string]map[string]string{}
	current  = English
	watchers []func()
)

func init() {
	entries, err := translationFS.ReadDir("translation")
	if err != nil {
		panic("i18n: read translations: " + err.Error())
	}

	for _, entry := range entries {
		data, err := translationFS.ReadFile("translation/" + entry.Name())
		if err != nil {
			panic("i18n: read " + entry.Name() + ": " + err.Error())
		}

		var messages map[string]string
		if err := json.Unmarshal(data, &messages); err != nil {
			panic("i18n: parse " + entry.Name() + ": " + err.Error())
		}

		catalog[strings.TrimSuffix(entry.Name(), ".json")] = messages
	}
}

// Supported reports the language codes that have a translation catalog, plus
// English, which needs none because the keys are already English.
func Supported() []string {
	return []string{Russian, English}
}

// catalogKeys returns the keys defined for a language. It exists for the
// coverage test, which asserts the catalog and the source agree in both
// directions.
func catalogKeys(language string) []string {
	mu.RLock()
	defer mu.RUnlock()

	messages := catalog[language]
	keys := make([]string, 0, len(messages))
	for key := range messages {
		keys = append(keys, key)
	}
	return keys
}

// SetLanguage switches the active language and notifies watchers. The code is
// "auto" (follow the OS locale), "ru" or "en"; anything else falls back to
// English. It returns the resolved code, which is never "auto".
func SetLanguage(code string) string {
	resolved := resolve(code)

	mu.Lock()
	changed := resolved != current
	current = resolved
	listeners := append([]func(){}, watchers...)
	mu.Unlock()

	if changed {
		for _, fn := range listeners {
			fn()
		}
	}
	return resolved
}

// Language returns the active language code ("ru" or "en").
func Language() string {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// resolve maps a config value onto a concrete language code.
func resolve(code string) string {
	switch strings.ToLower(strings.TrimSpace(code)) {
	case Russian:
		return Russian
	case English:
		return English
	case Auto:
		return systemLanguage()
	default:
		return English
	}
}

// systemLanguage picks a supported language from the OS locale, defaulting to
// English for locales we do not translate.
func systemLanguage() string {
	locale := strings.ToLower(lang.SystemLocale().LanguageString())
	if strings.HasPrefix(locale, Russian) {
		return Russian
	}
	return English
}

// OnChange registers a callback fired whenever the language actually changes.
// It is used to rebuild the window so switching applies without a restart.
func OnChange(fn func()) {
	mu.Lock()
	defer mu.Unlock()
	watchers = append(watchers, fn)
}

// T translates an English source string.
func T(key string) string {
	mu.RLock()
	defer mu.RUnlock()

	if messages, ok := catalog[current]; ok {
		if translated, ok := messages[key]; ok && translated != "" {
			return translated
		}
	}
	return key
}

// Tf translates a format string and applies the arguments. The placeholders
// live inside the translation, so a language may reorder them with explicit
// argument indexes.
func Tf(key string, args ...any) string {
	return fmt.Sprintf(T(key), args...)
}
