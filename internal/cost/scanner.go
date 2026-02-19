package cost

import (
	"fmt"
	"strconv"
	"time"

	"github.com/h2ik/claude-statusline/internal/cache"
)

const transcriptCacheTTL = 5 * time.Minute

// TranscriptScanner computes period costs by scanning Claude Code's native
// JSONL transcript files. Results are cached for 5 minutes.
type TranscriptScanner struct {
	projectsDir string
	cache       *cache.Cache
}

// NewTranscriptScanner creates a scanner reading from the given projects directory.
func NewTranscriptScanner(projectsDir string, c *cache.Cache) *TranscriptScanner {
	return &TranscriptScanner{projectsDir: projectsDir, cache: c}
}

// CalculatePeriod returns the total USD cost from all transcripts within the
// given duration. Results are cached per-duration with a 5 minute TTL.
func (s *TranscriptScanner) CalculatePeriod(duration time.Duration) float64 {
	cacheKey := fmt.Sprintf("transcript-cost:%s", duration.String())

	if data, err := s.cache.Get(cacheKey, transcriptCacheTTL); err == nil {
		if val, err := strconv.ParseFloat(string(data), 64); err == nil {
			return val
		}
	}

	total := ScanTranscripts(s.projectsDir, duration)
	_ = s.cache.Set(cacheKey, []byte(strconv.FormatFloat(total, 'f', 6, 64)), transcriptCacheTTL)
	return total
}

// CalculateToday returns the total USD cost from all transcripts since
// midnight local time today. Results are cached with a 5 minute TTL,
// keyed by the current date so the cache resets at midnight.
func (s *TranscriptScanner) CalculateToday() float64 {
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	cacheKey := fmt.Sprintf("transcript-cost:today:%s", midnight.Format("2006-01-02"))

	if data, err := s.cache.Get(cacheKey, transcriptCacheTTL); err == nil {
		if val, err := strconv.ParseFloat(string(data), 64); err == nil {
			return val
		}
	}

	total := ScanTranscriptsSince(s.projectsDir, midnight)
	_ = s.cache.Set(cacheKey, []byte(strconv.FormatFloat(total, 'f', 6, 64)), transcriptCacheTTL)
	return total
}
