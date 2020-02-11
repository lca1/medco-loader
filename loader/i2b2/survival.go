package loaderi2b2

import (
	"strings"

	libunlynx "github.com/ldsec/unlynx/lib"
	"go.dedis.ch/onet/v3"
)

// IsSurvivalFact returns true if the concept code off the observation fact PK relates to a survival analysis observation
func IsSurvivalFact(factKey *ObservationFactPK) bool {
	return strings.Contains(factKey.ConceptCD, ConceptIdentifier)
}

// EncryptEventBlob takes a set of textual representation of survival analysis event and returns the equivalent set of encrypted event
func EncryptEventBlob(eventBlobs map[*ObservationFactPK]string, group *onet.Roster) (cipherMap map[*ObservationFactPK]string, err error) {
	length := len(eventBlobs)
	cipherMap = make(map[*ObservationFactPK]string, length)
	for position, blob := range eventBlobs {
		var event *Event
		event, err = NewEventFromString(blob)
		if err != nil {
			return
		}
		var eventEncrypted, censoringEncrypted string
		eventEncrypted, err = libunlynx.EncryptInt(group.Aggregate, event.EventOfInterest).Serialize()
		if err != nil {
			return
		}

		censoringEncrypted, err = libunlynx.EncryptInt(group.Aggregate, event.CensoringEvent).Serialize()
		if err != nil {
			return
		}

		//same separator ?
		cipherMap[position] = eventEncrypted + EventSeparator + censoringEncrypted
	}
	return
}

//EventBlobForDummyPatient returns the enryption of 0 and 0
func EventBlobForDummyPatient(group *onet.Roster) (cipherBlob string, err error) {
	zeroEvent := NewEvent(false, false)
	var eventEncrypted, censoringEncrypted string
	eventEncrypted, err = libunlynx.EncryptInt(group.Aggregate, zeroEvent.EventOfInterest).Serialize()
	if err != nil {
		return
	}

	censoringEncrypted, err = libunlynx.EncryptInt(group.Aggregate, zeroEvent.CensoringEvent).Serialize()
	if err != nil {
		return
	}

	cipherBlob = eventEncrypted + EventSeparator + censoringEncrypted
	return
}
