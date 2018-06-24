package kobo

import "testing"

func TestDeviceByID(t *testing.T) {
	for _, d := range Devices {
		if dd, _ := DeviceByID(d.ID); dd.ID != d.ID || dd.Hardware != d.Hardware || dd.Name != d.Name {
			t.Errorf("id '%s should be device %v", d.ID, d)
		}
	}
	if d, ok := DeviceByID("asdasd"); ok || d != nil {
		t.Errorf("id 'asdasd' should not return a device")
	}
}
