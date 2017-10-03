package models

import (
	"time"
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	ScreenName		string
	Name			string
	TwitterID		int64
	LastLoginedAt	time.Time
}

type Broadcast struct {
	gorm.Model
	StartedAt		time.Time
	EndedAt			time.Time
	User			User
	RTMPURL			string
	PublishURL		string
	Password		string
}
