package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

type Graph uint64

var n int
var numEdges int
var edgeIndex [][]int
var edgePairs [][2]int
var allPerms [][]int

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
	allPerms = permutations(n)
}

func permutations(n int) [][]int {
	if n == 0 {
		return [][]int{{}}
	}
	var result [][]int
	arr := make([]int, n)
	for i := range arr {
		arr[i] = i
	}
	var generate func(k int)
	generate = func(k int) {
		if k == 1 {
			perm := make([]int, n)
			copy(perm, arr)
			result = append(result, perm)
			return
		}
		for i := 0; i < k; i++ {
			generate(k - 1)
			if k%2 == 0 {
				arr[i], arr[k-1] = arr[k-1], arr[i]
			} else {
				arr[0], arr[k-1] = arr[k-1], arr[0]
			}
		}
	}
	generate(n)
	return result
}

func (g Graph) edgeCount() int {
	count := 0
	tmp := g
	for tmp != 0 {
		count += int(tmp & 1)
		tmp >>= 1
	}
	return count
}

// Check if g is isomorphic to a subgraph of other
func (g Graph) isIsomorphicSubgraphOf(other Graph) bool {
	for _, perm := range allPerms {
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
		// Check if relabeled is a subset of other
		if relabeled&other == relabeled {
			return true
		}
	}
	return false
}

func parseGraph6(line string) Graph {
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return 0
	}
	nFromLine := int(line[0]) - 63
	if nFromLine != n {
		return 0
	}

	var bits []byte
	for i := 1; i < len(line); i++ {
		val := int(line[i]) - 63
		for b := 5; b >= 0; b-- {
			bits = append(bits, byte((val>>b)&1))
		}
	}

	var g Graph
	bitIdx := 0
	for j := 1; j < n; j++ {
		for i := 0; i < j; i++ {
			if bitIdx < len(bits) && bits[bitIdx] == 1 {
				g |= 1 << edgeIndex[i][j]
			}
			bitIdx++
		}
	}
	return g
}

func (g Graph) toGraph6() string {
	result := []byte{byte(n + 63)}
	var bits []byte
	for j := 1; j < n; j++ {
		for i := 0; i < j; i++ {
			if g&(1<<edgeIndex[i][j]) != 0 {
				bits = append(bits, 1)
			} else {
				bits = append(bits, 0)
			}
		}
	}
	for len(bits)%6 != 0 {
		bits = append(bits, 0)
	}
	for i := 0; i < len(bits); i += 6 {
		val := bits[i]<<5 | bits[i+1]<<4 | bits[i+2]<<3 | bits[i+3]<<2 | bits[i+4]<<1 | bits[i+5]
		result = append(result, byte(val+63))
	}
	return string(result)
}

func main() {
	nFlag := flag.Int("n", 8, "number of vertices")
	outputFile := flag.String("out", "", "output file for maximal graphs")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Println("Usage: filter_maximal -n <vertices> [-out output.g6] <input1.g6> [input2.g6] ...")
		fmt.Println("  Reads multiple g6 files and outputs only maximal graphs (not subgraph of any other)")
		os.Exit(1)
	}

	initEdges(*nFlag)

	// Read all graphs from all input files
	var allGraphs []Graph
	for _, inputFile := range flag.Args() {
		f, err := os.Open(inputFile)
		if err != nil {
			fmt.Printf("Error opening %s: %v\n", inputFile, err)
			continue
		}
		scanner := bufio.NewScanner(f)
		count := 0
		for scanner.Scan() {
			g := parseGraph6(scanner.Text())
			if g != 0 {
				allGraphs = append(allGraphs, g)
				count++
			}
		}
		f.Close()
		fmt.Printf("Read %d graphs from %s\n", count, inputFile)
	}

	fmt.Printf("Total: %d graphs\n", len(allGraphs))

	// Sort by edge count descending (larger graphs first)
	sort.Slice(allGraphs, func(i, j int) bool {
		return allGraphs[i].edgeCount() > allGraphs[j].edgeCount()
	})

	// Filter: keep only maximal graphs
	var maximal []Graph
	for i, g := range allGraphs {
		if i%100 == 0 {
			fmt.Printf("\rProcessing %d/%d, maximal so far: %d   ", i, len(allGraphs), len(maximal))
		}

		isSubgraph := false
		for _, m := range maximal {
			if g.isIsomorphicSubgraphOf(m) {
				isSubgraph = true
				break
			}
		}
		if !isSubgraph {
			maximal = append(maximal, g)
		}
	}
	fmt.Printf("\rProcessing %d/%d, maximal: %d           \n", len(allGraphs), len(allGraphs), len(maximal))

	// Group by edge count for summary
	byEdges := make(map[int]int)
	for _, g := range maximal {
		byEdges[g.edgeCount()]++
	}

	fmt.Printf("\nMaximal graphs by edge count:\n")
	var edgeCounts []int
	for e := range byEdges {
		edgeCounts = append(edgeCounts, e)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(edgeCounts)))
	for _, e := range edgeCounts {
		fmt.Printf("  %d edges: %d graphs\n", e, byEdges[e])
	}

	// Write output
	if *outputFile != "" {
		out, err := os.Create(*outputFile)
		if err != nil {
			fmt.Printf("Error creating %s: %v\n", *outputFile, err)
			os.Exit(1)
		}
		for _, g := range maximal {
			fmt.Fprintln(out, g.toGraph6())
		}
		out.Close()
		fmt.Printf("\nWrote %d maximal graphs to %s\n", len(maximal), *outputFile)
	}
}
