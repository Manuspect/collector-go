package main

import (
	"collector-go/controllers"
	nats "collector-go/internal/queue"
	"collector-go/internal/storage"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/joho/godotenv"
)

const DateTime = "2006-01-02 15:04:03"

var Version = "1.0.0"
var BuildDate = time.Now().Format(DateTime)

func main() {
	godotenv.Load()

	jetStream, err := nats.NatsConnect()
	if err != nil {
		log.Fatalln(err)
	}

	minio_client := storage.MinioConnect()

	app := fiber.New(fiber.Config{
		AppName:                      fmt.Sprintf("Ver: %s BuildDate: %s", Version, BuildDate),
		JSONEncoder:                  json.Marshal,
		JSONDecoder:                  json.Unmarshal,
		DisablePreParseMultipartForm: true,
		StreamRequestBody:            true,
	})

	app.Use(logger.New())
	app.Use(helmet.New())

	app.Post("/api/v1/upload", controllers.UploadFile(minio_client, jetStream))

	app.Get("/metrics", monitor.New())

	app.Listen(fmt.Sprintf("%s:%s", os.Getenv("APP_HOST"), os.Getenv("APP_PORT")))
}
