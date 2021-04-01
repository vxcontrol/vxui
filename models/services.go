package models

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/jinzhu/gorm"
)

// ServiceInfoDB is model to contain service external config to connetion to DB
type ServiceInfoDB struct {
	Name string `form:"name" json:"name" validate:"max=50,required"`
	User string `form:"user" json:"user" validate:"max=50,required"`
	Pass string `form:"pass" json:"pass" validate:"max=50,required"`
	Host string `form:"host" json:"host" validate:"max=50,required"`
	Port uint64 `form:"port" json:"port" validate:"min=1,max=65535,numeric,required"`
}

// Valid is function to control input/output data
func (sidb ServiceInfoDB) Valid() error {
	return validate.Struct(sidb)
}

// ServiceInfoS3 is model to contain service external config to connetion to S3
type ServiceInfoS3 struct {
	Endpoint   string `form:"endpoint" json:"endpoint" validate:"max=100,required"`
	AccessKey  string `form:"access_key" json:"access_key" validate:"max=50,required"`
	SecretKey  string `form:"secret_key" json:"secret_key" validate:"max=50,required"`
	BucketName string `form:"bucket_name" json:"bucket_name" validate:"max=30,required"`
}

// Valid is function to control input/output data
func (sis3 ServiceInfoS3) Valid() error {
	return validate.Struct(sis3)
}

// ServiceInfoServer is model to contain service external config to connetion to vxserver
type ServiceInfoServer struct {
	Proto string `form:"proto" json:"proto" validate:"oneof=ws wss,required"`
	Host  string `form:"host" json:"host" validate:"max=50,required"`
	Port  uint64 `form:"port" json:"port" validate:"min=1,max=65535,numeric,required"`
}

// Valid is function to control input/output data
func (sis ServiceInfoServer) Valid() error {
	return validate.Struct(sis)
}

// ServiceInfo is model to contain service external config to connetion (S3, DB, vxserver)
type ServiceInfo struct {
	DB     ServiceInfoDB     `form:"db" json:"db" validate:"required"`
	S3     ServiceInfoS3     `form:"s3" json:"s3" validate:"required"`
	Server ServiceInfoServer `form:"server" json:"server" validate:"required"`
}

// Valid is function to control input/output data
func (si ServiceInfo) Valid() error {
	return validate.Struct(si)
}

// Value is interface function to return current value to store to DB
func (si ServiceInfo) Value() (driver.Value, error) {
	b, err := json.Marshal(si)
	return string(b), err
}

// Scan is interface function to parse DB value when getting from DB
func (si *ServiceInfo) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), si)
}

// Service is model to contain service information
type Service struct {
	ID       uint64      `form:"id,omitempty" json:"id,omitempty" validate:"min=0,numeric" gorm:"type:INT(10) UNSIGNED;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	Name     string      `form:"name" json:"name" validate:"max=50,required" gorm:"type:VARCHAR(50);NOT NULL"`
	Type     string      `form:"type" json:"type" validate:"oneof=vxmonitor,required" gorm:"type:ENUM('vxmonitor');NOT NULL"`
	Status   string      `form:"status" json:"status" validate:"oneof=created active blocked removed,required" gorm:"type:ENUM('created','active','blocked','removed');NOT NULL"`
	TenantID uint64      `form:"tenant_id" json:"tenant_id" validate:"min=0,numeric" gorm:"type:INT(10) UNSIGNED;NOT NULL"`
	Info     ServiceInfo `form:"info" json:"info" validate:"required" gorm:"type:JSON;NOT NULL"`
}

// TableName returns the table name string to guaranty use correct table
func (s *Service) TableName() string {
	return "services"
}

// Valid is function to control input/output data
func (s Service) Valid() error {
	return validate.Struct(s)
}

// Validate is function to use callback to control input/output data
func (s Service) Validate(db *gorm.DB) {
	if err := s.Valid(); err != nil {
		db.AddError(err)
	}
}

// ServiceTenant is model to contain service information linked with service tenant
type ServiceTenant struct {
	Tenant  Tenant `form:"tenant,omitempty" json:"tenant,omitempty" gorm:"FOREIGNKEY:id;ASSOCIATION_FOREIGNKEY:tenant_id;ASSOCIATION_AUTOUPDATE:false;ASSOCIATION_AUTOCREATE:false"`
	Service `form:"" json:""`
}

// Valid is function to control input/output data
func (st ServiceTenant) Valid() error {
	if err := st.Tenant.Valid(); err != nil {
		return err
	}
	return st.Service.Valid()
}

// Validate is function to use callback to control input/output data
func (st ServiceTenant) Validate(db *gorm.DB) {
	if err := st.Valid(); err != nil {
		db.AddError(err)
	}
}
