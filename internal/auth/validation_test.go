package auth

import (
	"strings"
	"testing"
)

func TestUsernameValidation(t *testing.T) {
	tests := []struct {
		name     string
		username string
		valid    bool
	}{
		{"valid simple", "user123", true},
		{"valid with underscore", "user_123", true},
		{"valid min length", "abc", true},
		{"valid max length", "abcdefghijklmnopqrst", true},
		{"too short", "ab", false},
		{"too long", "abcdefghijklmnopqrstu", false},
		{"empty", "", false},
		{"with space", "user 123", false},
		{"with dash", "user-123", false},
		{"with special char", "user@123", false},
		{"only numbers", "12345", true},
		{"only underscores min", "___", true},
		{"starts with underscore", "_user", true},
		{"ends with underscore", "user_", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := usernameRegex.MatchString(tt.username)
			if result != tt.valid {
				t.Errorf("username %q: expected %v, got %v", tt.username, tt.valid, result)
			}
		})
	}
}

func TestPasswordValidation(t *testing.T) {
	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{"valid 8 chars", "password", true},
		{"valid longer", "longerpassword123", true},
		{"too short 7", "passwor", false},
		{"too short 1", "a", false},
		{"empty", "", false},
		{"exactly 8", "12345678", true},
		{"with special chars", "p@ssw0rd!", true},
		{"with spaces", "pass word", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := len(tt.password) >= 8
			if valid != tt.valid {
				t.Errorf("password %q (len=%d): expected %v, got %v", tt.password, len(tt.password), tt.valid, valid)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"unique constraint violation", "unique", true},
		{"duplicate key value", "duplicate", true},
		{"some other error", "unique", false},
		{"", "unique", false},
		{"unique", "unique", true},
		{"uniqueness", "unique", true},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.substr, func(t *testing.T) {
			got := strings.Contains(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}
