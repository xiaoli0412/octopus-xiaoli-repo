package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func TestImportDBDryRunReturnsSplitCredentialRebindCounts(t *testing.T) {
	setupHandlerTest(t)
	dump := &model.DBDump{
		Version: 1,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   `v1`,
			ExportSource:    `octopus`,
			ContainsSecrets: false,
		},
		Channels: []model.Channel{{ID: 101, Name: `preview-channel`, Enabled: true, Model: `gpt-4o`}},
		ChannelKeys: []model.ChannelKey{{ID: 201, ChannelID: 101, Enabled: true, ChannelKey: ``}},
		APIKeys: []model.APIKey{{ID: 301, Name: `preview-client`, APIKey: ``, Enabled: true}},
	}

	recorder := performImportMultipartHandlerRequest(t, `/api/v1/setting/import?dry_run=true&mode=incremental`, dump, nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf(`dry-run status = %d, want %d, body = %s`, recorder.Code, http.StatusOK, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	var result model.DBImportResult
	if err := json.Unmarshal(response.Data, &result); err != nil {
		t.Fatalf(`json.Unmarshal(result) error = %v`, err)
	}
	if result.Compatibility == nil || result.Compatibility.Summary == nil {
		t.Fatalf(`compatibility = %#v, want populated summary`, result.Compatibility)
	}
	if got := result.Compatibility.Summary.CredentialRebindTargets; got != 2 {
		t.Fatalf(`compatibility.summary.credential_rebind_targets = %d, want 2`, got)
	}
	if got := result.Compatibility.Summary.ChannelKeyRebindTargets; got != 1 {
		t.Fatalf(`compatibility.summary.channel_key_rebind_targets = %d, want 1`, got)
	}
	if got := result.Compatibility.Summary.APIKeyRebindTargets; got != 1 {
		t.Fatalf(`compatibility.summary.api_key_rebind_targets = %d, want 1`, got)
	}
	if len(result.Compatibility.CredentialRebindTargets) != 2 {
		t.Fatalf(`credential_rebind_targets = %#v, want 2`, result.Compatibility.CredentialRebindTargets)
	}
	foundChannelKey := false
	foundAPIKey := false
	for _, target := range result.Compatibility.CredentialRebindTargets {
		switch target.TargetType {
		case `channel_key`:
			foundChannelKey = true
		case `api_key`:
			foundAPIKey = true
		}
	}
	if !foundChannelKey || !foundAPIKey {
		t.Fatalf(`credential_rebind_targets = %#v, want both channel_key and api_key targets`, result.Compatibility.CredentialRebindTargets)
	}
}

func TestPreviewRollbackImportSnapshotReturnsSplitCredentialRebindCounts(t *testing.T) {
	setupHandlerTest(t)

	snapshotDir := filepath.Join(filepath.Dir(filepath.Clean(db.GetCurrentDSN())), `import-snapshots`)
	if err := os.MkdirAll(snapshotDir, 0o755); err != nil {
		t.Fatalf(`MkdirAll(snapshotDir) error = %v`, err)
	}
	snapshotName := `pre-import-redacted-rebind.json`
	snapshotPath := filepath.Join(snapshotDir, snapshotName)
	dump := &model.DBDump{
		Version: 1,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   `v1`,
			ExportSource:    `octopus`,
			ContainsSecrets: false,
		},
		Channels: []model.Channel{{ID: 401, Name: `rollback-preview-channel`, Enabled: true, Model: `gpt-4o`}},
		ChannelKeys: []model.ChannelKey{{ID: 501, ChannelID: 401, Enabled: true, ChannelKey: ``}},
		APIKeys: []model.APIKey{{ID: 601, Name: `rollback-preview-client`, APIKey: ``, Enabled: true}},
	}
	payload, err := json.MarshalIndent(dump, ``, `  `)
	if err != nil {
		t.Fatalf(`json.MarshalIndent(dump) error = %v`, err)
	}
	if err := os.WriteFile(snapshotPath, payload, 0o644); err != nil {
		t.Fatalf(`WriteFile(snapshotPath) error = %v`, err)
	}

	recorder := performJSONHandlerRequest(t, http.MethodPost, `/api/v1/setting/preview-rollback-import-snapshot`, map[string]string{
		`snapshot_name`: snapshotName,
	}, previewRollbackImportSnapshot)
	if recorder.Code != http.StatusOK {
		t.Fatalf(`preview status = %d, want %d, body = %s`, recorder.Code, http.StatusOK, recorder.Body.String())
	}
	response := decodeHandlerResponse(t, recorder)
	var preview model.DBRollbackPreviewResult
	if err := json.Unmarshal(response.Data, &preview); err != nil {
		t.Fatalf(`json.Unmarshal(preview) error = %v`, err)
	}
	if preview.Manifest == nil || preview.Manifest.ContainsSecrets {
		t.Fatalf(`manifest = %#v, want contains_secrets=false`, preview.Manifest)
	}
	if preview.Compatibility == nil || preview.Compatibility.Summary == nil {
		t.Fatalf(`compatibility = %#v, want populated summary`, preview.Compatibility)
	}
	if got := preview.Compatibility.Summary.CredentialRebindTargets; got != 2 {
		t.Fatalf(`compatibility.summary.credential_rebind_targets = %d, want 2`, got)
	}
	if got := preview.Compatibility.Summary.ChannelKeyRebindTargets; got != 1 {
		t.Fatalf(`compatibility.summary.channel_key_rebind_targets = %d, want 1`, got)
	}
	if got := preview.Compatibility.Summary.APIKeyRebindTargets; got != 1 {
		t.Fatalf(`compatibility.summary.api_key_rebind_targets = %d, want 1`, got)
	}
	if len(preview.Compatibility.CredentialRebindTargets) != 2 {
		t.Fatalf(`credential_rebind_targets = %#v, want 2`, preview.Compatibility.CredentialRebindTargets)
	}
	foundChannelKey := false
	foundAPIKey := false
	for _, target := range preview.Compatibility.CredentialRebindTargets {
		switch target.TargetType {
		case `channel_key`:
			foundChannelKey = true
		case `api_key`:
			foundAPIKey = true
		}
	}
	if !foundChannelKey || !foundAPIKey {
		t.Fatalf(`credential_rebind_targets = %#v, want both channel_key and api_key targets`, preview.Compatibility.CredentialRebindTargets)
	}
}
