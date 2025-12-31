package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Graph uint64

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

func (g Graph) hasEdge(i, j int) bool {
	if i > j {
		i, j = j, i
	}
	return g&(1<<edgeIndex[i][j]) != 0
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

func (g Graph) edges() [][2]int {
	var result [][2]int
	for idx := 0; idx < numEdges; idx++ {
		if g&(1<<idx) != 0 {
			result = append(result, edgePairs[idx])
		}
	}
	return result
}

// Check for K4 subgraph (4 nodes all pairwise connected)
// K4 is forbidden in penny graphs
func (g Graph) hasK4() bool {
	for a := 0; a < n; a++ {
		for b := a + 1; b < n; b++ {
			if !g.hasEdge(a, b) {
				continue
			}
			for c := b + 1; c < n; c++ {
				if !g.hasEdge(a, c) || !g.hasEdge(b, c) {
					continue
				}
				for d := c + 1; d < n; d++ {
					if g.hasEdge(a, d) && g.hasEdge(b, d) && g.hasEdge(c, d) {
						return true
					}
				}
			}
		}
	}
	return false
}

// Numerical embedding check using gradient descent
// Returns true if graph can be embedded with edges=1, non-edges>1
func (g Graph) isPennyGraph() bool {
	edges := g.edges()
	if len(edges) == 0 {
		return false
	}

	// Non-edges
	var nonEdges [][2]int
	for idx := 0; idx < numEdges; idx++ {
		if g&(1<<idx) == 0 {
			nonEdges = append(nonEdges, edgePairs[idx])
		}
	}

	// Try multiple random starts
	for attempt := 0; attempt < 20; attempt++ {
		pos := make([][2]float64, n)
		rng := rand.New(rand.NewSource(int64(42 + attempt)))

		// Initialize with spread-out random positions
		for i := 0; i < n; i++ {
			pos[i] = [2]float64{rng.Float64() * 2, rng.Float64() * 2}
		}

		// Gradient descent
		for iter := 0; iter < 3000; iter++ {
			grad := make([][2]float64, n)
			cost := 0.0

			// Edge constraints: distance should be 1
			for _, e := range edges {
				i, j := e[0], e[1]
				dx := pos[j][0] - pos[i][0]
				dy := pos[j][1] - pos[i][1]
				dist := math.Sqrt(dx*dx + dy*dy)
				if dist < 1e-10 {
					dist = 1e-10
				}
				err := dist - 1.0
				cost += err * err

				factor := 2 * err / dist
				grad[i][0] -= factor * dx
				grad[i][1] -= factor * dy
				grad[j][0] += factor * dx
				grad[j][1] += factor * dy
			}

			// Non-edge constraints: distance should be > 1
			for _, e := range nonEdges {
				i, j := e[0], e[1]
				dx := pos[j][0] - pos[i][0]
				dy := pos[j][1] - pos[i][1]
				dist := math.Sqrt(dx*dx + dy*dy)
				if dist < 1e-10 {
					dist = 1e-10
				}
				if dist < 1.0 {
					err := 1.0 - dist + 0.1
					cost += err * err

					factor := -2 * err / dist
					grad[i][0] -= factor * dx
					grad[i][1] -= factor * dy
					grad[j][0] += factor * dx
					grad[j][1] += factor * dy
				}
			}

			// Update positions
			lr := 0.1
			if iter > 1000 {
				lr = 0.01
			}
			if iter > 2000 {
				lr = 0.001
			}
			for i := 0; i < n; i++ {
				pos[i][0] -= lr * grad[i][0]
				pos[i][1] -= lr * grad[i][1]
			}

			if cost < 1e-10 {
				break
			}
		}

		// Verify solution
		valid := true
		for _, e := range edges {
			i, j := e[0], e[1]
			dx := pos[j][0] - pos[i][0]
			dy := pos[j][1] - pos[i][1]
			dist := math.Sqrt(dx*dx + dy*dy)
			if math.Abs(dist-1.0) > 0.001 {
				valid = false
				break
			}
		}
		if valid {
			for _, e := range nonEdges {
				i, j := e[0], e[1]
				dx := pos[j][0] - pos[i][0]
				dy := pos[j][1] - pos[i][1]
				dist := math.Sqrt(dx*dx + dy*dy)
				if dist <= 1.001 {
					valid = false
					break
				}
			}
		}
		if valid {
			return true
		}
	}
	return false
}

// Parse graph6 format to Graph
func parseGraph6(line string) Graph {
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return 0
	}

	// First byte encodes n (for n < 63)
	nFromLine := int(line[0]) - 63
	if nFromLine != n {
		return 0 // Skip if different n
	}

	// Decode the rest as 6-bit chunks representing upper triangle
	var bits []byte
	for i := 1; i < len(line); i++ {
		val := int(line[i]) - 63
		for b := 5; b >= 0; b-- {
			bits = append(bits, byte((val>>b)&1))
		}
	}

	// Build graph from upper triangle bits
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

// Convert Graph to graph6 format
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
	inputFile := flag.String("in", "", "input file (.g6 or .bin)")
	outputFile := flag.String("out", "", "output file (same format as input)")
	workers := flag.Int("workers", 0, "number of workers (default: NumCPU)")
	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Usage: verify_penny -n <vertices> -in <input> -out <output>")
		fmt.Println("  Supports .g6 (graph6) and .bin (binary) formats")
		os.Exit(1)
	}

	if *workers == 0 {
		*workers = runtime.NumCPU()
	}

	initEdges(*nFlag)
	bytesPerGraph := 4
	if numEdges > 32 {
		bytesPerGraph = 8
	}

	// Detect format from extension
	isG6 := strings.HasSuffix(*inputFile, ".g6")

	// Read graphs
	var graphs []Graph
	f, err := os.Open(*inputFile)
	if err != nil {
		fmt.Printf("Error opening %s: %v\n", *inputFile, err)
		os.Exit(1)
	}

	if isG6 {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			g := parseGraph6(scanner.Text())
			if g != 0 {
				graphs = append(graphs, g)
			}
		}
	} else {
		reader := bufio.NewReader(f)
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
	}
	f.Close()

	fmt.Printf("Loaded %d graphs from %s\n", len(graphs), *inputFile)
	fmt.Printf("Using %d workers\n", *workers)

	start := time.Now()

	// Phase 1: K4 pruning (fast, single-threaded)
	fmt.Println("\nPhase 1: K4 pruning...")
	var candidates []Graph
	for _, g := range graphs {
		if !g.hasK4() {
			candidates = append(candidates, g)
		}
	}
	fmt.Printf("After K4 prune: %d graphs (removed %d)\n", len(candidates), len(graphs)-len(candidates))

	// Phase 2: Parallel penny graph verification
	fmt.Println("\nPhase 2: Penny embedding verification...")
	var (
		checked atomic.Int64
		valid   atomic.Int64
		mu      sync.Mutex
		results []Graph
	)

	jobs := make(chan Graph, 1000)
	var wg sync.WaitGroup

	for w := 0; w < *workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for g := range jobs {
				checked.Add(1)
				if g.isPennyGraph() {
					valid.Add(1)
					mu.Lock()
					results = append(results, g)
					mu.Unlock()
				}
			}
		}()
	}

	// Progress reporter
	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				c := checked.Load()
				v := valid.Load()
				pct := float64(c) * 100 / float64(len(candidates))
				rate := float64(c) / time.Since(start).Seconds()
				eta := time.Duration(float64(len(candidates)-int(c))/rate) * time.Second
				fmt.Printf("\r  Progress: %d/%d (%.1f%%) | Valid: %d | Rate: %.1f/s | ETA: %v   ",
					c, len(candidates), pct, v, rate, eta)
			}
		}
	}()

	// Feed jobs
	for _, g := range candidates {
		jobs <- g
	}
	close(jobs)

	wg.Wait()
	done <- true

	fmt.Printf("\n\nDone in %v\n", time.Since(start))
	fmt.Printf("Total checked: %d\n", checked.Load())
	fmt.Printf("Valid penny graphs: %d\n", len(results))

	// Write output
	if *outputFile != "" {
		if strings.HasSuffix(*outputFile, ".g6") {
			out, err := os.Create(*outputFile)
			if err != nil {
				fmt.Printf("Error creating %s: %v\n", *outputFile, err)
				os.Exit(1)
			}
			for _, g := range results {
				fmt.Fprintln(out, g.toGraph6())
			}
			out.Close()
		} else {
			out, err := os.Create(*outputFile)
			if err != nil {
				fmt.Printf("Error creating %s: %v\n", *outputFile, err)
				os.Exit(1)
			}
			writer := bufio.NewWriter(out)
			for _, g := range results {
				if bytesPerGraph == 4 {
					binary.Write(writer, binary.LittleEndian, uint32(g))
				} else {
					binary.Write(writer, binary.LittleEndian, uint64(g))
				}
			}
			writer.Flush()
			out.Close()
		}
		fmt.Printf("Wrote %d penny graphs to %s\n", len(results), *outputFile)
	}
}
