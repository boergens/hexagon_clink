package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
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

func (g Graph) canonical() Graph {
	best := g
	perm := make([]int, n)
	for i := range perm {
		perm[i] = i
	}

	var generate func(k int)
	generate = func(k int) {
		if k == 1 {
			var relabeled Graph
			for idx := 0; idx < numEdges; idx++ {
				if g&(1<<idx) != 0 {
					i, j := edgePairs[idx][0], edgePairs[idx][1]
					ni, nj := perm[i], perm[j]
					if ni > nj {
						ni, nj = nj, ni
					}
					relabeled |= 1 << edgeIndex[ni][nj]
				}
			}
			if relabeled < best {
				best = relabeled
			}
			return
		}
		for i := 0; i < k; i++ {
			generate(k - 1)
			if k%2 == 0 {
				perm[i], perm[k-1] = perm[k-1], perm[i]
			} else {
				perm[0], perm[k-1] = perm[k-1], perm[0]
			}
		}
	}
	generate(n)
	return best
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: canonicalize <n> <input_grouped_wl.bin> <output_prefix>")
		fmt.Println("  n: number of vertices")
		fmt.Println("  input_grouped_wl.bin: WL-refined grouped file")
		fmt.Println("  output_prefix: prefix for output files (creates <prefix>.bin and <prefix>.txt)")
		os.Exit(1)
	}

	vertices, err := strconv.Atoi(os.Args[1])
	if err != nil || vertices < 2 {
		fmt.Println("Error: n must be an integer >= 2")
		os.Exit(1)
	}
	initEdges(vertices)

	inputFile := os.Args[2]
	outputPrefix := os.Args[3]

	bytesPerGraph := 4
	if numEdges > 32 {
		bytesPerGraph = 8
	}

	numWorkers := runtime.NumCPU()
	fmt.Printf("Using %d workers (n=%d, %d bytes/graph)\n", numWorkers, n, bytesPerGraph)

	f, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("Error opening input file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	reader := bufio.NewReader(f)

	var numGroups uint32
	binary.Read(reader, binary.LittleEndian, &numGroups)
	fmt.Printf("Canonicalizing %d groups...\n", numGroups)

	type group struct {
		graphs []Graph
	}
	groups := make([]group, numGroups)
	totalGraphs := 0
	for g := uint32(0); g < numGroups; g++ {
		var size uint32
		binary.Read(reader, binary.LittleEndian, &size)
		groups[g].graphs = make([]Graph, size)
		for i := uint32(0); i < size; i++ {
			if bytesPerGraph == 4 {
				var graph uint32
				binary.Read(reader, binary.LittleEndian, &graph)
				groups[g].graphs[i] = Graph(graph)
			} else {
				var graph uint64
				binary.Read(reader, binary.LittleEndian, &graph)
				groups[g].graphs[i] = Graph(graph)
			}
		}
		totalGraphs += int(size)
	}
	fmt.Printf("Loaded %d graphs in %d groups\n", totalGraphs, numGroups)

	start := time.Now()
	var canonCalls atomic.Int64
	var groupsDone atomic.Int64

	results := make(chan map[Graph]bool, numGroups)
	groupChan := make(chan int, numGroups)

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for gIdx := range groupChan {
				seen := make(map[Graph]bool)
				for _, gr := range groups[gIdx].graphs {
					canonCalls.Add(1)
					canon := gr.canonical()
					seen[canon] = true
				}
				results <- seen
				done := groupsDone.Add(1)
				if done%50 == 0 {
					fmt.Printf("  %d/%d groups done (%.1fs)\n", done, numGroups, time.Since(start).Seconds())
				}
			}
		}()
	}

	go func() {
		for i := 0; i < int(numGroups); i++ {
			groupChan <- i
		}
		close(groupChan)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	allUnique := make(map[Graph]bool)
	for seen := range results {
		for g := range seen {
			allUnique[g] = true
		}
	}

	fmt.Printf("\nDone in %v\n", time.Since(start))
	fmt.Printf("Total graphs: %d\n", totalGraphs)
	fmt.Printf("Canonical calls: %d\n", canonCalls.Load())
	fmt.Printf("Unique graphs: %d\n", len(allUnique))

	outFile, err := os.Create(outputPrefix + ".bin")
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	writer := bufio.NewWriter(outFile)
	for g := range allUnique {
		if bytesPerGraph == 4 {
			binary.Write(writer, binary.LittleEndian, uint32(g))
		} else {
			binary.Write(writer, binary.LittleEndian, uint64(g))
		}
	}
	writer.Flush()
	outFile.Close()
	fmt.Printf("Wrote %d unique graphs to %s.bin\n", len(allUnique), outputPrefix)

	txtFile, _ := os.Create(outputPrefix + ".txt")
	var sorted []Graph
	for g := range allUnique {
		sorted = append(sorted, g)
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	for _, g := range sorted {
		fmt.Fprintf(txtFile, "%d\n", g)
	}
	txtFile.Close()
	fmt.Printf("Wrote %d unique graphs to %s.txt\n", len(allUnique), outputPrefix)
}
