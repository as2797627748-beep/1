package domain

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"
)

type RunStatus string

const (
	RunQueued    RunStatus = "queued"
	RunRunning   RunStatus = "running"
	RunPaused    RunStatus = "paused"
	RunFailed    RunStatus = "failed"
	RunCompleted RunStatus = "completed"
)

type StageKind string

const (
	StageIntent    StageKind = "intent"
	StageContext   StageKind = "context"
	StagePlan      StageKind = "plan"
	StageResource  StageKind = "resource"
	StageModel     StageKind = "model"
	StageTool      StageKind = "tool"
	StageImplement StageKind = "implement"
	StageResult    StageKind = "result"
	StageTest      StageKind = "test"
	StageDeploy    StageKind = "deploy"
	StageRepair    StageKind = "repair"
	StageFinalize  StageKind = "finalize"

	// Legacy aliases kept for existing callers.
	StageAnalyze StageKind = StageIntent
	StageModify  StageKind = StageResult
)

type StageStatus string

const (
	StagePending   StageStatus = "pending"
	StageRunning   StageStatus = "running"
	StagePaused    StageStatus = "paused"
	StageFailed    StageStatus = "failed"
	StageCompleted StageStatus = "completed"
	StageSkipped   StageStatus = "skipped"
)

type EventType string

const (
	EventRunCreated         EventType = "run_created"
	EventRunUpdated         EventType = "run_updated"
	EventStageUpdated       EventType = "stage_updated"
	EventRequirementPatched EventType = "requirement_patched"
)

type ToolConfig struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type Stage struct {
	ID        string       `json:"id"`
	Kind      StageKind    `json:"kind"`
	Status    StageStatus  `json:"status"`
	Summary   string       `json:"summary"`
	Tools     []ToolConfig `json:"tools"`
	UpdatedAt time.Time    `json:"updatedAt"`
}

type TestReport struct {
	Name      string        `json:"name"`
	Layer     string        `json:"layer"`
	Status    string        `json:"status"`
	Duration  time.Duration `json:"duration"`
	Summary   string        `json:"summary"`
	Completed time.Time     `json:"completed"`
}

type RunLog struct {
	Stage   StageKind `json:"stage"`
	Level   string    `json:"level"`
	Message string    `json:"message"`
	At      time.Time `json:"at"`
}

type DevActivity struct {
	Kind    string    `json:"kind"`
	Target  string    `json:"target"`
	Detail  string    `json:"detail"`
	Status  string    `json:"status"`
	Command string    `json:"command,omitempty"`
	At      time.Time `json:"at"`
}

type FailureSummary struct {
	Stage      StageKind `json:"stage"`
	Category   string    `json:"category"`
	Target     string    `json:"target,omitempty"`
	Reason     string    `json:"reason"`
	Suggestion string    `json:"suggestion"`
	Resolved   bool      `json:"resolved"`
	At         time.Time `json:"at"`
}

type DeploymentRecord struct {
	Mode      string    `json:"mode"`
	Status    string    `json:"status"`
	Target    string    `json:"target"`
	Version   string    `json:"version"`
	Summary   string    `json:"summary"`
	CreatedAt time.Time `json:"createdAt"`
}

type AnalysisPlan struct {
	Intent            string           `json:"intent"`
	ProjectKind       string           `json:"projectKind"`
	Summary           string           `json:"summary"`
	RequirementSpec   RequirementSpec  `json:"requirementSpec"`
	TaskQueue         []TaskQueueItem  `json:"taskQueue"`
	Checkpoints       []FlowCheckpoint `json:"checkpoints"`
	LoopBlueprint     []LoopStep       `json:"loopBlueprint"`
	RecommendedStages []StageKind      `json:"recommendedStages"`
	RecommendedTools  []string         `json:"recommendedTools"`
	RecommendedTests  []string         `json:"recommendedTests"`
	RecommendedBundle BundleSuggestion `json:"recommendedBundle"`
	Risks             []string         `json:"risks"`
}

type RequirementSpec struct {
	Summary                string   `json:"summary"`
	FunctionalRequirements []string `json:"functionalRequirements"`
	TechStack              []string `json:"techStack"`
	NonFunctional          []string `json:"nonFunctional"`
	Constraints            []string `json:"constraints"`
	StylePreferences       []string `json:"stylePreferences"`
}

type TaskQueueItem struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Summary  string `json:"summary"`
	Agent    string `json:"agent"`
	Phase    string `json:"phase"`
	Priority int    `json:"priority"`
	Status   string `json:"status"`
}

type FlowCheckpoint struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
	Gate    string `json:"gate"`
	Status  string `json:"status"`
}

type BundleSuggestion struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

type LoopStep struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Summary string `json:"summary"`
}

type ToolProfile struct {
	ID          string `json:"id"`
	Category    string `json:"category"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Heavy       bool   `json:"heavy"`
}

type AtomicToolProfile struct {
	ID                string      `json:"id"`
	Category          string      `json:"category"`
	Name              string      `json:"name"`
	Summary           string      `json:"summary"`
	StageKinds        []StageKind `json:"stageKinds"`
	LoadTier          string      `json:"loadTier"`
	Priority          int         `json:"priority"`
	LocalFirst        bool        `json:"localFirst"`
	CoverageTags      []string    `json:"coverageTags"`
	PreferredProvider string      `json:"preferredProvider"`
	FallbackProviders []string    `json:"fallbackProviders"`
	ActivationMode    string      `json:"activationMode"`
	DedupGroup        string      `json:"dedupGroup"`
	Recommended       bool        `json:"recommended"`
	Allowed           bool        `json:"allowed"`
	Reason            string      `json:"reason"`
}

type ToolInvocation struct {
	ID        string    `json:"id"`
	ToolID    string    `json:"toolId"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	Summary   string    `json:"summary"`
	CreatedAt time.Time `json:"createdAt"`
}

type TestProfile struct {
	ID          string `json:"id"`
	Layer       string `json:"layer"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type WorkflowOption struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Name        string `json:"name"`
	Description string `json:"description"`
	DefaultOn   bool   `json:"defaultOn"`
}

type WorkflowOptionGroup struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Options     []WorkflowOption `json:"options"`
}

type FeatureToggle struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	DefaultOn   bool   `json:"defaultOn"`
	Recommended bool   `json:"recommended"`
	Allowed     bool   `json:"allowed"`
	Warning     string `json:"warning"`
}

type BuiltinModelPack struct {
	ID                string         `json:"id"`
	Provider          string         `json:"provider"`
	ModelName         string         `json:"modelName"`
	Version           string         `json:"version"`
	Name              string         `json:"name"`
	Description       string         `json:"description"`
	InstallState      string         `json:"installState"`
	StatusDetail      string         `json:"statusDetail"`
	Enabled           bool           `json:"enabled"`
	Downloaded        bool           `json:"downloaded"`
	DefaultOn         bool           `json:"defaultOn"`
	SizeTier          string         `json:"sizeTier"`
	Variant           string         `json:"variant"`
	SizeHint          string         `json:"sizeHint"`
	RuntimeHint       string         `json:"runtimeHint"`
	SystemRequirement string         `json:"systemRequirement"`
	Recommended       bool           `json:"recommended"`
	Allowed           bool           `json:"allowed"`
	Warning           string         `json:"warning"`
	ReviewScore       int            `json:"reviewScore"`
	FilterScore       int            `json:"filterScore"`
	AlignmentScore    int            `json:"alignmentScore"`
	PolicyHint        string         `json:"policyHint"`
	PolicyMode        string         `json:"policyMode"`
	PolicyDecision    PolicyDecision `json:"policyDecision"`
}

type SystemProfile struct {
	OS                     string   `json:"os"`
	Arch                   string   `json:"arch"`
	CPUCores               int      `json:"cpuCores"`
	MemoryMB               int      `json:"memoryMB"`
	Tier                   string   `json:"tier"`
	RecommendedConcurrency int      `json:"recommendedConcurrency"`
	LocalModelAllowed      bool     `json:"localModelAllowed"`
	RecommendedVPS         string   `json:"recommendedVps"`
	RecommendedModes       []string `json:"recommendedModes"`
	RecommendedFeatures    []string `json:"recommendedFeatures"`
	UpgradeSuggestions     []string `json:"upgradeSuggestions"`
	NoSSHReadiness         string   `json:"noSshReadiness"`
	StrategySummary        string   `json:"strategySummary"`
}

type AuditCheck struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Summary    string `json:"summary"`
	Suggestion string `json:"suggestion"`
}

type AuditReport struct {
	Score   int          `json:"score"`
	Summary string       `json:"summary"`
	Checks  []AuditCheck `json:"checks"`
}

type AdviceRequest struct {
	Mode   string `json:"mode"`
	Goal   string `json:"goal"`
	Target string `json:"target"`
}

type AdviceReport struct {
	Mode         string   `json:"mode"`
	Summary      string   `json:"summary"`
	PromptBundle string   `json:"promptBundle"`
	HostingMode  string   `json:"hostingMode"`
	Checklist    []string `json:"checklist"`
	ConsoleSteps []string `json:"consoleSteps"`
	Suggestions  []string `json:"suggestions"`
}

type PolicyDecision struct {
	Area    string `json:"area"`
	Action  string `json:"action"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

type Run struct {
	ID                  string              `json:"id"`
	Title               string              `json:"title"`
	Goal                string              `json:"goal"`
	TemplateID          string              `json:"templateId"`
	TemplateSource      string              `json:"templateSource"`
	Status              RunStatus           `json:"status"`
	DedupKey            string              `json:"dedupKey"`
	LoopBlueprint       []LoopStep          `json:"loopBlueprint"`
	AssembledTools      []AtomicToolProfile `json:"assembledTools"`
	Stages              []Stage             `json:"stages"`
	TestReports         []TestReport        `json:"testReports"`
	Logs                []RunLog            `json:"logs"`
	DevActivities       []DevActivity       `json:"devActivities"`
	Failures            []FailureSummary    `json:"failures"`
	Deployments         []DeploymentRecord  `json:"deployments"`
	PolicyDecisions     []PolicyDecision    `json:"policyDecisions"`
	Analysis            AnalysisPlan        `json:"analysis"`
	TestMode            string              `json:"testMode"`
	SelectedTests       []string            `json:"selectedTests"`
	AutoRepairEnabled   bool                `json:"autoRepairEnabled"`
	AutoRepairMode      string              `json:"autoRepairMode"`
	RemoteDeployEnabled bool                `json:"remoteDeployEnabled"`
	RepairAttempts      int                 `json:"repairAttempts"`
	CreatedAt           time.Time           `json:"createdAt"`
	UpdatedAt           time.Time           `json:"updatedAt"`
}

type ModelProfile struct {
	ID                 string   `json:"id"`
	Provider           string   `json:"provider"`
	Name               string   `json:"name"`
	Version            string   `json:"version"`
	FullName           string   `json:"fullName"`
	Website            string   `json:"website"`
	Region             string   `json:"region"`
	Tags               []string `json:"tags"`
	ReviewScore        int      `json:"reviewScore"`
	FilterScore        int      `json:"filterScore"`
	AlignmentScore     int      `json:"alignmentScore"`
	Configured         bool     `json:"configured"`
	CredentialEnv      string   `json:"credentialEnv"`
	RecommendedBaseURL string   `json:"recommendedBaseUrl"`
}

type Template struct {
	ID                string           `json:"id"`
	Kind              string           `json:"kind"`
	Name              string           `json:"name"`
	Description       string           `json:"description"`
	RecommendedBundle BundleSuggestion `json:"recommendedBundle"`
	DefaultStages     []StageKind      `json:"defaultStages"`
	DefaultTools      []string         `json:"defaultTools"`
	Defaults          interface{}      `json:"defaults"`
}

type ProviderStatus struct {
	Provider      string `json:"provider"`
	Configured    bool   `json:"configured"`
	CredentialEnv string `json:"credentialEnv"`
	Website       string `json:"website"`
	Summary       string `json:"summary"`
}

type CapabilityProfile struct {
	ID          string `json:"id"`
	Category    string `json:"category"`
	Name        string `json:"name"`
	Mode        string `json:"mode"`
	Allowed     bool   `json:"allowed"`
	Recommended bool   `json:"recommended"`
	Summary     string `json:"summary"`
	Reason      string `json:"reason"`
}

type OptimizationProfile struct {
	ID       string `json:"id"`
	Priority string `json:"priority"`
	Title    string `json:"title"`
	Action   string `json:"action"`
	Reason   string `json:"reason"`
}

type RuntimePolicy struct {
	Profile              string   `json:"profile"`
	MaxConcurrentRuns    int      `json:"maxConcurrentRuns"`
	MaxHeavyActions      int      `json:"maxHeavyActions"`
	AllowBackgroundJobs  bool     `json:"allowBackgroundJobs"`
	AllowLocalModels     bool     `json:"allowLocalModels"`
	CacheBudgetMB        int      `json:"cacheBudgetMB"`
	ValidationDepth      string   `json:"validationDepth"`
	DefaultEnabledTools  []string `json:"defaultEnabledTools"`
	DefaultDisabledTools []string `json:"defaultDisabledTools"`
	Summary              string   `json:"summary"`
}

type SystemSummary struct {
	Providers            []ProviderStatus      `json:"providers"`
	ToolCatalog          []ToolProfile         `json:"toolCatalog"`
	AtomicTools          []AtomicToolProfile   `json:"atomicTools"`
	ToolInvocations      []ToolInvocation      `json:"toolInvocations"`
	TestCatalog          []TestProfile         `json:"testCatalog"`
	WorkflowOptions      []WorkflowOptionGroup `json:"workflowOptions"`
	FeatureToggles       []FeatureToggle       `json:"featureToggles"`
	BuiltinModelPacks    []BuiltinModelPack    `json:"builtinModelPacks"`
	CapabilityProfiles   []CapabilityProfile   `json:"capabilityProfiles"`
	OptimizationProfiles []OptimizationProfile `json:"optimizationProfiles"`
	RuntimePolicy        RuntimePolicy         `json:"runtimePolicy"`
	SystemProfile        SystemProfile         `json:"systemProfile"`
	RunTotals            map[string]int        `json:"runTotals"`
	SchedulerMode        string                `json:"schedulerMode"`
	DeployMode           string                `json:"deployMode"`
	InterfaceModes       []string              `json:"interfaceModes"`
}

type Event struct {
	Type    EventType `json:"type"`
	RunID   string    `json:"runId,omitempty"`
	Message string    `json:"message"`
	At      time.Time `json:"at"`
}

type CreateRunInput struct {
	Title               string       `json:"title"`
	Goal                string       `json:"goal"`
	TemplateID          string       `json:"templateId"`
	TemplateSource      string       `json:"templateSource"`
	TestMode            string       `json:"testMode"`
	SelectedTests       []string     `json:"selectedTests"`
	Stages              []StageKind  `json:"stages"`
	Tools               []ToolConfig `json:"tools"`
	AutoRepairEnabled   bool         `json:"autoRepairEnabled"`
	AutoRepairMode      string       `json:"autoRepairMode"`
	RemoteDeployEnabled bool         `json:"remoteDeployEnabled"`
}

type DevActivityInput struct {
	Kind    string `json:"kind"`
	Target  string `json:"target"`
	Detail  string `json:"detail"`
	Status  string `json:"status"`
	Command string `json:"command"`
}

func NormalizeTools(tools []ToolConfig) []ToolConfig {
	seen := map[string]ToolConfig{}
	for _, tool := range tools {
		name := strings.TrimSpace(strings.ToLower(tool.Name))
		if name == "" {
			continue
		}
		tool.Name = name
		seen[name] = tool
	}
	out := make([]ToolConfig, 0, len(seen))
	for _, tool := range seen {
		out = append(out, tool)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}

func dedupeStrings(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(strings.ToLower(value))
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		out = append(out, trimmed)
	}
	sort.Strings(out)
	return out
}

func DedupKey(input CreateRunInput) string {
	stages := make([]string, 0, len(input.Stages))
	for _, stage := range input.Stages {
		stages = append(stages, string(stage))
	}
	sort.Strings(stages)
	tools := NormalizeTools(input.Tools)
	toolNames := make([]string, 0, len(tools))
	for _, tool := range tools {
		flag := "off"
		if tool.Enabled {
			flag = "on"
		}
		toolNames = append(toolNames, tool.Name+":"+flag)
	}
	payload := strings.ToLower(strings.TrimSpace(input.Title)) + "|" +
		strings.ToLower(strings.TrimSpace(input.Goal)) + "|" +
		strings.ToLower(strings.TrimSpace(input.TemplateID)) + "|" +
		strings.Join(stages, ",") + "|" + strings.Join(toolNames, ",") + "|" + strings.ToLower(strings.TrimSpace(input.AutoRepairMode)) + "|" + strings.ToLower(strings.TrimSpace(input.TestMode)) + "|" + strings.Join(dedupeStrings(input.SelectedTests), ",")
	hash := sha1.Sum([]byte(payload))
	return hex.EncodeToString(hash[:])
}

func SortModels(models []ModelProfile) []ModelProfile {
	out := append([]ModelProfile(nil), models...)
	sort.SliceStable(out, func(i, j int) bool {
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
		if left.Provider != right.Provider {
			return left.Provider < right.Provider
		}
		return left.Name < right.Name
	})
	return out
}

func VersionTag(now time.Time) string {
	return fmt.Sprintf("%04d%02d%02d-%02d%02d%02d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
}
