package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"autocode-platform/internal/domain"
	"autocode-platform/internal/service"
)

func TestCreateRunEndpoint(t *testing.T) {
	platform := service.NewPlatform()
	t.Cleanup(platform.Close)
	server := New(platform)
	payload := domain.CreateRunInput{
		Title:               "闭环平台",
		Goal:                "初始化系统",
		TemplateID:          "tpl-full-loop",
		TemplateSource:      "manual-select",
		TestMode:            "light",
		SelectedTests:       []string{"static-check", "review-check"},
		Stages:              []domain.StageKind{domain.StageAnalyze, domain.StageImplement, domain.StageTest},
		AutoRepairEnabled:   true,
		RemoteDeployEnabled: true,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/runs", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	var run domain.Run
	if err := json.Unmarshal(rec.Body.Bytes(), &run); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if run.Title != payload.Title {
		t.Fatalf("expected title %q, got %q", payload.Title, run.Title)
	}
	if len(run.Stages) != 3 {
		t.Fatalf("expected 3 stages, got %d", len(run.Stages))
	}
	if run.TemplateSource != payload.TemplateSource {
		t.Fatalf("expected template source %q, got %q", payload.TemplateSource, run.TemplateSource)
	}
	if run.TestMode != payload.TestMode {
		t.Fatalf("expected test mode %q, got %q", payload.TestMode, run.TestMode)
	}
}

func TestStaticIndexServed(t *testing.T) {
	platform := service.NewPlatform()
	t.Cleanup(platform.Close)
	server := New(platform)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("繁星")) {
		t.Fatal("expected index page content")
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store, no-cache, must-revalidate" {
		t.Fatalf("expected static cache control header, got %q", got)
	}
	if got := rec.Header().Get("Pragma"); got != "no-cache" {
		t.Fatalf("expected pragma header, got %q", got)
	}
	if got := rec.Header().Get("Expires"); got != "0" {
		t.Fatalf("expected expires header, got %q", got)
	}
}

func TestStaticAppScriptIncludesDeploymentInteractions(t *testing.T) {
	platform := service.NewPlatform()
	t.Cleanup(platform.Close)
	server := New(platform)
	req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.Bytes()
	checks := [][]byte{
		[]byte("const stageOptions = ['intent', 'context', 'plan', 'resource', 'model', 'tool', 'implement', 'result', 'test', 'deploy', 'repair', 'finalize'];"),
		[]byte("intent: '任务识别'"),
		[]byte("context: '上下文汇总'"),
		[]byte("plan: '方案规划'"),
		[]byte("resource: '资源收束'"),
		[]byte("model: '模型选择'"),
		[]byte("tool: '工具装配'"),
		[]byte("result: '结果整理'"),
		[]byte("finalize: '总结沉淀'"),
		[]byte("runDeploymentAction"),
		[]byte("focusDeploymentTimelineRecord"),
		[]byte("buildRunDeploymentSignature"),
		[]byte("statusKey"),
		[]byte("isDeploymentStatus"),
		[]byte("shouldShowDeploymentTimelinePreviewAction"),
		[]byte("deploymentTimelinePreviewActionLabel"),
		[]byte("deriveRollbackSuggestionAction"),
		[]byte("deploymentTimelineFocusedItem"),
		[]byte("suppressError = false"),
		[]byte("rethrow = false"),
		[]byte("suppressError: true, rethrow: true"),
		[]byte("deployment-version-link"),
		[]byte("action-button-busy"),
		[]byte("deployment-timeline-item-focused"),
		[]byte("data-action=\"view-deployment\""),
		[]byte("data-action=\"open-preview\""),
		[]byte("data-action=\"deploy-run\""),
		[]byte("data-action=\"revalidate-run\""),
		[]byte("data-action=\"rollback-deployment\""),
		[]byte("data-action=\"view-effective-version\""),
		[]byte("data-action=\"open-preview-from-timeline\""),
		[]byte("部署准备"),
		[]byte("回退完成，待复验"),
		[]byte("部署完成，待查看预览"),
		[]byte("稳定部署失败"),
		[]byte("function failureTargetLabel(target)"),
		[]byte("function sortFailures(items)"),
		[]byte("function blockerCopyForFailure(failure)"),
		[]byte("function failureSourceLabel(failure)"),
		[]byte("function failureRecordLabel(failure)"),
		[]byte("function devActivityLabel(kind)"),
		[]byte("function devActivitySummary(item)"),
		[]byte("function devActivityMeta(item)"),
		[]byte("function runLogLevelLabel(level)"),
		[]byte("function policyDecisionAreaLabel(area)"),
		[]byte("function policyDecisionActionLabel(action)"),
		[]byte("function policyDecisionReasonLabel(reason)"),
		[]byte("function dedupeFailureAdvice(items)"),
		[]byte("function runFailureHeadline(run)"),
		[]byte("function runFailureSummary(run)"),
		[]byte("function runFailureContextMeta(run)"),
		[]byte("function runFailureActionHint(failure)"),
		[]byte("function resolvedFailureHint(failure)"),
		[]byte("失败摘要"),
		[]byte("当前进展稳定"),
		[]byte("当前没有待处理问题，可继续推进。"),
		[]byte("当前可直接执行："),
		[]byte("当前暂无可直接执行的建议动作。"),
		[]byte("检查要求："),
		[]byte("覆盖范围："),
		[]byte("结构状态"),
		[]byte("保存文件"),
		[]byte("回滚文件"),
		[]byte("打开预览"),
		[]byte("执行命令"),
		[]byte("进展"),
		[]byte("提醒"),
		[]byte("异常"),
		[]byte("工具策略"),
		[]byte("阶段调整"),
		[]byte("部署策略"),
		[]byte("修复策略"),
		[]byte("能力装配"),
		[]byte("默认关闭"),
		[]byte("后台重任务受限"),
		[]byte("重动作数量受限"),
		[]byte("步骤 ${index + 1}"),
		[]byte("建议阶段"),
		[]byte("建议能力"),
		[]byte("当前验证"),
		[]byte("推荐验证"),
		[]byte("同类能力"),
		[]byte("已按同类能力整理"),
		[]byte("接入门槛"),
		[]byte("稳定性"),
		[]byte("任务贴合度"),
		[]byte("已完成接入识别"),
		[]byte("接入变量名"),
		[]byte("当前可用"),
		[]byte("当前建议"),
		[]byte("备选方式"),
		[]byte("function invocationStatusLabel(status)"),
		[]byte("function terminalSessionStatusLabel(status)"),
		[]byte("function auditStatusLabel(status)"),
		[]byte("function queueAgentLabel(agent)"),
		[]byte("function queuePhaseLabel(phase)"),
		[]byte("function providerLabel(provider)"),
		[]byte("已关联任务"),
		[]byte("当前机器默认关闭稳定部署"),
		[]byte("按机器调整"),
		[]byte("运行方式"),
		[]byte("机器级别"),
		[]byte("当前建议能力"),
		[]byte("能力总览"),
		[]byte("全系统运行方式"),
		[]byte("全局建议"),
		[]byte("工作台"),
		[]byte("任务中枢"),
		[]byte("轻量入口"),
		[]byte("已更新功能设置"),
		[]byte("功能设置更新失败"),
		[]byte("使用提示"),
		[]byte("当前方式"),
		[]byte("运行信息"),
		[]byte("模板来源"),
		[]byte("测试 "),
		[]byte("预览确认已完成"),
		[]byte("部署后检查已复验通过"),
		[]byte("稳定部署问题已处理"),
		[]byte("请先完成预览确认，再继续后续推进。"),
		[]byte("请先完成部署后检查，再继续后续推进。"),
		[]byte("请先处理稳定部署问题，再继续后续推进。"),
		[]byte("验证待处理"),
		[]byte("部署待处理"),
		[]byte("处理建议"),
		[]byte("已处理"),
		[]byte("处理记录"),
		[]byte("当前无失败摘要"),
		[]byte("当前没有额外上下文需要关注。"),
		[]byte("关联目标"),
		[]byte("当前暂无处理记录"),
		[]byte("处理阶段："),
		[]byte("来源："),
		[]byte("关联问题："),
		[]byte("记录时间："),
		[]byte("deploy-smoke"),
		[]byte("preview-check"),
		[]byte("验证失败"),
		[]byte("部署失败"),
		[]byte("修复建议"),
		[]byte("处理记录"),
		[]byte("暂无问题"),
		[]byte("当前没有待处理项。"),
		[]byte("当前没有新的处理建议。"),
		[]byte("版本回退失败"),
		[]byte("复验失败"),
		[]byte("打开预览失败"),
	}
	for _, check := range checks {
		if !bytes.Contains(body, check) {
			t.Fatalf("expected app.js to contain %q", string(check))
		}
	}
}

func TestStaticAppStylesIncludeDeploymentInteractionClasses(t *testing.T) {
	platform := service.NewPlatform()
	t.Cleanup(platform.Close)
	server := New(platform)
	req := httptest.NewRequest(http.MethodGet, "/app.css", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.Bytes()
	checks := [][]byte{
		[]byte(".deployment-version-link"),
		[]byte(".action-button-busy"),
		[]byte(".deployment-timeline-item-focused"),
	}
	for _, check := range checks {
		if !bytes.Contains(body, check) {
			t.Fatalf("expected app.css to contain %q", string(check))
		}
	}
}

func TestSystemSummaryEndpoint(t *testing.T) {
	if err := os.Setenv("OPENAI_API_KEY", "present"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("OPENAI_API_KEY") })
	platform := service.NewPlatform()
	t.Cleanup(platform.Close)
	server := New(platform)
	req := httptest.NewRequest(http.MethodGet, "/api/system", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("OPENAI_API_KEY")) {
		t.Fatal("expected credential env name in response")
	}
}

func TestAnalyzeEndpoint(t *testing.T) {
	platform := service.NewPlatform()
	t.Cleanup(platform.Close)
	server := New(platform)
	body := bytes.NewBufferString(`{"title":"一句话建站","goal":"做一个带部署的 Web 平台"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/intake/analyze", body)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("recommendedStages")) {
		t.Fatal("expected analysis payload")
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("loopBlueprint")) {
		t.Fatal("expected loop blueprint in analysis payload")
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("recommendedBundle")) {
		t.Fatal("expected recommended bundle in analysis payload")
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("requirementSpec")) {
		t.Fatal("expected requirement spec in analysis payload")
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("taskQueue")) {
		t.Fatal("expected task queue in analysis payload")
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("checkpoints")) {
		t.Fatal("expected checkpoints in analysis payload")
	}
}

func TestAtomicToolsEndpoint(t *testing.T) {
	platform := service.NewPlatform()
	t.Cleanup(platform.Close)
	server := New(platform)
	req := httptest.NewRequest(http.MethodGet, "/api/tools", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("tool-workspace-patch")) {
		t.Fatal("expected atomic tool registry payload")
	}
}

func TestAtomicToolInvokeEndpoint(t *testing.T) {
	platform := service.NewPlatform()
	t.Cleanup(platform.Close)
	server := New(platform)
	req := httptest.NewRequest(http.MethodPost, "/api/tools/tool-workspace-patch/invoke", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("tool-workspace-patch")) {
		t.Fatal("expected invocation payload")
	}
}

func TestTemplatesEndpointIncludesBundleAndStageDefaults(t *testing.T) {
	platform := service.NewPlatform()
	t.Cleanup(platform.Close)
	server := New(platform)
	req := httptest.NewRequest(http.MethodGet, "/api/templates", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("recommendedBundle")) {
		t.Fatal("expected recommended bundle in template payload")
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("defaultStages")) {
		t.Fatal("expected default stages in template payload")
	}
}

func TestAuditEndpoint(t *testing.T) {
	platform := service.NewPlatform()
	t.Cleanup(platform.Close)
	server := New(platform)
	req := httptest.NewRequest(http.MethodGet, "/api/system/audit", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("checks")) {
		t.Fatal("expected audit checks")
	}
}

func TestAdviceEndpoint(t *testing.T) {
	platform := service.NewPlatform()
	t.Cleanup(platform.Close)
	server := New(platform)
	body := bytes.NewBufferString(`{"mode":"ops-review","goal":"提升免SSH运维"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/system/advice", body)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("promptBundle")) {
		t.Fatal("expected prompt bundle")
	}
}

func TestAdviceEndpointExternalHostingMode(t *testing.T) {
	platform := service.NewPlatform()
	t.Cleanup(platform.Close)
	server := New(platform)
	body := bytes.NewBufferString(`{"mode":"external-hosting","goal":"为 70B 模型生成独立节点托管建议","target":"cluster"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/system/advice", body)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("external-hosting")) {
		t.Fatal("expected external hosting mode in response")
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("cluster")) {
		t.Fatal("expected hosting mode in response")
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("集群级托管")) {
		t.Fatal("expected hosting suggestion in response")
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("checklist")) {
		t.Fatal("expected hosting checklist in response")
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("consoleSteps")) {
		t.Fatal("expected console steps in response")
	}
}

func TestDevFileReadWriteAndRollback(t *testing.T) {
	platform := service.NewPlatform()
	t.Cleanup(platform.Close)
	server := New(platform)
	filePath := "internal/httpapi/web/.test-dev-file.txt"
	t.Cleanup(func() {
		_ = os.Remove("/workspace/" + filePath)
	})

	writeBody := bytes.NewBufferString(`{"path":"` + filePath + `","content":"first version"}`)
	writeReq := httptest.NewRequest(http.MethodPost, "/api/dev/file", writeBody)
	writeRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(writeRec, writeReq)
	if writeRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", writeRec.Code)
	}

	secondWriteBody := bytes.NewBufferString(`{"path":"` + filePath + `","content":"second version"}`)
	secondWriteReq := httptest.NewRequest(http.MethodPost, "/api/dev/file", secondWriteBody)
	secondWriteRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(secondWriteRec, secondWriteReq)
	if secondWriteRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on second write, got %d", secondWriteRec.Code)
	}

	rollbackReq := httptest.NewRequest(http.MethodPost, "/api/dev/file/rollback", bytes.NewBufferString(`{"path":"`+filePath+`"}`))
	rollbackRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rollbackRec, rollbackReq)
	if rollbackRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on rollback, got %d", rollbackRec.Code)
	}

	readReq := httptest.NewRequest(http.MethodGet, "/api/dev/file?path="+filePath, nil)
	readRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(readRec, readReq)
	if readRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on read, got %d", readRec.Code)
	}
	if !bytes.Contains(readRec.Body.Bytes(), []byte("first version")) {
		t.Fatal("expected rolled back file content")
	}
}

func TestDevTerminalExecAndSessions(t *testing.T) {
	platform := service.NewPlatform()
	t.Cleanup(platform.Close)
	server := New(platform)

	foregroundReq := httptest.NewRequest(http.MethodPost, "/api/dev/terminal/exec", bytes.NewBufferString(`{"command":"printf 'ok'","cwd":".","background":false}`))
	foregroundRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(foregroundRec, foregroundReq)
	if foregroundRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on foreground exec, got %d", foregroundRec.Code)
	}
	if !bytes.Contains(foregroundRec.Body.Bytes(), []byte(`"status":"completed"`)) {
		t.Fatal("expected completed foreground command")
	}

	backgroundReq := httptest.NewRequest(http.MethodPost, "/api/dev/terminal/exec", bytes.NewBufferString(`{"command":"printf 'bg'","cwd":".","background":true}`))
	backgroundRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(backgroundRec, backgroundReq)
	if backgroundRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on background exec, got %d", backgroundRec.Code)
	}
	var session map[string]any
	if err := json.Unmarshal(backgroundRec.Body.Bytes(), &session); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	id, _ := session["id"].(string)
	if strings.TrimSpace(id) == "" {
		t.Fatal("expected background session id")
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/dev/terminal/sessions", nil)
	listRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on session list, got %d", listRec.Code)
	}
	if !bytes.Contains(listRec.Body.Bytes(), []byte(id)) {
		t.Fatal("expected background session in list")
	}
}

func TestRunDevActivityEndpoint(t *testing.T) {
	platform := service.NewPlatform()
	t.Cleanup(platform.Close)
	server := New(platform)
	run, err := platform.CreateRun(domain.CreateRunInput{Title: "工作台绑定", Goal: "记录工作台活动"})
	if err != nil {
		t.Fatal(err)
	}
	body := bytes.NewBufferString(`{"kind":"file-save","target":"internal/httpapi/web/app.js","status":"completed","detail":"保存文件"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/runs/"+run.ID+"/dev-activity", body)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("devActivities")) {
		t.Fatal("expected dev activity payload")
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("internal/httpapi/web/app.js")) {
		t.Fatal("expected saved file target")
	}
}

func TestBackgroundSessionListSyncsRunUpdate(t *testing.T) {
	platform := service.NewPlatform()
	t.Cleanup(platform.Close)
	server := New(platform)
	run, err := platform.CreateRun(domain.CreateRunInput{Title: "后台命令回写", Goal: "把后台命令结果回写到任务"})
	if err != nil {
		t.Fatal(err)
	}
	execReq := httptest.NewRequest(http.MethodPost, "/api/dev/terminal/exec", bytes.NewBufferString(`{"command":"printf 'bg'","cwd":".","background":true,"runId":"`+run.ID+`"}`))
	execRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(execRec, execReq)
	if execRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on background exec, got %d", execRec.Code)
	}
	time.Sleep(80 * time.Millisecond)
	listReq := httptest.NewRequest(http.MethodGet, "/api/dev/terminal/sessions", nil)
	listRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on session list, got %d", listRec.Code)
	}
	if !bytes.Contains(listRec.Body.Bytes(), []byte("updatedRuns")) {
		t.Fatal("expected updated runs in session list response")
	}
	if !bytes.Contains(listRec.Body.Bytes(), []byte(run.ID)) {
		t.Fatal("expected bound run id in session sync response")
	}
	updated, err := platform.GetRun(run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.DevActivities) == 0 {
		t.Fatal("expected synced dev activity on run")
	}
}

func TestKillSessionSyncsRunUpdate(t *testing.T) {
	platform := service.NewPlatform()
	t.Cleanup(platform.Close)
	server := New(platform)
	run, err := platform.CreateRun(domain.CreateRunInput{Title: "停止后台命令", Goal: "把停止结果回写到任务"})
	if err != nil {
		t.Fatal(err)
	}
	execReq := httptest.NewRequest(http.MethodPost, "/api/dev/terminal/exec", bytes.NewBufferString(`{"command":"sleep 1","cwd":".","background":true,"runId":"`+run.ID+`"}`))
	execRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(execRec, execReq)
	if execRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on background exec, got %d", execRec.Code)
	}
	var session map[string]any
	if err := json.Unmarshal(execRec.Body.Bytes(), &session); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	id, _ := session["id"].(string)
	if strings.TrimSpace(id) == "" {
		t.Fatal("expected session id")
	}
	killReq := httptest.NewRequest(http.MethodPost, "/api/dev/terminal/sessions/"+id+"/kill", nil)
	killRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(killRec, killReq)
	if killRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on session kill, got %d", killRec.Code)
	}
	if !bytes.Contains(killRec.Body.Bytes(), []byte("updatedRuns")) {
		t.Fatal("expected updated runs in kill response")
	}
	updated, err := platform.GetRun(run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.DevActivities) == 0 {
		t.Fatal("expected synced dev activity after kill")
	}
	if updated.DevActivities[0].Status != "killed" {
		t.Fatalf("expected latest dev activity killed, got %q", updated.DevActivities[0].Status)
	}
}

func TestRollbackRunEndpoint(t *testing.T) {
	platform := service.NewPlatform()
	t.Cleanup(platform.Close)
	server := New(platform)
	run, err := platform.CreateRun(domain.CreateRunInput{
		Title:               "回退入口",
		Goal:                "执行稳定部署并回退到指定版本",
		Stages:              []domain.StageKind{domain.StageAnalyze, domain.StageDeploy},
		RemoteDeployEnabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	platform.TickForTest()
	platform.TickForTest()
	updated, err := platform.GetRun(run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Deployments) == 0 {
		t.Fatal("expected deployment record before rollback")
	}
	body := bytes.NewBufferString(`{"version":"` + updated.Deployments[0].Version + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/runs/"+run.ID+"/rollback", body)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("rollback")) {
		t.Fatal("expected rollback record in response")
	}
}

func TestRevalidateRunEndpoint(t *testing.T) {
	platform := service.NewPlatform()
	t.Cleanup(platform.Close)
	server := New(platform)
	run, err := platform.CreateRun(domain.CreateRunInput{
		Title:         "复验入口",
		Goal:          "回退后重新验证",
		SelectedTests: []string{"review-check", "security-scan"},
	})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/runs/"+run.ID+"/revalidate", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("testReports")) {
		t.Fatal("expected test reports in response")
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("revalidate")) {
		t.Fatal("expected revalidate record in response")
	}
}

func TestRevalidateEndpointResolvesMatchingFailures(t *testing.T) {
	platform := service.NewPlatform()
	t.Cleanup(platform.Close)
	server := New(platform)
	run, err := platform.CreateRun(domain.CreateRunInput{
		Title:         "复验联动结构",
		Goal:          "trigger [fail:test:security-scan]",
		Stages:        []domain.StageKind{domain.StageTest},
		SelectedTests: []string{"review-check", "security-scan"},
	})
	if err != nil {
		t.Fatal(err)
	}
	platform.TickForTest()
	req := httptest.NewRequest(http.MethodPost, "/api/runs/"+run.ID+"/revalidate", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var payload domain.Run
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(payload.Failures) == 0 {
		t.Fatal("expected failure history kept in response")
	}
	failure := payload.Failures[0]
	if failure.Category != "validation" {
		t.Fatalf("expected validation category, got %q", failure.Category)
	}
	if failure.Target != "security-scan" {
		t.Fatalf("expected security-scan target, got %q", failure.Target)
	}
	if !failure.Resolved {
		t.Fatal("expected revalidate to resolve matching failure in response")
	}
}

func TestDeployRunEndpoint(t *testing.T) {
	platform := service.NewPlatform()
	t.Cleanup(platform.Close)
	server := New(platform)
	run, err := platform.CreateRun(domain.CreateRunInput{
		Title:               "部署入口",
		Goal:                "手动执行稳定部署",
		Stages:              []domain.StageKind{domain.StageAnalyze, domain.StageDeploy},
		RemoteDeployEnabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/runs/"+run.ID+"/deploy", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("configured-vps")) {
		t.Fatal("expected deployment target in response")
	}
}

func TestRunsEndpointIncludesFailureFields(t *testing.T) {
	platform := service.NewPlatform()
	t.Cleanup(platform.Close)
	server := New(platform)
	run, err := platform.CreateRun(domain.CreateRunInput{
		Title:         "列表失败结构",
		Goal:          "trigger [fail:test:security-scan]",
		Stages:        []domain.StageKind{domain.StageTest},
		SelectedTests: []string{"review-check", "security-scan"},
	})
	if err != nil {
		t.Fatal(err)
	}
	platform.TickForTest()
	platform.TickForTest()
	payload := fetchRunsList(t, server)
	matched := findRunInList(t, payload.Items, run.ID)
	if matched == nil {
		t.Fatal("expected run in list response")
	}
	if len(matched.Failures) == 0 {
		t.Fatal("expected failures in list payload")
	}
	failure := matched.Failures[0]
	if failure.Category != "validation" {
		t.Fatalf("expected validation category, got %q", failure.Category)
	}
	if failure.Target != "security-scan" {
		t.Fatalf("expected security-scan target, got %q", failure.Target)
	}
	if failure.Resolved {
		t.Fatal("expected unresolved failure in list payload")
	}
}

func TestRunsEndpointKeepsResolvedFailureHistory(t *testing.T) {
	platform := service.NewPlatform()
	t.Cleanup(platform.Close)
	server := New(platform)
	run, err := platform.CreateRun(domain.CreateRunInput{
		Title:         "列表处理记录结构",
		Goal:          "trigger [fail:test:security-scan]",
		Stages:        []domain.StageKind{domain.StageTest},
		SelectedTests: []string{"review-check", "security-scan"},
	})
	if err != nil {
		t.Fatal(err)
	}
	platform.TickForTest()
	revalidateReq := httptest.NewRequest(http.MethodPost, "/api/runs/"+run.ID+"/revalidate", nil)
	revalidateRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(revalidateRec, revalidateReq)
	if revalidateRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", revalidateRec.Code)
	}
	payload := fetchRunsList(t, server)
	matched := findRunInList(t, payload.Items, run.ID)
	if len(matched.Failures) == 0 {
		t.Fatal("expected resolved failure history in list payload")
	}
	failure := matched.Failures[0]
	if failure.Category != "validation" {
		t.Fatalf("expected validation category, got %q", failure.Category)
	}
	if failure.Target != "security-scan" {
		t.Fatalf("expected security-scan target, got %q", failure.Target)
	}
	if !failure.Resolved {
		t.Fatal("expected resolved failure history in list payload")
	}
}

func TestRunsEndpointLeavesFailuresEmptyWhenNone(t *testing.T) {
	platform := service.NewPlatform()
	t.Cleanup(platform.Close)
	server := New(platform)
	run, err := platform.CreateRun(domain.CreateRunInput{
		Title:  "列表无失败结构",
		Goal:   "正常任务",
		Stages: []domain.StageKind{domain.StageIntent, domain.StageContext},
	})
	if err != nil {
		t.Fatal(err)
	}
	payload := fetchRunsList(t, server)
	matched := findRunInList(t, payload.Items, run.ID)
	if len(matched.Failures) != 0 {
		t.Fatalf("expected no failures in list payload, got %d", len(matched.Failures))
	}
}

func fetchRunsList(t *testing.T, server *Server) struct {
	Items []domain.Run `json:"items"`
} {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/runs", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var payload struct {
		Items []domain.Run `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	return payload
}

func findRunInList(t *testing.T, items []domain.Run, runID string) *domain.Run {
	t.Helper()
	for i := range items {
		if items[i].ID == runID {
			return &items[i]
		}
	}
	t.Fatalf("expected run %q in list response", runID)
	return nil
}
