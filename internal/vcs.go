package internal

import (
	"runtime/debug"
	"strconv"
	"time"
)

var vcs struct {
	revision string
	time     time.Time
	modified bool
}

func init() {
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				if v := s.Value; len(v) >= 20 {
					vcs.revision = v
				}
			case "vcs.time":
				if v, err := time.ParseInLocation(time.RFC3339Nano, s.Value, time.UTC); err == nil {
					vcs.time = v
				}
			case "vcs.modified":
				if v, err := strconv.ParseBool(s.Value); err == nil {
					vcs.modified = v
				}
			}
		}
	}
}

func VCS() (revision string, time time.Time, modified bool) {
	return vcs.revision, vcs.time, vcs.modified
}

func Version() string {
	if vcs.revision == "" {
		return "unknown"
	}
	if vcs.modified {
		return vcs.revision[:7] + " (dirty)"
	}
	return vcs.revision[:7]
}

func VersionName() string {
	return "koboutils " + Version()
}
