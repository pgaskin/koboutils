package kobo

import "testing"

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
