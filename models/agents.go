package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/jinzhu/gorm"
)

// AgentOS is model to contain agent OS information
type AgentOS struct {
	Type string `form:"type" json:"type" validate:"oneof=windows linux darwin,required"`
	Arch string `form:"arch" json:"arch" validate:"oneof=386 amd64,required"`
	Name string `form:"name" json:"name" validate:"oneof=unknown,required"`
}

// Valid is function to control input/output data
func (aos AgentOS) Valid() error {
	return validate.Struct(aos)
}

// AgentUser is model to contain agent User information
type AgentUser struct {
	Name  string `form:"name" json:"name" validate:"oneof=unknown,required"`
	Group string `form:"group" json:"group" validate:"oneof=unknown,required"`
}

// Valid is function to control input/output data
func (au AgentUser) Valid() error {
	return validate.Struct(au)
}

// AgentInfo is model to contain general agent information
type AgentInfo struct {
	OS   AgentOS   `form:"os" json:"os" validate:"required"`
	User AgentUser `form:"user" json:"user" validate:"required"`
}

// Valid is function to control input/output data
func (ai AgentInfo) Valid() error {
	return validate.Struct(ai)
}

// Value is interface function to return current value to store to DB
func (ai AgentInfo) Value() (driver.Value, error) {
	b, err := json.Marshal(ai)
	return string(b), err
}

// Scan is interface function to parse DB value when getting from DB
func (ai *AgentInfo) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), ai)
}

// Agent is model to contain agent information from instance DB
type Agent struct {
	ID            uint64    `form:"id,omitempty" json:"id,omitempty" validate:"min=0,numeric" gorm:"type:INT(10) UNSIGNED;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	Hash          string    `form:"hash" json:"hash" validate:"len=32,hexadecimal,lowercase,required" gorm:"type:VARCHAR(32);NOT NULL"`
	IP            string    `form:"ip" json:"ip" validate:"max=50,tcp_addr,required" gorm:"type:VARCHAR(50);NOT NULL"`
	Description   string    `form:"description" json:"description" validate:"max=255,required" gorm:"type:VARCHAR(255);NOT NULL"`
	Info          AgentInfo `form:"info" json:"info" validate:"required" gorm:"type:JSON;NOT NULL"`
	Status        string    `form:"status" json:"status" validate:"oneof=created connected disconnected removed,required" gorm:"type:ENUM('created','connected','disconnected','removed');NOT NULL"`
	ConnectedDate time.Time `form:"connected_date,omitempty" json:"connected_date,omitempty" validate:"omitempty" gorm:"type:DATETIME;DEFAULT:NULL"`
	CreatedDate   time.Time `form:"created_date,omitempty" json:"created_date,omitempty" validate:"omitempty" gorm:"type:DATETIME;NOT NULL;DEFAULT:CURRENT_TIMESTAMP"`
}

// TableName returns the table name string to guaranty use correct table
func (a *Agent) TableName() string {
	return "agents"
}

// BeforeDelete hook defined for cascade delete
func (a *Agent) BeforeDelete(db *gorm.DB) error {
	db = db.Model(&Event{}).Where("agent_id = ?", a.ID).Unscoped().Delete(&Event{})
	db = db.Model(&ModuleA{}).Where("agent_id = ?", a.ID).Unscoped().Delete(&ModuleA{})
	return db.Error
}

// Valid is function to control input/output data
func (a Agent) Valid() error {
	return validate.Struct(a)
}

// Validate is function to use callback to control input/output data
func (a Agent) Validate(db *gorm.DB) {
	if err := a.Valid(); err != nil {
		db.AddError(err)
	}
}
