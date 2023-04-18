package data

import (
	"github.com/liuyp5181/base/database"
	"time"
)

func InsertUserPower(userID, path string, power int32, status int) (int, error) {
	up := UserPower{
		UserID:     userID,
		Path:       path,
		Power:      power,
		Status:     status,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	db := database.GetMysql(DBName)
	err := db.Table(UserPowerTable).Create(&up).Error
	if err != nil {
		return 0, err
	}

	return up.Id, nil
}

func UpdateUserPower(userID, path string, power int32, status int) error {
	up := UserPower{
		UserID:     userID,
		Path:       path,
		Power:      power,
		Status:     status,
		UpdateTime: time.Now(),
	}

	db := database.GetMysql(DBName)
	err := db.Table(UserPowerTable).Omit("create_time").Updates(&up).Error
	if err != nil {
		return err
	}
	return nil
}

func QueryUserPower(userID string) (*UserPower, error) {
	var up UserPower
	db := database.GetMysql(DBName)
	err := db.Table(UserPowerTable).Where("user_id = ?", userID).Find(&up).Error
	if err != nil {
		return nil, err
	}

	return &up, nil
}

func DeleteUserPower(userID string) error {
	db := database.GetMysql(DBName)
	err := db.Table(UserPowerTable).Delete("user_id = ?", userID).Error
	if err != nil {
		return err
	}
	return nil
}
