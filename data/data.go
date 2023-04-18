package data

import "time"

const DBName = "test"

const (
	ExternalTable  = "external"
	UserPowerTable = "user_power"
)

type External struct {
	Id          int       `gorm:"column:id;primary_key" json:"id"`
	ServiceName string    `gorm:"column:service_name" json:"service_name"`
	Method      string    `gorm:"column:method" json:"method"`
	Verify      int       `gorm:"column:verify" json:"verify"`
	Power       int32     `gorm:"column:power" json:"power"`
	Status      int       `gorm:"column:status" json:"status"`
	CreateTime  time.Time `gorm:"column:create_time" json:"create_time"`
	UpdateTime  time.Time `gorm:"column:update_time" json:"update_time"`
}

type UserPower struct {
	Id         int       `gorm:"column:id;primary_key" json:"id"`
	UserID     string    `gorm:"column:user_id" json:"user_id"`
	Path       string    `gorm:"column:path" json:"path"`
	Power      int32     `gorm:"column:power" json:"power"`
	Status     int       `gorm:"column:status" json:"status"`
	CreateTime time.Time `gorm:"column:create_time" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time" json:"update_time"`
}
