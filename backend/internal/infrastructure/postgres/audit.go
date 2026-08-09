package postgres

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"sena-iam-api/internal/application"
	"sena-iam-api/internal/domain"
)

type auditCursor struct {
	AttemptedAt string `json:"attempted_at"`
	ID          string `json:"id"`
}

func (r *Repository) LoginAudit(ctx context.Context, filters application.AuditFilters) (domain.AuditPage, error) {
	params := []any{}
	conditions := []string{}
	if filters.UserID != "" {
		params = append(params, filters.UserID)
		conditions = append(conditions, fmt.Sprintf("al.user_id = $%d::uuid", len(params)))
	}
	if filters.Email != "" {
		params = append(params, filters.Email)
		conditions = append(conditions, fmt.Sprintf("lower(al.email_attempted) = $%d", len(params)))
	}
	if filters.Outcome != "" {
		params = append(params, filters.Outcome)
		conditions = append(conditions, fmt.Sprintf("al.outcome = $%d", len(params)))
	}
	if filters.From != nil {
		params = append(params, *filters.From)
		conditions = append(conditions, fmt.Sprintf("al.attempted_at >= $%d", len(params)))
	}
	if filters.To != nil {
		params = append(params, *filters.To)
		conditions = append(conditions, fmt.Sprintf("al.attempted_at < $%d", len(params)))
	}
	if cursor := decodeAuditCursor(filters.Cursor); cursor != nil {
		params = append(params, cursor.AttemptedAt, cursor.ID)
		conditions = append(conditions, fmt.Sprintf("(al.attempted_at, al.id) < ($%d::timestamptz, $%d::uuid)", len(params)-1, len(params)))
	}
	whereSQL := ""
	if len(conditions) > 0 {
		whereSQL = "WHERE " + strings.Join(conditions, " AND ")
	}
	params = append(params, filters.Limit+1)
	rows, err := r.pool.Query(ctx, `SELECT al.id::text, al.user_id::text, al.email_attempted, al.outcome, al.ip_address, al.user_agent, al.attempted_at
       FROM identity_audit.audit_login al
       `+whereSQL+`
      ORDER BY al.attempted_at DESC, al.id DESC
      LIMIT $`+fmt.Sprint(len(params)), params...)
	if err != nil {
		return domain.AuditPage{}, err
	}
	defer rows.Close()
	items := []domain.AuditLogin{}
	for rows.Next() {
		var item domain.AuditLogin
		var userID, ip, ua sql.NullString
		var attempted sql.NullTime
		if err := rows.Scan(&item.ID, &userID, &item.EmailAttempted, &item.Outcome, &ip, &ua, &attempted); err != nil {
			return domain.AuditPage{}, err
		}
		item.UserID = nullStringPtr(userID)
		item.IPAddress = nullStringPtr(ip)
		item.UserAgent = nullStringPtr(ua)
		item.AttemptedAt = nullTimePtr(attempted)
		items = append(items, item)
	}
	var next *string
	if len(items) > filters.Limit {
		last := items[filters.Limit-1]
		items = items[:filters.Limit]
		if last.AttemptedAt != nil {
			encoded := encodeAuditCursor(auditCursor{AttemptedAt: last.AttemptedAt.Format("2006-01-02T15:04:05.999999Z07:00"), ID: last.ID})
			next = &encoded
		}
	}
	return domain.AuditPage{Data: items, Pagination: domain.AuditPagination{Limit: filters.Limit, NextCursor: next}}, nil
}

func encodeAuditCursor(cursor auditCursor) string {
	data, _ := json.Marshal(cursor)
	return base64.RawURLEncoding.EncodeToString(data)
}

func decodeAuditCursor(value string) *auditCursor {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil
	}
	var cursor auditCursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil
	}
	return &cursor
}