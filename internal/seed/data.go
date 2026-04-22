package seed

import "autocode-platform/internal/domain"

func ModelCatalog() []domain.ModelProfile {
	return []domain.ModelProfile{
		{ID: "openai:gpt-4.1", Provider: "OpenAI", Name: "GPT-4.1", Version: "gpt-4.1", FullName: "OpenAI / GPT-4.1 / gpt-4.1", Website: "https://platform.openai.com/docs/models", Region: "global", Tags: []string{"coding", "reasoning"}, ReviewScore: 55, FilterScore: 58, AlignmentScore: 60},
		{ID: "openai:gpt-4.1-mini", Provider: "OpenAI", Name: "GPT-4.1 mini", Version: "gpt-4.1-mini", FullName: "OpenAI / GPT-4.1 mini / gpt-4.1-mini", Website: "https://platform.openai.com/docs/models", Region: "global", Tags: []string{"fast", "general"}, ReviewScore: 54, FilterScore: 57, AlignmentScore: 58},
		{ID: "anthropic:claude-sonnet-4", Provider: "Anthropic", Name: "Claude Sonnet 4", Version: "claude-sonnet-4", FullName: "Anthropic / Claude Sonnet 4 / claude-sonnet-4", Website: "https://docs.anthropic.com", Region: "global", Tags: []string{"coding", "agent"}, ReviewScore: 60, FilterScore: 62, AlignmentScore: 64},
		{ID: "anthropic:claude-opus-4", Provider: "Anthropic", Name: "Claude Opus 4", Version: "claude-opus-4", FullName: "Anthropic / Claude Opus 4 / claude-opus-4", Website: "https://docs.anthropic.com", Region: "global", Tags: []string{"premium", "reasoning"}, ReviewScore: 61, FilterScore: 63, AlignmentScore: 65},
		{ID: "google:gemini-2.5-pro", Provider: "Google", Name: "Gemini 2.5 Pro", Version: "gemini-2.5-pro", FullName: "Google / Gemini 2.5 Pro / gemini-2.5-pro", Website: "https://ai.google.dev", Region: "global", Tags: []string{"reasoning", "multimodal"}, ReviewScore: 57, FilterScore: 60, AlignmentScore: 61},
		{ID: "google:gemini-2.5-flash", Provider: "Google", Name: "Gemini 2.5 Flash", Version: "gemini-2.5-flash", FullName: "Google / Gemini 2.5 Flash / gemini-2.5-flash", Website: "https://ai.google.dev", Region: "global", Tags: []string{"fast", "multimodal"}, ReviewScore: 56, FilterScore: 58, AlignmentScore: 59},
		{ID: "deepseek:deepseek-r1", Provider: "DeepSeek", Name: "DeepSeek R1", Version: "deepseek-r1", FullName: "DeepSeek / DeepSeek R1 / deepseek-r1", Website: "https://platform.deepseek.com", Region: "global", Tags: []string{"reasoning", "cost-effective"}, ReviewScore: 38, FilterScore: 40, AlignmentScore: 42},
		{ID: "deepseek:deepseek-v3", Provider: "DeepSeek", Name: "DeepSeek V3", Version: "deepseek-v3", FullName: "DeepSeek / DeepSeek V3 / deepseek-v3", Website: "https://platform.deepseek.com", Region: "global", Tags: []string{"general", "coding"}, ReviewScore: 39, FilterScore: 41, AlignmentScore: 43},
		{ID: "zhipu:glm-4.5", Provider: "Zhipu", Name: "GLM-4.5", Version: "glm-4.5", FullName: "Zhipu / GLM-4.5 / glm-4.5", Website: "https://open.bigmodel.cn", Region: "domestic", Tags: []string{"domestic", "general"}, ReviewScore: 52, FilterScore: 54, AlignmentScore: 55},
		{ID: "moonshot:kimi-k2", Provider: "Moonshot", Name: "Kimi K2", Version: "kimi-k2", FullName: "Moonshot / Kimi K2 / kimi-k2", Website: "https://platform.moonshot.cn", Region: "domestic", Tags: []string{"domestic", "long-context"}, ReviewScore: 47, FilterScore: 49, AlignmentScore: 50},
		{ID: "qwen:qwen3-coder", Provider: "Qwen", Name: "Qwen3 Coder", Version: "qwen3-coder", FullName: "Qwen / Qwen3 Coder / qwen3-coder", Website: "https://help.aliyun.com/zh/model-studio", Region: "domestic", Tags: []string{"coding", "domestic"}, ReviewScore: 44, FilterScore: 46, AlignmentScore: 47},
		{ID: "qwen:qwen-max", Provider: "Qwen", Name: "Qwen Max", Version: "qwen-max", FullName: "Qwen / Qwen Max / qwen-max", Website: "https://help.aliyun.com/zh/model-studio", Region: "domestic", Tags: []string{"general", "domestic"}, ReviewScore: 45, FilterScore: 47, AlignmentScore: 48},
		{ID: "groq:llama-3.3-70b", Provider: "Groq", Name: "Llama 3.3 70B", Version: "llama-3.3-70b", FullName: "Groq / Llama 3.3 70B / llama-3.3-70b", Website: "https://console.groq.com/docs/models", Region: "global", Tags: []string{"fast", "open"}, ReviewScore: 42, FilterScore: 44, AlignmentScore: 45},
		{ID: "mistral:mistral-large", Provider: "Mistral", Name: "Mistral Large", Version: "mistral-large", FullName: "Mistral / Mistral Large / mistral-large", Website: "https://docs.mistral.ai", Region: "global", Tags: []string{"open-ecosystem", "general"}, ReviewScore: 43, FilterScore: 45, AlignmentScore: 46},
		{ID: "minimax:minimax-m1", Provider: "MiniMax", Name: "MiniMax M1", Version: "minimax-m1", FullName: "MiniMax / MiniMax M1 / minimax-m1", Website: "https://platform.minimaxi.com", Region: "domestic", Tags: []string{"domestic", "multimodal"}, ReviewScore: 50, FilterScore: 52, AlignmentScore: 53},
		{ID: "hunyuan:hunyuan-turbo", Provider: "Hunyuan", Name: "Hunyuan Turbo", Version: "hunyuan-turbo", FullName: "Hunyuan / Hunyuan Turbo / hunyuan-turbo", Website: "https://cloud.tencent.com/product/hunyuan", Region: "domestic", Tags: []string{"domestic", "general"}, ReviewScore: 53, FilterScore: 55, AlignmentScore: 56},
		{ID: "ernie:ernie-4.0", Provider: "ERNIE", Name: "ERNIE 4.0", Version: "ernie-4.0", FullName: "ERNIE / ERNIE 4.0 / ernie-4.0", Website: "https://cloud.baidu.com/product/wenxinworkshop", Region: "domestic", Tags: []string{"domestic", "general"}, ReviewScore: 54, FilterScore: 56, AlignmentScore: 57},
		{ID: "stepfun:step-2", Provider: "StepFun", Name: "Step 2", Version: "step-2", FullName: "StepFun / Step 2 / step-2", Website: "https://platform.stepfun.com", Region: "domestic", Tags: []string{"domestic", "reasoning"}, ReviewScore: 48, FilterScore: 50, AlignmentScore: 51},
		{ID: "siliconflow:deepseek-r1", Provider: "SiliconFlow", Name: "DeepSeek R1 via SiliconFlow", Version: "deepseek-r1", FullName: "SiliconFlow / DeepSeek R1 / deepseek-r1", Website: "https://siliconflow.cn/zh-cn/models", Region: "domestic", Tags: []string{"aggregator", "domestic"}, ReviewScore: 40, FilterScore: 42, AlignmentScore: 43},
		{ID: "openrouter:mixtral-large", Provider: "OpenRouter", Name: "Mixtral Large", Version: "mixtral-large", FullName: "OpenRouter / Mixtral Large / mixtral-large", Website: "https://openrouter.ai/models", Region: "global", Tags: []string{"aggregator", "experimentation"}, ReviewScore: 41, FilterScore: 43, AlignmentScore: 44},
		{ID: "xai:grok-3", Provider: "xAI", Name: "Grok 3", Version: "grok-3", FullName: "xAI / Grok 3 / grok-3", Website: "https://docs.x.ai", Region: "global", Tags: []string{"reasoning", "agent"}, ReviewScore: 46, FilterScore: 48, AlignmentScore: 49},
	}
}

func Templates() []domain.Template {
	return []domain.Template{
		{
			ID:          "tpl-full-loop",
			Kind:        "workflow",
			Name:        "全自动闭环",
			Description: "需求分析、实现、测试、部署、修复全链路模板，适合需要完整闭环与自动修复的任务。",
			RecommendedBundle: domain.BundleSuggestion{
				ID:     "bundle-local-delivery",
				Name:   "本地交付闭环",
				Reason: "模板默认覆盖实现、验证、交付与修复链路，适合直接走本地交付闭环。",
			},
			DefaultStages: []domain.StageKind{domain.StageIntent, domain.StageContext, domain.StagePlan, domain.StageResource, domain.StageModel, domain.StageTool, domain.StageImplement, domain.StageResult, domain.StageTest, domain.StageDeploy, domain.StageRepair, domain.StageFinalize},
			DefaultTools:  []string{"analysis", "workspace", "terminal", "tests", "deploy", "logs"},
			Defaults:      map[string]any{"stages": []string{"intent", "context", "plan", "resource", "model", "tool", "implement", "result", "test", "deploy", "repair", "finalize"}, "tools": []string{"analysis", "workspace", "terminal", "tests", "deploy", "logs"}},
		},
		{
			ID:          "tpl-safe-deploy",
			Kind:        "workflow",
			Name:        "安全发布",
			Description: "偏重验证、幂等发布和交付可回退证据，适合稳态上线。",
			RecommendedBundle: domain.BundleSuggestion{
				ID:     "bundle-local-delivery",
				Name:   "本地交付闭环",
				Reason: "模板以验证和发布为主，应优先使用包含构建、质量和交付入口的能力包。",
			},
			DefaultStages: []domain.StageKind{domain.StageIntent, domain.StageContext, domain.StagePlan, domain.StageResource, domain.StageModel, domain.StageTool, domain.StageImplement, domain.StageResult, domain.StageTest, domain.StageDeploy, domain.StageFinalize},
			DefaultTools:  []string{"analysis", "workspace", "terminal", "tests", "deploy", "logs"},
			Defaults:      map[string]any{"stages": []string{"intent", "context", "plan", "resource", "model", "tool", "implement", "result", "test", "deploy", "finalize"}, "tools": []string{"analysis", "workspace", "terminal", "tests", "deploy", "logs"}},
		},
		{
			ID:          "tpl-tools-full",
			Kind:        "toolset",
			Name:        "全工具集",
			Description: "启用工作区、命令、测试、部署和日志工具，适合先把底层能力栈一次性准备齐。",
			RecommendedBundle: domain.BundleSuggestion{
				ID:     "bundle-local-core",
				Name:   "本地核心闭环",
				Reason: "模板重点是补齐底层执行入口，适合先按本地核心闭环起步，再按任务追加交付能力。",
			},
			DefaultStages: []domain.StageKind{domain.StageIntent, domain.StageContext, domain.StagePlan, domain.StageResource, domain.StageModel, domain.StageTool, domain.StageImplement, domain.StageResult, domain.StageTest, domain.StageFinalize},
			DefaultTools:  []string{"analysis", "workspace", "terminal", "tests", "deploy", "logs"},
			Defaults:      map[string]any{"stages": []string{"intent", "context", "plan", "resource", "model", "tool", "implement", "result", "test", "finalize"}, "tools": []string{"analysis", "workspace", "terminal", "tests", "deploy", "logs"}},
		},
	}
}
