package times

import "time"

func GetMoscowLocation() *time.Location {
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		location = time.FixedZone("UTC+3", 3*60*60)
	}

	return location
}
