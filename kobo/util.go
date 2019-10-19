package kobo

import (
	"image"
	"strconv"
	"strings"
)

// resizeKeepAspectRatio resizes sz to fill bounds while keeping the aspect
// ratio. It is based on the code for QSize::scaled with the modes
// Qt::KeepAspectRatio and Qt::KeepAspectRatioByExpanding.
func resizeKeepAspectRatio(sz image.Point, bounds image.Point, expand bool) image.Point {
	if sz.X == 0 || sz.Y == 0 {
		return sz
	}

	var useHeight bool
	ar := float64(sz.X) / float64(sz.Y)
	rw := int(float64(bounds.Y) * ar)

	if !expand {
		useHeight = rw <= bounds.X
	} else {
		useHeight = rw >= bounds.X
	}

	if useHeight {
		return image.Pt(rw, bounds.Y)
	}
	return image.Pt(bounds.X, int(float64(bounds.X)/ar))
}

func strSplitInt(str string) []int64 {
	spl := strings.Split(str, ".")
	ints := make([]int64, len(spl))
	for i, p := range spl {
		ints[i], _ = strconv.ParseInt(p, 10, 64)
	}
	return ints
}
