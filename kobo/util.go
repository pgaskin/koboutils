package kobo

import (
	"strconv"
	"strings"
)

func strSplitInt(str string) []int64 {
	spl := strings.Split(str, ".")
	ints := make([]int64, len(spl))
	for i, p := range spl {
		ints[i], _ = strconv.ParseInt(p, 10, 64)
	}
	return ints
}
