"""Plot the n=17 solution found by find_fourth."""

from visualize_solution import visualize_solution

arrangements = [
    [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16],
    [0, 8, 15, 9, 16, 12, 10, 5, 13, 2, 6, 3, 7, 14, 11, 1, 4],
    [10, 13, 16, 6, 14, 1, 15, 11, 0, 7, 2, 4, 9, 12, 8, 3, 5],
    [11, 4, 7, 5, 9, 6, 8, 16, 3, 10, 15, 12, 2, 14, 0, 13, 1],
]

visualize_solution(arrangements, n=17, output_file="solution_17.png")
