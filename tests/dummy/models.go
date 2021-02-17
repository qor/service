package dummy

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/media/oss"
)

type CreditCard struct {
	gorm.Model
	Number string
	Issuer string
}

type Company struct {
	gorm.Model
	Name string
}

type Address struct {
	gorm.Model
	UserID   uint
	Address1 string
	Address2 string
}

type Language struct {
	gorm.Model
	Name string
}

type User struct {
	gorm.Model
	Name         string `gorm:"size:50"`
	Age          uint
	Role         string
	Active       bool
	RegisteredAt *time.Time
	Avatar       oss.OSS
	Profile      Profile // has one
	CreditCardID uint
	CreditCard   CreditCard // belongs to
	Addresses    []Address  // has many
	CompanyID    uint
	Company      *Company   // belongs to
	Languages    []Language `gorm:"many2many:user_languages;"` // many 2 many
}

type Profile struct {
	gorm.Model
	UserID uint
	Name   string
	Sex    string

	Phone Phone
}

type Phone struct {
	gorm.Model

	ProfileID uint64
	Num       string
}

func (u User) DisplayName() string {
	return u.Name
}

func (u User) GetID() uint {
	return u.ID
}

func (u User) GetUsersByIDs(db *gorm.DB, ids []string) interface{} {
	return u
}
