package service

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"autocode-platform/internal/domain"
	"autocode-platform/internal/seed"
)

var (
	ErrRunConflict       = errors.New("run already exists")
	ErrRunNotFound       = errors.New("run not found")
	ErrSettingNotAllowed = errors.New("setting not allowed on current system profile")
)

type Platform struct {
	mu                sync.RWMutex
	runs              map[string]*domain.Run
	models            []domain.ModelProfile
	templates         []domain.Template
	featureToggles    map[string]domain.FeatureToggle
	builtinModelPacks map[string]domain.BuiltinModelPack
	toolInvocations   []domain.ToolInvocation
	subscribers       map[chan domain.Event]struct{}
	counter           uint64
	stopCh            chan struct{}
}

func NewPlatform() *Platform {
	p := &Platform{
		runs:              map[string]*domain.Run{},
		models:            decorateModels(seed.ModelCatalog()),
		templates:         seed.Templates(),
		featureToggles:    map[string]domain.FeatureToggle{},
		builtinModelPacks: map[string]domain.BuiltinModelPack{},
		toolInvocations:   []domain.ToolInvocation{},
		subscribers:       map[chan domain.Event]struct{}{},
		stopCh:            make(chan struct{}),
	}
	for _, toggle := range featureToggles() {
		p.featureToggles[toggle.ID] = toggle
	}
	for _, pack := range builtinModelPacks() {
		p.builtinModelPacks[pack.ID] = pack
	}
	go p.scheduler()
	return p
}

func (p *Platform) Close() {
	close(p.stopCh)
}

func (p *Platform) ListModels() []domain.ModelProfile {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return append([]domain.ModelProfile(nil), decorateModels(p.models)...)
}

func (p *Platform) Summary() domain.SystemSummary {
	p.mu.RLock()
	defer p.mu.RUnlock()
	profile := detectSystemProfile()
	totals := map[string]int{}
	for _, run := range p.runs {
		totals[string(run.Status)]++
	}
	return domain.SystemSummary{
		Providers:            providerStatuses(p.models),
		ToolCatalog:          sortedToolCatalog(),
		AtomicTools:          atomicToolCatalog(profile),
		ToolInvocations:      append([]domain.ToolInvocation(nil), p.toolInvocations...),
		TestCatalog:          sortedTestCatalog(),
		WorkflowOptions:      workflowOptionGroups(),
		FeatureToggles:       p.listFeatureTogglesLocked(profile),
		BuiltinModelPacks:    p.listBuiltinModelPacksLocked(profile),
		CapabilityProfiles:   capabilityProfiles(profile),
		OptimizationProfiles: optimizationProfiles(profile),
		RuntimePolicy:        runtimePolicy(profile),
		SystemProfile:        profile,
		RunTotals:            totals,
		SchedulerMode:        "serial-lightweight",
		DeployMode:           "idempotent-app-only",
		InterfaceModes:       []string{"atelier", "mission-control", "pocket"},
	}
}

func (p *Platform) ListAtomicTools() []domain.AtomicToolProfile {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return atomicToolCatalog(detectSystemProfile())
}

func (p *Platform) InvokeAtomicTool(id string) (domain.ToolInvocation, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	profile := detectSystemProfile()
	registry := atomicToolCatalog(profile)
	for _, item := range registry {
		if item.ID != id {
			continue
		}
		if !item.Allowed {
			return domain.ToolInvocation{}, ErrSettingNotAllowed
		}
		now := time.Now()
		invocation := domain.ToolInvocation{
			ID:        fmt.Sprintf("tool-%06d", atomic.AddUint64(&p.counter, 1)),
			ToolID:    item.ID,
			Title:     item.Name,
			Status:    "completed",
			Summary:   buildAtomicToolInvocationSummary(item, runtimePolicy(profile)),
			CreatedAt: now,
		}
		p.toolInvocations = append([]domain.ToolInvocation{invocation}, p.toolInvocations...)
		if len(p.toolInvocations) > 12 {
			p.toolInvocations = p.toolInvocations[:12]
		}
		p.publish(domain.Event{Type: domain.EventRunUpdated, Message: "已单独调用原子能力: " + item.Name, At: now})
		return invocation, nil
	}
	return domain.ToolInvocation{}, ErrRunNotFound
}

func (p *Platform) Analyze(input AnalyzeInput) domain.AnalysisPlan {
	return AnalyzeGoal(input)
}

func (p *Platform) Audit() domain.AuditReport {
	return buildAudit(p.Summary())
}

func (p *Platform) Advice(req domain.AdviceRequest) domain.AdviceReport {
	return buildAdvice(p.Summary(), req)
}

func (p *Platform) SetFeatureToggle(id string, enabled bool) (domain.FeatureToggle, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	toggle, ok := p.featureToggles[id]
	if !ok {
		return domain.FeatureToggle{}, ErrRunNotFound
	}
	decorated := decorateFeatureToggle(toggle, detectSystemProfile())
	if enabled && !decorated.Allowed {
		return decorated, ErrSettingNotAllowed
	}
	toggle.Enabled = enabled
	p.featureToggles[id] = toggle
	return decorateFeatureToggle(toggle, detectSystemProfile()), nil
}

func (p *Platform) SetBuiltinModelPack(id string, enabled bool, downloaded bool, remove bool) (domain.BuiltinModelPack, error) {
	return p.setBuiltinModelPackWithPolicy(id, enabled, downloaded, remove, runtimePolicy(detectSystemProfile()))
}

func (p *Platform) setBuiltinModelPackWithPolicy(id string, enabled bool, downloaded bool, remove bool, policy domain.RuntimePolicy) (domain.BuiltinModelPack, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	pack, ok := p.builtinModelPacks[id]
	if !ok {
		return domain.BuiltinModelPack{}, ErrRunNotFound
	}
	if remove {
		pack.InstallState = "removed"
		pack.Enabled = false
		pack.Downloaded = false
		p.builtinModelPacks[id] = pack
		p.publish(domain.Event{Type: domain.EventRunUpdated, Message: "本地原生模型已移除: " + pack.Name, At: now})
		return decorateBuiltinModelPack(pack, detectSystemProfile()), nil
	}
	if (enabled || (downloaded && !pack.Downloaded)) && !policy.AllowLocalModels {
		return decorateBuiltinModelPack(pack, detectSystemProfile()), fmt.Errorf("%w: runtime policy keeps local model packs in display-only mode", ErrSettingNotAllowed)
	}
	decorated := decorateBuiltinModelPack(pack, detectSystemProfile())
	if (enabled || (downloaded && !pack.Downloaded)) && !decorated.Allowed {
		return decorated, fmt.Errorf("%w: current system profile cannot activate this local model pack", ErrSettingNotAllowed)
	}
	if enabled && downloaded && !pack.Downloaded {
		pack.Enabled = true
		pack.Downloaded = false
		pack.InstallState = "queued"
		p.builtinModelPacks[id] = pack
		p.publish(domain.Event{Type: domain.EventRunUpdated, Message: "本地原生模型已加入部署队列: " + pack.Name, At: now})
		return decorateBuiltinModelPack(pack, detectSystemProfile()), nil
	}
	pack.Enabled = enabled
	if !enabled {
		pack.Enabled = false
	}
	if pack.Enabled && pack.Downloaded {
		pack.InstallState = "ready"
	} else if pack.Downloaded {
		pack.InstallState = "disabled"
	} else {
		pack.InstallState = "inactive"
	}
	p.builtinModelPacks[id] = pack
	p.publish(domain.Event{Type: domain.EventRunUpdated, Message: "本地原生模型状态已更新: " + pack.Name, At: now})
	return decorateBuiltinModelPack(pack, detectSystemProfile()), nil
}

func (p *Platform) ListTemplates() []domain.Template {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return append([]domain.Template(nil), p.templates...)
}

func (p *Platform) templateByID(id string) (domain.Template, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.templateByIDLocked(id)
}

func (p *Platform) templateByIDLocked(id string) (domain.Template, bool) {
	target := strings.TrimSpace(id)
	for _, item := range p.templates {
		if item.ID == target {
			return item, true
		}
	}
	return domain.Template{}, false
}

func (p *Platform) ListRuns() []domain.Run {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]domain.Run, 0, len(p.runs))
	for _, run := range p.runs {
		out = append(out, cloneRun(run))
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func (p *Platform) GetRun(id string) (domain.Run, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	run, ok := p.runs[id]
	if !ok {
		return domain.Run{}, ErrRunNotFound
	}
	return cloneRun(run), nil
}

func (p *Platform) CreateRun(input domain.CreateRunInput) (domain.Run, error) {
	return p.createRunWithPolicy(input, runtimePolicy(detectSystemProfile()))
}

func (p *Platform) createRunWithPolicy(input domain.CreateRunInput, policy domain.RuntimePolicy) (domain.Run, error) {
	template, hasTemplate := p.templateByID(input.TemplateID)
	if hasTemplate {
		input = applyTemplateDefaults(input, template)
	}
	analysis := AnalyzeGoal(AnalyzeInput{Title: input.Title, Goal: input.Goal})
	input, decisions := normalizeRunInputWithPolicy(input, analysis, policy)
	if hasTemplate {
		decisions = append([]domain.PolicyDecision{{
			Area:    "template",
			Action:  "applied",
			Reason:  template.ID,
			Message: "运行单已按模板骨架初始化: " + template.Name,
		}}, decisions...)
	}
	if source := strings.TrimSpace(input.TemplateSource); source != "" {
		decisions = append([]domain.PolicyDecision{{
			Area:    "template",
			Action:  "source",
			Reason:  source,
			Message: "模板来源已记录为 " + source,
		}}, decisions...)
	}
	profile := profileForRuntimePolicy(policy)
	assembledTools, assemblyDecisions := assembleAtomicTools(input, analysis, profile, preferredBundleID(template, hasTemplate, analysis))
	decisions = append(decisions, assemblyDecisions...)
	dedupKey := domain.DedupKey(input)
	now := time.Now()

	p.mu.Lock()
	defer p.mu.Unlock()
	for _, existing := range p.runs {
		if existing.DedupKey == dedupKey && (existing.Status == domain.RunQueued || existing.Status == domain.RunRunning) {
			return cloneRun(existing), ErrRunConflict
		}
	}

	id := fmt.Sprintf("run-%06d", atomic.AddUint64(&p.counter, 1))
	stages := make([]domain.Stage, 0, len(input.Stages))
	for index, kind := range input.Stages {
		stages = append(stages, domain.Stage{
			ID:        fmt.Sprintf("%s-stage-%d", id, index+1),
			Kind:      kind,
			Status:    domain.StagePending,
			Tools:     append([]domain.ToolConfig(nil), input.Tools...),
			UpdatedAt: now,
		})
	}
	run := &domain.Run{
		ID:                  id,
		Title:               strings.TrimSpace(input.Title),
		Goal:                strings.TrimSpace(input.Goal),
		TemplateID:          strings.TrimSpace(input.TemplateID),
		TemplateSource:      strings.TrimSpace(input.TemplateSource),
		Status:              domain.RunQueued,
		DedupKey:            dedupKey,
		LoopBlueprint:       append([]domain.LoopStep(nil), analysis.LoopBlueprint...),
		AssembledTools:      append([]domain.AtomicToolProfile(nil), assembledTools...),
		Stages:              stages,
		PolicyDecisions:     append([]domain.PolicyDecision(nil), decisions...),
		Analysis:            analysis,
		TestMode:            input.TestMode,
		SelectedTests:       append([]string(nil), input.SelectedTests...),
		AutoRepairEnabled:   input.AutoRepairEnabled,
		AutoRepairMode:      input.AutoRepairMode,
		RemoteDeployEnabled: input.RemoteDeployEnabled,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	initializeCheckpointState(run)
	run.Logs = append(run.Logs, domain.RunLog{Stage: domain.StageIntent, Level: "info", Message: "运行单进入队列，等待调度", At: now})
	for _, item := range decisions {
		run.Logs = append(run.Logs, domain.RunLog{Stage: domain.StageIntent, Level: "warn", Message: item.Message, At: now})
	}
	p.runs[id] = run
	p.publish(domain.Event{Type: domain.EventRunCreated, RunID: id, Message: "运行单已创建", At: now})
	return cloneRun(run), nil
}

func applyTemplateDefaults(input domain.CreateRunInput, template domain.Template) domain.CreateRunInput {
	if len(input.Stages) == 0 && len(template.DefaultStages) > 0 {
		input.Stages = append([]domain.StageKind(nil), template.DefaultStages...)
	}
	if len(input.Tools) == 0 && len(template.DefaultTools) > 0 {
		for _, name := range template.DefaultTools {
			input.Tools = append(input.Tools, domain.ToolConfig{Name: strings.TrimSpace(strings.ToLower(name)), Enabled: true})
		}
	}
	return input
}

func normalizeRunInputWithPolicy(input domain.CreateRunInput, analysis domain.AnalysisPlan, policy domain.RuntimePolicy) (domain.CreateRunInput, []domain.PolicyDecision) {
	decisions := []domain.PolicyDecision{}
	if len(input.Stages) == 0 {
		input.Stages = analysis.RecommendedStages
	}
	if len(input.Tools) == 0 {
		for _, name := range analysis.RecommendedTools {
			input.Tools = append(input.Tools, domain.ToolConfig{Name: name, Enabled: true})
		}
	}
	if strings.TrimSpace(input.TestMode) == "" {
		input.TestMode = defaultTestMode(policy)
	}
	if len(input.SelectedTests) == 0 {
		input.SelectedTests = append([]string(nil), analysis.RecommendedTests...)
	}

	disabledSet := map[string]bool{}
	for _, name := range policy.DefaultDisabledTools {
		disabledSet[strings.TrimSpace(strings.ToLower(name))] = true
	}

	toolMeta := map[string]domain.ToolProfile{}
	for _, item := range toolCatalog() {
		toolMeta[item.Name] = item
	}

	heavyUsed := 0
	normalizedTools := make([]domain.ToolConfig, 0, len(input.Tools))
	for _, tool := range domain.NormalizeTools(input.Tools) {
		enabled := tool.Enabled
		if disabledSet[tool.Name] {
			if enabled {
				decisions = append(decisions, domain.PolicyDecision{Area: "tool", Action: "disabled", Reason: "default-disabled", Message: "运行策略已关闭工具: " + tool.Name})
			}
			enabled = false
		}
		if enabled {
			if meta, ok := toolMeta[tool.Name]; ok && meta.Heavy {
				if tool.Name == "deploy" && !policy.AllowBackgroundJobs {
					decisions = append(decisions, domain.PolicyDecision{Area: "tool", Action: "disabled", Reason: "background-jobs-disabled", Message: "当前机器不允许后台重任务，已关闭 deploy 工具"})
					enabled = false
				} else if heavyUsed >= policy.MaxHeavyActions {
					decisions = append(decisions, domain.PolicyDecision{Area: "tool", Action: "disabled", Reason: "heavy-limit", Message: fmt.Sprintf("当前机器最多允许 %d 个重动作工具并行，已关闭 %s", policy.MaxHeavyActions, tool.Name)})
					enabled = false
				} else {
					heavyUsed++
				}
			}
		}
		normalizedTools = append(normalizedTools, domain.ToolConfig{Name: tool.Name, Enabled: enabled})
	}
	input.Tools = normalizedTools

	enabledTools := map[string]bool{}
	for _, tool := range input.Tools {
		enabledTools[tool.Name] = tool.Enabled
	}

	normalizedStages := make([]domain.StageKind, 0, len(input.Stages))
	seenStages := map[domain.StageKind]bool{}
	for _, stage := range input.Stages {
		if seenStages[stage] {
			continue
		}
		switch stage {
		case domain.StageDeploy:
			if !policy.AllowBackgroundJobs || !enabledTools["deploy"] {
				decisions = append(decisions, domain.PolicyDecision{Area: "stage", Action: "removed", Reason: "deploy-unavailable", Message: "运行策略已移除 deploy 阶段"})
				continue
			}
		case domain.StageTest:
			if !enabledTools["tests"] {
				decisions = append(decisions, domain.PolicyDecision{Area: "stage", Action: "removed", Reason: "tests-unavailable", Message: "tests 工具未启用，已移除 test 阶段"})
				continue
			}
		}
		seenStages[stage] = true
		normalizedStages = append(normalizedStages, stage)
	}
	input.Stages = normalizedStages

	if !policy.AllowBackgroundJobs || !enabledTools["deploy"] {
		if input.RemoteDeployEnabled {
			decisions = append(decisions, domain.PolicyDecision{Area: "deploy", Action: "disabled", Reason: "deploy-unavailable", Message: "运行策略已关闭远程部署"})
		}
		input.RemoteDeployEnabled = false
	}

	input.AutoRepairMode = normalizeAutoRepairMode(input.AutoRepairEnabled, input.AutoRepairMode, policy)
	if !input.AutoRepairEnabled && input.AutoRepairMode != "off" {
		decisions = append(decisions, domain.PolicyDecision{Area: "repair", Action: "downgraded", Reason: "repair-disabled", Message: "自动修复已关闭，修复强度已收束为 off"})
	}
	if input.AutoRepairEnabled {
		decisions = append(decisions, domain.PolicyDecision{Area: "repair", Action: "set", Reason: "runtime-policy", Message: "自动修复强度已设为 " + input.AutoRepairMode})
	}

	return input, dedupePolicyDecisions(decisions)
}

func defaultTestMode(policy domain.RuntimePolicy) string {
	if policy.Profile == "home-lite" {
		return "light"
	}
	if policy.ValidationDepth == "extended" {
		return "full"
	}
	return "template"
}

func normalizeAutoRepairMode(enabled bool, requested string, policy domain.RuntimePolicy) string {
	if !enabled {
		return "off"
	}
	mode := strings.TrimSpace(strings.ToLower(requested))
	if mode == "" || mode == "off" {
		switch policy.Profile {
		case "home-lite":
			return "lite"
		case "adaptive-performance":
			return "aggressive"
		default:
			return "standard"
		}
	}
	if mode != "lite" && mode != "standard" && mode != "aggressive" {
		switch policy.Profile {
		case "home-lite":
			return "lite"
		case "adaptive-performance":
			return "aggressive"
		default:
			return "standard"
		}
	}
	if policy.Profile == "home-lite" {
		return "lite"
	}
	if policy.Profile == "balanced-hybrid" && mode == "aggressive" {
		return "standard"
	}
	return mode
}

func profileForRuntimePolicy(policy domain.RuntimePolicy) domain.SystemProfile {
	profile := domain.SystemProfile{LocalModelAllowed: policy.AllowLocalModels}
	switch policy.Profile {
	case "home-lite":
		profile.Tier = "compact"
	case "adaptive-performance":
		profile.Tier = "performance"
	default:
		profile.Tier = "balanced"
	}
	return profile
}

func preferredBundleID(template domain.Template, hasTemplate bool, analysis domain.AnalysisPlan) string {
	if hasTemplate && strings.TrimSpace(template.RecommendedBundle.ID) != "" {
		return strings.TrimSpace(template.RecommendedBundle.ID)
	}
	return strings.TrimSpace(analysis.RecommendedBundle.ID)
}

func assembleAtomicTools(input domain.CreateRunInput, analysis domain.AnalysisPlan, profile domain.SystemProfile, bundleID string) ([]domain.AtomicToolProfile, []domain.PolicyDecision) {
	registry := atomicToolCatalog(profile)
	enabledTools := map[string]bool{}
	for _, tool := range domain.NormalizeTools(input.Tools) {
		enabledTools[tool.Name] = tool.Enabled
	}
	activeStages := map[domain.StageKind]bool{}
	for _, stage := range input.Stages {
		activeStages[stage] = true
	}
	selectedByGroup := map[string]domain.AtomicToolProfile{}
	decisions := []domain.PolicyDecision{}
	for _, item := range registry {
		if !item.Allowed || !atomicToolMatchesStages(item, activeStages) {
			continue
		}
		if !atomicToolMatchesProviders(item, enabledTools) {
			continue
		}
		group := strings.TrimSpace(item.DedupGroup)
		if group == "" {
			group = item.ID
		}
		current, exists := selectedByGroup[group]
		if !exists || shouldReplaceAtomicTool(current, item, enabledTools, bundleID) {
			selectedByGroup[group] = item
		}
	}
	assembled := make([]domain.AtomicToolProfile, 0, len(selectedByGroup))
	for _, item := range selectedByGroup {
		assembled = append(assembled, item)
	}
	sort.Slice(assembled, func(i, j int) bool {
		leftBundleRank := atomicToolBundleRank(assembled[i], bundleID)
		rightBundleRank := atomicToolBundleRank(assembled[j], bundleID)
		if leftBundleRank != rightBundleRank {
			return leftBundleRank < rightBundleRank
		}
		if assembled[i].Priority != assembled[j].Priority {
			return assembled[i].Priority < assembled[j].Priority
		}
		if assembled[i].Category != assembled[j].Category {
			return assembled[i].Category < assembled[j].Category
		}
		return assembled[i].ID < assembled[j].ID
	})
	for _, item := range assembled {
		reason := item.DedupGroup
		if atomicToolBundleRank(item, bundleID) == 0 && bundleID != "" {
			reason = bundleID
		}
		decisions = append(decisions, domain.PolicyDecision{
			Area:    "atomic-tool",
			Action:  "selected",
			Reason:  reason,
			Message: "已装配原子能力: " + item.Name + "，入口策略为 " + atomicToolPreferenceLabel(item),
		})
	}
	if len(assembled) == 0 {
		for _, recommended := range analysis.RecommendedTools {
			decisions = append(decisions, domain.PolicyDecision{Area: "atomic-tool", Action: "missing", Reason: "no-assembly", Message: "当前未能基于已启用工具装配原子能力，建议检查工具开关: " + recommended})
			break
		}
	}
	return assembled, dedupePolicyDecisions(decisions)
}

func atomicToolMatchesStages(item domain.AtomicToolProfile, activeStages map[domain.StageKind]bool) bool {
	for _, stage := range item.StageKinds {
		if activeStages[stage] {
			return true
		}
	}
	return false
}

func atomicToolMatchesProviders(item domain.AtomicToolProfile, enabledTools map[string]bool) bool {
	if enabledTools[item.PreferredProvider] {
		return true
	}
	for _, fallback := range item.FallbackProviders {
		if enabledTools[fallback] {
			return true
		}
	}
	return false
}

func shouldReplaceAtomicTool(current, candidate domain.AtomicToolProfile, enabledTools map[string]bool, bundleID string) bool {
	currentPreferred := enabledTools[current.PreferredProvider]
	candidatePreferred := enabledTools[candidate.PreferredProvider]
	if currentPreferred != candidatePreferred {
		return candidatePreferred
	}
	currentBundleRank := atomicToolBundleRank(current, bundleID)
	candidateBundleRank := atomicToolBundleRank(candidate, bundleID)
	if currentBundleRank != candidateBundleRank {
		return candidateBundleRank < currentBundleRank
	}
	if current.LocalFirst != candidate.LocalFirst {
		return candidate.LocalFirst
	}
	if current.Recommended != candidate.Recommended {
		return candidate.Recommended
	}
	if current.Priority != candidate.Priority {
		return candidate.Priority < current.Priority
	}
	return candidate.ID < current.ID
}

func atomicToolBundleRank(item domain.AtomicToolProfile, bundleID string) int {
	switch strings.TrimSpace(bundleID) {
	case "bundle-local-delivery":
		switch item.Category {
		case "release", "ops":
			return 0
		case "planner", "code", "quality":
			return 1
		}
	case "bundle-knowledge-office":
		switch item.Category {
		case "research", "knowledge", "office", "daily":
			return 0
		case "planner":
			return 1
		}
	case "bundle-creative-assets":
		switch item.Category {
		case "assets", "media":
			return 0
		case "research", "planner":
			return 1
		}
	case "bundle-local-core":
		switch item.Category {
		case "planner", "code", "quality", "ops":
			return 0
		}
	}
	return 2
}

func atomicToolPreferenceLabel(item domain.AtomicToolProfile) string {
	if item.LocalFirst {
		return "本地优先"
	}
	return "候补优先"
}

func autoRepairAttemptLimit(mode string) int {
	switch strings.TrimSpace(strings.ToLower(mode)) {
	case "lite":
		return 1
	case "aggressive":
		return 3
	case "standard":
		return 2
	default:
		return 0
	}
}

func dedupePolicyDecisions(values []domain.PolicyDecision) []domain.PolicyDecision {
	seen := map[string]bool{}
	out := make([]domain.PolicyDecision, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value.Message)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		value.Message = trimmed
		seen[trimmed] = true
		out = append(out, value)
	}
	return out
}

func (p *Platform) PauseRun(id string) (domain.Run, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	run, ok := p.runs[id]
	if !ok {
		return domain.Run{}, ErrRunNotFound
	}
	run.Status = domain.RunPaused
	run.UpdatedAt = time.Now()
	run.Logs = append(run.Logs, domain.RunLog{Stage: domain.StageIntent, Level: "warn", Message: "用户手动暂停当前运行单", At: run.UpdatedAt})
	for i := range run.Stages {
		if run.Stages[i].Status == domain.StageRunning {
			run.Stages[i].Status = domain.StagePaused
			run.Stages[i].UpdatedAt = run.UpdatedAt
		}
	}
	p.publish(domain.Event{Type: domain.EventRunUpdated, RunID: id, Message: "运行单已暂停", At: run.UpdatedAt})
	return cloneRun(run), nil
}

func (p *Platform) ResumeRun(id string) (domain.Run, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	run, ok := p.runs[id]
	if !ok {
		return domain.Run{}, ErrRunNotFound
	}
	run.Status = domain.RunRunning
	run.UpdatedAt = time.Now()
	run.Logs = append(run.Logs, domain.RunLog{Stage: domain.StageIntent, Level: "info", Message: "运行单恢复执行", At: run.UpdatedAt})
	for i := range run.Stages {
		if run.Stages[i].Status == domain.StagePaused {
			run.Stages[i].Status = domain.StagePending
			run.Stages[i].UpdatedAt = run.UpdatedAt
		}
	}
	p.publish(domain.Event{Type: domain.EventRunUpdated, RunID: id, Message: "运行单已恢复", At: run.UpdatedAt})
	return cloneRun(run), nil
}

func (p *Platform) PatchRequirements(id, extra string) (domain.Run, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	run, ok := p.runs[id]
	if !ok {
		return domain.Run{}, ErrRunNotFound
	}
	trimmed := strings.TrimSpace(extra)
	if trimmed == "" {
		return cloneRun(run), nil
	}
	if run.Goal != "" {
		run.Goal += "\n\n补充需求:\n" + trimmed
	} else {
		run.Goal = trimmed
	}
	run.Analysis = AnalyzeGoal(AnalyzeInput{Title: run.Title, Goal: run.Goal})
	run.LoopBlueprint = append([]domain.LoopStep(nil), run.Analysis.LoopBlueprint...)
	profile := profileForRuntimePolicy(runtimePolicy(detectSystemProfile()))
	template, hasTemplate := p.templateByIDLocked(run.TemplateID)
	assembledTools, assemblyDecisions := assembleAtomicTools(domain.CreateRunInput{
		Title:               run.Title,
		Goal:                run.Goal,
		TemplateID:          run.TemplateID,
		Stages:              currentRunStages(run.Stages),
		Tools:               activeRunTools(run.Stages),
		AutoRepairEnabled:   run.AutoRepairEnabled,
		AutoRepairMode:      run.AutoRepairMode,
		RemoteDeployEnabled: run.RemoteDeployEnabled,
	}, run.Analysis, profile, preferredBundleID(template, hasTemplate, run.Analysis))
	run.AssembledTools = assembledTools
	run.PolicyDecisions = dedupePolicyDecisions(append(run.PolicyDecisions, assemblyDecisions...))
	resetKinds := map[domain.StageKind]bool{
		domain.StagePlan:      true,
		domain.StageResource:  true,
		domain.StageModel:     true,
		domain.StageTool:      true,
		domain.StageImplement: true,
		domain.StageResult:    true,
		domain.StageTest:      true,
		domain.StageDeploy:    true,
		domain.StageRepair:    true,
		domain.StageFinalize:  true,
	}
	now := time.Now()
	for i := range run.Stages {
		if resetKinds[run.Stages[i].Kind] {
			run.Stages[i].Status = domain.StagePending
			run.Stages[i].Summary = "等待基于最新需求重新执行"
			run.Stages[i].UpdatedAt = now
		}
	}
	run.TestReports = nil
	run.Status = domain.RunPaused
	run.UpdatedAt = now
	run.Failures = nil
	run.Deployments = nil
	run.Logs = append(run.Logs, domain.RunLog{Stage: domain.StagePlan, Level: "info", Message: "已接收补充需求并重置受影响阶段", At: now})
	setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-implementation", "pending")
	setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-validation", "pending")
	setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-delivery", "pending")
	setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-preview", "pending")
	setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-security", "pending")
	p.publish(domain.Event{Type: domain.EventRequirementPatched, RunID: id, Message: "运行单已补充需求", At: now})
	return cloneRun(run), nil
}

func (p *Platform) RollbackDeployment(id, version string) (domain.Run, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	run, ok := p.runs[id]
	if !ok {
		return domain.Run{}, ErrRunNotFound
	}
	targetVersion := strings.TrimSpace(version)
	if targetVersion == "" {
		return domain.Run{}, errors.New("missing rollback version")
	}
	var matched *domain.DeploymentRecord
	for i := range run.Deployments {
		if run.Deployments[i].Version == targetVersion {
			matched = &run.Deployments[i]
			break
		}
	}
	if matched == nil {
		return domain.Run{}, errors.New("deployment version not found")
	}
	now := time.Now()
	run.Deployments = append(run.Deployments, domain.DeploymentRecord{
		Mode:      "rollback",
		Status:    "completed",
		Target:    matched.Target,
		Version:   matched.Version,
		Summary:   "已执行版本回退，可继续补做验证与预览确认。",
		CreatedAt: now,
	})
	run.UpdatedAt = now
	run.Logs = append(run.Logs, domain.RunLog{Stage: domain.StageDeploy, Level: "warn", Message: "已回退到稳定部署版本 " + matched.Version, At: now})
	setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-delivery", "completed")
	setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-preview", "pending")
	resolveMatchingFailures(run, func(failure domain.FailureSummary) bool {
		return failure.Category == "delivery" && (failure.Target == "stable-deploy" || failure.Target == "deploy-smoke")
	})
	p.publish(domain.Event{Type: domain.EventRunUpdated, RunID: id, Message: "运行单已回退到版本 " + matched.Version, At: now})
	return cloneRun(run), nil
}

func (p *Platform) RevalidateRun(id string) (domain.Run, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	run, ok := p.runs[id]
	if !ok {
		return domain.Run{}, ErrRunNotFound
	}
	now := time.Now()
	latestTarget := "configured-vps"
	latestVersion := ""
	if len(run.Deployments) > 0 {
		latestTarget = run.Deployments[len(run.Deployments)-1].Target
		latestVersion = run.Deployments[len(run.Deployments)-1].Version
	}
	run.TestReports = synthesizeTestReports(run, now, false)
	run.Deployments = append(run.Deployments, domain.DeploymentRecord{
		Mode:      "revalidate",
		Status:    "completed",
		Target:    latestTarget,
		Version:   latestVersion,
		Summary:   "已完成关键复验，待确认预览与服务状态。",
		CreatedAt: now,
	})
	setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-validation", "completed")
	if runIncludesTest(run, "security-scan") {
		setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-security", "completed")
	}
	setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-preview", "pending")
	resolveMatchingFailures(run, func(failure domain.FailureSummary) bool {
		if failure.Category == "validation" {
			return true
		}
		return failure.Category == "delivery" && failure.Target == "deploy-smoke"
	})
	run.UpdatedAt = now
	run.Logs = append(run.Logs, domain.RunLog{Stage: domain.StageTest, Level: "info", Message: "已完成回退后的关键验证，待确认预览与服务状态", At: now})
	p.publish(domain.Event{Type: domain.EventRunUpdated, RunID: id, Message: "运行单已完成回退后复验", At: now})
	return cloneRun(run), nil
}

func (p *Platform) DeployRun(id string) (domain.Run, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	run, ok := p.runs[id]
	if !ok {
		return domain.Run{}, ErrRunNotFound
	}
	if !run.RemoteDeployEnabled {
		return domain.Run{}, errors.New("remote deploy is disabled for this run")
	}
	if len(unresolvedFailures(run.Failures)) > 0 {
		return domain.Run{}, errors.New("run still has unresolved failures")
	}
	now := time.Now()
	version := domain.VersionTag(now)
	run.Deployments = append(run.Deployments, domain.DeploymentRecord{
		Mode:      "remote",
		Status:    "completed",
		Target:    "configured-vps",
		Version:   version,
		Summary:   "已触发稳定部署，建议继续确认预览与服务状态。",
		CreatedAt: now,
	})
	for i := range run.Stages {
		if run.Stages[i].Kind != domain.StageDeploy {
			continue
		}
		run.Stages[i].Status = domain.StageCompleted
		run.Stages[i].Summary = "已手动执行稳定部署，待确认预览与服务状态"
		run.Stages[i].UpdatedAt = now
		break
	}
	setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-delivery", "completed")
	setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-preview", "pending")
	resolveMatchingFailures(run, func(failure domain.FailureSummary) bool {
		return failure.Category == "delivery" && failure.Target == "stable-deploy"
	})
	run.UpdatedAt = now
	run.Logs = append(run.Logs, domain.RunLog{Stage: domain.StageDeploy, Level: "info", Message: "已执行稳定部署，当前版本 " + version, At: now})
	p.publish(domain.Event{Type: domain.EventRunUpdated, RunID: id, Message: "运行单已执行稳定部署", At: now})
	return cloneRun(run), nil
}

func (p *Platform) RecordDevActivity(id string, input domain.DevActivityInput) (domain.Run, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	run, ok := p.runs[id]
	if !ok {
		return domain.Run{}, ErrRunNotFound
	}
	now := time.Now()
	activity := domain.DevActivity{
		Kind:    strings.TrimSpace(strings.ToLower(input.Kind)),
		Target:  strings.TrimSpace(input.Target),
		Detail:  strings.TrimSpace(input.Detail),
		Status:  strings.TrimSpace(strings.ToLower(input.Status)),
		Command: strings.TrimSpace(input.Command),
		At:      now,
	}
	if activity.Kind == "" {
		activity.Kind = "note"
	}
	if activity.Status == "" {
		activity.Status = "completed"
	}
	run.DevActivities = append([]domain.DevActivity{activity}, run.DevActivities...)
	if len(run.DevActivities) > 24 {
		run.DevActivities = run.DevActivities[:24]
	}
	stage, level, message := summarizeDevActivity(activity)
	run.Logs = append(run.Logs, domain.RunLog{Stage: stage, Level: level, Message: message, At: now})
	applyDevCheckpointProgress(run, activity)
	if activity.Kind == "preview-open" && activity.Status == "completed" {
		resolveMatchingFailures(run, func(failure domain.FailureSummary) bool {
			return failure.Category == "delivery" && (failure.Target == "preview-check" || failure.Target == "preview-confirmation")
		})
	}
	run.UpdatedAt = now
	p.publish(domain.Event{Type: domain.EventRunUpdated, RunID: id, Message: message, At: now})
	return cloneRun(run), nil
}

func (p *Platform) Subscribe() (chan domain.Event, func()) {
	ch := make(chan domain.Event, 16)
	p.mu.Lock()
	p.subscribers[ch] = struct{}{}
	p.mu.Unlock()
	return ch, func() {
		p.mu.Lock()
		delete(p.subscribers, ch)
		close(ch)
		p.mu.Unlock()
	}
}

func (p *Platform) scheduler() {
	ticker := time.NewTicker(800 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.tick()
		case <-p.stopCh:
			return
		}
	}
}

func (p *Platform) tick() {
	profile := detectSystemProfile()
	p.tickWithPolicy(runtimePolicy(profile))
}

func (p *Platform) TickForTest() {
	p.tick()
}

func (p *Platform) tickWithPolicy(policy domain.RuntimePolicy) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.tickLocked(policy)
}

func (p *Platform) tickLocked(policy domain.RuntimePolicy) {
	p.advanceBuiltinModelPacks()
	runs := p.orderedRunsLocked()
	activeRuns := p.activeRunCountLocked()
	for _, run := range runs {
		if run.Status == domain.RunPaused || run.Status == domain.RunCompleted || run.Status == domain.RunFailed {
			continue
		}
		if run.Status == domain.RunQueued {
			if activeRuns >= maxConcurrentRuns(policy) {
				continue
			}
			run.Status = domain.RunRunning
			run.Logs = append(run.Logs, domain.RunLog{Stage: domain.StageIntent, Level: "info", Message: fmt.Sprintf("调度器已接管运行单，当前按 %s 级别最多并行 %d 条运行单", policy.Profile, maxConcurrentRuns(policy)), At: time.Now()})
			activeRuns++
		}
		advanced := p.advanceRun(run)
		if advanced {
			return
		}
	}
}

func (p *Platform) orderedRunsLocked() []*domain.Run {
	out := make([]*domain.Run, 0, len(p.runs))
	for _, run := range p.runs {
		out = append(out, run)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

func (p *Platform) activeRunCountLocked() int {
	count := 0
	for _, run := range p.runs {
		if run.Status == domain.RunRunning {
			count++
		}
	}
	return count
}

func maxConcurrentRuns(policy domain.RuntimePolicy) int {
	if policy.MaxConcurrentRuns <= 0 {
		return 1
	}
	return policy.MaxConcurrentRuns
}

func (p *Platform) advanceBuiltinModelPacks() {
	now := time.Now()
	for id, pack := range p.builtinModelPacks {
		updated := false
		message := ""
		switch pack.InstallState {
		case "queued":
			pack.InstallState = "downloading"
			updated = true
			message = "本地原生模型开始下载: " + pack.Name
		case "downloading":
			pack.InstallState = "configuring"
			updated = true
			message = "本地原生模型开始配置: " + pack.Name
		case "configuring":
			pack.InstallState = "ready"
			pack.Downloaded = true
			pack.Enabled = true
			updated = true
			message = "本地原生模型已就绪: " + pack.Name
		}
		if updated {
			p.builtinModelPacks[id] = pack
			p.publish(domain.Event{Type: domain.EventRunUpdated, Message: message, At: now})
		}
	}
}

func (p *Platform) advanceRun(run *domain.Run) bool {
	stageIndex := nextRunnableStageIndex(run)
	if stageIndex < 0 {
		return false
	}
	stage := &run.Stages[stageIndex]
	{
		stage.Status = domain.StageRunning
		now := time.Now()
		stage.UpdatedAt = now
		stage.Summary = p.summaryForStage(run, stage.Kind)
		run.Logs = append(run.Logs, domain.RunLog{Stage: stage.Kind, Level: "info", Message: "阶段开始执行: " + string(stage.Kind), At: now})
		if p.shouldFail(run, stage.Kind) {
			stage.Status = domain.StageFailed
			stage.Summary = "阶段执行失败，已生成失败摘要"
			run.Status = domain.RunFailed
			failures := p.failuresForStage(run, stage.Kind, now)
			run.Failures = append(run.Failures, failures...)
			if stage.Kind == domain.StageTest {
				run.TestReports = synthesizeTestReports(run, now, true)
			}
			if len(failures) > 0 {
				run.Logs = append(run.Logs, domain.RunLog{Stage: stage.Kind, Level: "error", Message: failures[0].Reason, At: now})
			}
			if run.AutoRepairEnabled && stage.Kind != domain.StageRepair && run.RepairAttempts < autoRepairAttemptLimit(run.AutoRepairMode) {
				run.RepairAttempts++
				stage.Status = domain.StagePaused
				stage.Summary = "阶段执行失败，已转入修复回路"
				p.enqueueRepair(run, now)
				run.Status = domain.RunRunning
				run.Logs = append(run.Logs, domain.RunLog{Stage: domain.StageRepair, Level: "info", Message: "已自动加入修复阶段，当前强度: " + run.AutoRepairMode, At: now})
			}
		} else {
			stage.Status = domain.StageCompleted
			run.Logs = append(run.Logs, domain.RunLog{Stage: stage.Kind, Level: "info", Message: "阶段执行完成: " + string(stage.Kind), At: now})
			advanceCheckpointForStage(run, stage.Kind)
			if stage.Kind == domain.StageTest {
				run.TestReports = synthesizeTestReports(run, now, false)
			}
			if stage.Kind == domain.StageRepair {
				stage.Summary = p.completeRepair(run)
			}
			if stage.Kind == domain.StageDeploy && run.RemoteDeployEnabled {
				run.Deployments = append(run.Deployments, domain.DeploymentRecord{
					Mode:      "remote",
					Status:    "completed",
					Target:    "configured-vps",
					Version:   domain.VersionTag(now),
					Summary:   "已生成稳定部署记录，可映射到部署脚本执行",
					CreatedAt: now,
				})
			}
		}
		stage.UpdatedAt = now
		run.UpdatedAt = now
		if allStagesDone(run.Stages) {
			run.Status = domain.RunCompleted
		}
		p.publish(domain.Event{Type: domain.EventStageUpdated, RunID: run.ID, Message: string(stage.Kind) + " 已更新", At: now})
		return true
	}
}

func nextRunnableStageIndex(run *domain.Run) int {
	if len(unresolvedFailures(run.Failures)) > 0 {
		for i := range run.Stages {
			if run.Stages[i].Kind == domain.StageRepair && run.Stages[i].Status != domain.StageCompleted && run.Stages[i].Status != domain.StageSkipped {
				return i
			}
		}
	}
	for i := range run.Stages {
		stage := run.Stages[i]
		if stage.Status == domain.StageCompleted || stage.Status == domain.StageSkipped || stage.Status == domain.StagePaused {
			continue
		}
		return i
	}
	return -1
}

func (p *Platform) repairSuggestion(kind domain.StageKind) string {
	switch kind {
	case domain.StageTest:
		return "读取失败测试摘要，重新生成修复补丁并复跑测试"
	case domain.StageDeploy:
		return "检查部署脚本、目标目录权限和应用启动日志"
	case domain.StageImplement, domain.StageResult:
		return "回看本轮产出差异与相关文件，修正后重新整理结果"
	default:
		return "回溯该阶段输入上下文并重新执行"
	}
}

func (p *Platform) shouldFail(run *domain.Run, kind domain.StageKind) bool {
	if kind != domain.StageRepair && hasCompletedRepair(run) {
		return false
	}
	goal := strings.ToLower(run.Goal)
	if strings.Contains(goal, "[fail:"+string(kind)+"]") {
		return true
	}
	if stageFailureTarget(run, kind) != "" {
		return true
	}
	return false
}

func hasCompletedRepair(run *domain.Run) bool {
	for _, stage := range run.Stages {
		if stage.Kind == domain.StageRepair && stage.Status == domain.StageCompleted {
			return true
		}
	}
	return false
}

func (p *Platform) enqueueRepair(run *domain.Run, now time.Time) {
	for i := range run.Stages {
		if run.Stages[i].Kind == domain.StageRepair && run.Stages[i].Status == domain.StagePending {
			return
		}
	}
	run.Stages = append(run.Stages, domain.Stage{
		ID:        fmt.Sprintf("%s-repair-%d", run.ID, run.RepairAttempts),
		Kind:      domain.StageRepair,
		Status:    domain.StagePending,
		Summary:   "等待自动修复",
		UpdatedAt: now,
	})
}

func (p *Platform) summaryForStage(run *domain.Run, kind domain.StageKind) string {
	toolSummary := summarizeStageAtomicTools(run.AssembledTools, kind)
	switch kind {
	case domain.StageIntent:
		return "已确认任务意图、交付方式与优先方向" + toolSummary
	case domain.StageContext:
		return "已汇总现有结构、限制条件与关键上下文" + toolSummary
	case domain.StagePlan:
		return "已整理执行路径、阶段顺序与风险收口策略" + toolSummary
	case domain.StageResource:
		return "已根据当前机器规格收束并发、重动作和运行边界" + toolSummary
	case domain.StageModel:
		return "已选定本轮任务的模型入口与回退路线" + toolSummary
	case domain.StageTool:
		return "已装配当前任务真正需要的原子能力" + toolSummary
	case domain.StageImplement:
		return "已进入核心产出执行，准备落地代码、文档或配置变更" + toolSummary
	case domain.StageResult:
		return "已整理本轮产出、差异摘要与后续交付材料" + toolSummary
	case domain.StageTest:
		return "已按 " + run.TestMode + " 模式执行质量检查与测试链" + toolSummary
	case domain.StageDeploy:
		if run.RemoteDeployEnabled {
			return "已准备部署产物并等待或执行稳定部署" + toolSummary
		}
		return "部署阶段已启用，但当前运行单未开启稳定部署" + toolSummary
	case domain.StageRepair:
		return "已根据失败摘要进入自动修复闭环" + toolSummary
	case domain.StageFinalize:
		return "已沉淀本轮关键结论、证据与下一步建议" + toolSummary
	default:
		return "阶段已处理"
	}
}

func summarizeStageAtomicTools(tools []domain.AtomicToolProfile, kind domain.StageKind) string {
	matched := make([]string, 0, len(tools))
	for _, item := range tools {
		for _, stage := range item.StageKinds {
			if stage == kind {
				matched = append(matched, item.Name)
				break
			}
		}
	}
	if len(matched) == 0 {
		return ""
	}
	return "，当前装配能力包括 " + strings.Join(matched, "、")
}

func (p *Platform) publish(event domain.Event) {
	for ch := range p.subscribers {
		select {
		case ch <- event:
		default:
		}
	}
}

func allStagesDone(stages []domain.Stage) bool {
	for _, stage := range stages {
		if stage.Status != domain.StageCompleted && stage.Status != domain.StageSkipped {
			return false
		}
	}
	return true
}

func synthesizeTestReports(run *domain.Run, now time.Time, failing bool) []domain.TestReport {
	selected := map[string]bool{}
	for _, item := range run.SelectedTests {
		selected[strings.TrimSpace(strings.ToLower(item))] = true
	}
	if len(selected) == 0 {
		for _, item := range run.Analysis.RecommendedTests {
			selected[strings.TrimSpace(strings.ToLower(item))] = true
		}
	}
	catalog := sortedTestCatalog()
	reports := make([]domain.TestReport, 0, len(catalog))
	failingTargets := failingTestTargets(run)
	for _, item := range catalog {
		name := strings.TrimSpace(strings.ToLower(item.Name))
		if !selected[name] {
			continue
		}
		status := "passed"
		summary := testSummaryForName(item.Name)
		if failing && (len(failingTargets) == 0 || failingTargets[name]) {
			status = "failed"
			summary = testFailureSummaryForName(item.Name)
		}
		reports = append(reports, domain.TestReport{
			Name:      item.Name,
			Layer:     item.Layer,
			Status:    status,
			Duration:  testDurationForMode(run.TestMode, item.Name),
			Summary:   summary,
			Completed: now,
		})
	}
	if len(reports) == 0 {
		status := "passed"
		summary := "基础结构复核通过"
		if failing {
			status = "failed"
			summary = "基础结构复核未通过，需要先处理当前问题。"
		}
		reports = append(reports, domain.TestReport{Name: "review-check", Layer: "quality", Status: status, Duration: 160 * time.Millisecond, Summary: summary, Completed: now})
	}
	return reports
}

func failingTestTargets(run *domain.Run) map[string]bool {
	goal := strings.ToLower(run.Goal)
	targets := map[string]bool{}
	marker := "[fail:test:"
	remaining := goal
	for {
		start := strings.Index(remaining, marker)
		if start < 0 {
			break
		}
		fragment := remaining[start+len(marker):]
		end := strings.Index(fragment, "]")
		if end < 0 {
			break
		}
		target := strings.TrimSpace(fragment[:end])
		if target != "" {
			targets[target] = true
		}
		remaining = fragment[end+1:]
	}
	return targets
}

func testFailureSummaryForName(name string) string {
	switch name {
	case "static-check":
		return "静态结构检查未通过，需要先修正语法或结构问题"
	case "lint":
		return "规范检查未通过，需要先处理格式或风格问题"
	case "review-check":
		return "结构复核未通过，需要先补齐缺失内容或逻辑"
	case "code-review":
		return "代码审查发现阻断问题，需要先修正后再继续"
	case "security-scan":
		return "安全扫描发现风险，需要先处理敏感信息或潜在漏洞"
	case "unit-tests":
		return "单元测试未通过，需要先修正核心逻辑"
	case "integration-tests":
		return "集成测试未通过，需要先修正模块联调问题"
	case "e2e-smoke":
		return "关键流程验证未通过，需要先修正实际操作链路"
	case "deploy-smoke":
		return "部署后检查未通过，需要先确认服务启动与健康状态"
	default:
		return "当前验证项未通过，需要先修正后再继续"
	}
}

func (p *Platform) failuresForStage(run *domain.Run, kind domain.StageKind, now time.Time) []domain.FailureSummary {
	if kind == domain.StageTest {
		reports := synthesizeTestReports(run, now, true)
		failures := make([]domain.FailureSummary, 0, len(reports))
		for _, report := range reports {
			if report.Status != "failed" {
				continue
			}
			failures = append(failures, domain.FailureSummary{
				Stage:      kind,
				Category:   "validation",
				Target:     report.Name,
				Reason:     report.Name + " 未通过: " + report.Summary,
				Suggestion: p.repairSuggestion(kind),
				Resolved:   false,
				At:         now,
			})
		}
		if len(failures) > 0 {
			return failures
		}
	}
	return []domain.FailureSummary{{
		Stage:      kind,
		Category:   failureCategoryForStage(kind),
		Target:     failureTargetForStage(run, kind),
		Reason:     failureReasonForStage(run, kind),
		Suggestion: p.repairSuggestionForTarget(kind, failureTargetForStage(run, kind)),
		Resolved:   false,
		At:         now,
	}}
}

func (p *Platform) completeRepair(run *domain.Run) string {
	previousFailures := append([]domain.FailureSummary(nil), run.Failures...)
	affected := map[domain.StageKind]bool{}
	resolvedCount := 0
	for i, failure := range previousFailures {
		if failure.Resolved {
			continue
		}
		affected[failure.Stage] = true
		previousFailures[i].Resolved = true
		resolvedCount++
	}
	for i := range run.Stages {
		if !affected[run.Stages[i].Kind] {
			continue
		}
		if run.Stages[i].Status == domain.StagePaused || run.Stages[i].Status == domain.StageFailed {
			run.Stages[i].Status = domain.StagePending
			run.Stages[i].Summary = "已完成一轮修复，等待重新执行"
			run.Stages[i].UpdatedAt = time.Now()
		}
	}
	run.Failures = previousFailures
	run.TestReports = nil
	if resolvedCount == 0 {
		return "已完成一轮自动修复，待继续后续阶段"
	}
	return "已根据失败摘要完成一轮自动修复，待重新执行受影响阶段"
}

func unresolvedFailures(failures []domain.FailureSummary) []domain.FailureSummary {
	active := make([]domain.FailureSummary, 0, len(failures))
	for _, failure := range failures {
		if failure.Resolved {
			continue
		}
		active = append(active, failure)
	}
	return active
}

func resolveMatchingFailures(run *domain.Run, match func(domain.FailureSummary) bool) int {
	resolved := 0
	for i := range run.Failures {
		if run.Failures[i].Resolved {
			continue
		}
		if !match(run.Failures[i]) {
			continue
		}
		run.Failures[i].Resolved = true
		resolved++
	}
	return resolved
}

func stageFailureTarget(run *domain.Run, kind domain.StageKind) string {
	goal := strings.ToLower(run.Goal)
	marker := "[fail:" + string(kind) + ":"
	start := strings.Index(goal, marker)
	if start < 0 {
		return ""
	}
	fragment := goal[start+len(marker):]
	end := strings.Index(fragment, "]")
	if end < 0 {
		return ""
	}
	return strings.TrimSpace(fragment[:end])
}

func failureCategoryForStage(kind domain.StageKind) string {
	switch kind {
	case domain.StageTest:
		return "validation"
	case domain.StageDeploy:
		return "delivery"
	case domain.StageRepair:
		return "repair"
	default:
		return "execution"
	}
}

func failureTargetForStage(run *domain.Run, kind domain.StageKind) string {
	if target := stageFailureTarget(run, kind); target != "" {
		return target
	}
	switch kind {
	case domain.StageDeploy:
		return "stable-deploy"
	default:
		return string(kind)
	}
}

func failureReasonForStage(run *domain.Run, kind domain.StageKind) string {
	target := failureTargetForStage(run, kind)
	if kind == domain.StageDeploy {
		switch target {
		case "preview-check", "preview-confirmation":
			return "预览确认未完成，当前版本还不能作为稳定交付结果"
		case "deploy-smoke":
			return "部署后检查未通过，需要先确认服务启动与健康状态"
		case "stable-deploy":
			return "稳定部署执行异常，需要先检查部署脚本、目标目录和版本切换过程"
		}
	}
	return "检测到模拟失败标记或执行异常"
}

func (p *Platform) repairSuggestionForTarget(kind domain.StageKind, target string) string {
	if kind == domain.StageDeploy {
		switch target {
		case "preview-check", "preview-confirmation":
			return "先完成预览确认，再补记当前版本表现与服务状态"
		case "deploy-smoke":
			return "检查服务启动日志、健康检查结果和当前生效版本"
		case "stable-deploy":
			return "检查部署脚本、目标目录权限和应用启动日志"
		}
	}
	return p.repairSuggestion(kind)
}

func initializeCheckpointState(run *domain.Run) {
	setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-requirement", "completed")
	setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-design", "completed")
	if len(run.Analysis.Checkpoints) == 0 {
		return
	}
	if hasImplementedActivity(run) {
		setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-implementation", "in_progress")
	}
}

func advanceCheckpointForStage(run *domain.Run, kind domain.StageKind) {
	switch kind {
	case domain.StageIntent, domain.StageContext:
		setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-requirement", "completed")
	case domain.StagePlan, domain.StageResource, domain.StageModel, domain.StageTool:
		setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-design", "completed")
	case domain.StageImplement, domain.StageResult, domain.StageRepair:
		setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-implementation", "completed")
	case domain.StageTest:
		setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-validation", "completed")
		if runIncludesTest(run, "security-scan") {
			setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-security", "completed")
		}
	case domain.StageDeploy:
		setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-delivery", "completed")
	}
}

func applyDevCheckpointProgress(run *domain.Run, activity domain.DevActivity) {
	switch activity.Kind {
	case "file-save", "file-rollback":
		setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-implementation", "in_progress")
	case "command":
		command := strings.ToLower(activity.Command)
		if hasAny(command, "test", "lint", "review", "scan", "check") {
			setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-validation", "in_progress")
		}
		if hasAny(command, "security", "gosec", "npm audit") {
			setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-security", "in_progress")
		}
		if hasAny(command, "build", "deploy", "release") {
			setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-delivery", "in_progress")
		}
		if activity.Status == "completed" && hasAny(command, "test", "lint", "review", "scan", "check") {
			setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-validation", "completed")
		}
		if activity.Status == "completed" && hasAny(command, "security", "gosec", "npm audit") {
			setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-security", "completed")
		}
	case "preview-open":
		setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-preview", "completed")
		setCheckpointStatus(&run.Analysis.Checkpoints, "checkpoint-delivery", "in_progress")
	}
}

func summarizeDevActivity(activity domain.DevActivity) (domain.StageKind, string, string) {
	switch activity.Kind {
	case "file-save":
		return domain.StageResult, "info", "开发工作台已保存文件: " + fallbackActivityTarget(activity.Target)
	case "file-rollback":
		return domain.StageRepair, "warn", "开发工作台已回滚文件: " + fallbackActivityTarget(activity.Target)
	case "preview-open":
		return domain.StageDeploy, "info", "开发工作台已加载预览端口: " + fallbackActivityTarget(activity.Target)
	case "command":
		stage := inferCommandStage(activity.Command)
		prefix := "开发工作台已执行命令"
		if activity.Status == "running" {
			prefix = "开发工作台已启动后台命令"
		}
		if activity.Status == "failed" || activity.Status == "timeout" || activity.Status == "killed" {
			return stage, "warn", fmt.Sprintf("%s: %s (%s)", prefix, fallbackActivityTarget(activity.Command), activity.Status)
		}
		return stage, "info", fmt.Sprintf("%s: %s (%s)", prefix, fallbackActivityTarget(activity.Command), activity.Status)
	default:
		return domain.StageResult, "info", strings.TrimSpace(activity.Detail)
	}
}

func inferCommandStage(command string) domain.StageKind {
	trimmed := strings.ToLower(strings.TrimSpace(command))
	switch {
	case hasAny(trimmed, "test", "lint", "review", "scan", "check"):
		return domain.StageTest
	case hasAny(trimmed, "deploy", "release", "rsync"):
		return domain.StageDeploy
	default:
		return domain.StageImplement
	}
}

func fallbackActivityTarget(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "未命名目标"
	}
	return trimmed
}

func setCheckpointStatus(checkpoints *[]domain.FlowCheckpoint, id string, status string) {
	for i := range *checkpoints {
		if (*checkpoints)[i].ID == id {
			(*checkpoints)[i].Status = status
			return
		}
	}
}

func runIncludesTest(run *domain.Run, target string) bool {
	selected := run.SelectedTests
	if len(selected) == 0 {
		selected = run.Analysis.RecommendedTests
	}
	for _, item := range selected {
		if strings.EqualFold(strings.TrimSpace(item), target) {
			return true
		}
	}
	return false
}

func hasImplementedActivity(run *domain.Run) bool {
	for _, item := range run.DevActivities {
		if item.Kind == "file-save" || item.Kind == "command" {
			return true
		}
	}
	return false
}

func testDurationForMode(mode string, name string) time.Duration {
	base := map[string]time.Duration{
		"static-check":      120 * time.Millisecond,
		"lint":              160 * time.Millisecond,
		"review-check":      180 * time.Millisecond,
		"code-review":       240 * time.Millisecond,
		"security-scan":     260 * time.Millisecond,
		"unit-tests":        420 * time.Millisecond,
		"integration-tests": 680 * time.Millisecond,
		"e2e-smoke":         820 * time.Millisecond,
		"deploy-smoke":      420 * time.Millisecond,
		"consistency-check": 200 * time.Millisecond,
	}[name]
	if base == 0 {
		base = 200 * time.Millisecond
	}
	switch strings.TrimSpace(strings.ToLower(mode)) {
	case "light", "smoke":
		return base / 2
	case "full":
		return base + base/3
	default:
		return base
	}
}

func testSummaryForName(name string) string {
	switch name {
	case "static-check":
		return "语法与静态结构检查通过"
	case "lint":
		return "代码规范与格式检查通过"
	case "review-check":
		return "结构化复核通过"
	case "code-review":
		return "AI 代码审查未发现阻断问题"
	case "security-scan":
		return "敏感信息与安全风险扫描通过"
	case "unit-tests":
		return "核心逻辑单元测试通过"
	case "integration-tests":
		return "接口与模块联调通过"
	case "e2e-smoke":
		return "关键流程冒烟验证通过"
	case "deploy-smoke":
		return "部署后健康检查通过"
	case "consistency-check":
		return "交付一致性检查通过"
	default:
		return "测试项执行通过"
	}
}

func cloneRun(run *domain.Run) domain.Run {
	copyRun := *run
	copyRun.Analysis = cloneAnalysisPlan(run.Analysis)
	copyRun.Stages = append([]domain.Stage(nil), run.Stages...)
	copyRun.LoopBlueprint = append([]domain.LoopStep(nil), run.LoopBlueprint...)
	copyRun.AssembledTools = append([]domain.AtomicToolProfile(nil), run.AssembledTools...)
	copyRun.TestReports = append([]domain.TestReport(nil), run.TestReports...)
	copyRun.SelectedTests = append([]string(nil), run.SelectedTests...)
	copyRun.Logs = append([]domain.RunLog(nil), run.Logs...)
	copyRun.DevActivities = append([]domain.DevActivity(nil), run.DevActivities...)
	copyRun.Failures = append([]domain.FailureSummary(nil), run.Failures...)
	copyRun.Deployments = append([]domain.DeploymentRecord(nil), run.Deployments...)
	copyRun.PolicyDecisions = append([]domain.PolicyDecision(nil), run.PolicyDecisions...)
	return copyRun
}

func cloneAnalysisPlan(plan domain.AnalysisPlan) domain.AnalysisPlan {
	copyPlan := plan
	copyPlan.TaskQueue = append([]domain.TaskQueueItem(nil), plan.TaskQueue...)
	copyPlan.Checkpoints = append([]domain.FlowCheckpoint(nil), plan.Checkpoints...)
	copyPlan.LoopBlueprint = append([]domain.LoopStep(nil), plan.LoopBlueprint...)
	copyPlan.RecommendedStages = append([]domain.StageKind(nil), plan.RecommendedStages...)
	copyPlan.RecommendedTools = append([]string(nil), plan.RecommendedTools...)
	copyPlan.RecommendedTests = append([]string(nil), plan.RecommendedTests...)
	copyPlan.Risks = append([]string(nil), plan.Risks...)
	copyPlan.RequirementSpec.FunctionalRequirements = append([]string(nil), plan.RequirementSpec.FunctionalRequirements...)
	copyPlan.RequirementSpec.TechStack = append([]string(nil), plan.RequirementSpec.TechStack...)
	copyPlan.RequirementSpec.NonFunctional = append([]string(nil), plan.RequirementSpec.NonFunctional...)
	copyPlan.RequirementSpec.Constraints = append([]string(nil), plan.RequirementSpec.Constraints...)
	copyPlan.RequirementSpec.StylePreferences = append([]string(nil), plan.RequirementSpec.StylePreferences...)
	return copyPlan
}

func buildAtomicToolInvocationSummary(item domain.AtomicToolProfile, policy domain.RuntimePolicy) string {
	stageNames := make([]string, 0, len(item.StageKinds))
	for _, stage := range item.StageKinds {
		stageNames = append(stageNames, string(stage))
	}
	priority := "本地优先"
	if !item.LocalFirst {
		priority = "候补优先"
	}
	return fmt.Sprintf("已按 %s 策略调用 %s，适用阶段 %s，当前负载档位为 %s，入口策略为 %s。", policy.Profile, item.Name, strings.Join(stageNames, " / "), item.LoadTier, priority)
}

func currentRunStages(stages []domain.Stage) []domain.StageKind {
	out := make([]domain.StageKind, 0, len(stages))
	seen := map[domain.StageKind]bool{}
	for _, stage := range stages {
		if seen[stage.Kind] {
			continue
		}
		seen[stage.Kind] = true
		out = append(out, stage.Kind)
	}
	return out
}

func activeRunTools(stages []domain.Stage) []domain.ToolConfig {
	if len(stages) == 0 {
		return nil
	}
	return append([]domain.ToolConfig(nil), stages[0].Tools...)
}

func (p *Platform) listFeatureTogglesLocked(profile domain.SystemProfile) []domain.FeatureToggle {
	out := make([]domain.FeatureToggle, 0, len(p.featureToggles))
	for _, toggle := range p.featureToggles {
		out = append(out, decorateFeatureToggle(toggle, profile))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (p *Platform) listBuiltinModelPacksLocked(profile domain.SystemProfile) []domain.BuiltinModelPack {
	out := make([]domain.BuiltinModelPack, 0, len(p.builtinModelPacks))
	for _, pack := range p.builtinModelPacks {
		out = append(out, decorateBuiltinModelPack(pack, profile))
	}
	sort.Slice(out, func(i, j int) bool {
		left := out[i]
		right := out[j]
		if left.ReviewScore != right.ReviewScore {
			return left.ReviewScore < right.ReviewScore
		}
		if left.FilterScore != right.FilterScore {
			return left.FilterScore < right.FilterScore
		}
		if left.AlignmentScore != right.AlignmentScore {
			return left.AlignmentScore < right.AlignmentScore
		}
		return left.Name < right.Name
	})
	return out
}
