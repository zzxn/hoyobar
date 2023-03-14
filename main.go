package main

import (
	"fmt"
	"hoyobar/conf"
	"hoyobar/handler"
	"hoyobar/middleware"
	"hoyobar/model"
	"hoyobar/service"
	"hoyobar/util/idgen"
	"hoyobar/util/mycache"
	"hoyobar/util/myerr"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	startApp(readConfig())
}

func readConfig() conf.Config {
	filePath := "config.yaml"
	log.Println("read config from", filePath)
	r, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("fails to open config file %s, err: %v\n", filePath, err)
	}
	config := conf.FromYAML(r)
	conf.Global = &config
	return config
}

func startApp(config conf.Config) {
	idgen.Init("1970-01-01", 0)

	db := initDB(config)
	model.Init(db)

	cache := mycache.NewMemoryCache()

	r := gin.Default()
	api := r.Group("/api")
	api.Use(middleware.ErrorHandler())

	var (
		userHandler handler.Handler
		postHandler handler.Handler
	)

	// user API
	userService := service.NewUserService(cache)
	api.Use(middleware.ReadAuthToken(func(authToken string, c *gin.Context) {
		userID, err := userService.AuthTokenToUserID(authToken)
		if err != nil {
			if e, ok := err.(*myerr.MyError); ok {
				log.Println("fails to read user ID from auth token, cause:", e.Cause())
			} else {
				log.Println("fails to read user ID from auth token, err:", err)
			}
			c.Set("auth_err", err)
			return
		}
		c.Set("user_id", userID)
	}))
	userHandler = &handler.UserHandler{UserService: userService} // must be pointer, why?
	userHandler.AddRoute(api.Group("/user"))

	// post API
	postService := service.NewPostService(cache)
	postHandler = &handler.PostHandler{PostService: postService}
	postHandler.AddRoute(api.Group("/post"))

	r.Run(fmt.Sprintf(":%v", config.App.Port))
}

func initSqlite3(config conf.Config) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(config.DB.Sqlite3.DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("fails to connect sqlite db: %v\n", err)
	}
	return db
}

func initMySQL(config conf.Config) *gorm.DB {
	var err error
	c := config.DB.MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Pass, c.Host, c.Port, c.Host)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("fails to connect database %q, err=%v\n", dsn, err)
	}
	return db
}

func initDB(config conf.Config) *gorm.DB {
	log.Printf("connect db with type %v \n", config.DB.Type)
	var db *gorm.DB
	switch config.DB.Type {
	case "mysql":
		db = initMySQL(config)
	case "sqlite3":
		db = initSqlite3(config)
	}
	if db == nil {
		log.Fatalln("not recoginize db type:", config.DB.Type)
	}
	// TODO: do we need to do this?
	db.Debug().AutoMigrate(&model.User{}, &model.Post{}, &model.PostStat{})
	return db
}
