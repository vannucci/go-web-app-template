package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type Workspace struct {
	ID        int               `db:"id"`
	Name      string            `db:"name"`
	Slug      string            `db:"slug"`
	Features  WorkspaceFeatures `db:"features"`
	CreatedAt time.Time         `db:"created_at"`
	UpdatedAt time.Time         `db:"updated_at"`
}

type Company struct {
	ID        int               `db:"id"`
	Name      string            `db:"name"`
	Slug      string            `db:"slug"`
	Features  WorkspaceFeatures `db:"features"`
	CreatedAt time.Time         `db:"created_at"`
	UpdatedAt time.Time         `db:"updated_at"`
}

type WorkspaceFeatures struct {
	AdvancedReports    bool `json:"advanced_reports"`
	BulkExport         bool `json:"bulk_export"`
	APIAccess          bool `json:"api_access"`
	SSOEnabled         bool `json:"sso_enabled"`
	MaxUsersPerCompany int  `json:"max_users_per_company"`
	AuditRetentionDays int  `json:"audit_retention_days"`
}

func (f WorkspaceFeatures) Value() (driver.Value, error) {
	return json.Marshal(f)
}

func (f *WorkspaceFeatures) Scan(value interface{}) error {
	if value == nil {
		*f = WorkspaceFeatures{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, f)
	case string:
		return json.Unmarshal([]byte(v), f)
	default:
		return nil
	}
}
