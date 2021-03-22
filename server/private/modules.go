package private

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/vxcontrol/vxcommon/storage"
	"github.com/vxcontrol/vxui/models"
	"github.com/vxcontrol/vxui/utils"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type agentModuleDetails struct {
	Name   string `json:"name"`
	Today  uint64 `json:"today"`
	Total  uint64 `json:"total"`
	Active bool   `json:"active"`
	Update bool   `json:"update"`
}

type agentModules struct {
	Modules []models.ModuleA     `json:"modules"`
	Details []agentModuleDetails `json:"details"`
}

type agentModule struct {
	Module  models.ModuleA     `json:"module"`
	Details agentModuleDetails `json:"details"`
}

type agentModulePatch struct {
	// Action on agent module must be one of activate, deactivate, update, store
	Action string         `form:"action" json:"action" binding:"oneof=activate deactivate store update,required" default:"update" enums:"activate,deactivate,store,update"`
	Module models.ModuleA `form:"module,omitempty" json:"module,omitempty" binding:"required_if=Action store,omitempty"`
}

type systemModuleFile struct {
	Path string `form:"path" json:"path" binding:"required"`
	Data string `form:"data" json:"data" binding:"required" default:"base64"`
}

type systemModuleFilePatch struct {
	Action  string `form:"action" json:"action" binding:"oneof=move remove save,required" default:"save" enums:"move,remove,save"`
	Path    string `form:"path" json:"path" binding:"required"`
	Data    string `form:"data,omitempty" json:"data,omitempty" default:"base64" binding:"omitempty,required_if=Action save"`
	NewPath string `form:"newpath,omitempty" json:"newpath,omitempty" binding:"omitempty,required_if=Action move"`
}

const sqlAgentModuleDetails = `SELECT
		m.name,
		(m.status = "joined") as active,
		(SELECT COUNT(id) FROM events e1
			WHERE e1.module_id = m.id AND DATE(e1.date) = CURDATE()) as today,
		(SELECT COUNT(id) FROM events e2 WHERE e2.module_id = m.id) as total
	FROM modules m
		WHERE m.agent_id = ?`

// Template is container for all module files
type Template map[string]map[string][]byte

func joinPath(args ...string) string {
	tpath := filepath.Join(args...)
	return strings.Replace(tpath, "\\", "/", -1)
}

func removeLeadSlash(files map[string][]byte) map[string][]byte {
	rfiles := make(map[string][]byte)
	for name, data := range files {
		rfiles[name[1:]] = data
	}
	return rfiles
}

func readDir(s storage.IStorage, path string) ([]string, error) {
	var files []string
	list, err := s.ListDir(path)
	if err != nil {
		return files, err
	}
	for _, info := range list {
		if info.IsDir() {
			list, err := readDir(s, path+"/"+info.Name())
			if err != nil {
				return files, err
			}
			files = append(files, list...)
		} else {
			files = append(files, path+"/"+info.Name())
		}
	}
	return files, nil
}

func loadModuleSConfig(files map[string][]byte) (*models.ModuleS, error) {
	var module models.ModuleS
	targets := map[string]interface{}{
		"changelog":            &module.Changelog,
		"config_schema":        &module.ConfigSchema,
		"default_config":       &module.DefaultConfig,
		"default_event_config": &module.DefaultEventConfig,
		"event_config_schema":  &module.EventConfigSchema,
		"event_data_schema":    &module.EventDataSchema,
		"locale":               &module.Locale,
	}

	for filename, container := range targets {
		if err := json.Unmarshal(files[filename+".json"], container); err != nil {
			return nil, errors.New("failed unmarshal " + filename + ": " + err.Error())
		}
	}

	return &module, nil
}

func patchModuleSConfig(module *models.ModuleS) error {
	type patchLocaleCfg struct {
		src     map[string]map[string]models.LocaleDesc
		list    []string
		titles  map[string]string
		locType string
	}

	placeholder := "{{placeholder}}"
	languages := []string{"ru", "en"}
	patchLocaleCfgList := []patchLocaleCfg{
		{
			src:  module.Locale.Tags,
			list: module.Info.Tags,
			titles: map[string]string{
				"ru": " тег",
				"en": " tag",
			},
			locType: "tags",
		},
		{
			src:  module.Locale.Events,
			list: module.Info.Events,
			titles: map[string]string{
				"ru": " событие",
				"en": " event",
			},
			locType: "events",
		},
	}

	currentTime := time.Now()
	patchLocaleCl := map[string]string{
		"ru": currentTime.Format("02.01.2006"),
		"en": currentTime.Format("01-02-2006"),
	}

	module.EventConfigSchema.Required = module.Info.Events
	if ecsItems := module.EventConfigSchema.Properties; len(ecsItems) == 1 {
		if ecsItem, ok := ecsItems[placeholder]; ok {
			delete(ecsItems, placeholder)
			for _, eventID := range module.Info.Events {
				ecsItems[eventID] = ecsItem
			}
		} else {
			return errors.New("failed to get event_config_schema placeholder")
		}
	} else {
		return errors.New("event_config_schema is invalid format")
	}

	if decItems := module.DefaultEventConfig; len(decItems) == 1 {
		if decItem, ok := decItems[placeholder]; ok {
			delete(decItems, placeholder)
			for _, eventID := range module.Info.Events {
				decItems[eventID] = decItem
			}
		} else {
			return errors.New("failed to get default_event_config placeholder")
		}
	} else {
		return errors.New("default_event_config is invalid format")
	}

	if clItems := module.Changelog; len(clItems) == 1 {
		if clItem, ok := clItems[placeholder]; ok {
			delete(clItems, placeholder)
			for lng, date := range patchLocaleCl {
				if clDesc, ok := clItem[lng]; ok {
					clDesc.Date = date
					clItem[lng] = clDesc
				}
			}
			clItems[module.Info.Version] = clItem
		} else {
			return errors.New("failed to get changelog placeholder")
		}
	} else {
		return errors.New("changelog is invalid format")
	}

	for _, pLoc := range patchLocaleCfgList {
		if locItems := pLoc.src; len(locItems) == 1 {
			if locItemEtl, ok := locItems[placeholder]; ok {
				delete(locItems, placeholder)
				for _, itemID := range pLoc.list {
					locItem := make(map[string]models.LocaleDesc, 0)
					for _, lng := range languages {
						if itemDescEtl, ok := locItemEtl[lng]; ok {
							locItem[lng] = models.LocaleDesc{
								Title:       itemID + pLoc.titles[lng],
								Description: itemDescEtl.Description,
							}
						}
					}
					locItems[itemID] = locItem
				}
			} else {
				return errors.New("failed to get locale " + pLoc.locType + " placeholder")
			}
		} else {
			return errors.New("locale " + pLoc.locType + " is invalid format")
		}
	}

	if locItems := module.Locale.EventConfig; len(locItems) == 1 {
		if locItem, ok := locItems[placeholder]; ok {
			delete(locItems, placeholder)
			for _, itemID := range module.Info.Events {
				locItems[itemID] = locItem
			}
		} else {
			return errors.New("failed to get locale events config placeholder")
		}
	} else {
		return errors.New("locale events config is invalid format")
	}

	for _, lng := range languages {
		if itemDescEtl, ok := module.Locale.Module[lng]; ok {
			module.Locale.Module[lng] = models.LocaleDesc{
				Title:       module.Info.Name + " " + itemDescEtl.Title,
				Description: itemDescEtl.Description,
			}
		}
	}

	return nil
}

func buildModuleSConfig(module *models.ModuleS) (map[string][]byte, error) {
	files := make(map[string][]byte, 0)
	targets := map[string]interface{}{
		"changelog":            &module.Changelog,
		"config_schema":        &module.ConfigSchema,
		"current_config":       &module.DefaultConfig,
		"current_event_config": &module.DefaultEventConfig,
		"default_config":       &module.DefaultConfig,
		"default_event_config": &module.DefaultEventConfig,
		"event_config_schema":  &module.EventConfigSchema,
		"event_data_schema":    &module.EventDataSchema,
		"info":                 &module.Info,
		"locale":               &module.Locale,
	}

	for filename, container := range targets {
		var containerOut bytes.Buffer
		if containerData, err := json.Marshal(container); err == nil {
			json.Indent(&containerOut, containerData, "", "    ")
			files[filename+".json"] = containerOut.Bytes()
		} else {
			return nil, errors.New("failed marshal " + filename + ": " + err.Error())
		}
	}

	return files, nil
}

func buildModuleAConfig(module *models.ModuleA) (map[string][]byte, error) {
	files := make(map[string][]byte, 0)
	targets := map[string]interface{}{
		"changelog":            &module.Changelog,
		"config_schema":        &module.ConfigSchema,
		"current_config":       &module.CurrentConfig,
		"current_event_config": &module.CurrentEventConfig,
		"default_config":       &module.DefaultConfig,
		"default_event_config": &module.DefaultEventConfig,
		"event_config_schema":  &module.EventConfigSchema,
		"event_data_schema":    &module.EventDataSchema,
		"info":                 &module.Info,
		"locale":               &module.Locale,
	}

	for filename, container := range targets {
		var containerOut bytes.Buffer
		if containerData, err := json.Marshal(container); err == nil {
			json.Indent(&containerOut, containerData, "", "    ")
			files[filename+".json"] = containerOut.Bytes()
		} else {
			return nil, errors.New("failed marshal " + filename + ": " + err.Error())
		}
	}

	return files, nil
}

func loadModuleSTemplate(mi *models.ModuleInfo) (Template, *models.ModuleS, error) {
	fs, err := storage.NewFS()
	if err != nil {
		return nil, nil, errors.New("failed initialize FS driver: " + err.Error())
	}

	var module *models.ModuleS
	template := make(Template)
	loadModuleDir := func(dir string) (map[string][]byte, error) {
		tpath := joinPath("templates", mi.Template, dir)
		if fs.IsNotExist(tpath) {
			return nil, errors.New("template directory not found")
		}
		files, err := fs.ReadDirRec(tpath)
		if err != nil {
			return nil, errors.New("failed read template files: " + err.Error())
		}

		return removeLeadSlash(files), nil
	}

	for _, dir := range []string{"bmodule", "cmodule", "smodule"} {
		if files, err := loadModuleDir(dir); err == nil {
			template[dir] = files
		} else {
			return nil, nil, err
		}
	}

	if files, err := loadModuleDir("config"); err == nil {
		if module, err = loadModuleSConfig(files); err != nil {
			return nil, nil, err
		}
		module.Info = *mi
		if err = patchModuleSConfig(module); err != nil {
			return nil, nil, err
		}
		if cfiles, err := buildModuleSConfig(module); err == nil {
			template["config"] = cfiles
		} else {
			return nil, nil, err
		}
	} else {
		return nil, nil, err
	}

	return template, module, nil
}

func storeModuleToGlobalS3(mi *models.ModuleInfo, mf Template) error {
	s3, err := storage.NewS3()
	if err != nil {
		return errors.New("failed initialize S3 driver: " + err.Error())
	}

	for _, dir := range []string{"bmodule", "cmodule", "smodule", "config"} {
		for fpath, fdata := range mf[dir] {
			if err := s3.WriteFile(joinPath(mi.Name, mi.Version, dir, fpath), fdata); err != nil {
				return errors.New("failed write file to S3: " + err.Error())
			}
		}
	}

	return nil
}

func storeModuleToInstanceS3(mi *models.ModuleInfo, mf Template, sv *models.Service) error {
	s3, err := storage.NewS3(
		sv.Info.S3.Endpoint,
		sv.Info.S3.AccessKey,
		sv.Info.S3.SecretKey,
		sv.Info.S3.BucketName)
	if err != nil {
		return errors.New("failed initialize S3 driver: " + err.Error())
	}

	for _, dir := range []string{"bmodule", "cmodule", "smodule", "config"} {
		for fpath, fdata := range mf[dir] {
			if err := s3.WriteFile(joinPath(mi.Name, mi.Version, dir, fpath), fdata); err != nil {
				return errors.New("failed write file to S3: " + err.Error())
			}
		}
	}

	return nil
}

func copyModuleAFilesToInstanceS3(ma models.ModuleA, sv *models.Service) error {
	gS3, err := storage.NewS3()
	if err != nil {
		return errors.New("failed initialize global S3 driver: " + err.Error())
	}

	mfiles, err := gS3.ReadDirRec(joinPath(ma.Info.Name, ma.Info.Version))
	if err != nil {
		return errors.New("failed read system module files: " + err.Error())
	}

	ufiles, err := gS3.ReadDirRec("utils")
	if err != nil {
		return errors.New("failed read utils files: " + err.Error())
	}

	iS3, err := storage.NewS3(
		sv.Info.S3.Endpoint,
		sv.Info.S3.AccessKey,
		sv.Info.S3.SecretKey,
		sv.Info.S3.BucketName)
	if err != nil {
		return errors.New("failed initialize instance S3 driver: " + err.Error())
	}

	for fpath, fdata := range mfiles {
		if err := iS3.WriteFile(joinPath(ma.Info.Name, ma.Info.Version, fpath), fdata); err != nil {
			return errors.New("failed write system module file to S3: " + err.Error())
		}
	}

	for fpath, fdata := range ufiles {
		if err := iS3.WriteFile(joinPath("utils", fpath), fdata); err != nil {
			return errors.New("failed write utils file to S3: " + err.Error())
		}
	}

	return nil
}

// GetAgentModules is a function to return agent module list view on dashboard
// @Summary Retrieve agent modules by agent hash
// @Tags Agents,Modules
// @Produce json
// @Param hash path string true "agent hash in hex format (md5)" minlength(32) maxlength(32)
// @Success 200 {object} utils.successResp{data=agentModules} "agent modules received successful"
// @Failure 403 {object} utils.errorResp "getting agent modules not permitted"
// @Failure 404 {object} utils.errorResp "agent or modules not found"
// @Router /agents/{hash}/modules [get]
func GetAgentModules(c *gin.Context) {
	session := sessions.Default(c)
	tid := session.Get("tid").(uint64)
	hash := c.Param("hash")

	sv := c.Keys["SV"].(*models.Service)
	gDB := c.Keys["gDB"].(*gorm.DB)
	var modulesS []models.ModuleS
	if err := gDB.Omit("id").Order("name asc").Find(&modulesS, "tenant_id IN (0, ?) AND service_type = ?", tid, sv.Type).Error; err != nil {
		utils.HTTPError(c, http.StatusNotFound, "system modules not found")
		return
	}

	iDB := c.Keys["iDB"].(*gorm.DB)
	var agent models.Agent
	if err := iDB.Take(&agent, "hash = ?", hash).Error; err != nil {
		utils.HTTPError(c, http.StatusNotFound, "agent not found")
		return
	}

	var resp agentModules
	if err := iDB.Order("status asc").Order("name asc").Find(&resp.Modules, "agent_id = ?", agent.ID).Error; err != nil {
		utils.HTTPError(c, http.StatusNotFound, "agent modules not found")
		return
	}

	if err := iDB.Raw(sqlAgentModuleDetails, agent.ID).Scan(&resp.Details).Error; err != nil {
		utils.HTTPError(c, http.StatusNotFound, "failed to retrieve agents modules details")
		return
	}

	getModule := func(name string) *models.ModuleA {
		for _, m := range resp.Modules {
			if m.Info.Name == name {
				return &m
			}
		}
		return nil
	}
	for _, ms := range modulesS {
		if ma := getModule(ms.Info.Name); ma == nil {
			mt := ms.ToModuleA()
			mt.Status = "inactive"
			resp.Modules = append(resp.Modules, mt)
			resp.Details = append(resp.Details, agentModuleDetails{Name: ms.Info.Name})
		} else {
			for imd := range resp.Details {
				md := &resp.Details[imd]
				if md.Name == ms.Info.Name {
					md.Update = ma.Info.Version != ms.Info.Version || ma.LastUpdate != ms.LastUpdate
					break
				}
			}
		}
	}

	utils.HTTPSuccess(c, http.StatusOK, resp)
}

// GetAgentModule is a function to return agent module by name
// @Summary Retrieve agent module data by agent hash and module name
// @Tags Agents,Modules
// @Produce json
// @Param hash path string true "agent hash in hex format (md5)" minlength(32) maxlength(32)
// @Param module_name path string true "module name without spaces"
// @Success 200 {object} utils.successResp{data=models.ModuleA} "agent module data received successful"
// @Failure 403 {object} utils.errorResp "getting agent module data not permitted"
// @Failure 404 {object} utils.errorResp "agent or module not found"
// @Router /agents/{hash}/modules/{module_name} [get]
func GetAgentModule(c *gin.Context) {
	moduleName := c.Param("module_name")
	hash := c.Param("hash")

	iDB := c.Keys["iDB"].(*gorm.DB)
	var agent models.Agent
	if err := iDB.Take(&agent, "hash = ?", hash).Error; err != nil {
		utils.HTTPError(c, http.StatusNotFound, "agent not found")
		return
	}

	var module models.ModuleA
	if err := iDB.Take(&module, "agent_id = ? AND name = ?", agent.ID, moduleName).Error; err != nil {
		utils.HTTPError(c, http.StatusNotFound, "module not found")
		return
	}

	utils.HTTPSuccess(c, http.StatusOK, module)
}

// PatchAgentModule is a function to update agent module info and status
// @Summary Update or patch agent module data by agent hash and module name
// @Tags Agents,Modules
// @Accept json
// @Produce json
// @Param hash path string true "agent hash in hex format (md5)" minlength(32) maxlength(32)
// @Param module_name path string true "module name without spaces"
// @Param json body agentModulePatch true "action on agent module as JSON data (activate, deactivate, store, update)"
// @Success 200 {object} utils.successResp "agent module patched successful"
// @Failure 403 {object} utils.errorResp "updating agent module not permitted"
// @Failure 404 {object} utils.errorResp "agent or module not found"
// @Failure 500 {object} utils.errorResp "internal error on updating agent module"
// @Router /agents/{hash}/modules/{module_name} [post]
func PatchAgentModule(c *gin.Context) {
	session := sessions.Default(c)
	tid := session.Get("tid").(uint64)
	hash := c.Param("hash")
	moduleName := c.Param("module_name")

	iDB := c.Keys["iDB"].(*gorm.DB)
	var agent models.Agent
	if err := iDB.Take(&agent, "hash = ?", hash).Error; err != nil {
		utils.HTTPError(c, http.StatusNotFound, "agent not found")
		return
	}

	sv := c.Keys["SV"].(*models.Service)
	gDB := c.Keys["gDB"].(*gorm.DB)
	var moduleS models.ModuleS
	if err := gDB.Take(&moduleS, "name = ? AND tenant_id IN (0, ?) AND service_type = ?", moduleName, tid, sv.Type).Error; err != nil {
		utils.HTTPError(c, http.StatusNotFound, "system module not found")
		return
	}

	var moduleA models.ModuleA
	if err := iDB.Take(&moduleA, "agent_id = ? AND name = ?", agent.ID, moduleName).Error; err != nil {
		moduleA.FromModuleS(&moduleS)
		moduleA.AgentID = agent.ID
	}

	var form agentModulePatch
	if err := c.ShouldBindJSON(&form); err != nil {
		utils.HTTPError(c, http.StatusNotFound, "post form not found or unknown")
		return
	}

	incl := []interface{}{"status", "last_update"}
	excl := []string{
		"agent_id",
		"status",
		"join_date",
		"current_config",
		"current_event_config",
	}
	switch form.Action {
	case "activate":
		if moduleA.ID == 0 {
			if err := copyModuleAFilesToInstanceS3(moduleA, sv); err != nil {
				break
			}
			if err := iDB.Create(&moduleA).Error; err == nil {
				utils.HTTPSuccess(c, http.StatusOK, struct{}{})
				return
			}
		} else {
			moduleA.Status = "joined"
			if err := iDB.Model(&moduleA).Select("", incl...).UpdateColumns(moduleA).Error; err == nil {
				utils.HTTPSuccess(c, http.StatusOK, struct{}{})
				return
			}
		}

	case "deactivate":
		if moduleA.ID != 0 {
			moduleA.Status = "inactive"
			if err := iDB.Model(&moduleA).Select("", incl...).UpdateColumns(moduleA).Error; err == nil {
				utils.HTTPSuccess(c, http.StatusOK, struct{}{})
				return
			}
		}

	case "store":
		if form.Module.Valid() != nil {
			utils.HTTPError(c, http.StatusBadRequest, "failed to valid new module data")
			return
		}

		changes := []bool{
			moduleA.ID != form.Module.ID,
			moduleA.Info.Name != form.Module.Info.Name,
			moduleA.Info.System != form.Module.Info.System,
			moduleA.Info.Template != form.Module.Info.Template,
			moduleA.Info.Version != form.Module.Info.Version,
			moduleA.AgentID != form.Module.AgentID,
			moduleA.JoinDate != form.Module.JoinDate,
		}
		for _, ch := range changes {
			if ch {
				utils.HTTPError(c, http.StatusInternalServerError, "failed accept system changes")
				return
			}
		}

		template := make(Template)
		form.Module.LastUpdate = time.Now()
		cfiles, err := buildModuleAConfig(&form.Module)
		if err != nil {
			utils.HTTPError(c, http.StatusInternalServerError, "failed to build module files")
			return
		}

		template["config"] = cfiles
		if err := storeModuleToInstanceS3(&form.Module.Info, template, sv); err != nil {
			utils.HTTPError(c, http.StatusInternalServerError, "failed to update module to s3")
			return
		}

		if err := iDB.Save(&form.Module).Error; err == nil {
			utils.HTTPSuccess(c, http.StatusOK, struct{}{})
			return
		}

	case "update":
		if err := copyModuleAFilesToInstanceS3(moduleA, sv); err != nil {
			break
		}
		if moduleA.ID != 0 {
			if err := iDB.Model(&moduleA).Omit(excl...).UpdateColumns(moduleS.ToModuleA()).Error; err == nil {
				utils.HTTPSuccess(c, http.StatusOK, struct{}{})
				return
			}
		}

	default:
		utils.HTTPError(c, http.StatusNotFound, "action not found or unknown")
		return
	}

	utils.HTTPError(c, http.StatusInternalServerError, "internal error")
}

// GetAgentBModule is a function to return bmodule vue code as a file
// @Summary Retrieve browser module vue code by agent hash and module name
// @Tags Agents,Modules
// @Produce text/javascript,application/javascript,json
// @Param hash path string true "agent hash in hex format (md5)" minlength(32) maxlength(32)
// @Param module_name path string true "module name without spaces"
// @Success 200 {file} file "browser module vue code as a file"
// @Failure 403 {object} utils.errorResp "getting agent module data not permitted"
// @Router /agents/{hash}/modules/{module_name}/bmodule.vue [get]
func GetAgentBModule(c *gin.Context) {
	var data []byte
	hash := c.Param("hash")
	moduleName := c.Param("module_name")
	defer func() {
		moduleName += ".js"
		c.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%q", moduleName))
		c.Data(http.StatusOK, "text/javascript", data)
	}()

	iDB := c.Keys["iDB"].(*gorm.DB)
	var agent models.Agent
	if err := iDB.Take(&agent, "hash = ?", hash).Error; err != nil {
		return
	}

	var module models.ModuleA
	if err := iDB.Take(&module, "agent_id = ? AND name = ?", agent.ID, moduleName).Error; err != nil {
		return
	}

	sv := c.Keys["SV"].(*models.Service)
	s3, err := storage.NewS3(
		sv.Info.S3.Endpoint,
		sv.Info.S3.AccessKey,
		sv.Info.S3.SecretKey,
		sv.Info.S3.BucketName)
	if err != nil {
		return
	}

	path := moduleName + "/" + module.Info.Version + "/bmodule/main.vue"
	if data, err = s3.ReadFile(path); err != nil {
		return
	}
}

// GetModules is a function to return system module list
// @Summary Retrieve system modules
// @Tags Modules
// @Produce json
// @Success 200 {object} utils.successResp{data=[]models.ModuleS} "system modules received successful"
// @Failure 403 {object} utils.errorResp "getting system modules not permitted"
// @Failure 404 {object} utils.errorResp "system modules not found"
// @Router /modules/ [get]
func GetModules(c *gin.Context) {
	session := sessions.Default(c)
	tid := session.Get("tid").(uint64)

	sv := c.Keys["SV"].(*models.Service)
	gDB := c.Keys["gDB"].(*gorm.DB)
	var modules []models.ModuleS
	if err := gDB.Order("name asc").Find(&modules, "tenant_id IN (0, ?) AND service_type = ?", tid, sv.Type).Error; err != nil {
		utils.HTTPError(c, http.StatusNotFound, "modules not found")
		return
	}

	utils.HTTPSuccess(c, http.StatusOK, modules)
}

// GetModule is a function to return system module by name
// @Summary Retrieve system module data by module name
// @Tags Modules
// @Produce json
// @Param module_name path string true "module name without spaces"
// @Success 200 {object} utils.successResp{data=models.ModuleS} "system module data received successful"
// @Failure 403 {object} utils.errorResp "getting system module data not permitted"
// @Failure 404 {object} utils.errorResp "system module not found"
// @Router /modules/{module_name} [get]
func GetModule(c *gin.Context) {
	session := sessions.Default(c)
	tid := session.Get("tid").(uint64)
	moduleName := c.Param("module_name")

	sv := c.Keys["SV"].(*models.Service)
	gDB := c.Keys["gDB"].(*gorm.DB)
	var module models.ModuleS
	if err := gDB.Take(&module, "name = ? AND tenant_id IN (0, ?) AND service_type = ?", moduleName, tid, sv.Type).Error; err != nil {
		utils.HTTPError(c, http.StatusNotFound, "module not found")
		return
	}

	utils.HTTPSuccess(c, http.StatusOK, module)
}

// GetModuleOption is a function to return option of system module rendered on server side
// @Summary Retrieve rendered Event Config Schema of system module data by module name
// @Tags Modules
// @Produce json
// @Param module_name path string true "module name without spaces"
// @Param option_name path string true "module option without spaces" Enums(id, service_type, tenant_id, config_schema, default_config, event_data_schema, event_config_schema, default_event_config, changelog, locale, info, last_update, definitions)
// @Success 200 {object} utils.successResp{data=interface{}} "module option received successful"
// @Failure 403 {object} utils.errorResp "getting module option not permitted"
// @Failure 404 {object} utils.errorResp "system module not found"
// @Failure 500 {object} utils.errorResp "internal error"
// @Router /modules/{module_name}/options/{option_name} [get]
func GetModuleOption(c *gin.Context) {
	session := sessions.Default(c)
	tid := session.Get("tid").(uint64)
	moduleName := c.Param("module_name")
	optionName := c.Param("option_name")

	sv := c.Keys["SV"].(*models.Service)
	gDB := c.Keys["gDB"].(*gorm.DB)
	var module models.ModuleS
	if err := gDB.Take(&module, "name = ? AND tenant_id = ? AND service_type = ?", moduleName, tid, sv.Type).Error; err != nil {
		utils.HTTPError(c, http.StatusNotFound, "module not found")
		return
	}

	if optionName == "definitions" {
		utils.HTTPSuccess(c, http.StatusOK, models.GetECSDefinitions(nil))
		return
	}

	options := make(map[string]json.RawMessage, 0)
	if data, err := json.Marshal(module); err != nil {
		utils.HTTPError(c, http.StatusInternalServerError, "failed to make json")
		return
	} else if err := json.Unmarshal(data, &options); err != nil {
		utils.HTTPError(c, http.StatusInternalServerError, "failed to parse json")
		return
	} else if _, ok := options[optionName]; !ok {
		utils.HTTPError(c, http.StatusNotFound, "option not found")
		return
	}

	utils.HTTPSuccess(c, http.StatusOK, options[optionName])
}

// GetModuleFiles is a function to return system module file list
// @Summary Retrieve system module files (relative path) by module name
// @Tags Modules
// @Produce json
// @Param module_name path string true "module name without spaces"
// @Success 200 {object} utils.successResp{data=[]string} "system module files received successful"
// @Failure 403 {object} utils.errorResp "getting system module files not permitted"
// @Failure 404 {object} utils.errorResp "system module not found"
// @Failure 500 {object} utils.errorResp "internal error"
// @Router /modules/{module_name}/files [get]
func GetModuleFiles(c *gin.Context) {
	session := sessions.Default(c)
	tid := session.Get("tid").(uint64)
	moduleName := c.Param("module_name")

	sv := c.Keys["SV"].(*models.Service)
	gDB := c.Keys["gDB"].(*gorm.DB)
	var module models.ModuleS
	if err := gDB.Take(&module, "name = ? AND tenant_id = ? AND service_type = ?", moduleName, tid, sv.Type).Error; err != nil {
		utils.HTTPError(c, http.StatusNotFound, "module not found")
		return
	}

	s3, err := storage.NewS3()
	if err != nil {
		utils.HTTPError(c, http.StatusInternalServerError, "failed to initialize S3 driver")
		return
	}

	files, err := readDir(s3, moduleName)
	if err != nil {
		utils.HTTPError(c, http.StatusInternalServerError, "failed to listening module files")
		return
	}

	utils.HTTPSuccess(c, http.StatusOK, files)
}

// GetModuleFile is a function to return system module file content
// @Summary Retrieve system module file content (in base64) by module name and relative path
// @Tags Modules
// @Produce json
// @Param module_name path string true "module name without spaces"
// @Param path query string true "relative path to module file"
// @Success 200 {object} utils.successResp{data=systemModuleFile} "system module file content received successful"
// @Failure 403 {object} utils.errorResp "getting system module file content not permitted"
// @Failure 404 {object} utils.errorResp "system module not found"
// @Failure 500 {object} utils.errorResp "internal error"
// @Router /modules/{module_name}/files/file [get]
func GetModuleFile(c *gin.Context) {
	session := sessions.Default(c)
	tid := session.Get("tid").(uint64)
	moduleName := c.Param("module_name")

	sv := c.Keys["SV"].(*models.Service)
	gDB := c.Keys["gDB"].(*gorm.DB)
	var module models.ModuleS
	if err := gDB.Take(&module, "name = ? AND tenant_id = ? AND service_type = ?", moduleName, tid, sv.Type).Error; err != nil {
		utils.HTTPError(c, http.StatusNotFound, "module not found")
		return
	}

	s3, err := storage.NewS3()
	if err != nil {
		utils.HTTPError(c, http.StatusInternalServerError, "failed to initialize S3 driver")
		return
	}

	var filePath = c.Query("path")
	if !strings.HasPrefix(filePath, moduleName+"/") || strings.Contains(filePath, "..") {
		utils.HTTPError(c, http.StatusForbidden, "failed to parse path to file")
		return
	}

	fileData, err := s3.ReadFile(filePath)
	if err != nil {
		utils.HTTPError(c, http.StatusInternalServerError, "failed to read module file")
		return
	}

	data := base64.URLEncoding.EncodeToString(fileData)
	utils.HTTPSuccess(c, http.StatusOK, systemModuleFile{Path: filePath, Data: data})
}

// PatchModuleFile is a function to save, move, remove of system module file and its content
// @Summary Patch system module file and content (in base64) by module name and relative path
// @Tags Modules
// @Accept json
// @Produce json
// @Param module_name path string true "module name without spaces"
// @Param json body systemModuleFilePatch true "action, relative path and file content for module file"
// @Success 200 {object} utils.successResp "action on system module file did successful"
// @Failure 403 {object} utils.errorResp "action on system module file not permitted"
// @Failure 404 {object} utils.errorResp "system module not found"
// @Failure 500 {object} utils.errorResp "internal error"
// @Router /modules/{module_name}/files/file [post]
func PatchModuleFile(c *gin.Context) {
	session := sessions.Default(c)
	tid := session.Get("tid").(uint64)
	moduleName := c.Param("module_name")

	sv := c.Keys["SV"].(*models.Service)
	gDB := c.Keys["gDB"].(*gorm.DB)
	var module models.ModuleS
	if err := gDB.Take(&module, "name = ? AND tenant_id = ? AND service_type = ?", moduleName, tid, sv.Type).Error; err != nil {
		utils.HTTPError(c, http.StatusNotFound, "module not found")
		return
	}

	s3, err := storage.NewS3()
	if err != nil {
		utils.HTTPError(c, http.StatusInternalServerError, "failed to initialize S3 driver")
		return
	}

	var form systemModuleFilePatch
	if err = c.ShouldBindJSON(&form); err != nil {
		utils.HTTPError(c, http.StatusForbidden, "failed to bind input form")
		return
	}

	if !strings.HasPrefix(form.Path, moduleName+"/") || strings.Contains(form.Path, "..") {
		utils.HTTPError(c, http.StatusForbidden, "failed to parse path to object")
		return
	}

	switch form.Action {
	case "save":
		data, err := base64.StdEncoding.DecodeString(form.Data)
		if err != nil {
			utils.HTTPError(c, http.StatusInternalServerError, "failed to parse module file content")
			return
		}

		if err = s3.WriteFile(form.Path, data); err != nil {
			utils.HTTPError(c, http.StatusInternalServerError, "failed to write module file")
			return
		}

	case "remove":
		if err = s3.Remove(form.Path); err != nil {
			utils.HTTPError(c, http.StatusInternalServerError, "failed to write module object")
			return
		}

	case "move":
		if !strings.HasPrefix(form.NewPath, moduleName+"/") || strings.Contains(form.NewPath, "..") {
			utils.HTTPError(c, http.StatusForbidden, "failed to parse newpath to object")
			return
		}

		if info, err := s3.GetInfo(form.Path); err != nil {
			utils.HTTPError(c, http.StatusInternalServerError, "failed to find object by path")
			return
		} else if !info.IsDir() {
			if strings.HasSuffix(form.NewPath, "/") {
				form.NewPath += info.Name()
			}

			if form.Path == form.NewPath {
				utils.HTTPError(c, http.StatusInternalServerError, "newpath is identical to path")
				return
			}

			if err = s3.Rename(form.Path, form.NewPath); err != nil {
				utils.HTTPError(c, http.StatusInternalServerError, "failed to move object")
				return
			}
		} else {
			if !strings.HasSuffix(form.Path, "/") {
				form.Path += "/"
			}
			if !strings.HasSuffix(form.NewPath, "/") {
				form.NewPath += "/"
			}

			if form.Path == form.NewPath {
				utils.HTTPError(c, http.StatusInternalServerError, "newpath is identical to path")
				return
			}

			files, err := s3.ListDirRec(form.Path)
			if err != nil {
				utils.HTTPError(c, http.StatusInternalServerError, "failed to get files by path")
				return
			}

			for obj, info := range files {
				if !info.IsDir() {
					curfile := filepath.Join(form.Path, obj)
					newfile := filepath.Join(form.NewPath, obj)
					if err = s3.Rename(curfile, newfile); err != nil {
						utils.HTTPError(c, http.StatusInternalServerError, "failed to move object")
						return
					}
				}
			}
		}

	default:
		utils.HTTPError(c, http.StatusForbidden, "action not permitted")
		return
	}

	if err := gDB.Model(&module).UpdateColumn("last_update", gorm.Expr("NOW()")).Error; err != nil {
		utils.HTTPError(c, http.StatusInternalServerError, "failed to update system module")
		return
	}

	utils.HTTPSuccess(c, http.StatusOK, struct{}{})
}

// CreateModule is a function to create new system module
// @Summary Create new system module from template
// @Tags Modules
// @Accept json
// @Produce json
// @Param json body models.ModuleInfo true "module info to create one"
// @Success 201 {object} utils.successResp "system module created successful"
// @Failure 400 {object} utils.errorResp "invalid system module info"
// @Failure 403 {object} utils.errorResp "creating system module not permitted"
// @Failure 500 {object} utils.errorResp "internal error"
// @Router /modules/ [put]
func CreateModule(c *gin.Context) {
	session := sessions.Default(c)
	tid := session.Get("tid").(uint64)
	sv := c.Keys["SV"].(*models.Service)

	var info models.ModuleInfo
	if err := c.ShouldBindJSON(&info); err != nil || info.Valid() != nil {
		utils.HTTPError(c, http.StatusBadRequest, "failed to valid module info")
		return
	}

	info.System = false
	template, module, err := loadModuleSTemplate(&info)
	if err != nil {
		utils.HTTPError(c, http.StatusInternalServerError, "failed to load module")
		return
	}

	module.TenantID = tid
	module.ServiceType = sv.Type
	if err = module.Valid(); err != nil {
		utils.HTTPError(c, http.StatusInternalServerError, "failed to validate module")
		return
	}

	if err := storeModuleToGlobalS3(&info, template); err != nil {
		utils.HTTPError(c, http.StatusInternalServerError, "failed to store module to s3")
		return
	}

	gDB := c.Keys["gDB"].(*gorm.DB)
	if err := gDB.Create(module).Error; err != nil {
		utils.HTTPError(c, http.StatusInternalServerError, "failed to store module to db")
		return
	}

	utils.HTTPSuccess(c, http.StatusCreated, struct{}{})
}

// UpdateModule is a function to update system module
// @Summary Update current version of system module to global DB and global S3 storage
// @Tags Modules
// @Accept json
// @Produce json
// @Param json body models.ModuleS true "module info to create one"
// @Param module_name path string true "module name without spaces"
// @Success 200 {object} utils.successResp "system module updated successful"
// @Failure 403 {object} utils.errorResp "updating system module not permitted"
// @Failure 404 {object} utils.errorResp "system module not found"
// @Failure 500 {object} utils.errorResp "internal error"
// @Router /modules/{module_name} [post]
func UpdateModule(c *gin.Context) {
	session := sessions.Default(c)
	tid := session.Get("tid").(uint64)
	moduleName := c.Param("module_name")

	sv := c.Keys["SV"].(*models.Service)
	gDB := c.Keys["gDB"].(*gorm.DB)
	var module models.ModuleS
	if err := gDB.Take(&module, "name = ? AND tenant_id = ? AND service_type = ?", moduleName, tid, sv.Type).Error; err != nil {
		utils.HTTPError(c, http.StatusNotFound, "module not found")
		return
	}

	var nmodule models.ModuleS
	if err := c.ShouldBindJSON(&nmodule); err != nil || nmodule.Valid() != nil {
		utils.HTTPError(c, http.StatusBadRequest, "failed to valid new module data")
		return
	}

	changes := []bool{
		module.ID != nmodule.ID,
		module.Info.Name != nmodule.Info.Name,
		module.Info.System != nmodule.Info.System,
		module.Info.Template != nmodule.Info.Template,
		module.Info.Version != nmodule.Info.Version,
		module.ServiceType != nmodule.ServiceType,
		module.TenantID != nmodule.TenantID,
	}
	for _, ch := range changes {
		if ch {
			utils.HTTPError(c, http.StatusInternalServerError, "failed accept system changes")
			return
		}
	}

	template := make(Template)
	nmodule.LastUpdate = time.Now()
	cfiles, err := buildModuleSConfig(&nmodule)
	if err != nil {
		utils.HTTPError(c, http.StatusInternalServerError, "failed to build module files")
		return
	}

	template["config"] = cfiles
	if err := storeModuleToGlobalS3(&nmodule.Info, template); err != nil {
		utils.HTTPError(c, http.StatusInternalServerError, "failed to update module to s3")
		return
	}

	if err := gDB.Save(&nmodule).Error; err != nil {
		utils.HTTPError(c, http.StatusInternalServerError, "failed to update module into db")
		return
	}

	utils.HTTPSuccess(c, http.StatusOK, struct{}{})
}

// DeleteModule is a function to cascade delete system module
// @Summary Delete system module from all DBs and S3 storage
// @Tags Modules
// @Produce json
// @Param module_name path string true "module name without spaces"
// @Success 200 {object} utils.successResp "system module deleted successful"
// @Failure 403 {object} utils.errorResp "deleting system module not permitted"
// @Failure 404 {object} utils.errorResp "system module not found"
// @Failure 500 {object} utils.errorResp "internal error"
// @Router /modules/{module_name} [delete]
func DeleteModule(c *gin.Context) {
	session := sessions.Default(c)
	tid := session.Get("tid").(uint64)
	moduleName := c.Param("module_name")

	sv := c.Keys["SV"].(*models.Service)
	gDB := c.Keys["gDB"].(*gorm.DB)
	var module models.ModuleS
	if err := gDB.Take(&module, "name = ? AND tenant_id = ? AND service_type = ?", moduleName, tid, sv.Type).Error; err != nil {
		utils.HTTPError(c, http.StatusNotFound, "module not found")
		return
	}

	deleteAgentModule := func(s models.Service) error {
		iDB := utils.GetDB(s.Info.DB.User, s.Info.DB.Pass, s.Info.DB.Host,
			strconv.Itoa(int(s.Info.DB.Port)), s.Info.DB.Name)
		if iDB == nil {
			return errors.New("failed to connect to instance DB")
		}
		defer iDB.Close()

		var modules []models.ModuleA
		if err := iDB.Find(&modules, "name = ?", moduleName).Error; err != nil {
			return err
		}

		if err := iDB.Delete(&modules, "name = ?", moduleName).Error; err != nil {
			return err
		}

		s3, err := storage.NewS3(
			s.Info.S3.Endpoint,
			s.Info.S3.AccessKey,
			s.Info.S3.SecretKey,
			s.Info.S3.BucketName)
		if err != nil {
			return err
		}

		if err := s3.RemoveDir("/" + moduleName + "/"); err != nil && err.Error() != "not found" {
			return err
		}

		return nil
	}

	var services []models.Service
	if err := gDB.Find(&services, "tenant_id = ? AND type = ?", tid, module.ServiceType).Error; err != nil {
		utils.HTTPError(c, http.StatusNotFound, "services not found")
		return
	}

	for _, s := range services {
		if err := deleteAgentModule(s); err != nil {
			utils.HTTPError(c, http.StatusInternalServerError, "failed to delete agent module")
			return
		}
	}

	if err := gDB.Delete(models.ModuleS{}, "name = ?", moduleName).Error; err != nil {
		utils.HTTPError(c, http.StatusInternalServerError, "failed to delete system module")
		return
	}

	s3, err := storage.NewS3()
	if err != nil {
		utils.HTTPError(c, http.StatusInternalServerError, "failed to connect to S3")
		return
	}

	if err := s3.RemoveDir("/" + moduleName + "/"); err != nil && err.Error() != "not found" {
		utils.HTTPError(c, http.StatusInternalServerError, "failed to delete system module files")
		return
	}

	utils.HTTPSuccess(c, http.StatusOK, struct{}{})
}
