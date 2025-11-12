package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Caritas-Team/reviewer/internal/config"
	"github.com/Caritas-Team/reviewer/internal/model"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/google/uuid"
)

// mockCache - мок для кэша
type mockCache struct {
	storage map[string][]byte
}

func newMockCache() *mockCache {
	return &mockCache{
		storage: make(map[string][]byte),
	}
}

func (m *mockCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, ok := m.storage[key]
	if !ok {
		return nil, memcache.ErrCacheMiss
	}
	return val, nil
}

func (m *mockCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.storage[key] = value
	return nil
}

func (m *mockCache) Close() error {
	return nil
}

func (m *mockCache) IsHealthy(ctx context.Context) bool {
	return true
}

// createTestPDF создает минимальный валидный PDF файл в памяти
func createTestPDF() []byte {
	// Минимальный валидный PDF документ
	return []byte(`%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
2 0 obj
<<
/Type /Pages
/Kids [3 0 R]
/Count 1
>>
endobj
3 0 obj
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
/Contents 4 0 R
/Resources <<
/Font <<
/F1 <<
/Type /Font
/Subtype /Type1
/BaseFont /Helvetica
>>
>>
>>
>>
endobj
4 0 obj
<<
/Length 44
>>
stream
BT
/F1 12 Tf
100 700 Td
(Test PDF) Tj
ET
endstream
endobj
xref
0 5
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
0000000115 00000 n 
0000000306 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
398
%%EOF`)
}

// createMultipartRequest создает тестовый multipart/form-data запрос с файлами
func createMultipartRequest(t *testing.T, url string, files [][]byte, filenames []string, operationKey string) *http.Request {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for i, fileData := range files {
		part, err := writer.CreateFormFile("files", filenames[i])
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}
		_, err = part.Write(fileData)
		if err != nil {
			t.Fatalf("Failed to write file data: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, url, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if operationKey != "" {
		req.Header.Set("X-Operation-Key", operationKey)
	}

	return req
}

func TestUploadHandler_Success(t *testing.T) {
	cfg := config.Config{
		Files: config.Files{
			MaxFilesPerRequest: 20,
			MaxFileSize:        10 * 1024 * 1024, // 10MB
			AllowedMIMETypes:   []string{"application/pdf"},
		},
		Memcached: config.Memcached{
			DefaultTTL: 300,
		},
	}

	mockCache := newMockCache()
	handler := &Handler{
		cfg:   cfg,
		cache: mockCache,
	}

	// Создаем тестовые PDF файлы
	file1 := createTestPDF()
	file2 := createTestPDF()

	req := createMultipartRequest(t, "/upload", [][]byte{file1, file2}, []string{"test1.pdf", "test2.pdf"}, "op-test-12345")
	w := httptest.NewRecorder()

	handler.UploadHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response model.UploadResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Status != "NEW" {
		t.Errorf("Expected status NEW, got %s", response.Status)
	}

	if response.PairsCount != 1 {
		t.Errorf("Expected pairs_count 1, got %d", response.PairsCount)
	}

	if response.OperationID == "" {
		t.Error("Expected operation_id to be set")
	}

	// Проверяем, что операция сохранена в кэш
	ctx := context.Background()
	operationData, err := mockCache.Get(ctx, response.OperationID)
	if err != nil {
		t.Fatalf("Failed to get operation from cache: %v", err)
	}

	var operation OperationStatus
	if err := json.Unmarshal(operationData, &operation); err != nil {
		t.Fatalf("Failed to unmarshal operation: %v", err)
	}

	if operation.Status != "NEW" {
		t.Errorf("Expected operation status NEW, got %s", operation.Status)
	}
}

func TestUploadHandler_MissingOperationKey(t *testing.T) {
	cfg := config.Config{
		Files: config.Files{
			MaxFilesPerRequest: 20,
			MaxFileSize:        10 * 1024 * 1024,
			AllowedMIMETypes:   []string{"application/pdf"},
		},
	}

	mockCache := newMockCache()
	handler := &Handler{
		cfg:   cfg,
		cache: mockCache,
	}

	file1 := createTestPDF()
	file2 := createTestPDF()

	req := createMultipartRequest(t, "/upload", [][]byte{file1, file2}, []string{"test1.pdf", "test2.pdf"}, "")
	w := httptest.NewRecorder()

	handler.UploadHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Code != "MISSING_OPERATION_KEY" {
		t.Errorf("Expected code MISSING_OPERATION_KEY, got %s", response.Code)
	}
}

func TestUploadHandler_OddFileCount(t *testing.T) {
	cfg := config.Config{
		Files: config.Files{
			MaxFilesPerRequest: 20,
			MaxFileSize:        10 * 1024 * 1024,
			AllowedMIMETypes:   []string{"application/pdf"},
		},
	}

	mockCache := newMockCache()
	handler := &Handler{
		cfg:   cfg,
		cache: mockCache,
	}

	file1 := createTestPDF()

	req := createMultipartRequest(t, "/upload", [][]byte{file1}, []string{"test1.pdf"}, "op-test-12345")
	w := httptest.NewRecorder()

	handler.UploadHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Code != "INSUFFICIENT_FILES" && response.Code != "INVALID_FILE_COUNT" {
		t.Errorf("Expected code INSUFFICIENT_FILES or INVALID_FILE_COUNT, got %s", response.Code)
	}
}

func TestUploadHandler_DuplicateOperationKey(t *testing.T) {
	cfg := config.Config{
		Files: config.Files{
			MaxFilesPerRequest: 20,
			MaxFileSize:        10 * 1024 * 1024,
			AllowedMIMETypes:   []string{"application/pdf"},
		},
		Memcached: config.Memcached{
			DefaultTTL: 300,
		},
	}

	mockCache := newMockCache()
	handler := &Handler{
		cfg:   cfg,
		cache: mockCache,
	}

	file1 := createTestPDF()
	file2 := createTestPDF()
	operationKey := "op-test-duplicate-12345"

	// Первый запрос - должен быть успешным
	req1 := createMultipartRequest(t, "/upload", [][]byte{file1, file2}, []string{"test1.pdf", "test2.pdf"}, operationKey)
	w1 := httptest.NewRecorder()
	handler.UploadHandler(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("First request should succeed, got status %d. Body: %s", w1.Code, w1.Body.String())
	}

	// Второй запрос с тем же ключом - должен вернуть 409
	req2 := createMultipartRequest(t, "/upload", [][]byte{file1, file2}, []string{"test1.pdf", "test2.pdf"}, operationKey)
	w2 := httptest.NewRecorder()
	handler.UploadHandler(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d. Body: %s", w2.Code, w2.Body.String())
	}

	var response DuplicateOperationError
	if err := json.Unmarshal(w2.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Code != "DUPLICATE_OPERATION_KEY" {
		t.Errorf("Expected code DUPLICATE_OPERATION_KEY, got %s", response.Code)
	}
}

func TestUploadHandler_WithRealPDFFile(t *testing.T) {
	// Тест с реальным PDF файлом из docs/example.pdf
	pdfPath := filepath.Join("..", "..", "docs", "example.pdf")
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		t.Skip("Skipping test: example.pdf not found")
	}

	pdfData, err := os.ReadFile(pdfPath)
	if err != nil {
		t.Fatalf("Failed to read PDF file: %v", err)
	}

	cfg := config.Config{
		Files: config.Files{
			MaxFilesPerRequest: 20,
			MaxFileSize:        10 * 1024 * 1024,
			AllowedMIMETypes:   []string{"application/pdf", "application/octet-stream"},
		},
		Memcached: config.Memcached{
			DefaultTTL: 300,
		},
	}

	mockCache := newMockCache()
	handler := &Handler{
		cfg:   cfg,
		cache: mockCache,
	}

	// Используем реальный PDF дважды (как до и после)
	req := createMultipartRequest(t, "/upload", [][]byte{pdfData, pdfData}, []string{"example.pdf", "example_after.pdf"}, "op-test-real-pdf")
	w := httptest.NewRecorder()

	handler.UploadHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response UploadResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Status != "NEW" {
		t.Errorf("Expected status NEW, got %s", response.Status)
	}

	if response.PairsCount != 1 {
		t.Errorf("Expected pairs_count 1, got %d", response.PairsCount)
	}
}

func TestGetHandler_Success(t *testing.T) {
	cfg := config.Config{}
	mockCache := newMockCache()
	handler := &Handler{
		cfg:   cfg,
		cache: mockCache,
	}

	// Создаем тестовую операцию с валидным UUID
	operationID := uuid.New().String()
	operation := OperationStatus{
		OperationID: operationID,
		Status:      "NEW",
		Progress: &Progress{
			Processed: 0,
			Total:     1,
		},
	}

	operationJSON, _ := json.Marshal(operation)
	ctx := context.Background()
	mockCache.Set(ctx, operation.OperationID, operationJSON, time.Hour)

	req := httptest.NewRequest(http.MethodGet, "/get?id="+operationID, nil)
	w := httptest.NewRecorder()

	handler.GetHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response OperationInProgress
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Status != "NEW" {
		t.Errorf("Expected status NEW, got %s", response.Status)
	}
}

func TestGetHandler_NotFound(t *testing.T) {
	cfg := config.Config{}
	mockCache := newMockCache()
	handler := &Handler{
		cfg:   cfg,
		cache: mockCache,
	}

	// Используем валидный UUID, но не существующий в кэше
	nonExistentID := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/get?id="+nonExistentID, nil)
	w := httptest.NewRecorder()

	handler.GetHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Code != "OPERATION_NOT_FOUND" {
		t.Errorf("Expected code OPERATION_NOT_FOUND, got %s", response.Code)
	}
}
