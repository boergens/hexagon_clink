# Penny Graph & Hexagon Clink Project

## Coding Conventions

- **Go binaries**: Always use `.out` extension when compiling: `go build -o foo.out foo.go`
- **Python**: Use the project venv: `source venv/bin/activate` (has matplotlib, numpy)

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

    solver_general/
    (general solver for any n,k on hex spiral)
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
| 13-17 | 4 | **proven** (n=13 via solver_k, n=17 via find_fourth) |
| 19 | 5 | **proven** (via solver_19 with hexagonal symmetry) |

**n=13 proof**: 78 pairs, 26 edges per graph, so 3 arrangements need exactly 78 edges with zero overlap. `solver_k/` exhaustively checks all 4 maximal penny graphs in all configurations - no valid 3-arrangement exists.

**n=19 proof**: 171 pairs, 42 edges per graph. Since 19 is a hexagonal number (1+6+12), symmetry reduces the search space by restricting item 0 to 4 representative positions.

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

**n=17**: Solution found at candidate 317544 (4 arrangements cover all 136 pairs)
```
arr0: [0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]
arr1: [0,8,15,9,16,12,10,5,13,2,6,3,7,14,11,1,4]
arr2: [10,13,16,6,14,1,15,11,0,7,2,4,9,12,8,3,5]
arr3: [11,4,7,5,9,6,8,16,3,10,15,12,2,14,0,13,1]
```

---

## solver_general/ - General Solver

General-purpose solver for any n and k on the hexagon spiral graph. Uses backtracking with pruning.

### Algorithm
1. Fix arr0 = identity
2. For each subsequent arrangement, backtrack through all permutations
3. Prune branches that exceed max overlap (derived from min-edges constraint)
4. For final arrangement, use doomed-pair check: if placing an item leaves an uncoverable pair with an already-placed item, skip it

### Usage
```bash
cd solver_general
go build -o solver.out solver.go
./solver.out -n 12 -k 3 -workers 1
```

### Flags
- `-n`: Number of items (default 17)
- `-k`: Number of arrangements to find (default 4)
- `-workers`: Parallel workers with different random seeds (default 8)
- `-max-overlap`: Comma-separated max overlap per level (e.g., '5,5,5')

### Results
- **n=7 k=2**: No solution (proves k≥3 needed)
- **n=7 k=3**: Solution found instantly
- **n=12 k=3**: Solution found in ~20s
- **n=13 k=3**: No solution in ~200ms (proves k≥4 needed)

---

## solver_19/ - Specialized n=19 Solver

Specialized solver for n=19, k=5 exploiting hexagonal symmetry.

### Symmetry Breaking
Since 19 is a hexagonal number (1+6+12), the hex spiral has 6-fold symmetry. Item 0 is restricted to 4 representative positions:
- Position 0: center
- Position 1: middle ring
- Position 7: outer ring corner
- Position 8: outer ring edge-center

### Usage
```bash
cd solver_19
go build -o solver.out solver.go
./solver.out -workers 8 -max-overlap 0,0,12
```

### Results
**n=19**: Solution found (5 arrangements cover all 171 pairs)
```
arr0: [0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18]
arr1: [8,16,2,5,13,15,10,14,0,7,4,12,9,18,1,11,17,3,6]
arr2: [4,18,10,17,1,16,11,0,15,12,7,13,2,14,5,3,9,6,8]
arr3: [1,10,12,18,2,15,7,5,0,17,8,14,3,16,13,6,4,9,11]
arr4: [15,0,16,18,8,17,9,13,5,12,6,1,10,2,4,3,7,14,11]
```

---

## solver_20/ - Specialized n=20 Solver

Specialized solver for n=20, k=5 with optimization for the low-degree slot.

### Special Slot Optimization
Slot 19 has only degree 2 (neighbors: 7, 18). At the last arrangement level (arr4), slot 19 is filled first, and only items needing ≤2 more partners are tried there. This prunes the search space significantly.

### Usage
```bash
cd solver_20
go build -o solver.out solver.go
./solver.out -workers 8 -max-overlap 0,0,10
```

### Flags
- `-workers`: Parallel workers with different random seeds (default 8)
- `-max-overlap`: Comma-separated max overlap for arr1, arr2, arr3 (arr4 must cover remaining pairs exactly)

---

## plotting/ - Solution Visualization

Visualize arrangements on the penny spiral graph.

### Usage
```bash
./venv/bin/python plotting/plot_n17_solution.py
```

### Files
- `hex_spiral.py` - Builds penny spiral coordinates and adjacencies
- `visualize_solution.py` - Matplotlib visualization of arrangements
- `plot_n17_solution.py` - Plot the n=17 solution
- `solution_17.png` - Generated visualization

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
