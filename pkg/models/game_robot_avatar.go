package models

import (
	"errors"
	"github.com/jmoiron/sqlx"
)

// GameRobotAvatar 机器人头像url
type GameRobotAvatar struct {
	ID          int64  `json:"id" form:"id" db:"id"`
	ImgUrl 		string `json:"img_url" form:"img_url" db:"img_url"`
}

// GetAllRobotAvatar 得到所有机器人得url
func GetAllRobotAvatar(robotdb *sqlx.DB) (gameRobotAvatar []GameRobotAvatar, err error){
	if robotdb == nil {
		return gameRobotAvatar, errors.New("robotdb is null")
	}
	tpl := `select id,img_url from game_robot_avatar`
	if err := robotdb.Select(&gameRobotAvatar, tpl); err != nil {
		return gameRobotAvatar, err
	}
	return gameRobotAvatar, nil
}

