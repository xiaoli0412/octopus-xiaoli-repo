package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/observability"
)

// mockAuditRecorder 用于测试的 mock 审计记录器
type mockAuditRecorder struct {
	mu     sync.Mutex
	logs   []observability.AuditLog
	byID   map[uint]*observability.AuditLog
	nextID uint
}

func newMockAuditRecorder() *mockAuditRecorder {
	return &mockAuditRecorder{
		byID:   make(map[uint]*observability.AuditLog),
		nextID: 1,
	}
}

func (m *mockAuditRecorder) Record(ctx context.Context, log observability.AuditLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	log.ID = m.nextID
	m.nextID++
	m.logs = append(m.logs, log)
	m.byID[log.ID] = &log
	return nil
}

func (m *mockAuditRecorder) Query(ctx context.Context, filter observability.AuditQueryFilter) ([]observability.AuditLog, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.logs, int64(len(m.logs)), nil
}

func (m *mockAuditRecorder) GetByID(ctx context.Context, id uint) (*observability.AuditLog, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if log, ok := m.byID[id]; ok {
		return log, nil
	}
	return nil, nil
}

// Logs 返回当前已记录的审计日志副本（线程安全）
func (m *mockAuditRecorder) Logs() []observability.AuditLog {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]observability.AuditLog, len(m.logs))
	copy(cp, m.logs)
	return cp
}

func TestAuditMiddlewareSkipsGetRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := newMockAuditRecorder()

	router := gin.New()
	router.Use(AuditMiddleware(recorder))
	router.GET("/api/v1/channel/list", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/channel/list", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// GET 请求不应记录审计日志
	if logs := recorder.Logs(); len(logs) != 0 {
		t.Errorf("expected 0 audit logs for GET, got %d", len(logs))
	}
}

func TestAuditMiddlewareSkipsNonSensitivePaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := newMockAuditRecorder()

	router := gin.New()
	router.Use(AuditMiddleware(recorder))
	router.POST("/api/v1/non-sensitive", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/non-sensitive", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if logs := recorder.Logs(); len(logs) != 0 {
		t.Errorf("expected 0 audit logs for non-sensitive path, got %d", len(logs))
	}
}

func TestAuditMiddlewareRecordsPostToSensitivePath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := newMockAuditRecorder()

	router := gin.New()
	router.Use(AuditMiddleware(recorder))
	router.POST("/api/v1/channel/create", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	body := `{"name":"test-channel"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/channel/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 等待异步写入
	time.Sleep(100 * time.Millisecond)

	logs := recorder.Logs()
	if len(logs) != 1 {
		t.Fatalf("expected 1 audit log, got %d", len(logs))
	}

	log := logs[0]
	if log.Action != string(observability.AuditActionCreate) {
		t.Errorf("Action = %q, want %q", log.Action, observability.AuditActionCreate)
	}
	if log.ResourceType != string(observability.ResourceTypeChannel) {
		t.Errorf("ResourceType = %q, want %q", log.ResourceType, observability.ResourceTypeChannel)
	}
}

func TestAuditMiddlewareRecordsDeleteAction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := newMockAuditRecorder()

	router := gin.New()
	router.Use(AuditMiddleware(recorder))
	router.DELETE("/api/v1/apikey/delete/1", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/apikey/delete/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	time.Sleep(100 * time.Millisecond)

	logs := recorder.Logs()
	if len(logs) != 1 {
		t.Fatalf("expected 1 audit log, got %d", len(logs))
	}

	log := logs[0]
	if log.Action != string(observability.AuditActionDelete) {
		t.Errorf("Action = %q, want %q", log.Action, observability.AuditActionDelete)
	}
	if log.ResourceType != string(observability.ResourceTypeAPIKey) {
		t.Errorf("ResourceType = %q, want %q", log.ResourceType, observability.ResourceTypeAPIKey)
	}
}

func TestAuditMiddlewareNilRecorder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(AuditMiddleware(nil))
	router.POST("/api/v1/channel/create", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/channel/create", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status code = %v, want %v", w.Code, http.StatusOK)
	}
}

func TestIsAuditSensitiveRequest(t *testing.T) {
	tests := []struct {
		method   string
		path     string
		expected bool
	}{
		{http.MethodPost, "/api/v1/channel/create", true},
		{http.MethodPut, "/api/v1/group/update", true},
		{http.MethodDelete, "/api/v1/apikey/delete/1", true},
		{http.MethodPost, "/api/v1/setting/update", true},
		{http.MethodPost, "/api/v1/user/change-password", true},
		{http.MethodPost, "/api/v1/backup/restore", true},
		{http.MethodGet, "/api/v1/channel/list", false},
		{http.MethodPost, "/api/v1/non-sensitive", false},
		{http.MethodGet, "/healthz", false},
	}

	for _, tt := range tests {
		got := isAuditSensitiveRequest(tt.method, tt.path)
		if got != tt.expected {
			t.Errorf("isAuditSensitiveRequest(%q, %q) = %v, want %v", tt.method, tt.path, got, tt.expected)
		}
	}
}

func TestInferAuditAction(t *testing.T) {
	tests := []struct {
		method   string
		path     string
		expected string
	}{
		{http.MethodPost, "/api/v1/channel/create", string(observability.AuditActionCreate)},
		{http.MethodPost, "/api/v1/channel/delete/1", string(observability.AuditActionDelete)},
		{http.MethodDelete, "/api/v1/apikey/1", string(observability.AuditActionDelete)},
		{http.MethodPost, "/api/v1/channel/enable/1", string(observability.AuditActionEnable)},
		{http.MethodPost, "/api/v1/channel/disable/1", string(observability.AuditActionDisable)},
		{http.MethodPost, "/api/v1/backup/restore", string(observability.AuditActionRestore)},
		{http.MethodPost, "/api/v1/backup/create", string(observability.AuditActionBackup)},
		{http.MethodPost, "/api/v1/channel/update", string(observability.AuditActionUpdate)},
	}

	for _, tt := range tests {
		got := inferAuditAction(tt.method, tt.path)
		if got != tt.expected {
			t.Errorf("inferAuditAction(%q, %q) = %q, want %q", tt.method, tt.path, got, tt.expected)
		}
	}
}

func TestInferResourceType(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/api/v1/channel/create", string(observability.ResourceTypeChannel)},
		{"/api/v1/group/update", string(observability.ResourceTypeGroup)},
		{"/api/v1/apikey/delete/1", string(observability.ResourceTypeAPIKey)},
		{"/api/v1/setting/update", string(observability.ResourceTypeSetting)},
		{"/api/v1/user/change-password", string(observability.ResourceTypeUser)},
		{"/api/v1/unknown", ""},
	}

	for _, tt := range tests {
		got := inferResourceType(tt.path)
		if got != tt.expected {
			t.Errorf("inferResourceType(%q) = %q, want %q", tt.path, got, tt.expected)
		}
	}
}

func TestAuditMiddlewareWithDB(t *testing.T) {
	// 使用真实 DB 测试完整流程
	if db.GetDB() != nil {
		if err := db.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}
	dbPath := filepath.Join(t.TempDir(), "octopus-audit-mw.db")
	if err := db.InitDB("sqlite", dbPath, false); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}()

	recorder := observability.NewAuditRecorder(db.GetDB())

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuditMiddleware(recorder))
	router.POST("/api/v1/channel/create", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	body := `{"name":"test-channel"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/channel/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 等待异步写入
	time.Sleep(200 * time.Millisecond)

	// 验证审计日志已写入 DB
	logs, total, err := recorder.Query(context.Background(), observability.AuditQueryFilter{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if total != 1 {
		t.Fatalf("total = %v, want 1", total)
	}
	if len(logs) != 1 {
		t.Fatalf("len(logs) = %v, want 1", len(logs))
	}
	if logs[0].Action != string(observability.AuditActionCreate) {
		t.Errorf("Action = %q, want %q", logs[0].Action, observability.AuditActionCreate)
	}
}

// setupAuditTestDB 初始化一个独立的 SQLite 数据库用于审计中间件测试，返回清理函数。
func setupAuditTestDB(t *testing.T) func() {
	t.Helper()
	if db.GetDB() != nil {
		if err := db.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}
	dbPath := filepath.Join(t.TempDir(), "octopus-audit-before-test.db")
	if err := db.InitDB("sqlite", dbPath, false); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	return func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}
}

// TestAuditMiddlewareUpdateRecordsBeforeSnapshot 验证更新类操作记录 before + after 快照。
func TestAuditMiddlewareUpdateRecordsBeforeSnapshot(t *testing.T) {
	cleanup := setupAuditTestDB(t)
	defer cleanup()

	// 预置一条 APIKey 记录
	apiKey := model.APIKey{
		Name:    "test-key-before-update",
		APIKey:  "sk-test-before-update-secret",
		Enabled: true,
	}
	if err := db.GetDB().Create(&apiKey).Error; err != nil {
		t.Fatalf("create apikey error = %v", err)
	}

	recorder := newMockAuditRecorder()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuditMiddleware(recorder))
	router.POST("/api/v1/apikey/update", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 发送更新请求（body 包含 id）
	body := `{"id":` + strconv.Itoa(apiKey.ID) + `,"name":"renamed-key"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/apikey/update", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	time.Sleep(100 * time.Millisecond)

	logs := recorder.Logs()
	if len(logs) != 1 {
		t.Fatalf("expected 1 audit log, got %d", len(logs))
	}
	log := logs[0]
	if log.Action != string(observability.AuditActionUpdate) {
		t.Errorf("Action = %q, want %q", log.Action, observability.AuditActionUpdate)
	}
	if log.BeforeJSON == "" {
		t.Errorf("BeforeJSON should not be empty for update action")
	}
	if !strings.Contains(log.BeforeJSON, "test-key-before-update") {
		t.Errorf("BeforeJSON should contain original name, got %q", log.BeforeJSON)
	}
	if log.AfterJSON == "" {
		t.Errorf("AfterJSON should not be empty for update action")
	}
	// resourceID 应等于 apiKey.ID
	if log.ResourceID != strconv.Itoa(apiKey.ID) {
		t.Errorf("ResourceID = %q, want %d", log.ResourceID, apiKey.ID)
	}
}

// TestAuditMiddlewareDeleteRecordsBeforeSnapshot 验证删除类操作记录 before 快照，after 为 "[deleted]"。
func TestAuditMiddlewareDeleteRecordsBeforeSnapshot(t *testing.T) {
	cleanup := setupAuditTestDB(t)
	defer cleanup()

	// 预置一条 APIKey 记录
	apiKey := model.APIKey{
		Name:    "test-key-before-delete",
		APIKey:  "sk-test-before-delete-secret",
		Enabled: true,
	}
	if err := db.GetDB().Create(&apiKey).Error; err != nil {
		t.Fatalf("create apikey error = %v", err)
	}

	recorder := newMockAuditRecorder()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuditMiddleware(recorder))
	router.DELETE("/api/v1/apikey/delete/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/apikey/delete/"+strconv.Itoa(apiKey.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	time.Sleep(100 * time.Millisecond)

	logs := recorder.Logs()
	if len(logs) != 1 {
		t.Fatalf("expected 1 audit log, got %d", len(logs))
	}
	log := logs[0]
	if log.Action != string(observability.AuditActionDelete) {
		t.Errorf("Action = %q, want %q", log.Action, observability.AuditActionDelete)
	}
	if log.BeforeJSON == "" {
		t.Errorf("BeforeJSON should not be empty for delete action")
	}
	if !strings.Contains(log.BeforeJSON, "test-key-before-delete") {
		t.Errorf("BeforeJSON should contain original name, got %q", log.BeforeJSON)
	}
	if log.AfterJSON != "[deleted]" {
		t.Errorf("AfterJSON = %q, want %q", log.AfterJSON, "[deleted]")
	}
	if log.ResourceID != strconv.Itoa(apiKey.ID) {
		t.Errorf("ResourceID = %q, want %d", log.ResourceID, apiKey.ID)
	}
}

// TestAuditMiddlewareSensitiveFieldSanitization 验证 password、api_key 等敏感字段在 before/after 快照中被脱敏。
func TestAuditMiddlewareSensitiveFieldSanitization(t *testing.T) {
	cleanup := setupAuditTestDB(t)
	defer cleanup()

	// 预置一条 User 记录（包含 Password 字段）
	user := model.User{
		Username: "audit-test-user",
		Password: "super-secret-password-value",
	}
	if err := db.GetDB().Create(&user).Error; err != nil {
		t.Fatalf("create user error = %v", err)
	}

	// 预置一条 APIKey 记录（包含 api_key 字段）
	apiKey := model.APIKey{
		Name:    "sensitive-key",
		APIKey:  "sk-super-secret-api-key-value",
		Enabled: true,
	}
	if err := db.GetDB().Create(&apiKey).Error; err != nil {
		t.Fatalf("create apikey error = %v", err)
	}

	recorder := newMockAuditRecorder()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuditMiddleware(recorder))
	router.POST("/api/v1/apikey/update", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 更新 APIKey：before 快照应包含 api_key 字段并被脱敏
	body := `{"id":` + strconv.Itoa(apiKey.ID) + `,"name":"renamed"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/apikey/update", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	time.Sleep(100 * time.Millisecond)

	logs := recorder.Logs()
	if len(logs) != 1 {
		t.Fatalf("expected 1 audit log, got %d", len(logs))
	}
	log := logs[0]
	// before 快照中的 api_key 值应被替换为 ***
	if strings.Contains(log.BeforeJSON, "sk-super-secret-api-key-value") {
		t.Errorf("BeforeJSON should not contain raw api_key value, got %q", log.BeforeJSON)
	}
	if !strings.Contains(log.BeforeJSON, "***") {
		t.Errorf("BeforeJSON should contain masked value ***, got %q", log.BeforeJSON)
	}
}

// TestAuditMiddlewareQueryFailureDoesNotBlock 验证查询失败时不阻塞请求。
func TestAuditMiddlewareQueryFailureDoesNotBlock(t *testing.T) {
	cleanup := setupAuditTestDB(t)
	defer cleanup()

	recorder := newMockAuditRecorder()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuditMiddleware(recorder))
	router.POST("/api/v1/apikey/update", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 使用不存在的 id，fetchResourceSnapshot 会返回 record not found 错误
	body := `{"id":99999,"name":"nonexistent"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/apikey/update", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 请求不应被阻塞
	if w.Code != http.StatusOK {
		t.Errorf("status code = %v, want %v (request should not be blocked)", w.Code, http.StatusOK)
	}

	time.Sleep(100 * time.Millisecond)

	logs := recorder.Logs()
	if len(logs) != 1 {
		t.Fatalf("expected 1 audit log, got %d", len(logs))
	}
	log := logs[0]
	// BeforeJSON 应包含错误标记
	if !strings.Contains(log.BeforeJSON, "[error:") {
		t.Errorf("BeforeJSON should contain error marker, got %q", log.BeforeJSON)
	}
}

// TestShouldFetchBeforeSnapshot 验证 shouldFetchBeforeSnapshot 函数。
func TestShouldFetchBeforeSnapshot(t *testing.T) {
	tests := []struct {
		action       string
		resourceType string
		expected     bool
	}{
		{string(observability.AuditActionUpdate), string(observability.ResourceTypeChannel), true},
		{string(observability.AuditActionDelete), string(observability.ResourceTypeAPIKey), true},
		{string(observability.AuditActionEnable), string(observability.ResourceTypeChannel), true},
		{string(observability.AuditActionDisable), string(observability.ResourceTypeChannel), true},
		{string(observability.AuditActionCreate), string(observability.ResourceTypeChannel), false},
		{string(observability.AuditActionLogin), string(observability.ResourceTypeUser), false},
		{string(observability.AuditActionBackup), string(observability.ResourceTypeSetting), false},
		{string(observability.AuditActionRestore), string(observability.ResourceTypeSetting), false},
		{string(observability.AuditActionUpdate), "", false},
	}

	for _, tt := range tests {
		got := shouldFetchBeforeSnapshot(tt.action, tt.resourceType)
		if got != tt.expected {
			t.Errorf("shouldFetchBeforeSnapshot(%q, %q) = %v, want %v", tt.action, tt.resourceType, got, tt.expected)
		}
	}
}

// TestExtractResourceID 验证 extractResourceID 函数。
func TestExtractResourceID(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		resourceType string
		body         []byte
		userID       uint
		expected     string
	}{
		{
			name:         "body with numeric id",
			path:         "/api/v1/apikey/update",
			resourceType: string(observability.ResourceTypeAPIKey),
			body:         []byte(`{"id":42,"name":"test"}`),
			userID:       0,
			expected:     "42",
		},
		{
			name:         "body with string id",
			path:         "/api/v1/channel/update",
			resourceType: string(observability.ResourceTypeChannel),
			body:         []byte(`{"id":"abc"}`),
			userID:       0,
			expected:     "abc",
		},
		{
			name:         "path with numeric id",
			path:         "/api/v1/apikey/delete/7",
			resourceType: string(observability.ResourceTypeAPIKey),
			body:         nil,
			userID:       0,
			expected:     "7",
		},
		{
			name:         "setting uses key field",
			path:         "/api/v1/setting/set",
			resourceType: string(observability.ResourceTypeSetting),
			body:         []byte(`{"key":"cors_allow_origins","value":"*"}`),
			userID:       0,
			expected:     "cors_allow_origins",
		},
		{
			name:         "user falls back to JWT user ID",
			path:         "/api/v1/user/change-password",
			resourceType: string(observability.ResourceTypeUser),
			body:         []byte(`{"old_password":"x","new_password":"y"}`),
			userID:       5,
			expected:     "5",
		},
		{
			name:         "body id takes precedence over path",
			path:         "/api/v1/apikey/delete/99",
			resourceType: string(observability.ResourceTypeAPIKey),
			body:         []byte(`{"id":3}`),
			userID:       0,
			expected:     "3",
		},
		{
			name:         "empty body and non-numeric path",
			path:         "/api/v1/setting/import",
			resourceType: string(observability.ResourceTypeSetting),
			body:         nil,
			userID:       0,
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractResourceID(tt.path, tt.resourceType, tt.body, tt.userID)
			if got != tt.expected {
				t.Errorf("extractResourceID() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestFetchResourceSnapshotChannel 验证 fetchResourceSnapshot 对 channel 资源的查询。
func TestFetchResourceSnapshotChannel(t *testing.T) {
	cleanup := setupAuditTestDB(t)
	defer cleanup()

	// 预置一条 Channel 记录
	channel := model.Channel{
		Name:    "test-channel-snapshot",
		Enabled: true,
	}
	if err := db.GetDB().Create(&channel).Error; err != nil {
		t.Fatalf("create channel error = %v", err)
	}

	snapshot, err := fetchResourceSnapshot(string(observability.ResourceTypeChannel), strconv.Itoa(channel.ID))
	if err != nil {
		t.Fatalf("fetchResourceSnapshot error = %v", err)
	}
	if !strings.Contains(snapshot, "test-channel-snapshot") {
		t.Errorf("snapshot should contain channel name, got %q", snapshot)
	}
}

// TestFetchResourceSnapshotNilDB 验证 db 未初始化时返回错误。
func TestFetchResourceSnapshotNilDB(t *testing.T) {
	cleanup := setupAuditTestDB(t)
	defer cleanup()

	// 关闭 DB 后，GetDB 仍返回非 nil（已关闭的 gorm.DB），但查询会失败。
	// 这里通过不存在的 id 验证查询失败路径。
	_, err := fetchResourceSnapshot(string(observability.ResourceTypeChannel), "999999")
	if err == nil {
		t.Errorf("fetchResourceSnapshot with nonexistent id should return error")
	}
}

// TestFetchResourceSnapshotEmptyID 验证空 resourceID 返回空字符串无错误。
func TestFetchResourceSnapshotEmptyID(t *testing.T) {
	cleanup := setupAuditTestDB(t)
	defer cleanup()

	snapshot, err := fetchResourceSnapshot(string(observability.ResourceTypeChannel), "")
	if err != nil {
		t.Errorf("fetchResourceSnapshot with empty id should not return error, got %v", err)
	}
	if snapshot != "" {
		t.Errorf("fetchResourceSnapshot with empty id should return empty string, got %q", snapshot)
	}
}

// TestFetchResourceSnapshotSetting 验证 fetchResourceSnapshot 对 setting 资源的查询。
func TestFetchResourceSnapshotSetting(t *testing.T) {
	cleanup := setupAuditTestDB(t)
	defer cleanup()

	// 预置一条 Setting 记录
	setting := model.Setting{
		Key:   model.SettingKeyCORSAllowOrigins,
		Value: "https://example.com",
	}
	if err := db.GetDB().Create(&setting).Error; err != nil {
		t.Fatalf("create setting error = %v", err)
	}

	snapshot, err := fetchResourceSnapshot(string(observability.ResourceTypeSetting), string(model.SettingKeyCORSAllowOrigins))
	if err != nil {
		t.Fatalf("fetchResourceSnapshot error = %v", err)
	}
	if !strings.Contains(snapshot, "cors_allow_origins") {
		t.Errorf("snapshot should contain setting key, got %q", snapshot)
	}
	if !strings.Contains(snapshot, "https://example.com") {
		t.Errorf("snapshot should contain setting value, got %q", snapshot)
	}
}
