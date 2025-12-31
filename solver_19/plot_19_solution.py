"""Plot the n=19 solution."""

import sys
sys.path.insert(0, '../plotting')

from visualize_solution import visualize_solution

arrangements = [
    [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18],
    [8, 16, 2, 5, 13, 15, 10, 14, 0, 7, 4, 12, 9, 18, 1, 11, 17, 3, 6],
    [4, 18, 10, 17, 1, 16, 11, 0, 15, 12, 7, 13, 2, 14, 5, 3, 9, 6, 8],
    [1, 10, 12, 18, 2, 15, 7, 5, 0, 17, 8, 14, 3, 16, 13, 6, 4, 9, 11],
    [15, 0, 16, 18, 8, 17, 9, 13, 5, 12, 6, 1, 10, 2, 4, 3, 7, 14, 11],
]

visualize_solution(arrangements, n=19, output_file="solution_19.png")
