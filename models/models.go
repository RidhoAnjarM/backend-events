package models

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `gorm:"unique" json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type Event struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	DateStart         string    `json:"datestart"`
	DateEnd           string    `json:"dateend"`
	Time              string    `json:"time"`
	LocationID        uint      `json:"location_id"`
	Location          string    `json:"location"`
	Address           string    `json:"address"`
	Capacity          int       `json:"capacity"`
	RemainingCapacity int       `json:"remaining_capacity"`
	Photo             string    `json:"photo"`
	Price             string    `json:"price"`
	CategoryID        uint      `json:"category_id"`
	Benefits          string    `json:"benefits"`
	Mode              string    `json:"mode"`
	Link              string    `json:"link"`
	Status            string    `json:"status"`
	Sessions          []Session `gorm:"foreignKey:EventID" json:"sessions"`
	PopularityScore   float64   `json:"popularity_score"`
}

type Session struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	EventID  uint   `json:"event_id"`
	Date     string `json:"date"`
	Time     string `json:"time,omitempty"`
	Speaker  string `json:"speaker,omitempty"`
	Location string `json:"location,omitempty"`
}

type Registration struct {
	ID            uint   `gorm:"primaryKey"`
	UserID        uint   `json:"user_id"`
	User          User   `gorm:"foreignKey:UserID"`
	EventID       uint   `json:"event_id"`
	Event         Event  `gorm:"foreignKey:EventID"`
	Username      string `json:"username" `
	Name          string `json:"name"`
	Email         string `json:"email"`
	PhoneNumber   string `json:"phone"`
	Job           string `json:"job"`
	PaymentMethod string `json:"payment_method"`
	PaymentStatus string `json:"payment_status"`
}

type Category struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `json:"name"`
}

type Location struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	City string `gorm:"not null" json:"city"`
}

type Rating struct {
	ID      uint `gorm:"primaryKey" json:"id"`
	UserID  uint `gorm:"not null" json:"user_id"`
	EventID uint `gorm:"not null" json:"event_id"`
	Rating  int  `gorm:"not null" json:"rating"`
}
