package service

import (
	"os"
	"sort"

	"autocode-platform/internal/domain"
)

type ProviderProfile struct {
	Provider           string
	CredentialEnv      string
	Website            string
	RecommendedBaseURL string
	Summary            string
}

func providerProfiles() map[string]ProviderProfile {
	return map[string]ProviderProfile{
		"Anthropic":   {Provider: "Anthropic", CredentialEnv: "ANTHROPIC_API_KEY", Website: "https://docs.anthropic.com", RecommendedBaseURL: "https://api.anthropic.com", Summary: "适合长上下文与代码代理"},
		"DeepSeek":    {Provider: "DeepSeek", CredentialEnv: "DEEPSEEK_API_KEY", Website: "https://platform.deepseek.com", RecommendedBaseURL: "https://api.deepseek.com", Summary: "推理与成本平衡较好"},
		"ERNIE":       {Provider: "ERNIE", CredentialEnv: "ERNIE_API_KEY", Website: "https://cloud.baidu.com/product/wenxinworkshop", RecommendedBaseURL: "https://qianfan.baidubce.com", Summary: "适合百度生态与中文场景"},
		"Google":      {Provider: "Google", CredentialEnv: "GEMINI_API_KEY", Website: "https://ai.google.dev", RecommendedBaseURL: "https://generativelanguage.googleapis.com", Summary: "多模态与推理能力较强"},
		"Groq":        {Provider: "Groq", CredentialEnv: "GROQ_API_KEY", Website: "https://console.groq.com/docs/models", RecommendedBaseURL: "https://api.groq.com/openai/v1", Summary: "超低延迟推理体验突出"},
		"Hunyuan":     {Provider: "Hunyuan", CredentialEnv: "HUNYUAN_API_KEY", Website: "https://cloud.tencent.com/product/hunyuan", RecommendedBaseURL: "https://hunyuan.tencentcloudapi.com", Summary: "腾讯云生态接入友好"},
		"MiniMax":     {Provider: "MiniMax", CredentialEnv: "MINIMAX_API_KEY", Website: "https://platform.minimaxi.com", RecommendedBaseURL: "https://api.minimaxi.com", Summary: "多模态与中文对话表现均衡"},
		"Mistral":     {Provider: "Mistral", CredentialEnv: "MISTRAL_API_KEY", Website: "https://docs.mistral.ai", RecommendedBaseURL: "https://api.mistral.ai", Summary: "开放生态与多区域部署友好"},
		"Moonshot":    {Provider: "Moonshot", CredentialEnv: "MOONSHOT_API_KEY", Website: "https://platform.moonshot.cn", RecommendedBaseURL: "https://api.moonshot.cn", Summary: "长上下文中文场景较好"},
		"OpenAI":      {Provider: "OpenAI", CredentialEnv: "OPENAI_API_KEY", Website: "https://platform.openai.com/docs/models", RecommendedBaseURL: "https://api.openai.com", Summary: "通用能力稳定"},
		"OpenRouter":  {Provider: "OpenRouter", CredentialEnv: "OPENROUTER_API_KEY", Website: "https://openrouter.ai/models", RecommendedBaseURL: "https://openrouter.ai/api", Summary: "适合统一聚合多模型"},
		"Qwen":        {Provider: "Qwen", CredentialEnv: "DASHSCOPE_API_KEY", Website: "https://help.aliyun.com/zh/model-studio", RecommendedBaseURL: "https://dashscope.aliyuncs.com", Summary: "代码和中文能力均衡"},
		"SiliconFlow": {Provider: "SiliconFlow", CredentialEnv: "SILICONFLOW_API_KEY", Website: "https://siliconflow.cn/zh-cn/models", RecommendedBaseURL: "https://api.siliconflow.cn", Summary: "聚合式平台，适合灵活切换模型"},
		"StepFun":     {Provider: "StepFun", CredentialEnv: "STEPFUN_API_KEY", Website: "https://platform.stepfun.com", RecommendedBaseURL: "https://api.stepfun.com", Summary: "国内推理与通用任务均衡"},
		"xAI":         {Provider: "xAI", CredentialEnv: "XAI_API_KEY", Website: "https://docs.x.ai", RecommendedBaseURL: "https://api.x.ai", Summary: "偏新锐风格与高上下文任务"},
		"Zhipu":       {Provider: "Zhipu", CredentialEnv: "ZHIPU_API_KEY", Website: "https://open.bigmodel.cn", RecommendedBaseURL: "https://open.bigmodel.cn/api/paas/v4", Summary: "国内生态接入方便"},
	}
}

func configuredProviders() map[string]bool {
	out := map[string]bool{}
	for name, profile := range providerProfiles() {
		_, ok := os.LookupEnv(profile.CredentialEnv)
		out[name] = ok
	}
	return out
}

func decorateModels(models []domain.ModelProfile) []domain.ModelProfile {
	profiles := providerProfiles()
	configured := configuredProviders()
	out := append([]domain.ModelProfile(nil), models...)
	for i := range out {
		profile, ok := profiles[out[i].Provider]
		if !ok {
			continue
		}
		out[i].Configured = configured[out[i].Provider]
		out[i].CredentialEnv = profile.CredentialEnv
		out[i].RecommendedBaseURL = profile.RecommendedBaseURL
	}
	return domain.SortModels(out)
}

func providerStatuses(models []domain.ModelProfile) []domain.ProviderStatus {
	profiles := providerProfiles()
	configured := configuredProviders()
	used := map[string]bool{}
	for _, model := range models {
		used[model.Provider] = true
	}
	out := make([]domain.ProviderStatus, 0, len(used))
	for provider := range used {
		profile := profiles[provider]
		out = append(out, domain.ProviderStatus{
			Provider:      provider,
			Configured:    configured[provider],
			CredentialEnv: profile.CredentialEnv,
			Website:       profile.Website,
			Summary:       profile.Summary,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Provider < out[j].Provider
	})
	return out
}
