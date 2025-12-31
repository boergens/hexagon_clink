#!/usr/bin/env python3
"""
Plot the n=15 solution found by find_fourth SAT solver.
4 arrangements cover all 105 pairs.
"""

from visualize_solution import visualize_solution

arrangements = [
    (0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14),
    (4, 11, 7, 10, 6, 12, 1, 5, 14, 0, 9, 3, 8, 13, 2),
    (12, 14, 9, 5, 8, 0, 10, 1, 3, 6, 11, 13, 7, 2, 4),
    (8, 14, 11, 3, 5, 6, 7, 12, 2, 1, 13, 0, 9, 4, 10),
]

visualize_solution(arrangements, n=15, output_file="n15_solution.png")
