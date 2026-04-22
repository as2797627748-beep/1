package service

import (
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"autocode-platform/internal/domain"
)

func detectSystemProfile() domain.SystemProfile {
	memoryMB := detectMemoryMB()
	cpu := runtime.NumCPU()
	tier := "balanced"
	recommended := "2C4G"
	modes := []string{"atelier", "mission-control", "pocket"}
	features := []string{"workflow-config", "system-audit", "web-ops"}
	upgrade := []string{"保持应用级稳定部署，升级后可直接恢复当前版本目录"}
	readiness := "基础免 SSH 管理可用，适合继续向 Web 控制台集中运维演进"
	recommendedConcurrency := 2
	localModelAllowed := true
	strategySummary := "当前机器适合均衡闭环，默认保留常用自动化能力，并限制重动作的同时活跃数量。"

	if cpu <= 1 || memoryMB <= 1200 {
		tier = "compact"
		recommended = "1C1G 到 2C2G"
		features = append(features, "serial-scheduler", "lite-model-routing")
		upgrade = append(upgrade, "建议未来升级到 2C4G 以承载更多模型和并行任务")
		recommendedConcurrency = 1
		localModelAllowed = false
		strategySummary = "当前机器应采用轻载串行闭环，重能力默认关闭或外移，本地模型仅展示为主。"
	}
	if cpu >= 4 && memoryMB >= 7800 {
		tier = "performance"
		recommended = "4C8G 或更高"
		features = append(features, "parallel-jobs", "local-model-packs")
		upgrade = append(upgrade, "可开启更多内置模型包和并发工作流")
		recommendedConcurrency = 3
		localModelAllowed = true
		strategySummary = "当前机器可承载更积极的闭环并发与本地能力，但仍应坚持按需启用和限流治理。"
	}

	return domain.SystemProfile{
		OS:                     runtime.GOOS,
		Arch:                   runtime.GOARCH,
		CPUCores:               cpu,
		MemoryMB:               memoryMB,
		Tier:                   tier,
		RecommendedConcurrency: recommendedConcurrency,
		LocalModelAllowed:      localModelAllowed,
		RecommendedVPS:         recommended,
		RecommendedModes:       modes,
		RecommendedFeatures:    features,
		UpgradeSuggestions:     upgrade,
		NoSSHReadiness:         readiness,
		StrategySummary:        strategySummary,
	}
}

func detectMemoryMB() int {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 1024
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "MemTotal:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		kb, parseErr := strconv.Atoi(fields[1])
		if parseErr != nil {
			continue
		}
		return kb / 1024
	}
	return 1024
}

func capabilityProfiles(profile domain.SystemProfile) []domain.CapabilityProfile {
	compact := profile.Tier == "compact"
	balanced := profile.Tier == "balanced"
	performance := profile.Tier == "performance"
	profiles := []domain.CapabilityProfile{
		{
			ID:          "cap-control-plane",
			Category:    "core",
			Name:        "控制中枢与界面治理",
			Mode:        "resident",
			Allowed:     true,
			Recommended: true,
			Summary:     "控制台、编排入口与系统状态面板应持续可用。",
			Reason:      "超级智能体的核心必须轻量常驻，其余重能力按需唤起。",
		},
		{
			ID:          "cap-api-routing",
			Category:    "models",
			Name:        "API 模型接入与路由",
			Mode:        "on-demand",
			Allowed:     true,
			Recommended: true,
			Summary:     "外部模型平台优先按需接入，不在本机长期占用额外资源。",
			Reason:      "弱机环境下优先利用外部推理能力，保留本地控制面。",
		},
		{
			ID:          "cap-local-models",
			Category:    "models",
			Name:        "本地原生模型唤起",
			Mode:        ternaryMode(compact, balanced, performance, "display-only", "on-demand", "on-demand"),
			Allowed:     !compact,
			Recommended: performance,
			Summary:     ternaryText(compact, balanced, performance, "当前机器仅保留本地模型候选展示，不建议实际拉起。", "当前机器可按需短时唤起轻量本地模型，不建议长期常驻。", "当前机器可按需启用更多本地模型，但仍以按需唤起优先。"),
			Reason:      "本地模型最耗内存与磁盘，应始终延迟启动并限制常驻。",
		},
		{
			ID:          "cap-research",
			Category:    "knowledge",
			Name:        "研究检索与资料整理",
			Mode:        "on-demand",
			Allowed:     true,
			Recommended: true,
			Summary:     "检索、摘要和资料整理应按任务即时调用，不保留重后台进程。",
			Reason:      "研究任务频率波动大，按需调用可减少空转占用。",
		},
		{
			ID:          "cap-doc-processing",
			Category:    "knowledge",
			Name:        "文档解析与内容处理",
			Mode:        "on-demand",
			Allowed:     true,
			Recommended: true,
			Summary:     "解析、抽取、OCR 与转换链路应按需触发并及时清理缓存。",
			Reason:      "文档处理中间产物容易占用磁盘，不适合长期保留。",
		},
		{
			ID:          "cap-automation",
			Category:    "ops",
			Name:        "自动修复与自我调整",
			Mode:        ternaryMode(compact, balanced, performance, "on-demand", "on-demand", "resident"),
			Allowed:     true,
			Recommended: true,
			Summary:     ternaryText(compact, balanced, performance, "当前机器建议在需要时再触发自修复与自优化。", "当前机器可按需执行自修复、自调整与轻量优化闭环。", "当前机器可持续运行更积极的自修复与治理策略。"),
			Reason:      "自修复能力是超级智能体关键能力，但应随机器余量动态调整。",
		},
		{
			ID:          "cap-deployments",
			Category:    "ops",
			Name:        "部署、发布与环境接管",
			Mode:        ternaryMode(compact, balanced, performance, "on-demand", "on-demand", "on-demand"),
			Allowed:     true,
			Recommended: true,
			Summary:     "部署与环境配置应在用户触发时执行，并根据当前配置做最优轻量方案。",
			Reason:      "发布任务通常是峰值操作，不应长期占资源。",
		},
		{
			ID:          "cap-background-workers",
			Category:    "runtime",
			Name:        "后台工作进程",
			Mode:        ternaryMode(compact, balanced, performance, "display-only", "on-demand", "on-demand"),
			Allowed:     !compact,
			Recommended: balanced || performance,
			Summary:     ternaryText(compact, balanced, performance, "当前机器不建议维持额外后台工作进程常驻。", "当前机器可按需启动短生命周期后台工作进程。", "当前机器可承载更多后台任务，但仍应避免无意义常驻。"),
			Reason:      "后台进程最容易悄悄吃掉内存与 CPU，需要严格受控。",
		},
		{
			ID:          "cap-heavy-assets",
			Category:    "storage",
			Name:        "大文件缓存与外部资源",
			Mode:        "external-managed",
			Allowed:     true,
			Recommended: true,
			Summary:     "大型资源优先使用多源引用、短期缓存和自动替换，不默认长期落盘。",
			Reason:      "磁盘通常比功能野心更紧张，必须把大资源当作外部托管对象处理。",
		},
	}
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].ID < profiles[j].ID })
	return profiles
}

func optimizationProfiles(profile domain.SystemProfile) []domain.OptimizationProfile {
	profiles := []domain.OptimizationProfile{
		{
			ID:       "opt-control-plane",
			Priority: "high",
			Title:    "保持控制中枢轻量常驻",
			Action:   "仅保留界面、调度与状态面板常驻，其他重能力全部改为按需唤起。",
			Reason:   "这样才能在高低配机器、家用电脑和可变更环境里都维持稳定。",
		},
		{
			ID:       "opt-cache-policy",
			Priority: "high",
			Title:    "所有重资源走短驻留缓存",
			Action:   "模型、文档中间产物和外部大文件都不长期落盘，使用短期缓存与自动清理。",
			Reason:   "磁盘通常不宽裕，重资源必须默认可释放。",
		},
	}

	switch profile.Tier {
	case "compact":
		profiles = append(profiles,
			domain.OptimizationProfile{
				ID:       "opt-compact-home",
				Priority: "high",
				Title:    "启用家用轻载模式",
				Action:   "强制串行任务、停用后台常驻工作进程、本地模型仅展示或极轻量按需尝试。",
				Reason:   "1C1G 到 2C2G 级别机器更适合把重能力外移，避免影响日常使用。",
			},
			domain.OptimizationProfile{
				ID:       "opt-compact-api-first",
				Priority: "high",
				Title:    "优先使用外部 API 与外部托管",
				Action:   "研究、推理、重文档处理优先走外部能力，本机仅保留编排与治理。",
				Reason:   "弱机更适合做超级智能体控制台，而不是重执行节点。",
			},
			domain.OptimizationProfile{
				ID:       "opt-compact-burst",
				Priority: "medium",
				Title:    "禁止峰值并发",
				Action:   "避免同时执行构建、测试、部署和本地推理，所有重任务按顺序排队。",
				Reason:   "小机器最怕瞬时峰值把系统拖死。",
			},
		)
	case "balanced":
		profiles = append(profiles,
			domain.OptimizationProfile{
				ID:       "opt-balanced-hybrid",
				Priority: "high",
				Title:    "采用混合执行策略",
				Action:   "本机保留轻量自动化与短时本地能力，重推理与大资源仍优先按需或外部托管。",
				Reason:   "中配机器适合承担部分能力，但不适合把所有模块都做成常驻服务。",
			},
			domain.OptimizationProfile{
				ID:       "opt-balanced-local",
				Priority: "medium",
				Title:    "本地模型仅短时唤起",
				Action:   "只在明确需要时启动轻量本地模型，完成后回收资源。",
				Reason:   "这样能兼顾可用性与系统响应速度。",
			},
		)
	case "performance":
		profiles = append(profiles,
			domain.OptimizationProfile{
				ID:       "opt-performance-guardrails",
				Priority: "medium",
				Title:    "高配也保持按需优先",
				Action:   "即使在高配机器上，也只让真正核心能力常驻，其他模块按需启动并设资源上限。",
				Reason:   "高配不代表应该无限常驻，受控运行更稳定。",
			},
			domain.OptimizationProfile{
				ID:       "opt-performance-split",
				Priority: "medium",
				Title:    "重能力分机或分层运行",
				Action:   "将本地模型、大文件处理和后台工作进程尽量与主控制台解耦，避免互相抢资源。",
				Reason:   "超级系统要更看重治理稳定性，而不是单机堆满功能。",
			},
		)
	}

	sort.Slice(profiles, func(i, j int) bool {
		if profiles[i].Priority != profiles[j].Priority {
			return optimizationPriorityWeight(profiles[i].Priority) < optimizationPriorityWeight(profiles[j].Priority)
		}
		return profiles[i].ID < profiles[j].ID
	})
	return profiles
}

func optimizationPriorityWeight(priority string) int {
	switch priority {
	case "high":
		return 0
	case "medium":
		return 1
	default:
		return 2
	}
}

func runtimePolicy(profile domain.SystemProfile) domain.RuntimePolicy {
	switch profile.Tier {
	case "compact":
		return domain.RuntimePolicy{
			Profile:              "home-lite",
			MaxConcurrentRuns:    1,
			MaxHeavyActions:      1,
			AllowBackgroundJobs:  false,
			AllowLocalModels:     false,
			CacheBudgetMB:        512,
			ValidationDepth:      "essential",
			DefaultEnabledTools:  []string{"workspace", "terminal", "tests", "logs", "analysis"},
			DefaultDisabledTools: []string{"deploy", "build"},
			Summary:              "适合低配 VPS 或家用轻载环境，重能力默认关闭或转为外部托管。",
		}
	case "performance":
		return domain.RuntimePolicy{
			Profile:              "adaptive-performance",
			MaxConcurrentRuns:    3,
			MaxHeavyActions:      2,
			AllowBackgroundJobs:  true,
			AllowLocalModels:     true,
			CacheBudgetMB:        4096,
			ValidationDepth:      "extended",
			DefaultEnabledTools:  []string{"workspace", "terminal", "tests", "deploy", "logs", "analysis", "build"},
			DefaultDisabledTools: []string{},
			Summary:              "适合高配机器，允许更多本地能力，但仍优先按需唤起而非无限常驻。",
		}
	default:
		return domain.RuntimePolicy{
			Profile:              "balanced-hybrid",
			MaxConcurrentRuns:    2,
			MaxHeavyActions:      1,
			AllowBackgroundJobs:  true,
			AllowLocalModels:     true,
			CacheBudgetMB:        1536,
			ValidationDepth:      "standard",
			DefaultEnabledTools:  []string{"workspace", "terminal", "tests", "logs", "analysis", "deploy"},
			DefaultDisabledTools: []string{"build"},
			Summary:              "适合中配机器，保留常用自动化能力，重任务按需启用。",
		}
	}
}

func ternaryMode(compact, balanced, performance bool, compactMode, balancedMode, performanceMode string) string {
	if compact {
		return compactMode
	}
	if balanced {
		return balancedMode
	}
	if performance {
		return performanceMode
	}
	return balancedMode
}

func ternaryText(compact, balanced, performance bool, compactText, balancedText, performanceText string) string {
	if compact {
		return compactText
	}
	if balanced {
		return balancedText
	}
	if performance {
		return performanceText
	}
	return balancedText
}
