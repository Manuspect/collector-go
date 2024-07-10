package controllers

import (
	"context"
	"fmt"

	logFi "github.com/gofiber/fiber/v2/log"

	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/minio/minio-go/v7"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/oklog/ulid/v2"
)

// @Summary		UploadFile return file with bucket name
// @Description	UploadFile retrieves the value of the environment variable named by the key and return fileUpload
// @Tags		Upload
// @Accept		json
// @Produce		json
// @Param		fileUpload		formData	file					true	"path to upload formData file"
// @Success		200				{object}	entities.InfoFile
// @Failure		400				{object}	entities.ServerError
// @Failure		500				{object}	entities.ServerError
// @Router		/api/v1/upload	[post]
func UploadFile(m *minio.Client, js jetstream.JetStream) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		bucketName := os.Getenv("MINIO_BUCKET")
		file, err := c.FormFile("fileUpload")
		if err != nil {
			logFi.Error("UploadFile")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}

		buffer, err := file.Open()
		if err != nil {
			logFi.Error("UploadFile")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}
		defer buffer.Close()

		objectName := fmt.Sprintf("%s-%s", ulid.Make(), file.Filename)
		fileBuffer := buffer
		contentType := file.Header["Content-Type"][0]
		fileSize := file.Size

		info, err := m.PutObject(
			ctx,
			bucketName,
			objectName,
			fileBuffer,
			fileSize,
			minio.PutObjectOptions{ContentType: contentType},
		)
		if err != nil {
			logFi.Error("UploadFile")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}

		js.Publish(
			c.Context(),
			"BATCH.MAIN",
			[]byte(fmt.Sprintf(
				"{objectName: %s, bucketName: %s}",
				objectName,
				bucketName)),
		)

		logFi.Infof("Successfully uploaded %s of size %d\n", objectName, info.Size)

		return c.Status(200).JSON(fiber.Map{
			"error": false,
			"msg":   nil,
			"info":  info,
		})
	}
}
