package observability

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
)

func setupAuditTestDB(t *testing.T) {
	t.Helper()
	if db.GetDB() != nil {
		if err := db.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}
	dbPath := filepath.Join(t.TempDir(), "octopus-audit.db")
	if err := db.InitDB("sqlite", dbPath, false); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	t.Cleanup(func() {
		if db.GetDB() != nil {
			if err := db.Close(); err != nil {
				t.Fatalf("Close() error = %v", err)
			}
		}
	})
}

func TestAuditRecorderRecordAndQuery(t *testing.T) {
	setupAuditTestDB(t)

	recorder := NewAuditRecorder(db.GetDB())

	log := AuditLog{
		UserID:       1,
		Username:     "admin",
		Action:       string(AuditActionCreate),
		ResourceType: string(ResourceTypeChannel),
		ResourceID:   "1",
		ResourceName: "test-channel",
		AfterJSON:    `{"name":"test-channel"}`,
		IP:           "127.0.0.1",
		UserAgent:    "test-agent",
	}

	err := recorder.Record(context.Background(), log)
	if err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	// 查询
	logs, total, err := recorder.Query(context.Background(), AuditQueryFilter{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if total != 1 {
		t.Errorf("total = %v, want 1", total)
	}
	if len(logs) != 1 {
		t.Fatalf("len(logs) = %v, want 1", len(logs))
	}

	got := logs[0]
	if got.Username != "admin" {
		t.Errorf("Username = %q, want %q", got.Username, "admin")
	}
	if got.Action != string(AuditActionCreate) {
		t.Errorf("Action = %q, want %q", got.Action, AuditActionCreate)
	}
	if got.ResourceType != string(ResourceTypeChannel) {
		t.Errorf("ResourceType = %q, want %q", got.ResourceType, ResourceTypeChannel)
	}
	if got.ResourceName != "test-channel" {
		t.Errorf("ResourceName = %q, want %q", got.ResourceName, "test-channel")
	}
	if got.IP != "127.0.0.1" {
		t.Errorf("IP = %q, want %q", got.IP, "127.0.0.1")
	}
}

func TestAuditRecorderGetByID(t *testing.T) {
	setupAuditTestDB(t)

	recorder := NewAuditRecorder(db.GetDB())

	log := AuditLog{
		UserID:       1,
		Username:     "admin",
		Action:       string(AuditActionUpdate),
		ResourceType: string(ResourceTypeAPIKey),
		ResourceID:   "5",
		ResourceName: "test-key",
	}

	if err := recorder.Record(context.Background(), log); err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	// 查询所有以获取 ID
	logs, _, err := recorder.Query(context.Background(), AuditQueryFilter{
		Page:     1,
		PageSize: 10,
	})
	if err != nil || len(logs) == 0 {
		t.Fatalf("Query() error = %v, len = %d", err, len(logs))
	}

	id := logs[0].ID
	got, err := recorder.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetByID() returned nil")
	}
	if got.ID != id {
		t.Errorf("ID = %v, want %v", got.ID, id)
	}
	if got.Action != string(AuditActionUpdate) {
		t.Errorf("Action = %q, want %q", got.Action, AuditActionUpdate)
	}
}

func TestAuditRecorderQueryWithFilter(t *testing.T) {
	setupAuditTestDB(t)

	recorder := NewAuditRecorder(db.GetDB())

	// 写入多条不同类型的审计日志
	records := []AuditLog{
		{UserID: 1, Username: "admin", Action: string(AuditActionCreate), ResourceType: string(ResourceTypeChannel), ResourceID: "1"},
		{UserID: 1, Username: "admin", Action: string(AuditActionDelete), ResourceType: string(ResourceTypeChannel), ResourceID: "2"},
		{UserID: 2, Username: "operator", Action: string(AuditActionCreate), ResourceType: string(ResourceTypeAPIKey), ResourceID: "3"},
	}

	for _, log := range records {
		if err := recorder.Record(context.Background(), log); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}

	// 按 action 过滤
	logs, total, err := recorder.Query(context.Background(), AuditQueryFilter{
		Action:   string(AuditActionCreate),
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if total != 2 {
		t.Errorf("total for action=create = %v, want 2", total)
	}
	if len(logs) != 2 {
		t.Errorf("len(logs) = %v, want 2", len(logs))
	}

	// 按 resource_type 过滤
	logs, total, err = recorder.Query(context.Background(), AuditQueryFilter{
		ResourceType: string(ResourceTypeAPIKey),
		Page:         1,
		PageSize:     10,
	})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if total != 1 {
		t.Errorf("total for resource_type=apikey = %v, want 1", total)
	}

	// 按 user_id 过滤
	logs, total, err = recorder.Query(context.Background(), AuditQueryFilter{
		UserID:   2,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if total != 1 {
		t.Errorf("total for user_id=2 = %v, want 1", total)
	}
	if len(logs) != 1 {
		t.Fatalf("len(logs) = %v, want 1", len(logs))
	}
}

func TestAuditRecorderQueryWithTimeRange(t *testing.T) {
	setupAuditTestDB(t)

	recorder := NewAuditRecorder(db.GetDB())

	now := time.Now()
	log := AuditLog{
		UserID:       1,
		Username:     "admin",
		Action:       string(AuditActionCreate),
		ResourceType: string(ResourceTypeChannel),
		CreatedAt:    now,
	}
	if err := recorder.Record(context.Background(), log); err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	// 查询包含当前时间范围
	logs, total, err := recorder.Query(context.Background(), AuditQueryFilter{
		StartTime: now.Add(-1 * time.Hour),
		EndTime:   now.Add(1 * time.Hour),
		Page:      1,
		PageSize:  10,
	})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if total != 1 {
		t.Errorf("total = %v, want 1", total)
	}
	if len(logs) != 1 {
		t.Fatalf("len(logs) = %v, want 1", len(logs))
	}

	// 查询不包含的时间范围
	logs, total, err = recorder.Query(context.Background(), AuditQueryFilter{
		StartTime: now.Add(-2 * time.Hour),
		EndTime:   now.Add(-1 * time.Hour),
		Page:      1,
		PageSize:  10,
	})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if total != 0 {
		t.Errorf("total for excluded range = %v, want 0", total)
	}
}

func TestAuditRecorderNilDB(t *testing.T) {
	recorder := NewAuditRecorder(nil)

	// nil db 不应 panic
	err := recorder.Record(context.Background(), AuditLog{})
	if err != nil {
		t.Errorf("Record() with nil db error = %v", err)
	}

	logs, total, err := recorder.Query(context.Background(), AuditQueryFilter{})
	if err != nil {
		t.Errorf("Query() with nil db error = %v", err)
	}
	if total != 0 || logs != nil {
		t.Errorf("Query() with nil db: total = %v, logs = %v", total, logs)
	}

	got, err := recorder.GetByID(context.Background(), 1)
	if err != nil {
		t.Errorf("GetByID() with nil db error = %v", err)
	}
	if got != nil {
		t.Errorf("GetByID() with nil db: got = %v, want nil", got)
	}
}

func TestAuditLogTableName(t *testing.T) {
	if got := (AuditLog{}).TableName(); got != "audit_logs" {
		t.Errorf("TableName() = %q, want %q", got, "audit_logs")
	}
}

func TestAuditQueryFilterPagination(t *testing.T) {
	setupAuditTestDB(t)

	recorder := NewAuditRecorder(db.GetDB())

	// 写入 5 条记录
	for i := 0; i < 5; i++ {
		if err := recorder.Record(context.Background(), AuditLog{
			UserID:       1,
			Username:     "admin",
			Action:       string(AuditActionCreate),
			ResourceType: string(ResourceTypeChannel),
			ResourceID:   "1",
		}); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}

	// 第一页，每页 2 条
	logs, total, err := recorder.Query(context.Background(), AuditQueryFilter{
		Page:     1,
		PageSize: 2,
	})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if total != 5 {
		t.Errorf("total = %v, want 5", total)
	}
	if len(logs) != 2 {
		t.Errorf("len(logs) page 1 = %v, want 2", len(logs))
	}

	// 第三页，每页 2 条（应该只有 1 条）
	logs, _, err = recorder.Query(context.Background(), AuditQueryFilter{
		Page:     3,
		PageSize: 2,
	})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("len(logs) page 3 = %v, want 1", len(logs))
	}
}

func TestSanitizeAuditJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "password field masked",
			input: `{"username":"admin","password":"secret123"}`,
			want:  `{"password":"***","username":"admin"}`,
		},
		{
			name:  "api_key field masked",
			input: `{"name":"my-channel","api_key":"sk-abc123"}`,
			want:  `{"api_key":"***","name":"my-channel"}`,
		},
		{
			name:  "token field masked",
			input: `{"token":"bearer-xyz","data":"safe"}`,
			want:  `{"data":"safe","token":"***"}`,
		},
		{
			name:  "nested sensitive field",
			input: `{"config":{"secret":"my-secret","port":8080},"name":"app"}`,
			want:  `{"config":{"port":8080,"secret":"***"},"name":"app"}`,
		},
		{
			name:  "array with sensitive fields",
			input: `[{"key":"sk-1","name":"a"},{"key":"sk-2","name":"b"}]`,
			want:  `[{"key":"***","name":"a"},{"key":"***","name":"b"}]`,
		},
		{
			name:  "case insensitive key matching",
			input: `{"Password":"secret","API_KEY":"sk-xyz"}`,
			want:  `{"API_KEY":"***","Password":"***"}`,
		},
		{
			name:  "non-sensitive fields unchanged",
			input: `{"name":"channel","type":"openai","base_url":"https://api.openai.com"}`,
			want:  `{"base_url":"https://api.openai.com","name":"channel","type":"openai"}`,
		},
		{
			name:  "non-JSON text fallback",
			input: `password=test123&api_key=sk-abc`,
			want:  `password=test123&api_key=sk-abc`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeAuditJSON(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeAuditJSON(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanitizeAuditJSONMultipleSensitiveFields(t *testing.T) {
	input := `{"password":"p1","api_key":"k1","name":"test","secret":"s1","token":"t1"}`
	got := SanitizeAuditJSON(input)

	// 验证所有敏感字段都被脱敏
	if !strings.Contains(got, `"password":"***"`) {
		t.Errorf("password not masked: %s", got)
	}
	if !strings.Contains(got, `"api_key":"***"`) {
		t.Errorf("api_key not masked: %s", got)
	}
	if !strings.Contains(got, `"secret":"***"`) {
		t.Errorf("secret not masked: %s", got)
	}
	if !strings.Contains(got, `"token":"***"`) {
		t.Errorf("token not masked: %s", got)
	}
	if !strings.Contains(got, `"name":"test"`) {
		t.Errorf("non-sensitive field changed: %s", got)
	}
}
