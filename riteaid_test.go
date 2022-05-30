package riteaid

import (
	"fmt"
	"testing"
	"time"

	"github.com/zsefvlol/timezonemapper"
)

func Test__getStoreDataURL(t *testing.T) {
	// !! SKIP
	t.SkipNow()
	// Test for expected value
	address := "4 Walton St E, Willard, OH 44890"

	want := "https://www.riteaid.com/services/ext/v2/stores/getStores?pharmacyOnly=false&globalZipCodeRequired=true&address=4+Walton+St+E%2C+Willard%2C+OH+44890&radius=3.1"
	got, err := __getStoreDataURL(address, 3.1)
	if got != want || err != nil {
		t.Errorf("__getStoreDataURL(%q, 3) = %q, want %q", address, got, want)
	}

	// Test for radius at max value
	want = "https://www.riteaid.com/services/ext/v2/stores/getStores?pharmacyOnly=false&globalZipCodeRequired=true&address=4+Walton+St+E%2C+Willard%2C+OH+44890&radius=25"
	got, err = __getStoreDataURL(address, 25)
	if got != want || err != nil {
		t.Errorf("__getStoreDataURL(%q, 25) = %q, want %q", address, got, want)
	}

	// Test for Radius Over Max
	want = "https://www.riteaid.com/services/ext/v2/stores/getStores?pharmacyOnly=false&globalZipCodeRequired=true&address=4+Walton+St+E%2C+Willard%2C+OH+44890&radius=999"
	got, err = __getStoreDataURL(address, 999)
	if got != want || err != ErrRadiusOverMax {
		t.Errorf("__getStoreDataURL(%q, 999) = %q, want %q", address, got, want)
	}

	// Test for Radius Under Min
	want = ""
	got, err = __getStoreDataURL(address, -1)
	if got != want || err != ErrRadiusUnderMin {
		t.Errorf("__getStoreDataURL(%q, -1) = %q, want %q", address, got, want)
	}

	// Test GetStoreDataStruct
	want = ""
	result, err := GetStoreDataStruct(address, 0.5)
	// TODO: More Tests!
	if err != nil || result.Status != "SUCCESS" || result.Data.Stores == nil || len(result.Data.Stores) == 0 || result.Data.Stores[0].StoreNumber == 0 || result.Data.Stores[0].StoreNumber != 3357 {
		t.Errorf("GetStoreDataStruct(%q, 0.5) = <result>, want %q", address, want)
	}

	// Test MapURL
	// !! DEPRECATED !!
	want = "http://maps.google.com/maps?daddr=Rite+Aid%2C+4+East+Walton+Street%2C+Willard%2C+OH+44890-9419"
	got = GetMapURL(result.Data.Stores[0])
	if got != want {
		t.Errorf("GetMapURL(<store>) = %q, want %q", got, want)
	}

	// Test GetFedExPickupURL
	// !! DEPRECATED !!
	want = "https://www.fedex.com/grd/rpp/ShowRPP.do?pickupType=Business&contactName=Onsite%20Manager&state=OH&pickupLocation=0&weightOver150=No&companyName=Rite+Aid&trackingId=123456789012&address1=4+East+Walton+Street&city=Willard&zip=44890-9419&phoneNum=4199353900&numPackages=1"
	sent := "9622 0131 4 (000 000 0000) 0 00 1234 5678 9012"
	got = GetFedExPickupURL(result.Data.Stores[0], sent)
	if got != want {
		t.Errorf("GetFedExPickupURL(<store>, \"%s\") = %q, want %q", sent, got, want)
	}

	want = "https://www.fedex.com/grd/rpp/ShowRPP.do?pickupType=Business&contactName=Onsite%20Manager&state=OH&pickupLocation=0&weightOver150=No&companyName=Rite+Aid&trackingId=123456789012&address1=4+East+Walton+Street&city=Willard&zip=44890-9419&phoneNum=4199353900&numPackages=1"
	sent = "9622013140009780845100123456789012"
	got = GetFedExPickupURL(result.Data.Stores[0], sent)
	if got != want {
		t.Errorf("GetFedExPickupURL(<store>, \"%s\") = %q, want %q", sent, got, want)
	}

	// Test GetStoreAddress
	want = "Rite Aid, 4 East Walton Street, Willard, OH 44890-9419"
	got = GetStoreAddress(result.Data.Stores[0])
	if got != want {
		t.Errorf("GetStoreAddress(<store>) = %q, want %q", got, want)
	}

	// Test IsStoreOpen
	loc, _ := time.LoadLocation(result.Data.Stores[0].TimeZone)
	dateTime, _ := time.ParseInLocation(DateFormat, "2022-05-22 7:30PM", loc)
	isOpenStore, isOpenRX, err := IsStoreOpen(dateTime, result.Data.Stores[0])
	if err != nil || !isOpenStore || isOpenRX {
		t.Errorf("IsStoreOpen(%s, <store>) = [%t,%t], want [true,false]", dateTime.String(), isOpenStore, isOpenRX)
	}

	// Test removeNonNumeric
	want = "12345678901"
	got = removeNonNumeric("+1 (234) 567-8901")
	if got != want {
		t.Errorf("removeNonNumeric(\"%q\") = %q, want %q", "+1 (234) 567-8901", got, want)
	}

	// TODO: A LOT!

	// Test for failure
	// got, err = __getStoreDataURL(address, -1)
}

func TestGetStoreHours(t *testing.T) {
	storeData := Store{
		Latitude:  41.0428,
		Longitude: -82.7258,

		StoreHoursMonday:    "8:01am-10:01pm",
		StoreHoursTuesday:   "8:02am-10:02pm",
		StoreHoursWednesday: "8:03am-10:03pm",
		StoreHoursThursday:  "8:04am-10:04pm",
		StoreHoursFriday:    "8:05am-10:05pm",
		StoreHoursSaturday:  "8:06am-10:06pm",
		StoreHoursSunday:    "8:00am-10:00pm",
		RXHrsMon:            "9:01am-9:01pm",
		RXHrsTue:            "9:02am-9:02pm",
		RXHrsWed:            "9:03am-9:03pm",
		RXHrsThu:            "9:04am-9:04pm",
		RXHrsFri:            "9:05am-9:05pm",
		RXHrsSat:            "9:06am-9:06pm",
		RXHrsSun:            "9:00am-9:00pm",

		HolidayHours: []HolidayHours{
			{
				HolidayDate:   "2016-12-25",
				StoreHours:    "12:00pm-8:00pm",
				PharmacyHours: "1:00pm-7:00pm",
			},
			{
				HolidayDate:   "2017-12-25",
				StoreHours:    "11:00am-7:00pm",
				PharmacyHours: "12:00pm-6:00pm",
			},
		},
	}

	// Test Holiday Hours
	var storeHours [2]time.Time
	var rxHours [2]time.Time
	var err error

	const (
		DateFormat = "2006-01-02 3:04PM"
	)

	// Get the current date in the time zone / location specified
	locName := timezonemapper.LatLngToTimezoneString(storeData.Latitude, storeData.Longitude)
	loc, err := time.LoadLocation(locName)
	if err != nil {
		t.Errorf("Error loading location: %s", err)
	}

	// Test First Holiday Hours
	date := "2016-12-25"
	t.Log("Testing First Holiday Hours")
	storeHours, rxHours, err = GetStoreHours(date, storeData)
	// Check for Error
	if err != nil {
		t.Errorf("GetStoreHours(\"%q\", <storeData>) ERROR: %q", date, err)
	}
	// Check Store Hours
	if storeHours[0].Format("3:04PM") != "12:00PM" || storeHours[1].Format("3:04PM") != "8:00PM" {
		t.Errorf("GetStoreHours(\"%q\", <storeData>) = [%q,%q], want [12:00PM,8:00PM]", date, storeHours[0].Format("3:04PM"), storeHours[1].Format("3:04PM"))
	}
	// Check Store Date
	if storeHours[0].Format("2006-01-02") != "2016-12-25" || storeHours[1].Format("2006-01-02") != "2016-12-25" {
		t.Errorf("GetStoreHours(\"%q\", <storeData>) = [%q,%q], want [2016-12-25,2016-12-25]", date, storeHours[0].Format("2006-01-02"), storeHours[1].Format("2006-01-02"))
	}
	// Check RX Hours
	if rxHours[0].Format("3:04PM") != "1:00PM" || rxHours[1].Format("3:04PM") != "7:00PM" {
		t.Errorf("GetStoreHours(\"%q\", <storeData>) = [%q,%q], want [1:00PM,7:00PM]", date, rxHours[0].Format("3:04PM"), rxHours[1].Format("3:04PM"))
	}
	// Check RX Date
	if rxHours[0].Format("2006-01-02") != "2016-12-25" || rxHours[1].Format("2006-01-02") != "2016-12-25" {
		t.Errorf("GetStoreHours(\"%q\", <storeData>) = [%q,%q], want [2016-12-25,2016-12-25]", date, rxHours[0].Format("2006-01-02"), rxHours[1].Format("2006-01-02"))
	}

	// Test Second Holiday Hours
	date = "2017-12-25"
	t.Log("Testing Second Holiday Hours")
	storeHours, rxHours, err = GetStoreHours(date, storeData)
	// Check for Error
	if err != nil {
		t.Errorf("GetStoreHours(\"%q\", <storeData>) ERROR: %q", date, err)
	}
	// Check Store Hours
	if storeHours[0].Format("3:04PM") != "11:00AM" || storeHours[1].Format("3:04PM") != "7:00PM" {
		t.Errorf("GetStoreHours(\"%q\", <storeData>) = [%q,%q], want [11:00AM,7:00PM]", date, storeHours[0].Format("3:04PM"), storeHours[1].Format("3:04PM"))
	}
	// Check Store Date
	if storeHours[0].Format("2006-01-02") != "2017-12-25" || storeHours[1].Format("2006-01-02") != "2017-12-25" {
		t.Errorf("GetStoreHours(\"%q\", <storeData>) = [%q,%q], want [2017-12-25,2017-12-25]", date, storeHours[0].Format("2006-01-02"), storeHours[1].Format("2006-01-02"))
	}
	// Check RX Hours
	if rxHours[0].Format("3:04PM") != "12:00PM" || rxHours[1].Format("3:04PM") != "6:00PM" {
		t.Errorf("GetStoreHours(\"%q\", <storeData>) = [%q,%q], want [12:00PM,6:00PM]", date, rxHours[0].Format("3:04PM"), rxHours[1].Format("3:04PM"))
	}
	// Check RX Date
	if rxHours[0].Format("2006-01-02") != "2017-12-25" || rxHours[1].Format("2006-01-02") != "2017-12-25" {
		t.Errorf("GetStoreHours(\"%q\", <storeData>) = [%q,%q], want [2017-12-25,2017-12-25]", date, rxHours[0].Format("2006-01-02"), rxHours[1].Format("2006-01-02"))
	}

	// Test All Weekday Hours.  Targeting 2018-11-01 as it will also process daylight savings
	// NOTE: Time is set to 6AM to ensure that DST transition has occurred else calculations might be off
	startDate, err := time.ParseInLocation("2006-01-02 3:04 PM", "2018-11-01 6:00 AM", loc)
	if err != nil {
		t.Errorf("Error parsing start date: %s", err)
	}
	weekdayHours(startDate, storeData, t)

	// Test All Weekday Hours.  Targeting 2018-03-10 as it will also process daylight savings
	startDate, err = time.ParseInLocation("2006-01-02 3:04 PM", "2022-03-10 6:00 AM", loc)
	if err != nil {
		t.Errorf("Error parsing start date: %s", err)
	}
	// Call helper function to test all weekday hours
	weekdayHours(startDate, storeData, t)

}

func weekdayHours(date time.Time, storeData Store, t *testing.T) {
	initLoc := date.Location().String()
	var storeHours, rxHours [2]time.Time
	var storeHoursWant, rxHoursWant [2]string
	var err error

	const (
		DateFormat = "2006-01-02 3:04 PM MST"
	)

	for i := 0; i < 7; i++ {
		wd := date.Weekday()
		t.Logf("Testing %s Hours for date %q", wd, date.Format("2006-01-02 MST"))

		storeHoursWant = [2]string{fmt.Sprintf("%s 8:0%d AM %s", date.Format("2006-01-02"), wd, date.Format("MST")), fmt.Sprintf("%s 10:0%d PM %s", date.Format("2006-01-02"), wd, date.Format("MST"))}
		rxHoursWant = [2]string{fmt.Sprintf("%s 9:0%d AM %s", date.Format("2006-01-02"), wd, date.Format("MST")), fmt.Sprintf("%s 9:0%d PM %s", date.Format("2006-01-02"), wd, date.Format("MST"))}

		// Test for weekday hours
		storeHours, rxHours, err = GetStoreHours(date.Format("2006-01-02"), storeData)
		if err != nil {
			t.Errorf("GetStoreHours(%q, <storeData>) = Error: %q", date.Format(DateFormat), err)
			date.Zone()
		}
		if storeHours[0].Format(DateFormat) != storeHoursWant[0] || storeHours[1].Format(DateFormat) != storeHoursWant[1] {
			t.Errorf("GetStoreHours(%q, <storeData>) = [%q,%q], want [%q,%q]", date.Format(DateFormat), storeHours[0].Format(DateFormat), storeHours[1].Format(DateFormat), storeHoursWant[0], storeHoursWant[1])
		}
		if rxHours[0].Format(DateFormat) != rxHoursWant[0] || rxHours[1].Format(DateFormat) != rxHoursWant[1] {
			t.Errorf("GetStoreHours(%q, <storeData>) = [%q,%q], want [%q,%q]", date.Format(DateFormat), rxHours[0].Format(DateFormat), rxHours[1].Format(DateFormat), rxHoursWant[0], rxHoursWant[1])
		}

		// Exit if done
		if i == 7 {
			break
		}

		// Proceed to next dat
		date = date.AddDate(0, 0, 1)
		if date.Location().String() != initLoc {
			t.Logf("Location changed from %q to %q", initLoc, date.Location().String())
		}
	}
}

// func TestMain(m *testing.M) {
// 	scratch()
// }

// Test Get Store Data API call
// func TestGetStoreData(t *testing.T) {
// 	address := "4 Walton St E, Willard, OH 44890"

// 	want := ""
// 	got, err := GetStoreData(address, 3)
// 	if got != want || err != nil {
// 		t.Errorf("GetStoreData(%q, 3) = %q, want %q", address, got, want)
// 	}
// }
