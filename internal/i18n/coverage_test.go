package i18n_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/Lucky2356/system-hub/internal/i18n"
)

// TestRussianCatalogCoversEveryKey parses the whole module, collects every
// literal passed to i18n.T / i18n.Tf, and asserts the Russian catalog has an
// entry for it.
//
// Without this, forgetting a translation is invisible: T falls back to the key,
// so the string silently renders in English inside an otherwise Russian UI.
func TestRussianCatalogCoversEveryKey(t *testing.T) {
	keys, err := collectTranslationKeys(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("collect keys: %v", err)
	}
	if len(keys) == 0 {
		t.Fatal("no i18n.T calls found — the collector is broken, not the catalog")
	}

	i18n.SetLanguage(i18n.Russian)
	defer i18n.SetLanguage(i18n.English)

	var missing []string
	for _, key := range keys {
		if i18n.T(key) == key {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		t.Errorf("%d key(s) missing from translation/ru.json:", len(missing))
		for _, key := range missing {
			t.Errorf("  %q", key)
		}
	}
}

// TestRussianCatalogHasNoUnusedKeys is the other direction: an entry nobody
// looks up is dead weight, and usually means the source string was reworded
// while the catalog kept the old key.
func TestRussianCatalogHasNoUnusedKeys(t *testing.T) {
	keys, err := collectTranslationKeys(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("collect keys: %v", err)
	}

	used := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		used[key] = struct{}{}
	}

	for _, key := range i18n.CatalogKeys(i18n.Russian) {
		if _, ok := used[key]; !ok {
			t.Errorf("translation/ru.json defines %q, which no i18n.T call uses", key)
		}
	}
}

// TestFormatVerbsMatchTheKey guards the Tf keys: a translation that drops or
// reorders a verb renders as %!d(MISSING) at runtime, and only on that
// language.
func TestFormatVerbsMatchTheKey(t *testing.T) {
	i18n.SetLanguage(i18n.Russian)
	defer i18n.SetLanguage(i18n.English)

	for _, key := range i18n.CatalogKeys(i18n.Russian) {
		want := formatVerbs(key)
		got := formatVerbs(i18n.T(key))

		if len(want) != len(got) {
			t.Errorf("%q has verbs %v but its translation has %v", key, want, got)
			continue
		}
		for i := range want {
			if want[i] != got[i] {
				t.Errorf("%q has verbs %v but its translation has %v", key, want, got)
				break
			}
		}
	}
}

// formatVerbs extracts the fmt verbs from a format string, treating "%%" as a
// literal percent rather than a verb.
func formatVerbs(s string) []string {
	var verbs []string

	for i := 0; i < len(s); i++ {
		if s[i] != '%' || i+1 >= len(s) {
			continue
		}
		if s[i+1] == '%' {
			i++
			continue
		}

		j := i + 1
		for j < len(s) && strings.ContainsRune(".0123456789+-# ", rune(s[j])) {
			j++
		}
		if j < len(s) {
			verbs = append(verbs, s[i:j+1])
			i = j
		}
	}
	return verbs
}

// collectTranslationKeys walks the module and returns the literal first
// argument of every i18n.T / i18n.Tf call. Non-literal arguments are skipped:
// they cannot be checked statically.
func collectTranslationKeys(root string) ([]string, error) {
	seen := map[string]struct{}{}
	var keys []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if name := d.Name(); name == ".git" || name == "dist" || name == "build" {
				return fs.SkipDir
			}
			return nil
		}
		// Test files are skipped: they pass throwaway keys to T on purpose, and
		// those must not be mistaken for strings the UI shows.
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Parse every file regardless of build tags, so keys used only by the
		// Windows or Linux provider are checked on both platforms.
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}

		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 {
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkg, ok := sel.X.(*ast.Ident)
			if !ok || pkg.Name != "i18n" {
				return true
			}
			if sel.Sel.Name != "T" && sel.Sel.Name != "Tf" {
				return true
			}

			key, ok := stringLiteral(call.Args[0])
			if !ok {
				return true
			}
			if _, dup := seen[key]; !dup {
				seen[key] = struct{}{}
				keys = append(keys, key)
			}
			return true
		})

		return nil
	})

	return keys, err
}

// stringLiteral resolves a plain literal or a concatenation of literals
// ("a" + "b"), which is how the longer multi-line messages are written.
func stringLiteral(expr ast.Expr) (string, bool) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind != token.STRING {
			return "", false
		}
		value, err := strconv.Unquote(e.Value)
		if err != nil {
			return "", false
		}
		return value, true

	case *ast.BinaryExpr:
		if e.Op != token.ADD {
			return "", false
		}
		left, ok := stringLiteral(e.X)
		if !ok {
			return "", false
		}
		right, ok := stringLiteral(e.Y)
		if !ok {
			return "", false
		}
		return left + right, true
	}
	return "", false
}
