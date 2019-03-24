package api

import "time"

// DBModel is the base model for all db items
type DBModel struct {
	ID        uint       `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-"`
}

// User is the struct that holds user specific information
type User struct {
	DBModel
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
	Hash  string `json:"-"`
}

// Schedule is the struct that holds the schedule information
type Schedule struct {
	DBModel
	Time   time.Time `json:"time"`
	Source string    `json:"source,omitempty"`
	Status string    `json:"status"`
}

// MigrateDB creates all necessary database relations
func (routes *Routes) MigrateDB() {
	routes.db.AutoMigrate(&User{}, &Schedule{})
}
