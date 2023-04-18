package data

import (
	"github.com/liuyp5181/base/database"
	"time"
)

func InsertExternal(serviceName, method string, power int32, status int) (int, error) {
	ex := External{
		ServiceName: serviceName,
		Method:      method,
		Power:       power,
		Status:      status,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}
	db := database.GetMysql(DBName)
	err := db.Table(ExternalTable).Create(&ex).Error
	if err != nil {
		return 0, err
	}

	return ex.Id, nil
}

func UpdateExternal(id int, serviceName, method string, power int32, status int) error {
	ex := External{
		Id:          id,
		ServiceName: serviceName,
		Method:      method,
		Power:       power,
		Status:      status,
		UpdateTime:  time.Now(),
	}

	db := database.GetMysql(DBName)
	err := db.Table(ExternalTable).Omit("create_time").Updates(&ex).Error
	if err != nil {
		return err
	}
	return nil
}

func QueryExternal(id int) (*External, error) {
	var ex External
	db := database.GetMysql(DBName)
	err := db.Table(ExternalTable).Where("id = ?", id).Find(&ex).Error
	if err != nil {
		return nil, err
	}

	return &ex, nil
}

func QueryExternalList(serviceName ...string) ([]*External, error) {
	var exs []*External
	db := database.GetMysql(DBName)
	if len(serviceName) > 0 {
		db = db.Where("service_name in ?", serviceName)
	}
	err := db.Table(ExternalTable).Find(&exs).Error
	if err != nil {
		return nil, err
	}

	return exs, nil
}

func DeleteExternal(id int) error {
	db := database.GetMysql(DBName)
	err := db.Table(ExternalTable).Delete("id = ?", id).Error
	if err != nil {
		return err
	}
	return nil
}
