package domain

import "testing"

func TestNormalizeNIM(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "trims and uppercase", in: "  ab12  ", want: "AB12"},
		{name: "remove inner spaces", in: "20 24 001", want: "2024001"},
		{name: "tabs and new lines", in: "\t19 98\n77", want: "199877"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := NormalizeNIM(tc.in); got != tc.want {
				t.Fatalf("NormalizeNIM(%q) = %q; want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	t.Parallel()

	valid := []string{"A1B2C3D4", "1234ABCD", "ZZZZ9999"}
	for _, token := range valid {
		if err := ValidateToken(token); err != nil {
			t.Fatalf("ValidateToken(%q) returned error: %v", token, err)
		}
	}

	invalid := []string{"A1B2C3D", "a1b2c3d4", "ABCD-123", "ABCD123*"}
	for _, token := range invalid {
		if err := ValidateToken(token); err == nil {
			t.Fatalf("ValidateToken(%q) expected error", token)
		}
	}
}
