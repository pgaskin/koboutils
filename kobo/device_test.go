package kobo

import (
	"fmt"
	"image"
	"reflect"
	"testing"
)

func TestDeviceList(t *testing.T) {
	// check this manually (automatically doing this would just be a duplicate of tbe info)
	for _, d := range Devices() {
		fmt.Printf("Device %d (%s):\n  Family: %s (%s)\n  Hardware: %s\n  IDString: %s\n  Storage: %dGB\n  CodeNames: %s\n  Cover Types:\n", int(d), d.Name(), d.Family(), d.CodeNames().Family(), d.Hardware(), d.IDString(), d.StorageGB(), d.CodeNames())
		for _, c := range CoverTypes() {
			fmt.Printf("    %s: %s\n", c, d.CoverSize(c))
		}
		fmt.Println()
	}
	if d, ok := DeviceByID("asdasd"); ok || d != 0 {
		t.Errorf("id 'asdasd' should not return a device")
	}
}

func TestCoverGeneratePath(t *testing.T) {
	for _, tc := range []struct {
		ct  CoverType
		ext bool
		iid string
		out string
	}{
		// note: the image ids and content ids are already tested, no need to do that here; just test the hashing and file path
		{
			CoverTypeFull, false,
			"file____mnt_onboard_kepubify_Books_converted_Patrick_Gaskin_Test_Book_1_kepub_epub",
			".kobo-images/210/143/file____mnt_onboard_kepubify_Books_converted_Patrick_Gaskin_Test_Book_1_kepub_epub - N3_FULL.parsed",
		},
		{
			CoverTypeLibFull, false,
			"file____mnt_onboard_kepubify_Books_converted_Patrick_Gaskin_Test_Book_1_kepub_epub",
			".kobo-images/210/143/file____mnt_onboard_kepubify_Books_converted_Patrick_Gaskin_Test_Book_1_kepub_epub - N3_LIBRARY_FULL.parsed",
		},
		{
			CoverTypeLibGrid, false,
			"file____mnt_onboard_kepubify_Books_converted_Patrick_Gaskin_Test_Book_1_kepub_epub",
			".kobo-images/210/143/file____mnt_onboard_kepubify_Books_converted_Patrick_Gaskin_Test_Book_1_kepub_epub - N3_LIBRARY_GRID.parsed",
		},
		{
			CoverTypeLibList, false,
			"file____mnt_onboard_kepubify_Books_converted_Patrick_Gaskin_Test_Book_1_kepub_epub",
			".kobo-images/210/143/file____mnt_onboard_kepubify_Books_converted_Patrick_Gaskin_Test_Book_1_kepub_epub - N3_LIBRARY_LIST.parsed",
		},
		{
			CoverTypeLibList, true,
			"file____mnt_onboard_kepubify_Books_converted_Patrick_Gaskin_Test_Book_2__kepub_epub",
			"koboExtStorage/images-cache/82/246/file____mnt_onboard_kepubify_Books_converted_Patrick_Gaskin_Test_Book_2__kepub_epub - N3_LIBRARY_LIST.parsed",
		},
	} {
		if path := tc.ct.GeneratePath(tc.ext, tc.iid); path != tc.out {
			t.Errorf("(%s, ext: %t, iid: %#v): expected %#v, got %#v", tc.ct, tc.ext, tc.iid, tc.out, path)
		}
	}
}

func TestSwitchCases(t *testing.T) {
	for _, d := range Devices() {
		for _, fn := range []interface{}{
			d.CodeNames,
			d.Family,
			d.Hardware,
			d.ID,
			d.IDString,
			d.Name,
			d.StorageGB,
			d.String,
			d.DisplayPPI,
		} {
			if panics(fn) {
				t.Errorf("%s: %s panics", d, reflect.ValueOf(fn))
			}
		}
		for _, ct := range CoverTypes() {
			if panics(func() image.Point { return d.CoverSize(ct) }) {
				t.Errorf("%s: CoverSize panics for %s", d, ct)
			}
		}
	}
}

func panics(fn interface{}) (panicked bool) {
	v := reflect.ValueOf(fn)
	if t := v.Type(); t.Kind() != reflect.Func {
		panic("not a func")
	} else if t.NumIn() != 0 {
		panic("func requires args")
	}

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			panicked = true
		}
	}()
	v.Call(nil)

	return false
}
