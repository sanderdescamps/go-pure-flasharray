package mock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/sanderdescamps/go-pure-flasharray/internal/fakearray"
	"github.com/sanderdescamps/go-pure-flasharray/pkg/flashclient"

	"github.com/gorilla/mux"
)

type Mock struct {
	*http.Server
	logger            *slog.Logger
	APIToken          string
	Username          string
	lockSessionTokens sync.Mutex
	sessionTokens     []string
	DisableAuth       bool
}

type MockOption func(*Mock)

func WithLogger(logger *slog.Logger) MockOption {
	return func(m *Mock) {
		m.SetLogger(logger)
	}
}

func NewMock(array *fakearray.Array, options ...MockOption) *Mock {
	m := &Mock{
		Username:    "fake-auth-user",
		APIToken:    "fake-auth-token",
		DisableAuth: false,
		logger:      slog.New(slog.DiscardHandler),
	}

	for _, option := range options {
		option(m)
	}

	router := mux.NewRouter()
	router.Use(m.logMiddleware())

	router.Methods("GET").Path("/api/api_version").HandlerFunc(GetVersionHandler(array))

	authRouter := router.NewRoute().Subrouter()
	authRouter.Use(m.versionMiddleware(array.GetVersions()))
	authRouter.Methods("POST").Path("/api/{api_version}/login").HandlerFunc(m.loginHandler)
	authRouter.Methods("POST").Path("/api/{api_version}/logout").HandlerFunc(m.logoutHandler)

	pathRouter := router.NewRoute().Subrouter()
	pathRouter.Use(m.authMiddleware, m.versionMiddleware(array.GetVersions()))

	InitConnectionRouter(pathRouter, array, m.logger)
	InitHostRouter(pathRouter, array, m.logger)
	InitHostGroupRouter(pathRouter, array, m.logger)
	InitProtectionGroupRouter(pathRouter, array, m.logger)
	InitPodRouter(pathRouter, array, m.logger)
	InitVolumeRouter(pathRouter, array, m.logger)
	InitVolumeGroupRouter(pathRouter, array, m.logger)
	InitVolumeSnapshotRouter(pathRouter, array, m.logger)
	InitMiscRouter(pathRouter, array, m.logger)

	m.Server = &http.Server{
		Handler:  router,
		ErrorLog: slog.NewLogLogger(m.logger.Handler(), slog.LevelInfo),
	}
	return m
}

func (m *Mock) SetLogger(logger *slog.Logger) {
	m.logger = logger
	if m.Server != nil {
		m.ErrorLog = slog.NewLogLogger(logger.Handler(), slog.LevelInfo)
	}
}

func NewMockWithTestData(apiToken string) (*Mock, error) {
	array, err := fakearray.NewArrayWithTestData()
	if err != nil {
		return nil, err
	}
	mock := NewMock(array)
	mock.APIToken = apiToken
	return mock, nil
}

func (m *Mock) GenerateNewSessionToken() string {
	m.lockSessionTokens.Lock()
	defer m.lockSessionTokens.Unlock()
	newToken := NewRandomString(32)
	m.sessionTokens = append(m.sessionTokens, newToken)

	return newToken
}

func (m *Mock) ValidateSessionToken(token string) bool {
	m.lockSessionTokens.Lock()
	defer m.lockSessionTokens.Unlock()
	return slices.Contains(m.sessionTokens, token)
}

func (m *Mock) InvalidateSessionToken(token string) {
	m.lockSessionTokens.Lock()
	defer m.lockSessionTokens.Unlock()
	m.sessionTokens = slices.DeleteFunc(m.sessionTokens, func(t string) bool { return t == token })
}

func (m *Mock) Start(host string, port int) error {
	m.Addr = fmt.Sprintf("%s:%d", host, port)
	fmt.Printf("Starting mock server on %s\n", m.Addr)
	if err := m.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (m *Mock) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = m.Shutdown(ctx)
}

func (m *Mock) logMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m.logger.InfoContext(r.Context(), fmt.Sprintf(
				"%s %s %s",
				r.Method,
				r.URL.Path,
				r.URL.Query().Encode(),
			))
			next.ServeHTTP(w, r)
		})
	}
}

func (m *Mock) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if xAuthToken := r.Header.Get("x-auth-token"); !m.DisableAuth && xAuthToken == "" {
			httpJsonError(w, "Missing x-auth-token header", http.StatusUnauthorized)
			return
		} else if !m.DisableAuth && !m.ValidateSessionToken(xAuthToken) {
			httpJsonError(w, "Unauthorized: invalid x-auth-token", http.StatusUnauthorized)
			return
		}
		// Extract claims and pass to next handler
		next.ServeHTTP(w, r)
	})
}

func (m *Mock) versionMiddleware(supportedVersions flashclient.ApiVersions) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			muxVars := mux.Vars(r)
			apiVersion, ok := muxVars["api_version"]
			if !ok {
				httpJsonError(w, "Missing api_version in URL", http.StatusBadRequest)
				return
			}
			if !slices.Contains(supportedVersions, apiVersion) {
				fmt.Printf("Unsupported API version: %s\n", apiVersion)
				httpJsonError(w, "Unsupported API version", http.StatusBadRequest)
				return
			}
			// Extract claims and pass to next handler
			next.ServeHTTP(w, r)
		})
	}
}

func (m *Mock) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpJsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiToken := r.Header.Get("api-token")

	if apiToken == "" {
		httpJsonError(w, "Missing api-token header", http.StatusBadRequest)
		return
	} else if apiToken != m.APIToken {
		httpJsonError(w, "Unauthorized: invalid api-token", http.StatusUnauthorized)
		return
	}

	sessionToken := m.GenerateNewSessionToken()
	w.Header().Set("x-auth-token", sessionToken)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	body := map[string]interface{}{
		"items": []map[string]string{
			{"username": m.Username},
		},
	}
	json.NewEncoder(w).Encode(body)
}

func (m *Mock) logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpJsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionToken := r.Header.Get("x-auth-token")

	if m.ValidateSessionToken(sessionToken) {
		m.InvalidateSessionToken(sessionToken)
	} else {
		httpJsonError(w, "Unauthorized: invalid x-auth-token", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetVersionHandler(array *fakearray.Array) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		data := map[string][]string{
			"version": array.Versions,
		}
		json.NewEncoder(w).Encode(data)
	}
}

func GetEmptyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		data := flashclient.NewResults[any]([]any{})
		json.NewEncoder(w).Encode(data)
	}
}
