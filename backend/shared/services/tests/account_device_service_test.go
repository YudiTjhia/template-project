package services_test

import (
	"testing"

	"backend/shared/models"
	"backend/shared/services"
)

func TestAccountDeviceServiceUpsertAndUpdateLocation(t *testing.T) {
	if authServiceDB == nil {
		t.Skip("auth service db is not initialized")
	}

	_, _ = authServiceDB.Exec(`TRUNCATE TABLE account_devices RESTART IDENTITY CASCADE`)

	service := services.NewAccountDeviceService(authServiceDB, models.Auth{
		AccountID: 7,
		UserName:  "alice",
		Email:     "alice@example.com",
	})

	record, err := service.UpsertLogin("Device-ABC")
	if err != nil {
		t.Fatalf("expected no error upserting device, got %v", err)
	}
	if record == nil {
		t.Fatalf("expected device record")
	}
	if record.DeviceID != "device-abc" {
		t.Fatalf("expected normalized device id, got %q", record.DeviceID)
	}
	if record.LocationID != nil {
		t.Fatalf("expected location to be nil on login upsert")
	}

	locationID := int64(42)
	updated, err := service.UpdateLocation("Device-ABC", locationID)
	if err != nil {
		t.Fatalf("expected no error updating device location, got %v", err)
	}
	if updated == nil || updated.LocationID == nil || *updated.LocationID != locationID {
		t.Fatalf("expected location id %d, got %+v", locationID, updated)
	}

	found, err := service.FindByDevice("device-abc")
	if err != nil {
		t.Fatalf("expected no error finding device, got %v", err)
	}
	if found == nil || found.LocationID == nil || *found.LocationID != locationID {
		t.Fatalf("expected stored location id %d, got %+v", locationID, found)
	}
}
