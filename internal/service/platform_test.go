package service

import (
	"errors"
	"os"
	"strings"
	"testing"
	"testing/quick"

	"autocode-platform/internal/domain"
)

func setRunFailuresForTest(p *Platform, runID string, failures []domain.FailureSummary) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.runs[runID].Failures = failures
}

func createRunWithDeploymentVersionForTest(t *testing.T, p *Platform, input domain.CreateRunInput) (domain.Run, string) {
	t.Helper()
	run, err := p.CreateRun(input)
	if err != nil {
		t.Fatal(err)
	}
	p.tick()
	p.tick()
	p.tick()
	updated, err := p.GetRun(run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Deployments) == 0 {
		t.Fatal("expected deployment record")
	}
	return run, updated.Deployments[0].Version
}

func TestCreateRunDeduplicatesQueuedRuns(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	input := domain.CreateRunInput{
		Title:      "build platform",
		Goal:       "do the job",
		TemplateID: "tpl-full-loop",
		Stages:     []domain.StageKind{domain.StageAnalyze, domain.StageImplement, domain.StageTest},
		Tools: []domain.ToolConfig{
			{Name: "workspace", Enabled: true},
			{Name: "workspace", Enabled: true},
		},
	}
	if _, err := p.CreateRun(input); err != nil {
		t.Fatalf("first create failed: %v", err)
	}
	if _, err := p.CreateRun(input); err == nil {
		t.Fatal("expected dedup conflict")
	}
}

func TestPatchRequirementsResetsAffectedStages(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, err := p.CreateRun(domain.CreateRunInput{
		Title:      "demo",
		Goal:       "initial",
		TemplateID: "tpl-full-loop",
	})
	if err != nil {
		t.Fatalf("create run failed: %v", err)
	}
	for i := 0; i < 3; i++ {
		p.tick()
	}
	patched, err := p.PatchRequirements(run.ID, "补充一个新需求")
	if err != nil {
		t.Fatalf("patch failed: %v", err)
	}
	if patched.Status != domain.RunPaused {
		t.Fatalf("expected paused, got %s", patched.Status)
	}
	foundContextCompleted := false
	for _, stage := range patched.Stages {
		if stage.Kind == domain.StageContext {
			if stage.Status != domain.StageCompleted {
				t.Fatalf("expected context stage preserved, got %s", stage.Status)
			}
			foundContextCompleted = true
			break
		}
	}
	if !foundContextCompleted {
		t.Fatal("expected context stage to exist")
	}
	for _, stage := range patched.Stages {
		switch stage.Kind {
		case domain.StagePlan, domain.StageResource, domain.StageModel, domain.StageTool, domain.StageImplement, domain.StageResult, domain.StageTest, domain.StageDeploy, domain.StageRepair, domain.StageFinalize:
			if stage.Status != domain.StagePending {
				t.Fatalf("expected %s stage reset to pending, got %s", stage.Kind, stage.Status)
			}
		}
	}
}

func TestCreateRunAppliesTemplateDefaultsWhenInputEmpty(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, err := p.CreateRun(domain.CreateRunInput{
		Title:      "模板兜底",
		Goal:       "按模板直接创建运行单",
		TemplateID: "tpl-safe-deploy",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(run.Stages) == 0 {
		t.Fatal("expected template stages to be applied")
	}
	if run.Stages[0].Kind != domain.StageIntent {
		t.Fatalf("expected intent as first stage, got %s", run.Stages[0].Kind)
	}
	if run.Stages[len(run.Stages)-1].Kind != domain.StageFinalize {
		t.Fatalf("expected finalize as last stage, got %s", run.Stages[len(run.Stages)-1].Kind)
	}
	foundPlan := false
	foundDeploy := false
	for _, stage := range run.Stages {
		if stage.Kind == domain.StagePlan {
			foundPlan = true
		}
		if stage.Kind == domain.StageDeploy {
			foundDeploy = true
		}
	}
	if !foundPlan {
		t.Fatal("expected template stages to include plan")
	}
	if !foundDeploy {
		t.Fatal("expected template stages to include deploy")
	}
	if len(run.PolicyDecisions) == 0 || run.PolicyDecisions[0].Area != "template" {
		t.Fatal("expected template decision to be recorded")
	}
	foundWorkspace := false
	for _, stage := range run.Stages {
		for _, tool := range stage.Tools {
			if tool.Name == "workspace" && tool.Enabled {
				foundWorkspace = true
			}
		}
	}
	if !foundWorkspace {
		t.Fatal("expected template tools to be applied")
	}
}

func TestCreateRunPersistsTemplateSource(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, err := p.CreateRun(domain.CreateRunInput{
		Title:          "模板来源",
		Goal:           "记录模板来源",
		TemplateID:     "tpl-full-loop",
		TemplateSource: "recent-template",
	})
	if err != nil {
		t.Fatal(err)
	}
	if run.TemplateSource != "recent-template" {
		t.Fatalf("expected template source recorded, got %q", run.TemplateSource)
	}
	foundSourceDecision := false
	for _, item := range run.PolicyDecisions {
		if item.Area == "template" && item.Action == "source" && item.Reason == "recent-template" {
			foundSourceDecision = true
			break
		}
	}
	if !foundSourceDecision {
		t.Fatal("expected template source policy decision")
	}
}

func TestCreateRunPersistsTestModeAndSelectedTests(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, err := p.CreateRun(domain.CreateRunInput{
		Title:         "测试模式",
		Goal:          "执行登录与权限验证",
		TestMode:      "smoke",
		SelectedTests: []string{"static-check", "security-scan", "e2e-smoke"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if run.TestMode != "smoke" {
		t.Fatalf("expected smoke test mode, got %q", run.TestMode)
	}
	if len(run.SelectedTests) != 3 {
		t.Fatalf("expected selected tests persisted, got %d", len(run.SelectedTests))
	}
}

func TestRecordDevActivityPersistsWorkbenchTrace(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, err := p.CreateRun(domain.CreateRunInput{
		Title: "开发工作台链路",
		Goal:  "完成文件修改并执行测试",
	})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := p.RecordDevActivity(run.ID, domain.DevActivityInput{
		Kind:    "command",
		Command: "go test ./...",
		Status:  "completed",
		Detail:  "执行测试",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.DevActivities) == 0 {
		t.Fatal("expected dev activity recorded")
	}
	if updated.DevActivities[0].Command != "go test ./..." {
		t.Fatalf("expected command recorded, got %q", updated.DevActivities[0].Command)
	}
	foundValidationCompleted := false
	for _, item := range updated.Analysis.Checkpoints {
		if item.ID == "checkpoint-validation" && item.Status == "completed" {
			foundValidationCompleted = true
			break
		}
	}
	if !foundValidationCompleted {
		t.Fatal("expected validation checkpoint completed")
	}
}

func TestRollbackDeploymentAppendsRollbackRecord(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, err := p.CreateRun(domain.CreateRunInput{
		Title:               "版本回退",
		Goal:                "执行稳定部署并支持回退",
		Stages:              []domain.StageKind{domain.StageAnalyze, domain.StageDeploy},
		RemoteDeployEnabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	p.tick()
	p.tick()
	p.tick()
	updated, err := p.GetRun(run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Deployments) == 0 {
		t.Fatal("expected deployment record")
	}
	version := updated.Deployments[0].Version
	rolled, err := p.RollbackDeployment(run.ID, version)
	if err != nil {
		t.Fatal(err)
	}
	if len(rolled.Deployments) < 2 {
		t.Fatal("expected rollback record appended")
	}
	latest := rolled.Deployments[len(rolled.Deployments)-1]
	if latest.Mode != "rollback" {
		t.Fatalf("expected rollback mode, got %q", latest.Mode)
	}
	if latest.Version != version {
		t.Fatalf("expected rollback version %q, got %q", version, latest.Version)
	}
	if len(rolled.Logs) == 0 {
		t.Fatal("expected rollback log recorded")
	}
	if !strings.Contains(rolled.Logs[len(rolled.Logs)-1].Message, version) {
		t.Fatalf("expected rollback log to mention version %q", version)
	}
}

func TestRevalidateRunRefreshesReports(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, err := p.CreateRun(domain.CreateRunInput{
		Title:         "回退后复验",
		Goal:          "回退后补做关键验证",
		SelectedTests: []string{"review-check", "security-scan"},
	})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := p.RevalidateRun(run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.TestReports) != 2 {
		t.Fatalf("expected 2 test reports, got %d", len(updated.TestReports))
	}
	if len(updated.Deployments) != 1 {
		t.Fatalf("expected 1 deployment timeline record, got %d", len(updated.Deployments))
	}
	if updated.Deployments[0].Mode != "revalidate" {
		t.Fatalf("expected revalidate deployment mode, got %q", updated.Deployments[0].Mode)
	}
	foundValidation := false
	foundSecurityReport := false
	for _, item := range updated.Analysis.Checkpoints {
		if item.ID == "checkpoint-validation" && item.Status == "completed" {
			foundValidation = true
		}
	}
	for _, report := range updated.TestReports {
		if report.Name == "security-scan" && report.Status == "passed" {
			foundSecurityReport = true
		}
	}
	if !foundValidation {
		t.Fatal("expected validation checkpoint completed")
	}
	if !foundSecurityReport {
		t.Fatal("expected security report generated")
	}
}

func TestRevalidateRunResolvesValidationFailures(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, err := p.CreateRun(domain.CreateRunInput{
		Title:         "复验消化失败摘要",
		Goal:          "回退后补做关键验证",
		SelectedTests: []string{"review-check", "security-scan"},
	})
	if err != nil {
		t.Fatal(err)
	}
	setRunFailuresForTest(p, run.ID, []domain.FailureSummary{
		{Stage: domain.StageTest, Category: "validation", Target: "security-scan", Reason: "安全扫描未通过", Suggestion: "请先修正", Resolved: false},
		{Stage: domain.StageDeploy, Category: "delivery", Target: "deploy-smoke", Reason: "部署后检查未通过", Suggestion: "请复验", Resolved: false},
	})
	updated, err := p.RevalidateRun(run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Failures) != 2 {
		t.Fatalf("expected 2 failures retained, got %d", len(updated.Failures))
	}
	if !updated.Failures[0].Resolved || !updated.Failures[1].Resolved {
		t.Fatal("expected revalidate to resolve validation and deploy-smoke failures")
	}
}

func TestRevalidateRunDoesNotResolvePreviewFailures(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, err := p.CreateRun(domain.CreateRunInput{
		Title:         "复验仅处理匹配失败",
		Goal:          "区分复验和预览确认",
		SelectedTests: []string{"review-check", "security-scan"},
	})
	if err != nil {
		t.Fatal(err)
	}
	setRunFailuresForTest(p, run.ID, []domain.FailureSummary{
		{Stage: domain.StageTest, Category: "validation", Target: "security-scan", Reason: "安全扫描未通过", Suggestion: "请先修正", Resolved: false},
		{Stage: domain.StageDeploy, Category: "delivery", Target: "deploy-smoke", Reason: "部署后检查未通过", Suggestion: "请复验", Resolved: false},
		{Stage: domain.StageDeploy, Category: "delivery", Target: "preview-check", Reason: "预览确认未完成", Suggestion: "请打开预览", Resolved: false},
	})
	updated, err := p.RevalidateRun(run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Failures) != 3 {
		t.Fatalf("expected 3 failures retained, got %d", len(updated.Failures))
	}
	if !updated.Failures[0].Resolved || !updated.Failures[1].Resolved {
		t.Fatal("expected revalidate to resolve validation and deploy-smoke failures")
	}
	if updated.Failures[2].Resolved {
		t.Fatal("expected preview failure to remain unresolved after revalidate")
	}
}

func TestDeployRunAppendsDeploymentRecord(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, err := p.CreateRun(domain.CreateRunInput{
		Title:               "手动部署",
		Goal:                "执行稳定部署",
		Stages:              []domain.StageKind{domain.StageAnalyze, domain.StageTest, domain.StageDeploy},
		RemoteDeployEnabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := p.DeployRun(run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Deployments) != 1 {
		t.Fatalf("expected 1 deployment record, got %d", len(updated.Deployments))
	}
	if updated.Deployments[0].Mode != "remote" {
		t.Fatalf("expected remote deployment mode, got %q", updated.Deployments[0].Mode)
	}
	foundDeployStage := false
	for _, stage := range updated.Stages {
		if stage.Kind == domain.StageDeploy && stage.Status == domain.StageCompleted {
			foundDeployStage = true
			break
		}
	}
	if !foundDeployStage {
		t.Fatal("expected deploy stage completed")
	}
}

func TestDeployRunDoesNotResolvePreviewFailures(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, err := p.CreateRun(domain.CreateRunInput{
		Title:               "部署不处理预览失败",
		Goal:                "执行稳定部署但保留预览问题",
		Stages:              []domain.StageKind{domain.StageDeploy},
		RemoteDeployEnabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	setRunFailuresForTest(p, run.ID, []domain.FailureSummary{
		{Stage: domain.StageDeploy, Category: "delivery", Target: "preview-check", Reason: "预览确认未完成", Suggestion: "请打开预览", Resolved: true},
		{Stage: domain.StageTest, Category: "validation", Target: "security-scan", Reason: "安全扫描未通过", Suggestion: "请先修正", Resolved: true},
	})
	updated, err := p.DeployRun(run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Failures) != 2 {
		t.Fatalf("expected 2 failures retained, got %d", len(updated.Failures))
	}
	if !updated.Failures[0].Resolved || !updated.Failures[1].Resolved {
		t.Fatal("expected deploy run not to reopen unrelated resolved failures")
	}
	if updated.Failures[0].Target != "preview-check" {
		t.Fatalf("expected preview failure retained first, got %q", updated.Failures[0].Target)
	}
}

func TestRollbackDeploymentResolvesDeliveryFailures(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, version := createRunWithDeploymentVersionForTest(t, p, domain.CreateRunInput{
		Title:               "回退消化部署问题",
		Goal:                "执行稳定部署并支持回退",
		Stages:              []domain.StageKind{domain.StageAnalyze, domain.StageDeploy},
		RemoteDeployEnabled: true,
	})
	setRunFailuresForTest(p, run.ID, []domain.FailureSummary{
		{Stage: domain.StageDeploy, Category: "delivery", Target: "stable-deploy", Reason: "稳定部署执行异常", Suggestion: "请先回退", Resolved: false},
		{Stage: domain.StageDeploy, Category: "delivery", Target: "deploy-smoke", Reason: "部署后检查未通过", Suggestion: "请先回退", Resolved: false},
		{Stage: domain.StageDeploy, Category: "delivery", Target: "preview-check", Reason: "预览确认未完成", Suggestion: "请打开预览", Resolved: false},
	})
	rolled, err := p.RollbackDeployment(run.ID, version)
	if err != nil {
		t.Fatal(err)
	}
	if !rolled.Failures[0].Resolved || !rolled.Failures[1].Resolved {
		t.Fatal("expected rollback to resolve deploy execution and smoke failures")
	}
	if rolled.Failures[2].Resolved {
		t.Fatal("expected preview confirmation failure to remain unresolved")
	}
}

func TestRollbackDeploymentDoesNotResolveValidationFailures(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, version := createRunWithDeploymentVersionForTest(t, p, domain.CreateRunInput{
		Title:               "回退不处理验证失败",
		Goal:                "执行稳定部署并保留验证问题",
		Stages:              []domain.StageKind{domain.StageAnalyze, domain.StageDeploy},
		RemoteDeployEnabled: true,
	})
	setRunFailuresForTest(p, run.ID, []domain.FailureSummary{
		{Stage: domain.StageTest, Category: "validation", Target: "security-scan", Reason: "安全扫描未通过", Suggestion: "请先修正", Resolved: false},
		{Stage: domain.StageDeploy, Category: "delivery", Target: "stable-deploy", Reason: "稳定部署执行异常", Suggestion: "请先回退", Resolved: false},
		{Stage: domain.StageDeploy, Category: "delivery", Target: "deploy-smoke", Reason: "部署后检查未通过", Suggestion: "请先回退", Resolved: false},
	})
	rolled, err := p.RollbackDeployment(run.ID, version)
	if err != nil {
		t.Fatal(err)
	}
	if rolled.Failures[0].Resolved {
		t.Fatal("expected validation failure to remain unresolved after rollback")
	}
	if !rolled.Failures[1].Resolved || !rolled.Failures[2].Resolved {
		t.Fatal("expected rollback to resolve only delivery failures")
	}
}

func TestPreviewActivityResolvesPreviewFailures(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, err := p.CreateRun(domain.CreateRunInput{Title: "预览确认", Goal: "打开预览确认服务状态"})
	if err != nil {
		t.Fatal(err)
	}
	setRunFailuresForTest(p, run.ID, []domain.FailureSummary{
		{Stage: domain.StageDeploy, Category: "delivery", Target: "preview-check", Reason: "预览确认未完成", Suggestion: "请打开预览", Resolved: false},
	})
	updated, err := p.RecordDevActivity(run.ID, domain.DevActivityInput{Kind: "preview-open", Target: "8080", Status: "completed", Detail: "打开预览"})
	if err != nil {
		t.Fatal(err)
	}
	if !updated.Failures[0].Resolved {
		t.Fatal("expected preview-open activity to resolve preview failure")
	}
}

func TestPreviewActivityDoesNotResolveValidationFailures(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, err := p.CreateRun(domain.CreateRunInput{Title: "预览不处理验证失败", Goal: "打开预览但保留验证问题"})
	if err != nil {
		t.Fatal(err)
	}
	setRunFailuresForTest(p, run.ID, []domain.FailureSummary{
		{Stage: domain.StageTest, Category: "validation", Target: "security-scan", Reason: "安全扫描未通过", Suggestion: "请先修正", Resolved: false},
		{Stage: domain.StageDeploy, Category: "delivery", Target: "preview-check", Reason: "预览确认未完成", Suggestion: "请打开预览", Resolved: false},
	})
	updated, err := p.RecordDevActivity(run.ID, domain.DevActivityInput{Kind: "preview-open", Target: "8080", Status: "completed", Detail: "打开预览"})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Failures[0].Resolved {
		t.Fatal("expected validation failure to remain unresolved after preview-open activity")
	}
	if !updated.Failures[1].Resolved {
		t.Fatal("expected preview failure resolved after preview-open activity")
	}
}

func TestSortModelsPrefersLowerScores(t *testing.T) {
	models := []domain.ModelProfile{
		{Name: "A", Provider: "P", ReviewScore: 10, FilterScore: 10, AlignmentScore: 10},
		{Name: "B", Provider: "P", ReviewScore: 8, FilterScore: 20, AlignmentScore: 20},
	}
	sorted := domain.SortModels(models)
	if sorted[0].Name != "B" {
		t.Fatalf("expected B first, got %s", sorted[0].Name)
	}
}

func TestDedupKeyStableUnderToolOrdering(t *testing.T) {
	property := func(title, goal string) bool {
		left := domain.CreateRunInput{
			Title:  title,
			Goal:   goal,
			Stages: []domain.StageKind{domain.StageAnalyze, domain.StageTest},
			Tools:  []domain.ToolConfig{{Name: "tests", Enabled: true}, {Name: "workspace", Enabled: true}},
		}
		right := domain.CreateRunInput{
			Title:  title,
			Goal:   goal,
			Stages: []domain.StageKind{domain.StageTest, domain.StageAnalyze},
			Tools:  []domain.ToolConfig{{Name: "workspace", Enabled: true}, {Name: "tests", Enabled: true}, {Name: "workspace", Enabled: true}},
		}
		return domain.DedupKey(left) == domain.DedupKey(right)
	}
	if err := quick.Check(property, nil); err != nil {
		t.Fatal(err)
	}
}

func TestSummaryReflectsConfiguredProviderWithoutExposingSecrets(t *testing.T) {
	if err := os.Setenv("DEEPSEEK_API_KEY", "present"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("DEEPSEEK_API_KEY") })
	p := NewPlatform()
	t.Cleanup(p.Close)
	summary := p.Summary()
	found := false
	for _, provider := range summary.Providers {
		if provider.Provider == "DeepSeek" {
			found = true
			if !provider.Configured {
				t.Fatal("expected DeepSeek configured")
			}
			if provider.CredentialEnv != "DEEPSEEK_API_KEY" {
				t.Fatalf("unexpected credential env %s", provider.CredentialEnv)
			}
		}
	}
	if !found {
		t.Fatal("expected DeepSeek provider summary")
	}
}

func TestAutoRepairAppendsFailureSummary(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, err := p.CreateRun(domain.CreateRunInput{
		Title:             "repair demo",
		Goal:              "trigger [fail:test]",
		Stages:            []domain.StageKind{domain.StageAnalyze, domain.StageTest},
		AutoRepairEnabled: true,
		AutoRepairMode:    "standard",
	})
	if err != nil {
		t.Fatal(err)
	}
	p.tick()
	p.tick()
	updated, err := p.GetRun(run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Failures) == 0 {
		t.Fatal("expected failure summary")
	}
	if updated.Failures[0].Category != "validation" {
		t.Fatalf("expected validation failure category, got %q", updated.Failures[0].Category)
	}
	if updated.Failures[0].Resolved {
		t.Fatal("expected new failure to stay unresolved")
	}
	if updated.RepairAttempts == 0 {
		t.Fatal("expected repair attempts")
	}
	foundPausedTest := false
	for _, stage := range updated.Stages {
		if stage.Kind == domain.StageTest {
			if stage.Status != domain.StagePaused {
				t.Fatalf("expected failed test stage paused for repair, got %s", stage.Status)
			}
			foundPausedTest = true
		}
	}
	if !foundPausedTest {
		t.Fatal("expected test stage to exist")
	}
}

func TestTestStageFailureBuildsFailedReports(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, err := p.CreateRun(domain.CreateRunInput{
		Title:         "验证失败聚合",
		Goal:          "trigger [fail:test:security-scan]",
		Stages:        []domain.StageKind{domain.StageTest},
		SelectedTests: []string{"review-check", "security-scan"},
	})
	if err != nil {
		t.Fatal(err)
	}
	p.tick()
	updated, err := p.GetRun(run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != domain.RunFailed {
		t.Fatalf("expected run failed, got %s", updated.Status)
	}
	foundFailedSecurity := false
	for _, report := range updated.TestReports {
		if report.Name == "security-scan" && report.Status == "failed" {
			foundFailedSecurity = true
		}
	}
	if !foundFailedSecurity {
		t.Fatal("expected failed security-scan report")
	}
	if len(updated.Failures) == 0 {
		t.Fatal("expected failure summaries generated from failed reports")
	}
	if updated.Failures[0].Category != "validation" {
		t.Fatalf("expected validation failure category, got %q", updated.Failures[0].Category)
	}
	if updated.Failures[0].Target != "security-scan" {
		t.Fatalf("expected security-scan failure target, got %q", updated.Failures[0].Target)
	}
	if !strings.Contains(updated.Failures[0].Reason, "security-scan") {
		t.Fatalf("expected failure reason to mention failed report, got %q", updated.Failures[0].Reason)
	}
}

func TestAutoRepairRunsBeforeRetryingFailedStage(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, err := p.CreateRun(domain.CreateRunInput{
		Title:             "修复优先执行",
		Goal:              "trigger [fail:test]",
		Stages:            []domain.StageKind{domain.StageTest},
		SelectedTests:     []string{"review-check"},
		AutoRepairEnabled: true,
		AutoRepairMode:    "standard",
	})
	if err != nil {
		t.Fatal(err)
	}
	p.tick()
	first, err := p.GetRun(run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if first.Stages[0].Status != domain.StagePaused {
		t.Fatalf("expected failed stage paused for repair, got %s", first.Stages[0].Status)
	}
	p.tick()
	second, err := p.GetRun(run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Failures) == 0 {
		t.Fatal("expected failure history kept after repair stage")
	}
	if !second.Failures[0].Resolved {
		t.Fatal("expected failures marked resolved after repair stage")
	}
	if second.Stages[0].Status != domain.StagePending {
		t.Fatalf("expected failed stage reset to pending after repair, got %s", second.Stages[0].Status)
	}
	foundCompletedRepair := false
	for _, stage := range second.Stages {
		if stage.Kind == domain.StageRepair && stage.Status == domain.StageCompleted {
			foundCompletedRepair = true
		}
	}
	if !foundCompletedRepair {
		t.Fatal("expected repair stage completed before retry")
	}
	p.tick()
	finalRun, err := p.GetRun(run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if finalRun.Stages[0].Status != domain.StageCompleted {
		t.Fatalf("expected retried stage completed, got %s", finalRun.Stages[0].Status)
	}
}

func TestDeployRunOnlyBlocksUnresolvedFailures(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, err := p.CreateRun(domain.CreateRunInput{
		Title:               "部署失败阻断",
		Goal:                "允许部署校验",
		Stages:              []domain.StageKind{domain.StageDeploy},
		RemoteDeployEnabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	setRunFailuresForTest(p, run.ID, []domain.FailureSummary{{
		Stage:      domain.StageDeploy,
		Category:   "delivery",
		Target:     "stable-deploy",
		Reason:     "部署结果仍待确认",
		Suggestion: "请先处理部署异常",
		Resolved:   false,
	}})
	if _, err := p.DeployRun(run.ID); err == nil {
		t.Fatal("expected unresolved failures to block deploy")
	}
	p.mu.Lock()
	p.runs[run.ID].Failures[0].Resolved = true
	p.mu.Unlock()
	updated, err := p.DeployRun(run.ID)
	if err != nil {
		t.Fatalf("expected resolved failures not to block deploy, got %v", err)
	}
	if len(updated.Deployments) == 0 {
		t.Fatal("expected deployment record after deploy succeeds")
	}
}

func TestDeployStageFailureBuildsStructuredDeliveryFailure(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, err := p.CreateRun(domain.CreateRunInput{
		Title:               "部署失败聚合",
		Goal:                "trigger [fail:deploy:deploy-smoke]",
		Stages:              []domain.StageKind{domain.StageDeploy},
		RemoteDeployEnabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	p.tick()
	updated, err := p.GetRun(run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != domain.RunFailed {
		t.Fatalf("expected run failed, got %s", updated.Status)
	}
	if len(updated.Failures) != 1 {
		t.Fatalf("expected 1 deployment failure, got %d", len(updated.Failures))
	}
	failure := updated.Failures[0]
	if failure.Category != "delivery" {
		t.Fatalf("expected delivery category, got %q", failure.Category)
	}
	if failure.Target != "deploy-smoke" {
		t.Fatalf("expected deploy-smoke target, got %q", failure.Target)
	}
	if !strings.Contains(failure.Reason, "部署后检查未通过") {
		t.Fatalf("expected delivery reason, got %q", failure.Reason)
	}
	if !strings.Contains(failure.Suggestion, "健康检查") {
		t.Fatalf("expected deploy-specific suggestion, got %q", failure.Suggestion)
	}
}

func TestNormalizeAutoRepairModeByPolicy(t *testing.T) {
	if got := normalizeAutoRepairMode(true, "aggressive", domain.RuntimePolicy{Profile: "home-lite"}); got != "lite" {
		t.Fatalf("expected lite on home-lite, got %s", got)
	}
	if got := normalizeAutoRepairMode(true, "aggressive", domain.RuntimePolicy{Profile: "balanced-hybrid"}); got != "standard" {
		t.Fatalf("expected standard on balanced-hybrid, got %s", got)
	}
	if got := normalizeAutoRepairMode(true, "", domain.RuntimePolicy{Profile: "adaptive-performance"}); got != "aggressive" {
		t.Fatalf("expected aggressive default on adaptive-performance, got %s", got)
	}
	if got := normalizeAutoRepairMode(false, "aggressive", domain.RuntimePolicy{Profile: "adaptive-performance"}); got != "off" {
		t.Fatalf("expected off when disabled, got %s", got)
	}
}

func TestAutoRepairAttemptLimit(t *testing.T) {
	if got := autoRepairAttemptLimit("lite"); got != 1 {
		t.Fatalf("expected 1, got %d", got)
	}
	if got := autoRepairAttemptLimit("standard"); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
	if got := autoRepairAttemptLimit("aggressive"); got != 3 {
		t.Fatalf("expected 3, got %d", got)
	}
}

func TestAnalyzeGoalBuildsDeployAwarePlan(t *testing.T) {
	plan := AnalyzeGoal(AnalyzeInput{Title: "一键部署平台", Goal: "做一个 Web 平台并自动部署到 VPS"})
	if plan.ProjectKind != "web" {
		t.Fatalf("expected web project kind, got %s", plan.ProjectKind)
	}
	if plan.RecommendedBundle.ID != "bundle-local-delivery" {
		t.Fatalf("expected delivery bundle, got %s", plan.RecommendedBundle.ID)
	}
	foundDeploy := false
	for _, stage := range plan.RecommendedStages {
		if stage == domain.StageDeploy {
			foundDeploy = true
		}
	}
	if !foundDeploy {
		t.Fatal("expected deploy stage in analysis")
	}
	if len(plan.RecommendedStages) < 10 {
		t.Fatalf("expected broad 12-step skeleton, got %d stages", len(plan.RecommendedStages))
	}
	if plan.RecommendedStages[0] != domain.StageIntent {
		t.Fatalf("expected intent as first recommended stage, got %s", plan.RecommendedStages[0])
	}
	if plan.RecommendedStages[len(plan.RecommendedStages)-1] != domain.StageFinalize && plan.RecommendedStages[len(plan.RecommendedStages)-1] != domain.StageRepair {
		t.Fatalf("expected finalize or repair as last recommended stage, got %s", plan.RecommendedStages[len(plan.RecommendedStages)-1])
	}
	if len(plan.RecommendedTools) == 0 || len(plan.RecommendedTests) == 0 {
		t.Fatal("expected recommended tools and tests")
	}
	if len(plan.LoopBlueprint) != 12 {
		t.Fatalf("expected 12 loop blueprint steps, got %d", len(plan.LoopBlueprint))
	}
	if len(plan.RequirementSpec.FunctionalRequirements) == 0 {
		t.Fatal("expected requirement spec functional requirements")
	}
	if len(plan.TaskQueue) == 0 {
		t.Fatal("expected task queue")
	}
	if len(plan.Checkpoints) == 0 {
		t.Fatal("expected checkpoints")
	}
}

func TestAnalyzeGoalIncludesReviewAndSecurityForAuthTasks(t *testing.T) {
	plan := AnalyzeGoal(AnalyzeInput{Title: "登录权限系统", Goal: "开发带登录、权限和 token 校验的后台"})
	foundReview := false
	foundSecurity := false
	for _, item := range plan.RecommendedTests {
		if item == "code-review" {
			foundReview = true
		}
		if item == "security-scan" {
			foundSecurity = true
		}
	}
	if !foundReview {
		t.Fatal("expected code review in recommended tests")
	}
	if !foundSecurity {
		t.Fatal("expected security scan in recommended tests")
	}
}

func TestTickWithPolicyRespectsConcurrentRunLimit(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	first, err := p.createRunWithPolicy(domain.CreateRunInput{
		Title:  "任务一",
		Goal:   "执行第一条任务",
		Stages: []domain.StageKind{domain.StageAnalyze, domain.StageTest},
	}, domain.RuntimePolicy{Profile: "home-lite", MaxConcurrentRuns: 1})
	if err != nil {
		t.Fatal(err)
	}
	second, err := p.createRunWithPolicy(domain.CreateRunInput{
		Title:  "任务二",
		Goal:   "执行第二条任务",
		Stages: []domain.StageKind{domain.StageAnalyze, domain.StageTest},
	}, domain.RuntimePolicy{Profile: "home-lite", MaxConcurrentRuns: 1})
	if err != nil {
		t.Fatal(err)
	}
	p.tickWithPolicy(domain.RuntimePolicy{Profile: "home-lite", MaxConcurrentRuns: 1})
	updatedFirst, err := p.GetRun(first.ID)
	if err != nil {
		t.Fatal(err)
	}
	updatedSecond, err := p.GetRun(second.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updatedFirst.Status != domain.RunRunning {
		t.Fatalf("expected first run running, got %s", updatedFirst.Status)
	}
	if updatedSecond.Status != domain.RunQueued {
		t.Fatalf("expected second run queued under limit, got %s", updatedSecond.Status)
	}
}

func TestSummaryIncludesAdaptiveProfileDetails(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	summary := p.Summary()
	if summary.SystemProfile.RecommendedConcurrency <= 0 {
		t.Fatal("expected recommended concurrency in system profile")
	}
	if summary.SystemProfile.StrategySummary == "" {
		t.Fatal("expected strategy summary in system profile")
	}
	if summary.RuntimePolicy.MaxConcurrentRuns <= 0 {
		t.Fatal("expected runtime policy concurrent run limit")
	}
	if summary.RuntimePolicy.ValidationDepth == "" {
		t.Fatal("expected runtime policy validation depth")
	}
	if len(summary.AtomicTools) == 0 {
		t.Fatal("expected atomic tools in summary")
	}
}

func TestInvokeAtomicToolRecordsInvocation(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	invocation, err := p.InvokeAtomicTool("tool-workspace-patch")
	if err != nil {
		t.Fatal(err)
	}
	if invocation.ToolID != "tool-workspace-patch" {
		t.Fatalf("unexpected tool invocation id %s", invocation.ToolID)
	}
	summary := p.Summary()
	if len(summary.ToolInvocations) == 0 {
		t.Fatal("expected invocation recorded in summary")
	}
}

func TestAtomicToolCatalogCoversBroadCategoriesWithLocalFirst(t *testing.T) {
	tools := atomicToolCatalog(domain.SystemProfile{Tier: "balanced"})
	if len(tools) < 10 {
		t.Fatalf("expected broad atomic tool coverage, got %d", len(tools))
	}
	categories := map[string]bool{}
	localFirstCount := 0
	for _, item := range tools {
		categories[item.Category] = true
		if item.LocalFirst {
			localFirstCount++
		}
		if len(item.CoverageTags) == 0 {
			t.Fatalf("expected coverage tags for %s", item.ID)
		}
	}
	for _, expected := range []string{"planner", "code", "quality", "release", "ops", "research", "knowledge", "office", "daily", "assets", "media"} {
		if !categories[expected] {
			t.Fatalf("expected category %s in atomic tool catalog", expected)
		}
	}
	if localFirstCount < len(tools)-1 {
		t.Fatalf("expected most atomic tools local-first, got %d of %d", localFirstCount, len(tools))
	}
}

func TestInvokeAtomicToolRejectsRestrictedCapability(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	profile := domain.SystemProfile{Tier: "compact"}
	tools := atomicToolCatalog(profile)
	foundRestricted := false
	foundBuildOptional := false
	for _, item := range tools {
		if item.ID == "tool-release-delivery" {
			if item.Allowed {
				t.Fatal("expected delivery tool restricted on compact tier")
			}
			foundRestricted = true
		}
		if item.ID == "tool-build-artifact" {
			if item.ActivationMode != "optional" {
				t.Fatalf("expected build artifact optional on compact tier, got %s", item.ActivationMode)
			}
			foundBuildOptional = true
		}
	}
	if !foundRestricted {
		t.Fatal("expected delivery atomic tool in registry")
	}
	if !foundBuildOptional {
		t.Fatal("expected build atomic tool in registry")
	}
}

func TestDecorateFeatureToggleCompactWarning(t *testing.T) {
	profile := domain.SystemProfile{Tier: "compact"}
	toggle := decorateFeatureToggle(domain.FeatureToggle{ID: "builtin-model-packs"}, profile)
	if toggle.Recommended {
		t.Fatal("expected builtin model packs not recommended on compact tier")
	}
	if toggle.Warning == "" {
		t.Fatal("expected compact warning")
	}
}

func TestDecorateBuiltinModelPackRestriction(t *testing.T) {
	profile := domain.SystemProfile{Tier: "compact"}
	pack := decorateBuiltinModelPack(domain.BuiltinModelPack{ID: "pack-deepseek-u-7b", SizeTier: "7B"}, profile)
	if pack.Allowed {
		t.Fatal("expected deepseek 7b restricted on compact tier")
	}
	if pack.PolicyMode != "display-only" {
		t.Fatalf("expected compact tier pack display-only, got %s", pack.PolicyMode)
	}
	if pack.PolicyDecision.Action != "display-only" {
		t.Fatalf("expected display-only decision action, got %s", pack.PolicyDecision.Action)
	}
	if pack.Warning == "" {
		t.Fatal("expected warning for restricted pack")
	}
}

func TestDecorateBuiltinModelPackLargeTierWarning(t *testing.T) {
	profile := domain.SystemProfile{Tier: "balanced"}
	pack := decorateBuiltinModelPack(domain.BuiltinModelPack{ID: "pack-command-u-70b", SizeTier: "70B"}, profile)
	if pack.Allowed {
		t.Fatal("expected 70b pack restricted on balanced tier")
	}
	if pack.PolicyMode != "externally-managed" {
		t.Fatalf("expected 70b balanced pack externally-managed, got %s", pack.PolicyMode)
	}
	if pack.Warning == "" {
		t.Fatal("expected warning for large tier pack")
	}
}

func TestDecorateBuiltinModelPackBalancedOnDemand(t *testing.T) {
	profile := domain.SystemProfile{Tier: "balanced"}
	pack := decorateBuiltinModelPack(domain.BuiltinModelPack{ID: "pack-qwen-u-4b", SizeTier: "4B"}, profile)
	if !pack.Allowed {
		t.Fatal("expected 4b balanced pack allowed")
	}
	if pack.PolicyMode != "on-demand" {
		t.Fatalf("expected balanced 4b pack on-demand, got %s", pack.PolicyMode)
	}
	if pack.PolicyDecision.Action != "on-demand" {
		t.Fatalf("expected on-demand decision action, got %s", pack.PolicyDecision.Action)
	}
}

func TestFinalizeBuiltinModelPackState(t *testing.T) {
	ready := finalizeBuiltinModelPackState(domain.BuiltinModelPack{Enabled: true, Downloaded: true})
	if ready.InstallState != "ready" {
		t.Fatalf("expected ready state, got %s", ready.InstallState)
	}
	queued := finalizeBuiltinModelPackState(domain.BuiltinModelPack{InstallState: "queued"})
	if queued.StatusDetail == "" {
		t.Fatal("expected queued state detail")
	}
	disabled := finalizeBuiltinModelPackState(domain.BuiltinModelPack{Downloaded: true})
	if disabled.InstallState != "disabled" {
		t.Fatalf("expected disabled state, got %s", disabled.InstallState)
	}
	removed := finalizeBuiltinModelPackState(domain.BuiltinModelPack{InstallState: "removed"})
	if removed.StatusDetail == "" {
		t.Fatal("expected removed state detail")
	}
}

func TestSetBuiltinModelPackStartsInstallPipeline(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	pack, err := p.SetBuiltinModelPack("pack-qwen-u-4b", true, true, false)
	if err != nil {
		t.Fatalf("set builtin model pack failed: %v", err)
	}
	if pack.InstallState != "queued" {
		t.Fatalf("expected queued state, got %s", pack.InstallState)
	}
	p.tick()
	p.tick()
	p.tick()
	summary := p.Summary()
	for _, item := range summary.BuiltinModelPacks {
		if item.ID == "pack-qwen-u-4b" {
			if item.InstallState != "ready" {
				t.Fatalf("expected ready after pipeline, got %s", item.InstallState)
			}
			if !item.Downloaded || !item.Enabled {
				t.Fatal("expected pack downloaded and enabled after pipeline")
			}
			return
		}
	}
	t.Fatal("expected builtin model pack in summary")
}

func TestCreateRunWithCompactRuntimePolicyDisablesDeploy(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, err := p.createRunWithPolicy(domain.CreateRunInput{
		Title:               "低配运行策略",
		Goal:                "做一个平台并部署到 VPS",
		Stages:              []domain.StageKind{domain.StageAnalyze, domain.StageImplement, domain.StageTest, domain.StageDeploy},
		AutoRepairEnabled:   true,
		AutoRepairMode:      "aggressive",
		RemoteDeployEnabled: true,
		Tools: []domain.ToolConfig{
			{Name: "workspace", Enabled: true},
			{Name: "tests", Enabled: true},
			{Name: "deploy", Enabled: true},
		},
	}, runtimePolicy(domain.SystemProfile{Tier: "compact"}))
	if err != nil {
		t.Fatalf("create run failed: %v", err)
	}
	if run.RemoteDeployEnabled {
		t.Fatal("expected remote deploy disabled under compact runtime policy")
	}
	if run.AutoRepairMode != "lite" {
		t.Fatalf("expected auto repair mode narrowed to lite, got %s", run.AutoRepairMode)
	}
	if len(run.PolicyDecisions) == 0 {
		t.Fatal("expected structured policy decisions")
	}
	for _, stage := range run.Stages {
		if stage.Kind == domain.StageDeploy {
			t.Fatal("expected deploy stage removed under compact runtime policy")
		}
	}
	toolStates := map[string]bool{}
	for _, tool := range run.Stages[0].Tools {
		toolStates[tool.Name] = tool.Enabled
	}
	if toolStates["deploy"] {
		t.Fatal("expected deploy tool disabled by compact runtime policy")
	}
	foundAdjustment := false
	foundStructured := false
	for _, log := range run.Logs {
		if strings.Contains(log.Message, "自动修复强度已设为 lite") {
			foundAdjustment = true
			break
		}
	}
	for _, item := range run.PolicyDecisions {
		if item.Area == "repair" && strings.Contains(item.Message, "自动修复强度已设为 lite") {
			foundStructured = true
			break
		}
	}
	if !foundAdjustment {
		t.Fatal("expected runtime policy adjustment log")
	}
	if !foundStructured {
		t.Fatal("expected structured policy decision for repair mode")
	}
	if len(run.AssembledTools) == 0 {
		t.Fatal("expected assembled atomic tools")
	}
}

func TestCreateRunWithRuntimePolicyLimitsHeavyTools(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, err := p.createRunWithPolicy(domain.CreateRunInput{
		Title:  "限制重工具",
		Goal:   "做一次构建和部署",
		Stages: []domain.StageKind{domain.StageAnalyze, domain.StageImplement, domain.StageTest, domain.StageDeploy},
		Tools: []domain.ToolConfig{
			{Name: "workspace", Enabled: true},
			{Name: "tests", Enabled: true},
			{Name: "build", Enabled: true},
			{Name: "deploy", Enabled: true},
		},
	}, domain.RuntimePolicy{MaxHeavyActions: 1, AllowBackgroundJobs: true})
	if err != nil {
		t.Fatalf("create run failed: %v", err)
	}
	toolStates := map[string]bool{}
	for _, tool := range run.Stages[0].Tools {
		toolStates[tool.Name] = tool.Enabled
	}
	if !toolStates["build"] {
		t.Fatal("expected first heavy tool to remain enabled")
	}
	if toolStates["deploy"] {
		t.Fatal("expected deploy disabled after reaching heavy tool limit")
	}
	for _, stage := range run.Stages {
		if stage.Kind == domain.StageDeploy {
			t.Fatal("expected deploy stage removed when deploy tool is disabled")
		}
	}
}

func TestCreateRunAssemblesAtomicToolsByDedupGroup(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, err := p.createRunWithPolicy(domain.CreateRunInput{
		Title:  "能力装配",
		Goal:   "整理代码结构并完成验证",
		Stages: []domain.StageKind{domain.StageAnalyze, domain.StageImplement, domain.StageTest},
		Tools: []domain.ToolConfig{
			{Name: "analysis", Enabled: true},
			{Name: "workspace", Enabled: true},
			{Name: "format", Enabled: true},
			{Name: "tests", Enabled: true},
			{Name: "lint", Enabled: true},
		},
	}, domain.RuntimePolicy{Profile: "balanced-hybrid", MaxHeavyActions: 1, AllowBackgroundJobs: true, AllowLocalModels: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(run.AssembledTools) == 0 {
		t.Fatal("expected assembled tools")
	}
	groups := map[string]int{}
	foundWorkspacePrimary := false
	for _, item := range run.AssembledTools {
		groups[item.DedupGroup]++
		if item.ID == "tool-workspace-patch" {
			foundWorkspacePrimary = true
		}
	}
	if groups["workspace-edit"] != 1 {
		t.Fatalf("expected single assembled tool for workspace-edit group, got %d", groups["workspace-edit"])
	}
	if !foundWorkspacePrimary {
		t.Fatal("expected workspace patch selected as primary assembled tool")
	}
}

func TestTemplateBundleInfluencesAtomicToolOrdering(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, err := p.CreateRun(domain.CreateRunInput{
		Title:      "模板交付优先",
		Goal:       "按模板完成安全发布",
		TemplateID: "tpl-safe-deploy",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(run.AssembledTools) == 0 {
		t.Fatal("expected assembled tools")
	}
	if run.AssembledTools[0].Category != "release" && run.AssembledTools[0].Category != "ops" {
		t.Fatalf("expected delivery-oriented category first, got %s", run.AssembledTools[0].Category)
	}
	foundBundleDecision := false
	for _, item := range run.PolicyDecisions {
		if item.Area == "atomic-tool" && item.Reason == "bundle-local-delivery" {
			foundBundleDecision = true
			break
		}
	}
	if !foundBundleDecision {
		t.Fatal("expected atomic tool decisions to reflect template bundle preference")
	}
}

func TestPatchRequirementsRebuildsAssembledTools(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, err := p.CreateRun(domain.CreateRunInput{
		Title:  "补需重算",
		Goal:   "先做代码任务",
		Stages: []domain.StageKind{domain.StageAnalyze, domain.StageImplement, domain.StageTest},
		Tools: []domain.ToolConfig{
			{Name: "analysis", Enabled: true},
			{Name: "workspace", Enabled: true},
			{Name: "tests", Enabled: true},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	patched, err := p.PatchRequirements(run.ID, "补充文档整理与研究摘要")
	if err != nil {
		t.Fatal(err)
	}
	if len(patched.AssembledTools) == 0 {
		t.Fatal("expected assembled tools after patch")
	}
	foundResearchOrDocs := false
	for _, item := range patched.AssembledTools {
		if item.ID == "tool-doc-governance" || item.ID == "tool-research-brief" {
			foundResearchOrDocs = true
		}
	}
	if !foundResearchOrDocs {
		t.Fatal("expected patched run to rebuild assembled tools with updated analysis context")
	}
}

func TestStageSummaryIncludesAssembledToolNames(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	run, err := p.createRunWithPolicy(domain.CreateRunInput{
		Title:  "阶段摘要",
		Goal:   "执行代码改写并验证",
		Stages: []domain.StageKind{domain.StageAnalyze, domain.StageImplement, domain.StageTest},
		Tools: []domain.ToolConfig{
			{Name: "analysis", Enabled: true},
			{Name: "workspace", Enabled: true},
			{Name: "tests", Enabled: true},
		},
	}, domain.RuntimePolicy{Profile: "balanced-hybrid", MaxHeavyActions: 1, AllowBackgroundJobs: true, AllowLocalModels: true})
	if err != nil {
		t.Fatal(err)
	}
	summary := p.summaryForStage(&run, domain.StageImplement)
	if !strings.Contains(summary, "当前装配能力包括") {
		t.Fatalf("expected assembled tools in stage summary, got %s", summary)
	}
}

func TestSetBuiltinModelPackWithCompactRuntimePolicyRejectsActivation(t *testing.T) {
	p := NewPlatform()
	t.Cleanup(p.Close)
	_, err := p.setBuiltinModelPackWithPolicy("pack-qwen-u-4b", true, true, false, runtimePolicy(domain.SystemProfile{Tier: "compact"}))
	if err == nil {
		t.Fatal("expected compact runtime policy to reject local model activation")
	}
	if !errors.Is(err, ErrSettingNotAllowed) {
		t.Fatalf("expected ErrSettingNotAllowed, got %v", err)
	}
}

func TestAnalyzeGoalSupportsResearchAndOffice(t *testing.T) {
	plan := AnalyzeGoal(AnalyzeInput{Title: "整理研究报告", Goal: "汇总资料并生成办公汇报"})
	if plan.ProjectKind != "research" && plan.ProjectKind != "office" {
		t.Fatalf("expected research or office project kind, got %s", plan.ProjectKind)
	}
	if plan.RecommendedBundle.ID != "bundle-knowledge-office" {
		t.Fatalf("expected knowledge-office bundle, got %s", plan.RecommendedBundle.ID)
	}
	if len(plan.RecommendedTests) == 0 {
		t.Fatal("expected recommended tests for non-code tasks")
	}
	foundReview := false
	for _, item := range plan.RecommendedTests {
		if item == "review-check" {
			foundReview = true
		}
	}
	if !foundReview {
		t.Fatal("expected review-check in recommended tests")
	}
}

func TestCapabilityProfilesCompactTier(t *testing.T) {
	profiles := capabilityProfiles(domain.SystemProfile{Tier: "compact"})
	if len(profiles) == 0 {
		t.Fatal("expected capability profiles")
	}
	for _, profile := range profiles {
		if profile.ID == "cap-local-models" {
			if profile.Mode != "display-only" {
				t.Fatalf("expected local models display-only on compact tier, got %s", profile.Mode)
			}
			if profile.Allowed {
				t.Fatal("expected local models not allowed on compact tier")
			}
			return
		}
	}
	t.Fatal("expected cap-local-models profile")
}

func TestOptimizationProfilesCompactTier(t *testing.T) {
	profiles := optimizationProfiles(domain.SystemProfile{Tier: "compact"})
	if len(profiles) < 3 {
		t.Fatal("expected compact optimization profiles")
	}
	if profiles[0].Priority != "high" {
		t.Fatalf("expected highest priority first, got %s", profiles[0].Priority)
	}
	found := false
	for _, profile := range profiles {
		if profile.ID == "opt-compact-api-first" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected compact api-first optimization profile")
	}
}

func TestRuntimePolicyCompactTier(t *testing.T) {
	policy := runtimePolicy(domain.SystemProfile{Tier: "compact"})
	if policy.Profile != "home-lite" {
		t.Fatalf("expected home-lite profile, got %s", policy.Profile)
	}
	if policy.AllowLocalModels {
		t.Fatal("expected local models disabled in compact runtime policy")
	}
	if policy.MaxHeavyActions != 1 {
		t.Fatalf("expected max heavy actions 1, got %d", policy.MaxHeavyActions)
	}
	if policy.MaxConcurrentRuns != 1 {
		t.Fatalf("expected max concurrent runs 1, got %d", policy.MaxConcurrentRuns)
	}
	if policy.ValidationDepth != "essential" {
		t.Fatalf("expected essential validation depth, got %s", policy.ValidationDepth)
	}
}
