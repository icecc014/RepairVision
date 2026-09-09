package logic

import (
	"testing"

	"map/mapclient"
)

func TestValidateRoomForBuilding(t *testing.T) {
	b := &mapclient.Building{Id: 4, Name: "4号宿舍楼", Floors: 6, RoomsPerFloor: 16}
	if err := validateRoomForBuilding("401", 4, b); err != nil {
		t.Fatalf("401 should be valid: %v", err)
	}
	if err := validateRoomForBuilding("701", 7, b); err == nil {
		t.Fatal("7th floor should be invalid")
	}
	if err := validateRoomForBuilding("417", 4, b); err == nil {
		t.Fatal("room 417 should be out of range")
	}
}
