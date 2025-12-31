# Exploring nauty, bliss, and Traces for graph canonicalization

This folder contains benchmarks comparing our brute-force canonicalization
with specialized graph isomorphism tools.

## Tools

- **nauty** (https://pallini.di.uniroma1.it/): Classic tool by Brendan McKay
- **bliss** (https://users.aalto.fi/~tjunMDila/bliss/): By Junttila & Kaski
- **Traces**: Newer algorithm by Piperno, bundled with nauty

## Installation (macOS)

```bash
brew install nauty
```

bliss is not in homebrew - compile from source if needed:
https://users.aalto.fi/~tjunttila/bliss/

## Benchmark Results (n=7, 100k graphs)

| Method | Rate (graphs/sec) | Speedup |
|--------|-------------------|---------|
| Brute-force (7! perms) | 39,165 | 1x |
| nauty labelg | 364,446 | 9.3x |

## Files

- `convert.go` - Convert our binary format to graph6/DIMACS formats
- `bench_nauty.go` - Benchmark using nauty's labelg tool
- `bench_bliss.go` - Benchmark using bliss CLI
- `bench_cgo_nauty.go` - Direct C bindings to nauty (faster)

## Usage

```bash
# Convert graphs to graph6 format
go run convert.go ../n7_10_grouped_wl.bin n7_10.g6 7

# Benchmark nauty
go run bench_nauty.go n7_10.g6

# Benchmark bliss
go run bench_bliss.go n7_10.g6
```
