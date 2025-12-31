#!/usr/bin/env python3
"""
Plot polyiamond graphs with unit-distance triangular lattice embedding.
Reads vertex coordinates from a simple text format output by the Go enumerator.
"""

import sys
import math
import subprocess
import tempfile
import matplotlib.pyplot as plt


def to_cartesian(a, b):
    """Convert triangular lattice (a,b) to cartesian (x,y)."""
    x = a + b * 0.5
    y = b * math.sqrt(3) / 2
    return x, y


def graph_to_g6(n, edges):
    """Convert graph to graph6 format."""
    # Build adjacency matrix
    adj = [[False] * n for _ in range(n)]
    for e1, e2 in edges:
        if e1 > e2:
            e1, e2 = e2, e1
        adj[e1][e2] = True

    # Encode n
    result = [chr(n + 63)] if n <= 62 else [chr(126), chr((n >> 12) + 63), chr(((n >> 6) & 63) + 63), chr((n & 63) + 63)]

    # Encode upper triangle
    bits = []
    for j in range(1, n):
        for i in range(j):
            bits.append(adj[i][j])

    # Pad to multiple of 6
    while len(bits) % 6 != 0:
        bits.append(False)

    # Convert to characters
    for i in range(0, len(bits), 6):
        val = sum((1 << (5 - j)) for j in range(6) if bits[i + j])
        result.append(chr(val + 63))

    return ''.join(result)


def canonicalize_g6(g6_strings):
    """Use nauty's shortg to get canonical g6 forms."""
    with tempfile.NamedTemporaryFile(mode='w', suffix='.g6', delete=False) as f:
        for s in g6_strings:
            f.write(s + '\n')
        tmp_in = f.name

    result = subprocess.run(['shortg', '-q', tmp_in], capture_output=True, text=True)

    # Read the canonicalized file
    with open(tmp_in, 'r') as f:
        canonical = [line.strip() for line in f if line.strip()]

    import os
    os.unlink(tmp_in)

    return canonical


def parse_graph_file(filename):
    """Parse graph file with vertices and edges."""
    graphs = []

    with open(filename, 'r') as f:
        lines = [line.strip() for line in f if line.strip()]

    i = 0
    while i < len(lines):
        if lines[i].startswith('GRAPH'):
            graph = {'id': int(lines[i].split()[1]), 'vertices': [], 'edges': []}
            i += 1

            nv = int(lines[i].split()[1])
            i += 1
            for _ in range(nv):
                a, b = map(int, lines[i].split())
                graph['vertices'].append((a, b))
                i += 1

            ne = int(lines[i].split()[1])
            i += 1
            for _ in range(ne):
                e1, e2 = map(int, lines[i].split())
                graph['edges'].append((e1, e2))
                i += 1

            graphs.append(graph)
        else:
            i += 1

    return graphs


def filter_isomorphic(graphs):
    """Keep only one representative per isomorphism class."""
    # Convert each graph to g6
    g6_list = []
    for g in graphs:
        n = len(g['vertices'])
        g6 = graph_to_g6(n, g['edges'])
        g6_list.append(g6)

    # Get canonical forms
    canonical = canonicalize_g6(g6_list)

    # Keep first occurrence of each canonical form
    seen = set()
    unique_graphs = []
    for g, canon in zip(graphs, canonical):
        if canon not in seen:
            seen.add(canon)
            unique_graphs.append(g)

    return unique_graphs


def plot_graphs(graphs, output_file):
    """Plot graphs in a grid layout."""
    n = len(graphs)
    cols = min(n, 2)
    rows = (n + cols - 1) // cols

    fig, axes = plt.subplots(rows, cols, figsize=(6*cols, 6*rows))
    if n == 1:
        axes = [axes]
    else:
        axes = axes.flatten()

    for idx, graph in enumerate(graphs):
        ax = axes[idx]

        # Convert vertices to cartesian
        pos = [to_cartesian(a, b) for a, b in graph['vertices']]

        # Draw edges
        for e1, e2 in graph['edges']:
            x1, y1 = pos[e1]
            x2, y2 = pos[e2]
            ax.plot([x1, x2], [y1, y2], 'gray', linewidth=1.5, zorder=1)

        # Draw vertices
        xs = [p[0] for p in pos]
        ys = [p[1] for p in pos]
        ax.scatter(xs, ys, s=200, c='lightblue', edgecolors='black', zorder=2)

        # Label vertices
        for i, (x, y) in enumerate(pos):
            ax.annotate(str(i), (x, y), ha='center', va='center', fontsize=8)

        ax.set_aspect('equal')
        nv = len(graph['vertices'])
        ne = len(graph['edges'])
        ax.set_title(f"Graph {graph['id']}: {nv}v, {ne}e")
        ax.axis('off')

    # Hide unused axes
    for idx in range(n, len(axes)):
        axes[idx].axis('off')

    plt.tight_layout()
    plt.savefig(output_file, dpi=150)
    print(f"Saved to {output_file}")


def main():
    if len(sys.argv) < 3:
        print(f"Usage: {sys.argv[0]} <input.txt> <output.png> [--unique]")
        print("  --unique: filter to keep only non-isomorphic graphs")
        sys.exit(1)

    input_file = sys.argv[1]
    output_file = sys.argv[2]
    unique = '--unique' in sys.argv

    graphs = parse_graph_file(input_file)
    print(f"Loaded {len(graphs)} graphs")

    if unique:
        graphs = filter_isomorphic(graphs)
        print(f"After isomorphism filtering: {len(graphs)} unique graphs")

    plot_graphs(graphs, output_file)


if __name__ == "__main__":
    main()
