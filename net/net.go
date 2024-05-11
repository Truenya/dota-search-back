package net

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Truenya/dota-search-back/cache"
	"github.com/Truenya/dota-search-back/data"
	"github.com/Truenya/dota-search-back/db"
	"github.com/Truenya/dota-search-back/handlers"
	"github.com/gin-gonic/gin"

	grl "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-contrib/cors"
	log "github.com/sirupsen/logrus"
)

func SendOk(c *gin.Context, some interface{}) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("Access-Control-Request-Method", "POST")
	c.JSON(http.StatusOK, some)
}

func SendStatus(c *gin.Context, code int, err interface{}) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("Access-Control-Request-Method", "POST")
	c.JSON(code, err)
}

func SendNoContent(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("Access-Control-Request-Method", "POST")
	c.Writer.WriteHeader(http.StatusNoContent)
}

func GetUser(c *gin.Context) {
	id := c.Param("id")

	p, ok := cache.GetU(id)
	if !ok {
		SendStatus(c, http.StatusNotFound, `{"msg": "User not found `+id+`"}`)
		return
	}

	SendOk(c, p)
}

func GetCommand(c *gin.Context) {
	id := c.Param("id")

	com, ok := cache.GetC(id)
	if !ok {
		SendStatus(c, http.StatusNotFound, `{"msg": "Command not found `+id+`"}`)
		return
	}

	SendOk(c, com)
}

func GetMessage(c *gin.Context) {
	id := c.Param("id")

	m, ok := cache.GetM(id)
	if !ok {
		SendStatus(c, http.StatusNotFound, `{"msg": "Message not found `+id+`"}`)
		return
	}

	SendOk(c, m)
}

func GetUsers(c *gin.Context) {
	au := cache.GetAllU()
	if len(au) == 0 {
		SendNoContent(c)
		return
	}

	log.Debugf("Sending player list: %v", au)

	SendOk(c, au)
}

func GetCommands(c *gin.Context) {
	ac := cache.GetAllC()
	if len(ac) == 0 {
		SendStatus(c, http.StatusNoContent, `{"msg": "Commands not found"}`)
		return
	}

	SendOk(c, ac)
}

func GetMessages(c *gin.Context) {
	am := cache.GetAllM()
	if len(am) == 0 {
		log.Debugln("No messages stored in cache")
		log.Debugln(am)
		SendStatus(c, http.StatusNoContent, `{"msg": "Commands not found"}`)

		return
	}

	if c.Query("all") == "true" {
		log.Debugln("Sending all")
		log.Debugln(am)
		SendOk(c, am)

		return
	}

	for id, mes := range am {
		if time.Since(mes.Timestamp) >= time.Hour*72 {
			delete(am, id)
		}
	}

	SendOk(c, am)
}

func AddUser(c *gin.Context) {
	var p data.Player
	if err := json.NewDecoder(c.Request.Body).Decode(&p); err != nil {
		log.Debug("Failed to convert to player ", p)
		SendStatus(c, http.StatusBadRequest, fmt.Sprintf(`{"msg": "Failed to convert to player, error is: %s"}`, err))

		return
	}

	p.IP = c.ClientIP()

	action := handlers.AddPlayer(p)
	switch action {
	case handlers.TooManyRequests:
		SendStatus(c, http.StatusTooManyRequests, `{"msg": "Wait for 30s before storing data again"}`)
		return
	case handlers.Updated:
		SendStatus(c, http.StatusOK, `{"msg": "Updated"}`)
		return
	case handlers.DeletedAndAdded:
		fallthrough
	case handlers.Deleted:
		fallthrough
	case handlers.NotFound:
		fallthrough
	case handlers.JustAdded:
		SendStatus(c, http.StatusCreated, `{"msg": "created"}`)
		return
	}
}

func AddCommand(c *gin.Context) {
	var com data.Command
	if json.NewDecoder(c.Request.Body).Decode(&com) != nil {
		SendStatus(c, http.StatusBadRequest, `{"msg": "Failed to convert payload to Command"}`)
		return
	}

	com.IP = c.ClientIP()
	action := handlers.AddCommand(com)

	switch action {
	case handlers.TooManyRequests:
		SendStatus(c, http.StatusTooManyRequests, `{"msg": "Wait for 30s before storing data again"}`)
		return
	case handlers.Updated:
		SendStatus(c, http.StatusOK, `{"msg": "Updated"}`)
		return
	case handlers.DeletedAndAdded:
		fallthrough
	case handlers.Deleted:
		fallthrough
	case handlers.NotFound:
		fallthrough
	case handlers.JustAdded:
		SendStatus(c, http.StatusCreated, `{"msg": "created"}`)
		return
	}
}

func AddMessage(c *gin.Context) {
	var m data.Message
	if json.NewDecoder(c.Request.Body).Decode(&m) != nil {
		SendStatus(c, http.StatusBadRequest, `{"msg": "Failed to convert payload to Message"}`)
		return
	}

	m.CalculateHash()

	if cache.StoreM(m) == cache.Conflict {
		c.JSON(http.StatusConflict, []byte(`{"msg": "Such entity already exists."}`))
		return
	}

	db.StoreM(m)
	c.JSON(http.StatusOK, []byte("ok"))
}

func keyFunc(c *gin.Context) string {
	return c.ClientIP()
}

func errorHandler(c *gin.Context, info grl.Info) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("Access-Control-Request-Method", "POST, GET")
	c.String(429, "Too many requests. Try again in "+time.Until(info.ResetTime).String())
}

func RouteAll() {
	store := grl.InMemoryStore(&grl.InMemoryOptions{
		Rate:  time.Second,
		Limit: 5,
	})

	mw := grl.RateLimiter(store, &grl.Options{
		ErrorHandler: errorHandler,
		KeyFunc:      keyFunc,
	})

	router := gin.Default()
	// CORS for https://foo.com and https://github.com origins, allowing:
	// - PUT and PATCH methods
	// - Origin header
	// - Credentials share
	// - Preflight requests cached for 12 hours
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		AllowOriginFunc: func(_ string) bool {
			return true
		},
		MaxAge: 12 * time.Hour,
	}))
	router.GET("/player", mw, GetUsers)
	router.GET("/command", mw, GetCommands)
	router.GET("/message", mw, GetMessages)
	router.POST("/player", mw, AddUser)
	router.POST("/command", mw, AddCommand)
	router.GET("/player/:id", mw, GetUser)
	router.GET("/command/:id", mw, GetCommand)
	router.GET("/message/:id", mw, GetMessage)
	router.GET("/ws", mw, WsHandler)
	log.Error(router.Run(":322"))
}
