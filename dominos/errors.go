package dominos

import (
	"errors"
	"fmt"
)

var (
	InvalidCountry  = errors.New("Your Wii's country is not supported (US and Canada only).\nError Code: ")
	GenericError    = errors.New("An unknown error has occurred. Please contact WiiLink support.\nError Code: ")
	NoDeliveryHours = errors.New("No delivery hours are available.\nError Code: ")
)

func MakeError(err map[string]any) error {
	// Dominos error system is quite confusing, the more people that encounter errors the more we will learn.
	statuses := err["Order"].(map[string]any)["StatusItems"].([]any)

	for _, status := range statuses {
		// Our implementation should always have the first dictionary be "AutoAddedOrderId".
		// To be safe we will skip if encountered.
		code := status.(map[string]any)["Code"].(string)
		if code == "AutoAddedOrderId" {
			continue
		}

		// This is not guaranteed to exist. If it does, it is much more verbose than the code.
		pulseText := status.(map[string]any)["PulseText"]
		if pulseText != nil {
			return errors.New(fmt.Sprintf("An error has occured: %s\nError Code: ", pulseText))
		}

		// Default to the error code
		return errors.New(fmt.Sprintf("An error has occured: %s\nError Code: ", code))
	}

	// If somehow nothing existed, give generic error.
	return GenericError
}
