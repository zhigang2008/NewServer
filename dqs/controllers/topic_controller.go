package controllers

import (
	"dqs/dao"
	"dqs/models"
	"dqs/util"
	log "github.com/cihub/seelog"
	"time"
)

type TopicController struct {
	BaseController
}

func (this *TopicController) TopicView() {
	this.Data["title"] = "专题图"
	this.Data["author"] = "wangzhigang"

	this.CheckUser()

	//当前时间
	now := time.Now()
	CurrentTime := now.Format(CommonTimeLayout)
	this.Data["CurrentTime"] = CurrentTime

	paginationDevices := util.Pagination{PageSize: 10, CurrentPage: 1}
	err := dao.DeviceList(&paginationDevices)
	if err != nil {
		log.Warnf("查询设备信息失败:%s", err.Error())
	}
	paginationDevices.Compute()
	this.Data["devices"] = paginationDevices.Data
	this.Data["devicePages"] = paginationDevices.PageCount

	allDevices, err := dao.GetAllDevices()
	if err != nil {
		log.Warnf("查询所有设备信息失败:%s", err.Error())
	}
	this.Data["allDevices"] = allDevices

	this.Data["GoogleMap"] = !SystemConfigs.DisableGoogleMap
	this.Data["mapExtend"] = SystemConfigs.GisImageCfg.BBOX

	this.Data["gisServiceUrl"] = SystemConfigs.GisServiceUrl
	this.Data["gisServiceParams"] = SystemConfigs.GisServiceParams
	this.Data["gisBasicLayer"] = SystemConfigs.GisLayerBasic
	this.Data["gisChinaLayer"] = SystemConfigs.GisLayerChina

	this.TplNames = "topic.html"
	this.Render()
}

//查询关联数据
func (this *TopicController) TopicData() {
	beginTime := this.GetString("beginTime")
	endTime := this.GetString("endTime")
	sensors := this.GetStrings("Sensors")

	var eventSignal models.EventSignal = models.EventSignal{}
	//查找报警数据
	alarms, err := dao.GetTopicData(beginTime, endTime, sensors)
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
}
