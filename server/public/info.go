package public

import (
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/vxcontrol/vxui/models"
	"github.com/vxcontrol/vxui/utils"
)

type info struct {
	Type             string                   `json:"type"`
	RecaptchaHTMLKey string                   `json:"recaptcha_html_key"`
	Server           models.ServiceInfoServer `json:"server"`
	User             models.User              `json:"user"`
	Group            models.Group             `json:"group"`
	Tenant           models.Tenant            `json:"tenant"`
}

// Info is function to return settings and current information about system and config
// @Summary Retrieve current user and system settings
// @Tags Public
// @Produce json
// @Success 200 {object} utils.successResp{data=info} "info received successful"
// @Failure 403 {object} utils.errorResp "getting info not permitted"
// @Failure 404 {object} utils.errorResp "user not found"
// @Router /info [get]
func Info(c *gin.Context) {
	var resp info
	resp.RecaptchaHTMLKey = os.Getenv("RECAPTCHA_HTML_KEY")
	session := sessions.Default(c)
	uid := session.Get("uid")
	sid := session.Get("sid")
	if uid == nil {
		resp.Type = "guest"
	} else {
		resp.Type = "user"
		gDB := c.Keys["gDB"].(*gorm.DB)
		if err := gDB.Take(&resp.User, "id = ?", uid).Related(&resp.Group).Related(&resp.Tenant).Error; err != nil {
			utils.HTTPError(c, http.StatusNotFound, "user not found")
			return
		}
		if sid != 0 {
			var service models.Service
			if err := gDB.Take(&service, "id = ?", sid).Error; err == nil {
				resp.Server = service.Info.Server
			}
		}
	}

	utils.HTTPSuccess(c, http.StatusOK, resp)
}
