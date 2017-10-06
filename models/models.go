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
	StartedAt		*time.Time // Nullable when not started
	EndedAt			*time.Time // Nullable
	User			User
	UserID			uint
	RTMPName		string
	Password		string
}

type BroadcastApproval struct {
	gorm.Model
	Broadcast		Broadcast
	BroadcastID		uint
	User			User
	UserID			uint
}
