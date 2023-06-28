package formatting

import "strings"

// CleanUpPhoneNumber : Clean out to plain number text
func CleanUpPhoneNumber(countryCode, phoneNumber string) string {
	phoneNumber = strings.TrimSpace(phoneNumber)
	phoneNumber = strings.Replace(phoneNumber, " ", "", -1)
	phoneNumber = strings.Replace(phoneNumber, "+", "", -1)
	phoneNumber = strings.Replace(phoneNumber, "-", "", -1)
	phoneNumber = strings.TrimLeft(phoneNumber, "0")
	phoneNumber = strings.TrimLeft(phoneNumber, countryCode)
	return phoneNumber
}
