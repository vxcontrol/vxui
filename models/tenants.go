package models

import "github.com/jinzhu/gorm"

// Tenant is model to contain tenant information
type Tenant struct {
	ID     uint64 `form:"id,omitempty" json:"id,omitempty" validate:"min=0,numeric" gorm:"type:INT(10) UNSIGNED;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	Status string `form:"status" json:"status" validate:"oneof=active blocked,required" gorm:"type:ENUM('active','blocked');NOT NULL"`
}

// TableName returns the table name string to guaranty use correct table
func (t *Tenant) TableName() string {
	return "tenants"
}

// BeforeDelete hook defined for cascade delete
func (t *Tenant) BeforeDelete(db *gorm.DB) error {
	db = db.Model(&User{}).Where("tenant_id = ?", t.ID).Unscoped().Delete(&User{})
	db = db.Model(&Service{}).Where("tenant_id = ?", t.ID).Unscoped().Delete(&Service{})
	db = db.Model(&ModuleS{}).Where("tenant_id = ?", t.ID).Unscoped().Delete(&ModuleS{})

	return db.Error
}

// Valid is function to control input/output data
func (t Tenant) Valid() error {
	return validate.Struct(t)
}

// Validate is function to use callback to control input/output data
func (t Tenant) Validate(db *gorm.DB) {
	if err := t.Valid(); err != nil {
		db.AddError(err)
	}
}
