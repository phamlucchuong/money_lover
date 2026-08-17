package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string     `gorm:"type:varchar(255);not null"`
	Email     string     `gorm:"type:varchar(255);not null"`
	Password  string     `gorm:"type:varchar(255);not null"`
	CreatedAt *time.Time `gorm:"autoCreateTime"`
	UpdatedAt *time.Time `gorm:"autoUpdateTime"`
	DeletedAt *time.Time `gorm:"type:timestampz;default:null"`
}
