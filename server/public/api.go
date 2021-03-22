package public

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/vxcontrol/vxui/utils"
)

// DownloadAgent is a function to return agent binary file
// @Summary Retrieve agent binary file by OS and arch
// @Tags Public,Downloads
// @Produce octet-stream,json
// @Param os path string true "agent info OS" default(linux) Enums(windows, linux, darwin)
// @Param arch path string true "agent info arch" default(amd64) Enums(386, amd64)
// @Success 200 {file} file "agent binary as a file"
// @Failure 400 {object} utils.errorResp "invalid agent info"
// @Failure 403 {object} utils.errorResp "getting agent binary file not permitted"
// @Failure 404 {object} utils.errorResp "agent binary file not found"
// @Failure 500 {object} utils.errorResp "internal error"
// @Router /downloads/vxagent/{os}/{arch} [get]
func DownloadAgent(c *gin.Context) {
	validate := validator.New()
	agentOS := c.Param("os")
	agentArch := c.Param("arch")
	agentName := "vxagent"
	if agentOS == "windows" {
		agentName += ".exe"
	}

	if err := validate.Var(agentOS, "oneof=windows linux darwin,required"); err != nil {
		utils.HTTPError(c, http.StatusBadRequest, "failed to valid agent os")
		return
	}
	if err := validate.Var(agentArch, "oneof=386 amd64,required"); err != nil {
		utils.HTTPError(c, http.StatusBadRequest, "failed to valid agent arch")
		return
	}

	agentPath := filepath.Join("binaries", agentOS, agentArch, agentName)
	if _, err := os.Stat(agentPath); os.IsNotExist(err) {
		utils.HTTPError(c, http.StatusNotFound, "agent binary file not found")
		return
	} else if err != nil {
		utils.HTTPError(c, http.StatusInternalServerError, "failed to download agentt binary file")
		return
	}

	c.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%q", agentName))
	c.File(agentPath)
}
