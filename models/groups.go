package models

import "github.com/jinzhu/gorm"

// Group is model to contain group information
type Group struct {
	ID   uint64 `form:"id,omitempty" json:"id,omitempty" validate:"min=0,numeric" gorm:"type:INT(10) UNSIGNED;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	Name string `form:"name" json:"name" validate:"max=50,required" gorm:"type:VARCHAR(50);NOT NULL;UNIQUE_INDEX"`
}

// TableName returns the table name string to guaranty use correct table
func (g *Group) TableName() string {
	return "groups"
}

// Valid is function to control input/output data
func (g Group) Valid() error {
	return validate.Struct(g)
}

// Validate is function to use callback to control input/output data
func (g Group) Validate(db *gorm.DB) {
	if err := g.Valid(); err != nil {
		db.AddError(err)
	}
}
