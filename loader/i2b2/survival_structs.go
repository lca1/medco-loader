package loaderi2b2

import (
	"fmt"
	"strconv"
	"strings"

	libunlynx "github.com/ldsec/unlynx/lib"
)

//SurvivalKindPK
type SurvivalTypeDimensionPK struct {
	Path string
}

//SurvivalKind
type SurvivalTypeDimension struct {
	PK       *SurvivalTypeDimensionPK
	KindCode string
}

func (sk *SurvivalTypeDimension) ToCSVText() string {
	return fmt.Sprintf(`"%s","%s"`, sk.PK.Path, sk.KindCode)
}

//TimePointPK
type TimeDimensionPK struct {
	Path string
}

//TimePoint
type TimeDimension struct {
	PK       *TimeDimensionPK
	TimeCode string
}

func (td *TimeDimension) ToCSVText() string {
	return fmt.Sprintf(`"%s","%s"`, td.PK.Path, td.TimeCode)
}

//ObservationFactPK TODO
type SurvivalFactPK struct {
	PatientNum string
	TimeCode   string
	KindCode   string
}

//ObservationFact
type SurvivalFact struct {
	PK              *SurvivalFactPK
	EventOfInterest string
	CensoringEvent  string
}

func (sf *SurvivalFact) ToCSVText() string {
	return fmt.Sprintf(`"%s","%s","%s","%s","%s"`, sf.PK.PatientNum, sf.PK.KindCode, sf.PK.TimeCode, sf.EventOfInterest, sf.CensoringEvent)
}

var (
	TableSurvivalFact           map[*SurvivalFactPK]SurvivalFact
	MapPatientSurv              map[string][]*SurvivalFactPK
	MapDummySurv                map[string]struct{}
	TableTimeDimension          map[*TimeDimensionPK]TimeDimension
	TableSurvivalTypeDimension  map[*SurvivalTypeDimensionPK]SurvivalTypeDimension
	HeaderSurvivalTypeDimension []string
	HeaderTimeDimension         []string
	HeaderSurvivalFact          []string

	MapSurvivalTypeCodeToTag map[string]int64
	MapTimeCodeToTag         map[string]int64
)

func SurvivalTypeDimensionSensitiveToCSVText(tag *libunlynx.GroupingKey, tagID int64) string {
	finalString := `"\medco\tagged\concept\` + string(*tag) + `\","TAG_ID:` + strconv.FormatInt(tagID, 10) + `"`

	return strings.Replace(finalString, `"\N"`, "", -1)
}
func TimeDimensionSensitiveToCSVText(tag *libunlynx.GroupingKey, tagID int64) string {
	return SurvivalTypeDimensionSensitiveToCSVText(tag, tagID)
}
