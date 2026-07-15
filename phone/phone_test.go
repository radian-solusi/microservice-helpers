package phone

import "testing"

func TestValidate(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"0812345678", "62812345678", false},
		{"62812345678", "62812345678", false},
		{"812345678", "62812345678", false},
		{"", "", true},
		{"0812", "", true},                // too short
		{"0812345678901234567", "", true}, // too long
		{"08123abc78", "", true},          // non-digit
		{"0912345678", "", true},          // third digit 9 not allowed
		{"62112345678", "", true},         // third digit 1 not allowed
	}
	for _, c := range cases {
		got, err := Validate(c.in)
		if (err != nil) != c.wantErr {
			t.Errorf("Validate(%q) err = %v, wantErr %v", c.in, err, c.wantErr)
		}
		if err == nil && got != c.want {
			t.Errorf("Validate(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNormalize(t *testing.T) {
	cases := []struct{ in, want string }{
		{"+62 812 3456", "08123456"},
		{"628123456789", "08123456789"},
		{"+6281234", "081234"},
		{"5551234", "5551234"}, // no known prefix
		{"", ""},
		{"  +62812  ", "0812"},
	}
	for _, c := range cases {
		if got := Normalize(c.in); got != c.want {
			t.Errorf("Normalize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
