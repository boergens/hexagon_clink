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

func (g Graph) neighbors(v int) []int {
	var result []int
	for u := 0; u < n; u++ {
		if u != v && g.hasEdge(v, u) {
			result = append(result, u)
		}
	}
	return result
}

func (g Graph) degree(v int) int {
	return len(g.neighbors(v))
}

func (g Graph) fingerprint() string {
	type vertexInfo struct {
		degree    int
		triangles int
		neighDegs []int
	}

	infos := make([]vertexInfo, n)
	for v := 0; v < n; v++ {
		neighs := g.neighbors(v)
		infos[v].degree = len(neighs)

		for i := 0; i < len(neighs); i++ {
			for j := i + 1; j < len(neighs); j++ {
				if g.hasEdge(neighs[i], neighs[j]) {
					infos[v].triangles++
				}
			}
		}

		for _, u := range neighs {
			infos[v].neighDegs = append(infos[v].neighDegs, g.degree(u))
		}
		sort.Ints(infos[v].neighDegs)
	}

	type infoKey struct {
		degree    int
		triangles int
		neighDegs string
	}
	keys := make([]infoKey, n)
	for v := 0; v < n; v++ {
		keys[v] = infoKey{
			infos[v].degree,
			infos[v].triangles,
			fmt.Sprint(infos[v].neighDegs),
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].degree != keys[j].degree {
			return keys[i].degree > keys[j].degree
		}
		if keys[i].triangles != keys[j].triangles {
			return keys[i].triangles > keys[j].triangles
		}
		return keys[i].neighDegs < keys[j].neighDegs
	})

	return fmt.Sprint(keys)
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: refine_hash <n> <input.bin> <output.bin>")
		fmt.Println("  n: number of vertices")
		fmt.Println("  input.bin: binary file with graphs (each graph is uint32 or uint64)")
		fmt.Println("  output.bin: output file for grouped graphs")
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

	start := time.Now()
	groups := make(map[string][]Graph)
	total := 0

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
		fp := g.fingerprint()
		groups[fp] = append(groups[fp], g)
		total++
		if total%1000000 == 0 {
			fmt.Printf("  Processed %dM, %d groups so far...\n", total/1000000, len(groups))
		}
	}

	fmt.Printf("\nDone fingerprinting in %v\n", time.Since(start))
	fmt.Printf("n=%d, numEdges=%d, bytesPerGraph=%d\n", n, numEdges, bytesPerGraph)
	fmt.Printf("Total: %d\n", total)
	fmt.Printf("Fingerprint groups: %d\n", len(groups))

	outFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()
	writer := bufio.NewWriter(outFile)

	numGroups := uint32(len(groups))
	binary.Write(writer, binary.LittleEndian, numGroups)

	for _, gs := range groups {
		size := uint32(len(gs))
		binary.Write(writer, binary.LittleEndian, size)
		for _, g := range gs {
			if bytesPerGraph == 4 {
				binary.Write(writer, binary.LittleEndian, uint32(g))
			} else {
				binary.Write(writer, binary.LittleEndian, uint64(g))
			}
		}
	}
	writer.Flush()

	info, _ := outFile.Stat()
	fmt.Printf("Wrote grouped data to %s (%.1f MB)\n", outputFile, float64(info.Size())/1024/1024)

	sizeDist := make(map[int]int)
	for _, gs := range groups {
		sizeDist[len(gs)]++
	}

	fmt.Printf("\nGroup size distribution:\n")
	sizes := make([]int, 0)
	for size := range sizeDist {
		sizes = append(sizes, size)
	}
	sort.Ints(sizes)

	for _, size := range sizes {
		fmt.Printf("  size %6d: %d groups\n", size, sizeDist[size])
	}
}
