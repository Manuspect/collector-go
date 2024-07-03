package main

import (
	"collector-go/controllers"
	"collector-go/internal/database"
	nats "collector-go/internal/queue"
	"collector-go/internal/service"
	databasesqlc "collector-go/internal/sqlc"
	"collector-go/internal/storage"
	"strconv"

	"fmt"
	"log"

	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/goccy/go-json"
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/joho/godotenv"
)

const DateTime = "2006-01-02 15:04:03"

var Version = "1.0.0"
var BuildDate = time.Now().Format(DateTime)

//	@Title			Collector
//	@Version		Version 1.0
//	@Description	API server for Collector Application

//	@BasePath	/api_v1/

//	@securityDefinitions.ApiKey	JWT
//	@in							header
//	@name						Authorization

func main() {
	godotenv.Load()

	db, err := database.NewConnect(
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	if err != nil {
		log.Fatal("no connection to DB", err)
	}
	queries := databasesqlc.New(db)

	jetStream, err := nats.NatsConnect()
	if err != nil {
		log.Fatalln(err)
	}

	minio_client := storage.MinioConnect()

	redis_db, err := strconv.Atoi(os.Getenv("REDIS_DB_NAME"))
	if err != nil {
		log.Println("error: convert string to int for redis_db, location cmd/main")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       redis_db,
	})

	app := fiber.New(fiber.Config{
		AppName:                      fmt.Sprintf("Ver: %s BuildDate: %s", Version, BuildDate),
		JSONEncoder:                  json.Marshal,
		JSONDecoder:                  json.Unmarshal,
		DisablePreParseMultipartForm: true,
		StreamRequestBody:            true,
	})

	app.Use(swagger.New(swagger.Config{
		BasePath: os.Getenv("SWAGGER_BASE_PATH"),
		FilePath: os.Getenv("SWAGGER_FILE_PATH"),
		Path:     "docs",
	}))

	app.Use(logger.New())
	app.Use(helmet.New())

	v1 := app.Group("/api_v1")

	v1.Post("/registration", service.CreateUser(queries))
	v1.Post("/login", service.CheckLogin(queries))
	v1.Get("/user/:id?", service.GetUserById(queries))
	v1.Get("/users", service.GetUsers(queries))

	app.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(os.Getenv("JWT_KEY_ACCESS"))},
	}))

	v1.Delete("/user", service.DeleteUserById(queries))
	v1.Patch("/user", service.UpdateUserById(queries))
	v1.Patch("/change_password", service.EditPasswordUser(queries))
	v1.Patch("/change_password_old", service.EditPasswordUserByOld(queries))
	v1.Post("/user/new_pass", service.CreateRestorePasswordLink(queries, client))
	v1.Patch("/user/change_pass", service.ChangePasswordByLink(queries, client))

	app.Post("/api/v1/upload", controllers.UploadFile(minio_client, jetStream))

	app.Get("/metrics", monitor.New())

	app.Listen(fmt.Sprintf("%s:%s", os.Getenv("APP_HOST"), os.Getenv("APP_PORT")))
}
