package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"sync"
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

func (g Graph) neighbors(v int) []int {
	var result []int
	for u := 0; u < n; u++ {
		if u != v && g.hasEdge(v, u) {
			result = append(result, u)
		}
	}
	return result
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

func readGraphs(inputFile string) []Graph {
	bytesPerGraph := 4
	if numEdges > 32 {
		bytesPerGraph = 8
	}

	f, _ := os.Open(inputFile)
	defer f.Close()

	var graphs []Graph
	reader := bufio.NewReader(f)

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
	return graphs
}

// Our optimized pipeline: fingerprint -> WL -> canonical on groups
func benchOurPipeline(graphs []Graph) (int, time.Duration) {
	numWorkers := runtime.NumCPU()
	start := time.Now()

	// Step 1: Fingerprint grouping
	fpGroups := make(map[string][]Graph)
	for _, g := range graphs {
		fp := g.fingerprint()
		fpGroups[fp] = append(fpGroups[fp], g)
	}

	// Step 2: WL refinement
	type group struct {
		graphs []Graph
	}
	var wlGroups []group
	for _, gs := range fpGroups {
		subgroups := make(map[string][]Graph)
		for _, g := range gs {
			wl := g.wlFingerprint(3)
			subgroups[wl] = append(subgroups[wl], g)
		}
		for _, sg := range subgroups {
			wlGroups = append(wlGroups, group{sg})
		}
	}

	// Step 3: Canonical on each group (parallel)
	results := make(chan map[Graph]bool, len(wlGroups))
	groupChan := make(chan int, len(wlGroups))

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for gIdx := range groupChan {
				seen := make(map[Graph]bool)
				for _, gr := range wlGroups[gIdx].graphs {
					canon := gr.canonical()
					seen[canon] = true
				}
				results <- seen
			}
		}()
	}

	go func() {
		for i := range wlGroups {
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

	return len(allUnique), time.Since(start)
}

func benchNautyLabelg(graphs []Graph) (int, time.Duration) {
	tmpFile := "/tmp/bench_compare.g6"
	out, _ := os.Create(tmpFile)
	for _, g := range graphs {
		fmt.Fprintln(out, g.toGraph6())
	}
	out.Close()

	start := time.Now()
	cmd := exec.Command("labelg", "-q", tmpFile)
	outPipe, _ := cmd.StdoutPipe()
	cmd.Start()

	unique := make(map[string]bool)
	scanner := bufio.NewScanner(outPipe)
	for scanner.Scan() {
		unique[scanner.Text()] = true
	}
	cmd.Wait()
	elapsed := time.Since(start)

	os.Remove(tmpFile)
	return len(unique), elapsed
}

func benchNautyShortg(graphs []Graph) (int, time.Duration) {
	tmpFile := "/tmp/bench_compare.g6"
	outFile := "/tmp/bench_compare_out.g6"
	out, _ := os.Create(tmpFile)
	for _, g := range graphs {
		fmt.Fprintln(out, g.toGraph6())
	}
	out.Close()

	start := time.Now()
	cmd := exec.Command("shortg", "-q", tmpFile, outFile)
	cmd.Run()
	elapsed := time.Since(start)

	// Count lines in output file
	f, _ := os.Open(outFile)
	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		count++
	}
	f.Close()

	os.Remove(tmpFile)
	os.Remove(outFile)
	return count, elapsed
}

// Read pre-grouped WL file and only benchmark the canonicalization step
func readGroupedWL(inputFile string) [][]Graph {
	bytesPerGraph := 4
	if numEdges > 32 {
		bytesPerGraph = 8
	}

	f, _ := os.Open(inputFile)
	defer f.Close()
	reader := bufio.NewReader(f)

	var numGroups uint32
	binary.Read(reader, binary.LittleEndian, &numGroups)

	groups := make([][]Graph, numGroups)
	for i := uint32(0); i < numGroups; i++ {
		var size uint32
		binary.Read(reader, binary.LittleEndian, &size)
		groups[i] = make([]Graph, size)
		for j := uint32(0); j < size; j++ {
			if bytesPerGraph == 4 {
				var v uint32
				binary.Read(reader, binary.LittleEndian, &v)
				groups[i][j] = Graph(v)
			} else {
				var v uint64
				binary.Read(reader, binary.LittleEndian, &v)
				groups[i][j] = Graph(v)
			}
		}
	}
	return groups
}

// Benchmark just the canonicalization step on pre-grouped data
func benchCanonicalOnly(groups [][]Graph) (int, time.Duration) {
	numWorkers := runtime.NumCPU()
	start := time.Now()

	results := make(chan map[Graph]bool, len(groups))
	groupChan := make(chan int, len(groups))

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for gIdx := range groupChan {
				seen := make(map[Graph]bool)
				for _, gr := range groups[gIdx] {
					canon := gr.canonical()
					seen[canon] = true
				}
				results <- seen
			}
		}()
	}

	go func() {
		for i := range groups {
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

	return len(allUnique), time.Since(start)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: compare_all <input.bin> <n> [--raw]")
		fmt.Println("  Compares our pipeline vs nauty performance")
		fmt.Println("")
		fmt.Println("  If input is *_grouped_wl.bin, compares just canonicalization step")
		fmt.Println("  Use --raw to force full pipeline comparison on raw graphs")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	vertices, _ := strconv.Atoi(os.Args[2])
	initEdges(vertices)

	// Detect if this is a grouped file or raw file
	isGrouped := len(os.Args) <= 3 // no --raw flag

	var graphs []Graph
	var groups [][]Graph
	var totalGraphs int

	if isGrouped {
		// Try reading as grouped format
		groups = readGroupedWL(inputFile)
		if len(groups) > 0 && len(groups) < 10000000 {
			for _, g := range groups {
				totalGraphs += len(g)
			}
			fmt.Printf("Loaded %d graphs in %d pre-grouped WL groups (n=%d)\n\n", totalGraphs, len(groups), n)
		} else {
			isGrouped = false
		}
	}

	if !isGrouped {
		graphs = readGraphs(inputFile)
		totalGraphs = len(graphs)
		fmt.Printf("Loaded %d raw graphs (n=%d)\n\n", totalGraphs, n)
	}

	// Limit for benchmark
	limit := totalGraphs
	if limit > 300000 {
		limit = 300000
		fmt.Printf("Limiting to %d graphs for benchmark\n\n", limit)
		if isGrouped {
			// Truncate groups
			count := 0
			for i, g := range groups {
				if count+len(g) > limit {
					groups = groups[:i]
					break
				}
				count += len(g)
			}
			totalGraphs = count
		} else {
			graphs = graphs[:limit]
			totalGraphs = limit
		}
	}

	var ourUnique int
	var ourTime time.Duration

	if isGrouped {
		fmt.Println("=== Our canonicalization (on pre-grouped data) ===")
		ourUnique, ourTime = benchCanonicalOnly(groups)
	} else {
		fmt.Println("=== Our full pipeline (fingerprint + WL + canonical) ===")
		ourUnique, ourTime = benchOurPipeline(graphs)
	}
	fmt.Printf("  Time: %v\n", ourTime)
	fmt.Printf("  Rate: %.0f graphs/sec\n", float64(totalGraphs)/ourTime.Seconds())
	fmt.Printf("  Unique: %d\n\n", ourUnique)

	// Flatten groups for nauty comparison
	if isGrouped && graphs == nil {
		for _, g := range groups {
			graphs = append(graphs, g...)
		}
	}

	// Check if nauty is available
	if _, err := exec.LookPath("labelg"); err == nil {
		fmt.Println("=== nauty labelg ===")
		nautyUnique, nautyTime := benchNautyLabelg(graphs)
		fmt.Printf("  Time: %v\n", nautyTime)
		fmt.Printf("  Rate: %.0f graphs/sec\n", float64(len(graphs))/nautyTime.Seconds())
		fmt.Printf("  Unique: %d\n", nautyUnique)
		if nautyTime < ourTime {
			fmt.Printf("  nauty is %.1fx faster\n\n", ourTime.Seconds()/nautyTime.Seconds())
		} else {
			fmt.Printf("  Our method is %.1fx faster\n\n", nautyTime.Seconds()/ourTime.Seconds())
		}

		fmt.Println("=== nauty shortg (deduplicate) ===")
		shortgUnique, shortgTime := benchNautyShortg(graphs)
		fmt.Printf("  Time: %v\n", shortgTime)
		fmt.Printf("  Rate: %.0f graphs/sec\n", float64(len(graphs))/shortgTime.Seconds())
		fmt.Printf("  Unique: %d\n", shortgUnique)
		if shortgTime < ourTime {
			fmt.Printf("  nauty is %.1fx faster\n", ourTime.Seconds()/shortgTime.Seconds())
		} else {
			fmt.Printf("  Our method is %.1fx faster\n", shortgTime.Seconds()/ourTime.Seconds())
		}
	} else {
		fmt.Println("nauty not found. Install with: brew install nauty")
	}
}
