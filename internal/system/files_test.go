package system

import "testing"

func TestSanitizePath(t *testing.T) {
	abs := t.TempDir() // absolute, OS-appropriate

	if got, err := sanitizePath(abs); err != nil || got != abs {
		t.Errorf("sanitizePath(%q) = %q, %v; want %q, nil", abs, got, err, abs)
	}

	// Unix-style rooted paths must be accepted on every OS: filepath.IsAbs is
	// false for these on Windows, and rejecting them broke the file browser's
	// preset paths there.
	if _, err := sanitizePath("/etc/systemd/system"); err != nil {
		t.Errorf("sanitizePath(%q) unexpected error: %v", "/etc/systemd/system", err)
	}

	if _, err := sanitizePath(""); err == nil {
		t.Error("expected error for empty path")
	}
	if _, err := sanitizePath("relative/dir"); err == nil {
		t.Error("expected error for relative path")
	}
	if _, err := sanitizePath("with\x00nul"); err == nil {
		t.Error("expected error for NUL byte")
	}
}

func TestGetPresetPathsAreRooted(t *testing.T) {
	for _, p := range GetPresetPaths() {
		if p.Path == "" {
			continue // home dir may be unavailable in a sandbox
		}
		if _, err := sanitizePath(p.Path); err != nil {
			t.Errorf("preset %q (%s) rejected by sanitizePath: %v", p.Title, p.Path, err)
		}
	}
}

func TestIsTextContent(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{"empty", []byte{}, true},
		{"plain text", []byte("hello world\n"), true},
		{"tabs and crlf", []byte("a\tb\r\n"), true},
		{"nul byte", []byte("abc\x00def"), false},
		{"control byte", []byte{0x01, 0x02}, false},
	}
	for _, tt := range tests {
		if got := isTextContent(tt.data); got != tt.want {
			t.Errorf("%s: isTextContent = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.00 KB"},
		{1024 * 1024, "1.00 MB"},
		{1024 * 1024 * 1024, "1.00 GB"},
	}
	for _, tt := range tests {
		if got := FormatFileSize(tt.in); got != tt.want {
			t.Errorf("FormatFileSize(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
