package loaderi2b2

import (
	"github.com/ldsec/medco-loader/loader"
	"github.com/ldsec/unlynx/lib"
	"github.com/stretchr/testify/assert"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/onet/v3"
	"go.dedis.ch/onet/v3/app"
	"go.dedis.ch/onet/v3/log"
	"os"
	"testing"
)

var publicKey kyber.Point
var el *onet.Roster
var local *onet.LocalTest

func getRoster(groupFilePath string) (*onet.Roster, *onet.LocalTest, error) {
	// empty string: make localtest
	if len(groupFilePath) == 0 {
		log.Info("Creating local test roster")

		local := onet.NewLocalTest(libunlynx.SuiTe)
		_, el, _ := local.GenTree(3, true)
		return el, local, nil

		// generate el with group file
	} else {
		log.Info("Creating roster from group file path")

		f, err := os.Open(groupFilePath)
		if err != nil {
			log.Error("Error while opening group file", err)
			return nil, nil, err
		}
		el, err := app.ReadGroupDescToml(f)
		if err != nil {
			log.Error("Error while reading group file", err)
			return nil, nil, err
		}
		if len(el.Roster.List) <= 0 {
			log.Error("Empty or invalid group file", err)
			return nil, nil, err
		}

		return el.Roster, nil, nil
	}
}

func setupEncryptEnv() {
	elAux, localAux, err := getRoster("")
	if err != nil {
		log.Fatal("Something went wrong when creating a testing environment!")
	}
	el = elAux
	local = localAux

	_, publicKey = libunlynx.GenKey()
}

func TestConvertTableAccess(t *testing.T) {
	log.SetDebugVisible(2)

	assert.Nil(t, ParseTableAccess())
	assert.Nil(t, ConvertTableAccess())
}

func TestParseDummyToPatient(t *testing.T) {
	log.SetDebugVisible(2)

	assert.Nil(t, ParseDummyToPatient())
}

func TestConvertPatientDimension(t *testing.T) {
	log.SetDebugVisible(2)
	setupEncryptEnv()

	ParseDummyToPatient()

	assert.Nil(t, ParsePatientDimension(publicKey))
	assert.Nil(t, ConvertPatientDimension(publicKey))

	local.CloseAll()
}

func TestConvertVisitDimension(t *testing.T) {
	log.SetDebugVisible(2)
	setupEncryptEnv()

	ParseDummyToPatient()

	ParsePatientDimension(publicKey)
	ConvertPatientDimension(publicKey)

	assert.Nil(t, ParseVisitDimension())
	assert.Nil(t, ConvertVisitDimension())

	local.CloseAll()
}

func TestUpdateChildrenEncryptIDs(t *testing.T) {
	TablesMedCoOntology = make(map[string]MedCoTableInfo)
	tableMedCoOntologyConceptEnc := make(map[string]*MedCoOntology)
	TablesMedCoOntology["test"] = MedCoTableInfo{Sensitive: tableMedCoOntologyConceptEnc}

	so0 := MedCoOntology{Fullname: "\\a\\", NodeEncryptID: 0}
	so1 := MedCoOntology{Fullname: "\\a\\b\\", NodeEncryptID: 1}
	so2 := MedCoOntology{Fullname: "\\a\\c\\", NodeEncryptID: 2}
	so3 := MedCoOntology{Fullname: "\\a\\c\\d", NodeEncryptID: 3}
	so4 := MedCoOntology{Fullname: "\\a\\c\\f", NodeEncryptID: 4}

	tableMedCoOntologyConceptEnc["\\a\\"] = &so0
	tableMedCoOntologyConceptEnc["\\a\\b\\"] = &so1
	tableMedCoOntologyConceptEnc["\\a\\c\\"] = &so2
	tableMedCoOntologyConceptEnc["\\a\\c\\d"] = &so3
	tableMedCoOntologyConceptEnc["\\a\\c\\f"] = &so4

	UpdateChildrenEncryptIDs("test")

	assert.Equal(t, 4, len(TablesMedCoOntology["test"].Sensitive["\\a\\"].ChildrenEncryptIDs))
	assert.Equal(t, 0, len(TablesMedCoOntology["test"].Sensitive["\\a\\b\\"].ChildrenEncryptIDs))
	assert.Equal(t, 2, len(TablesMedCoOntology["test"].Sensitive["\\a\\c\\"].ChildrenEncryptIDs))
	assert.Equal(t, 0, len(TablesMedCoOntology["test"].Sensitive["\\a\\c\\d"].ChildrenEncryptIDs))
	assert.Equal(t, 0, len(TablesMedCoOntology["test"].Sensitive["\\a\\c\\f"].ChildrenEncryptIDs))
}

func TestStripByLevel(t *testing.T) {

	test := `\SHRINE\Diagnoses\Neoplasms (140-239.99)\Benign neoplasms (210-229.99)\Benign neoplasm of bone and articular cartilage (213)\(213.9) Benign neoplasm of bone and articular cartilage, site unspecified\`
	result := StripByLevel(test, 1, true)
	assert.Equal(t, `\Diagnoses\Neoplasms (140-239.99)\Benign neoplasms (210-229.99)\Benign neoplasm of bone and articular cartilage (213)\(213.9) Benign neoplasm of bone and articular cartilage, site unspecified\`, result)

	test = `\SHRINE\Diagnoses\Neoplasms (140-239.99)\Benign neoplasms (210-229.99)\Benign neoplasm of bone and articular cartilage (213)\(213.9) Benign neoplasm of bone and articular cartilage, site unspecified\`
	result = StripByLevel(test, 2, true)
	assert.Equal(t, `\Neoplasms (140-239.99)\Benign neoplasms (210-229.99)\Benign neoplasm of bone and articular cartilage (213)\(213.9) Benign neoplasm of bone and articular cartilage, site unspecified\`, result)

	test = `\SHRINE\Diagnoses\Neoplasms (140-239.99)\Benign neoplasms (210-229.99)\Benign neoplasm of bone and articular cartilage (213)\(213.9) Benign neoplasm of bone and articular cartilage, site unspecified\`
	result = StripByLevel(test, 3, true)
	assert.Equal(t, `\Benign neoplasms (210-229.99)\Benign neoplasm of bone and articular cartilage (213)\(213.9) Benign neoplasm of bone and articular cartilage, site unspecified\`, result)

	test = `\SHRINE\Diagnoses\Neoplasms (140-239.99)\Benign neoplasms (210-229.99)\Benign neoplasm of bone and articular cartilage (213)\(213.9) Benign neoplasm of bone and articular cartilage, site unspecified\`
	result = StripByLevel(test, 1, false)
	assert.Equal(t, `\SHRINE\Diagnoses\Neoplasms (140-239.99)\Benign neoplasms (210-229.99)\Benign neoplasm of bone and articular cartilage (213)\`, result)

	test = `\SHRINE\Diagnoses\Neoplasms (140-239.99)\Benign neoplasms (210-229.99)\Benign neoplasm of bone and articular cartilage (213)\(213.9) Benign neoplasm of bone and articular cartilage, site unspecified\`
	result = StripByLevel(test, 2, false)
	assert.Equal(t, `\SHRINE\Diagnoses\Neoplasms (140-239.99)\Benign neoplasms (210-229.99)\`, result)

	test = `\SHRINE\Diagnoses\Neoplasms (140-239.99)\Benign neoplasms (210-229.99)\Benign neoplasm of bone and articular cartilage (213)\(213.9) Benign neoplasm of bone and articular cartilage, site unspecified\`
	result = StripByLevel(test, 10, true)
	assert.Equal(t, "", result)

	test = `\SHRINE\Diagnoses\Neoplasms (140-239.99)\Benign neoplasms (210-229.99)\Benign neoplasm of bone and articular cartilage (213)\(213.9) Benign neoplasm of bone and articular cartilage, site unspecified\`
	result = StripByLevel(test, 6, true)
	assert.Equal(t, "", result)
}

func TestConvertOntology(t *testing.T) {
	log.SetDebugVisible(2)
	setupEncryptEnv()
	Testing = true
	EnabledModifiers = true

	ListSensitiveConcepts = make(map[string]struct{})
	ListSensitiveConcepts[`\i2b2\Diagnoses\Neoplasms (140-239)\Benign neoplasms (210-229)\(216) Benign neoplasm of skin\`] = struct{}{}

	assert.Nil(t, ConvertLocalOntology(el, 0))
	assert.Nil(t, GenerateMedCoOntology())

	local.CloseAll()
}

func TestConvertOntology2(t *testing.T) {
	log.SetDebugVisible(2)
	setupEncryptEnv()
	Testing = true
	EnabledModifiers = true

	// testing modifiers with m_exclusion_cd = "X"
	ListSensitiveConcepts = make(map[string]struct{})
	ListSensitiveConcepts[`\i2b2\Diagnoses\Neoplasms (140-239)\Malignant neoplasms (140-208)\`] = struct{}{}

	assert.Nil(t, ConvertLocalOntology(el, 0))
	assert.Nil(t, GenerateMedCoOntology())

	local.CloseAll()
}

func TestConvertConceptDimension(t *testing.T) {
	log.SetDebugVisible(2)
	setupEncryptEnv()
	Testing = true
	EnabledModifiers = true

	ListSensitiveConcepts = make(map[string]struct{})
	ListSensitiveConcepts[`\i2b2\Diagnoses\Neoplasms (140-239)\Benign neoplasms (210-229)\(216) Benign neoplasm of skin\`] = struct{}{}

	assert.Nil(t, ConvertLocalOntology(el, 0))
	assert.Nil(t, GenerateMedCoOntology())

	if EnabledModifiers {
		assert.Nil(t, ParseModifierDimension())
		assert.Nil(t, ConvertModifierDimension())
	}

	assert.Nil(t, ParseConceptDimension())
	assert.Nil(t, ConvertConceptDimension())

	local.CloseAll()

}

func TestConvertAll(t *testing.T) {
	log.SetDebugVisible(2)
	setupEncryptEnv()
	Testing = true
	AllSensitive = false
	EnabledModifiers = true

	ListSensitiveConcepts = make(map[string]struct{})
	ListSensitiveConcepts[`\i2b2\Diagnoses\`] = struct{}{}

	assert.Nil(t, ConvertLocalOntology(el, 0))

	log.LLvl1("--- Finished converting LOCAL_ONTOLOGY ---")

	assert.Nil(t, GenerateMedCoOntology())

	log.LLvl1("--- Finished generating MEDCO_ONTOLOGY ---")

	if EnabledModifiers {
		assert.Nil(t, ParseModifierDimension())
		assert.Nil(t, ConvertModifierDimension())
		log.LLvl1("--- Finished converting MODIFIER_DIMENSION ---")
	}

	assert.Nil(t, ParseConceptDimension())
	assert.Nil(t, ConvertConceptDimension())

	log.LLvl1("--- Finished converting CONCEPT_DIMENSION ---")

	assert.Nil(t, FilterOldObservationFact())

	log.Lvl1("--- Finished filtering OLD_OBSERVATION_FACT ---")

	assert.Nil(t, FilterPatientDimension(publicKey))

	log.Lvl1("--- Finished filtering PATIENT_DIMENSION ---")

	assert.Nil(t, CallGenerateDummiesScript())

	log.Lvl1("--- Finished dummies generation ---")

	assert.Nil(t, ParseDummyToPatient())

	assert.Nil(t, ParsePatientDimension(publicKey))
	assert.Nil(t, ConvertPatientDimension(publicKey))

	log.LLvl1("--- Finished converting PATIENT_DIMENSION ---")

	assert.Nil(t, ParseVisitDimension())
	assert.Nil(t, ConvertVisitDimension())

	log.LLvl1("--- Finished converting VISIT_DIMENSION ---")

	assert.Nil(t, ParseNonSensitiveObservationFact())
	assert.Nil(t, ConvertNonSensitiveObservationFact())

	log.LLvl1("--- Finished converting non sensitive OBSERVATION_FACT ---")

	assert.Nil(t, ParseSensitiveObservationFact())
	assert.Nil(t, ConvertSensitiveObservationFact())

	log.LLvl1("--- Finished converting sensitive OBSERVATION_FACT ---")

	local.CloseAll()
}

func TestGenerateLoadingDataScript(t *testing.T) {
	assert.Nil(t, GenerateLoadingDataScriptSensitive(loader.DBSettings{DBhost: "localhost", DBport: 5434, DBname: "i2b2medcosrv0", DBuser: "i2b2", DBpassword: "i2b2"}))
}
