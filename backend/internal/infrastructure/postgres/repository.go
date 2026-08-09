package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"sena-iam-api/internal/application"
	"sena-iam-api/internal/domain"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, databaseURL string) (*Repository, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 10
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Repository{pool: pool}, nil
}

func (r *Repository) Close() { r.pool.Close() }

func (r *Repository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func (r *Repository) BootstrapDemoUsers(ctx context.Context, passwordHash, trainingCenterID string) error {
	users := []struct {
		ID, Email, FirstName, LastName, ActorType string
		ActorID                                  *string
		RoleName                                 string
		TrainingCenterID                         *string
	}{
		{ID: "c0000000-0000-0000-0000-000000000001", Email: "admin@sena.edu.co", FirstName: "System", LastName: "Admin", ActorType: "USER", RoleName: "SYSTEM_ADMIN"},
		{ID: "c0000000-0000-0000-0000-000000000002", Email: "director@sena.edu.co", FirstName: "Carlos", LastName: "Torres", ActorType: "USER", RoleName: "CENTER_DIRECTOR", TrainingCenterID: &trainingCenterID},
		{ID: "c0000000-0000-0000-0000-000000000003", Email: "coordinador@sena.edu.co", FirstName: "Maria", LastName: "Garcia", ActorType: "USER", RoleName: "COORDINATOR", TrainingCenterID: &trainingCenterID},
		{ID: "c0000000-0000-0000-0000-000000000004", Email: "instructor@sena.edu.co", FirstName: "Juan", LastName: "Perez", ActorType: "INSTRUCTOR", ActorID: strPtr("22222222-2222-2222-2222-222222222222"), RoleName: "INSTRUCTOR", TrainingCenterID: &trainingCenterID},
		{ID: "c0000000-0000-0000-0000-000000000005", Email: "aprendiz@sena.edu.co", FirstName: "Ana", LastName: "Lopez", ActorType: "LEARNER", ActorID: strPtr("33333333-3333-3333-3333-333333333333"), RoleName: "LEARNER", TrainingCenterID: &trainingCenterID},
		{ID: "c0000000-0000-0000-0000-000000000006", Email: "administrativo@sena.edu.co", FirstName: "Sofia", LastName: "Morales", ActorType: "USER", RoleName: "ADMIN_STAFF"},
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer rollback(ctx, tx)
	for _, user := range users {
		_, err := tx.Exec(ctx, `INSERT INTO identity."user" (id, email, password_hash, first_name, last_name, actor_type, actor_id, is_active)
         VALUES ($1, $2, $3, $4, $5, $6, $7::uuid, true)
         ON CONFLICT (id) DO UPDATE SET
           email = EXCLUDED.email,
           password_hash = EXCLUDED.password_hash,
           first_name = EXCLUDED.first_name,
           last_name = EXCLUDED.last_name,
           actor_type = EXCLUDED.actor_type,
           actor_id = EXCLUDED.actor_id,
           is_active = true,
           updated_at = now()`, user.ID, user.Email, passwordHash, user.FirstName, user.LastName, user.ActorType, user.ActorID)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO rbac.user_role (user_id, role_id, training_center_id, assigned_by)
     SELECT $1::uuid, r.id, $3::uuid, 'c0000000-0000-0000-0000-000000000001'::uuid
       FROM rbac.role r
      WHERE r.name = $2
        AND NOT EXISTS (
          SELECT 1
            FROM rbac.user_role ur
           WHERE ur.user_id = $1::uuid
             AND ur.role_id = r.id
             AND COALESCE(ur.training_center_id::text, '') = COALESCE($3::uuid::text, '')
        )`, user.ID, user.RoleName, user.TrainingCenterID)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, userSelect()+` WHERE lower(email) = $1`, email)
	user, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return user, err
}

func (r *Repository) GetUserByID(ctx context.Context, id string, withRoles bool) (*domain.User, error) {
	if !withRoles {
		row := r.pool.QueryRow(ctx, userSelect()+` WHERE id = $1::uuid`, id)
		user, err := scanUser(row)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return user, err
	}
	row := r.pool.QueryRow(ctx, `SELECT u.id::text, u.email, u.password_hash, u.first_name, u.last_name,
            u.actor_type, u.actor_id::text, u.is_active, u.last_login_at, u.failed_attempts,
            u.locked_until, u.created_at, u.updated_at,
            COALESCE(
              jsonb_agg(DISTINCT jsonb_build_object(
                'id', r.id,
                'name', r.name,
                'display_name', r.display_name,
                'training_center_id', ur.training_center_id,
                'assigned_at', ur.assigned_at,
                'expires_at', ur.expires_at
              )) FILTER (WHERE r.id IS NOT NULL),
              '[]'::jsonb
            ) AS roles
       FROM identity."user" u
       LEFT JOIN rbac.user_role ur
         ON ur.user_id = u.id
        AND (ur.expires_at IS NULL OR ur.expires_at > now())
       LEFT JOIN rbac.role r ON r.id = ur.role_id
      WHERE u.id = $1::uuid
      GROUP BY u.id`, id)
	user, err := scanUserWithRoles(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return user, err
}

func (r *Repository) UpdateLoginSuccess(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE identity."user" SET failed_attempts = 0, locked_until = NULL, last_login_at = now(), updated_at = now() WHERE id = $1::uuid`, userID)
	return err
}

func (r *Repository) UpdateFailedLogin(ctx context.Context, userID string, attempts int, lockedUntil *time.Time) error {
	_, err := r.pool.Exec(ctx, `UPDATE identity."user" SET failed_attempts = $1, locked_until = $2, updated_at = now() WHERE id = $3::uuid`, attempts, lockedUntil, userID)
	return err
}

func (r *Repository) AuditLogin(ctx context.Context, userID *string, email, outcome string, ipAddress, userAgent *string) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO identity_audit.audit_login (user_id, email_attempted, outcome, ip_address, user_agent)
       VALUES ($1::uuid, $2, $3, $4, $5)`, userID, email, outcome, ipAddress, userAgent)
	return err
}

func (r *Repository) GetUserAccess(ctx context.Context, userID string) (domain.UserAccess, error) {
	rolesRows, err := r.pool.Query(ctx, `SELECT r.name, ur.training_center_id::text
       FROM rbac.user_role ur
       JOIN rbac.role r ON r.id = ur.role_id
      WHERE ur.user_id = $1::uuid
        AND (ur.expires_at IS NULL OR ur.expires_at > now())
      ORDER BY r.name`, userID)
	if err != nil {
		return domain.UserAccess{}, err
	}
	defer rolesRows.Close()
	roles := []domain.AccessRole{}
	for rolesRows.Next() {
		var role domain.AccessRole
		var center sql.NullString
		if err := rolesRows.Scan(&role.Name, &center); err != nil {
			return domain.UserAccess{}, err
		}
		role.TrainingCenterID = nullStringPtr(center)
		roles = append(roles, role)
	}
	featuresRows, err := r.pool.Query(ctx, `WITH role_features AS (
       SELECT f.code AS feature_code, rf.scope_type, m.code AS module_code, 1 AS precedence
         FROM rbac.user_role ur
         JOIN rbac.role_feature rf ON rf.role_id = ur.role_id
         JOIN rbac_catalog.feature f ON f.id = rf.feature_id
         JOIN rbac_catalog.module m ON m.id = f.module_id
        WHERE ur.user_id = $1::uuid
          AND f.is_active = true
          AND m.is_active = true
          AND (ur.expires_at IS NULL OR ur.expires_at > now())
     ),
     additive_overrides AS (
       SELECT f.code AS feature_code, uso.scope_type, m.code AS module_code, 2 AS precedence
         FROM rbac.user_scope_override uso
         JOIN rbac_catalog.feature f ON f.id = uso.feature_id
         JOIN rbac_catalog.module m ON m.id = f.module_id
        WHERE uso.user_id = $1::uuid
          AND uso.is_allowed = true
          AND (uso.expires_at IS NULL OR uso.expires_at > now())
     ),
     denied AS (
       SELECT f.code AS feature_code
         FROM rbac.user_scope_override uso
         JOIN rbac_catalog.feature f ON f.id = uso.feature_id
        WHERE uso.user_id = $1::uuid
          AND uso.is_allowed = false
          AND (uso.expires_at IS NULL OR uso.expires_at > now())
     ),
     all_features AS (
       SELECT * FROM role_features
       UNION ALL
       SELECT * FROM additive_overrides
     )
     SELECT DISTINCT ON (feature_code) feature_code, scope_type, module_code
       FROM all_features
      WHERE feature_code NOT IN (SELECT feature_code FROM denied)
      ORDER BY feature_code, precedence DESC, scope_type`, userID)
	if err != nil {
		return domain.UserAccess{}, err
	}
	defer featuresRows.Close()
	features := []domain.AccessFeature{}
	for featuresRows.Next() {
		var feature domain.AccessFeature
		if err := featuresRows.Scan(&feature.FeatureCode, &feature.ScopeType, &feature.ModuleCode); err != nil {
			return domain.UserAccess{}, err
		}
		features = append(features, feature)
	}
	return domain.UserAccess{Roles: roles, Features: features}, nil
}

func (r *Repository) ListUsers(ctx context.Context, filters application.UserFilters) (domain.UserPage, error) {
	whereSQL, params := buildUserFilters(filters)
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*)::int AS total FROM identity."user" u `+whereSQL, params...).Scan(&total); err != nil {
		return domain.UserPage{}, err
	}
	sort := parseSort(filters.Sort)
	limit := filters.PageSize
	offset := (filters.Page - 1) * filters.PageSize
	params = append(params, limit, offset)
	rows, err := r.pool.Query(ctx, `SELECT u.id::text, u.email, u.password_hash, u.first_name, u.last_name,
            u.actor_type, u.actor_id::text, u.is_active, u.last_login_at, u.failed_attempts,
            u.locked_until, u.created_at, u.updated_at,
            COALESCE(
              jsonb_agg(DISTINCT jsonb_build_object(
                'id', r.id,
                'name', r.name,
                'display_name', r.display_name,
                'training_center_id', ur.training_center_id,
                'assigned_at', ur.assigned_at,
                'expires_at', ur.expires_at
              )) FILTER (WHERE r.id IS NOT NULL),
              '[]'::jsonb
            ) AS roles
       FROM identity."user" u
       LEFT JOIN rbac.user_role ur
         ON ur.user_id = u.id
        AND (ur.expires_at IS NULL OR ur.expires_at > now())
       LEFT JOIN rbac.role r ON r.id = ur.role_id
       `+whereSQL+`
      GROUP BY u.id
      ORDER BY `+sort+`
      LIMIT $`+fmt.Sprint(len(params)-1)+` OFFSET $`+fmt.Sprint(len(params)), params...)
	if err != nil {
		return domain.UserPage{}, err
	}
	defer rows.Close()
	users := []domain.User{}
	for rows.Next() {
		user, err := scanUserWithRoles(rows)
		if err != nil {
			return domain.UserPage{}, err
		}
		users = append(users, *user)
	}
	return domain.UserPage{Data: users, Pagination: domain.Pagination{Page: filters.Page, PageSize: filters.PageSize, TotalItems: total, TotalPages: int(math.Ceil(float64(total) / float64(filters.PageSize)))}}, nil
}

func userSelect() string {
	return `SELECT id::text, email, password_hash, first_name, last_name, actor_type, actor_id::text, is_active, last_login_at, failed_attempts, locked_until, created_at, updated_at FROM identity."user"`
}

type scanner interface{ Scan(dest ...any) error }

func scanUser(row scanner) (*domain.User, error) {
	var user domain.User
	var actorID sql.NullString
	var lastLogin, lockedUntil, createdAt, updatedAt sql.NullTime
	if err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.ActorType, &actorID, &user.IsActive, &lastLogin, &user.FailedAttempts, &lockedUntil, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	user.ActorID = nullStringPtr(actorID)
	user.LastLoginAt = nullTimePtr(lastLogin)
	user.LockedUntil = nullTimePtr(lockedUntil)
	user.CreatedAt = nullTimePtr(createdAt)
	user.UpdatedAt = nullTimePtr(updatedAt)
	user.FullName = strings.TrimSpace(user.FirstName + " " + user.LastName)
	user.Roles = []domain.UserRole{}
	return &user, nil
}

func scanUserWithRoles(row scanner) (*domain.User, error) {
	var user domain.User
	var actorID sql.NullString
	var lastLogin, lockedUntil, createdAt, updatedAt sql.NullTime
	var rolesRaw []byte
	if err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.ActorType, &actorID, &user.IsActive, &lastLogin, &user.FailedAttempts, &lockedUntil, &createdAt, &updatedAt, &rolesRaw); err != nil {
		return nil, err
	}
	user.ActorID = nullStringPtr(actorID)
	user.LastLoginAt = nullTimePtr(lastLogin)
	user.LockedUntil = nullTimePtr(lockedUntil)
	user.CreatedAt = nullTimePtr(createdAt)
	user.UpdatedAt = nullTimePtr(updatedAt)
	user.FullName = strings.TrimSpace(user.FirstName + " " + user.LastName)
	user.Roles = []domain.UserRole{}
	_ = json.Unmarshal(rolesRaw, &user.Roles)
	return &user, nil
}

func nullStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func nullTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}

func strPtr(value string) *string { return &value }

func rollback(ctx context.Context, tx pgx.Tx) { _ = tx.Rollback(ctx) }

func isUnique(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func buildUserFilters(filters application.UserFilters) (string, []any) {
	conditions := []string{}
	params := []any{}
	if filters.IsActive != nil {
		params = append(params, *filters.IsActive)
		conditions = append(conditions, fmt.Sprintf("u.is_active = $%d", len(params)))
	}
	if filters.Role != "" {
		params = append(params, filters.Role)
		conditions = append(conditions, fmt.Sprintf(`EXISTS (SELECT 1 FROM rbac.user_role fur JOIN rbac.role fr ON fr.id = fur.role_id WHERE fur.user_id = u.id AND fr.name = $%d)`, len(params)))
	}
	if filters.TrainingCenterID != "" {
		params = append(params, filters.TrainingCenterID)
		conditions = append(conditions, fmt.Sprintf(`EXISTS (SELECT 1 FROM rbac.user_role fuc WHERE fuc.user_id = u.id AND fuc.training_center_id = $%d::uuid)`, len(params)))
	}
	if filters.Search != "" {
		params = append(params, "%"+strings.ToLower(filters.Search)+"%")
		conditions = append(conditions, fmt.Sprintf(`(lower(u.email) LIKE $%d OR lower(u.first_name || ' ' || u.last_name) LIKE $%d)`, len(params), len(params)))
	}
	if len(conditions) == 0 {
		return "", params
	}
	return "WHERE " + strings.Join(conditions, " AND "), params
}

func parseSort(raw string) string {
	allowed := map[string]string{"email": "u.email", "first_name": "u.first_name", "last_name": "u.last_name", "created_at": "u.created_at", "last_login_at": "u.last_login_at"}
	parts := strings.Split(raw, ":")
	if len(parts) == 0 || allowed[parts[0]] == "" {
		return "u.created_at DESC"
	}
	direction := "ASC"
	if len(parts) > 1 && strings.ToLower(parts[1]) == "desc" {
		direction = "DESC"
	}
	return allowed[parts[0]] + " " + direction
}