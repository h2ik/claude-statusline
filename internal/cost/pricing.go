package cost

// Pricing represents per-million-token rates for a model.
type Pricing struct {
	InputPerMillion      float64
	OutputPerMillion     float64
	CacheWritePerMillion float64
	CacheReadPerMillion  float64
}

var pricingTable = map[string]Pricing{
	"claude-opus-4-5-20251101":   {5.0, 25.0, 6.25, 0.50},
	"claude-opus-4-6":            {5.0, 25.0, 6.25, 0.50},
	"claude-sonnet-4-5-20251101": {3.0, 15.0, 3.75, 0.30},
	"claude-sonnet-4-5-20250929": {3.0, 15.0, 3.75, 0.30},
	"claude-sonnet-4-20250514":   {3.0, 15.0, 3.75, 0.30},
	"claude-haiku-4-5-20251101":  {1.0, 5.0, 1.25, 0.10},
	"claude-haiku-4-5-20251001":  {1.0, 5.0, 1.25, 0.10},
}

var defaultPricing = Pricing{3.0, 15.0, 3.75, 0.30}

var prefixPricing = []struct {
	prefix  string
	pricing Pricing
}{
	{"claude-opus", Pricing{5.0, 25.0, 6.25, 0.50}},
	{"claude-sonnet", Pricing{3.0, 15.0, 3.75, 0.30}},
	{"claude-haiku", Pricing{1.0, 5.0, 1.25, 0.10}},
}

// ModelPrice returns per-million-token pricing for a model identifier.
// Checks exact matches, then prefix matches, then falls back to Sonnet-tier default.
func ModelPrice(model string) Pricing {
	if p, ok := pricingTable[model]; ok {
		return p
	}
	for _, pp := range prefixPricing {
		if len(model) >= len(pp.prefix) && model[:len(pp.prefix)] == pp.prefix {
			return pp.pricing
		}
	}
	return defaultPricing
}

// CalculateEntryCost computes the USD cost for a single transcript entry.
func CalculateEntryCost(inputTokens, outputTokens, cacheWriteTokens, cacheReadTokens int, model string) float64 {
	p := ModelPrice(model)
	return (float64(inputTokens)*p.InputPerMillion +
		float64(outputTokens)*p.OutputPerMillion +
		float64(cacheWriteTokens)*p.CacheWritePerMillion +
		float64(cacheReadTokens)*p.CacheReadPerMillion) / 1_000_000
}
