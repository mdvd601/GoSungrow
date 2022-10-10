package getAreaList

import (
	"GoSungrow/iSolarCloud/api"
	"GoSungrow/iSolarCloud/api/apiReflect"
	"fmt"
)

const Url = "/v1/powerStationService/getAreaList"
const Disabled = false

type RequestData struct {
}

func (rd RequestData) IsValid() error {
	return apiReflect.VerifyOptionsRequired(rd)
}

func (rd RequestData) Help() string {
	ret := fmt.Sprintf("")
	return ret
}

type ResultData struct {
	PageList []struct {
		FaultStationCount api.Integer   `json:"fault_station_count"`
		IsHaveEsPs        string        `json:"is_have_es_ps"`
		IsLeaf            api.Integer   `json:"is_leaf"`
		OrgID             api.Integer   `json:"org_id"`
		OrgName           string        `json:"org_name"`
		P83048            api.UnitValue `json:"p83048"`
		P83049            api.UnitValue `json:"p83049"`
		P83050            api.UnitValue `json:"p83050"`
		P83051            api.UnitValue `json:"p83051"`
		PlanMonth         string        `json:"plan_month"`
		StationCount      api.Integer   `json:"station_count"`
		TodayEnergy       api.UnitValue `json:"today_energy"`
		TotalEnergy       api.UnitValue `json:"total_energy"`
	} `json:"pageList"`
	RowCount api.Integer `json:"rowCount"`
}

func (e *ResultData) IsValid() error {
	var err error
	// switch {
	// case e.Dummy == "":
	//	break
	// default:
	//	err = errors.New(fmt.Sprintf("unknown error '%s'", e.Dummy))
	// }
	return err
}

// type DecodeResultData ResultData
//
// func (e *ResultData) UnmarshalJSON(data []byte) error {
//	var err error
//
//	for range Only.Once {
//		if len(data) == 0 {
//			break
//		}
//		var pd DecodeResultData
//
//		// Store ResultData
//		_ = json.Unmarshal(data, &pd)
//		e.Dummy = pd.Dummy
//	}
//
//	return err
// }
