package loader

import (
	"encoding/base64"
	"github.com/dedis/kyber"
	"github.com/dedis/onet/log"
	"github.com/lca1/unlynx/lib"
	"strconv"
	"strings"
)

// ####----HELPER STRUCTS----####

// ListSensitiveConceptsShrine list all the sensitive concepts (paths) - SHRINE (the bool is for nothing)
var ListSensitiveConceptsShrine map[string]bool

// ListSensitiveConceptsLocal list all the sensitive concepts (paths) and the respective shrine equivalent - LOCAL (the bool is for nothing)
var ListSensitiveConceptsLocal map[string][]string

// IDModifiers used to assign IDs to the modifiers concepts
var IDModifiers int64

// IDConcepts used to assign IDs to the different concepts
var IDConcepts int64

// ####----DATA TYPES----####

// TableShrineOntologyClear is the shrine_ontology table (it maps the concept path to a concept) with only the NON_SENSITIVE concepts (it INCLUDES MODIFIER NON-SENSITIVE concepts)
var TableShrineOntologyClear map[string]*ShrineOntology

// TableShrineOntologyConceptEnc is the shrine_ontology table (it maps the concept path to a concept) with only the SENSITIVE concepts (NO MODIFIER SENSITIVE concepts)
var TableShrineOntologyConceptEnc map[string]*ShrineOntology

// TableShrineOntologyModifierEnc is the shrine_ontology table (it maps the concept path to a concept) with only the MODIFIER SENSITIVE concepts
var TableShrineOntologyModifierEnc map[string][]*ShrineOntology

// HeaderShrineOntology contains all the headers for the shrine table
var HeaderShrineOntology []string

// ShrineOntology is the table that contains all concept codes from the shrine ontology
type ShrineOntology struct {
	NodeEncryptID      int64
	ChildrenEncryptIDs []int64

	HLevel           string
	Fullname         string
	Name             string
	SynonymCD        string
	VisualAttributes string
	TotalNum         string
	BaseCode         string
	MetadataXML      string
	FactTableColumn  string
	Tablename        string
	ColumnName       string
	ColumnDataType   string
	Operator         string
	DimCode          string
	Comment          string
	Tooltip          string
	AdminColumns     AdministrativeColumns
	ValueTypeCD      string
	AppliedPath      string
	ExclusionCD      string
}

// ToCSVText writes the ShrineOntology object in a way that can be added to a .csv file - "","","", etc.
func (so ShrineOntology) ToCSVText() string {
	if so.NodeEncryptID != int64(-1) { // sensitive
		metadata := ""

		if so.VisualAttributes[:1] == "C" { // if concept_parent_node
			metadata += "<?xml version=\"\"1.0\"\"?><ValueMetadata><Version>MedCo-0.1</Version><EncryptedType>CONCEPT_PARENT_NODE</EncryptedType>"
		} else if so.VisualAttributes[:1] == "F" { // else if concept_internal_node
			metadata += "<?xml version=\"\"1.0\"\"?><ValueMetadata><Version>MedCo-0.1</Version><EncryptedType>CONCEPT_INTERNAL_NODE</EncryptedType><NodeEncryptID>" + strconv.FormatInt(so.NodeEncryptID, 10) + "</NodeEncryptID>"
		} else if so.VisualAttributes[:1] == "L" { // else if concept_leaf
			metadata += "<?xml version=\"\"1.0\"\"?><ValueMetadata><Version>MedCo-0.1</Version><EncryptedType>CONCEPT_LEAF</EncryptedType><NodeEncryptID>" + strconv.FormatInt(so.NodeEncryptID, 10) + "</NodeEncryptID>"
		} else if so.VisualAttributes[:1] == "O" { // else if modifier_parent_node
			metadata += "<?xml version=\"\"1.0\"\"?><ValueMetadata><Version>MedCo-0.1</Version><EncryptedType>MODIFIER_PARENT_NODE</EncryptedType>"
		} else if so.VisualAttributes[:1] == "D" { // else if modifier_internal_node
			metadata += "<?xml version=\"\"1.0\"\"?><ValueMetadata><Version>MedCo-0.1</Version><EncryptedType>MODIFIER_INTERNAL_NODE</EncryptedType><NodeEncryptID>" + strconv.FormatInt(so.NodeEncryptID, 10) + "</NodeEncryptID>"
		} else if so.VisualAttributes[:1] == "R" { // else if modifier_leaf
			metadata += "<?xml version=\"\"1.0\"\"?><ValueMetadata><Version>MedCo-0.1</Version><EncryptedType>MODIFIER_LEAF</EncryptedType><NodeEncryptID>" + strconv.FormatInt(so.NodeEncryptID, 10) + "</NodeEncryptID>"
		} else if so.VisualAttributes[:1] == "M" {
			log.Fatal("Not supported go fuck yourself!")
		} else {
			log.Fatal("Wrong VisualAttribute")
		}

		// only internal and parent nodes can have children ;)
		if len(so.ChildrenEncryptIDs) > 0 && so.VisualAttributes[:1] != "L" && so.VisualAttributes[:1] != "R" {
			metadata += "<ChildrenEncryptIDs>"
			for _, childID := range so.ChildrenEncryptIDs {
				metadata += "<ChildEncryptID>" + strconv.FormatInt(childID, 10) + "</ChildEncryptID>"
			}

			metadata += "</ChildrenEncryptIDs>"
		}
		so.MetadataXML = metadata + "</ValueMetadata>"
	}
	acString := "\"" + so.AdminColumns.UpdateDate + "\"," + "\"" + so.AdminColumns.DownloadDate + "\"," + "\"" + so.AdminColumns.ImportDate + "\"," + "\"" + so.AdminColumns.SourceSystemCD + "\""

	return "\"" + so.HLevel + "\"," + "\"" + so.Fullname + "\"," + "\"" + so.Name + "\"," + "\"" + so.SynonymCD + "\"," + "\"" + so.VisualAttributes + "\"," + "\"" + so.TotalNum + "\"," +
		"\"" + so.BaseCode + "\"," + "\"" + so.MetadataXML + "\"," + "\"" + so.FactTableColumn + "\"," + "\"" + so.Tablename + "\"," + "\"" + so.ColumnName + "\"," + "\"" + so.ColumnDataType + "\"," + "\"" + so.Operator + "\"," +
		"\"" + so.DimCode + "\"," + "\"" + so.Comment + "\"," + "\"" + so.Tooltip + "\"," + acString + "," + "\"" + so.ValueTypeCD + "\"," + "\"" + so.AppliedPath + "\"," + "\"" + so.ExclusionCD + "\""
}

//-------------------------------------//

// TableLocalOntologyClear is the local ontology table (it maps the concept path to a concept) with only the NON_SENSITIVE concepts (it INCLUDES MODIFIER NON-SENSITIVE concepts)
var TableLocalOntologyClear map[string]*LocalOntology

// TagAndID is a struct that contains both Tag and TagID for a concept or modifier
type TagAndID struct {
	Tag   libunlynx.GroupingKey
	TagID int64
}

// CodeID is a struct that contains both concept code and concept path
type CodeID struct {
	Path string
	Code string
}

// MapConceptPathToTag maps a sensitive concept path to its respective tag and tag_id
var MapConceptPathToTag map[string]TagAndID

// MapModifierPathToTag maps a sensitive modifier path to its respective tag and tag_id
var MapModifierPathToTag map[string]TagAndID

// HeaderLocalOntology contains all the headers for the i2b2 table
var HeaderLocalOntology []string

// LocalOntology is the table that contains all concept codes from the local ontology (i2b2)
type LocalOntology struct {
	HLevel           string
	Fullname         string
	Name             string
	SynonymCD        string
	VisualAttributes string
	TotalNum         string
	BaseCode         string
	MetadataXML      string
	FactTableColumn  string
	Tablename        string
	ColumnName       string
	ColumnDataType   string
	Operator         string
	DimCode          string
	Comment          string
	Tooltip          string
	AppliedPath      string
	AdminColumns     AdministrativeColumns
	ValueTypeCD      string
	ExclusionCD      string
	Path             string
	Symbol           string

	// this only exists in the sensitive tagged
	PCoriBasecode string
}

// ToCSVText writes the LocalOntology object in a way that can be added to a .csv file - "","","", etc.
func (lo LocalOntology) ToCSVText() string {
	acString := "\"" + lo.AdminColumns.UpdateDate + "\"," + "\"" + lo.AdminColumns.DownloadDate + "\"," + "\"" + lo.AdminColumns.ImportDate + "\"," + "\"" + lo.AdminColumns.SourceSystemCD + "\""

	return "\"" + lo.HLevel + "\"," + "\"" + lo.Fullname + "\"," + "\"" + lo.Name + "\"," + "\"" + lo.SynonymCD + "\"," + "\"" + lo.VisualAttributes + "\"," + "\"" + lo.TotalNum + "\"," +
		"\"" + lo.BaseCode + "\"," + "\"" + lo.MetadataXML + "\"," + "\"" + lo.FactTableColumn + "\"," + "\"" + lo.Tablename + "\"," + "\"" + lo.ColumnName + "\"," + "\"" + lo.ColumnDataType + "\"," + "\"" + lo.Operator + "\"," +
		"\"" + lo.DimCode + "\"," + "\"" + lo.Comment + "\"," + "\"" + lo.Tooltip + "\"," + "\"" + lo.AppliedPath + "\"," + acString + "," + "\"" + lo.ValueTypeCD + "\"," + "\"" + lo.ExclusionCD + "\"," +
		"\"" + lo.Path + "\"," + "\"" + lo.Symbol + "\""
}

// LocalOntologySensitiveConceptToCSVText writes the tagging information of a concept of the local ontology in a way that can be added to a .csv file - "","","", etc.
func LocalOntologySensitiveConceptToCSVText(tag *libunlynx.GroupingKey, tagID int64) string {
	return `"3", "\medco\tagged\concept\` + string(*tag) + `\", "", "N", "LA ", "\N", "TAG_ID:` + strconv.FormatInt(tagID, 10) + `", "\N", "concept_cd", "concept_dimension", "concept_path", "T", "LIKE", "\medco\tagged\concept\` + string(*tag) +
		`\", "\N", "\N", "NOW()", "\N", "\N", "\N", "TAG_ID", "@", "\N", "\N", "\N", "\N"`
}

// LocalOntologySensitiveModifierToCSVText writes the tagging information of a modifier of the local ontology in a way that can be added to a .csv file - "","","", etc.
func LocalOntologySensitiveModifierToCSVText(tag *libunlynx.GroupingKey, tagID int64) string {
	return `"3", "\medco\tagged\modifier\` + string(*tag) + `\", "", "N", "LA ", "\N", "TAG_ID:` + strconv.FormatInt(tagID, 10) + `", "\N", "MODIFIER_CD", "MODIFIER_DIMENSION", "MODIFIER_PATH", "T", "LIKE", "\medco\tagged\modifier\` + string(*tag) +
		`\", "\N", "\N", "NOW()", "\N", "\N", "\N", "TAG_ID", "@", "\N", "\N", "\N", "\N"`
}

//-------------------------------------//

var MapPatientObs map[string][]*ObservationFactPK

//-------------------------------------//

// TableObservationFact is observation_fact table
var TableObservationFact map[*ObservationFactPK]ObservationFact

// HeaderObservationFact contains all the headers for the observation_fact table
var HeaderObservationFact []string

// ObservationFact is the fact table of the CRC-I2B2 start schema
type ObservationFact struct {
	PK              *ObservationFactPK
	ValTypeCD       string
	TValChar        string
	NValNum         string
	ValueFlagCD     string
	QuantityNum     string
	UnitsCD         string
	EndDate         string
	LocationCD      string
	ObservationBlob string
	ConfidenceNum   string
	AdminColumns    AdministrativeColumns
}

// ObservationFactPK is the primary key of ObservationFact
type ObservationFactPK struct {
	EncounterNum string
	PatientNum   string
	ConceptCD    string
	ProviderID   string
	StartDate    string
	ModifierCD   string
	InstanceNum  string
}

// AdministrativeColumns are a set of columns that exist in every i2b2 table
type AdministrativeColumns struct {
	UpdateDate      string
	DownloadDate    string
	ImportDate      string
	SourceSystemCD  string
	UploadID        string
	TextSearchIndex string
}

// ToCSVText writes the ObservationFact object in a way that can be added to a .csv file - "","","", etc.
func (lo ObservationFact) ToCSVText() string {
	acString := "\"" + lo.AdminColumns.UpdateDate + "\"," + "\"" + lo.AdminColumns.DownloadDate + "\"," + "\"" + lo.AdminColumns.ImportDate + "\"," + "\"" + lo.AdminColumns.SourceSystemCD + "\"," + "\"" + lo.AdminColumns.UploadID + "\"," + "\"" + lo.AdminColumns.TextSearchIndex + "\""

	return "\"" + lo.PK.EncounterNum + "\"," + "\"" + lo.PK.PatientNum + "\"," + "\"" + lo.PK.ConceptCD + "\"," + "\"" + lo.PK.ProviderID + "\"," + "\"" + lo.PK.StartDate + "\"," + "\"" + lo.PK.ModifierCD + "\"," +
		"\"" + lo.PK.InstanceNum + "\"," + "\"" + lo.ValTypeCD + "\"," + "\"" + lo.TValChar + "\"," + "\"" + lo.NValNum + "\"," + "\"" + lo.ValueFlagCD + "\"," + "\"" + lo.QuantityNum + "\"," + "\"" + lo.UnitsCD + "\"," +
		"\"" + lo.EndDate + "\"," + "\"" + lo.LocationCD + "\"," + "\"" + lo.ObservationBlob + "\"," + "\"" + lo.ConfidenceNum + "\"," + acString
}

//-------------------------------------//

// TableDummyToPatient contains all dummies and the original patient that is associated with them
var TableDummyToPatient map[string]string

// MapNewPatientNum keeps track of the mapping between the old patient_num and the new one
var MapNewPatientNum map[string]string

//-------------------------------------//

// TablePatientDimension is patient_dimension table
var TablePatientDimension map[PatientDimensionPK]PatientDimension

// HeaderPatientDimension contains all the headers for the Patient_Dimension table
var HeaderPatientDimension []string

// PatientDimension table represents a patient in the database
type PatientDimension struct {
	PK             PatientDimensionPK
	VitalStatusCD  string
	BirthDate      string
	DeathDate      string
	OptionalFields []OptionalFields
	AdminColumns   AdministrativeColumns
	EncryptedFlag  libunlynx.CipherText
}

// PatientDimensionPK is the primary key of the Patient_Dimension table
type PatientDimensionPK struct {
	PatientNum string
}

// ToCSVText writes the PatientDimensionPK struct in a way that can be added to a .csv file - "","","", etc.
func (pdk PatientDimensionPK) ToCSVText() string {
	return "\"" + pdk.PatientNum + "\""
}

// ToCSVText writes the PatientDimension struct in a way that can be added to a .csv file - "","","", etc.
func (pd PatientDimension) ToCSVText(empty bool) string {
	b := pd.EncryptedFlag.ToBytes()
	encodedEncryptedFlag := "\"" + base64.StdEncoding.EncodeToString(b) + "\""

	of := pd.OptionalFields
	ofString := ""
	if empty == false {
		for i := 0; i < len(of); i++ {
			// +4 because there is one pk field and 3 mandatory fields
			ofString += "\"" + of[i].Value + "\","
		}

		acString := "\"" + pd.AdminColumns.UpdateDate + "\"," + "\"" + pd.AdminColumns.DownloadDate + "\"," + "\"" + pd.AdminColumns.ImportDate + "\"," + "\"" + pd.AdminColumns.SourceSystemCD + "\"," + "\"" + pd.AdminColumns.UploadID + "\""
		return pd.PK.ToCSVText() + ",\"" + pd.VitalStatusCD + "\"," + "\"" + pd.BirthDate + "\"," + "\"" + pd.DeathDate + "\"," + ofString[:len(ofString)-1] + "," + acString + "," + encodedEncryptedFlag
	} else {
		for i := 0; i < len(of); i++ {
			// +4 because there is one pk field and 3 mandatory fields
			ofString += "\"" + "\","
		}

		acString := "\"" + "\"," + "\"" + "\"," + "\"" + "\"," + "\"" + "\"," + "\"" + "\""
		return pd.PK.ToCSVText() + ",\"" + "\"," + "\"" + "\"," + "\"" + "\"," + ofString[:len(ofString)-1] + "," + acString + "," + encodedEncryptedFlag
	}
}

// OptionalFields table contains the optional fields
type OptionalFields struct {
	ValType string
	Value   string
}

//-------------------------------------//

// MapNewEncounterNum maps [old_patient_num old_encounter_num] -> [new_patient_num new_encounter_num].
// For the dummies the [old_patient_num old_encounter_num] refers to the original values
var MapNewEncounterNum map[VisitDimensionPK]VisitDimensionPK

// MapPatientVisits maps a patient_num to all its encounter_nums
var MapPatientVisits map[string][]string

// MaxVisits keeps track of the maximum number of visits of all the patients
var MaxVisits int

//-------------------------------------//

// TableVisitDimension is visit_dimension table
var TableVisitDimension map[VisitDimensionPK]VisitDimension

// HeaderVisitDimension contains all the headers for the visit_dimension table
var HeaderVisitDimension []string

// VisitDimension table represents a visit in the database
type VisitDimension struct {
	PK             	VisitDimensionPK
	ActiveStatusCD 	string
	StartDate      	string
	EndDate      	string
	OptionalFields 	[]OptionalFields
	AdminColumns   	AdministrativeColumns
}

// VisitDimensionPK is the primary key of the visit_dimension table
type VisitDimensionPK struct {
	EncounterNum 	string
	PatientNum 		string
}

// ToCSVText writes the VisitDimensionPK struct in a way that can be added to a .csv file - "","","", etc.
func (vdk VisitDimensionPK) ToCSVText() string {
	return "\"" + vdk.EncounterNum + "\"," + "\"" + vdk.PatientNum + "\""
}

// ToCSVText writes the VisitDimension struct in a way that can be added to a .csv file - "","","", etc.
func (vd VisitDimension) ToCSVText(empty bool) string {
	of := vd.OptionalFields
	ofString := ""
	if empty == false {
		for i := 0; i < len(of); i++ {
			// +4 because there is two pk field and 3 mandatory fields
			ofString += "\"" + of[i].Value + "\","
		}

		acString := "\"" + vd.AdminColumns.UpdateDate + "\"," + "\"" + vd.AdminColumns.DownloadDate + "\"," + "\"" + vd.AdminColumns.ImportDate + "\"," + "\"" + vd.AdminColumns.SourceSystemCD + "\"," + "\"" + vd.AdminColumns.UploadID + "\""
		return vd.PK.ToCSVText() + ",\"" + vd.ActiveStatusCD + "\"," + "\"" + vd.StartDate + "\"," + "\"" + vd.EndDate + "\"," + ofString[:len(ofString)-1] + "," + acString
	} else {
		for i := 0; i < len(of); i++ {
			// +4 because there is on pk field and 3 mandatory fields
			ofString += "\"" + "\","
		}

		acString := "\"" + "\"," + "\"" + "\"," + "\"" + "\"," + "\"" + "\"," + "\""  + "\""
		return vd.PK.ToCSVText() + ",\"" + "\"," + "\"" + "\"," + "\"" + "\"," + ofString[:len(ofString)-1] + "," + acString
	}
}

//-------------------------------------//

// TableConceptDimension is concept_dimension table
var TableConceptDimension map[*ConceptDimensionPK]ConceptDimension

// HeaderConceptDimension contains all the headers for the concept_dimension table
var HeaderConceptDimension []string

// MapConceptCodeToTag maps the concept code (in the concept dimension) to the tag ID value (for the sensitive terms)
var MapConceptCodeToTag map[string]int64

// ConceptDimension table contains one row for each concept
type ConceptDimension struct {
	PK           *ConceptDimensionPK
	ConceptCD    string
	NameChar     string
	ConceptBlob  string
	AdminColumns AdministrativeColumns
}

// ConceptDimensionPK is the primary key of the Concept_Dimension table
type ConceptDimensionPK struct {
	ConceptPath string
}

// ToCSVText writes the ConceptDimension object in a way that can be added to a .csv file - "","","", etc.
func (cd ConceptDimension) ToCSVText() string {
	acString := "\"" + cd.AdminColumns.UpdateDate + "\"," + "\"" + cd.AdminColumns.DownloadDate + "\"," + "\"" + cd.AdminColumns.ImportDate + "\"," + "\"" + cd.AdminColumns.SourceSystemCD + "\"," + "\"" + cd.AdminColumns.UploadID + "\""
	return "\"" + cd.PK.ConceptPath + "\"," + "\"" + cd.ConceptCD + "\"," + "\"" + cd.NameChar + "\"," + "\"" + cd.ConceptBlob + "\"," + acString
}

// ConceptDimensionSensitiveToCSVText writes the tagging information of a concept of the concept_dimension table in a way that can be added to a .csv file - "","","", etc.
func ConceptDimensionSensitiveToCSVText(tag *libunlynx.GroupingKey, tagID int64) string {
	return `"\medco\tagged\concept\` + string(*tag) + `\", "TAG_ID:` + strconv.FormatInt(tagID, 10) + `", "\N", "\N", "\N", "\N", "NOW()", "\N", "\N"`
}

//-------------------------------------//

// TableModifierDimension is modifier_dimension table
var TableModifierDimension map[*ModifierDimensionPK]ModifierDimension

// HeaderModifierDimension contains all the headers for the modifier_dimension table
var HeaderModifierDimension []string

// MapModifierCodeToTag maps the modifier code (in the concept dimension) to the tag ID value (for the sensitive terms)
var MapModifierCodeToTag map[string]int64

// ModifierDimension table contains one row for each modifier
type ModifierDimension struct {
	PK           *ModifierDimensionPK
	ModifierCD   string
	NameChar     string
	ModifierBlob string
	AdminColumns AdministrativeColumns
}

// ModifierDimensionPK is the primary key of the Modifier_Dimension table
type ModifierDimensionPK struct {
	ModifierPath string
}

// ToCSVText writes the ModifierDimension object in a way that can be added to a .csv file - "","","", etc.
func (md ModifierDimension) ToCSVText() string {
	acString := "\"" + md.AdminColumns.UpdateDate + "\"," + "\"" + md.AdminColumns.DownloadDate + "\"," + "\"" + md.AdminColumns.ImportDate + "\"," + "\"" + md.AdminColumns.SourceSystemCD + "\"," + "\"" + md.AdminColumns.UploadID + "\""
	return "\"" + md.PK.ModifierPath + "\"," + "\"" + md.ModifierCD + "\"," + "\"" + md.NameChar + "\"," + "\"" + md.ModifierBlob + "\"," + acString
}

// ModifierDimensionSensitiveToCSVText writes the tagging information of a concept of the modifier_dimension table in a way that can be added to a .csv file - "","","", etc.
func ModifierDimensionSensitiveToCSVText(tag *libunlynx.GroupingKey, tagID int64) string {
	return `"\medco\tagged\modifier\` + string(*tag) + `\", "TAG_ID:` + strconv.FormatInt(tagID, 10) + `", "\N", "\N", "\N", "\N", "NOW()", "\N", "\N"`
}

//-------------------------------------//

// Am maps a shrine ontology sensitive concept or modifier concept to the local ontology (we need this to know which concepts from the local ontology are sensitive)
var Am AdapterMappings

// AdapterMappings is the xml pre-generated struct to parse the AdapterMappings.xml
type AdapterMappings struct {
	Hostname    string  `xml:"hostname"`
	TimeStamp   string  `xml:"timestamp"`
	ListEntries []Entry `xml:"mappings>entry"`
}

// Entry is part of the AdapterMappings.xml
type Entry struct {
	Key           string   `xml:"key"`
	ListLocalKeys []string `xml:"value>local_key"`
}

// SUPPORT FUNCTIONS

// ShrineOntologyFromString generates a ShrineOntology struct from a parsed line of a .csv file
func ShrineOntologyFromString(line []string) *ShrineOntology {
	size := len(line)

	ac := AdministrativeColumns{
		UpdateDate:     line[size-7],
		DownloadDate:   line[size-6],
		ImportDate:     line[size-5],
		SourceSystemCD: line[size-4],
	}

	so := &ShrineOntology{
		NodeEncryptID:      int64(-1), //signals that this shrine ontology element is not sensitive so no need for an encrypt ID
		ChildrenEncryptIDs: nil,       //same thing as before
		HLevel:             line[0],
		Fullname:           line[1],
		Name:               line[2],
		SynonymCD:          line[3],
		VisualAttributes:   line[4],
		TotalNum:           line[5],
		BaseCode:           line[6],
		MetadataXML:        strings.Replace(line[7], "\"", "\"\"", -1),
		FactTableColumn:    line[8],
		Tablename:          line[9],
		ColumnName:         line[10],
		ColumnDataType:     line[11],
		Operator:           line[12],
		DimCode:            line[13],
		Comment:            line[14],
		Tooltip:            line[15],
		AdminColumns:       ac,
		ValueTypeCD:        line[20],
		AppliedPath:        line[21],
		ExclusionCD:        line[22],
	}

	return so
}

// LocalOntologyFromString generates a LocalOntology struct from a parsed line of a .csv file
func LocalOntologyFromString(line []string) *LocalOntology {

	size := len(line)

	ac := AdministrativeColumns{
		UpdateDate:     line[size-8],
		DownloadDate:   line[size-7],
		ImportDate:     line[size-6],
		SourceSystemCD: line[size-5],
	}

	so := &LocalOntology{
		HLevel:           line[0],
		Fullname:         line[1],
		Name:             line[2],
		SynonymCD:        line[3],
		VisualAttributes: line[4],
		TotalNum:         line[5],
		BaseCode:         line[6],
		MetadataXML:      strings.Replace(line[7], "\"", "\"\"", -1),
		FactTableColumn:  line[8],
		Tablename:        line[9],
		ColumnName:       line[10],
		ColumnDataType:   line[11],
		Operator:         line[12],
		DimCode:          line[13],
		Comment:          line[14],
		Tooltip:          line[15],
		AppliedPath:      line[16],
		AdminColumns:     ac,
		ValueTypeCD:      line[21],
		ExclusionCD:      line[22],
		Path:             line[23],
		Symbol:           line[24],
	}

	return so

}

// PatientDimensionFromString generates a PatientDimension struct from a parsed line of a .csv file
func PatientDimensionFromString(line []string, pk kyber.Point) (PatientDimensionPK, PatientDimension) {
	pdk := PatientDimensionPK{
		PatientNum: line[0],
	}

	pd := PatientDimension{
		PK:            pdk,
		VitalStatusCD: line[1],
		BirthDate:     line[2],
		DeathDate:     line[3],
	}

	size := len(line)

	// optional fields
	of := make([]OptionalFields, 0)

	for i := 4; i < size-5; i++ {
		of = append(of, OptionalFields{ValType: HeaderPatientDimension[i], Value: line[i]})
	}

	ac := AdministrativeColumns{
		UpdateDate:     line[size-5],
		DownloadDate:   line[size-4],
		ImportDate:     line[size-3],
		SourceSystemCD: line[size-2],
		UploadID:       line[size-1],
	}

	ef := libunlynx.EncryptInt(pk, 1)

	pd.OptionalFields = of
	pd.AdminColumns = ac
	pd.EncryptedFlag = *ef

	return pdk, pd
}

// VisitDimensionFromString generates a VisitDimension struct from a parsed line of a .csv file
func VisitDimensionFromString(line []string) (VisitDimensionPK, VisitDimension) {
	vdk := VisitDimensionPK{
		EncounterNum: 	line[0],
		PatientNum:  	line[1],
	}

	vd := VisitDimension{
		PK:            	vdk,
		ActiveStatusCD: line[2],
		StartDate:     	line[3],
		EndDate:     	line[4],
	}

	size := len(line)

	// optional fields
	of := make([]OptionalFields, 0)

	for i := 5; i < size-5; i++ {
		of = append(of, OptionalFields{ValType: HeaderPatientDimension[i], Value: line[i]})
	}

	ac := AdministrativeColumns{
		UpdateDate:     line[size-5],
		DownloadDate:   line[size-4],
		ImportDate:     line[size-3],
		SourceSystemCD: line[size-2],
		UploadID:       line[size-1],
	}

	vd.OptionalFields = of
	vd.AdminColumns = ac

	return vdk, vd
}

// ConceptDimensionFromString generates a ConceptDimension struct from a parsed line of a .csv file
func ConceptDimensionFromString(line []string) (*ConceptDimensionPK, ConceptDimension) {
	cdk := &ConceptDimensionPK{
		ConceptPath: line[0],
	}

	cd := ConceptDimension{
		PK:          cdk,
		ConceptCD:   line[1],
		NameChar:    line[2],
		ConceptBlob: line[3],
	}

	ac := AdministrativeColumns{
		UpdateDate:     line[4],
		DownloadDate:   line[5],
		ImportDate:     line[6],
		SourceSystemCD: line[7],
		UploadID:       line[8],
	}

	cd.AdminColumns = ac

	return cdk, cd
}

// ModifierDimensionFromString generates a ModifierDimension struct from a parsed line of a .csv file
func ModifierDimensionFromString(line []string) (*ModifierDimensionPK, ModifierDimension) {
	mdk := &ModifierDimensionPK{
		ModifierPath: line[0],
	}

	md := ModifierDimension{
		PK:           mdk,
		ModifierCD:   line[1],
		NameChar:     line[2],
		ModifierBlob: line[3],
	}

	ac := AdministrativeColumns{
		UpdateDate:     line[4],
		DownloadDate:   line[5],
		ImportDate:     line[6],
		SourceSystemCD: line[7],
		UploadID:       line[8],
	}

	md.AdminColumns = ac

	return mdk, md
}

// ObservationFactFromString generates a ObservationFact struct from a parsed line of a .csv file
func ObservationFactFromString(line []string) (*ObservationFactPK, ObservationFact) {
	ofk := &ObservationFactPK{
		EncounterNum: line[0],
		PatientNum:   line[1],
		ConceptCD:    line[2],
		ProviderID:   line[3],
		StartDate:    line[4],
		ModifierCD:   line[5],
		InstanceNum:  line[6],
	}

	of := ObservationFact{
		PK:              ofk,
		ValTypeCD:       line[7],
		TValChar:        line[8],
		NValNum:         line[9],
		ValueFlagCD:     line[10],
		QuantityNum:     line[11],
		UnitsCD:         line[12],
		EndDate:         line[13],
		LocationCD:      line[14],
		ObservationBlob: line[15],
		ConfidenceNum:   line[16],
	}

	ac := AdministrativeColumns{
		UpdateDate:      line[17],
		DownloadDate:    line[18],
		ImportDate:      line[19],
		SourceSystemCD:  line[20],
		UploadID:        line[21],
		TextSearchIndex: line[22],
	}

	of.AdminColumns = ac

	return ofk, of
}
