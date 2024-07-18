package controllers

import (
	"context"
	"fmt"
	"os"

	logFi "github.com/gofiber/fiber/v2/log"

	"github.com/gofiber/fiber/v2"
	"github.com/minio/minio-go/v7"
	"github.com/nats-io/nats.go/jetstream"
)

// @Summary		UploadFile return file with bucket name
// @Description	UploadFile retrieves the value of the environment variable named by the key and return fileUpload
// @Tags		Upload
// @Accept		json
// @Produce		json
// @Param		fileUpload		formData	file					true	"path to upload formData file"
// @Param		userId		formData	string					true	"id of user"
// @Param		recordId		formData	string					true	"id of record"
// @Param		timestamp		formData	string					true	"timestamp of batch"
// @Success		200				{object}	entities.InfoFile
// @Failure		400				{object}	entities.ServerError
// @Failure		500				{object}	entities.ServerError
// @Router		/upload	[post]
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

		form, err := c.MultipartForm()
		if err != nil {
			logFi.Error("MultipartForm")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}

		userId, ok := form.Value["userId"]
		if !ok || len(userId) < 1 {
			logFi.Error("userId")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": true,
				"msg":   "field userId required",
			})
		}

		recordId, ok := form.Value["recordId"]
		if !ok || len(recordId) < 1 {
			logFi.Error("recordId")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": true,
				"msg":   "field recordId required",
			})
		}

		timestamp, ok := form.Value["timestamp"]
		if !ok || len(timestamp) < 1 {
			logFi.Error("timestamp")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": true,
				"msg":   "field timestamp required",
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

		objectName := fmt.Sprintf("%s-%s-%s-%s", userId[0], recordId[0], timestamp[0], file.Filename)
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
				"{\"objectName\": \"%s\", \"bucketName\": \"%s\"}",
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
