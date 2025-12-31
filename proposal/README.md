# Compute Time Proposal

## Project Title

**Minimum Covering Arrangements for Unit Disk Contact Graphs on Hexagonal Lattices**

Alternative titles:
- Optimal Pair Coverage in Hexagonal Lattice Permutation Systems
- Extremal Bounds for Complete Adjacency Coverage in Planar Contact Graphs
- Computational Bounds on Minimum Arrangement Sequences for Hexagonal Configurations

## Problem Statement

Given n elements arranged on a hexagonal lattice (unit disk contact graph), we seek the minimum number of distinct arrangements k such that every pair of elements is adjacent in at least one arrangement.

## Key Questions

1. For n=17: Is k=4 achievable, or is k=5 required?
2. What is the growth rate of k(n) as n increases?
3. Do "perfect" arrangements (zero edge overlap between consecutive arrangements) always exist?

## Computational Challenge

- n=17 has ~26 million candidate arrangement triples to evaluate
- Each candidate requires SAT solving to determine if a 4th arrangement exists
- Current rate: ~13 candidates/sec (single-threaded due to solver limitations)
- Estimated sequential time: ~23 days

## Requested Resources

[TODO: Specify CPU hours, parallelization strategy, etc.]
