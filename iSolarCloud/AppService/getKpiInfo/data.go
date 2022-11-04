package getKpiInfo

import (
	"GoSungrow/iSolarCloud/api"
	"GoSungrow/iSolarCloud/api/GoStruct"
	"GoSungrow/iSolarCloud/api/GoStruct/reflection"
	"GoSungrow/iSolarCloud/api/GoStruct/valueTypes"
	"fmt"
	"github.com/MickMake/GoUnify/Only"
)

const Url = "/v1/powerStationService/getKpiInfo"
const Disabled = false

type RequestData struct {
}

func (rd RequestData) IsValid() error {
	return GoStruct.VerifyOptionsRequired(rd)
}

func (rd RequestData) Help() string {
	ret := fmt.Sprintf("")
	return ret
}

type ResultData struct {
	ActualEnergy             []valueTypes.Float   `json:"actual_energy" PointUnitFrom:"ActualEnergyUnit" DataTable:"true" DataTableMerge:"true" DataTableShowIndex:"true"`
	PlanEnergy               []valueTypes.Float   `json:"plan_energy" PointUnitFrom:"PlanEnergyUnit" DataTable:"true" DataTableMerge:"true" DataTableShowIndex:"true"`

	ActualEnergyUnit         valueTypes.String    `json:"actual_energy_unit" PointIgnore:"true"`
	ChargeTotalEnergy        valueTypes.Float     `json:"charge_total_energy" PointUnitFrom:"ChargeTotalEnergyUnit" PointUpdateFreq:"UpdateFreqTotal"`
	ChargeTotalEnergyUnit    valueTypes.String    `json:"charge_total_energy_unit" PointIgnore:"true"`
	DisChargeTotalEnergy     valueTypes.Float     `json:"dis_charge_total_energy" PointUnitFrom:"DisChargeTotalEnergyUnit" PointUpdateFreq:"UpdateFreqTotal"`
	DisChargeTotalEnergyUnit valueTypes.String    `json:"dis_charge_total_energy_unit" PointIgnore:"true"`
	MonthEnergy              valueTypes.UnitValue `json:"month_energy" PointUpdateFreq:"UpdateFreqMonth"`
	OrgName                  valueTypes.String    `json:"org_name"`
	P83024                   valueTypes.Float     `json:"p83024" PointUnit:"Wh"`
	PercentPlanMonth         valueTypes.Float     `json:"percent_plan_month" PointUnit:"%" PointUpdateFreq:"UpdateFreqMonth"`
	PercentPlanYear          valueTypes.Float     `json:"percent_plan_year" PointUnit:"%" PointUpdateFreq:"UpdateFreqYear"`
	PlanEnergyUnit           valueTypes.String    `json:"plan_energy_unit" PointIgnore:"true"`
	PsCount                  valueTypes.Integer   `json:"ps_count"`
	TodayEnergy              valueTypes.UnitValue `json:"today_energy" PointUpdateFreq:"UpdateFreqTotal"`
	TotalCapacity            valueTypes.UnitValue `json:"total_capcity" PointId:"total_capacity" PointUpdateFreq:"UpdateFreqTotal"`
	TotalDesignCapacity      valueTypes.UnitValue `json:"total_design_capacity" PointUpdateFreq:"UpdateFreqTotal"`
	TotalEnergy              valueTypes.UnitValue `json:"total_energy" PointUpdateFreq:"UpdateFreqTotal"`
	YearEnergy               valueTypes.UnitValue `json:"year_energy" PointUpdateFreq:"UpdateFreqYear"`
}

func (e *ResultData) IsValid() error {
	var err error
	return err
}

func (e *EndPoint) GetData() api.DataMap {
	entries := api.NewDataMap()

	for range Only.Once {
		pkg := reflection.GetName("", *e)
		// dt := valueTypes.NewDateTime(valueTypes.Now)
		entries.StructToDataMap(*e,  "system", GoStruct.EndPointPath{})

		_ = entries.CopyPoint(pkg + ".p83024", "virtual.system", "p83024", "")
	}

	return entries
}
