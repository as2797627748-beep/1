package service

import (
	"fmt"
	"strings"

	"autocode-platform/internal/domain"
)

func buildAudit(summary domain.SystemSummary) domain.AuditReport {
	checks := []domain.AuditCheck{
		{
			ID:         "workflow-groups",
			Name:       "工作流分组完整性",
			Status:     statusFrom(len(summary.WorkflowOptions) >= 4),
			Summary:    "工作流选项已按入口、生产、质量和运维分组整理",
			Suggestion: "可继续为更多任务类型补充分组说明和方案模板",
		},
		{
			ID:         "tool-catalog",
			Name:       "工具目录覆盖",
			Status:     statusFrom(len(summary.ToolCatalog) >= 8),
			Summary:    "工具目录已覆盖分析、工作区、测试、构建和部署",
			Suggestion: "可继续补充文档、知识整理和审计类工具入口",
		},
		{
			ID:         "test-catalog",
			Name:       "测试链完整性",
			Status:     statusFrom(len(summary.TestCatalog) >= 5),
			Summary:    "测试目录已覆盖静态、单元、集成、端到端和部署后验证",
			Suggestion: "后续可继续补充性能和回归测试分类",
		},
		{
			ID:         "no-ssh",
			Name:       "免 SSH 运维准备度",
			Status:     statusFrom(strings.Contains(summary.SystemProfile.NoSSHReadiness, "可用")),
			Summary:    summary.SystemProfile.NoSSHReadiness,
			Suggestion: "建议继续将部署、升级和模型管理动作收口到 Web 控制台",
		},
	}
	score := 0
	for _, check := range checks {
		if check.Status == "passed" {
			score += 25
		} else {
			score += 12
		}
	}
	return domain.AuditReport{
		Score:   score,
		Summary: "已完成平台完整性、运维能力和能力目录检查",
		Checks:  checks,
	}
}

func buildAdvice(summary domain.SystemSummary, req domain.AdviceRequest) domain.AdviceReport {
	mode := strings.TrimSpace(req.Mode)
	if mode == "" {
		mode = "review"
	}
	suggestions := adviceSuggestions(mode, summary, req)
	hostingMode := ""
	checklist := []string{}
	consoleSteps := []string{}
	prompt := "请基于当前平台摘要、工具目录、测试目录、工作流分组和系统配置，对 " + mode + " 模式给出改进建议。"
	if strings.TrimSpace(req.Goal) != "" {
		prompt += " 目标: " + strings.TrimSpace(req.Goal)
	}
	if strings.TrimSpace(req.Target) != "" {
		prompt += " 托管目标: " + strings.TrimSpace(req.Target)
	}
	summaryText := "已整理当前系统建议，可用于审查、规划更新或运维优化"
	if strings.Contains(mode, "external-hosting") {
		hostingMode = classifyHostingMode(req.Target)
		summaryText = "已整理外部托管建议，可用于独立节点、GPU 主机或集群级部署规划"
		checklist = hostingChecklist(hostingMode)
		consoleSteps = hostingConsoleSteps(hostingMode)
	}
	return domain.AdviceReport{
		Mode:         mode,
		Summary:      summaryText,
		PromptBundle: prompt,
		HostingMode:  hostingMode,
		Checklist:    checklist,
		ConsoleSteps: consoleSteps,
		Suggestions:  suggestions,
	}
}

func adviceSuggestions(mode string, summary domain.SystemSummary, req domain.AdviceRequest) []string {
	goal := strings.TrimSpace(req.Goal)
	if strings.Contains(mode, "external-hosting") {
		recommended := summary.SystemProfile.RecommendedVPS
		if recommended == "" {
			recommended = "4C8G 或更高"
		}
		hostingMode := classifyHostingMode(req.Target)
		hostingLabel := hostingModeLabel(hostingMode)
		shapeSuggestion := "优先准备独立的 CPU 节点部署、健康检查与幂等重启脚本，控制台只保留编排与状态透传"
		capacitySuggestion := fmt.Sprintf("建议目标规格从 %s 起步，并按 %s 准备对应部署脚本、监控与容量预留", recommended, hostingLabel)
		scalingSuggestion := "将模型服务和控制台资源隔离，避免推理进程与 Web 中枢争抢内存、磁盘和后台任务额度"
		switch hostingMode {
		case "gpu":
			shapeSuggestion = "优先准备单机 GPU 托管，明确显存预算、量化方案、驱动版本和热重启脚本，控制台只保留入口与状态治理"
			capacitySuggestion = "建议为该模型预留独立 GPU 主机与显存余量，并补齐下载、加载、健康检查和切换脚本"
			scalingSuggestion = "为 GPU 单机补齐显存水位、推理队列和故障回退监控，避免重载时拖垮主控制台"
		case "cluster":
			shapeSuggestion = "优先准备集群级托管，拆分调度、推理、缓存和监控职责，控制台只负责策略编排与多节点治理"
			capacitySuggestion = "建议按集群级托管准备节点分层、容量池、统一镜像和滚动发布脚本，而不是继续依赖单机托管"
			scalingSuggestion = "为集群补齐服务发现、负载均衡、弹性扩缩和跨节点健康检查，确保超重模型可持续运行"
		}
		suggestions := []string{
			"推荐托管形态: " + hostingLabel,
			shapeSuggestion,
			scalingSuggestion,
			capacitySuggestion,
		}
		if goal != "" {
			suggestions = append([]string{"当前规划目标: " + goal}, suggestions...)
		}
		return suggestions
	}
	return []string{
		"优先补齐真实持久化和真实 provider 执行链，避免平台只停留在流程编排层",
		"将模型包下载、解压、启停与资源检查完全图形化，进一步减少 SSH 依赖",
		"把文档、运维、发布和知识整理任务与代码任务统一到同一工作流编排器中",
	}
}

func classifyHostingMode(target string) string {
	normalized := strings.ToLower(strings.TrimSpace(target))
	switch normalized {
	case "cluster":
		return "cluster"
	case "gpu", "node":
		return normalized
	default:
		return "cpu"
	}
}

func hostingModeLabel(mode string) string {
	switch mode {
	case "cluster":
		return "集群级托管"
	case "gpu":
		return "GPU 单机"
	case "node":
		return "CPU 独立节点"
	default:
		return "CPU 独立节点"
	}
}

func hostingChecklist(mode string) []string {
	base := []string{
		"为托管节点单独预留模型缓存、日志目录和健康检查端点，避免与主控制台混用运行目录",
		"准备稳定部署脚本，确保重复执行时只做增量更新、重启和状态校验",
		"接入进程存活、磁盘占用和模型加载成功率监控，避免无感失败",
	}
	switch mode {
	case "gpu":
		return append(base,
			"固定驱动、CUDA 与量化格式版本，避免 GPU 主机升级后出现模型无法加载",
			"为显存水位和推理并发设置保护阈值，防止高峰时直接挤爆宿主机",
		)
	case "cluster":
		return append(base,
			"为集群准备统一镜像、服务发现和滚动发布策略，避免节点间环境漂移",
			"把推理入口、缓存层和监控层拆分部署，保证超重模型横向扩容时可治理",
		)
	default:
		return append(base,
			"为 CPU 独立节点预留足够内存与交换空间，避免模型初始化阶段直接 OOM",
			"限制同机后台任务数量，让模型托管节点只承担推理与健康检查职责",
		)
	}
}

func hostingConsoleSteps(mode string) []string {
	label := hostingModeLabel(mode)
	return []string{
		"在策略中心保留该模型的“生成托管建议”入口，作为后续扩容与迁移的统一出口",
		"将托管目标登记为“" + label + "”，让控制台后续可直接复用同一托管口径展示状态",
		"控制台只保留启停、健康状态、容量提示和迁移建议，不把超重推理进程放回主机常驻",
	}
}

func statusFrom(ok bool) string {
	if ok {
		return "passed"
	}
	return "partial"
}
