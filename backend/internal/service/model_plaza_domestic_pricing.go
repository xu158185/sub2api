package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ModelPlazaDomesticPricing is an editable domestic official price in CNY per
// million tokens. A nil field means the model has no published value for it.
type ModelPlazaDomesticPricing struct {
	InputPrice         *float64 `json:"input_price,omitempty"`
	OutputPrice        *float64 `json:"output_price,omitempty"`
	CacheWritePrice    *float64 `json:"cache_write_price,omitempty"`
	CacheWrite1hPrice  *float64 `json:"cache_write_1h_price,omitempty"`
	CacheWrite12hPrice *float64 `json:"cache_write_12h_price,omitempty"`
	CacheReadPrice     *float64 `json:"cache_read_price,omitempty"`
	Source             string   `json:"source,omitempty"`
	EffectiveDate      string   `json:"effective_date,omitempty"`
}

func (s *SettingService) GetModelPlazaDomesticPricing(ctx context.Context) (map[string]ModelPlazaDomesticPricing, error) {
	if s == nil || s.settingRepo == nil {
		return map[string]ModelPlazaDomesticPricing{}, nil
	}
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyModelPlazaDomesticPricing)
	if err != nil {
		return map[string]ModelPlazaDomesticPricing{}, nil
	}
	result := map[string]ModelPlazaDomesticPricing{}
	if strings.TrimSpace(raw) == "" {
		return result, nil
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("decode model plaza domestic pricing: %w", err)
	}
	return result, nil
}

func (s *SettingService) SetModelPlazaDomesticPricing(ctx context.Context, pricing map[string]ModelPlazaDomesticPricing) error {
	if s == nil || s.settingRepo == nil {
		return fmt.Errorf("setting repository is unavailable")
	}
	normalized := make(map[string]ModelPlazaDomesticPricing, len(pricing))
	for model, value := range pricing {
		model = strings.TrimSpace(model)
		if model == "" {
			return fmt.Errorf("model name is required")
		}
		for _, price := range []*float64{value.InputPrice, value.OutputPrice, value.CacheWritePrice, value.CacheWrite1hPrice, value.CacheWrite12hPrice, value.CacheReadPrice} {
			if price != nil && (*price < 0 || *price > 1e9) {
				return fmt.Errorf("price for %s is out of range", model)
			}
		}
		normalized[model] = value
	}
	raw, err := json.Marshal(normalized)
	if err != nil {
		return fmt.Errorf("encode model plaza domestic pricing: %w", err)
	}
	return s.settingRepo.Set(ctx, SettingKeyModelPlazaDomesticPricing, string(raw))
}
