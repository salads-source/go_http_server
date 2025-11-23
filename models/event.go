package models

import "time"

type Event struct {
	ID          int
	Name        string    `binding: "Required"`
	Description string    `binding: "Required"`
	Location    string    `binding: "Required"`
	DateTime    time.Time `binding: "Required"`
	UserID      int
}

var events = []Event{}

func (e Event) Save() {
	events = append(events, e)
}

func GetAllEvents() []Event {
	return events
}
