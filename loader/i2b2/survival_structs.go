package loaderi2b2

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// EventSeparator is the the substring that separate the clear or base4 representation of encrypted events used in survival analysis
	EventSeparator string = " "
	// ConceptIdentifier is the beginning of the observation code for all survival analysis entries
	ConceptIdentifier string = "SRVA:"
)

// Event holds the event of the survival analysis
type Event struct {
	EventOfInterest int64
	CensoringEvent  int64
}

// NewEvent Event creator
func NewEvent(event, censoring bool) *Event {
	var eventCode, censoringCode int64
	if event {
		eventCode = 1
	}
	if censoring {
		censoringCode = 1
	}

	return &Event{
		EventOfInterest: eventCode,
		CensoringEvent:  censoringCode,
	}
}

// NewEventFromString another Event creator
func NewEventFromString(str string) (evnt *Event, err error) {
	eventStrings := strings.Split(str, EventSeparator)
	length := len(eventStrings)
	if length != 2 {
		err = fmt.Errorf(`The string event %s contains an unexpected number of codes (%d against expected 2). Note that the current substring to separate event codes is "%s"`, str, length, EventSeparator)
		return
	}
	eventCode, err := strconv.ParseInt(eventStrings[0], 10, 64)
	if err != nil {
		return
	}
	censoringCode, err := strconv.ParseInt(eventStrings[1], 10, 64)
	if err != nil {
		return
	}
	if eventCode != int64(0) && eventCode != int64(1) && censoringCode != int64(0) && censoringCode != int64(1) {
		err = fmt.Errorf("Both event codes should be integers 0 or 1. Found %d and %d", eventCode, censoringCode)
		return
	}

	evnt = &Event{
		EventOfInterest: eventCode,
		CensoringEvent:  censoringCode,
	}

	return
}

// String is the string representation of the event
func (evnt *Event) String() string {
	return strconv.FormatInt(evnt.EventOfInterest, 10) + EventSeparator + strconv.FormatInt(evnt.CensoringEvent, 10)
}
