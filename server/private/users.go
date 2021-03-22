package private

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/vxcontrol/vxui/models"
	"github.com/vxcontrol/vxui/utils"
	"golang.org/x/crypto/bcrypt"
)

// GetCurrentUser is a function to return account information
// @Summary Retrieve current user information
// @Tags Users
// @Produce json
// @Success 200 {object} utils.successResp{data=models.UserGroup} "user info received successful"
// @Failure 403 {object} utils.errorResp "getting user not permitted"
// @Failure 404 {object} utils.errorResp "user not found"
// @Router /users/current [get]
func GetCurrentUser(c *gin.Context) {
	session := sessions.Default(c)
	uid := session.Get("uid").(uint64)

	gDB := c.Keys["gDB"].(*gorm.DB)
	var user models.UserGroup
	if err := gDB.Take(&user.User, "id = ?", uid).Related(&user.Group).Error; err != nil {
		utils.HTTPError(c, http.StatusNotFound, "user not found")
		return
	}

	utils.HTTPSuccess(c, http.StatusOK, user)
}

// ChangePasswordCurrentUser is a function to update account password
// @Summary Update password for current user (account)
// @Tags Users
// @Accept json
// @Produce json
// @Param json body models.Password true "container to validate and update account password"
// @Success 200 {object} utils.successResp "account password updated successful"
// @Failure 400 {object} utils.errorResp "account password form data invalid"
// @Failure 403 {object} utils.errorResp "updating account password not permitted"
// @Failure 500 {object} utils.errorResp "internal error on updating account password"
// @Router /users/current/password [post]
func ChangePasswordCurrentUser(c *gin.Context) {
	session := sessions.Default(c)
	uid := session.Get("uid").(uint64)

	var form models.Password
	if err := c.ShouldBindJSON(&form); err != nil || form.Valid() != nil {
		utils.HTTPError(c, http.StatusBadRequest, "post form invalid")
		return
	}

	gDB := c.Keys["gDB"].(*gorm.DB)
	var user models.User
	if err := gDB.Take(&user, "id = ?", uid).Error; err != nil {
		utils.HTTPError(c, http.StatusForbidden, "invalid current user")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.CurrentPassword)); err != nil {
		utils.HTTPError(c, http.StatusForbidden, "invalid current password")
		return
	}

	if encPass, err := utils.EncryptPassword(form.Password); err == nil {
		user.Password = string(encPass)
	} else {
		utils.HTTPError(c, http.StatusBadRequest, "invalid new password form data")
		return
	}

	if err := gDB.Model(&user).Select("password").Updates(user).Error; err != nil {
		utils.HTTPSuccess(c, http.StatusInternalServerError, "internal error")
		return
	}

	utils.HTTPSuccess(c, http.StatusOK, struct{}{})
}
