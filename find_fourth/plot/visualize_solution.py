"""
Visualize hexagon grid arrangements as PNG.
Shows which items are adjacent in each arrangement.
"""

import matplotlib.pyplot as plt
import matplotlib.patches as patches
from itertools import combinations

from hex_spiral import build_spiral


def draw_arrangement(ax, arrangement, positions, adjacencies, title, new_pairs=None):
    ax.set_aspect('equal')

    xs = [p[0] for p in positions.values()]
    ys = [p[1] for p in positions.values()]
    margin = 1.5
    ax.set_xlim(min(xs) - margin, max(xs) + margin)
    ax.set_ylim(min(ys) - margin, max(ys) + margin)
    ax.set_title(title, fontsize=14, fontweight='bold')
    ax.axis('off')

    slot_to_item = {}
    for slot, item in enumerate(arrangement):
        slot_to_item[slot] = item

    if new_pairs is None:
        new_pairs = set()

    for slot_a, slot_b in adjacencies:
        xa, ya = positions[slot_a]
        xb, yb = positions[slot_b]
        item_a = slot_to_item[slot_a]
        item_b = slot_to_item[slot_b]
        pair = (min(item_a, item_b), max(item_a, item_b))
        if pair in new_pairs:
            ax.plot([xa, xb], [ya, yb], color='#22aa22', linewidth=4, alpha=0.9, zorder=1)
        else:
            ax.plot([xa, xb], [ya, yb], color='#aaaaaa', linewidth=2, alpha=0.5, zorder=0)

    for slot, (x, y) in positions.items():
        hex_patch = patches.RegularPolygon(
            (x, y), numVertices=6, radius=0.7,
            orientation=0,
            facecolor='lightyellow', edgecolor='black', linewidth=2
        )
        ax.add_patch(hex_patch)

        item = slot_to_item[slot]
        ax.text(x, y, str(item), ha='center', va='center',
                fontsize=14, fontweight='bold')


def get_covered_pairs(arrangement, adjacencies):
    slot_to_item = {}
    for slot, item in enumerate(arrangement):
        slot_to_item[slot] = item

    pairs = set()
    for slot_a, slot_b in adjacencies:
        item_a = slot_to_item[slot_a]
        item_b = slot_to_item[slot_b]
        pairs.add((min(item_a, item_b), max(item_a, item_b)))
    return pairs


def visualize_solution(arrangements, n, output_file="solution.png"):
    positions, adjacencies = build_spiral(n)

    fig, axes = plt.subplots(1, len(arrangements), figsize=(6*len(arrangements), 8))
    if len(arrangements) == 1:
        axes = [axes]

    all_pairs = set(combinations(range(n), 2))
    covered_so_far = set()

    for i, (ax, arr) in enumerate(zip(axes, arrangements)):
        covered = get_covered_pairs(arr, adjacencies)
        new_covered = covered - covered_so_far
        covered_so_far |= covered

        title = f"Arrangement {i+1}\n({len(covered)} edges, {len(new_covered)} new)"
        draw_arrangement(ax, arr, positions, adjacencies, title, new_pairs=new_covered)

    missing = all_pairs - covered_so_far
    fig.suptitle(
        f"{n}-Node Solution: {len(covered_so_far)}/{len(all_pairs)} pairs covered" +
        (f"\nMissing: {missing}" if missing else " ✓"),
        fontsize=16, y=0.02
    )

    plt.tight_layout()
    plt.savefig(output_file, dpi=150, bbox_inches='tight', facecolor='white')
    print(f"Saved to {output_file}")

    print(f"\nTotal unique pairs: {len(covered_so_far)}/{len(all_pairs)}")
    if missing:
        print(f"Missing: {sorted(missing)}")
    else:
        print("All pairs covered!")
