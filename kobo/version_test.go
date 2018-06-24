package kobo

import (
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
	td, err := ioutil.TempDir("", "koboutils")
	if err != nil {
		t.Fatalf("could not make temp dir: %v", err)
	}
	defer os.RemoveAll(td)

	err = os.Mkdir(filepath.Join(td, ".kobo"), 0755)
	if err != nil {
		t.Fatalf("could make fake .kobo dir: %v", err)
	}

	err = ioutil.WriteFile(filepath.Join(td, ".kobo", "version"), []byte("N345345345,3.0.35+,4.8.11073,3.0.35+,3.0.35+,00000000-0000-0000-0000-000000000375"), 0644)
	if err != nil {
		t.Fatalf("could not write fake version file: %v", err)
	}

	serial, version, id, err := ParseKoboVersion(td)
	if err != nil {
		t.Error(err)
	}

	if serial != "N345345345" || version != "4.8.11073" || id != "00000000-0000-0000-0000-000000000375" {
		t.Errorf("unexpected result: %s, %s, %s, %v", serial, version, id, err)
	}
}
