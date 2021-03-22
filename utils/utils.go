package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/qor/validations"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type successResp struct {
	Status string      `json:"status" example:"success"`
	Data   interface{} `json:"data" swaggertype:"object"`
} //@name SuccessResponse

// @name ErrorResponse
type errorResp struct {
	Status string `json:"status" example:"error"`
	Msg    string `json:"msg" example:"error message text"`
} //@name ErrorResponse

// Logger is global object to use in logging subsystem
var Logger = logrus.New()

func init() {
	file, err := os.OpenFile("./logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err == nil {
		mw := io.MultiWriter(os.Stdout, file)
		Logger.Out = mw
	} else {
		log.Println("Failed to log to file, using default stderr")
	}
	Logger.Formatter = &logrus.JSONFormatter{}
}

// NoRobotResponse is struct to contain reCaptcha response
type NoRobotResponse struct {
	ChallengeTs string `json:"challenge_ts"`
	Hostname    string `json:"hostname"`
	Success     bool   `json:"success"`
}

// NoRobot is function to retrieve reCaptcha response from client request
func NoRobot(response, ip string) bool {
	apiKey := os.Getenv("RECAPTCHA_API_KEY")
	if apiKey == "" {
		return true
	}
	form := url.Values{
		"secret":   {apiKey},
		"response": {response},
		"remoteip": {ip},
	}
	body := bytes.NewBufferString(form.Encode())
	resp, err := http.Post("https://www.google.com/recaptcha/api/siteverify",
		"application/x-www-form-urlencoded", body)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	bbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	var recaptcha NoRobotResponse
	err = json.Unmarshal(bbody, &recaptcha)
	if err != nil {
		return false
	}

	if recaptcha.Success != true {
		return false
	}

	return true
}

// GetDB is function to make GORM DB connection
func GetDB(user, pass, host, port, name string) *gorm.DB {
	addr := fmt.Sprintf("%s:%s@%s/%s?parseTime=true",
		user, pass, fmt.Sprintf("tcp(%s:%s)", host, port), name)

	conn, err := gorm.Open("mysql", addr)
	if err != nil {
		log.Println("Failed open gorm connection: ", err)
		return nil
	}

	conn.LogMode(true)
	validations.RegisterCallbacks(conn)

	conn.DB().SetMaxIdleConns(10)
	conn.DB().SetMaxOpenConns(100)
	conn.DB().SetConnMaxLifetime(time.Hour)

	return conn
}

// Escape is function to filter and escape user data to valid SQL syntax
func Escape(source string) string {
	var j int = 0
	if len(source) == 0 {
		return ""
	}
	tempStr := source[:]
	desc := make([]byte, len(tempStr)*2)
	for i := 0; i < len(tempStr); i++ {
		flag := false
		var escape byte
		switch tempStr[i] {
		case '\r':
			flag = true
			escape = '\r'
			break
		case '\n':
			flag = true
			escape = '\n'
			break
		case '\\':
			flag = true
			escape = '\\'
			break
		case '\'':
			flag = true
			escape = '\''
			break
		case '"':
			flag = true
			escape = '"'
			break
		case '\032':
			flag = true
			escape = 'Z'
			break
		default:
		}
		if flag {
			desc[j] = '\\'
			desc[j+1] = escape
			j = j + 2
		} else {
			desc[j] = tempStr[i]
			j = j + 1
		}
	}
	return string(desc[0:j])
}

// EncryptPassword is function to prepare user data as a password
func EncryptPassword(password string) (hpass []byte, err error) {
	hpass, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return
}

// MakeAgentHash is function to generate agent hash from name
func MakeAgentHash(name string) string {
	currentTime := time.Now().Format("2006-01-02 15:04:05.000000000")
	salt := "1c2f46ffd8d859e69584cf6b1ab99e6043c5889c"
	hash := md5.Sum([]byte(currentTime + name + salt))
	return hex.EncodeToString(hash[:])
}

// HTTPSuccess is function as a main part of public REST API (success response)
func HTTPSuccess(c *gin.Context, code int, data interface{}) {
	c.JSON(code, gin.H{"status": "success", "data": data})
	c.Abort()
}

// HTTPError is function as a main part of public REST API (failed response)
func HTTPError(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{"status": "error", "msg": msg})
	c.Abort()
}
