"""
Core penny spiral building logic for hexagonal grids.
"""

HEX_DIRS = [
    (1.5, 0),
    (0.75, 1.3),
    (-0.75, 1.3),
    (-1.5, 0),
    (-0.75, -1.3),
    (0.75, -1.3),
]


def add_vec(a, b):
    return (a[0] + b[0], a[1] + b[1])


def vec_close(a, b, tol=0.1):
    return abs(a[0] - b[0]) < tol and abs(a[1] - b[1]) < tol


def find_neighbors(pos, positions):
    neighbors = []
    for node, node_pos in positions.items():
        for d in HEX_DIRS:
            if vec_close(add_vec(node_pos, d), pos):
                neighbors.append(node)
                break
    return neighbors


def get_neighbor_positions(pos):
    return [add_vec(pos, d) for d in HEX_DIRS]


def position_occupied(pos, positions, tol=0.1):
    for existing in positions.values():
        if vec_close(pos, existing, tol):
            return True
    return False


def build_spiral(n):
    if n < 1:
        return {}, set()

    positions = {0: (0, 0)}
    adjacencies = set()

    if n == 1:
        return positions, adjacencies

    for node in range(1, n):
        best_pos = None
        best_contacts = 0
        best_dist = float('inf')

        prev_pos = positions[node - 1]
        candidates = get_neighbor_positions(prev_pos)

        for cand in candidates:
            if position_occupied(cand, positions):
                continue

            neighbors = find_neighbors(cand, positions)
            num_contacts = len(neighbors)
            dist = cand[0]**2 + cand[1]**2

            if (num_contacts > best_contacts or
                (num_contacts == best_contacts and dist < best_dist)):
                best_pos = cand
                best_contacts = num_contacts
                best_dist = dist

        if best_pos is None:
            raise ValueError(f"Could not place node {node}")

        positions[node] = best_pos
        neighbors = find_neighbors(best_pos, positions)
        for neighbor in neighbors:
            adjacencies.add((min(node, neighbor), max(node, neighbor)))

    return positions, adjacencies
