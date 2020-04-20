package loaderi2b2

import (
	"os"
	"strconv"

	"go.dedis.ch/onet/v3/log"

	libunlynx "github.com/ldsec/unlynx/lib"
	"go.dedis.ch/onet/v3"
)

////////// new format
func ParseSurvivalTypeDimension() error {
	lines, err := readCSV("SURVIVAL_TYPE_DIMENSION")
	if err != nil {
		return err
	}

	TableSurvivalTypeDimension = make(map[*SurvivalTypeDimensionPK]SurvivalTypeDimension)
	HeaderSurvivalTypeDimension = make([]string, 0)
	MapSurvivalTypeCodeToTag = make(map[string]int64)

	/*
		Struct of survival type dimensiontable
		survival_type_path,   PK
		survival_type_code,
	*/

	for _, header := range lines[0] {
		HeaderSurvivalTypeDimension = append(HeaderSurvivalTypeDimension, header)
	}

	//skip header
	for _, line := range lines[1:] {
		//for the moment, flat path as single string
		stpk := &SurvivalTypeDimensionPK{Path: line[0]}

		TableSurvivalTypeDimension[stpk] = SurvivalTypeDimension{PK: stpk, KindCode: line[1]}
	}

	return nil
}

func ConvertSurvivalTypeDimension() error {
	csvOutputFile, err := os.Create(OutputFilePaths["SURVIVAL_TYPE_DIMENSION"].Path)
	if err != nil {
		log.Fatal("Error opening [" + OutputFilePaths["SURVIVAL_TYPE_DIMENSION"].TableName + "].csv " + err.Error())
		return err
	}
	defer csvOutputFile.Close()
	headerString := ""
	for _, header := range HeaderSurvivalTypeDimension {
		headerString += "\"" + header + "\","
	}
	csvOutputFile.WriteString(headerString[:len(headerString)-1] + "\n")

	for _, cd := range TableSurvivalTypeDimension {
		// if the concept is non-sensitive -> keep it as it is
		if _, ok := TableLocalOntologyClear[cd.PK.Path]; ok {
			csvOutputFile.WriteString(cd.ToCSVText() + "\n")
			// if the concept is sensitive -> fetch its encrypted tag and tag_id
		} else if _, ok := MapConceptPathToTag[cd.PK.Path]; ok {
			temp := MapConceptPathToTag[cd.PK.Path].Tag
			csvOutputFile.WriteString(SurvivalTypeDimensionSensitiveToCSVText(&temp, MapConceptPathToTag[cd.PK.Path].TagID) + "\n")
			MapSurvivalTypeCodeToTag[cd.KindCode] = MapConceptPathToTag[cd.PK.Path].TagID
			// if the concept does not exist in the LocalOntology and none of his siblings is sensitive
		} else if _, ok := HasSensitiveParents(cd.PK.Path); !ok {
			csvOutputFile.WriteString(cd.ToCSVText() + "\n")
		} else {
			ListConceptsToIgnore[cd.KindCode] = struct{}{}
		}
	}
	return nil
}

func ParseTimeDimension() error {
	lines, err := readCSV("TIME_DIMENSION")
	if err != nil {
		return err
	}

	TableTimeDimension = make(map[*TimeDimensionPK]TimeDimension)
	HeaderTimeDimension = make([]string, 0)
	MapTimeCodeToTag = make(map[string]int64)

	/*
		Struct of time dimension table

		time_path,   PK
		time_code

	*/

	for _, header := range lines[0] {
		HeaderTimeDimension = append(HeaderTimeDimension, header)
	}

	//skip header
	for _, line := range lines[1:] {
		//path as flat string for the moment
		tpk := &TimeDimensionPK{Path: line[0]}

		TableTimeDimension[tpk] = TimeDimension{PK: tpk, TimeCode: line[1]}
	}

	return nil

}

func ConvertTimeDimension() error {
	csvOutputFile, err := os.Create(OutputFilePaths["TIME_DIMENSION"].Path)
	if err != nil {
		log.Fatal("Error opening [" + OutputFilePaths["TIME_DIMENSION"].TableName + "].csv " + err.Error())
		return err
	}
	defer csvOutputFile.Close()
	headerString := ""
	for _, header := range HeaderTimeDimension {
		headerString += "\"" + header + "\","
	}
	csvOutputFile.WriteString(headerString[:len(headerString)-1] + "\n")

	for _, cd := range TableTimeDimension {
		// if the concept is non-sensitive -> keep it as it is
		if _, ok := TableLocalOntologyClear[cd.PK.Path]; ok {
			csvOutputFile.WriteString(cd.ToCSVText() + "\n")
			// if the concept is sensitive -> fetch its encrypted tag and tag_id
		} else if _, ok := MapConceptPathToTag[cd.PK.Path]; ok {
			temp := MapConceptPathToTag[cd.PK.Path].Tag
			csvOutputFile.WriteString(TimeDimensionSensitiveToCSVText(&temp, MapConceptPathToTag[cd.PK.Path].TagID) + "\n")
			MapTimeCodeToTag[cd.TimeCode] = MapConceptPathToTag[cd.PK.Path].TagID
			// if the concept does not exist in the LocalOntology and none of his siblings is sensitive
		} else if _, ok := HasSensitiveParents(cd.PK.Path); !ok {
			csvOutputFile.WriteString(cd.ToCSVText() + "\n")
		} else {
			ListConceptsToIgnore[cd.TimeCode] = struct{}{}
		}
	}
	return nil

}

func ParseSurvivalFact() error {
	lines, err := readCSV("SURVIVAL_FACT")
	if err != nil {
		return err
	}
	TableSurvivalFact = make(map[*SurvivalFactPK]SurvivalFact)
	EventObservationBlob = make(map[*ObservationFactPK]string)
	HeaderSurvivalFact = make([]string, 0)
	MapPatientObs = make(map[string][]*ObservationFactPK)
	MapDummyObs = make(map[string][]*ObservationFactPK)
	MapPatientSurv = make(map[string][]*SurvivalFactPK)
	MapDummySurv = make(map[string]struct{})

	/*

		structure of survivalFact table
		patient_num,        part of PK
		survival_type_code, part of PK
		time_code,          part of PK
		event_of_interest,
		censorig_event


	*/

	for _, header := range lines[0] {
		HeaderSurvivalFact = append(HeaderSurvivalFact, header)
	}

	for _, line := range lines[1:] {
		sfpk := &SurvivalFactPK{PatientNum: line[0], KindCode: line[1], TimeCode: line[2]}
		sf := SurvivalFact{PK: sfpk, EventOfInterest: line[3], CensoringEvent: line[4]}
		_, isTimeToIgnores := ListConceptsToIgnore[sfpk.TimeCode]
		_, isSurvivalTypeToIgnore := ListConceptsToIgnore[sfpk.KindCode]
		if !isTimeToIgnores && !isSurvivalTypeToIgnore {
			TableSurvivalFact[sfpk] = sf
			if obsList, ok := MapPatientSurv[sfpk.PatientNum]; !ok {
				MapPatientSurv[sfpk.PatientNum] = []*SurvivalFactPK{sfpk}
			} else {
				obsList = append(obsList, sfpk)
			}
			if originalPatient, ok := TableDummyToPatient[sfpk.PatientNum]; ok {
				MapDummyObs[sfpk.PatientNum] = MapPatientObs[originalPatient]
			}
		}
	}

	return nil

}

func ConvertSurvivalFact(group *onet.Roster) error {
	csvOutputFile, err := os.Create(OutputFilePaths["SURVIVAL_FACT"].Path)
	if err != nil {
		log.Fatal("Error opening [" + OutputFilePaths["SURVIVAL_FACT"].Path + "].csv")
		return err
	}
	defer csvOutputFile.Close()
	headerString := ""
	for _, header := range HeaderSurvivalFact {
		headerString += "\"" + header + "\","
	}
	// remove the last ,
	csvOutputFile.WriteString(headerString[:len(headerString)-1] + "\n")

	for sfk, sf := range TableSurvivalFact {

		copySurv := sf

		// if dummy observation
		if _, ok := TableDummyToPatient[sfk.PatientNum]; ok {

			// 3. change patient_num
			// 4. if the observation is a survival analysis recpord, add the blob
			if _, ok := MapDummySurv[sf.PK.PatientNum]; !ok {

				log.Fatal("assertion: the patient number should be in TableDummyToPatient as well as in MapdummySrv")

			}

			eoi := libunlynx.EncryptInt(group.Aggregate, int64(0))
			ce := libunlynx.EncryptInt(group.Aggregate, int64(0))
			copySurv.EventOfInterest, err = eoi.Serialize()
			if err != nil {
				return err
			}
			copySurv.CensoringEvent, err = ce.Serialize()
			if err != nil {
				return err
			}
		} else {
			eoiInt, err := strconv.ParseInt(sf.CensoringEvent, 10, 64)
			if err != nil {
				return err
			}
			ceInt, err := strconv.ParseInt(sf.EventOfInterest, 10, 64)
			if err != nil {
				return err
			}
			eoi := libunlynx.EncryptInt(group.Aggregate, eoiInt)
			ce := libunlynx.EncryptInt(group.Aggregate, ceInt)

			copySurv.EventOfInterest, err = eoi.Serialize()
			if err != nil {
				return err
			}
			copySurv.CensoringEvent, err = ce.Serialize()
			if err != nil {
				return err
			}

		}

		// if the concept is sensitive we replace its code with the correspondent tag ID
		if tag, ok := MapTimeCodeToTag[copySurv.PK.TimeCode]; ok {
			copySurv.PK.TimeCode = "TAG_ID:" + strconv.FormatInt(tag, 10)
		}

		if tag, ok := MapSurvivalTypeCodeToTag[copySurv.PK.KindCode]; ok {
			copySurv.PK.KindCode = "TAG_ID:" + strconv.FormatInt(tag, 10)
		}

		csvOutputFile.WriteString(copySurv.ToCSVText() + "\n")

	}
	return nil
}

func SurvivalTablesLoadingScript() (loading string) {

	loading += "CREATE TABLE IF NOT EXISTS " + I2B2METADATA + "survival_ontology (LIKE " + I2B2METADATA + "i2b2);"

	loading += "CREATE SCHEMA IF NOT EXISTS " + SURVIVALDEMODATA[:len(SURVIVALDEMODATA)-1] + " "
	loading += "AUTHORIZATION i2b2;\n"
	loading += "CREATE TABLE IF NOT EXISTS " + SURVIVALDEMODATA + "survival_fact (patient_num int, survival_type_code varchar(50), time_code  varchar(50), event_of_interest text, censoring_event text, PRIMARY KEY (patient_num, survival_type_code,time_code ));\n"
	loading += "CREATE TABLE IF NOT EXISTS " + SURVIVALDEMODATA + "survival_type_dimension (survival_type_path varchar(700) PRIMARY KEY,survival_type_code varchar(50));\n"
	loading += "CREATE TABLE IF NOT EXISTS " + SURVIVALDEMODATA + "time_dimension (time_path varchar(700) PRIMARY KEY, time_code varchar(50));\n"

	loading += "TRUNCATE TABLE " + SURVIVALDEMODATA + "survival_fact;\n"
	loading += "TRUNCATE TABLE " + SURVIVALDEMODATA + "survival_type_dimension;\n"
	loading += "TRUNCATE TABLE " + SURVIVALDEMODATA + "time_dimension;\n"

	loading += `CREATE OR REPLACE FUNCTION grouping_filter(patient integer,groupss integer[][])  RETURNS int AS \$\$` + "\n"
	loading += `	DECLARE` + "\n"
	loading += `		group_ int[];` + "\n"
	loading += `		res int;` + "\n"
	loading += `		patient int;` + "\n"
	loading += `		group_idx int := 0;` + "\n"
	loading += `    BEGIN` + "\n"
	loading += `		group_idx=0;` + "\n"
	loading += `		FOREACH group_ IN ARRAY groupss LOOP` + "\n"
	loading += `			IF patient = ANY (group_) THEN` + "\n"
	loading += `				res=group_idx;` + "\n"
	loading += `			END IF;`
	loading += `			group_idx = group_idx +1;` + "\n"
	loading += `		END LOOP;` + "\n"
	loading += `        RETURN  res;` + "\n"
	loading += `    END;` + "\n"
	loading += `\$\$ LANGUAGE plpgsql STABLE;` + "\n"

	loading += `\copy ` + OutputFilePaths["SURVIVAL_FACT"].TableName + ` FROM '` + OutputFilePaths["SURVIVAL_FACT"].Path + `' ESCAPE '"' DELIMITER ',' CSV HEADER;` + "\n"
	loading += `\copy ` + OutputFilePaths["SURVIVAL_TYPE_DIMENSION"].TableName + ` FROM '` + OutputFilePaths["SURVIVAL_TYPE_DIMENSION"].Path + `' ESCAPE '"' DELIMITER ',' CSV HEADER;` + "\n"
	loading += `\copy ` + OutputFilePaths["TIME_DIMENSION"].TableName + ` FROM '` + OutputFilePaths["TIME_DIMENSION"].Path + `' ESCAPE '"' DELIMITER ',' CSV HEADER;` + "\n"

	return

}
