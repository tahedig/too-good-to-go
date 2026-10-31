package validation

import "testing"

func TestNormalizePhone(t *testing.T) {
	valid := map[string]string{
		"+989123456789":     "+989123456789",
		"09123456789":       "+989123456789",
		"989123456789":      "+989123456789",
		"00989123456789":    "+989123456789",
		" 0912 345-67-89 ":  "+989123456789",
		"(0912) 345 6789":   "+989123456789",
		"+14155552671":      "+14155552671",
		"0044 20 7946 0958": "+442079460958",
	}
	for in, want := range valid {
		got, err := NormalizePhone(in)
		if err != nil || got != want {
			t.Errorf("NormalizePhone(%q) = %q, %v; want %q", in, got, err, want)
		}
	}

	for _, in := range []string{"", "abc", "12345", "0912", "+0123456789", "+98912345678901234", "0912345678a"} {
		if got, err := NormalizePhone(in); err == nil {
			t.Errorf("NormalizePhone(%q) = %q, want error", in, got)
		}
	}
}

func TestMaskPhone(t *testing.T) {
	cases := map[string]string{
		"+989123456789": "+98******6789",
		"+12345678":     "+12**5678",
		"short":         "***",
	}
	for in, want := range cases {
		if got := MaskPhone(in); got != want {
			t.Errorf("MaskPhone(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIsNumericCode(t *testing.T) {
	cases := []struct {
		code   string
		length int
		want   bool
	}{
		{"123456", 6, true},
		{"000000", 6, true},
		{"12345", 6, false},
		{"1234567", 6, false},
		{"12a456", 6, false},
		{"١٢٣٤٥٦", 6, false}, // Arabic-Indic digits are rejected
		{"", 6, false},
	}
	for _, tc := range cases {
		if got := IsNumericCode(tc.code, tc.length); got != tc.want {
			t.Errorf("IsNumericCode(%q, %d) = %v, want %v", tc.code, tc.length, got, tc.want)
		}
	}
}
