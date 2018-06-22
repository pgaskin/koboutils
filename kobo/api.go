package kobo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// UpgradeCheckResult represents an update check result from the Kobo API.
type UpgradeCheckResult struct {
	Data           interface{}
	ReleaseNoteURL string
	UpgradeType    UpgradeType
	UpgradeURL     string
}

// UpgradeType represents an upgrade type.
type UpgradeType int

// Upgrade types.
const (
	UpgradeTypeNone      UpgradeType = 0 // No upgrade available.
	UpgradeTypeAvailable UpgradeType = 1 // Optional update, but never seen this before.
	UpgradeTypeRequired  UpgradeType = 2 // Automatic update.
)

func (u UpgradeType) String() string {
	switch u {
	case UpgradeTypeNone:
		return "None"
	case UpgradeTypeAvailable:
		return "Available"
	case UpgradeTypeRequired:
		return "Required"
	default:
		return "Unknown (" + strconv.Itoa(int(u)) + ")"
	}
}

// IsUpdate checks if an UpdateType signifies an available update.
func (u UpgradeType) IsUpdate() bool {
	return u != UpgradeTypeNone
}

// CheckUpgrade queries the Kobo API for an update.
func CheckUpgrade(device, affiliate, curVersion, serial string) (*UpgradeCheckResult, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.kobobooks.com/1.0/UpgradeCheck/Device/%s/%s/%s/%s", device, affiliate, curVersion, serial))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("response status %d", resp.StatusCode)
	}

	var res UpgradeCheckResult
	err = json.NewDecoder(resp.Body).Decode(&res)

	return &res, err
}
