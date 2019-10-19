package kobo

import (
	"fmt"
	"image"
	"path/filepath"
	"strconv"
	"strings"
)

// PathToContentID generates the Kobo ContentId for a path relative to the
// internal storage root (slashes are converted to forward slashes automatically).
func PathToContentID(relpath string) string {
	return fmt.Sprintf("file:///mnt/onboard/%s", filepath.ToSlash(relpath))
}

// ContentIDToImageID converts the Kobo ContentId to the ImageId.
func ContentIDToImageID(contentID string) string {
	return strings.NewReplacer(
		" ", "_",
		"/", "_",
		":", "_",
		".", "_",
	).Replace(contentID)
}

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

// hashedImageParts returns the parts needed for constructing the path to the
// cached image. The result can be applied like:
// .kobo-images/{dir1}/{dir2}/{basename} - N3_SOMETHING.jpg
func hashedImageParts(imageID string) (dir1, dir2, basename string) {
	imgID := []byte(imageID)
	h := uint32(0x00000000)
	for _, x := range imgID {
		h = (h << 4) + uint32(x)
		h ^= (h & 0xf0000000) >> 23
		h &= 0x0fffffff
	}
	return fmt.Sprintf("%d", h&(0xff*1)), fmt.Sprintf("%d", (h&(0xff00*1))>>8), imageID
}

func strSplitInt(str string) []int64 {
	spl := strings.Split(str, ".")
	ints := make([]int64, len(spl))
	for i, p := range spl {
		ints[i], _ = strconv.ParseInt(p, 10, 64)
	}
	return ints
}
