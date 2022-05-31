package riteaid

import (
	"os"
	"testing"
	"time"
)

const (
	TestAddress = "4 Walton St E, Willard, OH 44890"
)

// Test the getStoreDataURL.
func Test_getStoreDataURL(t *testing.T) {
	const (
		radius   = 0.5
		expected = "https://www.riteaid.com/services/ext/v2/stores/getStores?pharmacyOnly=false&globalZipCodeRequired=true&address=4+Walton+St+E%2C+Willard%2C+OH+44890&radius=0.5"
	)
	url, err := getStoreDataURL(TestAddress, radius)
	if err != nil {
		t.Errorf("getStoreDataURL(%q, %f) Error: %v", TestAddress, radius, err)
	}
	if expected != url {
		t.Errorf("getStoreDataURL(%q, %f) = %q, want %q", TestAddress, radius, url, expected)
	}
}

func Test_GetTZLocation(t *testing.T) {
	const (
		latitudeAZ  = 33.604808
		longitudeAZ = -112.666466
	)
	loc, err := GetTZLocation(latitudeAZ, longitudeAZ)
	if err != nil {
		t.Errorf("GetTZLocation(%f, %f) Error: %v", latitudeAZ, longitudeAZ, err)
	}
	if loc.String() != "MST" {
		t.Errorf("GetTZLocation(%f, %f) = %q, want %q", latitudeAZ, longitudeAZ, loc.String(), "MST")
	}
}

func Test_(t *testing.T) {
	const (
		latitudeAZ     = 33.604808
		longitudeAZ    = -112.666466
		timeRange1     = "10:00am-9:00pm"
		timeRange2     = "9:00 AM-10:00 PM"
		wantFormatTime = "15:04:05-07:00"
		wantFormatDate = "2006-01-02"
		start1Want     = "10:00:00-07:00"
		end1Want       = "20:00:00-07:00"
		start2Want     = "09:00:00-07:00"
		end2Want       = "21:00:00-07:00"
	)
	var now = time.Now().Format(wantFormatDate)
	startTime, endTime, err := ParseTimeSpan(timeRange1, now, latitudeAZ, longitudeAZ)
	if err != nil {
		t.Errorf("ParseTimeSpan(%q, %q, %f, %f) Error: %v", timeRange1, now, latitudeAZ, longitudeAZ, err)
	}
	if startTime.Format(wantFormatTime) != start1Want {
		t.Errorf("ParseTimeSpan(%q, %q, %f, %f) = %q, want %q", timeRange1, now, latitudeAZ, longitudeAZ, startTime.Format(wantFormatTime), start1Want)
	}
	if startTime.Format(wantFormatTime) != end1Want {
		t.Errorf("ParseTimeSpan(%q, %q, %f, %f) = %q, want %q", timeRange1, now, latitudeAZ, longitudeAZ, endTime.Format(wantFormatTime), end1Want)
	}
	startTime, endTime, err = ParseTimeSpan(timeRange2, now, latitudeAZ, longitudeAZ)
	if err != nil {
		t.Errorf("ParseTimeSpan(%q, %q, %f, %f) Error: %v", timeRange2, now, latitudeAZ, longitudeAZ, err)
	}
	if startTime.Format(wantFormatTime) != start2Want {
		t.Errorf("ParseTimeSpan(%q, %q, %f, %f) = %q, want %q", timeRange2, now, latitudeAZ, longitudeAZ, startTime.Format(wantFormatTime), start2Want)
	}
	if startTime.Format(wantFormatTime) != end2Want {
		t.Errorf("ParseTimeSpan(%q, %q, %f, %f) = %q, want %q", timeRange2, now, latitudeAZ, longitudeAZ, endTime.Format(wantFormatTime), end2Want)
	}
}

// Test the GetStoreDataJSON. The results are saved to file
// "storeData.json". This test is skipped when testing in
// short mode.
func Test_GetStoreDataJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping GetStoreDataJSON")
	}
	var json, err = GetStoreDataJSON(TestAddress, 0.5)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if len(json) == 0 {
		t.Errorf("Error: %v", "json is empty")
	}
	if json[0] != '{' && json[len(json)-1] != '}' {
		t.Errorf("Error: %v", "json is not a JSON object")
	}

	// Output to storeData.json
	if !t.Failed() {
		os.WriteFile("storeData.json", []byte(json), 0644)
	}
}

// Test the GetStoreData. This test is skipped when testing in
// short mode.
func Test_GetStoreData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping GetStoreData")
	}
	var result, err = GetStoreData(TestAddress, 0.5)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if result.Status != "SUCCESS" {
		t.Errorf("Result Status: %v", result.Status)
	}
}
