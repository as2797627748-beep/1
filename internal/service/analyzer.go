package service

import (
	"sort"
	"strings"

	"autocode-platform/internal/domain"
)

type AnalyzeInput struct {
	Title string `json:"title"`
	Goal  string `json:"goal"`
}

func toolCatalog() []domain.ToolProfile {
	return []domain.ToolProfile{
		{ID: "workspace", Category: "core", Name: "workspace", Description: "读写工作区文件与生成补丁", Heavy: false},
		{ID: "terminal", Category: "core", Name: "terminal", Description: "执行受控命令与构建流程", Heavy: false},
		{ID: "research", Category: "knowledge", Name: "research", Description: "查阅资料、整理结论与生成摘要", Heavy: false},
		{ID: "office", Category: "productivity", Name: "office", Description: "纪要、表格、提纲和办公材料编排", Heavy: false},
		{ID: "assets", Category: "creative", Name: "assets", Description: "游戏、应用和内容资源规划与整理", Heavy: false},
		{ID: "tests", Category: "quality", Name: "tests", Description: "运行单元、集成与端到端测试", Heavy: false},
		{ID: "lint", Category: "quality", Name: "lint", Description: "代码风格与静态检查", Heavy: false},
		{ID: "format", Category: "quality", Name: "format", Description: "格式化与结构清理", Heavy: false},
		{ID: "build", Category: "release", Name: "build", Description: "生成发布产物和校验构建", Heavy: true},
		{ID: "deploy", Category: "release", Name: "deploy", Description: "触发稳定部署", Heavy: true},
		{ID: "logs", Category: "ops", Name: "logs", Description: "收集运行与部署日志", Heavy: false},
		{ID: "analysis", Category: "planner", Name: "analysis", Description: "将一句话需求拆解为流程与风险", Heavy: false},
	}
}

func atomicToolCatalog(profile domain.SystemProfile) []domain.AtomicToolProfile {
	compact := profile.Tier == "compact"
	tools := []domain.AtomicToolProfile{
		{
			ID:                "tool-intent-analysis",
			Category:          "planner",
			Name:              "任务识别与预演",
			Summary:           "把一句话目标拆成阶段、能力和验证建议，作为所有编排入口的起点。",
			StageKinds:        []domain.StageKind{domain.StageIntent, domain.StageContext, domain.StagePlan},
			LoadTier:          "light",
			Priority:          10,
			LocalFirst:        true,
			CoverageTags:      []string{"任务识别", "流程预演", "风险收束"},
			PreferredProvider: "analysis",
			FallbackProviders: []string{"logs"},
			ActivationMode:    "builtin",
			DedupGroup:        "analysis",
			Recommended:       true,
			Allowed:           true,
			Reason:            "所有任务都需要先完成意图识别和风险收束。",
		},
		{
			ID:                "tool-workspace-patch",
			Category:          "code",
			Name:              "工作区改写",
			Summary:           "围绕文件读取、补丁写入和结构改造完成核心产出。",
			StageKinds:        []domain.StageKind{domain.StageImplement, domain.StageResult, domain.StageRepair},
			LoadTier:          "light",
			Priority:          10,
			LocalFirst:        true,
			CoverageTags:      []string{"代码改写", "配置改写", "文档改写"},
			PreferredProvider: "workspace",
			FallbackProviders: []string{"terminal"},
			ActivationMode:    "builtin",
			DedupGroup:        "workspace-edit",
			Recommended:       true,
			Allowed:           true,
			Reason:            "这是代码、配置和文档改写的主入口。",
		},
		{
			ID:                "tool-structure-refine",
			Category:          "code",
			Name:              "结构整理",
			Summary:           "用于格式化、重排、瘦身和结构层面的轻量优化，避免和核心改写入口重复。",
			StageKinds:        []domain.StageKind{domain.StageResult, domain.StageRepair, domain.StageFinalize},
			LoadTier:          "light",
			Priority:          20,
			LocalFirst:        true,
			CoverageTags:      []string{"格式整理", "结构清理", "轻量重构"},
			PreferredProvider: "format",
			FallbackProviders: []string{"workspace"},
			ActivationMode:    "builtin",
			DedupGroup:        "workspace-edit",
			Recommended:       true,
			Allowed:           true,
			Reason:            "格式化和结构整理应统一放在一个轻量入口，而不是重复分散在多个动作中。",
		},
		{
			ID:                "tool-command-exec",
			Category:          "ops",
			Name:              "受控命令执行",
			Summary:           "用于构建、检查和有限的自动化命令编排。",
			StageKinds:        []domain.StageKind{domain.StageTool, domain.StageImplement, domain.StageTest, domain.StageDeploy, domain.StageRepair},
			LoadTier:          ternaryLoad(compact, "medium", "light"),
			Priority:          10,
			LocalFirst:        true,
			CoverageTags:      []string{"构建命令", "运行命令", "诊断命令"},
			PreferredProvider: "terminal",
			FallbackProviders: []string{"logs"},
			ActivationMode:    "builtin",
			DedupGroup:        "command",
			Recommended:       true,
			Allowed:           true,
			Reason:            "命令执行能力是构建、验证和排错的基础支撑。",
		},
		{
			ID:                "tool-build-artifact",
			Category:          "release",
			Name:              "构建产物",
			Summary:           "集中处理发布前构建、打包和产物检查，作为交付前的唯一重构建入口。",
			StageKinds:        []domain.StageKind{domain.StageResult, domain.StageTest, domain.StageDeploy},
			LoadTier:          "heavy",
			Priority:          10,
			LocalFirst:        true,
			CoverageTags:      []string{"构建产物", "打包产物", "发布准备"},
			PreferredProvider: "build",
			FallbackProviders: []string{"terminal"},
			ActivationMode:    ternaryActivation(compact),
			DedupGroup:        "build",
			Recommended:       !compact,
			Allowed:           !compact,
			Reason:            buildToolReason(compact),
		},
		{
			ID:                "tool-quality-validation",
			Category:          "quality",
			Name:              "质量验证链",
			Summary:           "统一承接静态检查、单测、集成和冒烟验证。",
			StageKinds:        []domain.StageKind{domain.StageTest, domain.StageRepair},
			LoadTier:          "medium",
			Priority:          10,
			LocalFirst:        true,
			CoverageTags:      []string{"静态检查", "单元测试", "集成测试", "冒烟验证"},
			PreferredProvider: "tests",
			FallbackProviders: []string{"lint", "format"},
			ActivationMode:    "builtin",
			DedupGroup:        "validation",
			Recommended:       true,
			Allowed:           true,
			Reason:            "所有可交付结果都必须经过统一验证链。",
		},
		{
			ID:                "tool-quality-review",
			Category:          "quality",
			Name:              "质量复核",
			Summary:           "用于规范检查、结构复核和非代码任务的交付校对，避免重复调用多个审查入口。",
			StageKinds:        []domain.StageKind{domain.StagePlan, domain.StageTest, domain.StageRepair, domain.StageFinalize},
			LoadTier:          "light",
			Priority:          20,
			LocalFirst:        true,
			CoverageTags:      []string{"规范复核", "结构复核", "交付校对"},
			PreferredProvider: "lint",
			FallbackProviders: []string{"tests", "format"},
			ActivationMode:    "builtin",
			DedupGroup:        "validation",
			Recommended:       true,
			Allowed:           true,
			Reason:            "非代码任务也需要统一复核入口，而不是单纯依赖测试链。",
		},
		{
			ID:                "tool-release-delivery",
			Category:          "release",
			Name:              "稳定部署",
			Summary:           "负责部署产物生成、稳定部署和版本记录沉淀。",
			StageKinds:        []domain.StageKind{domain.StageDeploy},
			LoadTier:          "heavy",
			Priority:          10,
			LocalFirst:        true,
			CoverageTags:      []string{"稳定部署", "版本记录", "部署产物"},
			PreferredProvider: "deploy",
			FallbackProviders: []string{"build", "logs"},
			ActivationMode:    "optional",
			DedupGroup:        "delivery",
			Recommended:       !compact,
			Allowed:           !compact,
			Reason:            deliveryToolReason(compact),
		},
		{
			ID:                "tool-observability",
			Category:          "ops",
			Name:              "运行观测",
			Summary:           "统一承接日志、阶段动态、故障线索和部署证据，不再分散到多个观测入口。",
			StageKinds:        []domain.StageKind{domain.StageContext, domain.StageTest, domain.StageDeploy, domain.StageRepair, domain.StageFinalize},
			LoadTier:          "light",
			Priority:          10,
			LocalFirst:        true,
			CoverageTags:      []string{"日志观测", "故障线索", "运行证据"},
			PreferredProvider: "logs",
			FallbackProviders: []string{"terminal"},
			ActivationMode:    "builtin",
			DedupGroup:        "observability",
			Recommended:       true,
			Allowed:           true,
			Reason:            "观测能力必须统一，否则故障上下文会被切碎。",
		},
		{
			ID:                "tool-research-brief",
			Category:          "research",
			Name:              "研究整理",
			Summary:           "负责检索、摘要、对比和知识沉淀，覆盖非代码任务。",
			StageKinds:        []domain.StageKind{domain.StageContext, domain.StagePlan, domain.StageImplement, domain.StageResult, domain.StageFinalize},
			LoadTier:          "light",
			Priority:          10,
			LocalFirst:        true,
			CoverageTags:      []string{"检索摘要", "资料对比", "知识沉淀"},
			PreferredProvider: "research",
			FallbackProviders: []string{"office"},
			ActivationMode:    "builtin",
			DedupGroup:        "research",
			Recommended:       true,
			Allowed:           true,
			Reason:            "研究和知识整理是统一自动化中枢的常见能力。",
		},
		{
			ID:                "tool-doc-governance",
			Category:          "knowledge",
			Name:              "文档治理",
			Summary:           "统一承接 README、规范、Wiki 和说明文档的生成与更新。",
			StageKinds:        []domain.StageKind{domain.StagePlan, domain.StageImplement, domain.StageResult, domain.StageRepair, domain.StageFinalize},
			LoadTier:          "light",
			Priority:          10,
			LocalFirst:        true,
			CoverageTags:      []string{"README", "Wiki", "规范文档"},
			PreferredProvider: "workspace",
			FallbackProviders: []string{"research", "office"},
			ActivationMode:    "builtin",
			DedupGroup:        "documentation",
			Recommended:       true,
			Allowed:           true,
			Reason:            "文档治理应独立成一类能力，但仍复用本地工作区作为主入口。",
		},
		{
			ID:                "tool-office-compose",
			Category:          "office",
			Name:              "办公整理",
			Summary:           "用于纪要、提纲、表格说明和结构化办公内容输出。",
			StageKinds:        []domain.StageKind{domain.StagePlan, domain.StageImplement, domain.StageResult, domain.StageFinalize},
			LoadTier:          "light",
			Priority:          10,
			LocalFirst:        true,
			CoverageTags:      []string{"纪要整理", "提纲输出", "汇报材料"},
			PreferredProvider: "office",
			FallbackProviders: []string{"research"},
			ActivationMode:    "builtin",
			DedupGroup:        "office",
			Recommended:       true,
			Allowed:           true,
			Reason:            "办公内容不该被拆成另一套孤立流程。",
		},
		{
			ID:                "tool-daily-automation",
			Category:          "daily",
			Name:              "日常编排",
			Summary:           "处理提醒、待办、轻量流程和个人助理类事务，避免另起一套小工具。",
			StageKinds:        []domain.StageKind{domain.StageIntent, domain.StagePlan, domain.StageImplement, domain.StageResult, domain.StageFinalize},
			LoadTier:          "light",
			Priority:          10,
			LocalFirst:        true,
			CoverageTags:      []string{"待办整理", "提醒流程", "日常事务"},
			PreferredProvider: "office",
			FallbackProviders: []string{"analysis", "research"},
			ActivationMode:    "builtin",
			DedupGroup:        "daily-assist",
			Recommended:       true,
			Allowed:           true,
			Reason:            "日常事务属于高频轻量能力，应直接纳入统一编排中枢。",
		},
		{
			ID:                "tool-asset-planning",
			Category:          "assets",
			Name:              "内容资产规划",
			Summary:           "覆盖界面资源、游戏素材、应用内容资产和交付素材组织。",
			StageKinds:        []domain.StageKind{domain.StageContext, domain.StagePlan, domain.StageImplement, domain.StageResult},
			LoadTier:          "light",
			Priority:          10,
			LocalFirst:        true,
			CoverageTags:      []string{"界面资源", "游戏素材", "内容资产"},
			PreferredProvider: "assets",
			FallbackProviders: []string{"workspace", "research"},
			ActivationMode:    "builtin",
			DedupGroup:        "asset-design",
			Recommended:       true,
			Allowed:           true,
			Reason:            "内容资产能力应该有统一入口，避免按项目类型重复造轮子。",
		},
		{
			ID:                "tool-media-production",
			Category:          "media",
			Name:              "多媒体编排",
			Summary:           "统一承接图片、视频、交互动效和媒体产出规划，默认只做中枢编排与本地优先组织。",
			StageKinds:        []domain.StageKind{domain.StageContext, domain.StagePlan, domain.StageImplement},
			LoadTier:          ternaryLoad(compact, "medium", "light"),
			Priority:          20,
			LocalFirst:        true,
			CoverageTags:      []string{"图片产出", "视频规划", "动效编排"},
			PreferredProvider: "assets",
			FallbackProviders: []string{"research", "office"},
			ActivationMode:    "optional",
			DedupGroup:        "media-production",
			Recommended:       !compact,
			Allowed:           true,
			Reason:            "媒体能力要覆盖，但默认以本地组织和按需接入为主，避免把平台拖成重工作站。",
		},
	}
	sort.Slice(tools, func(i, j int) bool { return tools[i].ID < tools[j].ID })
	return tools
}

func ternaryLoad(compact bool, compactValue, defaultValue string) string {
	if compact {
		return compactValue
	}
	return defaultValue
}

func ternaryActivation(compact bool) string {
	if compact {
		return "optional"
	}
	return "builtin"
}

func buildToolReason(compact bool) string {
	if compact {
		return "构建能力必须保留，但在低配机器上应作为按需重动作，而不是默认常驻入口。"
	}
	return "构建能力属于少数必须保留的重入口之一，应作为统一构建产物通道被集中治理。"
}

func deliveryToolReason(compact bool) string {
	if compact {
		return "当前机器不适合默认开启稳定部署，建议保持候选位。"
	}
	return "当前机器可按需执行交付，但应控制并发和重动作数量。"
}

func testCatalog() []domain.TestProfile {
	return []domain.TestProfile{
		{ID: "static-check", Layer: "quality", Name: "static-check", Description: "语法、类型和结构静态检查"},
		{ID: "lint", Layer: "quality", Name: "lint", Description: "代码规范与可维护性检查"},
		{ID: "review-check", Layer: "quality", Name: "review-check", Description: "对文档、方案和输出质量进行结构化复核"},
		{ID: "code-review", Layer: "quality", Name: "code-review", Description: "AI 代码审查，检查逻辑漏洞、规范偏差与性能风险"},
		{ID: "security-scan", Layer: "security", Name: "security-scan", Description: "检测敏感信息、依赖漏洞和常见安全风险"},
		{ID: "unit", Layer: "unit", Name: "unit-tests", Description: "最小逻辑单元测试"},
		{ID: "integration", Layer: "integration", Name: "integration-tests", Description: "模块与接口联调测试"},
		{ID: "e2e", Layer: "e2e", Name: "e2e-smoke", Description: "关键页面或关键流程冒烟验证"},
		{ID: "deploy-smoke", Layer: "ops", Name: "deploy-smoke", Description: "部署后健康检查与启动验证"},
		{ID: "consistency-check", Layer: "ops", Name: "consistency-check", Description: "检查配置、素材和交付内容的一致性"},
	}
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
	return out
}

func AnalyzeGoal(input AnalyzeInput) domain.AnalysisPlan {
	text := strings.ToLower(strings.TrimSpace(input.Title + "\n" + input.Goal))
	projectKind := "general"
	intent := "build"
	if hasAny(text, "网站", "web", "前端", "ui", "界面", "dashboard") {
		projectKind = "web"
	}
	if hasAny(text, "api", "后端", "服务", "server") {
		if projectKind == "general" {
			projectKind = "backend"
		}
	}
	if hasAny(text, "游戏", "game", "玩法", "关卡") {
		projectKind = "game"
	}
	if hasAny(text, "app", "安卓", "ios", "移动端", "客户端") {
		projectKind = "app"
	}
	if hasAny(text, "软件", "desktop", "客户端工具", "工具软件") {
		projectKind = "software"
	}
	if hasAny(text, "办公", "表格", "纪要", "汇报", "邮件") {
		projectKind = "office"
	}
	if hasAny(text, "研究", "调研", "查阅", "资料", "总结", "报告") {
		projectKind = "research"
	}
	if hasAny(text, "日常", "待办", "计划", "行程", "提醒") {
		projectKind = "daily"
	}
	if hasAny(text, "文档", "内容", "wiki", "readme", "文章") {
		projectKind = "docs-content"
	}
	if hasAny(text, "修复", "改错", "排错", "bug", "故障") {
		intent = "fix"
	}
	if hasAny(text, "整理", "总结", "归纳", "梳理") {
		intent = "organize"
	}
	if hasAny(text, "部署", "上线", "vps", "发布") {
		if intent == "build" {
			intent = "build-deploy"
		}
	}

	stages := []domain.StageKind{
		domain.StageIntent,
		domain.StageContext,
		domain.StagePlan,
		domain.StageResource,
		domain.StageModel,
		domain.StageTool,
		domain.StageImplement,
		domain.StageResult,
		domain.StageTest,
	}
	if hasAny(text, "部署", "上线", "vps", "发布") {
		stages = append(stages, domain.StageDeploy)
	}
	if intent == "fix" || hasAny(text, "自动修复", "排错", "闭环", "全自动") {
		stages = append(stages, domain.StageRepair)
	}
	stages = append(stages, domain.StageFinalize)
	stages = dedupeStages(stages)

	tools := []string{"analysis", "workspace", "terminal", "tests", "logs"}
	if hasAny(text, "格式", "整洁", "规范") {
		tools = append(tools, "format", "lint")
	}
	if hasAny(text, "构建", "打包", "上线", "部署", "vps") {
		tools = append(tools, "build", "deploy")
	}
	if hasAny(text, "测试", "验证", "闭环", "质量") {
		tools = append(tools, "lint")
	}
	if hasAny(text, "研究", "调研", "查阅", "资料", "总结", "报告") {
		tools = append(tools, "research")
	}
	if hasAny(text, "办公", "表格", "纪要", "汇报", "邮件", "日常") {
		tools = append(tools, "office")
	}
	if hasAny(text, "游戏", "素材", "界面", "app", "软件") {
		tools = append(tools, "assets")
	}

	tests := []string{"static-check", "review-check"}
	if projectKind == "web" || projectKind == "backend" || projectKind == "app" || projectKind == "software" || projectKind == "game" {
		tests = append(tests, "unit-tests", "integration-tests", "e2e-smoke", "code-review", "security-scan")
	} else {
		tests = append(tests, "consistency-check")
	}
	if hasAny(text, "部署", "vps", "上线") {
		tests = append(tests, "deploy-smoke")
	}
	if hasAny(text, "规范", "整洁") {
		tests = append(tests, "lint")
	}
	if hasAny(text, "登录", "权限", "认证", "token", "密钥", "安全", "审查", "review") {
		tests = append(tests, "code-review", "security-scan")
	}

	risks := []string{"需求可能在执行中途变化，需允许暂停补需", "1C1G VPS 需要控制并发与构建资源占用"}
	if hasAny(text, "部署", "vps") {
		risks = append(risks, "远程部署需保证可重复执行，并避免 SSH 中断导致半成品状态")
	}
	if hasAny(text, "多模型", "全平台", "provider") {
		risks = append(risks, "多平台模型接入需要处理配置缺失和能力差异")
	}
	if projectKind == "office" || projectKind == "research" || projectKind == "daily" || projectKind == "docs-content" {
		risks = append(risks, "非代码任务同样需要明确交付格式、引用边界和复核标准")
	}
	if hasAny(text, "登录", "权限", "认证", "token", "密钥", "安全") {
		risks = append(risks, "认证、权限和密钥配置需要纳入安全扫描与代码审查检查点")
	}
	if hasAny(text, "npm", "package.json", "依赖", "vue", "react", "go", "第三方") {
		risks = append(risks, "依赖安装与升级阶段需要同步检查第三方依赖风险和锁文件一致性")
	}

	requirementSpec := buildRequirementSpec(input, text, projectKind)
	taskQueue := buildTaskQueue(intent, projectKind, stages, tests)
	checkpoints := buildFlowCheckpoints(projectKind, intent, tests)

	return domain.AnalysisPlan{
		Intent:            intent,
		ProjectKind:       projectKind,
		Summary:           buildSummary(intent, projectKind, stages, tests),
		RequirementSpec:   requirementSpec,
		TaskQueue:         taskQueue,
		Checkpoints:       checkpoints,
		LoopBlueprint:     buildLoopBlueprint(intent, projectKind),
		RecommendedStages: stages,
		RecommendedTools:  dedupeStrings(tools),
		RecommendedTests:  dedupeStrings(tests),
		RecommendedBundle: recommendedBundle(intent, projectKind),
		Risks:             dedupeStrings(risks),
	}
}

func recommendedBundle(intent, projectKind string) domain.BundleSuggestion {
	switch projectKind {
	case "research", "office", "daily", "docs-content":
		return domain.BundleSuggestion{
			ID:     "bundle-knowledge-office",
			Name:   "知识与办公",
			Reason: "当前任务以研究、文档或办公交付为主，优先收束到研究整理、知识沉淀和办公协作能力。",
		}
	case "game":
		return domain.BundleSuggestion{
			ID:     "bundle-creative-assets",
			Name:   "内容与资产",
			Reason: "当前任务包含游戏或素材导向内容，适合优先组织资源规划、多媒体和内容资产链路。",
		}
	}
	if intent == "build-deploy" {
		return domain.BundleSuggestion{
			ID:     "bundle-local-delivery",
			Name:   "本地交付闭环",
			Reason: "当前任务明确包含发布或上线目标，适合直接装配构建、验证、交付和运维链路。",
		}
	}
	return domain.BundleSuggestion{
		ID:     "bundle-local-core",
		Name:   "本地核心闭环",
		Reason: "当前任务以分析、改写和验证为主，先装配本地优先的核心闭环可以减少额外重动作。",
	}
}

func buildLoopBlueprint(intent, projectKind string) []domain.LoopStep {
	steps := []domain.LoopStep{
		{ID: "intent-intake", Name: "任务识别", Summary: "先确认这是构建、修复、整理还是交付导向任务。"},
		{ID: "context-collection", Name: "上下文汇总", Summary: "收集目标、限制、现有结构和已知风险，避免盲目执行。"},
		{ID: "plan-formation", Name: "方案规划", Summary: "基于任务类型生成适合当前场景的执行路径与优先级。"},
		{ID: "resource-shaping", Name: "资源收束", Summary: "根据机器规格和运行策略收束并发、工具和负载等级。"},
		{ID: "model-routing", Name: "模型选择", Summary: "为当前任务挑选最适合的模型入口和回退方案。"},
		{ID: "tool-assembly", Name: "工具装配", Summary: "装配本次任务真正需要的原子能力，避免重复和空转。"},
		{ID: "task-execution", Name: "执行生成", Summary: "执行代码、文档、配置或研究产出，并持续记录阶段摘要。"},
		{ID: "result-structuring", Name: "结果整理", Summary: "把输出整理为可复核、可交付、可追踪的结构化结果。"},
		{ID: "validation", Name: "验证检查", Summary: validationSummary(projectKind)},
		{ID: "delivery", Name: "部署交付", Summary: deliverySummary(intent)},
		{ID: "repair-loop", Name: "修复回路", Summary: "当验证或交付失败时，自动生成失败摘要并进入修复链路。"},
		{ID: "knowledge-sync", Name: "总结沉淀", Summary: "沉淀关键决策、经验和后续建议，形成长期可复用上下文。"},
	}
	return steps
}

func validationSummary(projectKind string) string {
	switch projectKind {
	case "web", "backend", "app", "software", "game":
		return "执行静态检查、单元测试、集成测试和关键流程验证，确认结果可用。"
	default:
		return "执行结构复核、一致性检查和关键交付核验，确保输出可信。"
	}
}

func deliverySummary(intent string) string {
	if intent == "build-deploy" {
		return "在需要时进入稳定部署或交付整理，保留可重试、可回退的证据。"
	}
	return "按任务目标决定是否进入发布、移交或结果投递，不强制执行部署。"
}

func buildSummary(intent, projectKind string, stages []domain.StageKind, tests []string) string {
	stageNames := make([]string, 0, len(stages))
	for _, stage := range stages {
		stageNames = append(stageNames, stageSummaryLabel(stage))
	}
	testNames := make([]string, 0, len(tests))
	for _, test := range dedupeStrings(tests) {
		testNames = append(testNames, testSummaryLabel(test))
	}
	return "已根据一句话需求整理出" + intentSummaryLabel(intent) + "的" + projectKindSummaryLabel(projectKind) + "方案，覆盖阶段 " + strings.Join(stageNames, "、") + "，并补齐 " + strings.Join(testNames, "、") + " 等验证环节。"
}

func intentSummaryLabel(intent string) string {
	switch intent {
	case "fix":
		return "修复导向"
	case "organize":
		return "整理导向"
	case "build-deploy":
		return "构建与交付导向"
	default:
		return "构建导向"
	}
}

func projectKindSummaryLabel(projectKind string) string {
	switch projectKind {
	case "web":
		return "Web 项目"
	case "backend":
		return "后端服务"
	case "game":
		return "游戏项目"
	case "app":
		return "应用项目"
	case "software":
		return "软件工程"
	case "office":
		return "办公任务"
	case "research":
		return "研究任务"
	case "daily":
		return "日常事务"
	case "docs-content":
		return "文档内容"
	default:
		return "通用任务"
	}
}

func stageSummaryLabel(stage domain.StageKind) string {
	switch stage {
	case domain.StageIntent:
		return "任务识别"
	case domain.StageContext:
		return "上下文汇总"
	case domain.StagePlan:
		return "方案规划"
	case domain.StageResource:
		return "资源收束"
	case domain.StageModel:
		return "模型选择"
	case domain.StageTool:
		return "工具装配"
	case domain.StageImplement:
		return "执行生成"
	case domain.StageResult:
		return "结果整理"
	case domain.StageTest:
		return "验证检查"
	case domain.StageDeploy:
		return "部署交付"
	case domain.StageRepair:
		return "修复回路"
	case domain.StageFinalize:
		return "总结沉淀"
	default:
		return string(stage)
	}
}

func testSummaryLabel(test string) string {
	switch test {
	case "static-check":
		return "静态检查"
	case "review-check":
		return "结构复核"
	case "code-review":
		return "代码审查"
	case "security-scan":
		return "安全扫描"
	case "unit-tests":
		return "单元测试"
	case "integration-tests":
		return "集成测试"
	case "e2e-smoke":
		return "端到端验证"
	case "deploy-smoke":
		return "交付后检查"
	case "consistency-check":
		return "一致性检查"
	case "lint":
		return "规范检查"
	default:
		return test
	}
}

func buildRequirementSpec(input AnalyzeInput, text, projectKind string) domain.RequirementSpec {
	functional := []string{"支持自然语言驱动的任务初始化与阶段拆解"}
	techStack := []string{}
	nonFunctional := []string{"保持本地优先与轻量运行策略", "提供分阶段验证与状态可追踪能力"}
	constraints := []string{"工作目录统一收束在 /workspace", "需要兼顾 1C1G VPS 资源约束"}
	style := []string{}

	if hasAny(text, "登录", "注册", "权限", "认证") {
		functional = append(functional, "实现用户认证、权限控制与安全校验链路")
		nonFunctional = append(nonFunctional, "认证相关流程必须经过安全扫描与代码审查")
	}
	if hasAny(text, "后台", "管理系统", "dashboard", "admin") {
		functional = append(functional, "提供后台管理界面与核心业务操作入口")
	}
	if hasAny(text, "预览", "运行", "dev") {
		functional = append(functional, "构建后自动启动服务并提供预览验证入口")
	}
	if hasAny(text, "go", "golang") {
		techStack = append(techStack, "Go")
	}
	if hasAny(text, "vue3", "vue", "vite") {
		techStack = append(techStack, "Vue 3")
	}
	if hasAny(text, "react", "next") {
		techStack = append(techStack, "React")
	}
	if hasAny(text, "node", "npm", "pnpm") {
		techStack = append(techStack, "Node.js")
	}
	if hasAny(text, "docker", "容器") {
		constraints = append(constraints, "开发环境需要支持容器化隔离与端口映射")
	}
	if hasAny(text, "风格", "设计", "居中", "主题", "界面") {
		style = append(style, "界面风格与交互细节需要纳入任务约束")
	}
	if projectKind == "web" || projectKind == "app" {
		functional = append(functional, "前后端联调后需要通过预览面板验证可见效果")
	}
	if len(techStack) == 0 {
		techStack = append(techStack, "沿用现有项目技术栈")
	}

	return domain.RequirementSpec{
		Summary:                strings.TrimSpace(input.Title + " " + input.Goal),
		FunctionalRequirements: dedupeStrings(functional),
		TechStack:              dedupeStrings(techStack),
		NonFunctional:          dedupeStrings(nonFunctional),
		Constraints:            dedupeStrings(constraints),
		StylePreferences:       dedupeStrings(style),
	}
}

func buildTaskQueue(intent, projectKind string, stages []domain.StageKind, tests []string) []domain.TaskQueueItem {
	queue := []domain.TaskQueueItem{
		{ID: "task-intent-intake", Title: "收束任务目标", Summary: "确认任务意图、交付方式和关键边界。", Agent: "需求解析Agent", Phase: "intent", Priority: 10, Status: "ready"},
		{ID: "task-context-collection", Title: "补齐上下文", Summary: "整理现有结构、依赖条件和已知限制，避免盲目推进。", Agent: "需求解析Agent", Phase: "context", Priority: 20, Status: "ready"},
		{ID: "task-plan-formation", Title: "生成执行规划", Summary: "产出阶段顺序、风险收口和验证安排。", Agent: "需求解析Agent", Phase: "plan", Priority: 30, Status: "ready"},
	}
	if hasStage(stages, domain.StageResource) {
		queue = append(queue, domain.TaskQueueItem{ID: "task-resource-shaping", Title: "收束资源策略", Summary: "根据机器规格决定并发、重动作和交付范围。", Agent: "调度策略Agent", Phase: "resource", Priority: 35, Status: "ready"})
	}
	if hasStage(stages, domain.StageModel) {
		queue = append(queue, domain.TaskQueueItem{ID: "task-model-routing", Title: "选择模型路径", Summary: "选择最适合当前任务的模型入口与回退路线。", Agent: "模型治理Agent", Phase: "model", Priority: 38, Status: "ready"})
	}
	if hasStage(stages, domain.StageTool) {
		queue = append(queue, domain.TaskQueueItem{ID: "task-tool-assembly", Title: "装配原子能力", Summary: "把真正需要的原子能力装配到本次运行单。", Agent: "工具治理Agent", Phase: "tool", Priority: 39, Status: "ready"})
	}
	if hasStage(stages, domain.StageImplement) {
		queue = append(queue, domain.TaskQueueItem{ID: "task-code-implement", Title: "执行核心产出", Summary: "按任务队列逐步完成代码、文档、配置或研究内容生成。", Agent: "代码生成Agent", Phase: "implement", Priority: 40, Status: "ready"})
	}
	if hasStage(stages, domain.StageResult) {
		queue = append(queue, domain.TaskQueueItem{ID: "task-result-structuring", Title: "整理交付结果", Summary: "把产出整理成可复核、可追踪、可交付的结构化结果。", Agent: "结果整理Agent", Phase: "result", Priority: 50, Status: "ready"})
	}
	if hasStage(stages, domain.StageTest) {
		queue = append(queue, domain.TaskQueueItem{ID: "task-test-validate", Title: "执行验证检查", Summary: "执行单测、集成测试、代码审查和安全扫描。", Agent: "测试验证Agent", Phase: "test", Priority: 60, Status: "ready"})
	}
	if hasStage(stages, domain.StageDeploy) {
		queue = append(queue, domain.TaskQueueItem{ID: "task-preview-delivery", Title: "执行部署交付", Summary: "构建产物、启动服务并生成预览或部署配置。", Agent: "终端执行Agent", Phase: "deploy", Priority: 70, Status: "ready"})
	}
	if hasStage(stages, domain.StageRepair) {
		queue = append(queue, domain.TaskQueueItem{ID: "task-repair-loop", Title: "进入修复回路", Summary: "结合失败摘要、日志和验证结果做定向修复。", Agent: "错误修复Agent", Phase: "repair", Priority: 75, Status: "ready"})
	}
	if hasStage(stages, domain.StageFinalize) {
		queue = append(queue, domain.TaskQueueItem{ID: "task-knowledge-sync", Title: "沉淀结果与建议", Summary: "总结关键决策、证据和下一步建议，形成长期上下文。", Agent: "总结沉淀Agent", Phase: "finalize", Priority: 80, Status: "ready"})
	}
	if projectKind == "web" || projectKind == "backend" || projectKind == "app" || projectKind == "software" || projectKind == "game" {
		queue = append(queue, domain.TaskQueueItem{ID: "task-review-security", Title: "代码审查与安全扫描", Summary: "把代码审查与安全扫描作为标准质量闸门。", Agent: "测试验证Agent", Phase: "quality", Priority: 65, Status: "ready"})
	}
	_ = tests
	return queue
}

func buildFlowCheckpoints(projectKind, intent string, tests []string) []domain.FlowCheckpoint {
	items := []domain.FlowCheckpoint{
		{ID: "checkpoint-requirement", Title: "需求理解完成", Summary: "任务识别与上下文汇总已经收束，目标和边界可追踪。", Gate: "需求确认后方可进入规划与装配", Status: "pending"},
		{ID: "checkpoint-design", Title: "规划装配完成", Summary: "方案规划、资源收束、模型选择与工具装配已经就绪。", Gate: "规划完成后方可进入执行生成", Status: "pending"},
		{ID: "checkpoint-implementation", Title: "执行结果完成", Summary: "核心产出与结果整理已经完成，可进入验证链路。", Gate: "结果整理完成后进入验证", Status: "pending"},
		{ID: "checkpoint-validation", Title: "验证通过", Summary: "测试、代码审查和安全扫描通过后才允许构建交付。", Gate: "验证通过后方可构建或上线", Status: "pending"},
	}
	if intent == "build-deploy" {
		items = append(items, domain.FlowCheckpoint{ID: "checkpoint-delivery", Title: "交付准备完成", Summary: "构建产物、预览或部署配置准备完毕。", Gate: "交付证据齐全后方可上线", Status: "pending"})
	}
	if projectKind == "web" || projectKind == "app" {
		items = append(items, domain.FlowCheckpoint{ID: "checkpoint-preview", Title: "预览验证完成", Summary: "预览面板需要看到最新页面效果并完成关键流程检查。", Gate: "预览通过后才进入交付", Status: "pending"})
	}
	if containsString(tests, "security-scan") {
		items = append(items, domain.FlowCheckpoint{ID: "checkpoint-security", Title: "安全扫描通过", Summary: "敏感信息、依赖漏洞和常见风险检查通过。", Gate: "安全扫描通过后方可交付", Status: "pending"})
	}
	return items
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if strings.TrimSpace(strings.ToLower(value)) == strings.TrimSpace(strings.ToLower(target)) {
			return true
		}
	}
	return false
}

func hasStage(stages []domain.StageKind, target domain.StageKind) bool {
	for _, stage := range stages {
		if stage == target {
			return true
		}
	}
	return false
}

func hasAny(text string, keywords ...string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func dedupeStages(stages []domain.StageKind) []domain.StageKind {
	seen := map[domain.StageKind]bool{}
	out := make([]domain.StageKind, 0, len(stages))
	for _, stage := range stages {
		if seen[stage] {
			continue
		}
		seen[stage] = true
		out = append(out, stage)
	}
	return out
}

func sortedToolCatalog() []domain.ToolProfile {
	out := append([]domain.ToolProfile(nil), toolCatalog()...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func sortedTestCatalog() []domain.TestProfile {
	out := append([]domain.TestProfile(nil), testCatalog()...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
