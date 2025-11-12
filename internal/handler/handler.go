package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Caritas-Team/reviewer/internal/config"
	"github.com/Caritas-Team/reviewer/internal/memecached"
	mc "github.com/Caritas-Team/reviewer/internal/memecached"
	m "github.com/Caritas-Team/reviewer/internal/model"
	"github.com/Caritas-Team/reviewer/internal/usecase/file"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/google/uuid"
)

type Handler struct {
	cfg    config.Config
	cache  mc.CacheInterface
	parser *file.PDFParser
}

func NewHandler(cfg config.Config, cache *memecached.Cache) *Handler {
	return &Handler{
		cfg:    cfg,
		cache:  cache,
		parser: file.NewPDFParser(),
	}
}

// UploadHandler обрабатывает POST /upload
func (h *Handler) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed", "METHOD_NOT_ALLOWED")
		return
	}

	// Проверка заголовка X-Operation-Key
	operationKey := r.Header.Get("X-Operation-Key")
	if operationKey == "" {
		writeError(w, http.StatusBadRequest, "X-Operation-Key header is required", "MISSING_OPERATION_KEY")
		return
	}

	// Проверка идемпотентности
	ctx := r.Context()
	existingOpID, err := h.checkIdempotency(ctx, operationKey)
	if err == nil && existingOpID != "" {
		writeJSON(w, http.StatusConflict, m.DuplicateOperationError{
			Error:               "Operation with this key already exists",
			Code:                "DUPLICATE_OPERATION_KEY",
			ExistingOperationID: existingOpID,
		})
		return
	}

	// Парсинг multipart/form-data
	err = r.ParseMultipartForm(h.cfg.Files.MaxFileSize)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Failed to parse form: %v", err), "INVALID_FORM_DATA")
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		writeError(w, http.StatusBadRequest, "No files provided", "NO_FILES")
		return
	}

	// Валидация количества файлов
	if len(files) > h.cfg.Files.MaxFilesPerRequest {
		writeError(w, http.StatusBadRequest,
			fmt.Sprintf("Maximum %d files allowed per request. Got %d files",
				h.cfg.Files.MaxFilesPerRequest, len(files)),
			"FILE_LIMIT_EXCEEDED")
		return
	}

	if len(files) < 2 {
		writeError(w, http.StatusBadRequest, "At least 2 files required (minimum 1 pair)", "INSUFFICIENT_FILES")
		return
	}

	// Проверка четности количества файлов (должны быть пары)
	if len(files)%2 != 0 {
		writeError(w, http.StatusBadRequest,
			fmt.Sprintf("Files count must be even (pairs of before/after). Got %d files", len(files)),
			"INVALID_FILE_COUNT")
		return
	}

	// Валидация типов файлов
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("Failed to open file: %v", err), "FILE_OPEN_ERROR")
			return
		}
		defer file.Close()

		// Проверка MIME типа
		buffer := make([]byte, 512)
		_, err = file.Read(buffer)
		if err != nil && err != io.EOF {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("Failed to read file: %v", err), "FILE_READ_ERROR")
			return
		}
		_, err = file.Seek(0, 0)
		if err != nil {
			return
		}

		mimeType := http.DetectContentType(buffer)
		if !isAllowedMIMEType(mimeType, h.cfg.Files.AllowedMIMETypes) {
			writeError(w, http.StatusBadRequest,
				fmt.Sprintf("File %s has invalid MIME type: %s", fileHeader.Filename, mimeType),
				"INVALID_FILE_TYPE")
			return
		}

		// Проверка размера файла
		if fileHeader.Size > h.cfg.Files.MaxFileSize {
			writeError(w, http.StatusBadRequest,
				fmt.Sprintf("File %s exceeds maximum size of %d bytes", fileHeader.Filename, h.cfg.Files.MaxFileSize),
				"FILE_TOO_LARGE")
			return
		}
	}

	// Создание операции
	operationID := uuid.New().String()
	pairsCount := len(files) / 2

	// Сохранение файлов (временное решение - в памяти, позже можно сохранять на диск)
	filePaths := make([]string, 0, len(files))
	for _, fileHeader := range files {
		filePaths = append(filePaths, fileHeader.Filename)
	}

	// Сохранение статуса операции
	operation := m.OperationStatus{
		OperationID: operationID,
		Status:      "NEW",
		CreatedAt:   time.Now(),
		Files:       filePaths,
		Progress: &m.Progress{
			Processed: 0,
			Total:     pairsCount,
		},
	}

	operationJSON, err := json.Marshal(operation)
	if err != nil {
		slog.Error("Failed to marshal operation", "err", err)
		writeError(w, http.StatusInternalServerError, "Internal server error", "INTERNAL_ERROR")
		return
	}

	// Сохранение операции в кэш
	ttl := time.Duration(h.cfg.Memcached.DefaultTTL) * time.Second
	err = h.cache.Set(ctx, operationID, operationJSON, ttl)
	if err != nil {
		slog.Error("Failed to save operation to cache", "err", err)
		writeError(w, http.StatusInternalServerError, "Internal server error", "INTERNAL_ERROR")
		return
	}

	// Сохранение связи operationKey -> operationID для идемпотентности
	keyData := []byte(operationID)
	err = h.cache.Set(ctx, "opkey:"+operationKey, keyData, ttl)
	if err != nil {
		slog.Error("Failed to save operation key", "err", err)
		// Не критично, продолжаем
	}

	// TODO: Запуск асинхронной обработки файлов

	writeJSON(w, http.StatusOK, m.UploadResponse{
		OperationID: operationID,
		Status:      "NEW",
		PairsCount:  pairsCount,
		Message:     fmt.Sprintf("%d pairs of files accepted for processing", pairsCount),
	})
}

// GetHandler обрабатывает GET /get
func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed", "METHOD_NOT_ALLOWED")
		return
	}

	operationID := r.URL.Query().Get("id")
	if operationID == "" {
		writeError(w, http.StatusBadRequest, "id parameter is required", "MISSING_ID")
		return
	}

	// Валидация UUID
	_, err := uuid.Parse(operationID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid operation ID format", "INVALID_ID_FORMAT")
		return
	}

	ctx := r.Context()
	operationJSON, err := h.cache.Get(ctx, operationID)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			writeError(w, http.StatusNotFound, "Operation not found", "OPERATION_NOT_FOUND")
			return
		}
		slog.Error("Failed to get operation from cache", "err", err)
		writeError(w, http.StatusInternalServerError, "Internal server error", "INTERNAL_ERROR")
		return
	}

	var operation m.OperationStatus
	err = json.Unmarshal(operationJSON, &operation)
	if err != nil {
		slog.Error("Failed to unmarshal operation", "err", err)
		writeError(w, http.StatusInternalServerError, "Internal server error", "INTERNAL_ERROR")
		return
	}

	// Формирование ответа в зависимости от статуса
	switch operation.Status {
	case "NEW", "PROGRESS":
		writeJSON(w, http.StatusOK, m.OperationInProgress{
			OperationID:                operation.OperationID,
			Status:                     operation.Status,
			Progress:                   operation.Progress,
			EstimatedCompletionSeconds: h.estimateCompletionTime(operation.Progress),
		})
	case "DONE":
		writeJSON(w, http.StatusOK, m.OperationComplete{
			OperationID: operation.OperationID,
			Status:      operation.Status,
			Results:     operation.Results,
		})
	case "ERROR":
		writeJSON(w, http.StatusOK, m.OperationError{
			OperationID: operation.OperationID,
			Status:      operation.Status,
			Error:       operation.Error,
			ErrorCode:   operation.ErrorCode,
			FailedFiles: operation.FailedFiles,
		})
	default:
		writeError(w, http.StatusInternalServerError, "Unknown operation status", "UNKNOWN_STATUS")
	}
}

// Вспомогательные функции

func (h *Handler) checkIdempotency(ctx context.Context, operationKey string) (string, error) {
	keyData, err := h.cache.Get(ctx, "opkey:"+operationKey)
	if err != nil {
		return "", err
	}
	return string(keyData), nil
}

func (h *Handler) estimateCompletionTime(progress *m.Progress) int {
	if progress == nil || progress.Total == 0 {
		return 60 // Дефолтное время
	}
	if progress.Processed == 0 {
		return progress.Total * 10 // Примерно 10 секунд на пару
	}
	remaining := progress.Total - progress.Processed
	avgTimePerPair := 10 // секунд
	return remaining * avgTimePerPair
}

func isAllowedMIMEType(mimeType string, allowedTypes []string) bool {
	for _, allowed := range allowedTypes {
		if mimeType == allowed || strings.HasPrefix(mimeType, allowed) {
			return true
		}
	}
	return false
}

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("Failed to encode JSON response", "err", err)
	}
}

func writeError(w http.ResponseWriter, statusCode int, message, code string) {
	writeJSON(w, statusCode, m.ErrorResponse{
		Error: message,
		Code:  code,
	})
}
