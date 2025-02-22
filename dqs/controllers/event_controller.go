package controllers

import (
	"dqs/dao"
	"dqs/models"
	"dqs/util"
	"encoding/xml"
	//"github.com/astaxie/beego"
	//"bytes"
	"fmt"
	log "github.com/cihub/seelog"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	EarthQuakeTimeLayout = "20060102150405.000" //地震时间格式
)

type EventController struct {
	BaseController
}

var Local *time.Location

func init() {
	l, err := time.LoadLocation("Local")
	if err == nil {
		Local = l
		fmt.Println(Local)
	} else {
		fmt.Println("init:" + err.Error())
	}
}

//事件列表
func (this *EventController) EventPageList() {
	this.Data["title"] = "仪器观测事件"
	this.Data["author"] = "wangzhigang"
	//权限检查
	//this.AuthRoles("role_admin")
	this.CheckUser()

	pagination := util.Pagination{}
	page, err := this.GetInt("page")
	if err != nil {
		pagination.CurrentPage = 1
	} else {
		pagination.CurrentPage = int(page)
	}
	pagesize, err2 := this.GetInt("pagesize")
	if err2 != nil {
		pagination.PageSize = 10
	} else {
		pagination.PageSize = int(pagesize)
	}

	//查询参数

	eventid := this.GetString("eventid")
	if eventid != "" {
		pagination.AddParams("eventid", eventid)
	}
	begintime := this.GetString("begintime")
	if begintime != "" {
		pagination.AddParams("begintime", begintime)
	} else {
		now := time.Now()
		pagination.AddParams("begintime", now.Format(dao.EventTimeLayout))
	}
	endtime := this.GetString("endtime")
	if endtime != "" {
		pagination.AddParams("endtime", endtime)
	}

	//执行查询
	err = dao.EventPageList(&pagination)
	if err != nil {
		log.Warnf("查询震情事件列表失败:%s", err.Error())
	}
	pagination.Compute()

	this.Data["pagedata"] = pagination
	this.TplNames = "eventlist.html"
	this.Render()
}

//事件列表Json
func (this *EventController) EventJsonList() {

	pagination := util.Pagination{}
	page, err := this.GetInt("page")
	if err != nil {
		pagination.CurrentPage = 1
	} else {
		pagination.CurrentPage = int(page)
	}
	pagesize, err2 := this.GetInt("pagesize")
	if err2 != nil {
		pagination.PageSize = 10
	} else {
		pagination.PageSize = int(pagesize)
	}

	//查询参数
	/*
		eventid := this.GetString("eventid")
		if eventid != "" {
			pagination.AddParams("eventid", eventid)
		}
		begintime := this.GetString("begintime")
		if begintime != "" {
			pagination.AddParams("begintime", begintime)
		} else {
			now := time.Now()
			pagination.AddParams("begintime", now.Format(dao.EventTimeLayout))
		}
		endtime := this.GetString("endtime")
		if endtime != "" {
			pagination.AddParams("endtime", endtime)
		}
	*/
	//执行查询
	err = dao.EventPageList(&pagination)
	if err != nil {
		log.Warnf("查询震情事件列表失败:%s", err.Error())
	}
	pagination.Compute()

	this.Data["json"] = pagination.Data
	this.ServeJson()
}

//确认事件列表
func (this *EventController) EventSignalPageList() {
	this.Data["title"] = "地震事件列表"
	this.Data["author"] = "wangzhigang"
	//权限检查
	//this.AuthRoles("role_admin")
	this.CheckUser()

	pagination := util.Pagination{}
	page, err := this.GetInt("page")
	if err != nil {
		pagination.CurrentPage = 1
	} else {
		pagination.CurrentPage = int(page)
	}
	pagesize, err2 := this.GetInt("pagesize")
	if err2 != nil {
		pagination.PageSize = 10
	} else {
		pagination.PageSize = int(pagesize)
	}

	//查询参数

	signalid := this.GetString("signalid")
	if signalid != "" {
		pagination.AddParams("signalid", signalid)
	}
	level, err3 := this.GetInt("level")
	if err3 == nil {
		pagination.AddParams("level", int(level))
	}

	begintime := this.GetString("begintime")
	if begintime != "" {
		pagination.AddParams("begintime", begintime)
	} else {
		now := time.Now()
		pagination.AddParams("begintime", now.Format(dao.EventTimeLayout))
	}
	endtime := this.GetString("endtime")
	if endtime != "" {
		pagination.AddParams("endtime", endtime)
	}

	//执行查询
	err = dao.EventSignalPageList(&pagination)
	if err != nil {
		log.Warnf("查询地震事件列表失败:%s", err.Error())
	}
	pagination.Compute()

	this.Data["pagedata"] = pagination
	this.TplNames = "eventsignallist.html"
	this.Render()
}

//根据地震事件构造等值线
func (this *EventController) EventLine() {
	this.Data["title"] = "等值线"
	this.Data["author"] = "wangzhigang"
	//权限检查
	//this.AuthRoles("role_admin")
	this.CheckUser()

	eventid := this.GetString(":id")
	var eventSignal models.EventSignal = models.EventSignal{}
	//var dataArray []models.NetGrid
	//查找当前事件
	event, err0 := dao.GetEventById(eventid)
	if err0 != nil {
		log.Warnf("查找事件[%s]失败:%s", eventid, err0.Error())
	}
	if event.IsConfirm {
		eventSignal, err0 = dao.GetEventSignalById(event.SignalId)
	}

	//查找报警数据
	alarms, err := dao.GetAlarmsByEventId(eventid)
	if err != nil {
		log.Warnf("查找等值线的报警数据时出错:%s", err.Error())
	}

	//是否加入网格化虚拟站点
	dataArray := NetGridCompute(alarms, eventSignal)

	//传递的数据值
	DataArrayStr := ""
	DataArrayStrPGA := ""
	DataArrayStrSI := ""
	var lastlng, lastlat float32

	for k, v := range dataArray {
		if k < len(dataArray)-1 {
			DataArrayStr += v.String() + ","
			DataArrayStrPGA += v.StringPGA() + ","
			DataArrayStrSI += v.StringSI() + ","
		} else {
			DataArrayStr += v.String()
			DataArrayStrPGA += v.StringPGA()
			DataArrayStrSI += v.StringSI()
		}
		//设置地图中心点位置
		if k == 0 {
			lastlng = v.Longitude
			lastlat = v.Latitude
		}
	}
	//添加系统参数
	this.Data["dataArray"] = DataArrayStr
	this.Data["dataArrayPGA"] = DataArrayStrPGA
	this.Data["dataArraySI"] = DataArrayStrSI

	this.Data["dataSize"] = len(dataArray)

	this.Data["lastlng"] = lastlng
	this.Data["lastlat"] = lastlat

	usegis := SystemConfigs.UseGis
	if usegis {
		this.Data["gisServiceUrl"] = SystemConfigs.GisServiceUrl
		this.Data["gisServiceParams"] = SystemConfigs.GisServiceParams
		this.Data["gisBasicLayer"] = SystemConfigs.GisLayerBasic
		this.Data["gisChinaLayer"] = SystemConfigs.GisLayerChina
		this.TplNames = "eventline.html"
	} else {
		this.TplNames = "eventline-nogis.html"
	}
	this.Render()

}

//根据地震事件构造等值线
func (this *EventController) EventLineJson() {

	eventid := this.GetString(":id")
	var eventSignal models.EventSignal = models.EventSignal{}
	//查找当前事件
	event, err0 := dao.GetEventById(eventid)
	if err0 != nil {
		log.Warnf("查找事件[%s]失败:%s", eventid, err0.Error())
	}
	if event.IsConfirm {
		eventSignal, err0 = dao.GetEventSignalById(event.SignalId)
	}

	//查找报警数据
	alarms, err := dao.GetAlarmsByEventId(eventid)
	if err != nil {
		log.Warnf("查找等值线的报警数据时出错:%s", err.Error())
	}

	//是否加入网格化虚拟站点
	dataArray := NetGridCompute(alarms, eventSignal)

	//传递的数据值
	data := make(map[string]interface{})
	DataArrayStr := ""
	DataArrayStrPGA := ""
	DataArrayStrSI := ""
	var lastlng, lastlat float32

	for k, v := range dataArray {
		if k < len(dataArray)-1 {
			DataArrayStr += v.String() + ","
			DataArrayStrPGA += v.StringPGA() + ","
			DataArrayStrSI += v.StringSI() + ","
		} else {
			DataArrayStr += v.String()
			DataArrayStrPGA += v.StringPGA()
			DataArrayStrSI += v.StringSI()
		}
		//设置地图中心点位置
		if k == 0 {
			lastlng = v.Longitude
			lastlat = v.Latitude
		}
	}
	//添加系统参数
	data["dataArray"] = DataArrayStr
	data["dataArrayPGA"] = DataArrayStrPGA
	data["dataArraySI"] = DataArrayStrSI

	data["dataSize"] = len(dataArray)

	data["lastlng"] = lastlng
	data["lastlat"] = lastlat

	this.Data["json"] = data
	this.ServeJson()

}

//地震事件定位
func (this *EventController) QuakeLocation() {
	this.Data["title"] = "地震定位"
	this.Data["author"] = "wangzhigang"
	this.CheckUser()
	sid := this.GetString(":id")
	signal, err := dao.GetEventSignalById(sid)
	if err != nil {
		log.Warnf("获取地震数据失败:%s", err.Error())
	}
	this.Data["signal"] = signal

	usegis := SystemConfigs.UseGis
	if usegis {
		this.Data["gisServiceUrl"] = SystemConfigs.GisServiceUrl
		this.Data["gisServiceParams"] = SystemConfigs.GisServiceParams
		this.Data["gisBasicLayer"] = SystemConfigs.GisLayerBasic
		this.Data["gisChinaLayer"] = SystemConfigs.GisLayerChina
		this.TplNames = "quakelocation.html"
	} else {
		this.TplNames = "quakelocation-nogis.html"
	}

	this.Render()
}

//网格化计算数据点
func NetGridCompute(alarms *[]models.AlarmInfo, eventSignal models.EventSignal) (ng []models.NetGrid) {

	dataArray := make([]models.NetGrid, len(*alarms), len(*alarms)*2)
	for k, v := range *alarms {
		dataArray[k] = models.NetGrid{Longitude: v.Longitude, Latitude: v.Latitude, Value: v.Intensity, PGAValue: v.PGA, SIValue: v.SI}
	}
	//计算虚拟的网格值
	if eventSignal.Id != "" {
		//TODO待添加算法,转移到GIS服务中进行
	}
	return dataArray
}

//添加地震事件
func (this *EventController) AddEventSignal() {
	body := this.Ctx.Request.Body
	defer body.Close()

	content, err := ioutil.ReadAll(body)
	if err != nil {
		log.Errorf("地震事件接口读入数据错误:%s", err.Error())
		this.writeResponse(false, err.Error())
		return
	}
	//xml解析
	earthQuake := new(models.EarthQuake)
	if err = xml.Unmarshal(content, earthQuake); err != nil {
		log.Errorf("地震事件接口xml解析错误:%s", err.Error())
		this.writeResponse(false, err.Error())
		return
	}
	eventSignal := new(models.EventSignal)

	eventSignal.Id = util.GUID()
	eventSignal.Longitude = earthQuake.Longitude
	eventSignal.Latitude = earthQuake.Latitude
	eventSignal.Level = earthQuake.Level
	eventSignal.EventId = earthQuake.EVENT_ID
	eventSignal.EventSeq = earthQuake.EventSeq
	eventSignal.CODE = earthQuake.CODE
	eventSignal.CNAME = earthQuake.CNAME
	eventSignal.DEPTH = earthQuake.DEPTH
	eventSignal.LOCATION_CNAME = earthQuake.LOCATION_CNAME
	tm, errt := time.ParseInLocation(EarthQuakeTimeLayout, earthQuake.Time, Local)
	if errt != nil {
		eventSignal.Time = time.Now()
	}
	eventSignal.Time = tm
	eventSignal.ReceiveTime = time.Now()

	err = dao.EventSignalSave(eventSignal)
	if err != nil {
		log.Errorf("保存接收的地震事件失败:%s", err.Error())
		this.writeResponse(false, err.Error())
		return
	}

	log.Infof("成功接收了地震事件%s [%f,%f] %f级", earthQuake.Time, earthQuake.Longitude, earthQuake.Latitude, earthQuake.Level)
	this.writeResponse(true, "success")

	//提供回送数据
	go FeedbackData(eventSignal)
	return
}

//写入回复数据
func (this *EventController) writeResponse(ok bool, msg string) {
	reswriter := this.Ctx.ResponseWriter
	fb := models.Feedback{}
	fb.Ok = ok
	fb.Message = msg
	cont, err := xml.Marshal(fb)
	if err == nil {
		reswriter.Write(cont)
	}
}

//回送数据，调用川局服务器接口
func FeedbackData(eventSignal *models.EventSignal) {

	var sleepTime = SystemConfigs.QuakeReportCfg.SleepTime
	var timeSpan = SystemConfigs.QuakeReportCfg.TimeSpan
	var serviceUrl = strings.TrimSpace(SystemConfigs.QuakeReportCfg.ServiceURL)
	if serviceUrl == "" {
		return
	}
	if sleepTime < 0 {
		sleepTime = 5
	}

	if timeSpan == 0 {
		timeSpan = 5
	}
	if time.Now().Sub(eventSignal.Time) < time.Minute*time.Duration(sleepTime) {
		log.Infof("当前时刻与事件时间较近,延时%d分钟报送数据", sleepTime)
		time.Sleep(time.Minute * time.Duration(sleepTime))
	}
	alarmList, err1 := dao.FetchQuakeAlarms(eventSignal.Time, SystemConfigs.QuakeReportCfg.TimeSpan)
	if err1 != nil || alarmList == nil {
		log.Error("获取报警数据失败：停止发送")
		return
	}

	var quakeDataList models.AlarmDataList = models.AlarmDataList{}
	quakeDataList.EventId = eventSignal.EventId
	quakeDataList.Alarms = alarmList

	dataBody, err := xml.Marshal(quakeDataList)
	if err != nil {
		log.Errorf("数据解析错误：%s", err.Error())
		return
	}
	encodeData := url.QueryEscape(string(dataBody))
	/*if strings.TrimSpace(string(dataBody)) == "" {
		log.Warn("空数据，停止报送")
		return
	}
	*/

	datastr := `<?xml version="1.0" encoding="utf-8"?>
  <soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
   <soap:Body>
    <LDdatainput xmlns="http://tempuri.org/">
      <dzid>%s</dzid>
      <dzxl>%s</dzxl>
      <xml>%s</xml>
    </LDdatainput>
   </soap:Body>
  </soap:Envelope>`

	dzxh := eventSignal.EventSeq
	if dzxh == "" {
		dzxh = "001"
	}
	sendStr := fmt.Sprintf(datastr, eventSignal.EventId, dzxh, encodeData)
	log.Info(sendStr)
	client := &http.Client{}
	reqest, _ := http.NewRequest("POST", serviceUrl, strings.NewReader(sendStr))

	reqest.Header.Add("Content-Type", "text/xml; charset=utf-8")
	reqest.Header.Add("SOAPAction", "http://tempuri.org/LDdatainput")

	response, _ := client.Do(reqest)
	if response.StatusCode == 200 {
		body, _ := ioutil.ReadAll(response.Body)

		result := models.ResultInfo{}
		err := xml.Unmarshal(body, &result)
		if err == nil {
			if result.Body.LDdatainputResponse.LDdatainputResult == "0" {
				log.Info("回送数据接口调用成功")
			}
		}
		bodystr := string(body)
		log.Info(bodystr)
	} else {
		log.Error(response.StatusCode)
		body, _ := ioutil.ReadAll(response.Body)
		result := models.ResultInfo{}
		err := xml.Unmarshal(body, &result)
		if err == nil {
			log.Error(result.Body.Fault.Reason)
		}
		bodystr := string(body)
		log.Error(bodystr)
	}
}
