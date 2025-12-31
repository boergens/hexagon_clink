#!/usr/bin/env python3
"""
Enumerate all distinct polyiamonds (shapes made of N edge-connected equilateral triangles).
Accounts for rotation and reflection symmetry (D6 group = 12 operations).
"""

import sys
from typing import Set, FrozenSet, List, Tuple

# Represent triangles by their 3 vertices in a triangular lattice
# Vertices use (a, b) integer coordinates where:
#   real position = (a + b/2, b * sqrt(3)/2)
Vertex = Tuple[int, int]
Triangle = FrozenSet[Vertex]  # Always exactly 3 vertices
Polyiamond = FrozenSet[Triangle]


def make_triangle(v1: Vertex, v2: Vertex, v3: Vertex) -> Triangle:
    """Create a triangle from 3 vertices."""
    return frozenset([v1, v2, v3])


def get_grid_triangles_at(q: int, r: int) -> Tuple[Triangle, Triangle]:
    """Get the up and down triangles at grid position (q, r)."""
    # Up-pointing triangle: vertices at (q,r), (q+1,r), (q,r+1)
    up = make_triangle((q, r), (q + 1, r), (q, r + 1))
    # Down-pointing triangle: vertices at (q,r+1), (q+1,r), (q+1,r+1)
    down = make_triangle((q, r + 1), (q + 1, r), (q + 1, r + 1))
    return up, down


def is_up_pointing(tri: Triangle) -> bool:
    """Check if triangle is up-pointing based on vertex arrangement."""
    verts = sorted(tri)
    # Up: two vertices have same b, third has b+1
    # Down: two vertices have same b, third has b-1
    b_values = sorted(v[1] for v in verts)
    return b_values[0] == b_values[1]  # Two lower, one higher = up


def rotate_vertex_60(v: Vertex) -> Vertex:
    """Rotate vertex 60° counterclockwise around origin."""
    a, b = v
    return (-b, a + b)


def reflect_vertex(v: Vertex) -> Vertex:
    """Reflect vertex across the a-axis."""
    a, b = v
    return (a + b, -b)


def transform_triangle(tri: Triangle, rotate: int, do_reflect: bool) -> Triangle:
    """Apply rotation and optional reflection to triangle."""
    verts = list(tri)

    # Apply reflection first
    if do_reflect:
        verts = [reflect_vertex(v) for v in verts]

    # Apply rotation
    for _ in range(rotate % 6):
        verts = [rotate_vertex_60(v) for v in verts]

    return frozenset(verts)


def transform_shape(shape: Polyiamond, rotate: int, do_reflect: bool) -> Polyiamond:
    """Apply rotation and optional reflection to entire shape."""
    return frozenset(transform_triangle(tri, rotate, do_reflect) for tri in shape)


def normalize_position(shape: Polyiamond) -> Polyiamond:
    """Translate shape so minimum vertex coordinates are at (0, 0)."""
    if not shape:
        return shape

    all_verts = [v for tri in shape for v in tri]
    min_a = min(v[0] for v in all_verts)
    min_b = min(v[1] for v in all_verts)

    new_triangles = set()
    for tri in shape:
        new_verts = frozenset((a - min_a, b - min_b) for a, b in tri)
        new_triangles.add(new_verts)

    return frozenset(new_triangles)


def canonicalize(shape: Polyiamond) -> Polyiamond:
    """Return canonical form (lexicographically smallest under all symmetries)."""
    candidates = []

    # D6 group: 6 rotations × 2 (with/without reflection) = 12 operations
    for rot in range(6):
        for do_reflect in [False, True]:
            transformed = transform_shape(shape, rot, do_reflect)
            normalized = normalize_position(transformed)
            candidates.append(normalized)

    # Convert to sortable form for comparison
    def shape_key(s):
        return sorted(tuple(sorted(tri)) for tri in s)

    return min(candidates, key=shape_key)


def get_adjacent_triangles(tri: Triangle) -> List[Triangle]:
    """Get all triangles that share an edge with the given triangle."""
    verts = list(tri)
    neighbors = []

    # For each pair of vertices (edge), find the triangle on the other side
    for i in range(3):
        v1, v2 = verts[i], verts[(i + 1) % 3]
        v3 = verts[(i + 2) % 3]  # The vertex not on this edge

        # The fourth vertex that completes a parallelogram
        # v4 = v1 + v2 - v3
        v4 = (v1[0] + v2[0] - v3[0], v1[1] + v2[1] - v3[1])

        neighbor = make_triangle(v1, v2, v4)
        neighbors.append(neighbor)

    return neighbors


def get_boundary(shape: Polyiamond) -> Set[Triangle]:
    """Get all triangles adjacent to the shape but not in it."""
    boundary = set()
    for tri in shape:
        for neighbor in get_adjacent_triangles(tri):
            if neighbor not in shape:
                boundary.add(neighbor)
    return boundary


def enumerate_polyiamonds(n: int) -> Set[Polyiamond]:
    """Enumerate all distinct polyiamonds with n triangles."""
    if n < 1:
        return set()

    # Start with single up-pointing triangle
    up, _ = get_grid_triangles_at(0, 0)
    initial = frozenset([up])

    if n == 1:
        return {canonicalize(initial)}

    current_shapes: Set[Polyiamond] = {canonicalize(initial)}

    for size in range(2, n + 1):
        next_shapes: Set[Polyiamond] = set()

        for shape in current_shapes:
            for new_tri in get_boundary(shape):
                new_shape = shape | {new_tri}
                canonical = canonicalize(new_shape)
                next_shapes.add(canonical)

        current_shapes = next_shapes

    return current_shapes


def triangle_to_qr(tri: Triangle) -> Tuple[int, int, int]:
    """Convert vertex-based triangle to (q, r, orientation) format."""
    verts = sorted(tri)
    up = is_up_pointing(tri)

    if up:
        # Up triangle: bottom-left vertex is at (q, r)
        q, r = min(verts, key=lambda v: (v[1], v[0]))
    else:
        # Down triangle: top-left vertex is at (q, r+1), so q, r is one less in r
        top_left = min((v for v in verts if v[1] == max(v2[1] for v2 in verts)), key=lambda v: v[0])
        q, r = top_left[0], top_left[1] - 1

    return (q, r, 0 if up else 1)


def shape_to_ascii(shape: Polyiamond) -> str:
    """Convert polyiamond to ASCII art representation."""
    if not shape:
        return ""

    # Convert triangles to (q, r, orient) format
    qr_tris = [triangle_to_qr(tri) for tri in shape]

    min_q = min(t[0] for t in qr_tris)
    max_q = max(t[0] for t in qr_tris)
    min_r = min(t[1] for t in qr_tris)
    max_r = max(t[1] for t in qr_tris)

    height = max_r - min_r + 1
    width = (max_q - min_q + 1) * 2 + (max_r - min_r) + 2

    grid = [[' ' for _ in range(width)] for _ in range(height)]

    for q, r, orient in qr_tris:
        row = max_r - r
        col = (q - min_q) * 2 + (r - min_r)

        if orient == 0:
            char = '△'
        else:
            col += 1
            char = '▽'

        if 0 <= row < height and 0 <= col < width:
            grid[row][col] = char

    return '\n'.join(''.join(row).rstrip() for row in grid)


def polyiamond_to_graph(shape: Polyiamond) -> Tuple[Set[Vertex], Set[FrozenSet[Vertex]]]:
    """Convert polyiamond to graph (vertices and edges)."""
    vertices = set()
    edges = set()

    for tri in shape:
        verts = list(tri)
        for v in verts:
            vertices.add(v)
        # Add the 3 edges of this triangle
        for i in range(3):
            edge = frozenset([verts[i], verts[(i+1) % 3]])
            edges.add(edge)

    return vertices, edges


def main():
    if len(sys.argv) < 2:
        print(f"Usage: {sys.argv[0]} <n> [--show] [--edges E]")
        print("  n: number of triangles")
        print("  --show: display ASCII art of each polyiamond")
        print("  --edges E: filter to show only polyiamonds with E graph edges")
        sys.exit(1)

    n = int(sys.argv[1])
    show = '--show' in sys.argv

    edge_filter = None
    if '--edges' in sys.argv:
        idx = sys.argv.index('--edges')
        edge_filter = int(sys.argv[idx + 1])

    print(f"Enumerating polyiamonds with {n} triangles...")
    shapes = enumerate_polyiamonds(n)

    print(f"Found {len(shapes)} distinct polyiamond(s)")

    # Known values (OEIS A000577)
    known = {1: 1, 2: 1, 3: 1, 4: 3, 5: 4, 6: 12, 7: 24, 8: 66, 9: 160, 10: 448}
    if n in known:
        expected = known[n]
        status = "✓" if len(shapes) == expected else f"✗ (expected {expected})"
        print(f"Verification: {status}")

    # Count edges for each polyiamond
    edge_counts = {}
    for shape in shapes:
        _, edges = polyiamond_to_graph(shape)
        num_edges = len(edges)
        edge_counts[num_edges] = edge_counts.get(num_edges, 0) + 1

    print(f"\nEdge count distribution:")
    for e in sorted(edge_counts.keys()):
        print(f"  {e} edges: {edge_counts[e]} polyiamond(s)")

    if edge_filter is not None:
        filtered = [s for s in shapes if len(polyiamond_to_graph(s)[1]) == edge_filter]
        print(f"\nPolyiamonds with exactly {edge_filter} edges: {len(filtered)}")

        if show:
            for i, shape in enumerate(filtered, 1):
                print(f"\n--- Polyiamond {i} ---")
                print(shape_to_ascii(shape))
    elif show:
        shapes_list = sorted(shapes, key=lambda s: sorted(tuple(sorted(tri)) for tri in s))
        for i, shape in enumerate(shapes_list, 1):
            print(f"\n--- Polyiamond {i} ---")
            print(shape_to_ascii(shape))


if __name__ == "__main__":
    main()
