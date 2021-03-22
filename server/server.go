package server

import (
	"encoding/gob"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/vxcontrol/vxui/models"
	"github.com/vxcontrol/vxui/server/private"
	"github.com/vxcontrol/vxui/server/public"
	"github.com/vxcontrol/vxui/utils"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/jinzhu/gorm/dialects/mysql" // GORM backend
	_ "github.com/vxcontrol/vxui/docs"        // swagger docs
)

func index(c *gin.Context) {
	data, err := ioutil.ReadFile("./static/index.html")
	if err != nil {
		log.Println(err)
		return
	}
	c.Data(200, "text/html", []byte(data))
}

// Run is a main function to start vxui server logic
func Run() {
	gob.Register(map[string]interface{}{})

	f, err := os.OpenFile("./logs/api.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	store := cookie.NewStore([]byte("secret-store-session"))

	gin.SetMode(gin.DebugMode)
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(sessions.Sessions("auth", store))

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/app/")
	})

	// static files
	router.StaticFile("/favicon.png", "./static/favicon.png")
	router.StaticFile("/apple-touch-icon.png", "./static/apple-touch-icon.png")
	router.Static("/js", "./static/js")
	router.Static("/css", "./static/css")
	router.Static("/fonts", "./static/fonts")
	router.Static("/images", "./static/images")

	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.String(), "/app/") {
			index(c)
		}
	})

	// set api handlers
	api := router.Group("/api/v1")
	api.Use(setGlobalDB())
	publicGroup := api.Group("/")
	{
		publicGroup.GET("/info", public.Info)
		publicGroup.POST("/signin", public.SignIn)
		publicGroup.POST("/signup", public.SignUp)

		downloadsGroup := publicGroup.Group("/downloads")
		{
			downloadsGroup.GET("/vxagent/:os/:arch", public.DownloadAgent)
		}
	}

	privateGroup := api.Group("/")
	privateGroup.Use(authRequired())
	privateGroup.Use(setServiceInfo())
	{
		privateGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

		usersGroup := privateGroup.Group("/users")
		{
			usersGroup.GET("/current", private.GetCurrentUser)
			usersGroup.POST("/current/password", private.ChangePasswordCurrentUser)
		}

		agentsGroup := privateGroup.Group("/agents")
		{
			agentsGroup.GET("/", private.GetAgents)
			agentsGroup.GET("/:hash", private.GetAgent)
			agentsGroup.POST("/:hash", private.PatchAgent)
			agentsGroup.GET("/:hash/modules", private.GetAgentModules)
			agentsGroup.GET("/:hash/modules/:module_name", private.GetAgentModule)
			agentsGroup.POST("/:hash/modules/:module_name", private.PatchAgentModule)
			agentsGroup.GET("/:hash/modules/:module_name/bmodule.vue", private.GetAgentBModule)
			agentsGroup.PUT("/", private.CreateAgent)
			agentsGroup.DELETE("/:hash", private.DeleteAgent)
		}

		eventsGroup := privateGroup.Group("/events")
		{
			eventsGroup.GET("/", private.GetEvents)
		}

		modulesGroup := privateGroup.Group("/modules")
		{
			modulesGroup.GET("/", private.GetModules)
			modulesGroup.GET("/:module_name", private.GetModule)
			modulesGroup.GET("/:module_name/options/:option_name", private.GetModuleOption)
		}

		modulesAdminGroup := privateGroup.Group("/modules")
		modulesAdminGroup.Use(adminRequired())
		{
			modulesAdminGroup.PUT("/", private.CreateModule)
			modulesAdminGroup.POST("/:module_name", private.UpdateModule)
			modulesAdminGroup.DELETE("/:module_name", private.DeleteModule)
			modulesAdminGroup.GET("/:module_name/files", private.GetModuleFiles)
			modulesAdminGroup.GET("/:module_name/files/file", private.GetModuleFile)
			modulesAdminGroup.POST("/:module_name/files/file", private.PatchModuleFile)
		}
	}

	router.Run(":8080")
}

func authRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		uid := session.Get("uid")
		gid := session.Get("gid")
		sid := session.Get("sid")
		tid := session.Get("tid")
		if uid == nil || gid == nil || sid == nil || tid == nil {
			utils.HTTPError(c, http.StatusForbidden, "auth required")
			return
		}
		c.Next()
	}
}

func adminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		gid, ok := session.Get("gid").(uint64)
		if !ok || (gid != 0 && gid != 1) {
			utils.HTTPError(c, http.StatusForbidden, "admin required")
			return
		}
		c.Next()
	}
}

func setGlobalDB() gin.HandlerFunc {
	gDB := utils.GetDB(os.Getenv("DB_USER"), os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
	if gDB == nil {
		panic("can't connect to global DB")
	}

	return func(c *gin.Context) {
		c.Set("gDB", gDB)
		c.Next()
	}
}

func setServiceInfo() gin.HandlerFunc {
	var mx sync.Mutex
	mDB := make(map[uint64]*gorm.DB, 0)
	mSV := make(map[uint64]*models.Service, 0)
	getInstanceDB := func(c *gin.Context) (*gorm.DB, *models.Service) {
		mx.Lock()
		defer mx.Unlock()

		session := sessions.Default(c)
		sid, ok := session.Get("sid").(uint64)
		if !ok || sid == 0 {
			return nil, nil
		}

		gDB := c.Keys["gDB"].(*gorm.DB)
		if iDB, ok := mDB[sid]; !ok {
			var s models.Service
			if err := gDB.Take(&s, "id = ?", sid).Error; err != nil {
				return nil, nil
			}

			iDB = utils.GetDB(s.Info.DB.User, s.Info.DB.Pass, s.Info.DB.Host,
				strconv.Itoa(int(s.Info.DB.Port)), s.Info.DB.Name)
			if iDB != nil {
				mDB[sid] = iDB
				mSV[sid] = &s
				return iDB, &s
			}
		} else {
			if sv, ok := mSV[sid]; ok {
				return iDB, sv
			}
		}

		return nil, nil
	}

	return func(c *gin.Context) {
		iDB, sv := getInstanceDB(c)
		c.Set("iDB", iDB)
		c.Set("SV", sv)
		c.Next()
	}
}
