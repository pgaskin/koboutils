package kobo

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestVersionCompare(t *testing.T) {
	for _, c := range []struct {
		A, B string
		C    int
	}{
		{"1", "1", 0},
		{"1", "0", 1},
		{"0", "1", -1},
		{"0.1", "0.1", 0},
		{"0.1", "0.0", 1},
		{"0.0", "0.1", -1},
		{"3.4.1", "3.4.1", 0},
		{"3.4.1", "3.19.5761", -1},
		{"3.19.5761", "4.0.0", -1},
		{"0.sdf", "0.dfg", 0},
		{"0.1.sdf", "0.0.dfg", 1},
		{"0", "0.1", 0},
	} {
		r := VersionCompare(c.A, c.B)
		if r != c.C {
			t.Errorf("VersionCompare(%s, %s) should be %d, not %d", c.A, c.B, c.C, r)
		}

		ra := VersionCompare(c.B, c.A)
		if ra != c.C*-1 {
			t.Errorf("VersionCompare(%s, %s) should be %d, not %d", c.B, c.A, c.C*-1, ra)
		}
	}
}

func TestParseKoboVersion(t *testing.T) {
	if err := fakekobo(func(kpath string) {
		serial, version, id, err := ParseKoboVersion(kpath)
		if err != nil {
			t.Error(err)
		}

		if serial != "N345345345" || version != "4.8.11073" || id != "00000000-0000-0000-0000-000000000375" {
			t.Errorf("unexpected result: %s, %s, %s, %v", serial, version, id, err)
		}
	}); err != nil {
		t.Fatal(err)
	}
}

func TestParseKoboAffiliate(t *testing.T) {
	if err := fakekobo(func(kpath string) {
		aff, err := ParseKoboAffiliate(kpath)
		if err != nil {
			t.Error(err)
		} else if aff != "Kobo" {
			t.Errorf("expected Kobo, got %#v", aff)
		}
	}); err != nil {
		t.Fatal(err)
	}
}

func TestParseKoboUAString(t *testing.T) {
	tests := []struct {
		name         string
		ua           string
		expectedVers string
		expectedID   string
		expectedErr  error
	}{
		{
			name:         "Kobo H2O UA",
			ua:           "Mozilla/5.0 (Linux; U; Android 2.0; en-us;) AppleWebKit/538.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/538.1 (Kobo Touch 0370/4.20.14622)",
			expectedVers: "4.20.14622", expectedID: "00000000-0000-0000-0000-000000000370", expectedErr: nil,
		},
		{
			name:         "Desktop Chrome UA",
			ua:           "Mozilla/5.0 (Linux; Android 8.0.0; SM-G930F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.162 Mobile Safari/537.36",
			expectedVers: "", expectedID: "", expectedErr: errors.New("could not parse UA string"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vers, id, err := ParseKoboUAString(tt.ua)
			if vers != tt.expectedVers {
				t.Errorf("unexpected vers, want '%s' got '%s'\n", tt.expectedVers, vers)
			}
			if id != tt.expectedID {
				t.Errorf("unexpected ID, want '%s' got '%s'\n", tt.expectedID, id)
			}
			if (err != tt.expectedErr) && (err != nil && tt.expectedErr != nil) && (err.Error() != tt.expectedErr.Error()) {
				t.Errorf("unexpected err, want '%v' got '%v'\n", tt.expectedErr, err)
			}
		})
	}
}

func TestIsKobo(t *testing.T) {
	if err := fakekobo(func(kpath string) {
		if !IsKobo(kpath) {
			t.Errorf("expected fake kobo to be a kobo")
		}
	}); err != nil {
		t.Fatal(err)
	}
	if IsKobo(".") {
		t.Errorf("expected current dir not to be a kobo")
	}
}

func fakekobo(fn func(kpath string)) error {
	td, err := ioutil.TempDir("", "koboutils")
	if err != nil {
		return fmt.Errorf("could not make temp dir: %v", err)
	}
	defer os.RemoveAll(td)

	err = os.Mkdir(filepath.Join(td, ".kobo"), 0755)
	if err != nil {
		return fmt.Errorf("could make fake .kobo dir: %v", err)
	}

	// the actual file doesn't have a newline, but we're adding it here to test
	// trimming spaces (for people who edit the file manually)
	err = ioutil.WriteFile(filepath.Join(td, ".kobo", "version"), []byte("N345345345,3.0.35+,4.8.11073,3.0.35+,3.0.35+,00000000-0000-0000-0000-000000000375\n"), 0644)
	if err != nil {
		return fmt.Errorf("could not write fake version file: %v", err)
	}

	err = ioutil.WriteFile(filepath.Join(td, ".kobo", "affiliate.conf"), []byte("[General]\naffiliate=Kobo"), 0644)
	if err != nil {
		return fmt.Errorf("could not write fake affiliate file: %v", err)
	}

	fn(td)

	return nil
}
