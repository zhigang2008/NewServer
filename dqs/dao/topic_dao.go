package dao

import (
	"dqs/models"
	"labix.org/v2/mgo/bson"
	"time"
)

const (
	TopicTimeLayout = "2006-01-02 15:04:05"
)

//分页查询
func GetTopicData(beginTime string, endTime string, sensors []string) (*[]models.AlarmInfo, error) {
	c := GetSession().DB(DatabaseName).C(DataCollection)

	//构造查询参数
	m := bson.M{}

	timeparam := bson.M{}

	btime, _ := time.ParseInLocation(TopicTimeLayout, beginTime, Local)
	timeparam["$gte"] = btime

	etime, _ := time.ParseInLocation(TopicTimeLayout, endTime, Local)
	timeparam["$lte"] = etime

	m["initrealtime"] = timeparam
	m["sensorid"] = bson.M{"$in": sensors}

	//查询
	alist := []models.AlarmInfo{}
	err0 := c.Find(&m).All(&alist)
	if err0 != nil {
		return nil, err0
	}

	return &alist, nil
}
