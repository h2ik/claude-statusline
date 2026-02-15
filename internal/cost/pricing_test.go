package cost

import (
	"testing"
)

func TestModelPrice_KnownModels(t *testing.T) {
	tests := []struct {
		model      string
		wantInput  float64
		wantOutput float64
		wantCacheW float64
		wantCacheR float64
	}{
		{"claude-opus-4-5-20251101", 5.0, 25.0, 6.25, 0.50},
		{"claude-opus-4-6", 5.0, 25.0, 6.25, 0.50},
		{"claude-sonnet-4-5-20251101", 3.0, 15.0, 3.75, 0.30},
		{"claude-sonnet-4-5-20250929", 3.0, 15.0, 3.75, 0.30},
		{"claude-sonnet-4-20250514", 3.0, 15.0, 3.75, 0.30},
		{"claude-haiku-4-5-20251101", 1.0, 5.0, 1.25, 0.10},
		{"claude-haiku-4-5-20251001", 1.0, 5.0, 1.25, 0.10},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			p := ModelPrice(tt.model)
			if p.InputPerMillion != tt.wantInput {
				t.Errorf("InputPerMillion: got %v, want %v", p.InputPerMillion, tt.wantInput)
			}
			if p.OutputPerMillion != tt.wantOutput {
				t.Errorf("OutputPerMillion: got %v, want %v", p.OutputPerMillion, tt.wantOutput)
			}
			if p.CacheWritePerMillion != tt.wantCacheW {
				t.Errorf("CacheWritePerMillion: got %v, want %v", p.CacheWritePerMillion, tt.wantCacheW)
			}
			if p.CacheReadPerMillion != tt.wantCacheR {
				t.Errorf("CacheReadPerMillion: got %v, want %v", p.CacheReadPerMillion, tt.wantCacheR)
			}
		})
	}
}

func TestModelPrice_UnknownModelReturnsSonnetDefault(t *testing.T) {
	p := ModelPrice("claude-unknown-99")
	if p.InputPerMillion != 3.0 {
		t.Errorf("expected default input rate 3.0, got %v", p.InputPerMillion)
	}
	if p.OutputPerMillion != 15.0 {
		t.Errorf("expected default output rate 15.0, got %v", p.OutputPerMillion)
	}
}

func TestModelPrice_PrefixMatching(t *testing.T) {
	// Future model variant should match by prefix
	p := ModelPrice("claude-opus-4-7-20260301")
	if p.InputPerMillion != 5.0 {
		t.Errorf("expected Opus prefix match input rate 5.0, got %v", p.InputPerMillion)
	}
}

func TestCalculateEntryCost(t *testing.T) {
	// Opus pricing: (1000*5 + 500*25 + 200*6.25 + 10000*0.50) / 1M = 0.02375
	cost := CalculateEntryCost(1000, 500, 200, 10000, "claude-opus-4-5-20251101")
	expected := 0.02375
	if cost < expected-0.0001 || cost > expected+0.0001 {
		t.Errorf("expected %f, got %f", expected, cost)
	}
}

func TestCalculateEntryCost_ZeroTokens(t *testing.T) {
	cost := CalculateEntryCost(0, 0, 0, 0, "claude-opus-4-5-20251101")
	if cost != 0.0 {
		t.Errorf("expected 0.0 for zero tokens, got %f", cost)
	}
}
