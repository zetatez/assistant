package gcal

import "time"

type Event struct {
	ID          string
	Title       string
	Description string
	Start       time.Time
	End         time.Time
	Timezone    string
	Location    string
}
