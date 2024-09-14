package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"plugin"
	"strings"
	"time"

	"goction/internal/api/dashboard"
	"goction/internal/config"
	"goction/internal/stats"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type GoctionFunc func(...string) (string, error)

type Server struct {
	config           *config.Config
	router           *mux.Router
	logger           *logrus.Logger
	stats            *stats.Manager
	goctionsCache    map[string]GoctionFunc
	dashboardHandler *dashboard.Handler
}

func NewServer(cfg *config.Config) (*Server, error) {
	logger := logrus.New()
	setupLogger(logger, cfg)

	statsManager, err := stats.NewManager(cfg.StatsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create stats manager: %w", err)
	}

	dashboardHandler, err := dashboard.NewHandler(cfg, statsManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create dashboard handler: %w", err)
	}

	s := &Server{
		config:           cfg,
		router:           mux.NewRouter(),
		logger:           logger,
		stats:            statsManager,
		goctionsCache:    make(map[string]GoctionFunc),
		dashboardHandler: dashboardHandler,
	}
	s.routes()
	return s, nil
}

func setupLogger(logger *logrus.Logger, cfg *config.Config) {
	isService := os.Getenv("GOCTION_SERVICE") == "true"

	if isService {
		logger.SetOutput(os.Stdout)
	} else {
		logFile, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logger.Fatalf("failed to open log file: %v", err)
		}
		logger.SetOutput(logFile)
	}

	logger.SetLevel(logrus.DebugLevel)
}

func (s *Server) routes() {
	s.router.Use(s.loggingMiddleware)

	// Dashboard routes
	s.router.HandleFunc("/", s.dashboardHandler.AuthMiddleware(s.dashboardHandler.HandleDashboard)).Methods("GET")
	s.router.HandleFunc("/login", s.dashboardHandler.HandleLogin).Methods("GET", "POST")
	s.router.HandleFunc("/logout", s.dashboardHandler.HandleLogout).Methods("POST")

	// Goction API routes
	s.router.HandleFunc("/goctions/{goction}", s.apiAuthMiddleware(s.handleExecuteGoction)).Methods("POST")
	s.router.HandleFunc("/goctions", s.apiAuthMiddleware(s.handleListGoctions)).Methods("GET")
	s.router.HandleFunc("/goctions/{goction}/info", s.apiAuthMiddleware(s.handleGetGoctionInfo)).Methods("GET")
	s.router.HandleFunc("/goctions/{goction}/history", s.apiAuthMiddleware(s.handleGetGoctionHistory)).Methods("GET")
}

func (s *Server) Start() error {
	s.logger.Infof("Server starting on :%d", s.config.Port)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.config.Port), s.router)
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.logger.WithFields(logrus.Fields{
			"method":   r.Method,
			"path":     r.URL.Path,
			"duration": time.Since(start),
		}).Info("Request handled")
	})
}

func (s *Server) apiAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-API-Token")
		if !strings.EqualFold(strings.TrimSpace(token), strings.TrimSpace(s.config.APIToken)) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func (s *Server) handleExecuteGoction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	goctionName := vars["goction"]

	goction, err := s.getGoction(goctionName)
	if err != nil {
		s.logger.WithError(err).Errorf("Failed to get goction: %s", goctionName)
		http.Error(w, fmt.Sprintf("Goction not found: %v", err), http.StatusNotFound)
		return
	}

	var requestBody struct {
		Args []string `json:"args"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		s.logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	start := time.Now()
	result, err := goction(requestBody.Args...)
	duration := time.Since(start)

	s.stats.RecordExecution(goctionName, duration, err == nil, result)

	if err != nil {
		s.logger.WithError(err).Errorf("Goction execution failed: %s", goctionName)
		http.Error(w, fmt.Sprintf("Goction execution failed: %v", err), http.StatusInternalServerError)
		return
	}

	s.logger.WithFields(logrus.Fields{
		"goction":  goctionName,
		"duration": duration,
	}).Info("Goction executed successfully")

	json.NewEncoder(w).Encode(map[string]string{"result": result})
}

func (s *Server) handleListGoctions(w http.ResponseWriter, r *http.Request) {
	goctions, err := s.listGoctions()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list goctions: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string][]string{"goctions": goctions})
}

func (s *Server) handleGetGoctionInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	goctionName := vars["goction"]

	info, err := s.getGoctionInfo(goctionName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get goction info: %v", err), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(info)
}

func (s *Server) handleGetGoctionHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	goctionName := vars["goction"]

	history, err := s.getGoctionHistory(goctionName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get goction history: %v", err), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string][]stats.ExecutionRecord{"history": history})
}

func (s *Server) getGoction(name string) (GoctionFunc, error) {
	if goction, ok := s.goctionsCache[name]; ok {
		return goction, nil
	}

	goctionPath := filepath.Join(s.config.GoctionsDir, name, name+".so")
	plug, err := plugin.Open(goctionPath)
	if err != nil {
		return nil, fmt.Errorf("could not open goction plugin: %w", err)
	}

	sym, err := plug.Lookup(strings.Title(name))
	if err != nil {
		return nil, fmt.Errorf("could not find goction symbol: %w", err)
	}

	goction, ok := sym.(func(...string) (string, error))
	if !ok {
		return nil, fmt.Errorf("unexpected type from module symbol")
	}

	s.goctionsCache[name] = goction
	return goction, nil
}

func (s *Server) listGoctions() ([]string, error) {
	files, err := os.ReadDir(s.config.GoctionsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read goctions directory: %w", err)
	}

	var goctions []string
	for _, file := range files {
		if file.IsDir() {
			goctions = append(goctions, file.Name())
		}
	}
	return goctions, nil
}

func (s *Server) getGoctionInfo(name string) (map[string]interface{}, error) {
	goctionStats, ok := s.stats.GetStats(name)
	if !ok {
		return nil, fmt.Errorf("goction not found: %s", name)
	}

	info := map[string]interface{}{
		"name":        name,
		"description": "Goction description", // You might want to store and retrieve this from somewhere
		"lastUpdated": time.Now(),            // You might want to track this separately
		"stats": map[string]interface{}{
			"totalCalls":      goctionStats.TotalCalls,
			"successfulCalls": goctionStats.SuccessfulCalls,
			"totalDuration":   goctionStats.TotalDuration.String(),
			"lastExecuted":    goctionStats.LastExecuted,
		},
	}

	return info, nil
}

func (s *Server) getGoctionHistory(name string) ([]stats.ExecutionRecord, error) {
	history := s.stats.GetExecutionHistory(name)
	if history == nil {
		return nil, fmt.Errorf("no history found for goction: %s", name)
	}
	return history, nil
}
