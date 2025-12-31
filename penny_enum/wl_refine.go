package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"
)

var n int
var numEdges int
var edgeIndex [][]int

func initEdges(vertices int) {
	n = vertices
	numEdges = n * (n - 1) / 2
	edgeIndex = make([][]int, n)
	for i := 0; i < n; i++ {
		edgeIndex[i] = make([]int, n)
	}
	idx := 0
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			edgeIndex[i][j] = idx
			edgeIndex[j][i] = idx
			idx++
		}
	}
}

type Graph uint64

func (g Graph) hasEdge(i, j int) bool {
	if i > j {
		i, j = j, i
	}
	return g&(1<<edgeIndex[i][j]) != 0
}

func (g Graph) degree(v int) int {
	count := 0
	for u := 0; u < n; u++ {
		if u != v && g.hasEdge(v, u) {
			count++
		}
	}
	return count
}

func (g Graph) wlFingerprint(iterations int) string {
	colors := make([]int, n)
	for v := 0; v < n; v++ {
		colors[v] = g.degree(v)
	}

	for iter := 0; iter < iterations; iter++ {
		newColors := make([]int, n)
		colorMap := make(map[string]int)
		nextColor := 0

		for v := 0; v < n; v++ {
			var neighColors []int
			for u := 0; u < n; u++ {
				if u != v && g.hasEdge(v, u) {
					neighColors = append(neighColors, colors[u])
				}
			}
			sort.Ints(neighColors)
			sig := fmt.Sprintf("%d:%v", colors[v], neighColors)

			if c, ok := colorMap[sig]; ok {
				newColors[v] = c
			} else {
				colorMap[sig] = nextColor
				newColors[v] = nextColor
				nextColor++
			}
		}
		colors = newColors
	}

	sorted := make([]int, n)
	copy(sorted, colors)
	sort.Ints(sorted)
	return fmt.Sprint(sorted)
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: wl_refine <n> <input_grouped.bin> <output_grouped_wl.bin>")
		fmt.Println("  n: number of vertices")
		fmt.Println("  input_grouped.bin: grouped binary file from refine_hash")
		fmt.Println("  output_grouped_wl.bin: output file with WL-refined groups")
		os.Exit(1)
	}

	vertices, err := strconv.Atoi(os.Args[1])
	if err != nil || vertices < 2 {
		fmt.Println("Error: n must be an integer >= 2")
		os.Exit(1)
	}
	initEdges(vertices)

	inputFile := os.Args[2]
	outputFile := os.Args[3]

	bytesPerGraph := 4
	if numEdges > 32 {
		bytesPerGraph = 8
	}

	f, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("Error opening input file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	reader := bufio.NewReader(f)

	var numGroups uint32
	binary.Read(reader, binary.LittleEndian, &numGroups)
	fmt.Printf("Reading %d groups, refining with WL (n=%d)...\n", numGroups, n)

	start := time.Now()
	totalGraphs := 0
	splitCount := 0

	type groupResult struct {
		graphs []Graph
	}
	var allResults []groupResult

	for g := uint32(0); g < numGroups; g++ {
		var size uint32
		binary.Read(reader, binary.LittleEndian, &size)

		graphs := make([]Graph, size)
		for i := uint32(0); i < size; i++ {
			if bytesPerGraph == 4 {
				var graph uint32
				binary.Read(reader, binary.LittleEndian, &graph)
				graphs[i] = Graph(graph)
			} else {
				var graph uint64
				binary.Read(reader, binary.LittleEndian, &graph)
				graphs[i] = Graph(graph)
			}
		}
		totalGraphs += int(size)

		subgroups := make(map[string][]Graph)
		for _, gr := range graphs {
			fp := gr.wlFingerprint(3)
			subgroups[fp] = append(subgroups[fp], gr)
		}

		if len(subgroups) > 1 {
			splitCount++
			sizes := make([]int, 0, len(subgroups))
			for _, sg := range subgroups {
				sizes = append(sizes, len(sg))
			}
			sort.Sort(sort.Reverse(sort.IntSlice(sizes)))
			fmt.Printf("  Split! Group %d (size %d) -> %d subgroups: %v\n", g, size, len(subgroups), sizes)
		}

		for _, sg := range subgroups {
			allResults = append(allResults, groupResult{sg})
		}

		if (g+1)%100 == 0 {
			fmt.Printf("  Progress: %d/%d groups, %d total subgroups, %d splits (%.1fs)\n",
				g+1, numGroups, len(allResults), splitCount, time.Since(start).Seconds())
		}
	}

	fmt.Printf("\nDone in %v\n", time.Since(start))
	fmt.Printf("Total graphs: %d\n", totalGraphs)
	fmt.Printf("Original groups: %d\n", numGroups)
	fmt.Printf("Refined groups: %d (splits: %d)\n", len(allResults), splitCount)

	outFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	writer := bufio.NewWriter(outFile)
	binary.Write(writer, binary.LittleEndian, uint32(len(allResults)))
	for _, gr := range allResults {
		binary.Write(writer, binary.LittleEndian, uint32(len(gr.graphs)))
		for _, g := range gr.graphs {
			if bytesPerGraph == 4 {
				binary.Write(writer, binary.LittleEndian, uint32(g))
			} else {
				binary.Write(writer, binary.LittleEndian, uint64(g))
			}
		}
	}
	writer.Flush()
	outFile.Close()
	fmt.Printf("Wrote to %s\n", outputFile)

	sizeDist := make(map[int]int)
	for _, gr := range allResults {
		sizeDist[len(gr.graphs)]++
	}
	fmt.Println("\nGroup size distribution:")
	sizes := make([]int, 0)
	for size := range sizeDist {
		sizes = append(sizes, size)
	}
	sort.Ints(sizes)
	for _, size := range sizes {
		fmt.Printf("  size %6d: %d groups\n", size, sizeDist[size])
	}
}
