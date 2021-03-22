package models

import (
	"strings"

	"github.com/jinzhu/gorm"
)

// User is model to contain user information
type User struct {
	ID       uint64 `form:"id,omitempty" json:"id,omitempty" validate:"min=0,numeric" gorm:"type:INT(10) UNSIGNED;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT"`
	Mail     string `form:"mail" json:"mail" validate:"max=50,email,required" gorm:"type:VARCHAR(50);NOT NULL;UNIQUE_INDEX"`
	Name     string `form:"name,omitempty" json:"name,omitempty" validate:"max=70,omitempty" gorm:"type:VARCHAR(70);NOT NULL;DEFAULT:''"`
	Password string `form:"-" json:"-" validate:"max=100,omitempty" gorm:"type:VARCHAR(100);NOT NULL"`
	Status   string `form:"status" json:"status" validate:"oneof=created active blocked,required" gorm:"type:ENUM('created','active','blocked');NOT NULL"`
	GroupID  uint64 `form:"group_id" json:"group_id" validate:"min=0,numeric" gorm:"type:INT(10) UNSIGNED;NOT NULL;DEFAULT:2"`
	TenantID uint64 `form:"tenant_id" json:"tenant_id" validate:"min=0,numeric" gorm:"type:INT(10) UNSIGNED;NOT NULL"`
}

// TableName returns the table name string to guaranty use correct table
func (u *User) TableName() string {
	return "users"
}

// FromSignUp receive all properties from SignUp object to current object
func (u *User) FromSignUp(sup *SignUp) {
	*u = sup.ToUser()
}

// FromSignIn receive all properties from SignIn object to current object
func (u *User) FromSignIn(sin *SignIn) {
	*u = sin.ToUser()
}

// Valid is function to control input/output data
func (u User) Valid() error {
	return validate.Struct(u)
}

// Validate is function to use callback to control input/output data
func (u User) Validate(db *gorm.DB) {
	if err := u.Valid(); err != nil {
		db.AddError(err)
	}
}

// SignIn is model to contain user information on SignIn procedure
type SignIn struct {
	Mail     string `form:"mail" json:"mail" validate:"max=50,required" gorm:"type:VARCHAR(50);NOT NULL;UNIQUE_INDEX"`
	Password string `form:"password" json:"password" validate:"min=5,max=100,required" gorm:"type:VARCHAR(100);NOT NULL"`
	Token    string `form:"token" json:"token" gorm:"-"`
}

// TableName returns the table name string to guaranty use correct table
func (sin *SignIn) TableName() string {
	return "users"
}

// ToUser receive all properties from SignUp object and return User object
func (sin *SignIn) ToUser() User {
	return User{
		Mail:     sin.Mail,
		Password: sin.Password,
	}
}

// Valid is function to control input/output data
func (sin SignIn) Valid() error {
	return validate.Struct(sin)
}

// Validate is function to use callback to control input/output data
func (sin SignIn) Validate(db *gorm.DB) {
	if err := sin.Valid(); err != nil {
		db.AddError(err)
	}
}

// SignUp is model to contain user information on SignUp procedure
type SignUp struct {
	Mail            string `form:"mail" json:"mail" validate:"max=50,email,required" gorm:"type:VARCHAR(50);NOT NULL;UNIQUE_INDEX"`
	Password        string `form:"password" json:"password" validate:"stpass,max=100,required" gorm:"type:VARCHAR(100);NOT NULL"`
	ConfirmPassword string `form:"confirm_password" json:"confirm_password" validate:"eqfield=Password" gorm:"-"`
	Token           string `form:"token" json:"token" gorm:"-"`
}

// TableName returns the table name string to guaranty use correct table
func (sup *SignUp) TableName() string {
	return "users"
}

// ToUser receive all properties from SignUp object and return User object
func (sup *SignUp) ToUser() User {
	return User{
		Mail:     sup.Mail,
		Name:     strings.Split(sup.Mail, "@")[0],
		Password: sup.Password,
		Status:   "created",
		GroupID:  1,
	}
}

// Valid is function to control input/output data
func (sup SignUp) Valid() error {
	return validate.Struct(sup)
}

// Validate is function to use callback to control input/output data
func (sup SignUp) Validate(db *gorm.DB) {
	if err := sup.Valid(); err != nil {
		db.AddError(err)
	}
}

// Password is model to contain user password to change it
type Password struct {
	CurrentPassword string `form:"current_password" json:"current_password" validate:"nefield=Password,min=8,max=100,required" gorm:"-"`
	Password        string `form:"password" json:"password" validate:"stpass,max=100,required" gorm:"type:VARCHAR(100);NOT NULL"`
	ConfirmPassword string `form:"confirm_password" json:"confirm_password" validate:"eqfield=Password" gorm:"-"`
}

// TableName returns the table name string to guaranty use correct table
func (p *Password) TableName() string {
	return "users"
}

// Valid is function to control input/output data
func (p Password) Valid() error {
	return validate.Struct(p)
}

// Validate is function to use callback to control input/output data
func (p Password) Validate(db *gorm.DB) {
	if err := p.Valid(); err != nil {
		db.AddError(err)
	}
}

// UserGroup is model to contain user information linked with user group
type UserGroup struct {
	Group Group `form:"group,omitempty" json:"group,omitempty" gorm:"FOREIGNKEY:id;ASSOCIATION_FOREIGNKEY:group_id;ASSOCIATION_AUTOUPDATE:false;ASSOCIATION_AUTOCREATE:false"`
	User  `form:"" json:""`
}

// Valid is function to control input/output data
func (ug UserGroup) Valid() error {
	if err := ug.Group.Valid(); err != nil {
		return err
	}
	return ug.User.Valid()
}

// Validate is function to use callback to control input/output data
func (ug UserGroup) Validate(db *gorm.DB) {
	if err := ug.Valid(); err != nil {
		db.AddError(err)
	}
}

// UserTenant is model to contain user information linked with user tenant
type UserTenant struct {
	Tenant Tenant `form:"tenant,omitempty" json:"tenant,omitempty" gorm:"FOREIGNKEY:TenantID;ASSOCIATION_FOREIGNKEY:ID"`
	User   `form:"" json:""`
}

// Valid is function to control input/output data
func (ut UserTenant) Valid() error {
	if err := ut.Tenant.Valid(); err != nil {
		return err
	}
	return ut.User.Valid()
}

// Validate is function to use callback to control input/output data
func (ut UserTenant) Validate(db *gorm.DB) {
	if err := ut.Valid(); err != nil {
		db.AddError(err)
	}
}

// UserGroupTenant is model to contain user information linked with user group and tenant
type UserGroupTenant struct {
	Group  Group  `form:"group,omitempty" json:"group,omitempty" gorm:"FOREIGNKEY:id;ASSOCIATION_FOREIGNKEY:group_id;ASSOCIATION_AUTOUPDATE:false;ASSOCIATION_AUTOCREATE:false"`
	Tenant Tenant `form:"tenant,omitempty" json:"tenant,omitempty" gorm:"FOREIGNKEY:id;ASSOCIATION_FOREIGNKEY:tenant_id;ASSOCIATION_AUTOUPDATE:false;ASSOCIATION_AUTOCREATE:false"`
	User   `form:"" json:""`
}

// Valid is function to control input/output data
func (ugt UserGroupTenant) Valid() error {
	if err := ugt.Group.Valid(); err != nil {
		return err
	}
	if err := ugt.Tenant.Valid(); err != nil {
		return err
	}
	return ugt.User.Valid()
}

// Validate is function to use callback to control input/output data
func (ugt UserGroupTenant) Validate(db *gorm.DB) {
	if err := ugt.Valid(); err != nil {
		db.AddError(err)
	}
}
