# Unknowns Audit - SYN CLI

Unknown unknowns surfaced via Rumsfeld Matrix analysis.

**Generated:** 2026-01-22
**Methodology:** 5-probe parallel analysis (deps, veteran wisdom, ancient patterns, scale cliffs, implicit knowledge)

---

## Executive Summary

| Category | Findings | Critical |
|----------|----------|----------|
| Deps & Packages | 3 | 1 |
| Veteran Wisdom Gaps | 15 | 2 |
| Ancient Patterns | 4 | 1 |
| Scale Cliffs | 10 | 3 |
| Implicit Knowledge | 35+ | 3 |

**Overall Risk Score:** HIGH

---

## Deps & Packages (~40%)

| Issue | Type | Recommendation |
|-------|------|----------------|
| **Retry Logic** (50 LOC) | Reinvented Wheel | Replace with `github.com/cenkalti/backoff` v4 - battle-tested, correct jitter, metrics |
| **String Error Detection** | Fragile | Use `errors.As()` with typed `APIError` instead of string matching |
| **Jitter Formula** | Bug | Current formula allows negative jitter (50/50). Use `rand.Float64()` for additive-only |

**Dependencies Status:** All healthy (lipgloss v1.1.0, cobra v1.10.2, viper v1.21.0)

**Extract Candidate:** Model alias resolution (18 LOC) could be published as `github.com/dotcommander/model-alias` for Synthetic.new ecosystem.

---

## Veteran Wisdom Gaps

Missing production patterns a 20-year veteran would demand:

| Pattern | Why Critical | Missing In | Severity |
|---------|--------------|------------|----------|
| **Circuit Breaker** | Prevent cascading failures when API degrades | `client.go` retry logic | CRITICAL |
| **Goroutine Lifecycle** | Input reader goroutine leaks on Ctrl-C | `chat.go:84-90` | CRITICAL |
| **Health Check** | No pre-flight API connectivity test | None exists | HIGH |
| **Rate Limiting** | No client-side throttle; concurrent use triggers bans | All client methods | HIGH |
| **Response Size Limits** | `io.ReadAll()` allows OOM | 5 locations in client.go | HIGH |
| **File Size Validation** | `-f` flag accepts unlimited files | `client.go:129` | HIGH |
| **Connection Pooling** | New http.Client per command | `root.go:235` | MEDIUM |
| **Timeout Strategy** | Scattered timeouts (30s, 1min, 5min) | Inconsistent | MEDIUM |
| **Context Propagation** | Spinner ignores parent ctx | `chat.go:127` | MEDIUM |
| **Structured Logging** | No request/correlation IDs | `client.go` | MEDIUM |
| **Error Categorization** | 401/403 triggers retries | `isRetryableError()` | MEDIUM |
| **Metrics/Observability** | No latency/token tracking | None exists | MEDIUM |
| **SIGHUP Handling** | Config reload would crash | `chat.go:63` | LOW |

---

## Ancient Patterns

| Pattern | Age | Modern Alternative | Location |
|---------|-----|-------------------|----------|
| **Hardcoded API Key** | Security anti-pattern | Environment-only, no defaults | `config.go:12` |
| **interface{}** | Pre-Go 1.18 | `any` keyword | `root.go:278`, vision.go |
| **Pointer Helpers** | Pre-Go 1.22 | Generic `ptr[T]` | `types.go:149-151` |
| **Type Assertion** | Error-prone | `errors.As()` | `root.go:198` |

**Good News:** Codebase uses modern Go (1.25.4) - `log/slog`, `math/rand/v2`, `os.ReadFile`, `io.ReadAll` all present.

---

## Scale Cliffs

Where the system breaks under load:

| Cliff | Trigger | Breaks At |
|-------|---------|-----------|
| **Goroutine Leak** | Ctrl-C during chat | ~100 leaked goroutines/session |
| **io.ReadAll OOM** | Response >500MB | Process death |
| **File Size Bomb** | `syn -f 10GB.dat` | Process OOM |
| **No Rate Limit** | Script loop 1000x | Instant API ban |
| **Retry Storm** | API outage + 10 clients | Hammers dead API |
| **HTTP Port Exhaustion** | 100+ parallel commands | TCP port depletion |
| **Context Slice O(n)** | 100+ chat turns | Memory pressure |
| **Embedding Array** | 10k texts per call | OOM or API rejection |
| **String Retry Match** | Error format change | Silent retry failure |
| **Base64 Image** | 1GB image → 1.3GB RAM | Vision OOM |

---

## Implicit Knowledge

Critical knowledge existing only in code, not documentation:

### Security

| Knowledge | Where | Risk |
|-----------|-------|------|
| **Hardcoded API key** | `config.go:12` | Token exposed in source |
| **Bearer in plain header** | `client.go:212` | Depends on HTTPS |

### Magic Numbers

| Value | Location | Why This Value? |
|-------|----------|-----------------|
| 20 messages | `chat.go:72` | Context window limit (undocumented) |
| 3 attempts | `config.go:19` | Retry count |
| 30s max backoff | `config.go:21` | Retry ceiling |
| 4096 tokens | `client.go:511` | Vision hardcoded |
| 12.5% jitter | `client.go:343` | Backoff randomness |
| 80ms spinner | `chat.go:49` | Animation frame rate |

### Timeout Strategy (Undocumented)

| Endpoint | Timeout | Location |
|----------|---------|----------|
| Search | 30s | `search.go:67` |
| Embed | 1min | `embed.go:54` |
| Chat | 5min | `root.go:269` |
| Vision | 5min | `vision.go:59` |
| HTTP default | 60s | `client.go:59` |

### API Assumptions

| Assumption | Location | Risk If Wrong |
|------------|----------|---------------|
| Retryable codes: 429,502,503,504 only | `client.go:316-319` | 500/501 fail immediately |
| Search URL: `/v2/search` (not `/openai/v1`) | `client.go:579` | String replace breaks |
| Vision always Qwen3-VL | `client.go:484` | No model flexibility |
| Stream mode: always false | `client.go:173` | No streaming support |
| System prompt on ALL requests | `client.go:154-156` | Affects all models |

### Incomplete Features

| Feature | Evidence | Status |
|---------|----------|--------|
| Anthropic API | `config.go:14` URL configured | Never called |
| Streaming | `Stream: false` hardcoded | Not implemented |

### Environment Variables

Pattern: `SYN_` prefix + dot-to-underscore

| Config Key | Env Var |
|------------|---------|
| `api.key` | `SYN_API_KEY` |
| `api.base_url` | `SYN_API_BASE_URL` |
| `chat.temperature` | `SYN_CHAT_TEMPERATURE` |

---

## Suggested Actions

| # | Action | Category | Impact | Effort |
|---|--------|----------|--------|--------|
| 1 | **Remove hardcoded API key** | Security | Prevent leak | 5min |
| 2 | **Add io.LimitReader to all responses** | Scale | Prevent OOM | 30min |
| 3 | **Fix goroutine lifecycle in chat.go** | Veteran | Prevent leak | 30min |
| 4 | **Add file size validation** | Scale | Prevent OOM | 15min |
| 5 | **Replace retry logic with cenkalti/backoff** | Deps | Better jitter, metrics | 1hr |
| 6 | **Add client-side rate limiter** | Scale | Prevent API ban | 30min |
| 7 | **Add circuit breaker (gobreaker)** | Veteran | Prevent retry storms | 1hr |
| 8 | **Document all magic numbers** | Implicit | Maintainability | 30min |
| 9 | **Add health check command** | Veteran | Better UX | 30min |
| 10 | **Configure HTTP connection pool** | Scale | Resource efficiency | 15min |

---

## Pre-mortem Analysis

*"It's 6 months later and syn failed in production. Why?"*

| Predicted Failure | Likelihood | Impact | Prevention |
|-------------------|------------|--------|------------|
| API key leaked via hardcoded default | HIGH | HIGH | Remove from code, env-only |
| Long chat sessions crash from goroutine leak | HIGH | MEDIUM | errgroup lifecycle |
| Batch script triggers API ban | MEDIUM | HIGH | Rate limiter |
| Large vision upload OOMs process | MEDIUM | MEDIUM | File size check |
| API outage causes retry storm | MEDIUM | HIGH | Circuit breaker |
| Response body exhausts memory | LOW | HIGH | io.LimitReader |

---

## Success Criteria

- [x] True unknowns surfaced (not answerable by single command)
- [x] Both dimensions covered (introspective + extrospective)
- [x] Risk calibrated by severity
- [x] Pre-mortem included
- [x] Ecosystem compared (cenkalti/backoff, gobreaker alternatives identified)
