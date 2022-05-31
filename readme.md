# RiteAid Store Search WebAPI Wrapper
> ***WARNING!*** This is my _VERY_ first Go project so their ~~maybe~~ **defiantly are** dragons! 🔥🐉 
> Expect breaking changes until v1

>The test module is still being tweaked to ensure I have accounted for all the dragon eggs 🥚 that may hatch in the future.

## What it do? 🔎
This module is a wrapper for the RiteAid Store search API. The date and times are store location / timezone aware and will always be returned with the store locality in mind. This way calculations can traverse timezones accurately

## Why I do? ✨
I'm a field tech and I close a lot of repair tickets for RiteAid and the tools I have to managed routing/trip planning are _meh_ 🤷.

🤓 So, I am creating/created tools to streamline my workflow and this module is one of those tools.

## Who else would do?
Someone whom wants to query the RiteAid store data would find this useful. It's kinda niche so I don't expect much if any interest in the project.

## Example Usage:
```golang
// Get a specific store data by entering stores known address with a minimal search radius
storeAddress := "4 Walton St E, Willard, OH 44890"
searchRadius := 0.1
rawJsonResult, err := GetStoreData(storeAddress, searchRadius)
if err != nil {
    panic(err)
}
os.WriteFile("storeData.json", []byte(rawJsonResult), 0644)
```
```golang
storeAddress := "4 Walton St E, Willard, OH 44890"
searchRadius := 0.1

// Get Store Data as a struct
searchResults, err := GetStoreDataStruct(storeAddress, searchRadius)
if err != nil {
    panic(err)
}
storeData := searchResults.Data.Stores[0]

// Get the store hours for today
var storeHours [2]time.Time
var rxHours [2]time.Time
storeHours, rxHours, err := GetStoreHours(time.Now().Format(DateFormat), storeData)
if err != nil {
    panic(err)
}

fmt.Printf("Store Hours: %s\n", storeHours)
fmt.Printf("Rx Hours: %s\n", rxHours)

// Is the store and rx open?
isOpenStore, isOpenRX, err := IsStoreOpen(time.Now(), storeData)
if err != nil {
    panic(err)
}

fmt.Printf("Is Store Open: %t\n", isOpenStore)
fmt.Printf("Is RX Open: %t\n", isOpenRX)
```
## TODO / Known Issues:
- [ ] Initial Alpha release!
- [ ] FIX BUG: GetStoreHours fails to account for Daylight Savings 
  - [ ] Issue was corrected with inclusion of [github.com/zsefvlol/timezonemapper](https://github.com/zsefvlol/timezonemapper) which doesn't appear to be actively maintained. Thus, I'll need to look over the code and see if I need to make changes to it, but at the moment it works well.
  - [ ] BUG: Weekday tests are failing, this is due to issues implementing the new external module.
- [ ] Finish Test Routines
- [ ] Code Review
- [ ] Code Review AGAIN!
- [ ] Remove the "DEPRECIATED" functions and put them into their own module or snippets as they are for special use cases that really only apply to a field tech.