package kobo

import (
	"image"
	"path/filepath"
	"testing"
)

func TestPathContentIDImageID(t *testing.T) {
	for _, tc := range []struct {
		path string
		cid  string
		iid  string
	}{
		{
			"kepubify/Books_converted/Patrick Gaskin/Test Book 1.kepub.epub",
			"file:///mnt/onboard/kepubify/Books_converted/Patrick Gaskin/Test Book 1.kepub.epub",
			"file____mnt_onboard_kepubify_Books_converted_Patrick_Gaskin_Test_Book_1_kepub_epub",
		},
		{
			"kepubify/Books_converted/Patrick Gaskin/Test Book 2:.kepub.epub",
			"file:///mnt/onboard/kepubify/Books_converted/Patrick Gaskin/Test Book 2:.kepub.epub",
			"file____mnt_onboard_kepubify_Books_converted_Patrick_Gaskin_Test_Book_2__kepub_epub",
		},
	} {
		cid := PathToContentID(tc.path)
		iid := ContentIDToImageID(tc.cid)

		if cid != PathToContentID(filepath.FromSlash(tc.path)) {
			t.Errorf("incorrect native path separator conversion")
		}

		if tc.cid != cid {
			t.Errorf("cid of %#v: expected %#v, got %#v", tc.path, tc.cid, cid)
		}

		if tc.iid != iid {
			t.Errorf("iid of %#v: expected %#v, got %#v", tc.cid, tc.iid, iid)
		}
	}
}

func TestResizeKeepAspectRatioExpand(t *testing.T) {
	for _, tc := range []struct {
		sz     image.Point
		bounds image.Point
		rsz    image.Point
	}{
		// don't resize if width or height is zero
		{image.Pt(0, 0), image.Pt(0, 0), image.Pt(0, 0)},
		{image.Pt(1, 0), image.Pt(0, 0), image.Pt(1, 0)},
		{image.Pt(0, 1), image.Pt(0, 0), image.Pt(0, 1)},
		// same aspect ratio
		{image.Pt(1, 1), image.Pt(1, 1), image.Pt(1, 1)},
		{image.Pt(1, 1), image.Pt(5, 5), image.Pt(5, 5)},
		{image.Pt(5, 5), image.Pt(1, 1), image.Pt(1, 1)},
		// limited by width
		{image.Pt(2, 3), image.Pt(6, 6), image.Pt(6, 9)},
		{image.Pt(2, 4), image.Pt(6, 6), image.Pt(6, 12)},
		{image.Pt(6, 9), image.Pt(2, 3), image.Pt(2, 3)},
		{image.Pt(6, 12), image.Pt(2, 4), image.Pt(2, 4)},
		// limited by height
		{image.Pt(3, 2), image.Pt(6, 6), image.Pt(9, 6)},
		{image.Pt(4, 2), image.Pt(6, 6), image.Pt(12, 6)},
		{image.Pt(9, 6), image.Pt(3, 2), image.Pt(3, 2)},
		{image.Pt(12, 6), image.Pt(4, 2), image.Pt(4, 2)},
		// fractional stuff
		{image.Pt(1391, 2200), image.Pt(355, 530), image.Pt(355, 561)},
	} {
		if tsz := resizeKeepAspectRatio(tc.sz, tc.bounds, true); !tsz.Eq(tc.rsz) {
			t.Errorf("(%s, %s, %t): expected %s, got %s", tc.sz, tc.bounds, true, tc.rsz, tsz)
		}
	}
}

func TestResizeKeepAspectRatioShrink(t *testing.T) {
	for _, tc := range []struct {
		sz     image.Point
		bounds image.Point
		rsz    image.Point
	}{
		// don't resize if width or height is zero
		{image.Pt(0, 0), image.Pt(0, 0), image.Pt(0, 0)},
		{image.Pt(1, 0), image.Pt(0, 0), image.Pt(1, 0)},
		{image.Pt(0, 1), image.Pt(0, 0), image.Pt(0, 1)},
		// same aspect ratio
		{image.Pt(1, 1), image.Pt(1, 1), image.Pt(1, 1)},
		{image.Pt(1, 1), image.Pt(5, 5), image.Pt(5, 5)},
		{image.Pt(5, 5), image.Pt(1, 1), image.Pt(1, 1)},
		// limited by width
		{image.Pt(2, 3), image.Pt(6, 6), image.Pt(4, 6)},
		{image.Pt(2, 4), image.Pt(6, 6), image.Pt(3, 6)},
		{image.Pt(6, 9), image.Pt(2, 3), image.Pt(2, 3)},
		{image.Pt(6, 12), image.Pt(2, 4), image.Pt(2, 4)},
		// limited by height
		{image.Pt(3, 2), image.Pt(6, 6), image.Pt(6, 4)},
		{image.Pt(4, 2), image.Pt(6, 6), image.Pt(6, 3)},
		{image.Pt(9, 6), image.Pt(3, 2), image.Pt(3, 2)},
		{image.Pt(12, 6), image.Pt(4, 2), image.Pt(4, 2)},
		// fractional stuff
		{image.Pt(1391, 2200), image.Pt(355, 530), image.Pt(335, 530)},
	} {
		if tsz := resizeKeepAspectRatio(tc.sz, tc.bounds, false); !tsz.Eq(tc.rsz) {
			t.Errorf("(%s, %s, %t): expected %s, got %s", tc.sz, tc.bounds, false, tc.rsz, tsz)
		}
	}
}
