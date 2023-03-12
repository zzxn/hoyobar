package main

import (
	"fmt"
	"hoyobar/conf"
	"hoyobar/handler"
	"hoyobar/model"
	"hoyobar/service"
	"hoyobar/util/idgen"
	"hoyobar/util/middleware"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
    config := readConfig() 
    initApp(config)
}

func readConfig() conf.Config {
    filePath := "config.yaml"
    log.Println("read config from", filePath)
    r, err := os.Open(filePath)
    if err != nil {
        log.Fatalf("fails to open config file %s, err: %v\n", filePath, err)
    }
    config := conf.FromYAML(r)
    return config
}

func initApp(config conf.Config) {
    idgen.Init("1970-01-01", 0)
    db := initDB(config)
    model.Init(db)

    r := gin.Default()
    api := r.Group("/api")
    api.Use(middleware.Auth())

    var (
        userHandler handler.Handler
    )

    // user API
    userService := service.UserService{}    
    userHandler = &handler.UserHandler{UserService: &userService} // must be pointer, why?
    userHandler.AddRoute(api.Group("/user"))
    
    r.Run(":8080")
}

func initTestDB() *gorm.DB {
    db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
    if err != nil {
        log.Fatalf("fails to connect sqlite db: %v\n", err)
    }
    db.Debug().AutoMigrate(&model.User{})
    return db
}

func initDB(config conf.Config) *gorm.DB {
    log.Println("connect db...")
    if true {
        // use test config
        return initTestDB()
    }
    var err error
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", 
        config.DB.User, config.DB.Pass, config.DB.Host, config.DB.Port, config.DB.DBName)
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("fails to connect database %q, err=%v\n", dsn, err)
    }
    return db
}
