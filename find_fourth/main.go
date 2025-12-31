package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/crillab/gophersat/solver"
)

type candidate struct {
	index int
	line  string
}

type result struct {
	index          int
	found          bool
	uncoveredCount int
	elapsed        time.Duration
	arr1, arr2     []int
	arr3           []int
}

func main() {
	nFlag := flag.Int("n", 17, "Number of items")
	inDir := flag.String("in", "output_17", "Input directory")
	samples := flag.Int("samples", 0, "Number of samples to check (0 = all)")
	workers := flag.Int("workers", 0, "Number of workers (0 = NumCPU)")
	flag.Parse()

	n := *nFlag
	numPairs := n * (n - 1) / 2
	numWorkers := *workers
	if numWorkers == 0 {
		numWorkers = runtime.NumCPU()
	}

	edges, numEdges := buildSpiral(n)
	fmt.Printf("n=%d, edges=%d, pairs=%d\n", n, numEdges, numPairs)
	fmt.Printf("Using %d workers\n", numWorkers)

	// Build pair index lookup
	pairTable := make([][]int, n)
	for a := 0; a < n; a++ {
		pairTable[a] = make([]int, n)
		for b := 0; b < n; b++ {
			if a < b {
				pairTable[a][b] = a*n - a*(a+1)/2 + (b - a - 1)
			} else if b < a {
				pairTable[a][b] = b*n - b*(b+1)/2 + (a - b - 1)
			}
		}
	}

	// Full adjacency
	fullAdj := make([][]int, n)
	for s := 0; s < n; s++ {
		fullAdj[s] = []int{}
	}
	for _, e := range edges {
		fullAdj[e.a] = append(fullAdj[e.a], e.b)
		fullAdj[e.b] = append(fullAdj[e.b], e.a)
	}

	// Adjacency matrix
	adjMatrix := make([][]bool, n)
	for s := 0; s < n; s++ {
		adjMatrix[s] = make([]bool, n)
	}
	for _, e := range edges {
		adjMatrix[e.a][e.b] = true
		adjMatrix[e.b][e.a] = true
	}

	// arr0 = identity coverage
	covered0 := make([]bool, numPairs)
	for _, e := range edges {
		covered0[pairTable[e.a][e.b]] = true
	}

	// Load lines from input files
	var allLines []string
	files, _ := filepath.Glob(filepath.Join(*inDir, "item_*.txt"))
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			allLines = append(allLines, scanner.Text())
		}
		f.Close()
	}

	fmt.Printf("Loaded %d candidates\n", len(allLines))

	checkCount := *samples
	if checkCount == 0 || checkCount > len(allLines) {
		checkCount = len(allLines)
	}

	fmt.Printf("Checking %d candidates with SAT solver...\n\n", checkCount)

	work := make(chan candidate, 1000)
	results := make(chan result, 100)

	var stopFlag int32

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for cand := range work {
				if atomic.LoadInt32(&stopFlag) != 0 {
					continue
				}

				parts := strings.Split(cand.line, ";")
				if len(parts) != 2 {
					continue
				}

				arr1 := parseArray(parts[0])
				arr2 := parseArray(parts[1])
				if len(arr1) != n || len(arr2) != n {
					continue
				}

				// Compute covered pairs after arr0, arr1, arr2
				covered := make([]bool, numPairs)
				copy(covered, covered0)

				for slot := 0; slot < n; slot++ {
					item := arr1[slot]
					for _, adjSlot := range fullAdj[slot] {
						adjItem := arr1[adjSlot]
						covered[pairTable[item][adjItem]] = true
					}
				}

				for slot := 0; slot < n; slot++ {
					item := arr2[slot]
					for _, adjSlot := range fullAdj[slot] {
						adjItem := arr2[adjSlot]
						covered[pairTable[item][adjItem]] = true
					}
				}

				// Find uncovered pairs
				var uncoveredPairs [][2]int
				for a := 0; a < n; a++ {
					for b := a + 1; b < n; b++ {
						if !covered[pairTable[a][b]] {
							uncoveredPairs = append(uncoveredPairs, [2]int{a, b})
						}
					}
				}

				start := time.Now()
				found, arr3 := solveSAT(n, uncoveredPairs, adjMatrix)
				elapsed := time.Since(start)

				results <- result{
					index:          cand.index,
					found:          found,
					uncoveredCount: len(uncoveredPairs),
					elapsed:        elapsed,
					arr1:           arr1,
					arr2:           arr2,
					arr3:           arr3,
				}

				if found {
					atomic.StoreInt32(&stopFlag, 1)
				}
			}
		}()
	}

	var checkedCount int64
	var foundResult *result
	start := time.Now()

	// Progress ticker - update every second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	done := make(chan struct{})
	go func() {
		for {
			select {
			case res, ok := <-results:
				if !ok {
					close(done)
					return
				}
				atomic.AddInt64(&checkedCount, 1)

				if res.found {
					foundResult = &res
					fmt.Printf("\n*** SOLUTION FOUND at candidate %d! ***\n", res.index)
					fmt.Printf("arr0: identity [0,1,2,...,%d]\n", n-1)
					fmt.Printf("arr1: %v\n", res.arr1)
					fmt.Printf("arr2: %v\n", res.arr2)
					fmt.Printf("arr3: %v\n", res.arr3)
					fmt.Printf("Uncovered pairs before arr3: %d\n", res.uncoveredCount)
					fmt.Printf("SAT solve time: %v\n", res.elapsed)
					fmt.Printf("Total time to find: %v\n", time.Since(start).Round(time.Millisecond))
				}

			case <-ticker.C:
				count := atomic.LoadInt64(&checkedCount)
				if count > 0 {
					elapsed := time.Since(start)
					rate := float64(count) / elapsed.Seconds()
					remaining := float64(checkCount) - float64(count)
					eta := time.Duration(remaining/rate) * time.Second
					fmt.Printf("  Progress: %d/%d (%.2f%%), rate=%.1f/s, ETA=%v\n",
						count, checkCount, float64(count)/float64(checkCount)*100, rate, eta.Round(time.Second))
				}
			}
		}
	}()

	for i := 0; i < checkCount; i++ {
		if atomic.LoadInt32(&stopFlag) != 0 {
			break
		}
		work <- candidate{index: i, line: allLines[i]}
	}
	close(work)

	wg.Wait()
	close(results)
	<-done

	elapsed := time.Since(start)
	checked := atomic.LoadInt64(&checkedCount)

	fmt.Printf("\nResults:\n")
	fmt.Printf("  Checked: %d\n", checked)
	fmt.Printf("  Total time: %v\n", elapsed.Round(time.Millisecond))
	if checked > 0 {
		fmt.Printf("  Avg time per candidate: %v\n", elapsed/time.Duration(checked))
		fmt.Printf("  Rate: %.0f candidates/sec\n", float64(checked)/elapsed.Seconds())
	}

	if foundResult != nil {
		fmt.Printf("\n*** Solution exists! 4 arrangements cover all %d pairs ***\n", numPairs)
	} else {
		fmt.Printf("\n*** No solution found in %d candidates ***\n", checked)
	}
}

func solveSAT(n int, uncoveredPairs [][2]int, adjMatrix [][]bool) (bool, []int) {
	// Variables: x[item][slot] means item is placed in slot
	// Variable numbering: item*n + slot + 1 (SAT vars are 1-indexed)
	varIdx := func(item, slot int) int {
		return item*n + slot + 1
	}

	var clauses [][]int

	// Constraint 1: Each item in at least one slot
	for item := 0; item < n; item++ {
		clause := make([]int, n)
		for slot := 0; slot < n; slot++ {
			clause[slot] = varIdx(item, slot)
		}
		clauses = append(clauses, clause)
	}

	// Constraint 2: Each item in at most one slot
	for item := 0; item < n; item++ {
		for s1 := 0; s1 < n; s1++ {
			for s2 := s1 + 1; s2 < n; s2++ {
				clauses = append(clauses, []int{-varIdx(item, s1), -varIdx(item, s2)})
			}
		}
	}

	// Constraint 3: Each slot has at least one item
	for slot := 0; slot < n; slot++ {
		clause := make([]int, n)
		for item := 0; item < n; item++ {
			clause[item] = varIdx(item, slot)
		}
		clauses = append(clauses, clause)
	}

	// Constraint 4: Each slot has at most one item
	for slot := 0; slot < n; slot++ {
		for i1 := 0; i1 < n; i1++ {
			for i2 := i1 + 1; i2 < n; i2++ {
				clauses = append(clauses, []int{-varIdx(i1, slot), -varIdx(i2, slot)})
			}
		}
	}

	// Next available variable for auxiliaries
	nextVar := n*n + 1

	// Constraint 5: Each uncovered pair must be covered by arr3
	for _, pair := range uncoveredPairs {
		a, b := pair[0], pair[1]

		// Collect all ways this pair can be covered
		var auxVars []int
		for s1 := 0; s1 < n; s1++ {
			for s2 := 0; s2 < n; s2++ {
				if adjMatrix[s1][s2] {
					// aux <=> (a@s1 AND b@s2)
					aux := nextVar
					nextVar++
					auxVars = append(auxVars, aux)

					clauses = append(clauses, []int{-aux, varIdx(a, s1)})
					clauses = append(clauses, []int{-aux, varIdx(b, s2)})
					clauses = append(clauses, []int{-varIdx(a, s1), -varIdx(b, s2), aux})
				}
			}
		}

		// At least one aux must be true
		clauses = append(clauses, auxVars)
	}

	// Solve
	problem := solver.ParseSlice(clauses)
	s := solver.New(problem)
	status := s.Solve()

	if status != solver.Sat {
		return false, nil
	}

	// Extract solution
	model := s.Model()
	arr3 := make([]int, n)
	for item := 0; item < n; item++ {
		for slot := 0; slot < n; slot++ {
			v := varIdx(item, slot)
			if v <= len(model) && model[v-1] {
				arr3[slot] = item
				break
			}
		}
	}

	return true, arr3
}

func parseArray(s string) []int {
	parts := strings.Split(s, ",")
	result := make([]int, len(parts))
	for i, p := range parts {
		result[i], _ = strconv.Atoi(p)
	}
	return result
}
