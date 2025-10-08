package controllers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"mobile-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UploadController struct {
	uploadPath string
}

func NewUploadController(uploadPath string) *UploadController {
	return &UploadController{uploadPath: uploadPath}
}

type UploadResponse struct {
	Filename string `json:"filename"`
	URL      string `json:"url"`
	Size     int64  `json:"size"`
	Type     string `json:"type"`
}

// UploadFile godoc
// @Summary Upload a single file
// @Description Upload a single file to the server
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "File to upload"
// @Success 200 {object} utils.SuccessResponse{data=UploadResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/upload [post]
func (uc *UploadController) UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.SendErrorResponse(c, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type
	if !uc.isValidFileType(header) {
		utils.SendErrorResponse(c, "Invalid file type", http.StatusBadRequest)
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Create upload directory if not exists
	if err := os.MkdirAll(uc.uploadPath, 0755); err != nil {
		utils.SendInternalServerErrorResponse(c, "Failed to create upload directory")
		return
	}

	// Save file
	filepath := filepath.Join(uc.uploadPath, filename)
	if err := uc.saveFile(file, filepath); err != nil {
		utils.SendInternalServerErrorResponse(c, "Failed to save file")
		return
	}

	response := UploadResponse{
		Filename: filename,
		URL:      fmt.Sprintf("/uploads/%s", filename),
		Size:     header.Size,
		Type:     header.Header.Get("Content-Type"),
	}

	utils.SendSuccessResponse(c, response, "File uploaded successfully")
}

func (uc *UploadController) UploadMultipleFiles(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		utils.SendErrorResponse(c, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		utils.SendErrorResponse(c, "No files uploaded", http.StatusBadRequest)
		return
	}

	var responses []UploadResponse
	for _, header := range files {
		// Validate file type
		if !uc.isValidFileType(header) {
			continue // Skip invalid files
		}

		// Generate unique filename
		ext := filepath.Ext(header.Filename)
		filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

		// Create upload directory if not exists
		if err := os.MkdirAll(uc.uploadPath, 0755); err != nil {
			continue
		}

		// Save file
		filepath := filepath.Join(uc.uploadPath, filename)
		file, err := header.Open()
		if err != nil {
			continue
		}

		if err := uc.saveFile(file, filepath); err != nil {
			file.Close()
			continue
		}
		file.Close()

		responses = append(responses, UploadResponse{
			Filename: filename,
			URL:      fmt.Sprintf("/uploads/%s", filename),
			Size:     header.Size,
			Type:     header.Header.Get("Content-Type"),
		})
	}

	utils.SendSuccessResponse(c, responses, "Files uploaded successfully")
}

func (uc *UploadController) GetFile(c *gin.Context) {
	filename := c.Param("filename")
	filepath := filepath.Join(uc.uploadPath, filename)

	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		utils.SendNotFoundResponse(c, "File not found")
		return
	}

	c.File(filepath)
}

func (uc *UploadController) DeleteFile(c *gin.Context) {
	filename := c.Param("filename")
	filepath := filepath.Join(uc.uploadPath, filename)

	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		utils.SendNotFoundResponse(c, "File not found")
		return
	}

	// Delete file
	if err := os.Remove(filepath); err != nil {
		utils.SendInternalServerErrorResponse(c, "Failed to delete file")
		return
	}

	utils.SendSuccessResponse(c, nil, "File deleted successfully")
}

func (uc *UploadController) isValidFileType(header *multipart.FileHeader) bool {
	allowedTypes := []string{
		"image/jpeg", "image/png", "image/gif", "image/webp",
		"application/pdf", "text/plain", "application/json",
		"application/zip", "application/x-zip-compressed",
	}
	contentType := header.Header.Get("Content-Type")

	for _, allowedType := range allowedTypes {
		if contentType == allowedType {
			return true
		}
	}
	return false
}

func (uc *UploadController) saveFile(src multipart.File, dst string) error {
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}
