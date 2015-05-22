package controllers

import (
	"dqs/dao"
	"dqs/util"
	log "github.com/cihub/seelog"
)

type TopicController struct {
	BaseController
}

func (this *TopicController) Get() {
	this.Data["title"] = "专题图"
	this.Data["author"] = "wangzhigang"

	this.CheckUser()

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
