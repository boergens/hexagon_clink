package main

import (
	"bufio"
	"fmt"
	"os"
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

func (g Graph) degrees() []int {
	deg := make([]int, n)
	for idx := 0; idx < numEdges; idx++ {
		if g&(1<<idx) != 0 {
			deg[edgePairs[idx][0]]++
			deg[edgePairs[idx][1]]++
		}
	}
	return deg
}

func (g Graph) hasIsolated() bool {
	deg := g.degrees()
	for i := 0; i < n; i++ {
		if deg[i] == 0 {
			return true
		}
	}
	return false
}

func (g Graph) isConnected() bool {
	if g == 0 {
		return false
	}
	adj := make([]uint64, n)
	for idx := 0; idx < numEdges; idx++ {
		if g&(1<<idx) != 0 {
			i, j := edgePairs[idx][0], edgePairs[idx][1]
			adj[i] |= 1 << j
			adj[j] |= 1 << i
		}
	}
	visited := uint64(1)
	queue := []int{0}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		neighbors := adj[node]
		for neighbor := 0; neighbor < n; neighbor++ {
			if neighbors&(1<<neighbor) != 0 && visited&(1<<neighbor) == 0 {
				visited |= 1 << neighbor
				queue = append(queue, neighbor)
			}
		}
	}
	return visited == (1<<n)-1
}

func (g Graph) maxDegree() int {
	deg := g.degrees()
	max := 0
	for i := 0; i < n; i++ {
		if deg[i] > max {
			max = deg[i]
		}
	}
	return max
}

func (g Graph) hasK4() bool {
	for a := 0; a < n; a++ {
		for b := a + 1; b < n; b++ {
			if g&(1<<edgeIndex[a][b]) == 0 {
				continue
			}
			for c := b + 1; c < n; c++ {
				if g&(1<<edgeIndex[a][c]) == 0 || g&(1<<edgeIndex[b][c]) == 0 {
					continue
				}
				for d := c + 1; d < n; d++ {
					if g&(1<<edgeIndex[a][d]) != 0 && g&(1<<edgeIndex[b][d]) != 0 && g&(1<<edgeIndex[c][d]) != 0 {
						return true
					}
				}
			}
		}
	}
	return false
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: generate_edges <n> <edges> <output.bin>")
		fmt.Println("  n: number of vertices")
		fmt.Println("  edges: exact number of edges")
		fmt.Println("  output.bin: output file for candidate graphs")
		fmt.Println("\nFilters: connected, no isolated vertices, max degree <= 6, no K4")
		os.Exit(1)
	}

	vertices, err := strconv.Atoi(os.Args[1])
	if err != nil || vertices < 2 {
		fmt.Println("Error: n must be an integer >= 2")
		os.Exit(1)
	}
	initEdges(vertices)

	targetEdges, err := strconv.Atoi(os.Args[2])
	if err != nil || targetEdges < 1 || targetEdges > numEdges {
		fmt.Printf("Error: edges must be between 1 and %d\n", numEdges)
		os.Exit(1)
	}

	outputFile := os.Args[3]

	bytesPerGraph := 4
	if numEdges > 32 {
		bytesPerGraph = 8
	}

	fmt.Printf("=== Generating n=%d candidates with %d edges ===\n", n, targetEdges)
	fmt.Printf("Max possible edges: %d, bytes per graph: %d\n\n", numEdges, bytesPerGraph)

	outFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()
	writer := bufio.NewWriter(outFile)

	start := time.Now()
	total := 0
	written := 0

	var generate func(start int, current Graph, remaining int)
	generate = func(startIdx int, current Graph, remaining int) {
		if remaining == 0 {
			total++
			if !current.hasIsolated() && current.maxDegree() <= 6 && current.isConnected() && !current.hasK4() {
				if bytesPerGraph == 4 {
					b := []byte{byte(current), byte(current >> 8), byte(current >> 16), byte(current >> 24)}
					writer.Write(b)
				} else {
					b := []byte{
						byte(current), byte(current >> 8), byte(current >> 16), byte(current >> 24),
						byte(current >> 32), byte(current >> 40), byte(current >> 48), byte(current >> 56),
					}
					writer.Write(b)
				}
				written++
			}
			if total%10000000 == 0 {
				fmt.Printf("  Processed %dM, written %d...\n", total/1000000, written)
			}
			return
		}
		if startIdx+remaining > numEdges {
			return
		}
		for i := startIdx; i <= numEdges-remaining; i++ {
			generate(i+1, current|(1<<i), remaining-1)
		}
	}

	generate(0, 0, targetEdges)
	writer.Flush()

	elapsed := time.Since(start)
	fmt.Printf("\nDone in %v\n", elapsed)
	fmt.Printf("Total graphs checked: %d\n", total)
	fmt.Printf("Candidates written: %d\n", written)

	info, _ := outFile.Stat()
	fmt.Printf("File size: %.1f MB\n", float64(info.Size())/1024/1024)
}
