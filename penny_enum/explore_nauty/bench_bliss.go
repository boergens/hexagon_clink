package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

var n int
var numEdges int
var edgeIndex [][]int
var edgePairs [][2]int

func initEdges(vertices int) {
	n = vertices
	numEdges = n * (n - 1) / 2
	edgeIndex = make([][]int, n)
	for i := 0; i < n; i++ {
		edgeIndex[i] = make([]int, n)
	}
	edgePairs = make([][2]int, numEdges)
	idx := 0
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			edgeIndex[i][j] = idx
			edgeIndex[j][i] = idx
			edgePairs[idx] = [2]int{i, j}
			idx++
		}
	}
}

type Graph uint64

func (g Graph) toDIMACS() string {
	edgeCount := 0
	for idx := 0; idx < numEdges; idx++ {
		if g&(1<<idx) != 0 {
			edgeCount++
		}
	}

	result := fmt.Sprintf("p edge %d %d\n", n, edgeCount)
	for idx := 0; idx < numEdges; idx++ {
		if g&(1<<idx) != 0 {
			i, j := edgePairs[idx][0], edgePairs[idx][1]
			result += fmt.Sprintf("e %d %d\n", i+1, j+1)
		}
	}
	return result
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: bench_bliss <input.bin> <n>")
		fmt.Println("  Benchmarks bliss on binary graph file")
		fmt.Println("")
		fmt.Println("Install bliss: brew install bliss")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	vertices, _ := strconv.Atoi(os.Args[2])
	initEdges(vertices)

	bytesPerGraph := 4
	if numEdges > 32 {
		bytesPerGraph = 8
	}

	// Check if bliss exists
	blissPath, err := exec.LookPath("bliss")
	if err != nil {
		fmt.Println("Error: bliss not found. Install with: brew install bliss")
		os.Exit(1)
	}
	fmt.Printf("Using bliss: %s\n", blissPath)

	// Read graphs
	f, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	var graphs []Graph
	reader := bufio.NewReader(f)

	// Detect format (grouped vs raw)
	info, _ := f.Stat()
	fileSize := info.Size()
	var numGroups uint32
	binary.Read(reader, binary.LittleEndian, &numGroups)

	if int64(numGroups)*4 > fileSize || numGroups > 10000000 {
		f.Seek(0, 0)
		reader = bufio.NewReader(f)
		buf := make([]byte, bytesPerGraph)
		for {
			_, err := reader.Read(buf)
			if err != nil {
				break
			}
			var g Graph
			if bytesPerGraph == 4 {
				g = Graph(binary.LittleEndian.Uint32(buf))
			} else {
				g = Graph(binary.LittleEndian.Uint64(buf))
			}
			graphs = append(graphs, g)
		}
	} else {
		for i := uint32(0); i < numGroups; i++ {
			var size uint32
			binary.Read(reader, binary.LittleEndian, &size)
			for j := uint32(0); j < size; j++ {
				var g Graph
				if bytesPerGraph == 4 {
					var v uint32
					binary.Read(reader, binary.LittleEndian, &v)
					g = Graph(v)
				} else {
					var v uint64
					binary.Read(reader, binary.LittleEndian, &v)
					g = Graph(v)
				}
				graphs = append(graphs, g)
			}
		}
	}

	fmt.Printf("Read %d graphs (n=%d)\n", len(graphs), n)

	// Limit for benchmark
	limit := len(graphs)
	if limit > 10000 {
		limit = 10000
		fmt.Printf("Limiting to %d graphs for benchmark\n", limit)
	}

	// Create temp file for each graph and run bliss
	fmt.Println("\n=== bliss canonical hash ===")
	tmpFile := "/tmp/bench_graph.dimacs"

	unique := make(map[string]bool)
	start := time.Now()

	for i := 0; i < limit; i++ {
		// Write graph to temp file
		out, _ := os.Create(tmpFile)
		fmt.Fprint(out, graphs[i].toDIMACS())
		out.Close()

		// Run bliss with canonical hash output
		cmd := exec.Command("bliss", "-canonical", tmpFile)
		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error on graph %d: %v\n", i, err)
			continue
		}

		// Extract canonical hash from output
		unique[string(output)] = true

		if (i+1)%1000 == 0 {
			elapsed := time.Since(start)
			fmt.Printf("  %d/%d graphs (%.0f/sec)\n", i+1, limit, float64(i+1)/elapsed.Seconds())
		}
	}

	elapsed := time.Since(start)
	fmt.Printf("\nTime: %v\n", elapsed)
	fmt.Printf("Graphs/sec: %.0f\n", float64(limit)/elapsed.Seconds())
	fmt.Printf("Unique canonical forms: %d\n", len(unique))

	os.Remove(tmpFile)
}
