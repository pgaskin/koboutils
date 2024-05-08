package main

import (
	"cmp"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Version represents a firmware version.
type Version struct {
	Major int
	Minor int
	Patch int // is just the build number on v4 and likely v5 too
}

// ParseVersion parses s, ensuring it is canonical.
func ParseVersion(s string) (Version, error) {
	spl := strings.Split(s, ".")
	if len(spl) != 3 {
		return Version{}, fmt.Errorf("more than 3 components in version %q", s)
	}

	major, err := strconv.ParseInt(spl[0], 10, 0)
	if err == nil && major <= 0 {
		err = errors.New("major must be gt 1")
	}
	if err != nil {
		return Version{}, fmt.Errorf("invalid version %q: %w", s, err)
	}

	minor, err := strconv.ParseInt(spl[1], 10, 0)
	if err == nil && minor < 0 {
		err = errors.New("minor must be ge 1")
	}
	if err != nil {
		return Version{}, fmt.Errorf("invalid version %q: %w", s, err)
	}

	patch, err := strconv.ParseInt(spl[2], 10, 0)
	if err == nil && patch < 0 {
		err = errors.New("patch must be ge 1")
	}
	if err != nil {
		return Version{}, fmt.Errorf("invalid version %q: %w", s, err)
	}

	v := Version{
		Major: int(major),
		Minor: int(minor),
		Patch: int(patch),
	}
	if v.String() != s {
		return Version{}, fmt.Errorf("non-canonical version %q", s)
	}
	return v, nil
}

func (v Version) IsZero() bool {
	return v.Major == 0 && v.Minor == 0 && v.Patch == 0
}

func (v Version) String() string {
	return strconv.Itoa(v.Major) + "." + strconv.Itoa(v.Minor) + "." + strconv.Itoa(v.Patch)
}

func (v Version) Less(o Version) bool {
	return v.Compare(o) < 0
}

func (v Version) Compare(o Version) int {
	if v.Major != o.Major {
		return cmp.Compare(v.Major, o.Major)
	}
	if v.Minor != o.Minor {
		return cmp.Compare(v.Minor, o.Minor)
	}
	if v.Patch != o.Patch {
		return cmp.Compare(v.Patch, o.Patch)
	}
	return 0
}
