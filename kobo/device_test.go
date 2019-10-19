package kobo

import (
	"fmt"
	"testing"
)

func TestDeviceList(t *testing.T) {
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

// TODO: more tests
