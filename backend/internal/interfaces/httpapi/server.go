package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"sena-iam-api/internal/application"
	"sena-iam-api/internal/domain"
	"sena-iam-api/internal/infrastructure/security"
)

type Server struct {
	app     *application.Service
	tokens  *security.TokenManager
	origins map[string]bool
}

type ctxKey string

const traceKey ctxKey = "trace_id"

func New(app *application.Service, tokens *security.TokenManager, origins []string) *Server {
	allowed := map[string]bool{}
	for _, origin := range origins {
		allowed[origin] = true
	}
	return &Server{app: app, tokens: tokens, origins: allowed}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.wrap(s.health))
	mux.HandleFunc("GET /.well-known/jwks.json", s.wrap(s.jwks))
	mux.HandleFunc("GET /api/v1/auth/.well-known/jwks.json", s.wrap(s.jwks))
	mux.HandleFunc("POST /api/v1/auth/login", s.wrap(s.login))
	mux.HandleFunc("POST /api/v1/auth/refresh", s.wrap(s.refresh))
	mux.HandleFunc("POST /api/v1/auth/logout", s.auth(s.logout))
	mux.HandleFunc("GET /api/v1/auth/me", s.auth(s.me))
	mux.HandleFunc("POST /api/v1/auth/password-reset/request", s.wrap(s.passwordResetRequest))
	mux.HandleFunc("POST /api/v1/auth/password-reset/confirm", s.wrap(s.passwordResetConfirm))
	mux.HandleFunc("GET /api/v1/users", s.feature("IDENTITY_USER_VIEW", s.listUsers))
	mux.HandleFunc("POST /api/v1/users", s.feature("IDENTITY_USER_MANAGE", s.createUser))
	mux.HandleFunc("GET /api/v1/users/{id}", s.feature("IDENTITY_USER_VIEW", s.getUser))
	mux.HandleFunc("PUT /api/v1/users/{id}", s.feature("IDENTITY_USER_MANAGE", s.updateUser))
	mux.HandleFunc("POST /api/v1/users/{id}/deactivate", s.feature("IDENTITY_USER_MANAGE", s.deactivateUser))
	mux.HandleFunc("GET /api/v1/users/{id}/sessions", s.feature("IDENTITY_USER_VIEW", s.listSessions))
	mux.HandleFunc("DELETE /api/v1/users/{id}/sessions/{session_id}", s.auth(s.revokeSession))
	mux.HandleFunc("POST /api/v1/users/{id}/roles", s.feature("IDENTITY_ROLE_ASSIGN", s.assignRole))
	mux.HandleFunc("DELETE /api/v1/users/{id}/roles/{role_name}", s.feature("IDENTITY_ROLE_ASSIGN", s.revokeRole))
	mux.HandleFunc("GET /api/v1/users/{id}/scope-overrides", s.feature("IDENTITY_SCOPE_MANAGE", s.listScopeOverrides))
	mux.HandleFunc("POST /api/v1/users/{id}/scope-overrides", s.feature("IDENTITY_SCOPE_MANAGE", s.createScopeOverride))
	mux.HandleFunc("DELETE /api/v1/users/{id}/scope-overrides/{override_id}", s.feature("IDENTITY_SCOPE_MANAGE", s.deleteScopeOverride))
	mux.HandleFunc("GET /api/v1/roles", s.feature("IDENTITY_ROLE_VIEW", s.listRoles))
	mux.HandleFunc("GET /api/v1/roles/{id}/features", s.feature("IDENTITY_ROLE_VIEW", s.roleFeatures))
	mux.HandleFunc("GET /api/v1/modules", s.feature("IDENTITY_ROLE_VIEW", s.listModules))
	mux.HandleFunc("GET /api/v1/reports/login-audit", s.feature("AUDIT_LOG_VIEW", s.loginAudit))
	return s.recover(s.trace(s.cors(mux)))
}

type authedHandler func(http.ResponseWriter, *http.Request, *security.Claims) error
type publicHandler func(http.ResponseWriter, *http.Request) error

func (s *Server) wrap(next publicHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := next(w, r); err != nil {
			s.writeError(w, r, err)
		}
	}
}

func (s *Server) auth(next authedHandler) http.HandlerFunc {
	return s.wrap(func(w http.ResponseWriter, r *http.Request) error {
		claims, err := s.authenticate(r)
		if err != nil {
			return err
		}
		return next(w, r, claims)
	})
}

func (s *Server) feature(feature string, next authedHandler) http.HandlerFunc {
	return s.auth(func(w http.ResponseWriter, r *http.Request, claims *security.Claims) error {
		if !hasFeature(claims, feature) {
			return domain.Forbidden(feature)
		}
		return next(w, r, claims)
	})
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) error {
	if err := s.app.Health(r.Context()); err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "iam-api"})
}

func (s *Server) jwks(w http.ResponseWriter, _ *http.Request) error {
	return writeJSON(w, http.StatusOK, s.tokens.JWKS())
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := readJSON(r, &body); err != nil {
		return err
	}
	response, err := s.app.Login(r.Context(), body.Email, body.Password, meta(r))
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, response)
}

func (s *Server) refresh(w http.ResponseWriter, r *http.Request) error {
	var body struct{ RefreshToken string `json:"refresh_token"` }
	if err := readJSON(r, &body); err != nil {
		return err
	}
	response, err := s.app.Refresh(r.Context(), body.RefreshToken)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, response)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request, claims *security.Claims) error {
	var body struct{ RefreshToken string `json:"refresh_token"` }
	_ = readJSON(r, &body)
	if err := s.app.Logout(r.Context(), claims.Subject, body.RefreshToken); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Server) me(w http.ResponseWriter, r *http.Request, claims *security.Claims) error {
	response, err := s.app.Me(r.Context(), claims.Subject)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, response)
}

func (s *Server) passwordResetRequest(w http.ResponseWriter, r *http.Request) error {
	var body struct{ Email string `json:"email"` }
	if err := readJSON(r, &body); err != nil {
		return err
	}
	if err := s.app.RequestPasswordReset(r.Context(), body.Email, meta(r)); err != nil {
		return err
	}
	w.WriteHeader(http.StatusAccepted)
	return nil
}

func (s *Server) passwordResetConfirm(w http.ResponseWriter, r *http.Request) error {
	var body struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := readJSON(r, &body); err != nil {
		return err
	}
	if err := s.app.ConfirmPasswordReset(r.Context(), body.Token, body.NewPassword); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Server) listUsers(w http.ResponseWriter, r *http.Request, _ *security.Claims) error {
	q := r.URL.Query()
	filters := application.UserFilters{Page: clamp(q.Get("page"), 1, 1, 100000), PageSize: clamp(q.Get("page_size"), 20, 1, 100), Sort: q.Get("sort"), Search: q.Get("search"), Role: q.Get("role"), TrainingCenterID: q.Get("training_center_id")}
	filters.IsActive = application.ParseBool(q.Get("is_active"))
	page, err := s.app.ListUsers(r.Context(), filters)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, page)
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request, claims *security.Claims) error {
	var body struct {
		Email            string  `json:"email"`
		FirstName        string  `json:"first_name"`
		LastName         string  `json:"last_name"`
		ActorType        string  `json:"actor_type"`
		ActorID          *string `json:"actor_id"`
		InitialRole      *string `json:"initial_role"`
		TrainingCenterID *string `json:"training_center_id"`
	}
	if err := readJSON(r, &body); err != nil {
		return err
	}
	input := application.CreateUserInput{Email: body.Email, FirstName: body.FirstName, LastName: body.LastName, ActorType: body.ActorType, InitialRole: body.InitialRole}
	if body.ActorID != nil {
		input.ActorID = application.ValidateNullableUUID(*body.ActorID)
	}
	if body.TrainingCenterID != nil {
		input.TrainingCenterID = application.ValidateNullableUUID(*body.TrainingCenterID)
	}
	response, err := s.app.CreateUser(r.Context(), input, claims.Subject)
	if err != nil {
		return err
	}
	w.Header().Set("Location", "/api/v1/users/"+(*response)["id"])
	return writeJSON(w, http.StatusCreated, response)
}

func (s *Server) getUser(w http.ResponseWriter, r *http.Request, _ *security.Claims) error {
	user, err := s.appUser(r.Context(), r.PathValue("id"))
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, user)
}

func (s *Server) updateUser(w http.ResponseWriter, r *http.Request, _ *security.Claims) error {
	var body struct {
		FirstName *string `json:"first_name"`
		LastName  *string `json:"last_name"`
		ActorType *string `json:"actor_type"`
		ActorID   *string `json:"actor_id"`
		IsActive  *bool   `json:"is_active"`
	}
	if err := readJSON(r, &body); err != nil {
		return err
	}
	input := application.UpdateUserInput{FirstName: body.FirstName, LastName: body.LastName, ActorType: body.ActorType, IsActive: body.IsActive}
	if body.ActorID != nil {
		input.ActorID = application.ValidateNullableUUID(*body.ActorID)
	}
	user, err := s.app.UpdateUser(r.Context(), r.PathValue("id"), input)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, user)
}

func (s *Server) deactivateUser(w http.ResponseWriter, r *http.Request, _ *security.Claims) error {
	if err := s.app.DeactivateUser(r.Context(), r.PathValue("id")); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Server) listSessions(w http.ResponseWriter, r *http.Request, _ *security.Claims) error {
	items, err := s.app.ListSessions(r.Context(), r.PathValue("id"))
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, items)
}

func (s *Server) revokeSession(w http.ResponseWriter, r *http.Request, claims *security.Claims) error {
	if err := s.app.RevokeSession(r.Context(), claims.Subject, r.PathValue("id"), r.PathValue("session_id"), hasFeature(claims, "IDENTITY_USER_MANAGE")); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Server) assignRole(w http.ResponseWriter, r *http.Request, claims *security.Claims) error {
	var body struct {
		RoleName         string  `json:"role_name"`
		TrainingCenterID *string `json:"training_center_id"`
		ExpiresAt        string  `json:"expires_at"`
	}
	if err := readJSON(r, &body); err != nil {
		return err
	}
	var center *string
	if body.TrainingCenterID != nil {
		center = application.ValidateNullableUUID(*body.TrainingCenterID)
	}
	role, err := s.app.AssignRole(r.Context(), r.PathValue("id"), body.RoleName, center, claims.Subject, application.ParseTime(body.ExpiresAt))
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusCreated, role)
}

func (s *Server) revokeRole(w http.ResponseWriter, r *http.Request, _ *security.Claims) error {
	if err := s.app.RevokeRole(r.Context(), r.PathValue("id"), r.PathValue("role_name")); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Server) listScopeOverrides(w http.ResponseWriter, r *http.Request, _ *security.Claims) error {
	items, err := s.app.ListScopeOverrides(r.Context(), r.PathValue("id"))
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, items)
}

func (s *Server) createScopeOverride(w http.ResponseWriter, r *http.Request, claims *security.Claims) error {
	var body struct {
		FeatureCode string `json:"feature_code"`
		ScopeType   string `json:"scope_type"`
		IsAllowed   bool   `json:"is_allowed"`
		Reason      string `json:"reason"`
		ExpiresAt   string `json:"expires_at"`
	}
	if err := readJSON(r, &body); err != nil {
		return err
	}
	input := application.ScopeOverrideInput{UserID: r.PathValue("id"), FeatureCode: body.FeatureCode, ScopeType: body.ScopeType, IsAllowed: body.IsAllowed, Reason: body.Reason, ExpiresAt: application.ParseTime(body.ExpiresAt)}
	item, err := s.app.CreateScopeOverride(r.Context(), input, claims.Subject)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusCreated, item)
}

func (s *Server) deleteScopeOverride(w http.ResponseWriter, r *http.Request, _ *security.Claims) error {
	if err := s.app.DeleteScopeOverride(r.Context(), r.PathValue("id"), r.PathValue("override_id")); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Server) listRoles(w http.ResponseWriter, r *http.Request, _ *security.Claims) error {
	items, err := s.app.ListRoles(r.Context())
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, items)
}

func (s *Server) roleFeatures(w http.ResponseWriter, r *http.Request, _ *security.Claims) error {
	items, err := s.app.GetRoleFeatures(r.Context(), r.PathValue("id"))
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, items)
}

func (s *Server) listModules(w http.ResponseWriter, r *http.Request, _ *security.Claims) error {
	items, err := s.app.ListModules(r.Context())
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, items)
}

func (s *Server) loginAudit(w http.ResponseWriter, r *http.Request, _ *security.Claims) error {
	q := r.URL.Query()
	filters := application.AuditFilters{Limit: clamp(q.Get("limit"), 50, 1, 100), Cursor: q.Get("cursor"), UserID: q.Get("user_id"), Email: q.Get("email"), Outcome: q.Get("outcome"), From: application.ParseTime(q.Get("from")), To: application.ParseTime(q.Get("to"))}
	page, err := s.app.LoginAudit(r.Context(), filters)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, page)
}

func (s *Server) appUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.app.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.NotFound("USER_NOT_FOUND", "Usuario no encontrado")
	}
	return user, nil
}

func (s *Server) authenticate(r *http.Request) (*security.Claims, error) {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return nil, domain.Unauthorized("MISSING_TOKEN", "Falta el token Bearer")
	}
	claims, err := s.tokens.Verify(strings.TrimSpace(header[7:]))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domain.Unauthorized("TOKEN_EXPIRED", "Access token expirado")
		}
		return nil, domain.Unauthorized("TOKEN_INVALID", "Token malformado o firma invalida")
	}
	return claims, nil
}

func hasFeature(claims *security.Claims, feature string) bool {
	for _, entry := range claims.Features {
		code := strings.Split(entry, ":")[0]
		if code == feature {
			return true
		}
	}
	return false
}

func readJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		return domain.BadRequest("JSON invalido")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(value)
}

func (s *Server) writeError(w http.ResponseWriter, r *http.Request, err error) {
	var appErr *domain.AppError
	if !errors.As(err, &appErr) {
		log.Printf("internal error: %v", err)
		appErr = domain.NewError(http.StatusInternalServerError, "INTERNAL_ERROR", "Error interno del servidor")
	}
	traceID, _ := r.Context().Value(traceKey).(string)
	_ = writeJSON(w, appErr.Status, map[string]any{"error": map[string]any{"code": appErr.Code, "message": appErr.Message, "details": appErr.Details, "trace_id": traceID}})
}

func (s *Server) trace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("x-trace-id")
		if traceID == "" {
			traceID = randomTrace()
		}
		w.Header().Set("x-trace-id", traceID)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), traceKey, traceID)))
	})
}

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && (len(s.origins) == 0 || s.origins[origin]) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Trace-Id")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("panic: %v", recovered)
				s.writeError(w, r, domain.NewError(http.StatusInternalServerError, "INTERNAL_ERROR", "Error interno del servidor"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func meta(r *http.Request) application.RequestMeta {
	ip := clientIP(r)
	ua := strings.TrimSpace(r.UserAgent())
	return application.RequestMeta{IPAddress: application.PtrString(ip), UserAgent: application.PtrString(ua), DeviceHint: application.PtrString(ua)}
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func clamp(value string, fallback, min, max int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		parsed = fallback
	}
	if parsed < min {
		return min
	}
	if parsed > max {
		return max
	}
	return parsed
}

func randomTrace() string {
	return strconv.FormatInt(int64(len(strings.TrimSpace("iam")))+int64(len(strconv.Itoa(http.StatusOK))), 10) + "-" + strconv.FormatInt(int64(len(strings.Repeat("x", 12))), 10)
}