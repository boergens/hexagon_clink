"""
Core penny spiral building logic for hexagonal grids.
Reusable module - no visualization dependencies.
"""

# Hex direction vectors (flat-top hexagons)
# Directions: 0=right, 1=up-right, 2=up-left, 3=left, 4=down-left, 5=down-right
HEX_DIRS = [
    (1.5, 0),           # 0: right
    (0.75, 1.3),        # 1: up-right
    (-0.75, 1.3),       # 2: up-left
    (-1.5, 0),          # 3: left
    (-0.75, -1.3),      # 4: down-left
    (0.75, -1.3),       # 5: down-right
]


def add_vec(a, b):
    return (a[0] + b[0], a[1] + b[1])


def vec_close(a, b, tol=0.1):
    return abs(a[0] - b[0]) < tol and abs(a[1] - b[1]) < tol


def find_neighbors(pos, positions):
    """Find which existing nodes are adjacent to position pos."""
    neighbors = []
    for node, node_pos in positions.items():
        for d in HEX_DIRS:
            if vec_close(add_vec(node_pos, d), pos):
                neighbors.append(node)
                break
    return neighbors


def get_neighbor_positions(pos):
    """Get all 6 neighboring positions of pos."""
    return [add_vec(pos, d) for d in HEX_DIRS]


def position_occupied(pos, positions, tol=0.1):
    """Check if a position is already occupied."""
    for existing in positions.values():
        if vec_close(pos, existing, tol):
            return True
    return False


def build_spiral(n):
    """
    Build penny spiral for n nodes.
    Returns (positions, adjacencies) where:
      - positions: dict mapping node index -> (x, y) coordinates
      - adjacencies: set of (a, b) tuples where a < b

    Key constraint: each new node must be adjacent to the previous node (spiral property).
    Among valid positions, prefer more total contacts, then closest to origin.
    """
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

        # Must be adjacent to previous node (node-1)
        prev_pos = positions[node - 1]
        candidates = get_neighbor_positions(prev_pos)

        for cand in candidates:
            if position_occupied(cand, positions):
                continue

            neighbors = find_neighbors(cand, positions)
            num_contacts = len(neighbors)
            dist = cand[0]**2 + cand[1]**2

            # Prefer more contacts, then closer to origin
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


def get_adjacencies(n):
    """Convenience function: return just the adjacencies for n nodes."""
    _, adjacencies = build_spiral(n)
    return adjacencies


def get_positions(n):
    """Convenience function: return just the positions for n nodes."""
    positions, _ = build_spiral(n)
    return positions
