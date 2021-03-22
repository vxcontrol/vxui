package private

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/vxcontrol/vxui/models"
	"github.com/vxcontrol/vxui/utils"
)

type agentDetails struct {
	Hash             string `json:"hash"`
	ActiveModules    int    `json:"active_modules"`
	EventsPerLastDay int    `json:"events_per_last_day"`
}

type agents struct {
	Agents  []models.Agent `json:"agents"`
	Details []agentDetails `json:"details"`
}

type agent struct {
	Agent   models.Agent `json:"agent"`
	Details agentDetails `json:"details"`
}

type agentInfo struct {
	Name string `json:"name" binding:"max=255,required"`
	OS   string `json:"os" binding:"oneof=windows linux darwin,required" default:"linux" enums:"windows,linux,darwin"`
	Arch string `json:"arch" binding:"oneof=386 amd64,required" default:"amd64" enums:"386,amd64"`
}

const sqlAgentDetails = `SELECT a.hash,
	(SELECT COUNT(m.id) FROM modules m
		WHERE m.agent_id = a.id AND m.status = 'joined') as active_modules,
	(SELECT COUNT(e.id) FROM events e LEFT JOIN modules m ON e.module_id = m.id
		WHERE e.module_id = m.id AND m.agent_id = a.id AND
			e.date >= ( CURDATE() - INTERVAL 1 DAY )) as events_per_last_day
	FROM agents a`

// GetAgents is a function to return agent list view on dashboard
// @Summary Retrieve agents list
// @Tags Agents
// @Produce json
// @Success 200 {object} utils.successResp{data=agents} "agents list received successful"
// @Failure 403 {object} utils.errorResp "getting agents not permitted"
// @Failure 404 {object} utils.errorResp "agents not found"
// @Router /agents/ [get]
func GetAgents(c *gin.Context) {
	iDB := c.Keys["iDB"].(*gorm.DB)
	var resp agents
	if err := iDB.Order("description asc").Find(&resp.Agents).Error; err != nil {
		utils.HTTPError(c, http.StatusNotFound, "internal error on retrieving agent list")
		return
	}

	if err := iDB.Raw(sqlAgentDetails).Scan(&resp.Details).Error; err != nil {
		utils.HTTPError(c, http.StatusNotFound, "internal error on retrieving agents details")
		return
	}

	utils.HTTPSuccess(c, http.StatusOK, resp)
}

// GetAgent is a function to return agent info and details view
// @Summary Retrieve agent info by agent hash
// @Tags Agents
// @Produce json
// @Param hash path string true "agent hash in hex format (md5)" minlength(32) maxlength(32)
// @Success 200 {object} utils.successResp{data=agent} "agent info received successful"
// @Failure 403 {object} utils.errorResp "getting agent info not permitted"
// @Failure 404 {object} utils.errorResp "agent not found"
// @Router /agents/{hash} [get]
func GetAgent(c *gin.Context) {
	hash := c.Param("hash")

	iDB := c.Keys["iDB"].(*gorm.DB)
	var resp agent
	if err := iDB.Take(&resp.Agent, "hash = ?", hash).Error; err != nil {
		utils.HTTPError(c, http.StatusNotFound, "internal error on retrieving agent info")
		return
	}

	if err := iDB.Raw(sqlAgentDetails+` WHERE a.hash = ?`, hash).Scan(&resp.Details).Error; err != nil {
		utils.HTTPError(c, http.StatusNotFound, "internal error on retrieving agent details")
		return
	}

	utils.HTTPSuccess(c, http.StatusOK, resp)
}

// PatchAgent is a function to update agent public info only
// @Summary Update agent info by agent hash
// @Tags Agents
// @Accept json
// @Produce json
// @Param hash path string true "agent hash in hex format (md5)" minlength(32) maxlength(32)
// @Param json body models.Agent true "agent info as JSON data"
// @Success 200 {object} utils.successResp{data=agent} "agent info updated successful"
// @Failure 400 {object} utils.errorResp "invalid agent info"
// @Failure 403 {object} utils.errorResp "updating agent info not permitted"
// @Failure 404 {object} utils.errorResp "agent not found"
// @Router /agents/{hash} [post]
func PatchAgent(c *gin.Context) {
	hash := c.Param("hash")

	var agent models.Agent
	if err := c.ShouldBindJSON(&agent); err != nil || agent.Valid() != nil {
		utils.HTTPError(c, http.StatusBadRequest, "invalid agent data")
		return
	}

	iDB := c.Keys["iDB"].(*gorm.DB)
	if agent.Hash != hash {
		utils.HTTPError(c, http.StatusNotFound, "agent hash should be identical")
		return
	}

	if err := iDB.Model(&agent).Select("description").Updates(agent).Error; err != nil {
		utils.HTTPError(c, http.StatusNotFound, "internal error on updating agent info")
		return
	}

	utils.HTTPSuccess(c, http.StatusOK, agent)
}

// CreateAgent is a function to create new agent
// @Summary Create new agent in service
// @Tags Agents
// @Accept json
// @Produce json
// @Param json body agentInfo true "agent info to create one"
// @Success 201 {object} utils.successResp "agent created successful"
// @Failure 400 {object} utils.errorResp "invalid agent info"
// @Failure 403 {object} utils.errorResp "creating agent not permitted"
// @Failure 500 {object} utils.errorResp "internal error"
// @Router /agents/ [put]
func CreateAgent(c *gin.Context) {
	var info agentInfo
	if err := c.ShouldBindJSON(&info); err != nil {
		utils.HTTPError(c, http.StatusBadRequest, "failed to valid agent info")
		return
	}

	agent := models.Agent{
		Hash:        utils.MakeAgentHash(info.Name),
		IP:          "0.0.0.0:32768",
		Description: info.Name,
		Status:      "disconnected",
		Info: models.AgentInfo{
			OS: models.AgentOS{
				Type: info.OS,
				Arch: info.Arch,
				Name: "unknown",
			},
			User: models.AgentUser{
				Name:  "unknown",
				Group: "unknown",
			},
		},
	}

	iDB := c.Keys["iDB"].(*gorm.DB)
	if err := iDB.Create(&agent).Error; err != nil {
		utils.HTTPError(c, http.StatusInternalServerError, "failed to create agent to db")
		return
	}

	utils.HTTPSuccess(c, http.StatusCreated, struct{}{})
}

// DeleteAgent is a function to cascade delete agent
// @Summary Delete agent from instance DB
// @Tags Agents
// @Produce json
// @Param hash path string true "agent hash in hex format (md5)" minlength(32) maxlength(32)
// @Success 200 {object} utils.successResp "agent deleted successful"
// @Failure 403 {object} utils.errorResp "deleting agent not permitted"
// @Failure 404 {object} utils.errorResp "agent not found"
// @Failure 500 {object} utils.errorResp "internal error"
// @Router /agents/{hash} [delete]
func DeleteAgent(c *gin.Context) {
	hash := c.Param("hash")

	iDB := c.Keys["iDB"].(*gorm.DB)
	var agent models.Agent
	if err := iDB.Take(&agent, "hash = ?", hash).Error; err != nil {
		utils.HTTPError(c, http.StatusNotFound, "agent not found")
		return
	}

	if err := iDB.Delete(&agent).Error; err != nil {
		utils.HTTPError(c, http.StatusInternalServerError, "failed to delete agent")
		return
	}

	utils.HTTPSuccess(c, http.StatusOK, struct{}{})
}
