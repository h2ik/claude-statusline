# Future Enhancements

Potential improvements requiring significant architectural changes.

## Session State Snapshots

**Current limitation:** `burn_rate` and `block_projection` compute from
single-point-in-time data. This provides instantaneous metrics but
cannot show trends or acceleration.

**Enhancement:** Add session snapshot recording — periodic writes to a
JSONL state file (similar to cost history) that records:
- Timestamp
- Context usage percentage
- Total cost
- Token counts
- Rate limit utilization

**Benefits:**
- `burn_rate` could show acceleration/deceleration trends
- `block_projection` could extrapolate from observed usage curves
- Time-to-limit predictions based on actual session behavior
- Historical context usage patterns

**Implementation approach:**
1. Add `internal/session/` package with `Snapshot` and `StateWriter` types
2. Write snapshots every N seconds (e.g., 30s) during statusline renders
3. Components read recent snapshots from JSONL for trend calculation
4. Auto-compact old session files (keep last 24h)

**Trade-offs:**
- Adds I/O overhead on every render (mitigated by buffered writes)
- Increases storage requirements (~1KB per session)
- More accurate metrics vs. simpler single-point computation

**Decision:** Deferred for v1. Single-point metrics are sufficient for
initial release. Revisit if users request trend-based features.
