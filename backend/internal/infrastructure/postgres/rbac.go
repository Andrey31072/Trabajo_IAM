package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"sena-iam-api/internal/application"
	"sena-iam-api/internal/domain"
)

func (r *Repository) AssignRole(ctx context.Context, userID, roleName string, trainingCenterID *string, assignedBy string, expiresAt *time.Time) (*domain.AssignedRole, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(ctx, tx)
	exists := ""
	if err := tx.QueryRow(ctx, `SELECT id::text FROM identity."user" WHERE id = $1::uuid`, userID).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("USER_NOT_FOUND", "Usuario no encontrado")
		}
		return nil, err
	}
	role, err := assignRoleTx(ctx, tx, userID, roleName, trainingCenterID, assignedBy, expiresAt)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return role, nil
}

func (r *Repository) RevokeRole(ctx context.Context, userID, roleName string) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM rbac.user_role ur
      USING rbac.role r
      WHERE ur.role_id = r.id
        AND ur.user_id = $1::uuid
        AND r.name = $2`, userID, roleName)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.NotFound("ROLE_ASSIGNMENT_NOT_FOUND", "Asignacion de rol no encontrada")
	}
	return nil
}

func (r *Repository) ListScopeOverrides(ctx context.Context, userID string) ([]domain.ScopeOverride, error) {
	rows, err := r.pool.Query(ctx, `SELECT uso.id::text, f.code AS feature_code, uso.scope_type, uso.is_allowed, uso.reason,
            uso.granted_by::text, uso.expires_at, uso.created_at
       FROM rbac.user_scope_override uso
       JOIN rbac_catalog.feature f ON f.id = uso.feature_id
      WHERE uso.user_id = $1::uuid
        AND (uso.expires_at IS NULL OR uso.expires_at > now())
      ORDER BY uso.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.ScopeOverride{}
	for rows.Next() {
		var item domain.ScopeOverride
		var expires, created sql.NullTime
		if err := rows.Scan(&item.ID, &item.FeatureCode, &item.ScopeType, &item.IsAllowed, &item.Reason, &item.GrantedBy, &expires, &created); err != nil {
			return nil, err
		}
		item.ExpiresAt = nullTimePtr(expires)
		item.CreatedAt = nullTimePtr(created)
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) CreateScopeOverride(ctx context.Context, input application.ScopeOverrideInput, grantedBy string) (*domain.ScopeOverride, error) {
	var item domain.ScopeOverride
	var expires, created sql.NullTime
	err := r.pool.QueryRow(ctx, `INSERT INTO rbac.user_scope_override (user_id, feature_id, scope_type, is_allowed, reason, granted_by, expires_at)
     SELECT $1::uuid, f.id, $2, $3, $4, $5::uuid, $6::timestamptz
       FROM rbac_catalog.feature f
      WHERE f.code = $7
      RETURNING id::text, $7::text AS feature_code, scope_type, is_allowed, reason, granted_by::text, expires_at, created_at`,
		input.UserID, input.ScopeType, input.IsAllowed, input.Reason, grantedBy, input.ExpiresAt, input.FeatureCode).
		Scan(&item.ID, &item.FeatureCode, &item.ScopeType, &item.IsAllowed, &item.Reason, &item.GrantedBy, &expires, &created)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.NotFound("FEATURE_NOT_FOUND", "Feature no encontrado")
	}
	if err != nil {
		return nil, err
	}
	item.ExpiresAt = nullTimePtr(expires)
	item.CreatedAt = nullTimePtr(created)
	return &item, nil
}

func (r *Repository) DeleteScopeOverride(ctx context.Context, userID, overrideID string) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM rbac.user_scope_override WHERE id = $1::uuid AND user_id = $2::uuid`, overrideID, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.NotFound("OVERRIDE_NOT_FOUND", "Override no encontrado")
	}
	return nil
}

func (r *Repository) ListRoles(ctx context.Context) ([]domain.Role, error) {
	rows, err := r.pool.Query(ctx, `SELECT r.id::text, r.name, r.display_name, r.description, r.is_system_role, r.created_at,
            COALESCE(
              jsonb_agg(jsonb_build_object(
                'feature_code', f.code,
                'feature_name', f.name,
                'action_level', f.action_level,
                'scope_type', rf.scope_type,
                'module_code', m.code
              ) ORDER BY m.display_order, f.code) FILTER (WHERE f.id IS NOT NULL),
              '[]'::jsonb
            ) AS features
       FROM rbac.role r
       LEFT JOIN rbac.role_feature rf ON rf.role_id = r.id
       LEFT JOIN rbac_catalog.feature f ON f.id = rf.feature_id AND f.is_active = true
       LEFT JOIN rbac_catalog.module m ON m.id = f.module_id
      GROUP BY r.id
      ORDER BY r.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	roles := []domain.Role{}
	for rows.Next() {
		var role domain.Role
		var description sql.NullString
		var created sql.NullTime
		var featuresRaw []byte
		if err := rows.Scan(&role.ID, &role.Name, &role.DisplayName, &description, &role.IsSystemRole, &created, &featuresRaw); err != nil {
			return nil, err
		}
		role.Description = nullStringPtr(description)
		role.CreatedAt = nullTimePtr(created)
		role.Features = []domain.RoleFeature{}
		_ = json.Unmarshal(featuresRaw, &role.Features)
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *Repository) GetRoleFeatures(ctx context.Context, roleID string) ([]domain.RoleFeature, error) {
	rows, err := r.pool.Query(ctx, `SELECT f.code AS feature_code, f.name AS feature_name, f.action_level, rf.scope_type
       FROM rbac.role_feature rf
       JOIN rbac_catalog.feature f ON f.id = rf.feature_id
      WHERE rf.role_id = $1::uuid
      ORDER BY f.code`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.RoleFeature{}
	for rows.Next() {
		var item domain.RoleFeature
		if err := rows.Scan(&item.FeatureCode, &item.FeatureName, &item.ActionLevel, &item.ScopeType); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		var id string
		if err := r.pool.QueryRow(ctx, `SELECT id::text FROM rbac.role WHERE id = $1::uuid`, roleID).Scan(&id); errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFound("ROLE_NOT_FOUND", "Rol no encontrado")
		}
	}
	return items, nil
}

func (r *Repository) ListModules(ctx context.Context) ([]domain.Module, error) {
	rows, err := r.pool.Query(ctx, `SELECT m.id::text, m.code, m.name, m.description, m.display_order, m.icon_key,
            COALESCE(
              jsonb_agg(jsonb_build_object(
                'id', f.id,
                'code', f.code,
                'name', f.name,
                'description', f.description,
                'action_level', f.action_level,
                'is_active', f.is_active
              ) ORDER BY f.code) FILTER (WHERE f.id IS NOT NULL),
              '[]'::jsonb
            ) AS features
       FROM rbac_catalog.module m
       LEFT JOIN rbac_catalog.feature f ON f.module_id = m.id
      WHERE m.is_active = true
      GROUP BY m.id
      ORDER BY m.display_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	modules := []domain.Module{}
	for rows.Next() {
		var module domain.Module
		var description, icon sql.NullString
		var featuresRaw []byte
		if err := rows.Scan(&module.ID, &module.Code, &module.Name, &description, &module.DisplayOrder, &icon, &featuresRaw); err != nil {
			return nil, err
		}
		module.Description = nullStringPtr(description)
		module.IconKey = nullStringPtr(icon)
		module.Features = []domain.ModuleFeature{}
		_ = json.Unmarshal(featuresRaw, &module.Features)
		modules = append(modules, module)
	}
	return modules, nil
}