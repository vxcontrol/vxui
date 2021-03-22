package private

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/vxcontrol/vxui/models"
	"github.com/vxcontrol/vxui/utils"
)

type events struct {
	Events []models.Event `json:"events"`
	Total  uint64         `json:"total"`
}

// GetEvents is a function to return event list view on dashboard
// @Summary Retrieve events list by filters
// @Tags Events
// @Produce json
// @Param request query utils.TableQuery true "query table params"
// @Success 200 {object} utils.successResp{data=events} "events list received successful"
// @Failure 400 {object} utils.errorResp "invalid event request data"
// @Failure 403 {object} utils.errorResp "getting events not permitted"
// @Failure 500 {object} utils.errorResp "invalid event data or query"
// @Router /events/ [get]
func GetEvents(c *gin.Context) {
	var query utils.TableQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.HTTPError(c, http.StatusBadRequest, "invalid event request data")
		return
	}

	query.Init("events", map[string]string{
		"id":                        "`{{table}}`.id",
		"hash":                      "`agents`.hash",
		"name":                      "`modules`.name",
		"data":                      "CONCAT(`{{table}}`.name, ' | ', `{{table}}`.data_text, ' | ', `{{table}}`.date)",
		"localizedDate":             "`{{table}}`.date",
		"localizedModuleName":       "JSON_EXTRACT(`modules`.locale, '$.module.{{lang}}.title')",
		"localizedEventName":        "JSON_EXTRACT(`modules`.locale, CONCAT('$.events.', `{{table}}`.name, '.{{lang}}.title'))",
		"localizedEventDescription": "JSON_EXTRACT(`modules`.locale, CONCAT('$.events.', `{{table}}`.name, '.{{lang}}.description'))",
	})

	iDB := c.Keys["iDB"].(*gorm.DB)
	var resp events
	funcs := []func(db *gorm.DB) *gorm.DB{
		func(db *gorm.DB) *gorm.DB {
			return db.Joins("JOIN agents ON agents.id = agent_id")
		},
		func(db *gorm.DB) *gorm.DB {
			return db.Joins("JOIN modules ON modules.id = module_id")
		}}

	var err error
	if resp.Total, err = query.Query(iDB, &resp.Events, funcs...); err != nil {
		utils.HTTPError(c, http.StatusInternalServerError, "invalid event query")
		return
	}

	for i := 0; i < len(resp.Events); i++ {
		if resp.Events[i].Valid() != nil {
			utils.HTTPError(c, http.StatusInternalServerError, "invalid event data")
			return
		}
	}

	utils.HTTPSuccess(c, http.StatusOK, resp)
}
