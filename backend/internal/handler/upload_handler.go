package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	maxUploadSize = 2 << 20 // 2MB
	uploadsDir    = "uploads"
)

// allowedTypes 定義允許的圖片類型子目錄
var allowedTypes = map[string]string{
	"avatar": "avatars",
	"cover":  "covers",
}

// UploadHandler 圖片上傳 HTTP Handler
type UploadHandler struct{}

// NewUploadHandler 建立 UploadHandler
func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

// UploadImage POST /api/v1/upload/image — multipart/form-data 圖片上傳
func (h *UploadHandler) UploadImage(c *gin.Context) {
	// 限制 body 大小為 2MB + 一點 overhead（form fields）
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize+512)

	// 取得 type 參數
	uploadType := c.PostForm("type")
	subDir, ok := allowedTypes[uploadType]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type must be avatar or cover"})
		return
	}

	// 取得上傳的檔案
	fileHeader, err := c.FormFile("file")
	if err != nil {
		if err.Error() == "http: request body too large" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file size exceeds 2MB limit"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "file and type are required"})
		return
	}

	// 檢查檔案大小
	if fileHeader.Size > maxUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file size exceeds 2MB limit"})
		return
	}

	// 開啟檔案以讀取內容
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	defer file.Close()

	// 讀取前 512 bytes 做 MIME type 偵測（不信任 Content-Type header）
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	contentType := http.DetectContentType(buf[:n])

	// 驗證 MIME type
	var ext string
	switch contentType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "only JPEG and PNG images are allowed"})
		return
	}

	// 回到檔案開頭
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// 產生隨機 UUID 檔名
	filename := uuid.New().String() + ext

	// 確保目錄存在
	dirPath := filepath.Join(uploadsDir, subDir)
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// 寫入檔案
	destPath := filepath.Join(dirPath, filename)
	dest, err := os.Create(destPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	defer dest.Close()

	if _, err := io.Copy(dest, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// 組裝 URL
	url := fmt.Sprintf("/static/uploads/%s/%s", subDir, filename)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"url": url,
		},
	})
}

// EnsureUploadsDir 確保 uploads 目錄結構存在
func EnsureUploadsDir() error {
	for _, subDir := range allowedTypes {
		if err := os.MkdirAll(filepath.Join(uploadsDir, subDir), 0o755); err != nil {
			return fmt.Errorf("建立上傳目錄失敗 %s: %w", subDir, err)
		}
	}
	return nil
}

// StaticUploadsPath 取得 uploads 目錄的絕對路徑（供 gin.Static 使用）
func StaticUploadsPath() string {
	return uploadsDir
}

// SanitizeFilename 清理檔名中的路徑穿越字元（雖然我們用 UUID 不會有此問題，但多一層防護）
func SanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, "..", "")
	return name
}
