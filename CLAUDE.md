# Penny Graph & Hexagon Clink Project

## Coding Conventions

- **Go binaries**: Always use `.out` extension when compiling: `go build -o foo.out foo.go`

## Overview

This project has three related threads:

### 1. Penny Graph Enumeration
Enumerating **penny graphs** (unit coin graphs) - graphs realizable by placing non-overlapping unit circles in 2D, where vertices are circle centers and edges connect touching circles.

Key constraints:
- **Edge distance = 1** (touching circles)
- **Non-edge distance > 1** (non-overlapping)
- **No K4 subgraph** (4 mutually touching circles impossible in 2D)
- **Max degree ≤ 6** (hexagonal packing limit)

### 2. Hexagon Clink Problem
Given n items, find the minimum number of **penny spiral** arrangements needed so every pair of items is adjacent at least once.

#### Spiral Construction Rules
1. Node 0 at center
2. Each new node must be **adjacent to the previous node** (spiral property)
3. Among valid positions, prefer more total contacts, then closer to origin

This produces nodes 1-6 as a ring around center, then nodes 7+ spiral outward.

#### Adjacency Counts (OEIS A047932)

| n | edges | pairs | lower bound |
|---|-------|-------|-------------|
| 7 | 12 | 21 | 2 |
| 8 | 14 | 28 | 2 |
| 9 | 16 | 36 | 3 |
| 15 | 31 | 105 | 4 |
| 17 | 36 | 136 | 4 |
| 19 | 40 | 171 | 5 |

Lower bound = ceil(pairs / edges)

#### Known Results

| n | min arrangements | status |
|---|------------------|--------|
| 3 | 1 | proven |
| 4-6 | 2 | proven |
| 7-12 | 3 | verified |
| 13-16 | 4 | verified |
| 17 | 4 or 5? | **in progress** |

#### Perfect k-Arrangements
k arrangements where consecutive ones share **no edges** - maximizes coverage efficiency.

For n=17: if we find perfect-3 (arr0, arr1, arr2 with zero overlap), we cover 3×36=108 pairs, leaving 28 for arr3.

### 3. Polyiamond Enumeration
Enumerating **polyiamonds** - shapes made of edge-connected equilateral triangles (OEIS A000577).

Used to find unit-distance graph embeddings on the triangular lattice.

| n triangles | polyiamonds |
|-------------|-------------|
| 1-3 | 1 |
| 4 | 3 |
| 5 | 4 |
| 6 | 12 |
| 7 | 24 |
| 8 | 66 |
| 9 | 160 |
| 10 | 448 |
| 13 | 9,235 |

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

### n=9 (9 vertices)
- 88,958 unique candidate graphs
- 3,136 valid penny graphs
- **16 maximal penny graphs** (4 with 16 edges, 12 with 15 edges)
- Pipeline: ~60B combinations checked in 12h46m

### Summary Table

| n | Candidates | Penny | Maximal | Max Edges | Time |
|---|------------|-------|---------|-----------|------|
| 8 | 5,481 | 677 | 9 | 14 | ~4 min |
| 9 | 88,958 | 3,136 | 16 | 16 | ~13 hrs |

Results in `penny_enum/n*_maximal_penny.g6` and `penny_enum/*.png`

## Dependencies

- **Go** - All tools written in Go
- **nauty** - `brew install nauty` - provides `shortg` for isomorphism
- **Python** (optional) - For plotting, needs scipy and matplotlib from `../hexagon_clink/venv`

## Quick Start

```bash
cd penny_enum

# Build tools
go build -o pipeline_nauty.out pipeline_nauty.go
go build -o verify_penny.out verify_penny.go
go build -o filter_maximal.out filter_maximal.go

# Run full pipeline for n=8
./pipeline_nauty.out -n 8 -out n8_unique.g6
./verify_penny.out -n 8 -in n8_unique.g6 -out n8_penny.g6
./filter_maximal.out -n 8 -out n8_maximal.g6 n8_penny.g6

# Plot results (requires scipy, matplotlib)
python3 plot_penny.py 8 n8_maximal.g6 n8_maximal.png
```

## Notes

- Penny graph verification uses gradient descent with multiple random restarts
- For large n, the pipeline writes batches to temp files and runs nauty incrementally
- The `filter_maximal` tool checks isomorphic subgraph relationships (O(n!) per pair)
- Reference implementation in `../hexagon_clink/explore_n7_k2/`

---

## find_fourth/ - SAT Search for 4th Arrangement

Given perfect-3 candidates (arr0, arr1, arr2 with no pairwise edge overlap), use SAT to find arr3 that covers all remaining pairs.

### Data
- `output_15/` - Perfect-3 pairs for n=15 (16 candidates)
- `output_17/` - Perfect-3 pairs for n=17 (~26M candidates)

Format: `arr1[0],arr1[1],...;arr2[0],arr2[1],...` (arr0 is always identity)

### Usage
```bash
cd find_fourth
go build -o find_fourth.out .
./find_fourth.out -n 15 -in output_15 -workers 1   # test with n=15
./find_fourth.out -n 17 -in output_17 -workers 1   # full n=17 search
```

**Note**: gophersat has threading bugs, must use `-workers 1`

### Results

**n=15**: Solution found (4 arrangements cover all 105 pairs)
```
arr0: [0,1,2,3,4,5,6,7,8,9,10,11,12,13,14]
arr1: [4,11,7,10,6,12,1,5,14,0,9,3,8,13,2]
arr2: [12,14,9,5,8,0,10,1,3,6,11,13,7,2,4]
arr3: [8,14,11,3,5,6,7,12,2,1,13,0,9,4,10]
```

**n=17**: In progress (~26M candidates at ~13/sec)

### Plotting
```bash
cd find_fourth/plot
../../../hexagon_clink/venv/bin/python plot_n15_solution.py
```

---

## polyiamond_enum/ - Polyiamond Enumeration

Enumerate polyiamonds and convert to graphs for finding unit-distance embeddings.

### Tools

| Tool | Description |
|------|-------------|
| `enumerate.py` | Python implementation (slower, good for small n) |
| `enumerate.go` | Go implementation (fast, multithreaded) |
| `plot_polyiamonds.py` | Plot graphs with triangular lattice embedding |

### Usage
```bash
cd polyiamond_enum

# Python version
./venv/bin/python enumerate.py 10 --show

# Go version (faster)
go build -o enumerate_fast.out enumerate.go
./enumerate_fast.out -min 13 -max 14 -v 13 -e 26 -coords output.txt -g6 output.g6

# Plot with isomorphism filtering
./venv/bin/python plot_polyiamonds.py output.txt output.png --unique
```

### Key Options (Go version)
- `-min`, `-max`: Triangle count range
- `-v`, `-e`: Filter by vertex/edge count
- `-g6`: Output graph6 format
- `-coords`: Output vertex coordinates for plotting
- `-w`: Number of workers (default: CPU count)

### Results

Polyiamonds with **13 vertices and 26 edges**: 4 unique non-isomorphic graphs
- Found from polyiamonds with 13-14 triangles
- All embeddable on triangular lattice with unit-distance edges
