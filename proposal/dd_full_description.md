# Project Description and Resources Justification

## Project Description

### Scientific Background

We address an open problem in discrete geometry and combinatorial optimization: given n elements on a hexagonal lattice (unit disk contact graph), what is the minimum number of distinct arrangements k(n) such that every pair of elements is adjacent in at least one arrangement? This problem connects to covering designs, permutation group theory, and has applications in experimental design where pairwise interactions must be tested under geometric constraints.

The sequence k(n) has been verified for n≤16 (where k=4 suffices), but n=17 remains open. The lower bound from edge-counting gives k≥4, but whether k=4 is achievable or k=5 is required is unknown.

### Technical Approach

Our approach exploits a structural property: we search for "perfect-3" arrangement triples where three consecutive arrangements share zero edges, maximizing coverage efficiency. We have enumerated all ~26 million such triples for n=17. For each candidate triple, we use a SAT solver to determine whether a fourth arrangement exists that covers all remaining uncovered pairs.

The SAT instances are moderately sized (~300 variables, ~50,000 clauses) but the sheer number of candidates (26M) makes sequential evaluation infeasible. Each SAT solve takes ~70ms, giving a sequential runtime of ~21 days.

### Expected Results and Timeline

**Month 1-2:**
- Port solver to Crux architecture, implement MPI-based parallelization
- Validate correctness against known n=15 solution
- Begin n=17 exhaustive search

**Month 3-4:**
- Complete n=17 search (definitive answer: k=4 achievable or k=5 required)
- If k=4 solution found: characterize solution space, count total solutions
- Begin preliminary runs for n=18, n=19

**Month 5-6:**
- Prepare INCITE proposal based on results
- Draft publication for Journal of Combinatorial Theory or Discrete Mathematics
- Profile and optimize solver for next-generation systems

### Proposal Plans

- **Target Program:** INCITE 2026
- **Submission Deadline:** June 2025 (anticipated)
- **PI Name:** [YOUR NAME]
- **Expected Title:** "Computational Determination of Minimum Covering Sequences for Hexagonal Lattice Contact Graphs"
- **Scope:** Full sequence determination for n≤30, requiring evaluation of exponentially larger candidate spaces

### Software Requirements

- **Commercial Software:** None required
- **Open Source Dependencies:**
  - Go compiler (standard)
  - gophersat SAT solver (BSD license)
  - Python/matplotlib for visualization (optional, post-processing only)

No special licensing or commercial packages needed.

### Storage Estimates

- **Input Data:** ~2GB (precomputed perfect-3 candidates for n=17), ~50GB projected for n=18-19
- **Candidate Enumeration:** We will generate perfect-3 candidates for n=18 and n=19 on Crux, producing ~500GB-1TB of candidate files
- **Intermediate Results:** SAT solver outputs, partial solution logs
- **Checkpointing:** Distributed checkpoint files across nodes for fault tolerance
- **Output Files:** Solution databases, statistical summaries, profiling data

**Estimate: 10-100K files, up to 2TB total storage**

---

## Resource Justification

### Crux (CPU Cluster)

**Importance to Project Goals:**
Crux is essential for this project. Our workload consists of millions of independent SAT instances—an embarrassingly parallel problem ideally suited to CPU clusters. The problem scales exponentially with n, and we aim to resolve n=17 definitively while making substantial progress on n=18 and n=19.

**Resource Request: 25,000 node-hours (2.5M core-hours at 100 cores/node)**

**Workload Breakdown:**
| Problem Size | Candidates | Est. Core-Hours | Purpose |
|--------------|------------|-----------------|---------|
| n=17 | 26M | 500K | Complete exhaustive search |
| n=17 enumeration variants | - | 200K | Alternative candidate generation strategies |
| n=18 candidate generation | - | 300K | Enumerate perfect-3 triples for n=18 |
| n=18 SAT search | ~200M (est.) | 1M | Partial search, characterize difficulty |
| n=19 preliminary | ~1B (est.) | 300K | Scaling studies, profiling |
| Development & profiling | - | 200K | Optimization, load balancing tuning |
| **Total** | | **2.5M** | |

**Codes Being Used:**
- Custom Go solver (`find_fourth`) implementing:
  - Candidate parsing and coverage computation
  - SAT encoding for arrangement existence
  - gophersat solver integration
- Candidate enumerator (`count_perfect3`) for generating perfect-3 triples
- MPI wrapper for work distribution (to be developed)

**Development Work Planned:**
1. Replace current threading model with MPI-based distribution (gophersat has thread-safety issues requiring process-level parallelism)
2. Implement dynamic load balancing—SAT solve times vary by instance (some UNSAT proofs are fast, SAT solutions require more exploration)
3. Add hierarchical checkpointing for fault tolerance on long runs
4. Profile and optimize memory usage to maximize candidates per node
5. Develop streaming candidate generation to reduce I/O bottlenecks for n≥18

**Experience on Similar Architectures:**
I have extensive experience running parallel workloads on the Max Planck Computing and Data Facility (MPCDF) in Garching, Germany, including MPI-parallelized combinatorial search codes on their HPC clusters. The Crux architecture (x86 CPU cluster) is similar to systems I have used at MPCDF, and I am familiar with job scheduling, MPI debugging, and performance optimization on such platforms.

**Scaling Strategy:**
- Initial runs: 10-100 nodes to validate parallelization
- Production n=17: 250-500 nodes for rapid completion
- n=18 exploration: 500-1000 nodes to make meaningful progress within allocation

### Eagle (Data Storage)

**Resource Request: 2TB**

**Importance to Project Goals:**
Eagle provides essential storage for the full project scope:

1. **Input Candidates:**
   - n=17: ~2GB (existing)
   - n=18: ~50-100GB (to be generated)
   - n=19: ~500GB-1TB (to be generated)

2. **Candidate Generation Output:**
   The perfect-3 enumeration for larger n produces substantial intermediate data. Storing candidates on Eagle allows separation of enumeration and SAT-solving phases, enabling restarts and iterative refinement.

3. **Checkpointing:**
   Distributed checkpoint files across long-running jobs for fault tolerance. With 500+ node jobs, robust checkpointing is critical.

4. **Results Database:**
   All solutions found, solve-time statistics, and coverage analysis for publication and INCITE proposal preparation.

**Storage Breakdown:**
| Data Type | Size | Purpose |
|-----------|------|---------|
| n=17 candidates | 2GB | Existing input |
| n=18 candidates | 100GB | Generated on Crux |
| n=19 candidates | 1TB | Generated on Crux |
| Checkpoints | 100GB | Fault tolerance |
| Results & logs | 50GB | Analysis, publication |
| **Total** | **~1.3TB** | (requesting 2TB for headroom) |

**File Access Pattern:**
- Write-heavy during candidate enumeration phase
- Read-heavy during SAT solving phase
- Periodic checkpoint writes during long jobs
- Modest metadata load (10-100K files)

---

## Long-Term Goals

This DD allocation is preparation for a larger INCITE proposal. The n=17 result will:

1. **Validate our approach:** Confirm SAT-based search is effective for this problem class
2. **Provide scaling data:** Measure solve-time distributions to estimate INCITE resource requirements for n≤30
3. **Establish feasibility:** A definitive n=17 result (either way) is publishable and demonstrates scientific impact

**Post-DD Requirements:**
Yes, we expect to require additional allocations. If n=17 is resolved, the natural next step is n=18-20 (where k likely increases to 5), requiring 10-100x more candidates. An INCITE allocation would target complete sequence determination up to n=30, potentially requiring millions of core-hours.

---

## Summary

| Resource | Request | Justification |
|----------|---------|---------------|
| Crux | 25,000 node-hours (2.5M core-hours) | Complete n=17, substantial progress on n=18-19, INCITE preparation |
| Eagle | 2TB | Candidate storage for n=17-19, checkpoints, results database |

This DD allocation will resolve an open combinatorics problem (n=17), make significant progress toward n=18-19, produce publication-ready results, and prepare a competitive INCITE proposal for complete sequence determination up to n=30.
