package main

import (
	"fmt"
	"hoyobar/conf"
	"hoyobar/handler"
	"hoyobar/middleware"
	"hoyobar/model"
	"hoyobar/service"
	"hoyobar/storage"
	"hoyobar/util/idgen"
	"hoyobar/util/mycache"
	"hoyobar/util/myerr"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	// "github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	rand.Seed(time.Now().Unix())
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
	fmt.Printf("config: %#v\n", config)
	return config
}

func startApp(config conf.Config) {
	idgen.Init("2020-01-01", 0)

	db := initDB(config)
	if conf.Global.DB.AutoMigrate {
		model.Migrate(db)
	}

	cache := initCache(config)

	r := gin.Default()
	r.ContextWithFallback = true
	// pprof.Register(r)
	r.Use(cors.Default())
	api := r.Group("/api")
	api.Use(
		middleware.ErrorHandler(),
		middleware.Timeout(conf.Global.App.Timeout.Default),
	)

	var (
		userHandler handler.Handler
		postHandler handler.Handler
	)

	userStorage := storage.NewUserStorageMySQL(db)
	postStorage := storage.NewPostStorageMySQL(db)
	replyStorage := storage.NewPostReplyStorageMySQL(db)

	// user API
	userService := service.NewUserService(cache, userStorage)
	api.Use(middleware.ReadAuthToken(func(authToken string, c *gin.Context) {
		log.Println("found auth token, checking user")
		userID, err := userService.AuthTokenToUserID(c, authToken)
		// check timeout
		select {
		case <-c.Done():
			log.Println("timeout during check user auth token")
			c.Error(myerr.ErrTimeout.WithCause(c.Err())) // nolint:errcheck
			c.Abort()
			return
		default: // do nothing
		}
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
	postService := service.NewPostService(cache, userStorage, postStorage, replyStorage)
	postHandler = &handler.PostHandler{
		PostService: postService,
		UserService: userService,
	}
	postHandler.AddRoute(api.Group("/post"))

	err := r.Run(fmt.Sprintf(":%v", config.App.Port))
	if err != nil {
		log.Fatalf("app exit with err: %v\n", err)
	}
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
		c.User, c.Pass, c.Host, c.Port, c.DBName)
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
	return db
}

func initCache(config conf.Config) mycache.Cache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.Global.Redis.Addr,
		Username: conf.Global.Redis.Username,
		Password: conf.Global.Redis.Password,
	})
	return mycache.NewRedisCache(rdb)
}
