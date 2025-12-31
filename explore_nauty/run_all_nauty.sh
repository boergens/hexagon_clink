#!/bin/bash

# Process n8 graphs with various edge counts using nauty
# Runs multiple edge counts in parallel

set -e

cd "$(dirname "$0")/.."

N=8

# Build tools if needed
if [ ! -f generate_edges ]; then
    echo "Building generate_edges..."
    go build -o generate_edges generate_edges.go
fi

if [ ! -f explore_nauty/convert ]; then
    echo "Building convert..."
    (cd explore_nauty && go build -o convert convert.go)
fi

# Function to process one edge count
process_edges() {
    local edges=$1
    local logfile="explore_nauty/log_${edges}.txt"

    echo "[edges=$edges] Starting..." | tee "$logfile"

    edges_file="n${N}_${edges}edges.bin"
    g6_file="explore_nauty/n${N}_${edges}.g6"
    unique_file="explore_nauty/n${N}_${edges}_unique.g6"

    # Generate edges file if it doesn't exist
    if [ ! -f "$edges_file" ]; then
        echo "[edges=$edges] Generating $edges_file..." | tee -a "$logfile"
        ./generate_edges $N $edges "$edges_file" 2>&1 | tee -a "$logfile"
    fi

    # Check if file has any graphs
    filesize=$(stat -f%z "$edges_file" 2>/dev/null || stat -c%s "$edges_file" 2>/dev/null)
    if [ "$filesize" -lt 4 ]; then
        echo "[edges=$edges] No graphs, skipping..." | tee -a "$logfile"
        return
    fi

    # Convert to graph6
    echo "[edges=$edges] Converting to g6..." | tee -a "$logfile"
    ./explore_nauty/convert "$edges_file" "$g6_file" $N raw 2>&1 | tee -a "$logfile"

    input_count=$(wc -l < "$g6_file")
    echo "[edges=$edges] Input: $input_count graphs" | tee -a "$logfile"

    # Run nauty shortg
    echo "[edges=$edges] Running shortg..." | tee -a "$logfile"
    start_time=$(date +%s)
    shortg -q "$g6_file" "$unique_file" 2>&1 | tee -a "$logfile"
    end_time=$(date +%s)

    unique_count=$(wc -l < "$unique_file")
    elapsed=$((end_time - start_time))
    echo "[edges=$edges] Done: $unique_count unique in ${elapsed}s" | tee -a "$logfile"
}

export -f process_edges
export N

echo "=== Processing n=$N graphs with nauty (parallel) ==="
echo ""

# Run edges 13,12,11,10,9,8 in parallel (6 processes)
for edges in 13 12 11 10 9 8; do
    process_edges $edges &
done

echo "Waiting for all jobs to complete..."
wait

echo ""
echo "=========================================="
echo "Summary"
echo "=========================================="
for edges in 13 12 11 10 9 8 7; do
    unique_file="explore_nauty/n${N}_${edges}_unique.g6"
    if [ -f "$unique_file" ]; then
        count=$(wc -l < "$unique_file")
        echo "n${N}_${edges}: $count unique graphs"
    fi
done
