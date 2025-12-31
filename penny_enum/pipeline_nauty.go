package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	return g&(1<<edgeIndex[i][j]) != 0
}

func (g Graph) degree(v int) int {
	deg := 0
	for u := 0; u < n; u++ {
		if u != v && g.hasEdge(v, u) {
			deg++
		}
	}
	return deg
}

func (g Graph) isConnected() bool {
	if g == 0 {
		return false
	}
	visited := make([]bool, n)
	queue := []int{0}
	visited[0] = true
	count := 1
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		for u := 0; u < n; u++ {
			if !visited[u] && g.hasEdge(node, u) {
				visited[u] = true
				count++
				queue = append(queue, u)
			}
		}
	}
	return count == n
}

func (g Graph) hasIsolatedVertex() bool {
	for v := 0; v < n; v++ {
		if g.degree(v) == 0 {
			return true
		}
	}
	return false
}

func (g Graph) maxDegree() int {
	maxDeg := 0
	for v := 0; v < n; v++ {
		d := g.degree(v)
		if d > maxDeg {
			maxDeg = d
		}
	}
	return maxDeg
}

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

func (g Graph) edgeCount() int {
	count := 0
	tmp := g
	for tmp != 0 {
		count += int(tmp & 1)
		tmp >>= 1
	}
	return count
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
	nFlag := flag.Int("n", 9, "number of vertices")
	minEdges := flag.Int("min", 0, "minimum edges (default: n-1)")
	maxEdgesFlag := flag.Int("max", 0, "maximum edges (default: 3n-6 for planar)")
	batchSize := flag.Int("batch", 10000000, "graphs per batch")
	outputFile := flag.String("out", "", "output file for unique graphs")
	tmpDir := flag.String("tmp", "tmp_nauty", "temp directory for intermediate files")
	workers := flag.Int("workers", 0, "workers for candidate generation")
	flag.Parse()

	if *workers == 0 {
		*workers = runtime.NumCPU()
	}

	initEdges(*nFlag)

	minE := *minEdges
	if minE == 0 {
		minE = n - 1 // minimum for connected graph
	}
	maxE := *maxEdgesFlag
	if maxE == 0 {
		maxE = 3*n - 6 // planar graph bound
	}

	fmt.Printf("=== Pipeline for n=%d ===\n", n)
	fmt.Printf("Edge range: %d to %d\n", minE, maxE)
	fmt.Printf("Batch size: %d graphs\n", *batchSize)
	fmt.Printf("Workers: %d\n", *workers)

	os.MkdirAll(*tmpDir, 0755)

	start := time.Now()

	// Generate candidates and write in batches
	var (
		totalChecked  atomic.Int64
		totalWritten  atomic.Int64
		batchNum      atomic.Int32
		currentBatch  []string
		batchMu       sync.Mutex
		batchFiles    []string
		batchFilesMu  sync.Mutex
	)

	flushBatch := func(batch []string, num int) {
		if len(batch) == 0 {
			return
		}
		batchFile := filepath.Join(*tmpDir, fmt.Sprintf("batch_%04d.g6", num))
		f, _ := os.Create(batchFile)
		w := bufio.NewWriter(f)
		for _, line := range batch {
			fmt.Fprintln(w, line)
		}
		w.Flush()
		f.Close()

		// Run shortg on this batch
		uniqueFile := filepath.Join(*tmpDir, fmt.Sprintf("unique_%04d.g6", num))
		cmd := exec.Command("shortg", "-q", batchFile, uniqueFile)
		cmd.Run()

		// Count unique
		uf, _ := os.Open(uniqueFile)
		scanner := bufio.NewScanner(uf)
		count := 0
		for scanner.Scan() {
			count++
		}
		uf.Close()

		fmt.Printf("  Batch %d: %d -> %d unique\n", num, len(batch), count)

		// Remove batch file, keep unique file
		os.Remove(batchFile)

		batchFilesMu.Lock()
		batchFiles = append(batchFiles, uniqueFile)
		batchFilesMu.Unlock()
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
				c := totalChecked.Load()
				w := totalWritten.Load()
				rate := float64(c) / time.Since(start).Seconds()
				fmt.Printf("\r  Checked: %dM, candidates: %dM, rate: %.1fM/s   ",
					c/1000000, w/1000000, rate/1000000)
			}
		}
	}()

	// Generate all candidate graphs
	fmt.Println("\nPhase 1: Generating candidates...")

	// We'll iterate through all possible edge combinations
	// Use recursive generation with pruning
	var generate func(edgeIdx int, g Graph, edgeCount int)
	generate = func(edgeIdx int, g Graph, edgeCount int) {
		// Pruning: if we can't reach minE edges, skip
		remaining := numEdges - edgeIdx
		if edgeCount+remaining < minE {
			return
		}
		// If we have too many edges, skip
		if edgeCount > maxE {
			return
		}

		if edgeIdx == numEdges {
			totalChecked.Add(1)

			// Check candidate filters
			if edgeCount < minE || edgeCount > maxE {
				return
			}
			if g.hasIsolatedVertex() {
				return
			}
			if g.maxDegree() > 6 {
				return
			}
			if !g.isConnected() {
				return
			}
			if g.hasK4() {
				return
			}

			// Valid candidate
			g6 := g.toGraph6()
			totalWritten.Add(1)

			batchMu.Lock()
			currentBatch = append(currentBatch, g6)
			if len(currentBatch) >= *batchSize {
				batch := currentBatch
				num := int(batchNum.Add(1))
				currentBatch = nil
				batchMu.Unlock()
				flushBatch(batch, num)
			} else {
				batchMu.Unlock()
			}
			return
		}

		// Don't include this edge
		generate(edgeIdx+1, g, edgeCount)

		// Include this edge
		generate(edgeIdx+1, g|(1<<edgeIdx), edgeCount+1)
	}

	generate(0, 0, 0)

	// Flush remaining batch
	batchMu.Lock()
	if len(currentBatch) > 0 {
		batch := currentBatch
		num := int(batchNum.Add(1))
		currentBatch = nil
		batchMu.Unlock()
		flushBatch(batch, num)
	} else {
		batchMu.Unlock()
	}

	done <- true

	fmt.Printf("\n\nPhase 1 complete: %d candidates in %d batches\n",
		totalWritten.Load(), len(batchFiles))

	// Phase 2: Merge all unique files and run shortg again
	if len(batchFiles) > 1 {
		fmt.Println("\nPhase 2: Merging batches...")

		// Concatenate all unique files
		mergedFile := filepath.Join(*tmpDir, "merged.g6")
		mf, _ := os.Create(mergedFile)
		mw := bufio.NewWriter(mf)
		totalMerged := 0
		for _, uf := range batchFiles {
			f, _ := os.Open(uf)
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				fmt.Fprintln(mw, scanner.Text())
				totalMerged++
			}
			f.Close()
		}
		mw.Flush()
		mf.Close()

		fmt.Printf("  Merged %d graphs from %d batch files\n", totalMerged, len(batchFiles))

		// Final shortg
		finalFile := *outputFile
		if finalFile == "" {
			finalFile = fmt.Sprintf("n%d_unique.g6", n)
		}
		fmt.Println("  Running final shortg...")
		cmd := exec.Command("shortg", "-q", mergedFile, finalFile)
		cmd.Run()

		// Count final
		f, _ := os.Open(finalFile)
		scanner := bufio.NewScanner(f)
		finalCount := 0
		for scanner.Scan() {
			finalCount++
		}
		f.Close()

		fmt.Printf("\n=== Result ===\n")
		fmt.Printf("Total unique graphs: %d\n", finalCount)
		fmt.Printf("Output: %s\n", finalFile)
		fmt.Printf("Time: %v\n", time.Since(start))

		// Cleanup
		for _, uf := range batchFiles {
			os.Remove(uf)
		}
		os.Remove(mergedFile)

	} else if len(batchFiles) == 1 {
		// Just one batch, rename it
		finalFile := *outputFile
		if finalFile == "" {
			finalFile = fmt.Sprintf("n%d_unique.g6", n)
		}
		os.Rename(batchFiles[0], finalFile)

		f, _ := os.Open(finalFile)
		scanner := bufio.NewScanner(f)
		count := 0
		for scanner.Scan() {
			count++
		}
		f.Close()

		fmt.Printf("\n=== Result ===\n")
		fmt.Printf("Total unique graphs: %d\n", count)
		fmt.Printf("Output: %s\n", finalFile)
		fmt.Printf("Time: %v\n", time.Since(start))
	}

	os.Remove(*tmpDir)
}
