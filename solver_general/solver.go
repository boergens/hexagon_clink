package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

var hexDirs = [6][2]float64{
	{1.5, 0}, {0.75, 1.3}, {-0.75, 1.3},
	{-1.5, 0}, {-0.75, -1.3}, {0.75, -1.3},
}

type Edge struct{ a, b int }

func buildSpiral(n int) []Edge {
	if n < 2 {
		return nil
	}

	positions := make([][2]float64, n)
	edges := make([]Edge, 0, n*3)
	positions[0] = [2]float64{0, 0}

	for node := 1; node < n; node++ {
		prev := positions[node-1]
		var bestPos [2]float64
		bestContacts, bestDist := -1, 1e9

		for d := 0; d < 6; d++ {
			cand := [2]float64{prev[0] + hexDirs[d][0], prev[1] + hexDirs[d][1]}

			occupied := false
			for i := 0; i < node; i++ {
				if math.Abs(cand[0]-positions[i][0]) < 0.1 && math.Abs(cand[1]-positions[i][1]) < 0.1 {
					occupied = true
					break
				}
			}
			if occupied {
				continue
			}

			contacts := 0
			for i := 0; i < node; i++ {
				for dd := 0; dd < 6; dd++ {
					neighbor := [2]float64{positions[i][0] + hexDirs[dd][0], positions[i][1] + hexDirs[dd][1]}
					if math.Abs(cand[0]-neighbor[0]) < 0.1 && math.Abs(cand[1]-neighbor[1]) < 0.1 {
						contacts++
						break
					}
				}
			}

			dist := cand[0]*cand[0] + cand[1]*cand[1]
			if contacts > bestContacts || (contacts == bestContacts && dist < bestDist) {
				bestPos, bestContacts, bestDist = cand, contacts, dist
			}
		}

		positions[node] = bestPos

		for i := 0; i < node; i++ {
			for d := 0; d < 6; d++ {
				neighbor := [2]float64{positions[i][0] + hexDirs[d][0], positions[i][1] + hexDirs[d][1]}
				if math.Abs(bestPos[0]-neighbor[0]) < 0.1 && math.Abs(bestPos[1]-neighbor[1]) < 0.1 {
					edges = append(edges, Edge{i, node})
					break
				}
			}
		}
	}
	return edges
}

type Solver struct {
	n, k      int
	numPairs  int
	numEdges  int
	edges     []Edge
	slotAdj   [][]int
	remEdges  []int
	pairTable [][]int

	solution [][]int
	found    int32
	mu       sync.Mutex
}

func NewSolver(n, k int) *Solver {
	edges := buildSpiral(n)

	slotAdj := make([][]int, n)
	for s := 0; s < n; s++ {
		for _, e := range edges {
			if e.a == s && e.b < s {
				slotAdj[s] = append(slotAdj[s], e.b)
			} else if e.b == s && e.a < s {
				slotAdj[s] = append(slotAdj[s], e.a)
			}
		}
	}

	remEdges := make([]int, n+1)
	for slot := 0; slot <= n; slot++ {
		for _, e := range edges {
			if e.a >= slot || e.b >= slot {
				remEdges[slot]++
			}
		}
	}

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

	return &Solver{
		n:         n,
		k:         k,
		numPairs:  n * (n - 1) / 2,
		numEdges:  len(edges),
		edges:     edges,
		slotAdj:   slotAdj,
		remEdges:  remEdges,
		pairTable: pairTable,
		solution:  make([][]int, k),
	}
}

func (s *Solver) pairIndex(a, b int) int {
	return s.pairTable[a][b]
}

func (s *Solver) solve(level int, covered []bool, coveredCount int, parentArrs [][]int, rng *rand.Rand) {
	if atomic.LoadInt32(&s.found) != 0 {
		return
	}

	remaining := s.k - level - 1
	missing := s.numPairs - coveredCount

	if missing > remaining*s.numEdges {
		return
	}

	minNewEdges := (missing + remaining - 1) / remaining
	maxOverlap := s.numEdges - minNewEdges

	arr := make([]int, s.n)
	used := make([]bool, s.n)
	usedItems := make([]int, 0, s.n)
	coveredSet := make([]bool, s.numPairs)
	copy(coveredSet, covered)

	order := make([]int, s.n)
	for i := 0; i < s.n; i++ {
		order[i] = i
	}
	rng.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })

	var enumerate func(slot, overlap, localCovered int)
	enumerate = func(slot, overlap, localCovered int) {
		if atomic.LoadInt32(&s.found) != 0 {
			return
		}

		missingNow := s.numPairs - localCovered
		maxPossible := s.remEdges[slot] + (remaining-1)*s.numEdges
		if missingNow > maxPossible {
			return
		}

		if slot == s.n {
			arrCopy := make([]int, s.n)
			copy(arrCopy, arr)
			coveredCopy := make([]bool, s.numPairs)
			copy(coveredCopy, coveredSet)

			newParentArrs := append(parentArrs, arrCopy)

			if level == s.k-2 {
				if localCovered == s.numPairs {
					s.mu.Lock()
					if atomic.LoadInt32(&s.found) == 0 {
						for i, perm := range newParentArrs {
							s.solution[i+1] = perm
						}
						atomic.StoreInt32(&s.found, 1)
					}
					s.mu.Unlock()
				}
			} else {
				s.solve(level+1, coveredCopy, localCovered, newParentArrs, rng)
			}
			return
		}

		for _, item := range order {
			if atomic.LoadInt32(&s.found) != 0 {
				return
			}
			if used[item] {
				continue
			}

			newOverlap := 0
			var newPairs []int
			for _, adjSlot := range s.slotAdj[slot] {
				adjItem := arr[adjSlot]
				pi := s.pairIndex(item, adjItem)
				if coveredSet[pi] {
					newOverlap++
				} else {
					newPairs = append(newPairs, pi)
				}
			}

			if overlap+newOverlap > maxOverlap {
				continue
			}

			if remaining == 1 {
				doomed := false
				for _, other := range usedItems {
					pi := s.pairIndex(item, other)
					if coveredSet[pi] {
						continue
					}
					found := false
					for _, cpi := range newPairs {
						if cpi == pi {
							found = true
							break
						}
					}
					if !found {
						doomed = true
						break
					}
				}
				if doomed {
					continue
				}
			}

			arr[slot] = item
			used[item] = true
			usedItems = append(usedItems, item)
			for _, pi := range newPairs {
				coveredSet[pi] = true
			}

			enumerate(slot+1, overlap+newOverlap, localCovered+len(newPairs))

			used[item] = false
			usedItems = usedItems[:len(usedItems)-1]
			for _, pi := range newPairs {
				coveredSet[pi] = false
			}
		}
	}

	enumerate(0, 0, coveredCount)
}

func (s *Solver) Solve(numWorkers int) bool {
	arr0 := make([]int, s.n)
	for i := 0; i < s.n; i++ {
		arr0[i] = i
	}
	s.solution[0] = arr0

	covered := make([]bool, s.numPairs)
	coveredCount := 0
	for _, e := range s.edges {
		pi := s.pairIndex(e.a, e.b)
		if !covered[pi] {
			covered[pi] = true
			coveredCount++
		}
	}

	if s.k == 1 {
		return coveredCount == s.numPairs
	}

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(seed int64) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(seed))
			s.solve(0, covered, coveredCount, nil, rng)
		}(time.Now().UnixNano() + int64(w)*12345)
	}
	wg.Wait()

	return atomic.LoadInt32(&s.found) != 0
}

func main() {
	n := flag.Int("n", 17, "Number of items")
	k := flag.Int("k", 4, "Number of arrangements")
	workers := flag.Int("workers", 8, "Number of parallel workers")
	flag.Parse()

	fmt.Printf("Searching for %d arrangements of %d items\n", *k, *n)

	solver := NewSolver(*n, *k)
	fmt.Printf("Edges per arrangement: %d, Total pairs: %d\n", solver.numEdges, solver.numPairs)
	fmt.Printf("Lower bound: ceil(%d/%d) = %d arrangements\n",
		solver.numPairs, solver.numEdges, (solver.numPairs+solver.numEdges-1)/solver.numEdges)
	fmt.Printf("Workers: %d\n\n", *workers)

	start := time.Now()
	found := solver.Solve(*workers)
	elapsed := time.Since(start)

	if found {
		fmt.Println("\n*** SOLUTION FOUND ***")
		for i, arr := range solver.solution {
			fmt.Printf("  Arr%d: %v\n", i, arr)
		}
	} else {
		fmt.Println("\nNo solution found.")
	}

	fmt.Printf("\nTime: %v\n", elapsed.Round(time.Millisecond))
}
