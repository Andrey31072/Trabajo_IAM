package domain

import "time"

type User struct {
	ID             string     `json:"id"`
	Email          string     `json:"email"`
	PasswordHash   string     `json:"-"`
	FirstName      string     `json:"first_name"`
	LastName       string     `json:"last_name"`
	FullName       string     `json:"full_name"`
	ActorType      string     `json:"actor_type"`
	ActorID        *string    `json:"actor_id"`
	IsActive       bool       `json:"is_active"`
	LastLoginAt    *time.Time `json:"last_login_at"`
	FailedAttempts int        `json:"-"`
	LockedUntil    *time.Time `json:"-"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
	Roles          []UserRole `json:"roles"`
}

type UserRole struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	DisplayName      string     `json:"display_name"`
	TrainingCenterID *string    `json:"training_center_id"`
	AssignedAt       *time.Time `json:"assigned_at"`
	ExpiresAt        *time.Time `json:"expires_at"`
}

type AssignedRole struct {
	ID               string     `json:"id"`
	UserID           string     `json:"user_id"`
	RoleID           string     `json:"role_id"`
	RoleName         string     `json:"role_name"`
	TrainingCenterID *string    `json:"training_center_id"`
	AssignedBy       string     `json:"assigned_by"`
	AssignedAt       *time.Time `json:"assigned_at"`
	ExpiresAt        *time.Time `json:"expires_at"`
}

type Role struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	DisplayName  string        `json:"display_name"`
	Description  *string       `json:"description"`
	IsSystemRole bool          `json:"is_system_role"`
	CreatedAt    *time.Time    `json:"created_at"`
	Features     []RoleFeature `json:"features"`
}

type RoleFeature struct {
	FeatureCode string `json:"feature_code"`
	FeatureName string `json:"feature_name"`
	ActionLevel string `json:"action_level"`
	ScopeType   string `json:"scope_type"`
	ModuleCode  string `json:"module_code,omitempty"`
}

type Module struct {
	ID           string          `json:"id"`
	Code         string          `json:"code"`
	Name         string          `json:"name"`
	Description  *string         `json:"description"`
	DisplayOrder int             `json:"display_order"`
	IconKey      *string         `json:"icon_key"`
	Features     []ModuleFeature `json:"features"`
}

type ModuleFeature struct {
	ID          string  `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	ActionLevel string  `json:"action_level"`
	IsActive    bool    `json:"is_active"`
}

type Session struct {
	ID         string     `json:"id"`
	DeviceHint *string    `json:"device_hint"`
	IPAddress  *string    `json:"ip_address"`
	CreatedAt  *time.Time `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
	IsRevoked  bool       `json:"is_revoked"`
	RevokedAt  *time.Time `json:"revoked_at"`
}

type ScopeOverride struct {
	ID          string     `json:"id"`
	FeatureCode string     `json:"feature_code"`
	ScopeType   string     `json:"scope_type"`
	IsAllowed   bool       `json:"is_allowed"`
	Reason      string     `json:"reason"`
	GrantedBy   string     `json:"granted_by"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   *time.Time `json:"created_at"`
}

type AuditLogin struct {
	ID             string     `json:"id"`
	UserID         *string    `json:"user_id"`
	EmailAttempted string     `json:"email_attempted"`
	Outcome        string     `json:"outcome"`
	IPAddress      *string    `json:"ip_address"`
	UserAgent      *string    `json:"user_agent"`
	AttemptedAt    *time.Time `json:"attempted_at"`
}

type AccessRole struct {
	Name             string
	TrainingCenterID *string
}

type AccessFeature struct {
	FeatureCode string
	ScopeType   string
	ModuleCode  string
}

type UserAccess struct {
	Roles    []AccessRole
	Features []AccessFeature
}

type AuthUser struct {
	ID               string   `json:"id"`
	Email            string   `json:"email"`
	FullName         string   `json:"full_name"`
	Roles            []string `json:"roles"`
	TrainingCenterID *string  `json:"training_center_id"`
	Features         []string `json:"features"`
}

type Me struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	FullName  string     `json:"full_name"`
	ActorType string     `json:"actor_type"`
	ActorID   *string    `json:"actor_id"`
	Roles     []UserRole `json:"roles"`
	Features  []string   `json:"features"`
	Modules   []string   `json:"modules"`
}

type AuthResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int      `json:"expires_in"`
	User         AuthUser `json:"user"`
}

type RefreshTokenRecord struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
}

type PasswordResetRequest struct {
	ID        string
	UserID    string
	Email     string
	FirstName string
	LastName  string
}

type UserPage struct {
	Data       []User     `json:"data"`
	Pagination Pagination `json:"pagination"`
}

type AuditPage struct {
	Data       []AuditLogin    `json:"data"`
	Pagination AuditPagination `json:"pagination"`
}

type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

type AuditPagination struct {
	Limit      int     `json:"limit"`
	NextCursor *string `json:"next_cursor"`
}