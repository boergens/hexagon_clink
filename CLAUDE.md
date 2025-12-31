# Penny Graph & Hexagon Clink Project

## Coding Conventions

- **Go binaries**: Always use `.out` extension when compiling: `go build -o foo.out foo.go`

## Overview

The **Hexagon Clink Problem**: Given n items, find the minimum number of arrangements on a penny graph such that every pair of items is adjacent at least once.

This project attacks the problem from multiple angles:

```
                    Hexagon Clink Problem
                            |
        +-------------------+-------------------+
        |                   |                   |
   penny_enum/         polyiamond_enum/     solver_k/
   (enumerate all      (find ALL maximal    (prove lower
    penny graphs)       penny graphs via     bounds via
                        triangular lattice)  exhaustive
        |                   |                backtracking)
        v                   v                    |
    find_fourth/  <----  4 graphs for n=13  --->+
    (SAT solver to find k-th arrangement
     given perfect (k-1) arrangements)
```

**Key insight**: The penny spiral (used in original problem) is just ONE maximal penny graph. To prove lower bounds, we must check ALL maximal penny graphs for that vertex count.

## The Core Problem

### Penny Graphs
Graphs realizable by placing non-overlapping unit circles in 2D, where vertices are circle centers and edges connect touching circles.

Constraints:
- **Edge distance = 1** (touching circles)
- **Non-edge distance > 1** (non-overlapping)
- **No K4 subgraph** (4 mutually touching circles impossible in 2D)
- **Max degree ≤ 6** (hexagonal packing limit)

### Counting Argument

For n items:
- Pairs to cover: C(n,2) = n(n-1)/2
- Edges in maximal penny graph: E(n)
- Lower bound on arrangements: ceil(pairs / edges)

If `k * edges = pairs` exactly, then k arrangements work **only if** they have zero pairwise overlap (every edge covers a unique pair).

### Known Results

| n | min arrangements | status |
|---|------------------|--------|
| 3 | 1 | trivial |
| 4-6 | 2 | proven |
| 7-12 | 3 | proven |
| 13-16 | 4 | **proven** (n=13 via solver_k) |
| 17 | 4 or 5? | in progress |

**n=13 proof**: 78 pairs, 26 edges per graph, so 3 arrangements need exactly 78 edges with zero overlap. `solver_k/` exhaustively checks all 4 maximal penny graphs in all configurations - no valid 3-arrangement exists.

---

## penny_enum/ - Penny Graph Enumeration

Enumerate all penny graphs on n vertices via candidate generation + verification.

### Pipeline
1. **Generate candidates** - All graphs with filters (connected, max degree ≤6, no K4)
2. **Remove isomorphisms** - Use nauty's `shortg`
3. **Verify penny embedding** - Gradient descent to find valid 2D embedding
4. **Filter maximal** - Keep only graphs not subgraphs of larger ones

### Usage
```bash
cd penny_enum
go build -o pipeline_nauty.out pipeline_nauty.go
go build -o verify_penny.out verify_penny.go
go build -o filter_maximal.out filter_maximal.go

./pipeline_nauty.out -n 8 -out n8_unique.g6
./verify_penny.out -n 8 -in n8_unique.g6 -out n8_penny.g6
./filter_maximal.out -n 8 -out n8_maximal.g6 n8_penny.g6
```

### Results

| n | Candidates | Penny | Maximal | Max Edges |
|---|------------|-------|---------|-----------|
| 8 | 5,481 | 677 | 9 | 14 |
| 9 | 88,958 | 3,136 | 16 | 16 |

---

## polyiamond_enum/ - Polyiamond Enumeration

Enumerate **polyiamonds** (shapes made of edge-connected equilateral triangles) and extract their contact graphs.

**Key insight**: Polyiamond contact graphs are automatically valid penny graphs embeddable on the triangular lattice. This finds ALL maximal penny graphs for small n.

### Usage
```bash
cd polyiamond_enum
go build -o enumerate_fast.out enumerate.go
./enumerate_fast.out -min 13 -max 14 -v 13 -e 26 -coords output.txt -g6 output.g6
```

### Results

**n=13**: Found exactly **4 non-isomorphic maximal penny graphs** with 26 edges each.
These are used by `solver_k/` and `find_fourth/`.

---

## solver_k/ - Exhaustive Backtracking Solver

Prove that k arrangements are insufficient for n items by exhaustive search over all maximal penny graphs.

### Algorithm (for n=13, k=3)
1. Enumerate all graph triples (shape0, shape1, shape2) with symmetry breaking: shape0 ≤ shape1 ≤ shape2 (20 combinations instead of 64)
2. Fix arr0 = identity on shape0
3. Backtrack to find arr1 on shape1 with zero overlap with arr0
4. Backtrack to find arr2 on shape2 covering exactly the remaining pairs
5. Prune aggressively: any "wasted" edge (covering an already-covered pair) terminates that branch

### Usage
```bash
cd solver_k
go build -o solver_13_3.out solver_13_3.go
./solver_13_3.out  # uses 13 parallel workers
```

### Results
**n=13**: No valid 3-arrangement exists. Proves n=13 requires at least 4 arrangements.
- Checked all 10 shape-pair combinations (with symmetry)
- ~27,000 valid arr1 permutations total
- Runtime: ~1m40s on modern hardware

---

## find_fourth/ - SAT Search for k-th Arrangement

Given perfect-(k-1) candidates (arrangements with zero pairwise overlap), use SAT to find the k-th arrangement covering remaining pairs.

### Data
- `output_15/` - Perfect-3 pairs for n=15 (16 candidates)
- `output_17/` - Perfect-3 pairs for n=17 (~26M candidates)

### Usage
```bash
cd find_fourth
go build -o find_fourth.out .
./find_fourth.out -n 15 -in output_15 -workers 1
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

**n=17**: In progress (~26M candidates)

---

## mathematica/ - Prototype Scripts

Mathematica implementations of the backtracking solver (used for prototyping before Go version).

- `prove_n13_needs_4.m` - Single spiral graph version
- `prove_n13_needs_4_all_graphs.m` - All 4 maximal graphs version
- `decode_g6.go` - Convert graph6 to Mathematica edge list format

---

## Dependencies

- **Go** - All tools written in Go
- **nauty** - `brew install nauty` - provides `shortg` for isomorphism
- **Python** (optional) - For plotting, needs scipy and matplotlib

## File Formats

- **Graph6 (.g6)** - Text format used by nauty, one graph per line
- **Binary (.bin)** - Compact edge bitmask format for large enumerations
