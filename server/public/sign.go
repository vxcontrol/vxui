package public

import (
	"net/http"

	"github.com/vxcontrol/vxui/models"
	"github.com/vxcontrol/vxui/utils"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// SignIn is function to login user in the system
// @Summary Login user into system
// @Tags Public
// @Accept json
// @Produce json
// @Param json body models.SignIn true "Sign In form JSON data"
// @Success 200 {object} utils.successResp "login successful"
// @Failure 400 {object} utils.errorResp "invalid login data"
// @Failure 401 {object} utils.errorResp "invalid login or password"
// @Failure 403 {object} utils.errorResp "login not permitted"
// @Router /signin [post]
func SignIn(c *gin.Context) {
	var data models.SignIn
	if err := c.ShouldBindJSON(&data); err != nil || data.Valid() != nil {
		utils.HTTPError(c, http.StatusBadRequest, "invalid login data")
		return
	}

	if !utils.NoRobot(data.Token, c.ClientIP()) {
		utils.HTTPError(c, http.StatusForbidden, "invalid login data")
		return
	}

	var user models.User
	gDB := c.Keys["gDB"].(*gorm.DB)
	if err := gDB.Take(&user, "mail = ?", data.Mail).Error; err != nil {
		utils.HTTPError(c, http.StatusUnauthorized, "invalid login or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data.Password)); err != nil {
		utils.HTTPError(c, http.StatusUnauthorized, "invalid login or password")
		return
	}

	if user.Status != "active" {
		utils.HTTPError(c, http.StatusForbidden, "user is inactive")
		return
	}

	var sid uint64
	var s models.Service
	if err := gDB.Take(&s, "tenant_id = ?", user.TenantID).Error; err == nil {
		sid = s.ID
	}

	session := sessions.Default(c)
	session.Set("uid", user.ID)
	session.Set("gid", user.GroupID)
	session.Set("tid", user.TenantID)
	session.Set("sid", sid)
	session.Save()
	utils.HTTPSuccess(c, http.StatusOK, struct{}{})
}

// SignUp is function to register user in the system
// @Summary Register user into system
// @Tags Public
// @Accept json
// @Produce json
// @Param json body models.SignUp true "Sign Up form JSON data"
// @Success 200 {object} utils.successResp "register successful"
// @Failure 400 {object} utils.errorResp "invalid registration data"
// @Failure 403 {object} utils.errorResp "register not permitted"
// @Failure 500 {object} utils.errorResp "couldn't perform insert in DB"
// @Router /signup [post]
func SignUp(c *gin.Context) {
	var data models.SignUp
	if err := c.ShouldBindJSON(&data); err != nil || data.Valid() != nil {
		utils.HTTPError(c, http.StatusBadRequest, "invalid registration data")
		return
	}

	if !utils.NoRobot(data.Token, c.ClientIP()) {
		utils.HTTPError(c, http.StatusForbidden, "invalid registration data")
		return
	}

	var user models.UserTenant
	user.User.FromSignUp(&data)
	user.Tenant.Status = "active"
	gDB := c.Keys["gDB"].(*gorm.DB)
	if encPass, err := utils.EncryptPassword(data.Password); err == nil {
		user.Password = string(encPass)
	} else {
		utils.HTTPError(c, http.StatusBadRequest, "invalid password value")
		return
	}

	if err := gDB.Create(&user).Error; err != nil {
		utils.HTTPError(c, http.StatusInternalServerError, "invalid registration data")
		return
	}

	utils.HTTPSuccess(c, http.StatusOK, struct{}{})
}
