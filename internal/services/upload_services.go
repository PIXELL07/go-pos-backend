package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/prayosha/go-pos-backend/internal/models"
	"gorm.io/gorm"
)

type UploadService struct {
	db        *gorm.DB
	uploadDir string
	baseURL   string
}

func NewUploadService(db *gorm.DB, uploadDir, baseURL string) *UploadService {
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	_ = os.MkdirAll(uploadDir, 0755)
	return &UploadService{db: db, uploadDir: uploadDir, baseURL: baseURL}
}

const maxUploadSize = 5 << 20 // 5 MB

var allowedMimes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

func (s *UploadService) Save(
	file multipart.File,
	header *multipart.FileHeader,
	ownerType string,
	ownerID *uuid.UUID,
) (*models.Upload, error) {
	if header.Size > maxUploadSize {
		return nil, fmt.Errorf("file too large (max 5 MB)")
	}

	// Detect MIME
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	_, _ = file.Seek(0, io.SeekStart)
	mime := detectMIME(buf[:n])
	ext, ok := allowedMimes[mime]
	if !ok {
		return nil, fmt.Errorf("unsupported file type: %s", mime)
	}

	// Build unique filename
	newName := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().UnixMilli(), ext)
	subDir := ownerType
	if subDir == "" {
		subDir = "misc"
	}
	dir := filepath.Join(s.uploadDir, subDir)
	_ = os.MkdirAll(dir, 0755)
	fullPath := filepath.Join(dir, newName)

	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	relPath := strings.TrimPrefix(fullPath, s.uploadDir)
	publicURL := fmt.Sprintf("%s/static%s", strings.TrimRight(s.baseURL, "/"), relPath)

	upload := &models.Upload{
		OwnerType: ownerType,
		FileName:  header.Filename,
		MimeType:  mime,
		SizeBytes: header.Size,
		URL:       publicURL,
		StorePath: fullPath,
	}
	if ownerID != nil {
		upload.OwnerID = *ownerID
	}

	if err := s.db.Create(upload).Error; err != nil {
		return nil, err
	}
	return upload, nil
}

func detectMIME(b []byte) string {
	if len(b) >= 3 && b[0] == 0xFF && b[1] == 0xD8 {
		return "image/jpeg"
	}
	if len(b) >= 8 && string(b[:8]) == "\x89PNG\r\n\x1a\n" {
		return "image/png"
	}
	if len(b) >= 4 && string(b[:4]) == "GIF8" {
		return "image/gif"
	}
	if len(b) >= 4 && string(b[:4]) == "RIFF" {
		return "image/webp"
	}
	return "application/octet-stream"
}
