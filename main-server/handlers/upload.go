package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
)

type UploadHandler struct {
	uploadDir string
}

func NewUploadHandler(uploadDir string) *UploadHandler {
	return &UploadHandler{
		uploadDir: uploadDir,
	}
}

func (h *UploadHandler) Upload(c echo.Context) error {
	// Get the file from form
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "No file provided",
		})
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to open file",
		})
	}
	defer src.Close()

	// Create destination file path
	filename := file.Filename
	dst := filepath.Join(h.uploadDir, filename)

	// Create the destination file
	out, err := os.Create(dst)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create file",
		})
	}
	defer out.Close()

	// Copy file contents
	_, err = io.Copy(out, src)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to save file",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":  "File uploaded successfully",
		"filename": filename,
		"url":      fmt.Sprintf("/uploads/%s", filename),
	})
}

func (h *UploadHandler) Serve(c echo.Context) error {
	// Get filename from URL parameter
	filename := c.Param("*")

	// Clean the filename to prevent directory traversal
	filename = filepath.Clean(filename)
	if strings.Contains(filename, "..") {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid filename",
		})
	}

	// Build full file path
	filePath := filepath.Join(h.uploadDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "File not found",
		})
	}

	// Serve the file
	return c.File(filePath)
}
