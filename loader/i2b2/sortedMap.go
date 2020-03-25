package loaderi2b2

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"

	libunlynx "github.com/ldsec/unlynx/lib"
	"github.com/sirupsen/logrus"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/onet/v3/log"
)

type entryType struct {
	key   string
	value PatientDimension
}
type SortedTablePatientDimensionType struct {
	entries []entryType
}

var SortedTablePatientDimension *SortedTablePatientDimensionType

func (sortedTablePatientDimension *SortedTablePatientDimensionType) Len() int {
	return len(sortedTablePatientDimension.entries)
}
func (sortedTablePatientDimension *SortedTablePatientDimensionType) Less(i, j int) bool {
	a, err := strconv.Atoi(sortedTablePatientDimension.entries[i].key)
	if err != nil {
		logrus.Panic("error in Atoi")
	}
	b, err := strconv.Atoi(sortedTablePatientDimension.entries[j].key)
	if err != nil {
		logrus.Panic("error in Atoi")
	}

	if a <= b {
		return true
	}
	return false

}
func (sortedTablePatientDimension *SortedTablePatientDimensionType) Swap(i, j int) {
	tmpCopy := sortedTablePatientDimension.entries[i]
	sortedTablePatientDimension.entries[i] = sortedTablePatientDimension.entries[j]
	sortedTablePatientDimension.entries[j] = tmpCopy
}

func NewSortedTablePatientDimensionType(tablePatientDimension map[PatientDimensionPK]PatientDimension) *SortedTablePatientDimensionType {
	newTable := &SortedTablePatientDimensionType{}
	for patientPK, patientDimension := range tablePatientDimension {
		newTable.entries = append(newTable.entries, entryType{key: patientPK.PatientNum, value: patientDimension})
	}
	sort.Sort(newTable)

	return newTable
}

//func (sortedTablePatientDimension *SortedTablePatientDimensionType) Get() (PatientDimension,bool)

func (sortedTablePatientDimension *SortedTablePatientDimensionType) ForEach(callBack func(entry entryType) error) (err error) {
	for _, elm := range sortedTablePatientDimension.entries {
		logrus.Trace(elm)
		err = callBack(elm)
		if err != nil {
			return
		}
	}
	return
}

//---------------------------------------- CallBack from loaderI2B2.go

type internalStates struct {
	pk            kyber.Point
	csvOutputFile *os.File
	i             *int
	perm          []int
	empty         bool
}

// -----------------------------

func (internals *internalStates) LoaderI2B2CallBack(entry entryType) (err error) {

	pd := entry.value
	MapNewPatientNum[pd.PK.PatientNum] = strconv.FormatInt(int64(internals.perm[*internals.i]), 10)
	logrus.Tracef("patient number %s has permutation %d", pd.PK.PatientNum, internals.perm[*internals.i])
	pd.PK.PatientNum = strconv.FormatInt(int64(internals.perm[*internals.i]), 10)
	internals.csvOutputFile.WriteString(pd.ToCSVText(internals.empty) + "\n")
	logrus.Trace("value of i ", strconv.Itoa(*internals.i))
	*internals.i++
	return nil

}

//---------------------------------------------------------------same for dummy

type factEntryType struct {
	key   *ObservationFactPK
	value ObservationFact
}

type SortedTableObservationFactType struct {
	entries []factEntryType
}

var SortedTableObservationFact *SortedTableObservationFactType

func (sortedTableObservationFact *SortedTableObservationFactType) Len() int {
	return len(sortedTableObservationFact.entries)
}
func (sortedTableObservationFact *SortedTableObservationFactType) Less(i, j int) bool {
	a, err := strconv.Atoi(sortedTableObservationFact.entries[i].key.PatientNum)
	if err != nil {
		logrus.Panic("error in Atoi")
	}
	b, err := strconv.Atoi(sortedTableObservationFact.entries[j].key.PatientNum)
	if err != nil {
		logrus.Panic("error in Atoi")
	}

	if a <= b {
		return true
	}
	return false

}
func (sortedTableObservationFact *SortedTableObservationFactType) Swap(i, j int) {
	tmpCopy := sortedTableObservationFact.entries[i]
	sortedTableObservationFact.entries[i] = sortedTableObservationFact.entries[j]
	sortedTableObservationFact.entries[j] = tmpCopy
}

func (sortedTableObservationFact *SortedTableObservationFactType) ForEach(callBack func(entry factEntryType) error) (err error) {
	for _, elm := range sortedTableObservationFact.entries {
		logrus.Trace(elm)
		err = callBack(elm)
		if err != nil {
			return
		}
	}
	return
}

func NewSortedTableObservationFactType(tableObservationFact map[*ObservationFactPK]ObservationFact) *SortedTableObservationFactType {
	newTable := &SortedTableObservationFactType{}
	for ofpk, of := range tableObservationFact {
		newTable.entries = append(newTable.entries, factEntryType{key: ofpk, value: of})
	}
	sort.Sort(newTable)

	return newTable
}

//--------------------fact callback
var factInternals *internalStates

func (internals *internalStates) FactCallback(entry factEntryType) (err error) {

	//for ofk, of := range TableObservationFact
	of := entry.value
	ofk := entry.key
	copyObs := of
	survFact := IsSurvivalFact(ofk)

	//observation blob for survival analysis
	if survFact {
		//ok is a extra check
		cipherBlob, ok := EventObservationBlobEncrypted[ofk]
		if !ok {
			return fmt.Errorf("Key for %s was not found. Was the encrpytion of the observation blob performed ?", fmt.Sprint(*ofk))
		}
		copyObs.ObservationBlob = cipherBlob

	}

	// if dummy observation
	if _, ok := TableDummyToPatient[of.PK.PatientNum]; ok {
		// 1. choose a random observation from the original patient
		// 2. copy the data
		// 3. change patient_num and encounter_num
		// 4. if the observation is a survival analysis recpord, add the blob
		listObs := MapDummyObs[of.PK.PatientNum]

		// TODO: find out why this can be 0 (the generation should not allow this
		if len(listObs) == 0 {
			return
		}

		index := rand.Intn(len(listObs))
		logrus.Tracef("Dummy patient %s for index %d, has obs %v ", of.PK.PatientNum, index, *listObs[index])
		copyObs = TableObservationFact[listObs[index]]

		// change patient_num and encounter_num
		tmp := MapNewEncounterNum[VisitDimensionPK{EncounterNum: copyObs.PK.EncounterNum, PatientNum: of.PK.PatientNum}]
		logrus.Tracef("Copy PatientNum %s EncounterNum %s", copyObs.PK.PatientNum, copyObs.PK.EncounterNum)
		copyObs.PK = regenerateObservationPK(copyObs.PK, tmp.PatientNum, tmp.EncounterNum)
		logrus.Tracef("Tmp PatientNum %s EncounterNum %s", copyObs.PK.PatientNum, copyObs.PK.EncounterNum)
		// keep the same concept (and text_search_index) that was already there
		copyObs.PK.ConceptCD = of.PK.ConceptCD
		copyObs.AdminColumns.TextSearchIndex = of.AdminColumns.TextSearchIndex
		logrus.Tracef("Dummy tmp has value %v", tmp)
		//Encrypts a "0 0" event and writes it in the blob of the copyobs
		if survFact {
			copyObs.ObservationBlob, err = EncryptZeroEvent()

			if err != nil {
				return err
			}
		}
		// delete observation from the list (so we don't choose it again)
		listObs[index] = listObs[len(listObs)-1]
		listObs = listObs[:len(listObs)-1]
		MapDummyObs[of.PK.PatientNum] = listObs

	} else { // if real observation
		// change patient_num and encounter_num
		tmp := MapNewEncounterNum[VisitDimensionPK{EncounterNum: of.PK.EncounterNum, PatientNum: of.PK.PatientNum}]
		copyObs.PK = regenerateObservationPK(copyObs.PK, tmp.PatientNum, tmp.EncounterNum)
	}

	// if the concept is sensitive we replace its code with the correspondent tag ID
	if _, ok := MapConceptCodeToTag[copyObs.PK.ConceptCD]; ok {
		copyObs.PK.ConceptCD = "TAG_ID:" + strconv.FormatInt(MapConceptCodeToTag[copyObs.PK.ConceptCD], 10)

	}
	//this should not happen
	if _, isSensi := MapConceptCodeToTag[copyObs.PK.ConceptCD]; strings.Contains(copyObs.PK.ConceptCD, "SRVA") && !isSensi {
		log.Fatalf("A concept identifies as a survival concept, but is not in the MapConceptCodeToTag %s", copyObs.PK.ConceptCD)
	}

	// TODO: connected with the previous TODO
	if copyObs.PK.EncounterNum != "" {
		logrus.Tracef("Final patient for  observation key %v, value %v is %s", *ofk, of, copyObs.PK.PatientNum)
		internals.csvOutputFile.WriteString(copyObs.ToCSVText() + "\n")
	}
	return

}

//---------------------------------------**********Same for dummy

type dummyEntryType struct {
	key   string
	value string
}

type SortedTableDummyToPatientType struct {
	entries []dummyEntryType
}

var SortedTableDummyToPatient *SortedTableDummyToPatientType

func NewSortedTableDummyToPatientType(sortedTableDummyToPatient map[string]string) *SortedTableDummyToPatientType {
	newTable := &SortedTableDummyToPatientType{}
	for patientA, patientB := range sortedTableDummyToPatient {
		newTable.entries = append(newTable.entries, dummyEntryType{key: patientA, value: patientB})
	}
	sort.Sort(newTable)

	return newTable

}

func (sortedTableDummyToPatient *SortedTableDummyToPatientType) Len() int {
	return len(sortedTableDummyToPatient.entries)
}
func (sortedTableDummyToPatient *SortedTableDummyToPatientType) Less(i, j int) bool {
	a, err := strconv.Atoi(sortedTableDummyToPatient.entries[i].key)
	if err != nil {
		logrus.Panic("error in Atoi")
	}
	b, err := strconv.Atoi(sortedTableDummyToPatient.entries[j].key)
	if err != nil {
		logrus.Panic("error in Atoi")
	}

	if a <= b {
		return true
	}
	return false

}
func (sortedTableDummyToPatient *SortedTableDummyToPatientType) Swap(i, j int) {
	tmpCopy := sortedTableDummyToPatient.entries[i]
	sortedTableDummyToPatient.entries[i] = sortedTableDummyToPatient.entries[j]
	sortedTableDummyToPatient.entries[j] = tmpCopy
}

func (sortedTableDummyToPatient *SortedTableDummyToPatientType) ForEach(callBack func(entry dummyEntryType) error) (err error) {
	for _, elm := range sortedTableDummyToPatient.entries {
		logrus.Trace(elm)
		err = callBack(elm)
		if err != nil {
			return
		}
	}
	return
}

//---------------------------------------***** same for visits

type visitEntryType struct {
	key   VisitDimensionPK
	value VisitDimension
}

type SortedTableVisitDimensionType struct {
	entries []visitEntryType
}

var SortedTableVisitDimension *SortedTableVisitDimensionType

func NewSortedTableVisitDimensionType(sortedTableVisitDimension map[VisitDimensionPK]VisitDimension) *SortedTableVisitDimensionType {
	newTable := &SortedTableVisitDimensionType{}
	for visitPK, visitDim := range sortedTableVisitDimension {
		newTable.entries = append(newTable.entries, visitEntryType{key: visitPK, value: visitDim})
	}
	sort.Sort(newTable)

	return newTable

}

func (sortedTableVisitDimension *SortedTableVisitDimensionType) Len() int {
	return len(sortedTableVisitDimension.entries)
}
func (sortedTableVisitDimension *SortedTableVisitDimensionType) Less(i, j int) bool {
	a, err := strconv.Atoi(sortedTableVisitDimension.entries[i].key.PatientNum)
	if err != nil {
		logrus.Panic("error in Atoi")
	}
	b, err := strconv.Atoi(sortedTableVisitDimension.entries[j].key.PatientNum)
	if err != nil {
		logrus.Panic("error in Atoi")
	}

	if a <= b {
		return true
	}
	return false

}
func (sortedTableVisitDimension *SortedTableVisitDimensionType) Swap(i, j int) {
	tmpCopy := sortedTableVisitDimension.entries[i]
	sortedTableVisitDimension.entries[i] = sortedTableVisitDimension.entries[j]
	sortedTableVisitDimension.entries[j] = tmpCopy
}

func (sortedTableVisitDimension *SortedTableVisitDimensionType) ForEach(callBack func(entry visitEntryType) error) (err error) {
	for _, elm := range sortedTableVisitDimension.entries {
		logrus.Trace(elm)
		err = callBack(elm)
		if err != nil {
			return
		}
	}
	return
}

// visit callback
var visitAndDummyInternals *internalStates

func (internals *internalStates) VisitCallback(entry visitEntryType) error {
	vd := entry.value
	MapNewEncounterNum[VisitDimensionPK{EncounterNum: vd.PK.EncounterNum, PatientNum: vd.PK.PatientNum}] = VisitDimensionPK{EncounterNum: strconv.FormatInt(int64(internals.perm[*internals.i]), 10), PatientNum: MapNewPatientNum[vd.PK.PatientNum]}
	vd.PK.EncounterNum = strconv.FormatInt(int64(internals.perm[*internals.i]), 10)
	vd.PK.PatientNum = MapNewPatientNum[vd.PK.PatientNum]
	internals.csvOutputFile.WriteString(vd.ToCSVText(internals.empty) + "\n")
	*internals.i++
	return nil
}

func (internals *internalStates) DummyConversionCallback(entry dummyEntryType) error {
	dummyNum := entry.key
	patientNum := entry.value
	logrus.Tracef("Patient dummy num %s permi %d", dummyNum, internals.perm[*internals.i])
	MapNewPatientNum[dummyNum] = strconv.FormatInt(int64(internals.perm[*internals.i]), 10)

	patient := TablePatientDimension[PatientDimensionPK{PatientNum: patientNum}]
	patient.PK.PatientNum = strconv.FormatInt(int64(internals.perm[*internals.i]), 10)
	ef := libunlynx.EncryptInt(internals.pk, 0)
	patient.EncryptedFlag = *ef

	internals.csvOutputFile.WriteString(patient.ToCSVText(internals.empty) + "\n")
	*internals.i++
	return nil

}

func (internals *internalStates) DummyVisitCallback(entry dummyEntryType) error {
	dummyNum := entry.key
	patientNum := entry.value

	listPatientVisits := MapPatientVisits[patientNum]

	for _, el := range listPatientVisits {
		logrus.Tracef("EncounterNum %s PatientNum %s permi %d mapNewPatient %s ", el, dummyNum, internals.perm[*internals.i], MapNewPatientNum[dummyNum])
		MapNewEncounterNum[VisitDimensionPK{EncounterNum: el, PatientNum: dummyNum}] = VisitDimensionPK{EncounterNum: strconv.FormatInt(int64(internals.perm[*internals.i]), 10), PatientNum: MapNewPatientNum[dummyNum]}
		visit := TableVisitDimension[VisitDimensionPK{EncounterNum: el, PatientNum: patientNum}]
		visit.PK.EncounterNum = strconv.FormatInt(int64(internals.perm[*internals.i]), 10)
		visit.PK.PatientNum = MapNewPatientNum[dummyNum]
		internals.csvOutputFile.WriteString(visit.ToCSVText(internals.empty) + "\n")
		*internals.i++
	}

	return nil

}

//
/*
i := 0
	for _, vd := range TableVisitDimension {
		MapNewEncounterNum[VisitDimensionPK{EncounterNum: vd.PK.EncounterNum, PatientNum: vd.PK.PatientNum}] = VisitDimensionPK{EncounterNum: strconv.FormatInt(int64(perm[i]), 10), PatientNum: MapNewPatientNum[vd.PK.PatientNum]}
		vd.PK.EncounterNum = strconv.FormatInt(int64(perm[i]), 10)
		vd.PK.PatientNum = MapNewPatientNum[vd.PK.PatientNum]
		csvOutputFile.WriteString(vd.ToCSVText(empty) + "\n")
		i++
	}

	// add dummies
	for dummyNum, patientNum := range TableDummyToPatient {
		listPatientVisits := MapPatientVisits[patientNum]

		for _, el := range listPatientVisits {
			MapNewEncounterNum[VisitDimensionPK{EncounterNum: el, PatientNum: dummyNum}] = VisitDimensionPK{EncounterNum: strconv.FormatInt(int64(perm[i]), 10), PatientNum: MapNewPatientNum[dummyNum]}
			visit := TableVisitDimension[VisitDimensionPK{EncounterNum: el, PatientNum: patientNum}]
			visit.PK.EncounterNum = strconv.FormatInt(int64(perm[i]), 10)
			visit.PK.PatientNum = MapNewPatientNum[dummyNum]
			csvOutputFile.WriteString(visit.ToCSVText(empty) + "\n")
			i++
		}
	}
*/
