package kobo

import (
	"fmt"
	"image"
)

// See https://gist.github.com/geek1011/613b34c23f026f7c39c50ee32f5e167e and
// https://github.com/shermp/Kobo-UNCaGED/issues/16

type Device int
type Hardware int
type CodeName string
type CodeNameTriplet [3]CodeName
type CoverType string

// Devices (not including really old ones, like Kobo eReader, Wireless, Literati, and Vox).
const (
	DeviceTouchAB               Device = 310
	DeviceTouchC                Device = 320
	DeviceGlo                   Device = 330
	DeviceMini                  Device = 340
	DeviceAuraHD                Device = 350
	DeviceAura                  Device = 360
	DeviceAuraH2O               Device = 370
	DeviceGloHD                 Device = 371
	DeviceTouch2                Device = 372
	DeviceAuraONE               Device = 373
	DeviceAuraH2OEdition2v1     Device = 374
	DeviceAuraEdition2v1        Device = 375
	DeviceClaraHD               Device = 376
	DeviceForma                 Device = 377
	DeviceAuraH2OEdition2v2     Device = 378
	DeviceAuraEdition2v2        Device = 379
	DeviceForma32               Device = 380
	DeviceAuraONELimitedEdition Device = 381
	DeviceLibraH2O              Device = 384
)

// Hardware revisions.
const (
	HardwareKobo3 Hardware = 3
	HardwareKobo4 Hardware = 4
	HardwareKobo5 Hardware = 5
	HardwareKobo6 Hardware = 6
	HardwareKobo7 Hardware = 7
)

// Codenames.
const (
	CodeNameNone          CodeName = ""
	CodeNameDesktop       CodeName = "desktop"
	CodeNameNickel1       CodeName = "nickel1"
	CodeNameNickel2       CodeName = "nickel2"
	CodeNameMerch         CodeName = "merch"
	CodeNameVox           CodeName = "vox"
	CodeNameTrilogy       CodeName = "trilogy"
	CodeNamePixie         CodeName = "pixie"
	CodeNamePika          CodeName = "pika"
	CodeNameDragon        CodeName = "dragon"
	CodeNameDahlia        CodeName = "dahlia"
	CodeNameAlyssum       CodeName = "alyssum"
	CodeNameSnow          CodeName = "snow"
	CodeNameNova          CodeName = "nova"
	CodeNameStorm         CodeName = "storm"
	CodeNameDaylight      CodeName = "daylight"
	CodeNameSuperDaylight CodeName = "superDaylight"
	CodeNameFrost         CodeName = "frost"
	CodeNameFrost32       CodeName = "frost32"
	CodeNamePhoenix       CodeName = "phoenix"
	CodeNameKraken        CodeName = "kraken"
	CodeNameStar          CodeName = "star"
)

// Cover types.
const (
	CoverTypeLibFull CoverType = "N3_LIBRARY_FULL"
	CoverTypeLibList CoverType = "N3_LIBRARY_LIST"
	CoverTypeLibGrid CoverType = "N3_LIBRARY_GRID"
)

func Devices() []Device {
	return []Device{DeviceTouchAB, DeviceTouchC, DeviceGlo, DeviceMini, DeviceAuraHD, DeviceAura, DeviceAuraH2O, DeviceGloHD, DeviceTouch2, DeviceAuraONE, DeviceAuraH2OEdition2v1, DeviceAuraEdition2v1, DeviceClaraHD, DeviceForma, DeviceAuraH2OEdition2v2, DeviceAuraEdition2v2, DeviceForma32, DeviceAuraONELimitedEdition, DeviceLibraH2O}
}

func CoverTypes() []CoverType {
	return []CoverType{CoverTypeLibFull, CoverTypeLibList, CoverTypeLibGrid}
}

func DeviceByID(id string) (Device, bool) {
	for _, device := range Devices() {
		if device.IDString() == id {
			return device, true
		}
	}
	return 0, false
}

func (d Device) ID() int {
	return int(d)
}

func (d Device) IDString() string {
	return fmt.Sprintf("00000000-0000-0000-0000-%012d", d.ID())
}

func (d Device) String() string {
	return d.Name()
}

func (d Device) Name() string {
	cd := d.CodeNames()
	dev := cd.FamilyString()
	if sec := cd.SecondaryString(); sec != "" {
		dev += " " + sec
	}
	switch d {
	case DeviceTouchAB:
		dev += " A/B"
	case DeviceTouchC:
		dev += " C"
	case DeviceAuraEdition2v1, DeviceAuraEdition2v2:
		dev += " Edition 2"
	}
	switch d {
	case DeviceAuraEdition2v1, DeviceAuraH2OEdition2v1:
		dev += " v1"
	case DeviceAuraEdition2v2, DeviceAuraH2OEdition2v2:
		dev += " v2"
	}
	return dev
}

func (d Device) Hardware() Hardware {
	switch d {
	case DeviceTouchAB:
		return HardwareKobo3
	case DeviceTouchC, DeviceMini, DeviceGlo, DeviceAuraHD:
		return HardwareKobo4
	case DeviceAura, DeviceAuraH2O:
		return HardwareKobo5
	case DeviceGloHD, DeviceTouch2, DeviceAuraH2OEdition2v1, DeviceAuraONE, DeviceAuraONELimitedEdition, DeviceAuraEdition2v1:
		return HardwareKobo6
	case DeviceAuraH2OEdition2v2, DeviceAuraEdition2v2, DeviceClaraHD, DeviceForma, DeviceForma32, DeviceLibraH2O:
		return HardwareKobo7
	}
	panic("unknown device")
}

func (h Hardware) Hardware() int {
	return int(h)
}

func (h Hardware) String() string {
	return fmt.Sprintf("kobo%d", int(h))
}

func (d Device) CodeNames() CodeNameTriplet {
	switch d {
	case DeviceTouchAB, DeviceTouchC:
		return CodeNameTriplet{CodeNameTrilogy, CodeNameTrilogy, CodeNameNone}
	case DeviceMini:
		return CodeNameTriplet{CodeNameTrilogy, CodeNamePixie, CodeNameNone}
	case DeviceTouch2:
		return CodeNameTriplet{CodeNameTrilogy, CodeNamePika, CodeNameNone}

	case DeviceAuraHD:
		return CodeNameTriplet{CodeNameDragon, CodeNameDragon, CodeNameNone}
	case DeviceAuraH2O:
		return CodeNameTriplet{CodeNameDragon, CodeNameDahlia, CodeNameNone}
	case DeviceGloHD:
		return CodeNameTriplet{CodeNameDragon, CodeNameAlyssum, CodeNameNone}
	case DeviceAuraH2OEdition2v1, DeviceAuraH2OEdition2v2:
		return CodeNameTriplet{CodeNameDragon, CodeNameSnow, CodeNameNone}
	case DeviceClaraHD:
		return CodeNameTriplet{CodeNameDragon, CodeNameNova, CodeNameNone}
	case DeviceLibraH2O:
		return CodeNameTriplet{CodeNameDragon, CodeNameStorm, CodeNameNone}

	case DeviceAuraONE:
		return CodeNameTriplet{CodeNameDaylight, CodeNameDaylight, CodeNameNone}
	case DeviceAuraONELimitedEdition:
		return CodeNameTriplet{CodeNameDaylight, CodeNameDaylight, CodeNameSuperDaylight}
	case DeviceForma:
		return CodeNameTriplet{CodeNameDaylight, CodeNameFrost, CodeNameNone}
	case DeviceForma32:
		return CodeNameTriplet{CodeNameDaylight, CodeNameFrost, CodeNameFrost32}

	case DeviceAura:
		return CodeNameTriplet{CodeNamePhoenix, CodeNamePhoenix, CodeNameNone}
	case DeviceGlo:
		return CodeNameTriplet{CodeNamePhoenix, CodeNameKraken, CodeNameNone}
	case DeviceAuraEdition2v1, DeviceAuraEdition2v2:
		return CodeNameTriplet{CodeNamePhoenix, CodeNameStar, CodeNameNone}
	}
	panic("unknown device")
}

func (c CodeName) String() string {
	return string(c)
}

func (c CodeNameTriplet) String() string {
	if c[2] != CodeNameNone {
		return fmt.Sprintf("class=%s family=%s secondary=%s", c[0], c[1], c[2])
	}
	return fmt.Sprintf("class=%s family=%s", c[0], c[1])
}

func (d Device) Family() string {
	return d.CodeNames().FamilyString()
}

func (c CodeNameTriplet) Class() CodeName {
	return c[0]
}

func (c CodeNameTriplet) Family() CodeName {
	return c[1]
}

func (c CodeNameTriplet) FamilyString() string {
	switch c.Family() {
	case CodeNameDesktop:
		return "Kobo Desktop"
	case CodeNameNickel1:
		return "Kobo eReader"
	case CodeNameNickel2:
		return "Kobo Wireless eReader"
	case CodeNameMerch:
		return "Literati / LookBook eReader"
	case CodeNameVox:
		return "Kobo Vox"
	case CodeNameTrilogy:
		return "Kobo Touch"
	case CodeNamePixie:
		return "Kobo Mini"
	case CodeNamePika:
		return "Kobo Touch 2.0"
	case CodeNameDragon:
		return "Kobo Aura HD"
	case CodeNameDahlia:
		return "Kobo Aura H2O"
	case CodeNameAlyssum:
		return "Kobo Glo HD"
	case CodeNameSnow:
		return "Kobo Aura H2O Edition 2"
	case CodeNameNova:
		return "Kobo Clara HD"
	case CodeNameStorm:
		return "Kobo Libra H2O"
	case CodeNameDaylight:
		return "Kobo Aura ONE"
	case CodeNameFrost:
		return "Kobo Forma"
	case CodeNamePhoenix:
		return "Kobo Aura"
	case CodeNameKraken:
		return "Kobo Glo"
	case CodeNameStar:
		return "Kobo Aura"
	}
	panic("unknown family")
}

func (c CodeNameTriplet) Secondary() CodeName {
	return c[2]
}

func (c CodeNameTriplet) SecondaryString() string {
	switch c.Secondary() {
	case CodeNameNone:
		return ""
	case CodeNameSuperDaylight:
		return "Limited Edition"
	case CodeNameFrost32:
		return "32GB"
	}
	panic("unknown secondary")
}

func (d Device) CoverSize(t CoverType) image.Point {
	if t == CoverTypeLibList {
		return image.Pt(60, 90)
	} else if t == CoverTypeLibGrid {
		return image.Pt(149, 223)
	} else if t != CoverTypeLibFull {
		panic("unknown cover type")
	}

	switch d.CodeNames().Family() {
	case CodeNameDragon, CodeNameSnow:
		return image.Pt(1080, 1440)
	case CodeNameDahlia:
		return image.Pt(1080, 1429)
	case CodeNameAlyssum, CodeNameNova:
		return image.Pt(1072, 1448)
	case CodeNameStorm:
		return image.Pt(1264, 1680)
	case CodeNameDaylight:
		return image.Pt(1404, 1872)
	case CodeNameFrost:
		return image.Pt(1440, 1920)
	case CodeNamePhoenix:
		return image.Pt(758, 1014)
	case CodeNameKraken, CodeNameStar:
		return image.Pt(758, 1024)
	default:
		return image.Pt(600, 800)
	}
}

func (d Device) CoverSized(t CoverType, orig image.Point) image.Point {
	return t.Resize(d.CoverSize(t), orig)
}

func (c CoverType) NickelString() string {
	return string(c)
}

func (c CoverType) String() string {
	return c.NickelString()
}

// Resize returns the dimensions to resize sz to for the cover type and target size.
func (c CoverType) Resize(target image.Point, sz image.Point) image.Point {
	switch c {
	case CoverTypeLibList:
		return resizeKeepAspectRatio(sz, target, false)
	case CoverTypeLibFull, CoverTypeLibGrid:
		return resizeKeepAspectRatio(sz, target, true)
	}
	panic("unknown cover type")
}

// GeneratePath generates the path for the cover of an ImageID. The path is always
// separated with forward slashes.
func (c CoverType) GeneratePath(external bool, iid string) string {
	cdir := ".kobo-images"
	if external {
		cdir = "koboExtStorage/images-cache"
	}
	dir1, dir2, base := hashedImageParts(iid)
	return fmt.Sprintf("%s/%s/%s/%s - %s.parsed", cdir, dir1, dir2, base, c.NickelString())
}

func (d Device) StorageGB() int {
	switch d {
	case DeviceTouchAB, DeviceTouchC, DeviceMini:
		return 2
	case DeviceTouch2, DeviceAuraHD, DeviceAuraH2O, DeviceGloHD, DeviceAura, DeviceGlo, DeviceAuraEdition2v1, DeviceAuraEdition2v2:
		return 4
	case DeviceAuraH2OEdition2v1, DeviceAuraH2OEdition2v2, DeviceClaraHD, DeviceLibraH2O, DeviceAuraONE, DeviceForma:
		return 8
	case DeviceAuraONELimitedEdition, DeviceForma32:
		return 32
	}
	panic("unknown device")
}
