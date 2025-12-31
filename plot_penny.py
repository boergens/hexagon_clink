#!/usr/bin/env python3
"""Plot penny graphs from graph6 files."""

import sys
import numpy as np
from scipy.optimize import minimize
import matplotlib.pyplot as plt

def parse_graph6(line, n):
    """Parse graph6 format to edge set."""
    line = line.strip()
    if not line:
        return None

    n_from_line = ord(line[0]) - 63
    if n_from_line != n:
        return None

    # Decode bits
    bits = []
    for c in line[1:]:
        val = ord(c) - 63
        for b in range(5, -1, -1):
            bits.append((val >> b) & 1)

    # Build edge set from upper triangle
    edges = set()
    bit_idx = 0
    for j in range(1, n):
        for i in range(j):
            if bit_idx < len(bits) and bits[bit_idx] == 1:
                edges.add((i, j))
            bit_idx += 1
    return edges

def embedding_cost(positions, edges, non_edges, n):
    pos = positions.reshape(n, 2)
    cost = 0.0
    for (i, j) in edges:
        dist = np.sqrt(np.sum((pos[i] - pos[j])**2))
        cost += (dist - 1.0)**2
    for (i, j) in non_edges:
        dist = np.sqrt(np.sum((pos[i] - pos[j])**2))
        if dist <= 1.0:
            cost += (1.0 - dist + 0.1)**2
    return cost

def find_embedding(n, edges, attempts=20, tol=1e-6):
    if len(edges) == 0:
        return None

    edge_set = set(edges)
    all_pairs = [(i, j) for i in range(n) for j in range(i+1, n)]
    non_edges = [p for p in all_pairs if p not in edge_set]
    edges_list = list(edges)

    for attempt in range(attempts):
        np.random.seed(42 + attempt)
        init_pos = np.random.randn(n, 2) * 1.5
        result = minimize(
            embedding_cost, init_pos.flatten(), args=(edges_list, non_edges, n),
            method='L-BFGS-B', options={'maxiter': 5000}
        )
        if result.fun < tol:
            pos = result.x.reshape(n, 2)
            # Verify
            valid = True
            for (i, j) in edges_list:
                d = np.sqrt(np.sum((pos[i] - pos[j])**2))
                if abs(d - 1.0) > 0.002:
                    valid = False
                    break
            if valid:
                for (i, j) in non_edges:
                    d = np.sqrt(np.sum((pos[i] - pos[j])**2))
                    if d <= 1.002:
                        valid = False
                        break
            if valid:
                return pos
    return None

def main():
    if len(sys.argv) < 3:
        print("Usage: plot_penny.py <n> <input.g6> [output.png]")
        sys.exit(1)

    n = int(sys.argv[1])
    input_file = sys.argv[2]
    output_file = sys.argv[3] if len(sys.argv) > 3 else input_file.replace('.g6', '.png')

    # Read graphs
    graphs = []
    with open(input_file) as f:
        for line in f:
            edges = parse_graph6(line, n)
            if edges:
                graphs.append(edges)

    print(f"Loaded {len(graphs)} graphs from {input_file}")

    if len(graphs) == 0:
        print("No graphs to plot")
        sys.exit(1)

    # Calculate grid size
    cols = min(5, len(graphs))
    rows = (len(graphs) + cols - 1) // cols

    fig, axes = plt.subplots(rows, cols, figsize=(3 * cols, 3 * rows))
    if len(graphs) == 1:
        axes = np.array([axes])
    axes = axes.flatten() if hasattr(axes, 'flatten') else [axes]

    for i, edges in enumerate(graphs):
        ax = axes[i]
        pos = find_embedding(n, edges)

        if pos is None:
            ax.set_title(f"Graph {i+1}\n(embedding failed)")
            ax.axis('off')
            continue

        # Draw edges
        for (a, b) in edges:
            ax.plot([pos[a, 0], pos[b, 0]], [pos[a, 1], pos[b, 1]], 'b-', lw=1.5)

        # Draw nodes
        ax.scatter(pos[:, 0], pos[:, 1], s=200, c='lightblue', edgecolors='black', zorder=5)
        for j in range(n):
            ax.annotate(str(j), (pos[j, 0], pos[j, 1]), ha='center', va='center', fontsize=9, fontweight='bold')

        # Degree sequence
        degrees = [0] * n
        for (a, b) in edges:
            degrees[a] += 1
            degrees[b] += 1

        ax.set_aspect('equal')
        ax.set_title(f"{len(edges)}e, deg={sorted(degrees, reverse=True)}", fontsize=8)
        ax.axis('off')

    # Hide unused axes
    for i in range(len(graphs), len(axes)):
        axes[i].axis('off')

    plt.tight_layout()
    plt.savefig(output_file, dpi=150)
    print(f"Saved to {output_file}")

if __name__ == "__main__":
    main()
