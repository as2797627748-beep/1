package service

import (
	"autocode-platform/internal/domain"
	"strconv"
	"strings"
)

func workflowOptionGroups() []domain.WorkflowOptionGroup {
	return []domain.WorkflowOptionGroup{
		{
			ID:          "group-intake",
			Name:        "需求入口",
			Description: "适合新手的一句话入口与任务拆解",
			Options: []domain.WorkflowOption{
				{ID: "one-shot-analysis", Kind: "stage", Name: "一句话自动分析", Description: "自动识别任务类型、流程、工具和测试建议", DefaultOn: true},
				{ID: "pause-and-refine", Kind: "stage", Name: "暂停补充需求", Description: "运行中可暂停并追加约束、目标和补充说明", DefaultOn: true},
			},
		},
		{
			ID:          "group-production",
			Name:        "内容生产",
			Description: "不只覆盖代码，也覆盖文档、配置和知识内容",
			Options: []domain.WorkflowOption{
				{ID: "code-flow", Kind: "task-kind", Name: "代码写改测", Description: "生成、修改、测试和修复代码闭环", DefaultOn: true},
				{ID: "docs-flow", Kind: "task-kind", Name: "文档生成与更新", Description: "编写 README、Wiki、规范和项目说明", DefaultOn: true},
				{ID: "ops-flow", Kind: "task-kind", Name: "运维与发布任务", Description: "发布、巡检、迁移建议和运行维护", DefaultOn: true},
				{ID: "config-flow", Kind: "task-kind", Name: "配置与环境整理", Description: "配置优化、环境建议和启动项调整", DefaultOn: true},
			},
		},
		{
			ID:          "group-product",
			Name:        "产品与应用",
			Description: "覆盖游戏、软件、App 与交互产品的完整编排",
			Options: []domain.WorkflowOption{
				{ID: "game-flow", Kind: "task-kind", Name: "游戏创作与改造", Description: "玩法、界面、资源组织与版本验证", DefaultOn: true},
				{ID: "app-flow", Kind: "task-kind", Name: "App 与移动体验", Description: "客户端能力、移动体验与发布准备", DefaultOn: true},
				{ID: "software-flow", Kind: "task-kind", Name: "桌面软件与工具", Description: "软件功能、打包、安装与升级建议", DefaultOn: true},
			},
		},
		{
			ID:          "group-office",
			Name:        "办公与研究",
			Description: "让平台适合查阅、整理、写作与日常事务",
			Options: []domain.WorkflowOption{
				{ID: "office-flow", Kind: "task-kind", Name: "办公协同整理", Description: "表格、纪要、汇报、流程整理与自动补全", DefaultOn: true},
				{ID: "research-flow", Kind: "task-kind", Name: "查阅研究与摘要", Description: "资料梳理、结论提炼、风险对比与知识沉淀", DefaultOn: true},
				{ID: "daily-flow", Kind: "task-kind", Name: "日常事务编排", Description: "重复性任务、个人助理流程与轻量自动化", DefaultOn: true},
			},
		},
		{
			ID:          "group-quality",
			Name:        "质量与验证",
			Description: "按步骤组织检查和测试，避免重复执行",
			Options: []domain.WorkflowOption{
				{ID: "static-quality", Kind: "test", Name: "静态检查与规范验证", Description: "语法、风格、结构和类型检查", DefaultOn: true},
				{ID: "behavior-tests", Kind: "test", Name: "行为测试链", Description: "单元、集成、端到端和部署后验证", DefaultOn: true},
				{ID: "self-audit", Kind: "audit", Name: "系统完整性自检", Description: "检查当前平台功能是否缺失或失衡", DefaultOn: true},
			},
		},
		{
			ID:          "group-ops",
			Name:        "运维与控制台",
			Description: "尽量让部署后日常操作不再依赖 SSH",
			Options: []domain.WorkflowOption{
				{ID: "web-ops", Kind: "ops", Name: "Web 运维面板", Description: "通过界面执行审计、建议、开关和模型包管理", DefaultOn: true},
				{ID: "adaptive-deploy", Kind: "ops", Name: "自适应部署", Description: "根据系统配置自动推荐部署模式和资源策略", DefaultOn: true},
				{ID: "upgrade-guidance", Kind: "ops", Name: "升级迁移建议", Description: "为未来更高配置 VPS 提供迁移建议", DefaultOn: true},
			},
		},
	}
}

func decorateFeatureToggle(toggle domain.FeatureToggle, profile domain.SystemProfile) domain.FeatureToggle {
	toggle.Allowed = true
	switch toggle.ID {
	case "builtin-model-packs":
		toggle.Recommended = profile.Tier == "performance"
		if profile.Tier == "compact" {
			toggle.Warning = "当前机器偏紧凑，内置模型包建议保持关闭，避免内存被本地模型占满。"
		} else if profile.Tier == "balanced" {
			toggle.Warning = "当前机器可按需开启少量轻量模型包，但应避免同时启用多个本地推理包。"
		}
	case "auto-ops":
		toggle.Recommended = true
		if profile.Tier == "compact" {
			toggle.Warning = "建议保留自动运维，但以串行巡检和轻量建议为主。"
		}
	case "adaptive-config":
		toggle.Recommended = true
	case "docs-automation":
		toggle.Recommended = true
	}
	return toggle
}

func decorateBuiltinModelPack(pack domain.BuiltinModelPack, profile domain.SystemProfile) domain.BuiltinModelPack {
	pack = finalizeBuiltinModelPackState(pack)
	tierValue := builtinModelTierValue(pack.SizeTier)
	mode, decision, allowed, recommended, warning := builtinModelPackPolicy(pack, profile, tierValue)
	pack.PolicyMode = mode
	pack.PolicyDecision = decision
	pack.Allowed = allowed
	pack.Recommended = recommended
	pack.Warning = warning
	pack.PolicyHint = strings.TrimSpace(decision.Message + " 当前候选定位：" + pack.PolicyHint)
	return pack
}

func builtinModelPackPolicy(pack domain.BuiltinModelPack, profile domain.SystemProfile, tierValue float64) (string, domain.PolicyDecision, bool, bool, string) {
	mode := "display-only"
	decision := domain.PolicyDecision{
		Area:   "local-model",
		Action: mode,
		Reason: "runtime-tier",
	}
	allowed := false
	recommended := false
	warning := ""

	switch {
	case tierValue > 0 && tierValue <= 4:
		switch profile.Tier {
		case "compact":
			decision.Message = "当前机器仅保留轻量本地模型展示位，不建议直接启用。"
			warning = "当前 1C1G 级别机器不建议启用该档本地原生模型，至少升级到 2C2G 后再尝试。"
		case "balanced":
			mode = "on-demand"
			decision.Action = mode
			decision.Message = "当前机器可按需唤起轻量本地模型，但应避免与其他重任务并发。"
			allowed = true
			recommended = true
			warning = "该档可按需启用，但建议串行运行，避免与构建任务并发。"
		case "performance":
			mode = "on-demand"
			decision.Action = mode
			decision.Message = "当前机器可把轻量本地模型作为按需唤起能力使用。"
			allowed = true
			recommended = true
		}
	case tierValue > 4 && tierValue <= 14:
		switch profile.Tier {
		case "compact":
			decision.Message = "当前机器仅展示中量本地模型目录，避免控制台被本地推理占满。"
			warning = "该体量原生本地模型不适合当前 1C1G 机器，建议至少升级到 4C8G。"
		case "balanced":
			mode = "on-demand"
			decision.Action = mode
			decision.Message = "当前机器允许按需唤起这一级模型，但更适合作为短时任务能力。"
			allowed = true
			recommended = tierValue <= 8
			warning = "该档可作为候选，但建议在 4C8G 以上机器使用以获得稳定体验。"
		case "performance":
			mode = "on-demand"
			decision.Action = mode
			decision.Message = "当前机器可按需唤起中量本地模型，用于高频本地实验。"
			allowed = true
			recommended = true
		}
	case tierValue > 14 && tierValue < 70:
		switch profile.Tier {
		case "compact":
			decision.Message = "当前机器仅展示高配本地模型候选，不允许在控制台所在主机直接启用。"
			warning = "该体量属于高配本地模型，不适合当前 1C1G 机器，建议至少升级到 12C32G 或独立 GPU。"
		case "balanced":
			mode = "externally-managed"
			decision.Action = mode
			decision.Reason = "dedicated-node-required"
			decision.Message = "该档模型建议交由独立高配节点或 GPU 主机托管，当前控制台仅保留候选位。"
			warning = "该档更适合独立高配节点或 GPU，当前机器建议仅保留候选展示。"
		case "performance":
			mode = "on-demand"
			decision.Action = mode
			decision.Message = "当前机器可按需唤起高配本地模型，但仍建议避免与控制台重任务长期共机。"
			allowed = true
			recommended = tierValue <= 32
			warning = "该档已接近高配区间，建议用于短时任务或迁移到独立推理节点。"
		}
	case tierValue >= 70:
		mode = "externally-managed"
		decision.Action = mode
		decision.Reason = "dedicated-node-required"
		if profile.Tier == "performance" {
			decision.Message = "即使当前机器允许管理该模型，也建议由独立高配节点或集群托管。"
			warning = "即使当前分层允许展示，该模型仍建议在独立高配节点上部署，不应与主控制台共机运行。"
		} else {
			decision.Message = "该体量属于超重本地模型，当前控制台仅负责展示与外部托管入口。"
			warning = "该体量属于高配或超高配本地模型，请在更高配置机器上部署。"
		}
	}

	if warning == "" && mode == "display-only" {
		warning = "当前机器以展示为主，建议优先使用 API 模型或更轻量候选。"
	}
	return mode, decision, allowed, recommended, warning
}

func builtinModelTierValue(sizeTier string) float64 {
	normalized := strings.ToLower(strings.TrimSpace(sizeTier))
	normalized = strings.TrimSuffix(normalized, "+")
	if strings.Contains(normalized, "x") {
		parts := strings.SplitN(normalized, "x", 2)
		left, _ := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(parts[0]), "b"), 64)
		right, _ := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(parts[1]), "b"), 64)
		return left * right
	}
	value, _ := strconv.ParseFloat(strings.TrimSuffix(normalized, "b"), 64)
	return value
}

func finalizeBuiltinModelPackState(pack domain.BuiltinModelPack) domain.BuiltinModelPack {
	if pack.InstallState == "removed" {
		pack.StatusDetail = "模型文件与本地配置已移除，保留候选位供后续重新部署。"
		return pack
	}
	if pack.InstallState == "queued" {
		pack.StatusDetail = "已加入本地部署队列，等待系统分配下载与配置阶段。"
		return pack
	}
	if pack.InstallState == "downloading" {
		pack.StatusDetail = "系统正在拉取模型文件与运行时依赖，请保持当前页面开启。"
		return pack
	}
	if pack.InstallState == "configuring" {
		pack.StatusDetail = "模型文件已就位，系统正在生成本地配置并接入调度。"
		return pack
	}
	if pack.Enabled && pack.Downloaded {
		pack.InstallState = "ready"
		pack.StatusDetail = "本地模型已完成下载与配置，可直接参与后续调度。"
		return pack
	}
	if pack.Downloaded {
		pack.InstallState = "disabled"
		pack.StatusDetail = "模型文件已保留在本地，当前处于停用状态。"
		return pack
	}
	pack.InstallState = "inactive"
	pack.StatusDetail = "当前尚未下载。启用后系统会自动完成拉取、配置与接入。"
	return pack
}

func featureToggles() []domain.FeatureToggle {
	return []domain.FeatureToggle{
		{ID: "builtin-model-packs", Name: "内置模型包", Description: "允许按需下载、部署和启用内置模型包，默认关闭", Enabled: false, DefaultOn: false},
		{ID: "auto-ops", Name: "自动运维", Description: "允许系统生成巡检、修复和运维建议", Enabled: true, DefaultOn: true},
		{ID: "adaptive-config", Name: "自适应配置", Description: "根据系统资源自动推荐并应用轻量配置", Enabled: true, DefaultOn: true},
		{ID: "docs-automation", Name: "文档自动化", Description: "自动生成和更新项目说明、Wiki 和交付文档", Enabled: true, DefaultOn: true},
	}
}

func builtinModelPacks() []domain.BuiltinModelPack {
	return []domain.BuiltinModelPack{
		{ID: "pack-qwen-u-1.5b", Provider: "Qwen", ModelName: "Qwen U", Version: "1.5B Uncensored", Name: "Qwen U 1.5B Uncensored", Description: "最小档中文原生候选，适合作为本地验证入口。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "1.5B", Variant: "GGUF Q4_K_M", SizeHint: "1-2GB", RuntimeHint: "建议 2C2G / 4GB RAM", SystemRequirement: "最低 2C2G，推荐 2C4G", Recommended: false, ReviewScore: 14, FilterScore: 14, AlignmentScore: 14, PolicyHint: "最轻量档，优先用于本地试跑。"},
		{ID: "pack-gemma-u-2.7b", Provider: "Gemma-compatible", ModelName: "Gemma U", Version: "2.7B Uncensored", Name: "Gemma U 2.7B Uncensored", Description: "2.7B 档轻量多语候选，适合做小型本地实验与摘要生成。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "2.7B", Variant: "GGUF Q4_K_M", SizeHint: "2-3GB", RuntimeHint: "建议 2C4G / 6GB RAM", SystemRequirement: "最低 2C4G，推荐 4C8G", Recommended: false, ReviewScore: 13, FilterScore: 13, AlignmentScore: 13, PolicyHint: "轻量多语原生候选。"},
		{ID: "pack-llama-u-3b", Provider: "Meta-compatible", ModelName: "Llama U", Version: "3B Uncensored", Name: "Llama U 3B Uncensored", Description: "3B 档通用原生候选，适合轻量开放问答。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "3B", Variant: "GGUF Q4_K_M", SizeHint: "2-4GB", RuntimeHint: "建议 2C4G / 6GB RAM", SystemRequirement: "最低 2C4G，推荐 4C8G", Recommended: false, ReviewScore: 13, FilterScore: 13, AlignmentScore: 13, PolicyHint: "轻量开放候选。"},
		{ID: "pack-phi-u-3.8b", Provider: "Phi-compatible", ModelName: "Phi U", Version: "3.8B Uncensored", Name: "Phi U 3.8B Uncensored", Description: "3.8B 档小型推理候选，适合作为轻量补充。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "3.8B", Variant: "GGUF Q4_K_M", SizeHint: "2-4GB", RuntimeHint: "建议 2C4G / 6GB RAM", SystemRequirement: "最低 2C4G，推荐 4C8G", Recommended: false, ReviewScore: 12, FilterScore: 12, AlignmentScore: 12, PolicyHint: "轻量推理补充位。"},
		{ID: "pack-qwen-u-4b", Provider: "Qwen", ModelName: "Qwen U", Version: "4B Uncensored", Name: "Qwen U 4B Uncensored", Description: "4B 档原生中文候选，适合做本地低限制试跑。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "4B", Variant: "GGUF Q4_K_M", SizeHint: "3-5GB", RuntimeHint: "建议 2C4G / 8GB RAM", SystemRequirement: "最低 2C4G，推荐 4C8G", Recommended: false, ReviewScore: 10, FilterScore: 10, AlignmentScore: 11, PolicyHint: "优先收录低限制原生变体，仅作本地候选展示。"},
		{ID: "pack-llama-u-4b", Provider: "Meta-compatible", ModelName: "Llama U", Version: "4B Uncensored", Name: "Llama U 4B Uncensored", Description: "4B 档英文与通用任务候选，适合做轻量本地实验。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "4B", Variant: "GGUF Q4_K_M", SizeHint: "3-5GB", RuntimeHint: "建议 2C4G / 8GB RAM", SystemRequirement: "最低 2C4G，推荐 4C8G", Recommended: false, ReviewScore: 11, FilterScore: 11, AlignmentScore: 12, PolicyHint: "低限制优先，保留原生候选位。"},
		{ID: "pack-phi-u-4b", Provider: "Phi-compatible", ModelName: "Phi U", Version: "4B Uncensored", Name: "Phi U 4B Uncensored", Description: "4B 档推理增强候选，适合做本地快速分析与草拟。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "4B", Variant: "GGUF Q4_K_M", SizeHint: "3-5GB", RuntimeHint: "建议 2C4G / 8GB RAM", SystemRequirement: "最低 2C4G，推荐 4C8G", Recommended: false, ReviewScore: 11, FilterScore: 11, AlignmentScore: 11, PolicyHint: "轻量推理增强位。"},
		{ID: "pack-openchat-u-4b", Provider: "OpenChat-compatible", ModelName: "OpenChat U", Version: "4B Uncensored", Name: "OpenChat U 4B Uncensored", Description: "4B 档开放对话候选，适合做轻量本地交流与草稿生成。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "4B", Variant: "GGUF Q4_K_M", SizeHint: "3-5GB", RuntimeHint: "建议 2C4G / 8GB RAM", SystemRequirement: "最低 2C4G，推荐 4C8G", Recommended: false, ReviewScore: 9, FilterScore: 9, AlignmentScore: 10, PolicyHint: "4B 档低限制对话候选。"},
		{ID: "pack-stablelm-u-4b", Provider: "StableLM-compatible", ModelName: "StableLM U", Version: "4B Uncensored", Name: "StableLM U 4B Uncensored", Description: "4B 档通用原生候选，适合补齐英文与多任务场景。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "4B", Variant: "GGUF Q4_K_M", SizeHint: "3-5GB", RuntimeHint: "建议 2C4G / 8GB RAM", SystemRequirement: "最低 2C4G，推荐 4C8G", Recommended: false, ReviewScore: 10, FilterScore: 9, AlignmentScore: 10, PolicyHint: "4B 档通用低限制补充位。"},
		{ID: "pack-qwen-u-7b", Provider: "Qwen", ModelName: "Qwen U", Version: "7B Uncensored", Name: "Qwen U 7B Uncensored", Description: "7B 档中文与代码混合候选，适合作为中阶本地原生库。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "7B", Variant: "GGUF Q4_K_M", SizeHint: "5-8GB", RuntimeHint: "建议 4C8G / 12GB RAM", SystemRequirement: "最低 4C8G，推荐 8C16G", Recommended: false, ReviewScore: 9, FilterScore: 9, AlignmentScore: 10, PolicyHint: "低审查低对齐优先的 7B 档候选。"},
		{ID: "pack-deepseek-u-7b", Provider: "DeepSeek", ModelName: "DeepSeek U", Version: "7B Uncensored", Name: "DeepSeek U 7B Uncensored", Description: "7B 档推理候选，面向更开放的本地问答与生成场景。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "7B", Variant: "GGUF Q4_K_M", SizeHint: "5-8GB", RuntimeHint: "建议 4C8G / 12GB RAM", SystemRequirement: "最低 4C8G，推荐 8C16G", Recommended: false, ReviewScore: 8, FilterScore: 8, AlignmentScore: 9, PolicyHint: "低限制优先的推理型候选。"},
		{ID: "pack-mistral-u-7b", Provider: "Mistral-compatible", ModelName: "Mistral U", Version: "7B Uncensored", Name: "Mistral U 7B Uncensored", Description: "7B 档通用原生候选，适合多语言内容生成。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "7B", Variant: "GGUF Q4_K_M", SizeHint: "5-8GB", RuntimeHint: "建议 4C8G / 12GB RAM", SystemRequirement: "最低 4C8G，推荐 8C16G", Recommended: false, ReviewScore: 12, FilterScore: 12, AlignmentScore: 13, PolicyHint: "保留原生多语言候选位。"},
		{ID: "pack-solar-u-7b", Provider: "Solar-compatible", ModelName: "Solar U", Version: "7B Uncensored", Name: "Solar U 7B Uncensored", Description: "7B 档开放通用候选，适合做本地多用途补充。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "7B", Variant: "GGUF Q4_K_M", SizeHint: "5-8GB", RuntimeHint: "建议 4C8G / 12GB RAM", SystemRequirement: "最低 4C8G，推荐 8C16G", Recommended: false, ReviewScore: 10, FilterScore: 10, AlignmentScore: 11, PolicyHint: "多用途中阶补充位。"},
		{ID: "pack-openhermes-u-7b", Provider: "OpenHermes-compatible", ModelName: "OpenHermes U", Version: "7B Uncensored", Name: "OpenHermes U 7B Uncensored", Description: "7B 档开放指令候选，适合作为对话与生成的低限制补充。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "7B", Variant: "GGUF Q4_K_M", SizeHint: "5-8GB", RuntimeHint: "建议 4C8G / 12GB RAM", SystemRequirement: "最低 4C8G，推荐 8C16G", Recommended: false, ReviewScore: 8, FilterScore: 8, AlignmentScore: 8, PolicyHint: "7B 档旗舰低限制候选。"},
		{ID: "pack-zephyr-u-7b", Provider: "Zephyr-compatible", ModelName: "Zephyr U", Version: "7B Uncensored", Name: "Zephyr U 7B Uncensored", Description: "7B 档多语开放候选，适合整理、草拟和通用问答。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "7B", Variant: "GGUF Q4_K_M", SizeHint: "5-8GB", RuntimeHint: "建议 4C8G / 12GB RAM", SystemRequirement: "最低 4C8G，推荐 8C16G", Recommended: false, ReviewScore: 9, FilterScore: 8, AlignmentScore: 9, PolicyHint: "7B 档多语低限制候选。"},
		{ID: "pack-neural-u-7b", Provider: "Neural-compatible", ModelName: "Neural U", Version: "7B Uncensored", Name: "Neural U 7B Uncensored", Description: "7B 档开放通用候选，适合作为中阶多任务工作马。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "7B", Variant: "GGUF Q4_K_M", SizeHint: "5-8GB", RuntimeHint: "建议 4C8G / 12GB RAM", SystemRequirement: "最低 4C8G，推荐 8C16G", Recommended: false, ReviewScore: 9, FilterScore: 9, AlignmentScore: 9, PolicyHint: "7B 档均衡低限制工作马。"},
		{ID: "pack-llama-u-8b", Provider: "Meta-compatible", ModelName: "Llama U", Version: "8B Uncensored", Name: "Llama U 8B Uncensored", Description: "8B 档开放通用候选，适合作为 7B 之上的补充。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "8B", Variant: "GGUF Q4_K_M", SizeHint: "6-9GB", RuntimeHint: "建议 4C8G / 16GB RAM", SystemRequirement: "最低 4C8G，推荐 8C16G", Recommended: false, ReviewScore: 9, FilterScore: 9, AlignmentScore: 10, PolicyHint: "中阶开放通用候选。"},
		{ID: "pack-qwen-u-14b", Provider: "Qwen", ModelName: "Qwen U", Version: "14B Uncensored", Name: "Qwen U 14B Uncensored", Description: "14B 档中文增强候选，适合独立中配节点。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "14B", Variant: "AWQ / GGUF", SizeHint: "10-18GB", RuntimeHint: "建议 8C16G / 24GB RAM", SystemRequirement: "最低 8C16G，推荐 12C32G", Recommended: false, ReviewScore: 8, FilterScore: 8, AlignmentScore: 9, PolicyHint: "中高体量中文原生候选。"},
		{ID: "pack-yi-u-14b", Provider: "Yi-compatible", ModelName: "Yi U", Version: "14B Uncensored", Name: "Yi U 14B Uncensored", Description: "14B 档多语原生候选，适合作为中高配通用服务。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "14B", Variant: "AWQ / GGUF", SizeHint: "10-18GB", RuntimeHint: "建议 8C16G / 24GB RAM", SystemRequirement: "最低 8C16G，推荐 12C32G", Recommended: false, ReviewScore: 9, FilterScore: 9, AlignmentScore: 10, PolicyHint: "多语中高配候选。"},
		{ID: "pack-openchat-u-14b", Provider: "OpenChat-compatible", ModelName: "OpenChat U", Version: "14B Uncensored", Name: "OpenChat U 14B Uncensored", Description: "14B 档开放对话候选，适合做独立多轮服务。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "14B", Variant: "AWQ / GGUF", SizeHint: "10-18GB", RuntimeHint: "建议 8C16G / 24GB RAM", SystemRequirement: "最低 8C16G，推荐 12C32G", Recommended: false, ReviewScore: 8, FilterScore: 8, AlignmentScore: 8, PolicyHint: "14B 档旗舰低限制对话候选。"},
		{ID: "pack-nemotron-u-14b", Provider: "Nemotron-compatible", ModelName: "Nemotron U", Version: "14B Uncensored", Name: "Nemotron U 14B Uncensored", Description: "14B 档推理与生成平衡候选，适合作为中高配核心服务。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "14B", Variant: "AWQ / GGUF", SizeHint: "10-18GB", RuntimeHint: "建议 8C16G / 24GB RAM", SystemRequirement: "最低 8C16G，推荐 12C32G", Recommended: false, ReviewScore: 8, FilterScore: 9, AlignmentScore: 8, PolicyHint: "14B 档平衡型低限制候选。"},
		{ID: "pack-mixtral-u-8x7b", Provider: "Mixtral-compatible", ModelName: "Mixtral U", Version: "8x7B Uncensored", Name: "Mixtral U 8x7B Uncensored", Description: "MoE 路线候选，适合需要更大上下文和多语能力的场景。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "8x7B", Variant: "AWQ / EXL2", SizeHint: "26-40GB", RuntimeHint: "建议 12C32G 或独立 GPU", SystemRequirement: "最低 12C32G，推荐 16C64G 或 GPU", Recommended: false, ReviewScore: 7, FilterScore: 7, AlignmentScore: 8, PolicyHint: "保留低限制 MoE 候选位。"},
		{ID: "pack-qwen-u-32b", Provider: "Qwen", ModelName: "Qwen U", Version: "32B Uncensored", Name: "Qwen U 32B Uncensored", Description: "32B 档高质量中文候选，适合作为高配本地服务。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "32B", Variant: "AWQ / GGUF", SizeHint: "20-40GB", RuntimeHint: "建议 12C32G 或独立 GPU", SystemRequirement: "最低 12C32G，推荐 16C64G 或 GPU", Recommended: false, ReviewScore: 7, FilterScore: 7, AlignmentScore: 8, PolicyHint: "高配中文原生候选。"},
		{ID: "pack-deepseek-u-32b", Provider: "DeepSeek", ModelName: "DeepSeek U", Version: "32B Uncensored", Name: "DeepSeek U 32B Uncensored", Description: "32B 档推理增强候选，适合作为高配独立生成与分析节点。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "32B", Variant: "AWQ / GGUF", SizeHint: "20-40GB", RuntimeHint: "建议 12C32G 或独立 GPU", SystemRequirement: "最低 12C32G，推荐 16C64G 或 GPU", Recommended: false, ReviewScore: 6, FilterScore: 6, AlignmentScore: 7, PolicyHint: "32B 档旗舰低限制推理候选。"},
		{ID: "pack-command-u-34b", Provider: "Command-compatible", ModelName: "Command U", Version: "34B Uncensored", Name: "Command U 34B Uncensored", Description: "34B 档长上下文候选，适合独立知识整理与长文生成。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "34B", Variant: "AWQ / EXL2", SizeHint: "22-42GB", RuntimeHint: "建议 12C32G 或独立 GPU", SystemRequirement: "最低 12C32G，推荐 16C64G 或 GPU", Recommended: false, ReviewScore: 7, FilterScore: 7, AlignmentScore: 8, PolicyHint: "长上下文高配候选。"},
		{ID: "pack-yi-u-34b", Provider: "Yi-compatible", ModelName: "Yi U", Version: "34B Uncensored", Name: "Yi U 34B Uncensored", Description: "34B 档多语高配候选，适合作为知识整理与创作的独立节点。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "34B", Variant: "AWQ / EXL2", SizeHint: "22-42GB", RuntimeHint: "建议 12C32G 或独立 GPU", SystemRequirement: "最低 12C32G，推荐 16C64G 或 GPU", Recommended: false, ReviewScore: 6, FilterScore: 7, AlignmentScore: 7, PolicyHint: "34B 档多语旗舰候选。"},
		{ID: "pack-llama-u-70b", Provider: "Meta-compatible", ModelName: "Llama U", Version: "70B Uncensored", Name: "Llama U 70B Uncensored", Description: "70B 档高配原生候选，用于独立推理节点。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "70B", Variant: "AWQ / GGUF", SizeHint: "40-80GB", RuntimeHint: "建议 16C64G 或独立 GPU", SystemRequirement: "最低 16C64G，推荐 24C128G 或 GPU", Recommended: false, ReviewScore: 7, FilterScore: 7, AlignmentScore: 8, PolicyHint: "仅作为高配本地候选展示，不适合轻量 VPS。"},
		{ID: "pack-command-u-70b", Provider: "Command-compatible", ModelName: "Command U", Version: "70B Uncensored", Name: "Command U 70B Uncensored", Description: "70B 档长上下文高配候选，适合独立知识与助手服务。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "70B", Variant: "AWQ / EXL2", SizeHint: "40-80GB", RuntimeHint: "建议 16C64G 或独立 GPU", SystemRequirement: "最低 16C64G，推荐 24C128G 或 GPU", Recommended: false, ReviewScore: 6, FilterScore: 6, AlignmentScore: 7, PolicyHint: "70B 档长上下文低限制候选。"},
		{ID: "pack-qwen-u-72b", Provider: "Qwen", ModelName: "Qwen U", Version: "72B Uncensored", Name: "Qwen U 72B Uncensored", Description: "72B 档中文高配原生候选，适合作为独立模型服务。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "72B", Variant: "AWQ / GGUF", SizeHint: "42-85GB", RuntimeHint: "建议 16C64G 或独立 GPU", SystemRequirement: "最低 16C64G，推荐 24C128G 或 GPU", Recommended: false, ReviewScore: 6, FilterScore: 6, AlignmentScore: 7, PolicyHint: "高体量低限制候选，应与主控制台分机部署。"},
		{ID: "pack-falcon-u-180b", Provider: "Falcon-compatible", ModelName: "Falcon U", Version: "180B Uncensored", Name: "Falcon U 180B Uncensored", Description: "180B 档超大候选，用于更重的本地推理规划。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "180B", Variant: "FP8 / Cluster Quant", SizeHint: "120-200GB", RuntimeHint: "建议多 GPU / 集群", SystemRequirement: "最低 4x GPU，推荐专用推理集群", Recommended: false, ReviewScore: 6, FilterScore: 6, AlignmentScore: 7, PolicyHint: "超大体量候选，仅作规划展示。"},
		{ID: "pack-llama-u-405b", Provider: "Meta-compatible", ModelName: "Llama U", Version: "405B Uncensored", Name: "Llama U 405B Uncensored", Description: "405B 档超大原生候选，仅用于超高配集群级部署规划。", Enabled: false, Downloaded: false, DefaultOn: false, SizeTier: "405B", Variant: "FP8 / Quantized Cluster", SizeHint: "240GB+", RuntimeHint: "建议多 GPU / 集群节点", SystemRequirement: "最低 8x GPU 或等效集群，推荐专用推理集群", Recommended: false, ReviewScore: 5, FilterScore: 5, AlignmentScore: 6, PolicyHint: "超高配展示位，只用于规划与候选管理。"},
	}
}
