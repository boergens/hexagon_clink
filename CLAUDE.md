# Penny Graph Enumeration Project

## Overview

This project enumerates **penny graphs** (also called unit coin graphs) - graphs that can be realized by placing non-overlapping unit circles in the 2D plane, where vertices are circle centers and edges connect touching circles.

Key constraints for penny graphs:
- **Edge distance = 1** (touching circles)
- **Non-edge distance > 1** (non-overlapping)
- **No K4 subgraph** (4 mutually touching circles impossible in 2D)
- **Max degree ≤ 6** (hexagonal packing limit)

## Pipeline

The enumeration pipeline has these stages:

1. **Generate candidates** - Enumerate graphs with filters (connected, no isolated vertices, max degree ≤6, no K4)
2. **Remove isomorphisms** - Use nauty's `shortg` to find canonical representatives
3. **Verify penny embedding** - Gradient descent to find valid 2D embedding
4. **Filter maximal** - Remove graphs that are subgraphs of larger penny graphs

## Tools

### Core Pipeline

| Tool | Description | Usage |
|------|-------------|-------|
| `pipeline_nauty` | All-in-one streaming pipeline for large n | `./pipeline_nauty -n 9 -out n9_unique.g6` |
| `generate_edges` | Generate candidate graphs with k edges | `./generate_edges <n> <edges> <output.bin>` |
| `verify_penny` | Check if graphs are valid penny graphs | `./verify_penny -n 8 -in input.g6 -out penny.g6` |
| `filter_maximal` | Keep only maximal graphs (not subgraph of another) | `./filter_maximal -n 8 -out maximal.g6 *.g6` |

### Older Pipeline (separate steps)

| Tool | Description |
|------|-------------|
| `refine_hash` | Group graphs by fingerprint (degree, triangles) |
| `wl_refine` | Further refine groups using Weisfeiler-Lehman coloring |
| `canonicalize` | Brute-force n! permutation canonicalization |

### Utilities

| Tool | Description |
|------|-------------|
| `explore_nauty/convert` | Convert binary ↔ graph6 format |
| `plot_penny.py` | Visualize penny graphs (requires scipy, matplotlib) |

## File Formats

- **Binary (.bin)** - 4 or 8 bytes per graph (uint32/uint64 edge bitmask)
- **Graph6 (.g6)** - Text format used by nauty, one graph per line

Graph representation: Edge bitmask where bit `edgeIndex[i][j]` indicates edge (i,j).

## Key Results

### n=8 (8 vertices)
- 5,481 unique candidate graphs (edges 7-18)
- 677 valid penny graphs
- **9 maximal penny graphs** (1 with 14 edges, 8 with 13 edges)
- Results in `explore_nauty/n8_*_penny.g6` and `n8_maximal_penny.png`

### n=9 (in progress)
- ~60 billion candidate combinations to check
- Pipeline running with 10M batch size
- Estimated ~15 hours total

## Dependencies

- **Go** - All tools written in Go
- **nauty** - `brew install nauty` - provides `shortg` for isomorphism
- **Python** (optional) - For plotting, needs scipy and matplotlib from `../hexagon_clink/venv`

## Quick Start

```bash
# Build all tools
go build -o generate_edges generate_edges.go
go build -o verify_penny verify_penny.go
go build -o filter_maximal filter_maximal.go
go build -o pipeline_nauty pipeline_nauty.go
(cd explore_nauty && go build -o convert convert.go)

# Run full pipeline for n=8
./pipeline_nauty -n 8 -out n8_unique.g6

# Verify penny graphs
./verify_penny -n 8 -in n8_unique.g6 -out n8_penny.g6

# Find maximal
./filter_maximal -n 8 -out n8_maximal.g6 n8_penny.g6

# Plot results
source ../hexagon_clink/venv/bin/activate
python3 plot_penny.py 8 n8_maximal.g6 n8_maximal.png
```

## Notes

- Penny graph verification uses gradient descent with multiple random restarts
- For large n, the pipeline writes batches to temp files and runs nauty incrementally
- The `filter_maximal` tool checks isomorphic subgraph relationships (O(n!) per pair)
- Reference implementation in `../hexagon_clink/explore_n7_k2/`
