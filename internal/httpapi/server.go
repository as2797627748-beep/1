package httpapi

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"

	"autocode-platform/internal/domain"
	"autocode-platform/internal/service"
)

//go:embed web/*
var webFS embed.FS

type Server struct {
	platform *service.Platform
	dev      *devWorkspace
	mux      *http.ServeMux
}

func New(platform *service.Platform) *Server {
	root := "/workspace"
	if cwd, err := os.Getwd(); err == nil && strings.TrimSpace(cwd) != "" {
		root = cwd
	}
	s := &Server{platform: platform, dev: newDevWorkspace(root), mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/intake/analyze", s.handleAnalyze)
	s.mux.HandleFunc("/api/system", s.handleSystem)
	s.mux.HandleFunc("/api/system/audit", s.handleAudit)
	s.mux.HandleFunc("/api/system/advice", s.handleAdvice)
	s.mux.HandleFunc("/api/workflow/options", s.handleWorkflowOptions)
	s.mux.HandleFunc("/api/settings/toggles/", s.handleToggleActions)
	s.mux.HandleFunc("/api/settings/model-packs/", s.handleModelPackActions)
	s.mux.HandleFunc("/api/models", s.handleModels)
	s.mux.HandleFunc("/api/tools", s.handleTools)
	s.mux.HandleFunc("/api/tools/", s.handleToolActions)
	s.mux.HandleFunc("/api/templates", s.handleTemplates)
	s.mux.HandleFunc("/api/runs", s.handleRuns)
	s.mux.HandleFunc("/api/runs/", s.handleRunActions)
	s.mux.HandleFunc("/api/dev/files", s.handleDevFiles)
	s.mux.HandleFunc("/api/dev/file", s.handleDevFile)
	s.mux.HandleFunc("/api/dev/file/rollback", s.handleDevFileRollback)
	s.mux.HandleFunc("/api/dev/terminal/exec", s.handleDevTerminalExec)
	s.mux.HandleFunc("/api/dev/terminal/sessions", s.handleDevTerminalSessions)
	s.mux.HandleFunc("/api/dev/terminal/sessions/", s.handleDevTerminalSessionActions)
	s.mux.HandleFunc("/api/dev/preview/", s.handleDevPreview)
	s.mux.HandleFunc("/api/events", s.handleEvents)
	s.mux.HandleFunc("/", s.handleStatic)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (s *Server) handleSystem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, s.platform.Summary())
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var input service.AnalyzeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	writeJSON(w, http.StatusOK, s.platform.Analyze(input))
}

func (s *Server) handleAudit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, s.platform.Audit())
}

func (s *Server) handleAdvice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var req domain.AdviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	writeJSON(w, http.StatusOK, s.platform.Advice(req))
}

func (s *Server) handleWorkflowOptions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": s.platform.Summary().WorkflowOptions})
}

func (s *Server) handleToggleActions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/settings/toggles/"), "/")
	var payload struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	item, err := s.platform.SetFeatureToggle(id, payload.Enabled)
	if err != nil {
		if errors.Is(err, service.ErrSettingNotAllowed) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) handleModelPackActions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/settings/model-packs/"), "/")
	var payload struct {
		Enabled    bool `json:"enabled"`
		Downloaded bool `json:"downloaded"`
		Remove     bool `json:"remove"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	item, err := s.platform.SetBuiltinModelPack(id, payload.Enabled, payload.Downloaded, payload.Remove)
	if err != nil {
		if errors.Is(err, service.ErrSettingNotAllowed) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": s.platform.ListModels()})
}

func (s *Server) handleTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": s.platform.ListAtomicTools()})
}

func (s *Server) handleToolActions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	relative := strings.TrimPrefix(r.URL.Path, "/api/tools/")
	parts := strings.Split(strings.Trim(relative, "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] != "invoke" {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "tool action not found"})
		return
	}
	invocation, err := s.platform.InvokeAtomicTool(parts[0])
	if err != nil {
		if errors.Is(err, service.ErrSettingNotAllowed) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "current system profile does not allow this atomic tool"})
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, invocation)
}

func (s *Server) handleTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": s.platform.ListTemplates()})
}

func (s *Server) handleRuns(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"items": s.platform.ListRuns()})
	case http.MethodPost:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
			return
		}
		var input domain.CreateRunInput
		if err := json.Unmarshal(body, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		run, err := s.platform.CreateRun(input)
		if errors.Is(err, service.ErrRunConflict) {
			writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error(), "run": run})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, run)
	default:
		writeMethodNotAllowed(w)
	}
}

func (s *Server) handleRunActions(w http.ResponseWriter, r *http.Request) {
	relative := strings.TrimPrefix(r.URL.Path, "/api/runs/")
	parts := strings.Split(strings.Trim(relative, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "run not found"})
		return
	}
	id := parts[0]
	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			writeMethodNotAllowed(w)
			return
		}
		run, err := s.platform.GetRun(id)
		if errors.Is(err, service.ErrRunNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, run)
		return
	}
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	action := parts[1]
	var (
		run domain.Run
		err error
	)
	switch action {
	case "pause":
		run, err = s.platform.PauseRun(id)
	case "resume":
		run, err = s.platform.ResumeRun(id)
	case "dev-activity":
		var payload domain.DevActivityInput
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		run, err = s.platform.RecordDevActivity(id, payload)
	case "requirements":
		var payload struct {
			Extra string `json:"extra"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		run, err = s.platform.PatchRequirements(id, payload.Extra)
	case "rollback":
		var payload struct {
			Version string `json:"version"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		run, err = s.platform.RollbackDeployment(id, payload.Version)
	case "deploy":
		run, err = s.platform.DeployRun(id)
	case "revalidate":
		run, err = s.platform.RevalidateRun(id)
	default:
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "unsupported action"})
		return
	}
	if errors.Is(err, service.ErrRunNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (s *Server) handleDevFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	items, err := s.dev.list(r.URL.Query().Get("path"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) handleDevFile(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		payload, err := s.dev.read(r.URL.Query().Get("path"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, payload)
	case http.MethodPost:
		var input struct {
			Path    string `json:"path"`
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		payload, err := s.dev.write(input.Path, input.Content)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, payload)
	default:
		writeMethodNotAllowed(w)
	}
}

func (s *Server) handleDevFileRollback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var input struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	payload, err := s.dev.rollback(input.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func (s *Server) handleDevTerminalExec(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var input struct {
		Command    string `json:"command"`
		CWD        string `json:"cwd"`
		Background bool   `json:"background"`
		RunID      string `json:"runId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	payload, err := s.dev.exec(input.Command, input.CWD, input.Background, input.RunID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func (s *Server) handleDevTerminalSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": s.dev.sessionsList(), "updatedRuns": s.syncTerminalRunUpdates()})
}

func (s *Server) handleDevTerminalSessionActions(w http.ResponseWriter, r *http.Request) {
	relative := strings.TrimPrefix(r.URL.Path, "/api/dev/terminal/sessions/")
	parts := strings.Split(strings.Trim(relative, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "session not found"})
		return
	}
	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			writeMethodNotAllowed(w)
			return
		}
		payload, err := s.dev.session(parts[0])
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"session": payload, "updatedRuns": s.syncTerminalRunUpdates()})
		return
	}
	if r.Method != http.MethodPost || parts[1] != "kill" {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "unsupported session action"})
		return
	}
	payload, err := s.dev.kill(parts[0])
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"session": payload, "updatedRuns": s.syncTerminalRunUpdates()})
}

func (s *Server) syncTerminalRunUpdates() []domain.Run {
	updates := s.dev.consumeSessionRunUpdates()
	runs := make([]domain.Run, 0, len(updates))
	for _, item := range updates {
		run, err := s.platform.RecordDevActivity(item.runID, domain.DevActivityInput{
			Kind:    "command",
			Target:  item.cwd,
			Command: item.command,
			Status:  item.status,
			Detail:  "后台命令状态已更新",
		})
		if err == nil {
			runs = append(runs, run)
		}
	}
	return runs
}

func (s *Server) handleDevPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	port := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/dev/preview/"), "/")
	if port == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing preview port"})
		return
	}
	proxy, err := newPreviewProxy(port)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	proxy.ServeHTTP(w, r)
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "stream unsupported"})
		return
	}
	ch, unsubscribe := s.platform.Subscribe()
	defer unsubscribe()
	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-ch:
			payload, _ := json.Marshal(event)
			fmt.Fprintf(w, "data: %s\n\n", payload)
			flusher.Flush()
		}
	}
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if name == "" || name == "." {
		name = "index.html"
	}
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	file, err := webFS.ReadFile("web/" + name)
	if err != nil {
		file, err = webFS.ReadFile("web/index.html")
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "missing web assets"})
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(file)
		return
	}
	if contentType := mime.TypeByExtension(path.Ext(name)); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(file)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeMethodNotAllowed(w http.ResponseWriter) {
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}
