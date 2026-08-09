package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"sena-iam-api/internal/application"
	"sena-iam-api/internal/domain"
)

func (r *Repository) CreateRefreshToken(ctx context.Context, userID, tokenHash string, deviceHint, ipAddress *string, ttlDays int) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO session.refresh_token (user_id, token_hash, device_hint, ip_address, expires_at)
     VALUES ($1::uuid, $2, $3, $4, now() + ($5::int * interval '1 day'))`, userID, tokenHash, deviceHint, ipAddress, ttlDays)
	return err
}

func (r *Repository) FindRefreshToken(ctx context.Context, tokenHash string) (*domain.RefreshTokenRecord, error) {
	var record domain.RefreshTokenRecord
	err := r.pool.QueryRow(ctx, `SELECT id::text, user_id::text, expires_at
       FROM session.refresh_token
      WHERE token_hash = $1
        AND is_revoked = false
        AND expires_at > now()`, tokenHash).Scan(&record.ID, &record.UserID, &record.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &record, err
}

func (r *Repository) RevokeRefreshToken(ctx context.Context, userID, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `UPDATE session.refresh_token SET is_revoked = true, revoked_at = now() WHERE token_hash = $1 AND user_id = $2::uuid`, tokenHash, userID)
	return err
}

func (r *Repository) RevokeAllRefreshTokens(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE session.refresh_token SET is_revoked = true, revoked_at = now() WHERE user_id = $1::uuid AND is_revoked = false`, userID)
	return err
}

func (r *Repository) InsertPasswordReset(ctx context.Context, userID, tokenHash string, ipAddress *string) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO session.password_reset_request (user_id, token_hash, ip_address)
       VALUES ($1::uuid, $2, $3)`, userID, tokenHash, ipAddress)
	return err
}

func (r *Repository) FindPasswordReset(ctx context.Context, tokenHash string) (*domain.PasswordResetRequest, error) {
	var req domain.PasswordResetRequest
	err := r.pool.QueryRow(ctx, `SELECT pr.id::text, pr.user_id::text, u.email, u.first_name, u.last_name
       FROM session.password_reset_request pr
       JOIN identity."user" u ON u.id = pr.user_id
      WHERE pr.token_hash = $1
        AND pr.is_used = false
        AND pr.expires_at > now()
      ORDER BY pr.requested_at DESC
      LIMIT 1`, tokenHash).Scan(&req.ID, &req.UserID, &req.Email, &req.FirstName, &req.LastName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &req, err
}

func (r *Repository) ConfirmPasswordReset(ctx context.Context, requestID, userID, passwordHash string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer rollback(ctx, tx)
	if _, err := tx.Exec(ctx, `UPDATE identity."user" SET password_hash = $1, failed_attempts = 0, locked_until = NULL, updated_at = now() WHERE id = $2::uuid`, passwordHash, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE session.password_reset_request SET is_used = true WHERE id = $1::uuid`, requestID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE session.refresh_token SET is_revoked = true, revoked_at = now() WHERE user_id = $1::uuid AND is_revoked = false`, userID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) CreateUser(ctx context.Context, input application.CreateUserInput, passwordHash, assignedBy string) (string, string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", "", err
	}
	defer rollback(ctx, tx)
	var id, email string
	err = tx.QueryRow(ctx, `INSERT INTO identity."user" (email, password_hash, first_name, last_name, actor_type, actor_id)
         VALUES ($1, $2, $3, $4, $5, $6::uuid)
         RETURNING id::text, email`, input.Email, passwordHash, input.FirstName, input.LastName, input.ActorType, input.ActorID).Scan(&id, &email)
	if err != nil {
		if isUnique(err) {
			return "", "", domain.Conflict("USER_ALREADY_EXISTS", "Ya existe un usuario con ese correo")
		}
		return "", "", err
	}
	if input.InitialRole != nil && strings.TrimSpace(*input.InitialRole) != "" {
		if _, err := assignRoleTx(ctx, tx, id, *input.InitialRole, input.TrainingCenterID, assignedBy, nil); err != nil {
			return "", "", err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return "", "", err
	}
	return id, email, nil
}

func (r *Repository) UpdateUser(ctx context.Context, id string, input application.UpdateUserInput) (*domain.User, error) {
	sets := []string{}
	params := []any{}
	add := func(field string, value any) {
		params = append(params, value)
		sets = append(sets, fmt.Sprintf("%s = $%d", field, len(params)))
	}
	if input.FirstName != nil {
		add("first_name", *input.FirstName)
	}
	if input.LastName != nil {
		add("last_name", *input.LastName)
	}
	if input.ActorType != nil {
		add("actor_type", *input.ActorType)
	}
	if input.ActorID != nil {
		add("actor_id", *input.ActorID)
	}
	if input.IsActive != nil {
		add("is_active", *input.IsActive)
	}
	if len(sets) > 0 {
		params = append(params, id)
		result, err := r.pool.Exec(ctx, `UPDATE identity."user" SET `+strings.Join(sets, ", ")+`, updated_at = now() WHERE id = $`+fmt.Sprint(len(params))+`::uuid`, params...)
		if err != nil {
			return nil, err
		}
		if result.RowsAffected() == 0 {
			return nil, domain.NotFound("USER_NOT_FOUND", "Usuario no encontrado")
		}
	}
	return r.GetUserByID(ctx, id, true)
}

func (r *Repository) DeactivateUser(ctx context.Context, id string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer rollback(ctx, tx)
	result, err := tx.Exec(ctx, `UPDATE identity."user" SET is_active = false, updated_at = now() WHERE id = $1::uuid`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.NotFound("USER_NOT_FOUND", "Usuario no encontrado")
	}
	if _, err := tx.Exec(ctx, `UPDATE session.refresh_token SET is_revoked = true, revoked_at = now() WHERE user_id = $1::uuid AND is_revoked = false`, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) ListSessions(ctx context.Context, userID string) ([]domain.Session, error) {
	rows, err := r.pool.Query(ctx, `SELECT id::text, device_hint, ip_address, created_at, expires_at, is_revoked, revoked_at
       FROM session.refresh_token
      WHERE user_id = $1::uuid
        AND expires_at > now()
        AND is_revoked = false
      ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sessions := []domain.Session{}
	for rows.Next() {
		var item domain.Session
		var device, ip sql.NullString
		var created, expires, revoked sql.NullTime
		if err := rows.Scan(&item.ID, &device, &ip, &created, &expires, &item.IsRevoked, &revoked); err != nil {
			return nil, err
		}
		item.DeviceHint = nullStringPtr(device)
		item.IPAddress = nullStringPtr(ip)
		item.CreatedAt = nullTimePtr(created)
		item.ExpiresAt = nullTimePtr(expires)
		item.RevokedAt = nullTimePtr(revoked)
		sessions = append(sessions, item)
	}
	return sessions, nil
}

func (r *Repository) RevokeSession(ctx context.Context, userID, sessionID string) error {
	result, err := r.pool.Exec(ctx, `UPDATE session.refresh_token SET is_revoked = true, revoked_at = now()
      WHERE id = $1::uuid AND user_id = $2::uuid
      RETURNING id`, sessionID, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.NotFound("SESSION_NOT_FOUND", "Sesion no encontrada")
	}
	return nil
}

func assignRoleTx(ctx context.Context, tx pgx.Tx, userID, roleName string, trainingCenterID *string, assignedBy string, expiresAt *time.Time) (*domain.AssignedRole, error) {
	var role domain.AssignedRole
	var center sql.NullString
	var assignedAt, expires sql.NullTime
	err := tx.QueryRow(ctx, `INSERT INTO rbac.user_role (user_id, role_id, training_center_id, assigned_by, expires_at)
     SELECT $1::uuid, r.id, $3::uuid, $4::uuid, $5::timestamptz
       FROM rbac.role r
      WHERE r.name = $2
        AND NOT EXISTS (
          SELECT 1
            FROM rbac.user_role ur
           WHERE ur.user_id = $1::uuid
             AND ur.role_id = r.id
             AND COALESCE(ur.training_center_id::text, '') = COALESCE($3::uuid::text, '')
        )
      RETURNING id::text, user_id::text, role_id::text, $2::text AS role_name, training_center_id::text, assigned_by::text, assigned_at, expires_at`, userID, roleName, trainingCenterID, assignedBy, expiresAt).
		Scan(&role.ID, &role.UserID, &role.RoleID, &role.RoleName, &center, &role.AssignedBy, &assignedAt, &expires)
	if err == nil {
		role.TrainingCenterID = nullStringPtr(center)
		role.AssignedAt = nullTimePtr(assignedAt)
		role.ExpiresAt = nullTimePtr(expires)
		return &role, nil
	}
	var roleID string
	checkErr := tx.QueryRow(ctx, `SELECT id::text FROM rbac.role WHERE name = $1`, roleName).Scan(&roleID)
	if errors.Is(checkErr, pgx.ErrNoRows) {
		return nil, domain.NotFound("ROLE_NOT_FOUND", "Rol no encontrado")
	}
	return nil, domain.Conflict("ROLE_ALREADY_ASSIGNED", "El usuario ya tiene ese rol asignado")
}