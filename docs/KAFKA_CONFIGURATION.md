# Kafka Configuration Guide

## Overview

All Kafka producer and consumer settings are now centralized in `config/config.yaml` for easy management and tuning without code changes.

## Configuration Structure

```yaml
kafka:
  brokers: localhost:9099
  groupId: user-worker-group

  # Producer Configuration
  producer:
    requiredAcks: all         # Durability level
    retryMax: 5               # Retry attempts for transient errors
    compression: none         # Compression algorithm
    maxMessageBytes: 1000000  # Max message size (1MB)

  # Consumer Configuration
  consumer:
    sessionTimeoutMs: 10000      # Session timeout (10s)
    heartbeatIntervalMs: 3000    # Heartbeat interval (3s)
    maxProcessingTimeMs: 300000  # Max processing time (5min)

  # Retry/DLQ Configuration
  retry:
    retrySuffix: ".retry"
    dlqSuffix: ".dlq"
    topics:
      user_created:
        enableRetry: true
        maxAttempts: 3
        backoffMs: 2000
```

## Producer Configuration

### requiredAcks

Controls durability vs throughput trade-off:

| Value | Behavior | Durability | Throughput | Use Case |
|-------|----------|------------|------------|----------|
| `all` | Wait for all in-sync replicas | **Highest** ⭐⭐⭐ | Low | Critical data (orders, payments) |
| `local` | Wait for leader only | Medium ⭐⭐ | Medium | Standard events |
| `none` | Fire and forget | Lowest ⭐ | **Highest** | Logs, metrics, non-critical data |

**Default:** `all` (most durable)

**Example:**
```yaml
producer:
  requiredAcks: all  # No message loss even if leader fails
```

### retryMax

Number of retry attempts for transient errors (network issues, broker not available).

| Value | Behavior | Use Case |
|-------|----------|----------|
| 0 | No retries | Testing |
| 3 | 3 retries | Fast-fail scenarios |
| **5** | 5 retries | **Recommended default** ⭐ |
| 10+ | Many retries | High-latency networks |

**Default:** 5

**Example:**
```yaml
producer:
  retryMax: 5  # Retry up to 5 times before failing
```

### compression

Message compression algorithm.

| Value | Ratio | CPU | Speed | Use Case |
|-------|-------|-----|-------|----------|
| `none` | 1x | None | **Fastest** ⚡ | Small messages, local dev |
| `gzip` | **Best** 🏆 | High | Slow | Large payloads, limited bandwidth |
| `snappy` | Good | Low | **Fast** | **Recommended for production** ⭐ |
| `lz4` | Good | Low | **Fastest** | High-throughput systems |
| `zstd` | **Better** | Medium | Fast | Modern systems (Kafka 2.1+) |

**Default:** `none`

**Recommendations:**
- Development: `none` (easier debugging)
- Production: `snappy` (best balance)
- Large messages (>10KB): `gzip` or `zstd`

**Example:**
```yaml
producer:
  compression: snappy  # Good compression with low CPU cost
```

### maxMessageBytes

Maximum message size in bytes.

| Value | Description | Use Case |
|-------|-------------|----------|
| 1000000 | 1 MB | **Default** ⭐ |
| 5000000 | 5 MB | Large JSON payloads |
| 10000000 | 10 MB | File attachments |

**Important:**
- Broker's `message.max.bytes` must be >= this value
- Topic's `max.message.bytes` must be >= this value
- Larger messages = higher latency

**Example:**
```yaml
producer:
  maxMessageBytes: 1000000  # 1MB max
```

## Consumer Configuration

### sessionTimeoutMs

Maximum time between heartbeats before consumer is considered dead and rebalancing occurs.

| Value | Description | Use Case |
|-------|-------------|----------|
| 6000 | 6 seconds | Fast rebalancing (dev) |
| **10000** | 10 seconds | **Recommended default** ⭐ |
| 30000 | 30 seconds | Slow/unreliable networks |

**Trade-offs:**
- **Lower**: Faster failure detection, more rebalancing
- **Higher**: Fewer rebalances, tolerates network hiccups

**Default:** 10000 (10 seconds)

**Example:**
```yaml
consumer:
  sessionTimeoutMs: 10000  # Consumer must heartbeat every 10s
```

### heartbeatIntervalMs

How often consumer sends heartbeats to broker.

**Rule:** Must be < `sessionTimeoutMs / 3`

| sessionTimeoutMs | Recommended heartbeatIntervalMs |
|------------------|---------------------------------|
| 6000 | 2000 (2s) |
| **10000** | **3000 (3s)** ⭐ |
| 30000 | 10000 (10s) |

**Default:** 3000 (3 seconds)

**Example:**
```yaml
consumer:
  sessionTimeoutMs: 10000
  heartbeatIntervalMs: 3000  # 3 heartbeats per session timeout
```

### maxProcessingTimeMs

Maximum time a message batch can be processed before consumer is considered stuck.

| Value | Description | Use Case |
|-------|-------------|----------|
| 60000 | 1 minute | Fast message processing |
| **300000** | 5 minutes | **Default** ⭐ |
| 600000 | 10 minutes | Long-running operations |

**Important:**
- If your handler takes longer than this, consumer will be kicked from group
- Set this based on your slowest expected handler

**Example:**
```yaml
consumer:
  maxProcessingTimeMs: 300000  # Handlers have up to 5 minutes
```

## Retry/DLQ Configuration

### Topic-Level Retry Settings

```yaml
retry:
  retrySuffix: ".retry"  # Retry topic suffix
  dlqSuffix: ".dlq"      # Dead letter queue suffix
  topics:
    user_created:          # Base topic name
      enableRetry: true    # Enable retry mechanism
      maxAttempts: 3       # Max retry attempts
      backoffMs: 2000      # 2 seconds between retries
```

### enableRetry

| Value | Behavior |
|-------|----------|
| `true` | Failed messages → retry topic → DLQ (after max attempts) |
| `false` | Failed messages → skip (logged) |

### maxAttempts

Number of retry attempts before sending to DLQ.

| Value | Total Tries | Use Case |
|-------|-------------|----------|
| 0 | 1 (no retry) | No retry needed |
| **3** | 4 (1 + 3 retries) | **Recommended** ⭐ |
| 5 | 6 (1 + 5 retries) | Transient failures |
| 10+ | 11+ | Persistent errors |

### backoffMs

Delay between retry attempts (in milliseconds).

| Value | Delay | Use Case |
|-------|-------|----------|
| 1000 | 1 second | Fast retry |
| **2000** | 2 seconds | **Default** ⭐ |
| 5000 | 5 seconds | Rate-limited APIs |
| 10000+ | 10+ seconds | External service cooldown |

**Example Flow (maxAttempts=3, backoffMs=2000):**

```
Message → Handler → ❌ Error (attempt 0)
  ↓ (wait 2s)
Retry Topic → Handler → ❌ Error (attempt 1)
  ↓ (wait 2s)
Retry Topic → Handler → ❌ Error (attempt 2)
  ↓ (wait 2s)
Retry Topic → Handler → ❌ Error (attempt 3)
  ↓
DLQ Topic (final destination)
```

## Configuration Examples

### Development Environment

Fast, loose, easy debugging:

```yaml
kafka:
  brokers: localhost:9099
  groupId: dev-worker-group
  producer:
    requiredAcks: local     # Faster, less durable
    retryMax: 3             # Fail fast
    compression: none       # No compression (easier to debug)
    maxMessageBytes: 1000000
  consumer:
    sessionTimeoutMs: 6000       # Fast rebalancing
    heartbeatIntervalMs: 2000
    maxProcessingTimeMs: 60000   # 1 minute max
  retry:
    topics:
      user_created:
        enableRetry: true
        maxAttempts: 2      # Quick retry
        backoffMs: 1000     # 1 second
```

### Production Environment

Durable, reliable, optimized:

```yaml
kafka:
  brokers: kafka-1:9092,kafka-2:9092,kafka-3:9092
  groupId: prod-worker-group
  producer:
    requiredAcks: all           # Maximum durability ⭐
    retryMax: 5                 # Retry transient errors
    compression: snappy         # Good compression ⭐
    maxMessageBytes: 5000000    # 5MB for larger payloads
  consumer:
    sessionTimeoutMs: 10000      # Balanced
    heartbeatIntervalMs: 3000
    maxProcessingTimeMs: 300000  # 5 minutes
  retry:
    topics:
      user_created:
        enableRetry: true
        maxAttempts: 3
        backoffMs: 2000
      order_created:
        enableRetry: true
        maxAttempts: 5           # More retries for critical events
        backoffMs: 5000          # Longer backoff
```

### High-Throughput Environment

Optimize for speed:

```yaml
kafka:
  brokers: kafka-cluster:9092
  groupId: high-throughput-worker
  producer:
    requiredAcks: local         # Faster (leader only)
    retryMax: 3
    compression: lz4            # Fastest compression ⚡
    maxMessageBytes: 1000000
  consumer:
    sessionTimeoutMs: 30000      # Tolerate slow processing
    heartbeatIntervalMs: 10000
    maxProcessingTimeMs: 600000  # 10 minutes
  retry:
    topics:
      metrics:
        enableRetry: false       # No retry for metrics (acceptable loss)
      events:
        enableRetry: true
        maxAttempts: 2           # Quick retry
        backoffMs: 1000
```

## Tuning Guidelines

### When to Increase sessionTimeoutMs

✅ Increase if you see frequent rebalancing:
```
Consumer group rebalanced
Session context cancelled, exiting consume loop
```

✅ Increase if network is slow/unreliable

❌ Don't increase too much (slows failure detection)

### When to Increase maxProcessingTimeMs

✅ Increase if handlers are slow but legitimate:
```
Handler takes 2-3 minutes to process
External API calls take time
Database operations are slow
```

❌ Don't use as band-aid for inefficient handlers (optimize instead!)

### When to Change requiredAcks

| Scenario | Recommendation |
|----------|----------------|
| Critical data (payments, orders) | `all` ⭐ |
| Standard events | `local` |
| Logs, metrics, analytics | `none` |
| Development/testing | `local` or `none` |

### When to Enable Compression

| Avg Message Size | Recommendation |
|------------------|----------------|
| < 1 KB | `none` |
| 1-10 KB | `snappy` ⭐ |
| 10-100 KB | `snappy` or `lz4` |
| > 100 KB | `gzip` or `zstd` |

## Monitoring

After changing config, monitor these metrics:

1. **Producer Metrics**
   - Request latency (ms)
   - Produce rate (msg/s)
   - Error rate
   - Compression ratio

2. **Consumer Metrics**
   - Consumer lag
   - Rebalance rate
   - Processing time per message
   - Error rate

3. **Retry/DLQ Metrics**
   - Messages in retry topic
   - Messages in DLQ
   - Retry success rate

## Best Practices

1. **Start with defaults** - Only tune when you have metrics showing issues
2. **Test changes in staging** - Config changes affect reliability
3. **Document changes** - Track why each value was chosen
4. **Monitor after changes** - Watch for unintended side effects
5. **Use compression in production** - `snappy` is almost always worth it

## Troubleshooting

### Consumer keeps rebalancing

**Symptoms:**
```
Consumer group rebalanced
Session context cancelled
```

**Solutions:**
- Increase `sessionTimeoutMs` (e.g., 10s → 30s)
- Decrease `heartbeatIntervalMs` (e.g., 3s → 2s)
- Check if handlers are blocking too long

### Messages timing out

**Symptoms:**
```
Context deadline exceeded
Consumer kicked from group
```

**Solutions:**
- Increase `maxProcessingTimeMs`
- Optimize slow handlers
- Add more parallel consumers

### Low throughput

**Solutions:**
- Enable compression (`snappy` or `lz4`)
- Increase `maxMessageBytes` if messages are large
- Use `requiredAcks: local` for non-critical data
- Add more partitions to topic

### High latency

**Solutions:**
- Disable compression (`none`)
- Use `requiredAcks: local` or `none`
- Reduce `retryMax` for faster failure
- Optimize handler code

## Summary

| Config | Default | Production Recommendation |
|--------|---------|---------------------------|
| **requiredAcks** | `all` | `all` (critical data), `local` (standard) |
| **retryMax** | 5 | 5 |
| **compression** | `none` | `snappy` ⭐ |
| **maxMessageBytes** | 1MB | 1-5MB |
| **sessionTimeoutMs** | 10s | 10s |
| **heartbeatIntervalMs** | 3s | 3s |
| **maxProcessingTimeMs** | 5min | 5min |
| **retry.maxAttempts** | 3 | 3-5 |
| **retry.backoffMs** | 2s | 2-5s |

All settings are in `config/config.yaml` - no code changes needed! 🎉
