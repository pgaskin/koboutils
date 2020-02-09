package kobo

import (
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
