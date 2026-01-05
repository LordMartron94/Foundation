# foundation

Foundation library providing core utilities and building blocks.

## Packages

### `benchmarking`

Provides helpful utilities for benchmarking Go code with comprehensive metrics collection.

#### FLOPS (Floating Point Operations Per Second)

FLOPS is a measure of computational throughput, representing the number of floating point mathematical operations performed per second. This metric is useful for evaluating the computational performance of algorithms that perform significant numerical work.

**What Counts as a Floating Point Operation?**

A floating point operation is a **mathematical operation** performed on floating point numbers, not a CPU instruction. Each of the following counts as one floating point operation:

- Addition: `a + b` (where a and b are floating point)
- Subtraction: `a - b`
- Multiplication: `a * b`
- Division: `a / b`
- Square root: `math.Sqrt(x)`
- Trigonometric functions: `math.Sin(x)`, `math.Cos(x)`, `math.Tan(x)`, etc.
- Exponential functions: `math.Exp(x)`, `math.Log(x)`, `math.Pow(x, y)`, etc.
- Other mathematical functions: `math.Abs(x)`, `math.Min(a, b)`, `math.Max(a, b)`, etc.

**Important Notes:**

- **Not CPU instructions**: FLOPS counts mathematical operations, not the underlying CPU instructions. A single mathematical operation may require multiple CPU instructions to execute.
- **Per operation**: When configuring `FLOPSPerOp` in `BenchmarkMetricsConfig`, count the total number of floating point mathematical operations performed in a single benchmark operation (one iteration of your benchmark function).
- **Accuracy**: For accurate FLOPS reporting, carefully count all floating point mathematical operations in your benchmark. This includes operations in loops, function calls, and any mathematical computations.

**Example:**

```go
// If your benchmark performs 1024 floating point multiplications per operation:
BenchmarkWithMetricsConfig(b, BenchmarkMetricsConfig{FLOPSPerOp: 1024.0}, prepareFn, testFn, cleanupFn)
```

This will report `flops/sec` which is calculated as `FLOPSPerOp * ops/sec`.

#### Manual Memory Throughput

For benchmarks using manual memory management (memcore, memforge, etc.) outside the Go GC, the standard `bytes/op` metric from `b.ReportAllocs()` will not accurately reflect memory usage. Use `BytesPerOp` in `BenchmarkMetricsConfig` to specify the actual bytes allocated per operation.

**When to Use Manual Bytes/Op:**

- Benchmarks using `memcore`, `memforge`, or other manual memory allocators
- Benchmarks that allocate memory outside Go's GC heap
- When you need accurate memory throughput measurements for manual memory management

**How It Works:**

- `BytesPerOp` is reported as `manual.bytes/op` metric
- Memory throughput calculation prefers `manual.bytes/op` over standard `bytes/op` when available
- Both metrics are displayed separately in Anvil's console and Web UI

**Example:**

```go
// If your benchmark allocates 4096 bytes per operation using manual memory:
BenchmarkWithMetricsConfig(b, BenchmarkMetricsConfig{BytesPerOp: 4096.0}, prepareFn, testFn, cleanupFn)
```

This will report `manual.bytes/op` and use it for accurate memory throughput calculation (`manual.bytes/op * ops/sec = bytes/sec`).
