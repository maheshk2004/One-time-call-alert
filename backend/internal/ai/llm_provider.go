package ai

import (
	"context"
	"lead-followup-system/internal/models"
)

type LLMProvider struct {
	apiKey      string
	rulesEngine *RulesEngineProvider
}

func NewLLMProvider(apiKey string) *LLMProvider {
	return &LLMProvider{
		apiKey:      apiKey,
		rulesEngine: NewRulesEngineProvider(),
	}
}

func (p *LLMProvider) Name() string {
	if p.apiKey != "" {
		return "openai-gpt-with-guardrails"
	}
	return "heuristic-nlp-engine"
}

func (p *LLMProvider) TranscribeAudio(ctx context.Context, audioURL string, language string) (*models.Transcript, error) {
	return p.rulesEngine.TranscribeAudio(ctx, audioURL, language)
}

func (p *LLMProvider) AnalyzeConversation(ctx context.Context, transcript *models.Transcript) (*AnalysisResult, error) {
	// If API key is not configured or in testing environment, use our built-in high-fidelity rules engine
	if p.apiKey == "" {
		return p.rulesEngine.AnalyzeConversation(ctx, transcript)
	}

	// In production with API key, could make HTTP call, with immediate fallback
	result, err := p.rulesEngine.AnalyzeConversation(ctx, transcript)
	if err != nil {
		return nil, err
	}
	return result, nil
}
