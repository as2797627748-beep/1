package httpapi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type devWorkspace struct {
	root         string
	mu           sync.RWMutex
	history      map[string][]fileSnapshot
	sessions     map[string]*terminalSession
	sessionOrder []string
	counter      uint64
}

type fileSnapshot struct {
	content   string
	createdAt time.Time
}

type terminalSession struct {
	id        string
	runID     string
	command   string
	cwd       string
	status    string
	exitCode  int
	output    bytes.Buffer
	startedAt time.Time
	endedAt   time.Time
	cancel    context.CancelFunc
	reported  bool
	mu        sync.RWMutex
}

func newDevWorkspace(root string) *devWorkspace {
	return &devWorkspace{
		root:     root,
		history:  map[string][]fileSnapshot{},
		sessions: map[string]*terminalSession{},
	}
}

func (d *devWorkspace) normalizePath(target string) (string, string, error) {
	rel := strings.TrimSpace(target)
	if rel == "" || rel == "." || rel == "/" {
		rel = "."
	}
	cleaned := filepath.Clean(rel)
	abs := filepath.Join(d.root, cleaned)
	abs, err := filepath.Abs(abs)
	if err != nil {
		return "", "", err
	}
	rootAbs, err := filepath.Abs(d.root)
	if err != nil {
		return "", "", err
	}
	if abs != rootAbs && !strings.HasPrefix(abs, rootAbs+string(os.PathSeparator)) {
		return "", "", fmt.Errorf("path out of workspace")
	}
	relToRoot, err := filepath.Rel(rootAbs, abs)
	if err != nil {
		return "", "", err
	}
	if relToRoot == "." {
		return abs, ".", nil
	}
	return abs, filepath.ToSlash(relToRoot), nil
}

func (d *devWorkspace) list(rel string) ([]map[string]any, error) {
	abs, normalized, err := d.normalizePath(rel)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(entries))
	for _, entry := range entries {
		info, infoErr := entry.Info()
		if infoErr != nil {
			continue
		}
		childPath := entry.Name()
		if normalized != "." {
			childPath = normalized + "/" + entry.Name()
		}
		items = append(items, map[string]any{
			"name":      entry.Name(),
			"path":      childPath,
			"isDir":     entry.IsDir(),
			"size":      info.Size(),
			"updatedAt": info.ModTime(),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		leftDir := items[i]["isDir"].(bool)
		rightDir := items[j]["isDir"].(bool)
		if leftDir != rightDir {
			return leftDir
		}
		return items[i]["name"].(string) < items[j]["name"].(string)
	})
	return items, nil
}

func (d *devWorkspace) read(rel string) (map[string]any, error) {
	abs, normalized, err := d.normalizePath(rel)
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(abs)
	if err != nil {
		return nil, err
	}
	d.mu.RLock()
	historyDepth := len(d.history[normalized])
	d.mu.RUnlock()
	return map[string]any{
		"path":         normalized,
		"content":      string(content),
		"historyDepth": historyDepth,
	}, nil
}

func (d *devWorkspace) write(rel string, content string) (map[string]any, error) {
	abs, normalized, err := d.normalizePath(rel)
	if err != nil {
		return nil, err
	}
	if existing, readErr := os.ReadFile(abs); readErr == nil {
		d.mu.Lock()
		d.history[normalized] = append([]fileSnapshot{{content: string(existing), createdAt: time.Now()}}, d.history[normalized]...)
		if len(d.history[normalized]) > 10 {
			d.history[normalized] = d.history[normalized][:10]
		}
		d.mu.Unlock()
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		return nil, err
	}
	d.mu.RLock()
	historyDepth := len(d.history[normalized])
	d.mu.RUnlock()
	return map[string]any{"path": normalized, "saved": true, "historyDepth": historyDepth}, nil
}

func (d *devWorkspace) rollback(rel string) (map[string]any, error) {
	abs, normalized, err := d.normalizePath(rel)
	if err != nil {
		return nil, err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	versions := d.history[normalized]
	if len(versions) == 0 {
		return nil, fmt.Errorf("no history available")
	}
	last := versions[0]
	d.history[normalized] = versions[1:]
	if err := os.WriteFile(abs, []byte(last.content), 0o644); err != nil {
		return nil, err
	}
	return map[string]any{"path": normalized, "restored": true, "historyDepth": len(d.history[normalized])}, nil
}

type sessionRunUpdate struct {
	runID   string
	command string
	cwd     string
	status  string
}

func (d *devWorkspace) exec(command string, cwd string, background bool, runID string) (map[string]any, error) {
	workingDir, normalized, err := d.normalizePath(cwd)
	if err != nil {
		return nil, err
	}
	if background {
		return d.startBackground(command, workingDir, normalized, strings.TrimSpace(runID))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	shell, shellArgs := shellCommand(command)
	cmd := exec.CommandContext(ctx, shell, shellArgs...)
	cmd.Dir = workingDir
	output, err := cmd.CombinedOutput()
	status := "completed"
	exitCode := 0
	if err != nil {
		status = "failed"
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}
	if ctx.Err() == context.DeadlineExceeded {
		status = "timeout"
		exitCode = 124
	}
	return map[string]any{
		"mode":     "foreground",
		"cwd":      normalized,
		"command":  command,
		"status":   status,
		"exitCode": exitCode,
		"output":   string(output),
	}, nil
}

func (d *devWorkspace) startBackground(command string, workingDir string, normalized string, runID string) (map[string]any, error) {
	ctx, cancel := context.WithCancel(context.Background())
	id := fmt.Sprintf("term-%06d", atomic.AddUint64(&d.counter, 1))
	shell, shellArgs := shellCommand(command)
	cmd := exec.CommandContext(ctx, shell, shellArgs...)
	cmd.Dir = workingDir
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	session := &terminalSession{id: id, runID: runID, command: command, cwd: normalized, status: "running", startedAt: time.Now(), cancel: cancel}
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, err
	}
	d.mu.Lock()
	d.sessions[id] = session
	d.sessionOrder = append([]string{id}, d.sessionOrder...)
	if len(d.sessionOrder) > 12 {
		d.sessionOrder = d.sessionOrder[:12]
	}
	d.mu.Unlock()
	go d.captureSessionOutput(session, stdout, stderr, cmd)
	return d.sessionSnapshot(session), nil
}

func (d *devWorkspace) captureSessionOutput(session *terminalSession, stdout io.ReadCloser, stderr io.ReadCloser, cmd *exec.Cmd) {
	defer stdout.Close()
	defer stderr.Close()
	_, _ = io.Copy(session, io.MultiReader(stdout, stderr))
	err := cmd.Wait()
	session.mu.Lock()
	defer session.mu.Unlock()
	session.endedAt = time.Now()
	if err != nil {
		session.status = "failed"
		if exitErr, ok := err.(*exec.ExitError); ok {
			session.exitCode = exitErr.ExitCode()
		} else {
			session.exitCode = 1
		}
		return
	}
	session.status = "completed"
}

func (s *terminalSession) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.output.Len()+len(p) > 256*1024 {
		trim := s.output.Len() + len(p) - 256*1024
		if trim > 0 {
			current := s.output.Bytes()
			s.output.Reset()
			if trim < len(current) {
				s.output.Write(current[trim:])
			}
		}
	}
	return s.output.Write(p)
}

func (d *devWorkspace) sessionSnapshot(session *terminalSession) map[string]any {
	session.mu.RLock()
	defer session.mu.RUnlock()
	return map[string]any{
		"id":        session.id,
		"runId":     session.runID,
		"command":   session.command,
		"cwd":       session.cwd,
		"status":    session.status,
		"exitCode":  session.exitCode,
		"output":    session.output.String(),
		"startedAt": session.startedAt,
		"endedAt":   session.endedAt,
	}
}

func (d *devWorkspace) consumeSessionRunUpdates() []sessionRunUpdate {
	d.mu.RLock()
	defer d.mu.RUnlock()
	updates := make([]sessionRunUpdate, 0, len(d.sessions))
	for _, session := range d.sessions {
		session.mu.Lock()
		if strings.TrimSpace(session.runID) == "" || session.reported || session.status == "running" {
			session.mu.Unlock()
			continue
		}
		updates = append(updates, sessionRunUpdate{
			runID:   session.runID,
			command: session.command,
			cwd:     session.cwd,
			status:  session.status,
		})
		session.reported = true
		session.mu.Unlock()
	}
	return updates
}

func (d *devWorkspace) sessionsList() []map[string]any {
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := make([]map[string]any, 0, len(d.sessionOrder))
	for _, id := range d.sessionOrder {
		if session, ok := d.sessions[id]; ok {
			out = append(out, d.sessionSnapshot(session))
		}
	}
	return out
}

func (d *devWorkspace) session(id string) (map[string]any, error) {
	d.mu.RLock()
	session, ok := d.sessions[id]
	d.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("session not found")
	}
	return d.sessionSnapshot(session), nil
}

func (d *devWorkspace) kill(id string) (map[string]any, error) {
	d.mu.RLock()
	session, ok := d.sessions[id]
	d.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("session not found")
	}
	if session.cancel != nil {
		session.cancel()
	}
	session.mu.Lock()
	session.status = "killed"
	session.endedAt = time.Now()
	session.exitCode = 130
	session.mu.Unlock()
	return d.sessionSnapshot(session), nil
}

func shellCommand(command string) (string, []string) {
	if _, err := exec.LookPath("zsh"); err == nil {
		return "zsh", []string{"-lc", command}
	}
	return "sh", []string{"-c", command}
}

func newPreviewProxy(port string) (*httputil.ReverseProxy, error) {
	target, err := url.Parse("http://127.0.0.1:" + strings.TrimSpace(port))
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
	}
	return proxy, nil
}

func isTextFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".go", ".js", ".ts", ".tsx", ".jsx", ".json", ".md", ".txt", ".html", ".css", ".yml", ".yaml", ".toml", ".sh", ".env", ".vue":
		return true
	default:
		return ext == ""
	}
}

func directoryHasTextChildren(path string) bool {
	found := false
	_ = filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if isTextFile(d.Name()) {
			found = true
			return io.EOF
		}
		return nil
	})
	return found
}
