package iSolarCloud

import (
	"GoSungrow/Only"
	"GoSungrow/iSolarCloud/AppService/getDeviceList"
	"GoSungrow/iSolarCloud/AppService/getDeviceModelInfoList"
	"GoSungrow/iSolarCloud/AppService/getHouseholdStoragePsReport"
	"GoSungrow/iSolarCloud/AppService/getKpiInfo"
	"GoSungrow/iSolarCloud/AppService/getPowerDevicePointInfo"
	"GoSungrow/iSolarCloud/AppService/getPowerDevicePointNames"
	"GoSungrow/iSolarCloud/AppService/getPowerStationData"
	"GoSungrow/iSolarCloud/AppService/getPsDetail"
	"GoSungrow/iSolarCloud/AppService/getPsDetailWithPsType"
	"GoSungrow/iSolarCloud/AppService/getPsList"
	"GoSungrow/iSolarCloud/AppService/getTemplateList"
	"GoSungrow/iSolarCloud/AppService/queryDeviceList"
	"GoSungrow/iSolarCloud/AppService/queryDeviceRealTimeDataByPsKeys"
	"GoSungrow/iSolarCloud/AppService/queryMutiPointDataList"
	"GoSungrow/iSolarCloud/WebAppService/getMqttConfigInfoByAppkey"
	"GoSungrow/iSolarCloud/WebAppService/queryUserCurveTemplateData"
	"GoSungrow/iSolarCloud/api"
	"GoSungrow/iSolarCloud/api/output"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)


func (sg *SunGrow) GetPointNamesFromTemplate(template string) api.TemplatePoints {
	var ret api.TemplatePoints

	for range Only.Once {
		if template == "" {
			sg.Error = errors.New("no template defined")
			break
		}

		ep := sg.GetByStruct(
			"WebAppService.queryUserCurveTemplateData",
			queryUserCurveTemplateData.RequestData{TemplateID: template},
			time.Hour,
		)
		if sg.Error != nil {
			break
		}

		data := queryUserCurveTemplateData.AssertResultData(ep)
		for dn, dr := range data.PointsData.Devices {
			for _, pr := range dr.Points {
				if pr.Unit == "null" {
					pr.Unit = ""
				}
					ret = append(ret, api.TemplatePoint {
					PsKey:       dn,
					PointId:     "p" + pr.PointID,
					Description: pr.PointName,
					Unit:        pr.Unit,
				})
			}
		}
	}

	return ret
}

func (sg *SunGrow) GetTemplateData(template string, date string, filter string) error {
	for range Only.Once {
		if template == "" {
			template = "8042"
		}

		if date == "" {
			date = api.NewDateTime("").String()
		}
		when := api.NewDateTime(date)

		var psIds []int64
		psIds, sg.Error = sg.StringToPids()
		if sg.Error != nil {
			break
		}

		for _, psId := range psIds {
			pointNames := sg.GetPointNamesFromTemplate(template)
			if sg.Error != nil {
				break
			}

			ep := sg.GetByStruct(
				"AppService.queryMutiPointDataList",
				queryMutiPointDataList.RequestData{
					PsID:           psId,
					PsKey:          pointNames.PrintKeys(),
					Points:         pointNames.PrintPoints(),
					MinuteInterval: "5",
					StartTimeStamp: when.GetDayStartTimestamp(),
					EndTimeStamp:   when.GetDayEndTimestamp(),
				},
				DefaultCacheTimeout,
			)
			if sg.Error != nil {
				break
			}

			// data := queryMutiPointDataList.AssertResultData(ep)
			data := queryMutiPointDataList.Assert(ep)
			table := data.GetDataTable(pointNames)
			if table.Error != nil {
				sg.Error = table.Error
				break
			}

			table.SetTitle("Template %s on %s for %d", template, when.String(), psId)
			table.SetFilePrefix(data.SetFilenamePrefix("%s-%s-%d", when.String(), template, psId))
			table.SetGraphFilter(filter)
			table.SetSaveFile(sg.SaveAsFile)
			table.OutputType = sg.OutputType
			sg.Error = table.Output()
			if sg.Error != nil {
				break
			}
		}
	}

	return sg.Error
}

func (sg *SunGrow) GetTemplatePoints(template string) error {
	for range Only.Once {
		if template == "" {
			template = "8042"
		}

		table := output.NewTable()
		sg.Error = table.SetHeader(
			"PointStruct Id",
			"Description",
			"Unit",
			)
		if sg.Error != nil {
			break
		}

		ss := sg.GetPointNamesFromTemplate(template)
		for _, s := range ss {
			sg.Error = table.AddRow(
				api.NameDevicePoint(s.PsKey, s.PointId),
				s.Description,
				s.Unit,
			)
			if sg.Error != nil {
				break
			}
		}
		if sg.Error != nil {
			break
		}

		table.SetTitle("Template %s", template)
		table.SetFilePrefix(template)
		table.SetGraphFilter("")
		table.SetSaveFile(sg.SaveAsFile)
		table.OutputType = sg.OutputType
		sg.Error = table.Output()
		if sg.Error != nil {
			break
		}
	}

	return sg.Error
}

func (sg *SunGrow) AllCritical() error {
	var ep api.EndPoint
	for range Only.Once {
		ep = sg.GetByJson("AppService.powerDevicePointList", "")
		if ep.IsError() {
			break
		}

		ep = sg.GetByJson("AppService.getPsList", "")
		if ep.IsError() {
			break
		}

		_getPsList := getPsList.AssertResultData(ep)

		for _, psId := range _getPsList.GetPsIds() {
			ep = sg.GetByJson("AppService.queryDeviceList", fmt.Sprintf(`{"ps_id":"%d"}`, psId))
			if ep.IsError() {
				break
			}

			ep = sg.GetByJson("AppService.queryDeviceListForApp", fmt.Sprintf(`{"ps_id":"%d"}`, psId))
			if ep.IsError() {
				break
			}

			ep = sg.GetByJson("WebAppService.showPSView", fmt.Sprintf(`{"ps_id":"%d"}`, psId))
			if ep.IsError() {
				break
			}

			// ep = sg.GetByJson("AppService.findPsType", fmt.Sprintf(`{"ps_id":"%d"}`, psId))
			// if ep.IsError() {
			// 	break
			// }

			ep = sg.GetByJson("AppService.getPowerStatistics", fmt.Sprintf(`{"ps_id":"%d"}`, psId))
			if ep.IsError() {
				break
			}

			ep = sg.GetByJson("AppService.getPsDetail", fmt.Sprintf(`{"ps_id":"%d"}`, psId))
			if ep.IsError() {
				break
			}

			ep = sg.GetByJson("AppService.getPsDetailWithPsType", fmt.Sprintf(`{"ps_id":"%d"}`, psId))
			if ep.IsError() {
				break
			}

			ep = sg.GetByJson("AppService.getPsHealthState", fmt.Sprintf(`{"ps_id":"%d"}`, psId))
			if ep.IsError() {
				break
			}

			ep = sg.GetByJson("AppService.getPsListStaticData", fmt.Sprintf(`{"ps_id":"%d"}`, psId))
			if ep.IsError() {
				break
			}

			ep = sg.GetByJson("AppService.getPsWeatherList", fmt.Sprintf(`{"ps_id":"%d"}`, psId))
			if ep.IsError() {
				break
			}

			// ep = sg.GetByJson("AppService.queryAllPsIdAndName", fmt.Sprintf(`{"ps_id":"%d"}`, psId))
			// if ep.IsError() {
			// 	break
			// }

			// ep = sg.GetByJson("AppService.queryDeviceListByUserId", fmt.Sprintf(`{"ps_id":"%d"}`, psId))
			// if ep.IsError() {
			// 	break
			// }

			ep = sg.GetByJson("AppService.queryDeviceListForApp", fmt.Sprintf(`{"ps_id":"%d"}`, psId))
			if ep.IsError() {
				break
			}
		}
	}

	sg.Error = ep.GetError()
	return sg.Error
}

func (sg *SunGrow) PrintCurrentStats() error {
	var ep api.EndPoint
	for range Only.Once {
		ep = sg.GetByStruct("AppService.getPsList", nil, DefaultCacheTimeout)
		if ep.IsError() {
			break
		}

		_getPsList := getPsList.Assert(ep)
		table := _getPsList.GetDataTable()
		if table.Error != nil {
			sg.Error = table.Error
			break
		}

		table.SetTitle("getPsList")
		table.SetFilePrefix(_getPsList.SetFilenamePrefix(""))
		table.SetGraphFilter("")
		table.SetSaveFile(sg.SaveAsFile)
		table.OutputType = sg.OutputType
		sg.Error = table.Output()
		if sg.Error != nil {
			break
		}

		for _, psId := range _getPsList.GetPsIds() {
			ep = sg.GetByStruct(
				"AppService.queryDeviceList",
				queryDeviceList.RequestData{PsId: strconv.FormatInt(psId, 10)},
				time.Second*60,
			)
			if sg.Error != nil {
				break
			}

			data := queryDeviceList.Assert(ep)
			data.GetData()

			table = data.GetDataTable()
			if table.Error != nil {
				sg.Error = table.Error
				break
			}

			table.SetTitle("queryDeviceList %s", psId)
			table.SetFilePrefix(data.SetFilenamePrefix("%s", psId))
			table.SetGraphFilter("")
			table.SetSaveFile(sg.SaveAsFile)
			table.OutputType = sg.OutputType
			sg.Error = table.Output()
			if sg.Error != nil {
				break
			}
		}
	}

	return sg.Error
}

func (sg *SunGrow) QueryDevice(psId int64) queryDeviceList.EndPoint {
	var ret queryDeviceList.EndPoint
	for range Only.Once {
		if psId == 0 {
			psId, sg.Error = sg.GetPsId()
			if sg.Error != nil {
				break
			}
		}

		// ep = sg.GetByJson("AppService.queryDeviceList", fmt.Sprintf(`{"ps_id":"%d"}`, psId))
		ep := sg.GetByStruct(
			"AppService.queryDeviceList",
			queryDeviceList.RequestData{PsId: strconv.FormatInt(psId, 10)},
			time.Second * 60,
		)
		// if sg.Error != nil {
		// 	break
		// }

		ret = queryDeviceList.Assert(ep)
	}

	return ret
}

func (sg *SunGrow) QueryPs(psId int64) getPsList.EndPoint {
	var ret getPsList.EndPoint
	for range Only.Once {
		if psId == 0 {
			psId, sg.Error = sg.GetPsId()
			if sg.Error != nil {
				break
			}
		}

		// ep = sg.GetByJson("AppService.queryDeviceList", fmt.Sprintf(`{"ps_id":"%d"}`, psId))
		ep := sg.GetByStruct(
			"AppService.getPsList",
			getPsList.RequestData{},
			time.Second * 60,
		)
		// if sg.Error != nil {
		// 	break
		// }

		ret = getPsList.Assert(ep)
	}

	return ret
}

func (sg *SunGrow) GetPointNames(devices ...string) error {
	for range Only.Once {
		if len(devices) == 0 {
			devices = getPowerDevicePointNames.DeviceTypes
		}
		for _, dt := range devices {
			ep := sg.GetByStruct(
				"AppService.getPowerDevicePointNames",
				getPowerDevicePointNames.RequestData{DeviceType: dt},
				DefaultCacheTimeout,
			)
			if sg.Error != nil {
				break
			}

			data := getPowerDevicePointNames.Assert(ep)
			table := data.GetDataTable()
			if table.Error != nil {
				sg.Error = table.Error
				break
			}

			table.SetTitle("getPowerDevicePointNames %s", dt)
			table.SetFilePrefix(data.SetFilenamePrefix(""))
			table.SetGraphFilter("")
			table.SetSaveFile(sg.SaveAsFile)
			table.OutputType = sg.OutputType
			sg.Error = table.Output()
			if sg.Error != nil {
				break
			}
		}
	}

	return sg.Error
}

func (sg *SunGrow) GetTemplates() error {
	for range Only.Once {
		ep := sg.GetByStruct(
			"AppService.getTemplateList",
			getTemplateList.RequestData{},
			DefaultCacheTimeout,
		)
		if sg.Error != nil {
			break
		}

		data := getTemplateList.Assert(ep)
		table := data.GetDataTable()
		if table.Error != nil {
			sg.Error = table.Error
			break
		}

		table.SetTitle("getTemplateList")
		table.SetFilePrefix(data.SetFilenamePrefix(""))
		table.SetGraphFilter("")
		table.SetSaveFile(sg.SaveAsFile)
		table.OutputType = sg.OutputType
		sg.Error = table.Output()
		if sg.Error != nil {
			break
		}
	}

	return sg.Error
}

func (sg *SunGrow) GetIsolarcloudMqtt(appKey string) error {
	for range Only.Once {
		if appKey == "" {
			appKey = sg.GetAppKey()
		}

		ep := sg.GetByStruct(
			"WebAppService.getMqttConfigInfoByAppkey",
			getMqttConfigInfoByAppkey.RequestData{AppKey: appKey},
			DefaultCacheTimeout,
		)
		if sg.Error != nil {
			break
		}

		data := getMqttConfigInfoByAppkey.Assert(ep)
		table := data.GetDataTable()
		if table.Error != nil {
			sg.Error = table.Error
			break
		}

		table.SetTitle("MQTT info")
		table.SetFilePrefix(data.SetFilenamePrefix(""))
		table.SetGraphFilter("")
		table.SetSaveFile(sg.SaveAsFile)
		table.OutputType = sg.OutputType
		sg.Error = table.Output()
		if sg.Error != nil {
			break
		}
	}

	return sg.Error
}

func (sg *SunGrow) GetRealTimeData(psKey string) error {
	for range Only.Once {
		if psKey == "" {
			var psKeys []string
			psKeys, sg.Error = sg.GetPsKeys()
			if sg.Error != nil {
				break
			}
			fmt.Printf("psKeys: %v\n", psKeys)
			psKey = strings.Join(psKeys, ",")
		}

		ep := sg.GetByStruct(
			"AppService.queryDeviceRealTimeDataByPsKeys",
			queryDeviceRealTimeDataByPsKeys.RequestData{PsKeyList: psKey},
			DefaultCacheTimeout,
		)
		if sg.Error != nil {
			break
		}

		data := queryDeviceRealTimeDataByPsKeys.Assert(ep)
		table := data.GetDataTable()
		if table.Error != nil {
			sg.Error = table.Error
			break
		}

		table.SetTitle("Real Time Data %s", psKey)
		table.SetFilePrefix(data.SetFilenamePrefix(""))
		table.SetGraphFilter("")
		table.SetSaveFile(sg.SaveAsFile)
		table.OutputType = sg.OutputType
		sg.Error = table.Output()
		if sg.Error != nil {
			break
		}
	}

	return sg.Error
}

func (sg *SunGrow) GetPsDetails(psIds ...int64) error {
	for range Only.Once {
		for _, psId := range psIds {
			ep := sg.GetByStruct(
				"AppService.getPsDetailWithPsType",
				getPsDetailWithPsType.RequestData{PsId: strconv.FormatInt(psId, 10)},
				DefaultCacheTimeout)
			if ep.IsError() {
				sg.Error = ep.GetError()
				break
			}

			data := getPsDetailWithPsType.Assert(ep)
			table := data.GetDataTable()
			if table.Error != nil {
				sg.Error = table.Error
				break
			}

			table.SetTitle("PS Details %s", psId)
			table.SetFilePrefix(data.SetFilenamePrefix("%d", psId))
			table.SetGraphFilter("")
			table.SetSaveFile(sg.SaveAsFile)
			table.OutputType = sg.OutputType
			sg.Error = table.Output()
			if sg.Error != nil {
				break
			}
		}
	}

	return sg.Error
}

func (sg *SunGrow) GetPointData(date string, pointNames api.TemplatePoints, psIds ...int64) error {
	for range Only.Once {
		if len(pointNames) == 0 {
			sg.Error = errors.New("no points defined")
			break
		}

		if date == "" {
			date = api.NewDateTime("").String()
		}
		when := api.NewDateTime(date)

		if len(psIds) == 0 {
			psIds, sg.Error = sg.StringToPids()
			if sg.Error != nil {
				break
			}
		}

		for _, psId := range psIds {
			ep := sg.GetByStruct(
				"AppService.queryMutiPointDataList",
				queryMutiPointDataList.RequestData{
					PsID:           psId,
					PsKey:          pointNames.PrintKeys(),
					Points:         pointNames.PrintPoints(),
					MinuteInterval: "5",
					StartTimeStamp: when.GetDayStartTimestamp(),
					EndTimeStamp:   when.GetDayEndTimestamp(),
				},
				DefaultCacheTimeout,
			)
			if sg.Error != nil {
				break
			}

			data := queryMutiPointDataList.Assert(ep)
			table := data.GetDataTable(pointNames)
			if table.Error != nil {
				sg.Error = table.Error
				break
			}

			table.SetTitle("Point Data %s", psId)
			table.SetFilePrefix(data.SetFilenamePrefix("%d", psId))
			table.SetGraphFilter("")
			table.SetSaveFile(sg.SaveAsFile)
			table.OutputType = sg.OutputType
			sg.Error = table.Output()
			if sg.Error != nil {
				break
			}
		}
	}

	return sg.Error
}

func (sg *SunGrow) SearchPointNames(pns ...string) error {
	for range Only.Once {
		table := output.NewTable()

		_ = table.SetHeader(
			"DeviceType",
			"Id",
			"Period",
			"Point Id",
			"Point Name",
			"Show Point Name",
			"Translation Id",
		)

		if len(pns) == 0 {
			fmt.Println("Searching up to id 1000 within getPowerDevicePointInfo")
			for pni := 0; pni < 1000; pni++ {
				PrintPause(pni, 20)

				ep := sg.GetByStruct(
					"AppService.getPowerDevicePointInfo",
					getPowerDevicePointInfo.RequestData{Id: strconv.FormatInt(int64(pni), 10)},
					DefaultCacheTimeout,
				)
				if sg.Error != nil {
					break
				}

				data := getPowerDevicePointInfo.Assert(ep)
				table = data.AddDataTable(table)
				if table.Error != nil {
					sg.Error = table.Error
					break
				}
			}
			fmt.Println("")
		} else {
			fmt.Printf("Searching for %v within getPowerDevicePointInfo\n", pns)
			for _, pn := range pns {
				ep := sg.GetByStruct(
					"AppService.getPowerDevicePointInfo",
					getPowerDevicePointInfo.RequestData{Id: pn},
					DefaultCacheTimeout,
				)
				if sg.Error != nil {
					break
				}

				data := getPowerDevicePointInfo.Assert(ep)
				table = data.GetDataTable()
				if table.Error != nil {
					sg.Error = table.Error
					break
				}
			}
			fmt.Println("")
		}

		table.SetTitle("Point Names")
		table.SetFilePrefix("AppService.getPowerDevicePointInfo")
		table.SetGraphFilter("")
		table.SetSaveFile(sg.SaveAsFile)
		table.OutputType = sg.OutputType
		sg.Error = table.Output()
		if sg.Error != nil {
			break
		}
	}

	return sg.Error
}

func (sg *SunGrow) GetDevices(psIds ...int64) error {
	for range Only.Once {
		if len(psIds) == 0 {
			psIds, sg.Error = sg.GetPsIds()
			if sg.Error != nil {
				break
			}
		}

		for _, psId := range psIds {
			ep := sg.GetByStruct(
				"AppService.getDeviceList",
				getDeviceList.RequestData{PsId: strconv.FormatInt(psId, 10)},
				DefaultCacheTimeout,
			)
			if sg.Error != nil {
				break
			}

			data := getDeviceList.Assert(ep)
			table := data.GetDataTable()
			if table.Error != nil {
				sg.Error = table.Error
				break
			}

			table.SetTitle("Device Info %s", psId)
			table.SetFilePrefix(data.SetFilenamePrefix("%d", psId))
			table.SetGraphFilter("")
			table.SetSaveFile(sg.SaveAsFile)
			table.OutputType = sg.OutputType
			sg.Error = table.Output()
			if sg.Error != nil {
				break
			}
		}
	}

	return sg.Error
}

func (sg *SunGrow) GetDeviceModels() error {
	for range Only.Once {
		ep := sg.GetByStruct(
			"AppService.getDeviceModelInfoList",
			getDeviceModelInfoList.RequestData{},
			DefaultCacheTimeout,
		)
		if sg.Error != nil {
			break
		}

		data := getDeviceModelInfoList.Assert(ep)
		table := data.GetDataTable()
		if table.Error != nil {
			sg.Error = table.Error
			break
		}

		table.SetTitle("Models")
		table.SetFilePrefix(data.SetFilenamePrefix(""))
		table.SetGraphFilter("")
		table.SetSaveFile(sg.SaveAsFile)
		table.OutputType = sg.OutputType
		sg.Error = table.Output()
		if sg.Error != nil {
			break
		}
	}

	return sg.Error
}

func (sg *SunGrow) GetDevicePoints(psIds ...int64) error {
	for range Only.Once {
		var points api.DataMap

		if len(psIds) == 0 {
			psIds, sg.Error = sg.GetPsIds()
			if sg.Error != nil {
				break
			}
		}

		// getPsList
		ep := sg.GetByStruct("AppService.getPsList", getPsList.RequestData{}, DefaultCacheTimeout)
		if sg.Error != nil {
			break
		}
		PsList := getPsList.Assert(ep)
		data := PsList.GetData()
		if PsList.Error != nil {
			sg.Error = PsList.Error
			break
		}
		points.AppendMap(data)

		for _, psId := range psIds {
			pid := strconv.Itoa(int(psId))

			// GoSungrow api raw getIncomeSettingInfos '{"ps_id":"1171348"}'
			// // getPowerStationForHousehold
			// // GoSungrow api raw getPowerStationForHousehold '{"ps_id":"1171348"}'


			// // getAreaList
			// getAreaList.Disabled
			//
			// // getPsReport
			// getPsReport.Disabled


			// // getPowerStatistics
			// // api raw getPowerStatistics '{"ps_id":"1171348"}'
			// ep = sg.GetByStruct("AppService.getPowerStatistics", getPowerStatistics.RequestData{ PsId: pid }, DefaultCacheTimeout)
			// if sg.Error != nil {
			// 	break
			// }
			// PowerStatistics := getPowerStatistics.Assert(ep)
			// data = PowerStatistics.GetData()
			// if PowerStatistics.Error != nil {
			// 	sg.Error = PowerStatistics.Error
			// 	break
			// }
			// points.AppendMap(data)


			// getPowerStationData
			// api raw getPowerStationData '{"date_id":"20221007","date_type":"1","ps_id":"1171348"}'
			ep = sg.GetByStruct("AppService.getPowerStationData", getPowerStationData.RequestData{ PsId: pid, DateType: "1", DateID: "20221007"}, DefaultCacheTimeout)
			if sg.Error != nil {
				break
			}
			PowerStationData := getPowerStationData.Assert(ep)
			data = PowerStationData.GetData()
			if PowerStationData.Error != nil {
				sg.Error = PowerStationData.Error
				break
			}
			points.AppendMap(data)
			// api raw getPowerStationData '{"date_id":"202210","date_type":"2","ps_id":"1171348"}'
			ep = sg.GetByStruct("AppService.getPowerStationData", getPowerStationData.RequestData{ PsId: pid, DateType: "2", DateID: "202210"}, DefaultCacheTimeout)
			if sg.Error != nil {
				break
			}
			PowerStationData = getPowerStationData.Assert(ep)
			data = PowerStationData.GetData()
			if PowerStationData.Error != nil {
				sg.Error = PowerStationData.Error
				break
			}
			points.AppendMap(data)
			// api raw getPowerStationData '{"date_id":"2022","date_type":"3","ps_id":"1171348"}'
			ep = sg.GetByStruct("AppService.getPowerStationData", getPowerStationData.RequestData{ PsId: pid, DateType: "3", DateID: "2022"}, DefaultCacheTimeout)
			if sg.Error != nil {
				break
			}
			PowerStationData = getPowerStationData.Assert(ep)
			data = PowerStationData.GetData()
			if PowerStationData.Error != nil {
				sg.Error = PowerStationData.Error
				break
			}
			points.AppendMap(data)


			// queryDeviceList
			// api get AppService.queryDeviceList '{"ps_id":"1171348"}'
			ep = sg.GetByStruct("AppService.queryDeviceList", queryDeviceList.RequestData{ PsId: pid }, DefaultCacheTimeout)
			if sg.Error != nil {
				break
			}
			DeviceList := queryDeviceList.Assert(ep)
			data = DeviceList.GetData()
			if DeviceList.Error != nil {
				sg.Error = DeviceList.Error
				break
			}
			points.AppendMap(data)


			// getHouseholdStoragePsReport
			// api get getHouseholdStoragePsReport '{"date_id":"20221001","date_type":"1","ps_id":"1129147"}'
			ep = sg.GetByStruct(
				"AppService.getHouseholdStoragePsReport",
				getHouseholdStoragePsReport.RequestData{ DateID: "20221001", DateType: "1", PsID: pid },
				DefaultCacheTimeout,
				)
			if sg.Error != nil {
				break
			}
			HouseholdStoragePsReport := getHouseholdStoragePsReport.Assert(ep)
			data = HouseholdStoragePsReport.GetData()
			if HouseholdStoragePsReport.Error != nil {
				sg.Error = HouseholdStoragePsReport.Error
				break
			}
			points.AppendMap(data)


			// getPsDetailWithPsType
			ep = sg.GetByStruct("AppService.getPsDetailWithPsType", getPsDetailWithPsType.RequestData{ PsId: pid }, DefaultCacheTimeout)
			if sg.Error != nil {
				break
			}
			PsDetailWithPsType := getPsDetailWithPsType.Assert(ep)
			data = PsDetailWithPsType.GetData()
			if PsDetailWithPsType.Error != nil {
				sg.Error = PsDetailWithPsType.Error
				break
			}
			points.AppendMap(data)


			// getPsDetail
			ep = sg.GetByStruct("AppService.getPsDetail", getPsDetail.RequestData{ PsId: pid }, DefaultCacheTimeout)
			if sg.Error != nil {
				break
			}
			PsDetail := getPsDetail.Assert(ep)
			data = PsDetail.GetData()
			if PsDetail.Error != nil {
				sg.Error = PsDetailWithPsType.Error
				break
			}
			points.AppendMap(data)


			// getKpiInfo
			ep = sg.GetByStruct("AppService.getKpiInfo", getKpiInfo.RequestData{ }, DefaultCacheTimeout)
			if sg.Error != nil {
				break
			}
			KpiInfo := getKpiInfo.Assert(ep)
			data = KpiInfo.GetData()
			if KpiInfo.Error != nil {
				sg.Error = KpiInfo.Error
				break
			}
			points.AppendMap(data)


			// // getPowerDevicePointInfo
			// ep = sg.GetByStruct("AppService.getPowerDevicePointInfo", getPowerDevicePointInfo.RequestData{ Id: pid }, DefaultCacheTimeout)
			// if sg.Error != nil {
			// 	break
			// }
			// ep4 := getPowerDevicePointInfo.Assert(ep)
			// data = ep4.GetData()
			// if ep4.Error != nil {
			// 	sg.Error = ep4.Error
			// 	break
			// }
			// points.AppendMap(data)


			// // queryDeviceRealTimeDataByPsKeys
			// ep = sg.GetByStruct("AppService.queryDeviceRealTimeDataByPsKeys", queryDeviceRealTimeDataByPsKeys.RequestData{ PsKeyList: pid }, DefaultCacheTimeout)
			// if sg.Error != nil {
			// 	break
			// }
			// DeviceRealTimeDataByPsKeys := queryDeviceRealTimeDataByPsKeys.Assert(ep)
			// data = DeviceRealTimeDataByPsKeys.GetData()
			// if DeviceRealTimeDataByPsKeys.Error != nil {
			// 	sg.Error = DeviceRealTimeDataByPsKeys.Error
			// 	break
			// }
			// points.AppendMap(data)

		}

		points.Print()
	}

	return sg.Error
}


func PrintPause(index int, max int) {
	for range Only.Once {
		if index == 0 {
			fmt.Printf("\n%.3d ", index)
			break
		}

		m := math.Mod(float64(index), float64(max))
		if m == 0 {
			fmt.Printf("PAUSE")
			time.Sleep(time.Millisecond * 500)
			// fmt.Printf("\r%s%.3d ", strings.Repeat(" ", 4), pni)
			fmt.Printf("\r%.3d ", index)
		} else {
			time.Sleep(time.Millisecond * 100)
			fmt.Printf(".")
		}
	}
}

func (sg *SunGrow) GetPointName(pn string) error {
	for range Only.Once {
		ep := sg.GetByStruct(
			"AppService.getPowerDevicePointInfo",
			getPowerDevicePointInfo.RequestData{Id: pn},
			DefaultCacheTimeout,
		)
		if sg.Error != nil {
			break
		}

		data := getPowerDevicePointInfo.Assert(ep)
		table := data.GetDataTable()
		if table.Error != nil {
			sg.Error = table.Error
			break
		}

		// table2 := data.GetData()
		// fmt.Printf("%v\n", table2)

		table.SetTitle("Point Name %s", pn)
		table.SetFilePrefix(data.SetFilenamePrefix(""))
		table.SetGraphFilter("")
		table.SetSaveFile(sg.SaveAsFile)
		table.OutputType = sg.OutputType
		sg.Error = table.Output()
		if sg.Error != nil {
			break
		}
	}

	return sg.Error
}


func (sg *SunGrow) GetPsIds() ([]int64, error) {
	var ret []int64

	for range Only.Once {
		ep := sg.GetByStruct("AppService.getPsList", nil, DefaultCacheTimeout)
		if ep.IsError() {
			sg.Error = ep.GetError()
			break
		}

		_getPsList := getPsList.AssertResultData(ep)
		ret = _getPsList.GetPsIds()
	}

	return ret, sg.Error
}

func (sg *SunGrow) GetPsId() (int64, error) {
	var ret int64

	for range Only.Once {

		ep := sg.GetByStruct("AppService.getPsList", nil, DefaultCacheTimeout)
		if ep.IsError() {
			sg.Error = ep.GetError()
			break
		}

		_getPsList := getPsList.AssertResultData(ep)
		psIds := _getPsList.GetPsIds()
		if len(psIds) == 0 {
			break
		}

		ret = psIds[0]
	}

	return ret, sg.Error
}

func (sg *SunGrow) StringToPids(pids ...string) ([]int64, error) {
	var psIds []int64
	for range Only.Once {
		for _, pid := range pids {
			p, err := strconv.ParseInt(pid, 10, 64)
			if err != nil {
				continue
			}
			psIds = append(psIds, p)
		}
		if len(psIds) == 0 {
			psIds, sg.Error = sg.GetPsIds()
			if sg.Error != nil {
				break
			}
		}
	}

	return psIds, sg.Error
}

func (sg *SunGrow) GetPsName() ([]string, error) {
	var ret []string

	for range Only.Once {
		ep := sg.GetByStruct("AppService.getPsList", nil, DefaultCacheTimeout)
		if ep.IsError() {
			sg.Error = ep.GetError()
			break
		}

		_getPsList := getPsList.AssertResultData(ep)
		ret = _getPsList.GetPsName()
	}

	return ret, sg.Error
}

func (sg *SunGrow) GetPsModel() ([]string, error) {
	var ret []string

	for range Only.Once {
		var psIds []int64
		psIds, sg.Error = sg.StringToPids()
		if sg.Error != nil {
			break
		}

		for _, psId := range psIds {
			ep := sg.GetByStruct(
				"AppService.getPsDetailWithPsType",
				getPsDetailWithPsType.RequestData{PsId: strconv.FormatInt(psId, 10)},
				DefaultCacheTimeout)
			if ep.IsError() {
				sg.Error = ep.GetError()
				break
			}

			data := getPsDetailWithPsType.Assert(ep)
			ret = append(ret, data.GetDeviceName())
		}
	}

	return ret, sg.Error
}

func (sg *SunGrow) GetPsSerial() ([]string, error) {
	var ret []string

	for range Only.Once {
		var psIds []int64
		psIds, sg.Error = sg.StringToPids()
		if sg.Error != nil {
			break
		}

		for _, psId := range psIds {
			ep := sg.GetByStruct(
				"AppService.getPsDetailWithPsType",
				getPsDetailWithPsType.RequestData{PsId: strconv.FormatInt(psId, 10)},
				DefaultCacheTimeout)
			if ep.IsError() {
				sg.Error = ep.GetError()
				break
			}

			data := getPsDetailWithPsType.Assert(ep)
			ret = append(ret, data.GetDeviceSerial())
		}
	}

	return ret, sg.Error
}

func (sg *SunGrow) GetPsKeys() ([]string, error) {
	var ret []string

	for range Only.Once {
		var psIds []int64
		psIds, sg.Error = sg.StringToPids()
		if sg.Error != nil {
			break
		}

		for _, psId := range psIds {
			ep := sg.GetByStruct(
				"AppService.getPsDetailWithPsType",
				getPsDetailWithPsType.RequestData{PsId: strconv.FormatInt(psId, 10)},
				DefaultCacheTimeout)
			if ep.IsError() {
				sg.Error = ep.GetError()
				break
			}

			data := getPsDetailWithPsType.Assert(ep)
			ret = append(ret, data.GetPsKeys()...)
		}
	}

	return ret, sg.Error
}
