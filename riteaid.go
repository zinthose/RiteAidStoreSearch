package riteaid

// TODO: Adjust to a allow for channels and function caching

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/zsefvlol/timezonemapper"
)

const (
	TimeFormat   = "3:04pm"
	TimeFormat_M = "3:04 PM"

	// DateFormat is the format used by the RiteAid API
	DateFormat = "2006-01-02"

	// DateFormatm is the format used by the RiteAid API for the date with the time
	//  i.e. "2006-01-02 3:04pm"
	DateTimeFormat = DateFormat + " " + TimeFormat

	// DateFormat_M is the format used by the RiteAid API for the date with the time
	//  i.e. "2006-01-02 3:04 PM"
	DateTimeFormat_M = DateFormat + " " + TimeFormat_M

	// Private constants
	riteAidAPIURL  = "https://www.riteaid.com/services/ext/v2/stores/getStores?pharmacyOnly=false&globalZipCodeRequired=true&address=%s&radius=%.3g"
	fedExPickupURL = `https://www.fedex.com/grd/rpp/ShowRPP.do?pickupType=Business&contactName=Onsite%%20Manager&state=%s&pickupLocation=0&weightOver150=No&companyName=%s&trackingId=%s&address1=%s&city=%s&zip=%s&phoneNum=%s&numPackages=%d`
	googleMapURL   = "http://maps.google.com/maps?daddr=%s"
)

// Error returned when the radius is under the minimum value
var ErrRadiusUnderMin = errors.New("radius is less than 0 this MUST be 0 or greater")

// Error returned when the radius is over the maximum value (This error can be ignored as it will default to 25 when over)
var ErrRadiusOverMax = errors.New("radius is greater than 25 the API will only return a max of 25 stores")

// Error returned when a parsed time span is out of expected order.
//  Expected order is: "8:00am-5:00pm" || "8:00 AM-5:00 PM"
var ErrTimeParseOrder = errors.New("parsed start time is after parsed end time")

// Error returned when RiteAid API returns an error
var ErrRiteAidAPIError = errors.New("RiteAid API returned an error")

// RiteAid store api returned a defined JSON structure.
// As of 5/29/2022 the structure is accurate but is subject to change as the API evolves.
//
// Known issues:
//  - The API returns are not fully mapped and some are being omitted from the struct
//  - Warnings is not verified
//  - AmbiguousAddresses is being omitted from import
type Result struct {
	Data      Data   `json:"data,omitempty"`
	Status    string `json:"Status"`
	ErrCde    string `json:"ErrCde,omitempty"`
	ErrMsg    string `json:"ErrMsg,omitempty"`
	ErrMsgDtl string `json:"ErrMsgDtl,omitempty"`
}

type Data struct {
	Stores          []Store         `json:"stores"`
	GlobalZipCode   string          `json:"globalZipCode,omitempty"`
	ResolvedAddress ResolvedAddress `json:"resolvedAddress"`
	Warnings        []string        `json:"warnings"` // TODO: This is a guess and needs verified
	// AmbiguousAddresses []AmbiguousAddress `json:"ambiguousAddresses"` // TODO: What is this?
}

type HolidayHours struct {
	HolidayDate   string `json:"holidayDate,omitempty"`
	StoreHours    string `json:"storeHours"`
	PharmacyHours string `json:"pharmacyHours"`
}

// SpecialHours is mapped as the key names are dynamic and not known ahead of time
//  i.e. "2006-01-02" -> "3:04pm-5:00pm"
type PickupDateAndTimes struct {
	RegularHours []string          `json:"regularHours"`
	DefaultTime  string            `json:"defaultTime"`
	Earliest     string            `json:"earliest"`
	SpecialHours map[string]string `json:"specialHours,omitempty"` // SpecialHours is returned with multiple values. Ex. {"2022-05-28": "1:00 PM-5:00 PM"}
}

type ResolvedAddress struct {
	AddressLine       string  `json:"addressLine"`
	AdminDistrict     string  `json:"adminDistrict"`
	Altitude          float64 `json:"altitude"`
	Confidence        string  `json:"confidence"`
	CalculationMethod string  `json:"calculationMethod"`
	CountryRegion     string  `json:"countryRegion"`
	DisplayName       string  `json:"displayName"`
	District          string  `json:"district"`
	FormattedAddress  string  `json:"formattedAddress"`
	GeocodeBestView   struct {
		NorthEastElements struct {
			Altitude  float64 `json:"altitude"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		}
		SouthWestElements struct {
			Altitude  float64 `json:"altitude"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		}
	}
	Latitude   float64 `json:"latitude"`
	Locality   string  `json:"locality"`
	Longitude  float64 `json:"longitude"`
	PostalCode string  `json:"postalCode,omitempty"`
	PostalTown string  `json:"postalTown,omitempty"`
}

// Event is not imported as the format is now known as of 2022/05/29
type Store struct {
	StoreNumber         uint32   `json:"storeNumber"`
	Address             string   `json:"address"`
	City                string   `json:"city"`
	State               string   `json:"state"`
	Zipcode             string   `json:"zipcode"`
	TimeZone            string   `json:"timeZone"`
	FullZipCode         string   `json:"fullZipCode"`
	FullPhone           string   `json:"fullPhone"`
	LocationDescription string   `json:"locationDescription"`
	StoreHoursMonday    string   `json:"storeHoursMonday"`
	StoreHoursTuesday   string   `json:"storeHoursTuesday"`
	StoreHoursWednesday string   `json:"storeHoursWednesday"`
	StoreHoursThursday  string   `json:"storeHoursThursday"`
	StoreHoursFriday    string   `json:"storeHoursFriday"`
	StoreHoursSaturday  string   `json:"storeHoursSaturday"`
	StoreHoursSunday    string   `json:"storeHoursSunday"`
	RXHrsMon            string   `json:"rxHrsMon"`
	RXHrsTue            string   `json:"rxHrsTue"`
	RXHrsWed            string   `json:"rxHrsWed"`
	RXHrsThu            string   `json:"rxHrsThu"`
	RXHrsFri            string   `json:"rxHrsFri"`
	RXHrsSat            string   `json:"rxHrsSat"`
	RXHrsSun            string   `json:"rxHrsSun"`
	StoreType           string   `json:"storeType"`
	Latitude            float64  `json:"latitude"`
	Longitude           float64  `json:"longitude"`
	Name                string   `json:"name"`
	MilesFromCenter     float64  `json:"milesFromCenter"`
	SpecialServicesKeys []string `json:"specialServicesKeys,omitempty"`
	// Event string `json:"event"` // TODO - Not sure what this is.
	HolidayHours       []HolidayHours     `json:"holidayHours,omitempty"`
	PickupDateAndTimes PickupDateAndTimes `json:"pickupDateAndTimes"`
}

// Function will place call to RiteAid API and return the store location data.
// It is recommended to keep the radius as small as possible to minimize the
// data received unless importing multiple stores for caching purposes.
//
// ADDRESS: The address of the store. i.e.
// "4 Walton St E, Willard, OH 44890" (required)
//
// RADIUS: The radius in miles of the address to search. i.e. 3
// (required - must be between 0.01 and 25)
//
// RETURNS: The raw store location data. This is the string value returned
// from the API call in JSON format
func GetStoreData(address string, radius float64) (string, error) {
	url, err := __getStoreDataURL(address, radius)
	if err != nil && err != ErrRadiusOverMax {
		return "", err
	}

	// Make the request
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	// Ready the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	sb := string(body)

	return sb, nil
}

// Function will place call to RiteAid API and return the store location data
// as a struct.
// It is recommended to keep the radius as small as possible to minimize the
// data received unless importing multiple stores for caching purposes.
//
// ADDRESS: The address of the store. i.e.
// "4 Walton St E, Willard, OH 44890" (required)
//
// RADIUS: The radius in miles of the address to search. i.e. 3
// (required - must be between 0.01 and 25)
//
// RETURNS: The store location data as a struct. In addition,
// if the API call fails, a ErrRiteAidAPIError will be raised
func GetStoreDataStruct(address string, radius float64) (Result, error) {
	var result Result
	var err error

	// Get Store Data
	sb, err := GetStoreData(address, radius)
	if err != nil {
		return result, err
	}

	// Unmarshal the response
	err = json.Unmarshal([]byte(sb), &result)
	if err != nil {
		return result, err
	}

	// If the API call failed, return the error
	if result.Status != "SUCCESS" {
		log.Println("RiteAid API call failed: " + result.ErrMsg)
		return result, ErrRiteAidAPIError
	}

	// Return the result
	return result, nil
}

// ParseTimeSpan takes a time range string and parses it into a start and end time.
// The time zone is required to ensure proper calculations based on the locality of
// the user, the store, and the server.
//  timeRange i.e. "8:00am-5:00pm" || "8:00 AM-5:00 PM"
//       date i.e. "2006-01-02"
//   timeZone i.e. "MST"
//
//  startTime, endTime, err := ParseTimeSpan("8:00am-5:00pm", "2006-01-02", )
func ParseTimeSpan(timeRange string, date string, latitude float64, longitude float64) (time.Time, time.Time, error) {
	const (
		dtz = "%s %s"
	)

	// Split the time span into start and end times
	times := strings.Split(timeRange, "-")

	// Detect format used for the time
	var form string
	if times[0][len(times[0])-1:] == "m" {
		form = DateTimeFormat
	} else {
		form = DateTimeFormat_M
	}

	// Get the time zone location of the store
	loc, err := GetTZLocation(latitude, longitude)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	// Parse the start time
	start, err := time.ParseInLocation(form, fmt.Sprintf(dtz, date, times[0]), loc)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	// Parse the end time
	end, err := time.ParseInLocation(form, fmt.Sprintf(dtz, date, times[1]), loc)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	// Require the start time to be before the end time
	if start.After(end) {
		return time.Time{}, time.Time{}, ErrTimeParseOrder
	}

	// Return the start and end times
	return start, end, nil
}

// ParseWeekDayHours takes a time range string and parses it into a start and end time for a given weekday
func ParseWeekDayHours(weekday time.Weekday, timeRange string, longitude float64, latitude float64) (time.Time, time.Time, error) {

	// Get the current date in the time zone / location specified
	loc, err := GetTZLocation(latitude, longitude)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	now := time.Now().In(loc)

	// Get the current weekday
	now_weekday := now.Weekday()
	if now_weekday == weekday {
		// It's today!
		return ParseTimeSpan(timeRange, now.Format(DateFormat), latitude, longitude)
	} else {
		// It's not today.
		// Get the next weekday
		next_weekday := now.AddDate(0, 0, int(weekday-now_weekday)).Format(DateFormat)
		return ParseTimeSpan(timeRange, next_weekday, latitude, longitude)
	}
}

// Retrieves the store hours for a given date. This takes into account the store's
// TimeZone and holiday hours.
//  // First return pair is the store hours.
//  // Second return pair is the rx hours.
//  var storeHours [2]time.Time
//  var rxHours [2]time.Time
//  storeHours, rxHours, err := GetStoreHours("2022-05-30", storeData)
func GetStoreHours(date string, storeData Store) ([2]time.Time, [2]time.Time, error) {

	// Verify date is in the correct format
	_, err := time.Parse(DateFormat, date)
	if err != nil {
		return [2]time.Time{}, [2]time.Time{}, err
	}

	// Get the current date in the time zone / location specified
	loc, err := GetTZLocation(storeData.Latitude, storeData.Longitude)
	if err != nil {
		return [2]time.Time{}, [2]time.Time{}, err
	}

	// Return holiday hours for target date if any
	for i := 0; i < len(storeData.HolidayHours); i++ {
		holiday := storeData.HolidayHours[i]

		// Verify holiday date is in the correct format
		_, err := time.Parse(DateFormat, holiday.HolidayDate)
		if err != nil {
			return [2]time.Time{}, [2]time.Time{}, err
		}

		// Check if the holiday date matches the target date
		if holiday.HolidayDate == date {
			// Return the holiday hours
			storeStart, storeEnd, err := ParseTimeSpan(holiday.StoreHours, date, storeData.Latitude, storeData.Longitude)
			if err != nil {
				return [2]time.Time{}, [2]time.Time{}, err
			}
			rxStart, rxEnd, err := ParseTimeSpan(holiday.PharmacyHours, date, storeData.Latitude, storeData.Longitude)
			if err != nil {
				return [2]time.Time{}, [2]time.Time{}, err
			}
			return [2]time.Time{storeStart, storeEnd}, [2]time.Time{rxStart, rxEnd}, nil
		}
	}

	// Return standard hours for target date
	dt, err := time.ParseInLocation(DateFormat, date, loc)
	if err != nil {
		return [2]time.Time{}, [2]time.Time{}, err
	}
	weekday := dt.Weekday()

	// Get unparsed RX and Store hours for a given weekday by using a little bit of code trickery.
	rxHrs := [7]string{
		storeData.RXHrsSun,
		storeData.RXHrsMon,
		storeData.RXHrsTue,
		storeData.RXHrsWed,
		storeData.RXHrsThu,
		storeData.RXHrsFri,
		storeData.RXHrsSat,
	}[weekday]
	storeHours := [7]string{
		storeData.StoreHoursSunday,
		storeData.StoreHoursMonday,
		storeData.StoreHoursTuesday,
		storeData.StoreHoursWednesday,
		storeData.StoreHoursThursday,
		storeData.StoreHoursFriday,
		storeData.StoreHoursSaturday,
	}[weekday]

	// Parse the store hours
	storeStart, storeEnd, err := ParseTimeSpan(storeHours, date, storeData.Latitude, storeData.Longitude)
	if err != nil {
		return [2]time.Time{}, [2]time.Time{}, err
	}

	// Parse the RX hours
	rxStart, rxEnd, err := ParseTimeSpan(rxHrs, date, storeData.Latitude, storeData.Longitude)
	if err != nil {
		return [2]time.Time{}, [2]time.Time{}, err
	}

	// log.Printf("Using standard hours %s\n", date)
	return [2]time.Time{storeStart, storeEnd}, [2]time.Time{rxStart, rxEnd}, nil
}

// Returns true if the store is open at the given date and time.
//  // First return is the store.
//  // Second return is the pharmacy.
//  loc, _ := time.LoadLocation(storeData.TimeZone)
//  dateTime, _ := time.ParseInLocation("2006-01-02 3:04PM", "2022-05-29 8:00PM", loc)
//  isOpenStore, isOpenRX, _ := IsStoreOpen(dateTime, storeData)
//  fmt.Printf("Is Store Open: %t\n", isOpenStore)
//  fmt.Printf("Is RX Open: %t\n", isOpenRX)
func IsStoreOpen(dateTime time.Time, storeData Store) (bool, bool, error) {
	var storeHours [2]time.Time
	var rxHours [2]time.Time

	storeHours, rxHours, err := GetStoreHours(dateTime.Format(DateFormat), storeData)
	if err != nil {
		return false, false, err
	}

	return dateTime.After(storeHours[0]) && dateTime.Before(storeHours[1]), dateTime.After(rxHours[0]) && dateTime.Before(rxHours[1]), nil
}

// *****************************************************************************
// * Private functions
// *****************************************************************************

// Private function to strip all non numeric characters from a string
//  removeNonNumeric("(419) 555-1212") -> "4195551212"
func removeNonNumeric(s string) string {
	return regexp.MustCompile(`[^\d]+`).ReplaceAllString(s, "")
}

// Returns the store address for the given store.
// This is primarily a helper function used internally but may be useful.
//  storeAddress, err := GetStoreAddress(storeData) ->
//    "Rite Aid, 4 East Walton Street, Willard, OH 44890-9419"
func GetStoreAddress(storeData Store) string {
	return fmt.Sprintf("%s, %s, %s, %s %s", storeData.Name, storeData.Address, storeData.City, storeData.State, storeData.FullZipCode)
}

// Private function to build the URL for the getStoreData API call.
//  ADDRESS: The address of the store. i.e. "4 Walton St E, Willard, OH 44890" (required)
//  RADIUS: The radius of the store. i.e. 3 (required - must be between 0 and 25. 0 = max radius)
//  RETURNS: The URL for the API call.
func __getStoreDataURL(address string, radius float64) (string, error) {
	// Require radius to be between 0 and 25 (0 is default will list all withing 25 mile radius)
	var err error
	if radius <= 0 {
		return "", ErrRadiusUnderMin
	} else if radius > 25 {
		err = ErrRadiusOverMax // Non critical error
	}
	encodedAddress := url.QueryEscape(address)
	url := fmt.Sprintf(riteAidAPIURL, encodedAddress, radius)

	return url, err
}

func GetTZLocation(longitude float64, latitude float64) (*time.Location, error) {
	// Get the current date in the time zone / location specified
	locName := timezonemapper.LatLngToTimezoneString(latitude, longitude)
	loc, err := time.LoadLocation(locName)
	if err != nil {
		return &time.Location{}, err
	}
	return loc, nil
}

// *****************************************************************************
// !! DEPRECIATED functions
// *****************************************************************************

// Private function that returns the 12 digit FedEx tracking number from
// the barcode data or OCR label data
//  !! DEPRECIATED: Function has a very specific use case.
//
//  fedexTracking("9622 0131 4 (000 000 0000) 0 00 1234 5678 9012") -> "123456789012"
//  fedexTracking("9622013140009780845100123456789012") -> "123456789012"
// TODO: Add support for 3D barcode
func fedexTracking(trackNum string) string {
	trackNum = removeNonNumeric(trackNum)
	if len(trackNum) > 12 {
		return trackNum[len(trackNum)-12:]
	}
	return trackNum
}

// Returns a FedEx URL for scheduling in store package pickup
//  !! DEPRECIATED: Function has a very specific use case.
//
//  // This is a standard tracking number
//  GetFedExPickupURL(storeData, "123456789012")
//
//  // This is a 2D barcode
//  GetFedExPickupURL(storeData, "9622013140009780845100123456789012")
//
//  // This is the OCR from above the 2D barcode.
//  GetFedExPickupURL(storeData, "9622 0131 4 (000 000 0000) 0 00 1234 5678 9012")
// The FedEx Tracking Number can be data directly from the label's barcode.
func GetFedExPickupURL(storeData Store, FedExTracking string) string {
	return fmt.Sprintf(fedExPickupURL,
		url.QueryEscape(storeData.State),
		url.QueryEscape(storeData.Name),
		url.QueryEscape(fedexTracking(FedExTracking)),
		url.QueryEscape(storeData.Address),
		url.QueryEscape(storeData.City),
		url.QueryEscape(storeData.FullZipCode),
		url.QueryEscape(removeNonNumeric(storeData.FullPhone)),
		1,
	)
}

// Ancillary function that returns a Google Maps URL to the store.
// This function is depreciated and will be removed in the future.
//  !! DEPRECIATED: Function has a very specific use case.
func GetMapURL(storeData Store) string {
	return fmt.Sprintf(googleMapURL, url.QueryEscape(GetStoreAddress(storeData)))
}
