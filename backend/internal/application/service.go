package application

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"sena-iam-api/internal/domain"
	"sena-iam-api/internal/infrastructure/security"
)

type Repository interface {
	Ping(ctx context.Context) error
	BootstrapDemoUsers(ctx context.Context, passwordHash, trainingCenterID string) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id string, withRoles bool) (*domain.User, error)
	UpdateLoginSuccess(ctx context.Context, userID string) error
	UpdateFailedLogin(ctx context.Context, userID string, attempts int, lockedUntil *time.Time) error
	AuditLogin(ctx context.Context, userID *string, email, outcome string, ipAddress, userAgent *string) error
	GetUserAccess(ctx context.Context, userID string) (domain.UserAccess, error)
	CreateRefreshToken(ctx context.Context, userID, tokenHash string, deviceHint, ipAddress *string, ttlDays int) error
	FindRefreshToken(ctx context.Context, tokenHash string) (*domain.RefreshTokenRecord, error)
	RevokeRefreshToken(ctx context.Context, userID, tokenHash string) error
	RevokeAllRefreshTokens(ctx context.Context, userID string) error
	InsertPasswordReset(ctx context.Context, userID, tokenHash string, ipAddress *string) error
	FindPasswordReset(ctx context.Context, tokenHash string) (*domain.PasswordResetRequest, error)
	ConfirmPasswordReset(ctx context.Context, requestID, userID, passwordHash string) error
	ListUsers(ctx context.Context, filters UserFilters) (domain.UserPage, error)
	CreateUser(ctx context.Context, input CreateUserInput, passwordHash, assignedBy string) (string, string, error)
	UpdateUser(ctx context.Context, id string, input UpdateUserInput) (*domain.User, error)
	DeactivateUser(ctx context.Context, id string) error
	ListSessions(ctx context.Context, userID string) ([]domain.Session, error)
	RevokeSession(ctx context.Context, userID, sessionID string) error
	AssignRole(ctx context.Context, userID, roleName string, trainingCenterID *string, assignedBy string, expiresAt *time.Time) (*domain.AssignedRole, error)
	RevokeRole(ctx context.Context, userID, roleName string) error
	ListScopeOverrides(ctx context.Context, userID string) ([]domain.ScopeOverride, error)
	CreateScopeOverride(ctx context.Context, input ScopeOverrideInput, grantedBy string) (*domain.ScopeOverride, error)
	DeleteScopeOverride(ctx context.Context, userID, overrideID string) error
	ListRoles(ctx context.Context) ([]domain.Role, error)
	GetRoleFeatures(ctx context.Context, roleID string) ([]domain.RoleFeature, error)
	ListModules(ctx context.Context) ([]domain.Module, error)
	LoginAudit(ctx context.Context, filters AuditFilters) (domain.AuditPage, error)
}

type TokenManager interface {
	TTLSeconds() int
	Sign(user domain.User, auth domain.AuthUser) (string, error)
}

type Mailer interface {
	PasswordReset(to, name, link string) error
	PasswordChanged(to, name string) error
	Welcome(to, name, temporaryPassword, appURL string) error
}

type Service struct {
	repo        Repository
	tokens      TokenManager
	mailer      Mailer
	appURL      string
	refreshDays int
	demoPass    string
	demoCenter  string
}

type RequestMeta struct {
	IPAddress  *string
	UserAgent  *string
	DeviceHint *string
}

type UserFilters struct {
	Page             int
	PageSize         int
	Sort             string
	Search           string
	Role             string
	IsActive         *bool
	TrainingCenterID string
}

type CreateUserInput struct {
	Email            string
	FirstName        string
	LastName         string
	ActorType        string
	ActorID          *string
	InitialRole      *string
	TrainingCenterID *string
}

type UpdateUserInput struct {
	FirstName *string
	LastName  *string
	ActorType *string
	ActorID   *string
	IsActive  *bool
}

type ScopeOverrideInput struct {
	UserID      string
	FeatureCode string
	ScopeType   string
	IsAllowed   bool
	Reason      string
	ExpiresAt   *time.Time
}

type AuditFilters struct {
	Limit   int
	Cursor  string
	UserID  string
	Email   string
	Outcome string
	From    *time.Time
	To      *time.Time
}

func New(repo Repository, tokens TokenManager, mailer Mailer, appURL string, refreshDays int, demoPass, demoCenter string) *Service {
	return &Service{repo: repo, tokens: tokens, mailer: mailer, appURL: appURL, refreshDays: refreshDays, demoPass: demoPass, demoCenter: demoCenter}
}

func (s *Service) Health(ctx context.Context) error { return s.repo.Ping(ctx) }

func (s *Service) Bootstrap(ctx context.Context) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(s.demoPass), 12)
	if err != nil {
		return err
	}
	return s.repo.BootstrapDemoUsers(ctx, string(hash), s.demoCenter)
}

func (s *Service) Login(ctx context.Context, email, password string, meta RequestMeta) (*domain.AuthResponse, error) {
	email = NormalizeEmail(email)
	if email == "" || password == "" {
		return nil, domain.BadRequest("Correo y contrasena son obligatorios")
	}
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		_ = s.repo.AuditLogin(ctx, nil, email, "USER_NOT_FOUND", meta.IPAddress, meta.UserAgent)
		return nil, domain.Unauthorized("INVALID_CREDENTIALS", "Email o contrasena incorrectos")
	}
	if !user.IsActive {
		_ = s.repo.AuditLogin(ctx, &user.ID, email, "ACCOUNT_LOCKED", meta.IPAddress, meta.UserAgent)
		return nil, domain.Unauthorized("ACCOUNT_INACTIVE", "La cuenta esta desactivada")
	}
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		_ = s.repo.AuditLogin(ctx, &user.ID, email, "ACCOUNT_LOCKED", meta.IPAddress, meta.UserAgent)
		return nil, domain.Unauthorized("ACCOUNT_LOCKED", "La cuenta esta bloqueada temporalmente")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		attempts := user.FailedAttempts + 1
		var lockedUntil *time.Time
		if attempts >= 10 {
			t := time.Now().Add(24 * time.Hour)
			lockedUntil = &t
		} else if attempts >= 5 {
			t := time.Now().Add(15 * time.Minute)
			lockedUntil = &t
		}
		_ = s.repo.UpdateFailedLogin(ctx, user.ID, attempts, lockedUntil)
		_ = s.repo.AuditLogin(ctx, &user.ID, email, "INVALID_PASSWORD", meta.IPAddress, meta.UserAgent)
		return nil, domain.Unauthorized("INVALID_CREDENTIALS", "Email o contrasena incorrectos")
	}
	if err := s.repo.UpdateLoginSuccess(ctx, user.ID); err != nil {
		return nil, err
	}
	_ = s.repo.AuditLogin(ctx, &user.ID, email, "SUCCESS", meta.IPAddress, meta.UserAgent)
	return s.buildTokenResponse(ctx, *user, meta)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*domain.AuthResponse, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return nil, domain.BadRequest("refresh_token es obligatorio")
	}
	record, err := s.repo.FindRefreshToken(ctx, security.HashToken(refreshToken))
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, domain.Unauthorized("TOKEN_REVOKED", "Refresh token invalido, expirado o revocado")
	}
	user, err := s.repo.GetUserByID(ctx, record.UserID, false)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.IsActive {
		return nil, domain.Unauthorized("ACCOUNT_INACTIVE", "La cuenta esta desactivada")
	}
	authUser, err := s.BuildAuthUser(ctx, *user)
	if err != nil {
		return nil, err
	}
	access, err := s.tokens.Sign(*user, authUser)
	if err != nil {
		return nil, err
	}
	return &domain.AuthResponse{AccessToken: access, RefreshToken: refreshToken, TokenType: "Bearer", ExpiresIn: s.tokens.TTLSeconds(), User: authUser}, nil
}

func (s *Service) Logout(ctx context.Context, userID string, refreshToken string) error {
	if strings.TrimSpace(refreshToken) != "" {
		return s.repo.RevokeRefreshToken(ctx, userID, security.HashToken(refreshToken))
	}
	return s.repo.RevokeAllRefreshTokens(ctx, userID)
}

func (s *Service) Me(ctx context.Context, userID string) (*domain.Me, error) {
	user, err := s.repo.GetUserByID(ctx, userID, true)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.Unauthorized("TOKEN_INVALID", "Usuario no encontrado")
	}
	access, err := s.repo.GetUserAccess(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	features := make([]string, 0, len(access.Features))
	moduleSeen := map[string]bool{}
	modules := []string{}
	for _, feature := range access.Features {
		features = append(features, feature.FeatureCode+":"+feature.ScopeType)
		if feature.ModuleCode != "" && !moduleSeen[feature.ModuleCode] {
			moduleSeen[feature.ModuleCode] = true
			modules = append(modules, feature.ModuleCode)
		}
	}
	return &domain.Me{ID: user.ID, Email: user.Email, FullName: fullName(user.FirstName, user.LastName), ActorType: user.ActorType, ActorID: user.ActorID, Roles: user.Roles, Features: features, Modules: modules}, nil
}

func (s *Service) RequestPasswordReset(ctx context.Context, email string, meta RequestMeta) error {
	email = NormalizeEmail(email)
	if email == "" {
		return domain.BadRequest("Correo invalido")
	}
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}
	token, err := security.NewOpaqueToken(32)
	if err != nil {
		return err
	}
	if err := s.repo.InsertPasswordReset(ctx, user.ID, security.HashToken(token), meta.IPAddress); err != nil {
		return err
	}
	link := s.appURL + "/#/reset-password?token=" + token
	return s.mailer.PasswordReset(user.Email, fullName(user.FirstName, user.LastName), link)
}

func (s *Service) ConfirmPasswordReset(ctx context.Context, token, newPassword string) error {
	if strings.TrimSpace(token) == "" || len(newPassword) < 8 {
		return domain.BadRequest("Token y nueva contrasena valida son obligatorios")
	}
	request, err := s.repo.FindPasswordReset(ctx, security.HashToken(token))
	if err != nil {
		return err
	}
	if request == nil {
		return domain.BadRequest("El token expiro o ya fue usado")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}
	if err := s.repo.ConfirmPasswordReset(ctx, request.ID, request.UserID, string(hash)); err != nil {
		return err
	}
	return s.mailer.PasswordChanged(request.Email, fullName(request.FirstName, request.LastName))
}

func (s *Service) ListUsers(ctx context.Context, filters UserFilters) (domain.UserPage, error) {
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 {
		filters.PageSize = 20
	}
	if filters.PageSize > 100 {
		filters.PageSize = 100
	}
	return s.repo.ListUsers(ctx, filters)
}

func (s *Service) CreateUser(ctx context.Context, input CreateUserInput, assignedBy string) (*map[string]string, error) {
	input.Email = NormalizeEmail(input.Email)
	input.FirstName = clean(input.FirstName, 100)
	input.LastName = clean(input.LastName, 100)
	input.ActorType = clean(input.ActorType, 20)
	if input.Email == "" || input.FirstName == "" || input.LastName == "" || !validActorType(input.ActorType) {
		return nil, domain.Validation("Datos de usuario invalidos")
	}
	temporaryPassword, err := generateTemporaryPassword()
	if err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(temporaryPassword), 12)
	if err != nil {
		return nil, err
	}
	id, email, err := s.repo.CreateUser(ctx, input, string(hash), assignedBy)
	if err != nil {
		return nil, err
	}
	_ = s.mailer.Welcome(email, fullName(input.FirstName, input.LastName), temporaryPassword, s.appURL)
	return &map[string]string{"id": id, "email": email, "temporary_password": temporaryPassword}, nil
}

func (s *Service) GetUser(ctx context.Context, id string) (*domain.User, error) {
    user, err := s.repo.GetUserByID(ctx, id, true)
    if err != nil {
        return nil, err
    }
    if user == nil {
        return nil, domain.NotFound("USER_NOT_FOUND", "Usuario no encontrado")
    }
    return user, nil
}

func (s *Service) UpdateUser(ctx context.Context, id string, input UpdateUserInput) (*domain.User, error) {
	if input.ActorType != nil && !validActorType(*input.ActorType) {
		return nil, domain.Validation("actor_type invalido")
	}
	return s.repo.UpdateUser(ctx, id, input)
}

func (s *Service) DeactivateUser(ctx context.Context, id string) error { return s.repo.DeactivateUser(ctx, id) }
func (s *Service) ListSessions(ctx context.Context, userID string) ([]domain.Session, error) { return s.repo.ListSessions(ctx, userID) }
func (s *Service) RevokeSession(ctx context.Context, requesterID, userID, sessionID string, hasManage bool) error {
	if requesterID != userID && !hasManage {
		return domain.Forbidden("IDENTITY_USER_MANAGE")
	}
	return s.repo.RevokeSession(ctx, userID, sessionID)
}
func (s *Service) AssignRole(ctx context.Context, userID, roleName string, trainingCenterID *string, assignedBy string, expiresAt *time.Time) (*domain.AssignedRole, error) {
	if clean(roleName, 50) == "" {
		return nil, domain.BadRequest("role_name es obligatorio")
	}
	return s.repo.AssignRole(ctx, userID, roleName, trainingCenterID, assignedBy, expiresAt)
}
func (s *Service) RevokeRole(ctx context.Context, userID, roleName string) error { return s.repo.RevokeRole(ctx, userID, roleName) }
func (s *Service) ListScopeOverrides(ctx context.Context, userID string) ([]domain.ScopeOverride, error) { return s.repo.ListScopeOverrides(ctx, userID) }
func (s *Service) CreateScopeOverride(ctx context.Context, input ScopeOverrideInput, grantedBy string) (*domain.ScopeOverride, error) {
	input.FeatureCode = clean(input.FeatureCode, 60)
	input.ScopeType = clean(input.ScopeType, 30)
	input.Reason = clean(input.Reason, 500)
	if input.FeatureCode == "" || input.ScopeType == "" || input.Reason == "" {
		return nil, domain.Validation("feature_code, scope_type y reason son obligatorios")
	}
	return s.repo.CreateScopeOverride(ctx, input, grantedBy)
}
func (s *Service) DeleteScopeOverride(ctx context.Context, userID, overrideID string) error { return s.repo.DeleteScopeOverride(ctx, userID, overrideID) }
func (s *Service) ListRoles(ctx context.Context) ([]domain.Role, error) { return s.repo.ListRoles(ctx) }
func (s *Service) GetRoleFeatures(ctx context.Context, roleID string) ([]domain.RoleFeature, error) { return s.repo.GetRoleFeatures(ctx, roleID) }
func (s *Service) ListModules(ctx context.Context) ([]domain.Module, error) { return s.repo.ListModules(ctx) }
func (s *Service) LoginAudit(ctx context.Context, filters AuditFilters) (domain.AuditPage, error) {
	if filters.Limit < 1 {
		filters.Limit = 50
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}
	filters.Email = NormalizeEmail(filters.Email)
	return s.repo.LoginAudit(ctx, filters)
}

func (s *Service) BuildAuthUser(ctx context.Context, user domain.User) (domain.AuthUser, error) {
	access, err := s.repo.GetUserAccess(ctx, user.ID)
	if err != nil {
		return domain.AuthUser{}, err
	}
	roles := make([]string, 0, len(access.Roles))
	var trainingCenterID *string
	for _, role := range access.Roles {
		roles = append(roles, role.Name)
		if trainingCenterID == nil && role.TrainingCenterID != nil {
			trainingCenterID = role.TrainingCenterID
		}
	}
	features := make([]string, 0, len(access.Features))
	for _, feature := range access.Features {
		features = append(features, feature.FeatureCode+":"+feature.ScopeType)
	}
	return domain.AuthUser{ID: user.ID, Email: user.Email, FullName: fullName(user.FirstName, user.LastName), Roles: roles, TrainingCenterID: trainingCenterID, Features: features}, nil
}

func (s *Service) buildTokenResponse(ctx context.Context, user domain.User, meta RequestMeta) (*domain.AuthResponse, error) {
	authUser, err := s.BuildAuthUser(ctx, user)
	if err != nil {
		return nil, err
	}
	access, err := s.tokens.Sign(user, authUser)
	if err != nil {
		return nil, err
	}
	refresh, err := security.NewOpaqueToken(48)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateRefreshToken(ctx, user.ID, security.HashToken(refresh), meta.DeviceHint, meta.IPAddress, s.refreshDays); err != nil {
		return nil, err
	}
	return &domain.AuthResponse{AccessToken: access, RefreshToken: refresh, TokenType: "Bearer", ExpiresIn: s.tokens.TTLSeconds(), User: authUser}, nil
}

func NormalizeEmail(value string) string {
	email := strings.ToLower(strings.TrimSpace(value))
	if regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`).MatchString(email) {
		return email
	}
	return ""
}

func PtrString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func ValidateNullableUUID(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(value) {
		return &value
	}
	return nil
}

func ParseTime(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04", "2006-01-02"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return &parsed
		}
	}
	return nil
}

func ParseBool(value string) *bool {
	if value == "true" {
		v := true
		return &v
	}
	if value == "false" {
		v := false
		return &v
	}
	return nil
}

func clean(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) > max {
		return value[:max]
	}
	return value
}

func fullName(first, last string) string {
	return strings.TrimSpace(strings.TrimSpace(first) + " " + strings.TrimSpace(last))
}

func validActorType(value string) bool {
	return value == "USER" || value == "INSTRUCTOR" || value == "LEARNER"
}

func generateTemporaryPassword() (string, error) {
	bytes := make([]byte, 9)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes) + "A1!", nil
}