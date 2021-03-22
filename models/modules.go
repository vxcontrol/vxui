package models

import (
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

// ModuleConfig is a proprietary structure to contain module config
// E.x. {"property_1": "some property value", "property_2": 42, ...}
type ModuleConfig map[string]interface{}

// Valid is function to control input/output data
func (mc ModuleConfig) Valid() error {
	if err := validate.Var(mc, "required,dive,keys,solid,endkeys,required,valid"); err != nil {
		return err
	}
	return nil
}

// Value is interface function to return current value to store to DB
func (mc ModuleConfig) Value() (driver.Value, error) {
	b, err := json.Marshal(mc)
	return string(b), err
}

// Scan is interface function to parse DB value when getting from DB
func (mc *ModuleConfig) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), mc)
}

// EventConfigAction is a proprietary structure to describe an event config action
// E.x. {"name": "log_to_db", "type": "db"}
type EventConfigAction struct {
	Name string `form:"name" json:"name" validate:"solid,required"`
	Type string `form:"type" json:"type" validate:"oneof=db,required"`
}

// Valid is function to control input/output data
func (eca EventConfigAction) Valid() error {
	return validate.Struct(eca)
}

// EventConfigSeq is a proprietary structure to describe one of events sequence
// E.x. {"name": "log_to_db", "type": "db"}
type EventConfigSeq struct {
	Name     string `form:"name" json:"name" validate:"solid,required"`
	MinCount uint64 `form:"min_count" json:"min_count" validate:"min=1,numeric,required"`
}

// Valid is function to control input/output data
func (ecs EventConfigSeq) Valid() error {
	return validate.Struct(ecs)
}

// EventConfigItem is a proprietary structure to contain an event config
// It has to "type" and "actions" keys for atomic events
type EventConfigItem struct {
	Type     string                 `form:"type" json:"type" validate:"oneof=atomic aggregation correlation,required"`
	Actions  []EventConfigAction    `form:"actions" json:"actions" validate:"required,dive,valid"`
	Seq      []EventConfigSeq       `form:"seq,omitempty" json:"seq,omitempty" validate:"required_with=GroupBy,omitempty,unique,dive,required,valid"`
	GroupBy  []string               `form:"group_by,omitempty" json:"group_by,omitempty" validate:"required_with=Seq,unique,solid,omitempty"`
	MaxCount uint64                 `form:"max_count,omitempty" json:"max_count,omitempty" validate:"min=0,max=10000000,numeric,omitempty"`
	MaxTime  uint64                 `form:"max_time,omitempty" json:"max_time,omitempty" validate:"min=0,max=10000000,numeric,omitempty"`
	Config   map[string]interface{} `form:"-" json:"-" validate:"omitempty,dive,keys,solid,endkeys,required"`
}

// Valid is function to control input/output data
func (eci EventConfigItem) Valid() error {
	return validate.Struct(eci)
}

// MarshalJSON is a JSON interface function to make JSON data bytes array from the struct object
func (eci EventConfigItem) MarshalJSON() ([]byte, error) {
	var err error
	var data []byte
	raw := make(map[string]interface{}, 0)
	type eventConfigItem EventConfigItem
	if data, err = json.Marshal((*eventConfigItem)(&eci)); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	for k, v := range eci.Config {
		raw[k] = v
	}
	if eci.Type != "atomic" {
		raw["max_count"] = eci.MaxCount
		raw["max_time"] = eci.MaxTime
		raw["group_by"] = eci.GroupBy
		raw["seq"] = eci.Seq
	} else {
		delete(raw, "max_count")
		delete(raw, "max_time")
		delete(raw, "group_by")
		delete(raw, "seq")
	}
	return json.Marshal(raw)
}

// UnmarshalJSON is a JSON interface function to parse JSON data bytes array and to get struct object
func (eci *EventConfigItem) UnmarshalJSON(input []byte) error {
	var excludeKeys []string
	tp := reflect.TypeOf(EventConfigItem{})
	for i := 0; i < tp.NumField(); i++ {
		field := tp.Field(i)
		excludeKeys = append(excludeKeys, strings.Split(field.Tag.Get("json"), ",")[0])
	}
	type eventConfigItem EventConfigItem
	if err := json.Unmarshal(input, (*eventConfigItem)(eci)); err != nil {
		return err
	}
	raw := make(map[string]interface{}, 0)
	if err := json.Unmarshal(input, &raw); err != nil {
		return err
	}
	eci.Config = make(map[string]interface{}, 0)
	for k, v := range raw {
		if !stringInSlice(k, excludeKeys) {
			eci.Config[k] = v
		}
	}
	return nil
}

// EventConfig is a proprietary structure to contain events config of module
// E.x. {"event_id": {"type": "atomic|aggregation|correlation", "actions": [...], ...}}
type EventConfig map[string]EventConfigItem

// Valid is function to control input/output data
func (ec EventConfig) Valid() error {
	if err := validate.Var(ec, "required,dive,keys,solid,endkeys,required,valid"); err != nil {
		return err
	}
	return nil
}

// Value is interface function to return current value to store to DB
func (ec EventConfig) Value() (driver.Value, error) {
	b, err := json.Marshal(ec)
	return string(b), err
}

// Scan is interface function to parse DB value when getting from DB
func (ec *EventConfig) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), ec)
}

// ChangelogDesc is model to contain description of a version on a language
type ChangelogDesc struct {
	Date        string `form:"date" json:"date" validate:"cldate,required"`
	Title       string `form:"title" json:"title" validate:"max=300,required"`
	Description string `form:"description" json:"description" validate:"max=10000,required"`
}

// Valid is function to control input/output data
func (cld ChangelogDesc) Valid() error {
	return validate.Struct(cld)
}

// Changelog is a proprietary structure to contain changelog of module
// E.x. {"0.1.0": {"en": {...}, "ru": {...}}}
type Changelog map[string]map[string]ChangelogDesc

// Valid is function to control input/output data
func (cl Changelog) Valid() error {
	if err := validate.Var(cl, "required,dive,keys,semver,endkeys,required,len=2,dive,keys,oneof=ru en,endkeys,valid"); err != nil {
		return err
	}
	return nil
}

// Value is interface function to return current value to store to DB
func (cl Changelog) Value() (driver.Value, error) {
	b, err := json.Marshal(cl)
	return string(b), err
}

// Scan is interface function to parse DB value when getting from DB
func (cl *Changelog) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), cl)
}

// LocaleDesc is model to contain description of something on a language
type LocaleDesc struct {
	Title       string `form:"title" json:"title" validate:"max=300,required"`
	Description string `form:"description" json:"description" validate:"max=10000,required"`
}

// Valid is function to control input/output data
func (ld LocaleDesc) Valid() error {
	return validate.Struct(ld)
}

// Locale is a proprietary structure to contain locale of module
type Locale struct {
	Module      map[string]LocaleDesc                       `form:"module" json:"module" validate:"required,len=2,dive,keys,oneof=ru en,endkeys,valid"`
	Config      map[string]map[string]LocaleDesc            `form:"config" json:"config" validate:"required,dive,keys,solid,endkeys,required,len=2,dive,keys,oneof=ru en,endkeys,valid"`
	Events      map[string]map[string]LocaleDesc            `form:"events" json:"events" validate:"required,dive,keys,solid,endkeys,required,len=2,dive,keys,oneof=ru en,endkeys,valid"`
	EventConfig map[string]map[string]map[string]LocaleDesc `form:"event_config" json:"event_config" validate:"required,dive,keys,solid,endkeys,required,dive,keys,solid,endkeys,required,len=2,dive,keys,oneof=ru en,endkeys,valid"`
	EventData   map[string]map[string]LocaleDesc            `form:"event_data" json:"event_data" validate:"required,dive,keys,solid,endkeys,required,len=2,dive,keys,oneof=ru en,endkeys,valid"`
	Tags        map[string]map[string]LocaleDesc            `form:"tags" json:"tags" validate:"required,dive,keys,solid,endkeys,required,len=2,dive,keys,oneof=ru en,endkeys,valid"`
}

// Valid is function to control input/output data
func (l Locale) Valid() error {
	return validate.Struct(l)
}

// Value is interface function to return current value to store to DB
func (l Locale) Value() (driver.Value, error) {
	b, err := json.Marshal(l)
	return string(b), err
}

// Scan is interface function to parse DB value when getting from DB
func (l *Locale) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), l)
}

// ModuleInfo is model to contain general module information
type ModuleInfo struct {
	Name     string              `form:"name" json:"name" validate:"solid,required"`
	Template string              `form:"template" json:"template" validate:"oneof=generic empty custom,required"`
	Version  string              `form:"version" json:"version" validate:"semver,required"`
	OS       map[string][]string `form:"os" json:"os" validate:"min=1,dive,keys,oneof=windows linux darwin,endkeys,min=1,unique,required,dive,oneof=386 amd64"`
	System   bool                `form:"system" json:"system,required"`
	Tags     []string            `form:"tags" json:"tags" validate:"solid,omitempty"`
	Events   []string            `form:"events" json:"events" validate:"solid,omitempty"`
}

// Valid is function to control input/output data
func (mi ModuleInfo) Valid() error {
	return validate.Struct(mi)
}

// Value is interface function to return current value to store to DB
func (mi ModuleInfo) Value() (driver.Value, error) {
	b, err := json.Marshal(mi)
	return string(b), err
}

// Scan is interface function to parse DB value when getting from DB
func (mi *ModuleInfo) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), mi)
}

// ModuleS is model to contain system module information from global DB
type ModuleS struct {
	ID                 uint64       `form:"id,omitempty" json:"id,omitempty" validate:"min=0,numeric" gorm:"type:INT(10) UNSIGNED;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	ServiceType        string       `form:"service_type" json:"service_type" validate:"oneof=vxmonitor" gorm:"type:ENUM('vxmonitor');NOT NULL"`
	TenantID           uint64       `form:"tenant_id" json:"tenant_id" validate:"min=0,numeric" gorm:"type:INT(10) UNSIGNED;NOT NULL"`
	ConfigSchema       Schema       `form:"config_schema" json:"config_schema" validate:"required" gorm:"type:JSON;NOT NULL" swaggertype:"object"`
	DefaultConfig      ModuleConfig `form:"default_config" json:"default_config" validate:"required,valid" gorm:"type:JSON;NOT NULL"`
	EventDataSchema    Schema       `form:"event_data_schema" json:"event_data_schema" validate:"required" gorm:"type:JSON;NOT NULL" swaggertype:"object"`
	EventConfigSchema  Schema       `form:"event_config_schema" json:"event_config_schema" validate:"required" gorm:"type:JSON;NOT NULL" swaggertype:"object"`
	DefaultEventConfig EventConfig  `form:"default_event_config" json:"default_event_config" validate:"required,valid" gorm:"type:JSON;NOT NULL"`
	Changelog          Changelog    `form:"changelog" json:"changelog" validate:"required,valid" gorm:"type:JSON;NOT NULL"`
	Locale             Locale       `form:"locale" json:"locale" validate:"required" gorm:"type:JSON;NOT NULL"`
	Info               ModuleInfo   `form:"info" json:"info" validate:"required" gorm:"type:JSON;NOT NULL"`
	LastUpdate         time.Time    `form:"last_update,omitempty" json:"last_update,omitempty" validate:"omitempty" gorm:"type:DATETIME;NOT NULL;DEFAULT:CURRENT_TIMESTAMP"`
}

// TableName returns the table name string to guaranty use correct table
func (ms *ModuleS) TableName() string {
	return "modules"
}

// ToModuleA receive all properties from ModuleS object and return ModuleA object
func (ms *ModuleS) ToModuleA() ModuleA {
	return ModuleA{
		Status:             "joined",
		ConfigSchema:       ms.ConfigSchema,
		DefaultConfig:      ms.DefaultConfig,
		CurrentConfig:      ms.DefaultConfig,
		EventDataSchema:    ms.EventDataSchema,
		EventConfigSchema:  ms.EventConfigSchema,
		DefaultEventConfig: ms.DefaultEventConfig,
		CurrentEventConfig: ms.DefaultEventConfig,
		Changelog:          ms.Changelog,
		Locale:             ms.Locale,
		Info:               ms.Info,
		LastUpdate:         ms.LastUpdate,
	}
}

// Valid is function to control input/output data
func (ms ModuleS) Valid() error {
	return validate.Struct(ms)
}

// Validate is function to use callback to control input/output data
func (ms ModuleS) Validate(db *gorm.DB) {
	if err := ms.Valid(); err != nil {
		db.AddError(err)
	}
}

// ModuleSTenant is model to contain system module information linked with module tenant
type ModuleSTenant struct {
	Tenant  Tenant `form:"tenant,omitempty" json:"tenant,omitempty" gorm:"FOREIGNKEY:id;ASSOCIATION_FOREIGNKEY:tenant_id;ASSOCIATION_AUTOUPDATE:false;ASSOCIATION_AUTOCREATE:false"`
	ModuleS `form:"" json:""`
}

// Valid is function to control input/output data
func (mst ModuleSTenant) Valid() error {
	if err := mst.Tenant.Valid(); err != nil {
		return err
	}
	return mst.ModuleS.Valid()
}

// Validate is function to use callback to control input/output data
func (mst ModuleSTenant) Validate(db *gorm.DB) {
	if err := mst.Valid(); err != nil {
		db.AddError(err)
	}
}

// ModuleA is model to contain agent module information from instance DB
type ModuleA struct {
	ID                 uint64       `form:"id,omitempty" json:"id,omitempty" validate:"min=0,numeric" gorm:"type:INT(10) UNSIGNED;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	AgentID            uint64       `form:"agent_id" json:"agent_id" validate:"min=0,numeric" gorm:"type:INT(10) UNSIGNED;NOT NULL"`
	Status             string       `form:"status" json:"status" validate:"oneof=joined inactive,required" gorm:"type:ENUM('joined','inactive');NOT NULL"`
	JoinDate           time.Time    `form:"join_date,omitempty" json:"join_date,omitempty" validate:"omitempty" gorm:"type:DATETIME;NOT NULL;DEFAULT:CURRENT_TIMESTAMP"`
	ConfigSchema       Schema       `form:"config_schema" json:"config_schema" validate:"required" gorm:"type:JSON;NOT NULL" swaggertype:"object"`
	DefaultConfig      ModuleConfig `form:"default_config" json:"default_config" validate:"required,valid" gorm:"type:JSON;NOT NULL"`
	CurrentConfig      ModuleConfig `form:"current_config" json:"current_config" validate:"required,valid" gorm:"type:JSON;NOT NULL"`
	EventDataSchema    Schema       `form:"event_data_schema" json:"event_data_schema" validate:"required" gorm:"type:JSON;NOT NULL" swaggertype:"object"`
	EventConfigSchema  Schema       `form:"event_config_schema" json:"event_config_schema" validate:"required" gorm:"type:JSON;NOT NULL" swaggertype:"object"`
	DefaultEventConfig EventConfig  `form:"default_event_config" json:"default_event_config" validate:"required,valid" gorm:"type:JSON;NOT NULL"`
	CurrentEventConfig EventConfig  `form:"current_event_config" json:"current_event_config" validate:"required,valid" gorm:"type:JSON;NOT NULL"`
	Changelog          Changelog    `form:"changelog" json:"changelog" validate:"required,valid" gorm:"type:JSON;NOT NULL"`
	Locale             Locale       `form:"locale" json:"locale" validate:"required" gorm:"type:JSON;NOT NULL"`
	Info               ModuleInfo   `form:"info" json:"info" validate:"required" gorm:"type:JSON;NOT NULL"`
	LastUpdate         time.Time    `form:"last_update,omitempty" json:"last_update,omitempty" validate:"omitempty" gorm:"type:DATETIME;NOT NULL;DEFAULT:CURRENT_TIMESTAMP"`
}

// TableName returns the table name string to guaranty use correct table
func (ma *ModuleA) TableName() string {
	return "modules"
}

// FromModuleS receive all properties from ModuleS object to current object
func (ma *ModuleA) FromModuleS(ms *ModuleS) {
	*ma = ms.ToModuleA()
}

// BeforeDelete hook defined for cascade delete
func (ma *ModuleA) BeforeDelete(db *gorm.DB) error {
	return db.Model(&Event{}).Where("module_id = ?", ma.ID).Unscoped().Delete(&Event{}).Error
}

// Valid is function to control input/output data
func (ma ModuleA) Valid() error {
	return validate.Struct(ma)
}

// Validate is function to use callback to control input/output data
func (ma ModuleA) Validate(db *gorm.DB) {
	if err := ma.Valid(); err != nil {
		db.AddError(err)
	}
}

// ModuleAAgent is model to contain agent module information linked with module agent
type ModuleAAgent struct {
	Agent   Agent `form:"agent,omitempty" json:"agent,omitempty" gorm:"FOREIGNKEY:id;ASSOCIATION_FOREIGNKEY:agent_id;ASSOCIATION_AUTOUPDATE:false;ASSOCIATION_AUTOCREATE:false"`
	ModuleA `form:"" json:""`
}

// Valid is function to control input/output data
func (maa ModuleAAgent) Valid() error {
	if err := maa.Agent.Valid(); err != nil {
		return err
	}
	return maa.ModuleA.Valid()
}

// Validate is function to use callback to control input/output data
func (maa ModuleAAgent) Validate(db *gorm.DB) {
	if err := maa.Valid(); err != nil {
		db.AddError(err)
	}
}
