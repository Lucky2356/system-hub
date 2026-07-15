package system

import "testing"

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"simple", "nginx.service", false},
		{"with dashes inside", "my-app.service", false},
		{"docker image with tag", "nginx:1.25", false},
		{"empty", "", true},
		{"whitespace only", "   ", true},
		{"leading dash flag", "--version", true},
		{"single dash flag", "-H", true},
		{"internal space", "foo bar", true},
		{"tab", "foo\tbar", true},
		{"control char", "foo\x00bar", true},
		{"newline", "foo\nbar", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateName("service", tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateName(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
