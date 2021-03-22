package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/jinzhu/gorm"
)

// EventInfo is model to contain event information
type EventInfo struct {
	Data map[string]interface{} `form:"data" json:"data" validate:"required,dive,keys,solid,endkeys,required,valid"`
	Name string                 `form:"name" json:"name" validate:"max=1024,solid,required"`
	Time uint64                 `form:"time" json:"time" validate:"min=0,omitempty"`
	Uniq string                 `form:"uniq" json:"uniq" validate:"max=1024,required"`
}

// Valid is function to control input/output data
func (ei EventInfo) Valid() error {
	return validate.Struct(ei)
}

// Value is interface function to return current value to store to DB
func (ei EventInfo) Value() (driver.Value, error) {
	b, err := json.Marshal(ei)
	return string(b), err
}

// Scan is interface function to parse DB value when getting from DB
func (ei *EventInfo) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), ei)
}

// Event is model to contain event data from instance DB
type Event struct {
	ID       uint64    `form:"id,omitempty" json:"id,omitempty" validate:"min=0,numeric" gorm:"type:INT(10) UNSIGNED;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	ModuleID uint64    `form:"module_id" json:"module_id" validate:"min=0,numeric" gorm:"type:INT(10) UNSIGNED;NOT NULL"`
	AgentID  uint64    `form:"agent_id" json:"agent_id" validate:"min=0,numeric" gorm:"type:INT(10) UNSIGNED;NOT NULL"`
	Info     EventInfo `form:"info" json:"info" validate:"required" gorm:"type:JSON;NOT NULL"`
	Date     time.Time `form:"date,omitempty" json:"date,omitempty" validate:"omitempty" gorm:"type:DATETIME;NOT NULL;DEFAULT:CURRENT_TIMESTAMP"`
}

// TableName returns the table name string to guaranty use correct table
func (e *Event) TableName() string {
	return "events"
}

// Valid is function to control input/output data
func (e Event) Valid() error {
	return validate.Struct(e)
}

// Validate is function to use callback to control input/output data
func (e Event) Validate(db *gorm.DB) {
	if err := e.Valid(); err != nil {
		db.AddError(err)
	}
}

// EventModuleA is model to contain event data linked with module agent
type EventModuleA struct {
	ModuleA ModuleA `form:"module,omitempty" json:"module,omitempty" gorm:"FOREIGNKEY:id;ASSOCIATION_FOREIGNKEY:module_id;ASSOCIATION_AUTOUPDATE:false;ASSOCIATION_AUTOCREATE:false"`
	Event   `form:"" json:""`
}

// Valid is function to control input/output data
func (ema EventModuleA) Valid() error {
	if err := ema.ModuleA.Valid(); err != nil {
		return err
	}
	return ema.Event.Valid()
}

// Validate is function to use callback to control input/output data
func (ema EventModuleA) Validate(db *gorm.DB) {
	if err := ema.Valid(); err != nil {
		db.AddError(err)
	}
}

// EventAgent is model to contain event data linked with agent
type EventAgent struct {
	Agent Agent `form:"agent,omitempty" json:"agent,omitempty" gorm:"FOREIGNKEY:id;ASSOCIATION_FOREIGNKEY:agent_id;ASSOCIATION_AUTOUPDATE:false;ASSOCIATION_AUTOCREATE:false"`
	Event `form:"" json:""`
}

// Valid is function to control input/output data
func (ea EventAgent) Valid() error {
	if err := ea.Agent.Valid(); err != nil {
		return err
	}
	return ea.Event.Valid()
}

// Validate is function to use callback to control input/output data
func (ea EventAgent) Validate(db *gorm.DB) {
	if err := ea.Valid(); err != nil {
		db.AddError(err)
	}
}

// EventModuleAAgent is model to contain event data linked with agent and module agent
type EventModuleAAgent struct {
	ModuleA ModuleA `form:"module,omitempty" json:"module,omitempty" gorm:"FOREIGNKEY:id;ASSOCIATION_FOREIGNKEY:module_id;ASSOCIATION_AUTOUPDATE:false;ASSOCIATION_AUTOCREATE:false"`
	Agent   Agent   `form:"agent,omitempty" json:"agent,omitempty" gorm:"FOREIGNKEY:id;ASSOCIATION_FOREIGNKEY:agent_id;ASSOCIATION_AUTOUPDATE:false;ASSOCIATION_AUTOCREATE:false"`
	Event   `form:"" json:""`
}

// Valid is function to control input/output data
func (ema EventModuleAAgent) Valid() error {
	if err := ema.ModuleA.Valid(); err != nil {
		return err
	}
	if err := ema.Agent.Valid(); err != nil {
		return err
	}
	return ema.Event.Valid()
}

// Validate is function to use callback to control input/output data
func (ema EventModuleAAgent) Validate(db *gorm.DB) {
	if err := ema.Valid(); err != nil {
		db.AddError(err)
	}
}
