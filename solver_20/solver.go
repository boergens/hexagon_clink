package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	N = 20
	K = 5
)

var hexDirs = [6][2]float64{
	{1.5, 0}, {0.75, 1.3}, {-0.75, 1.3},
	{-1.5, 0}, {-0.75, -1.3}, {0.75, -1.3},
}

type Edge struct{ a, b int }

func buildSpiral() []Edge {
	positions := make([][2]float64, N)
	edges := make([]Edge, 0, N*3)
	positions[0] = [2]float64{0, 0}

	for node := 1; node < N; node++ {
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
	numPairs      int
	numEdges      int
	edges         []Edge
	slotAdj       [][]int // full adjacency for each slot
	slotDeg       []int   // degree of each slot
	pairTable     [][]int
	maxOverlapArr []int // per-level overlap limits

	solution     [][]int
	found        int32
	printedLevel []int32
	mu           sync.Mutex
}

func NewSolver() *Solver {
	edges := buildSpiral()

	// Build full adjacency for each slot
	slotAdj := make([][]int, N)
	for _, e := range edges {
		slotAdj[e.a] = append(slotAdj[e.a], e.b)
		slotAdj[e.b] = append(slotAdj[e.b], e.a)
	}

	// Compute degree for each slot
	slotDeg := make([]int, N)
	for i := 0; i < N; i++ {
		slotDeg[i] = len(slotAdj[i])
	}

	pairTable := make([][]int, N)
	for a := 0; a < N; a++ {
		pairTable[a] = make([]int, N)
		for b := 0; b < N; b++ {
			if a < b {
				pairTable[a][b] = a*N - a*(a+1)/2 + (b - a - 1)
			} else if b < a {
				pairTable[a][b] = b*N - b*(b+1)/2 + (a - b - 1)
			}
		}
	}

	return &Solver{
		numPairs:     N * (N - 1) / 2,
		numEdges:     len(edges),
		edges:        edges,
		slotAdj:      slotAdj,
		slotDeg:      slotDeg,
		pairTable:    pairTable,
		solution:     make([][]int, K),
		printedLevel: make([]int32, K),
	}
}

func (s *Solver) pairIndex(a, b int) int {
	return s.pairTable[a][b]
}

func (s *Solver) SetMaxOverlap(limits []int) {
	s.maxOverlapArr = limits
}

// countNeededPartners returns how many uncovered pairs item has with other items
func (s *Solver) countNeededPartners(item int, coveredSet []bool) int {
	count := 0
	for other := 0; other < N; other++ {
		if other == item {
			continue
		}
		pi := s.pairIndex(item, other)
		if !coveredSet[pi] {
			count++
		}
	}
	return count
}

// Special slot with minimum degree (slot 19 has degree 2)
const specialSlot = 19
const specialSlotDegree = 2

func (s *Solver) solve(level int, covered []bool, coveredCount int, parentArrs [][]int, rng *rand.Rand) {
	if atomic.LoadInt32(&s.found) != 0 {
		return
	}

	remaining := K - level - 1
	missing := s.numPairs - coveredCount

	if missing > remaining*s.numEdges {
		return
	}

	var maxOverlap int
	if s.maxOverlapArr != nil && level < len(s.maxOverlapArr) {
		maxOverlap = s.maxOverlapArr[level]
	} else {
		minNewEdges := (missing + remaining - 1) / remaining
		maxOverlap = s.numEdges - minNewEdges
	}

	arr := make([]int, N)
	for i := range arr {
		arr[i] = -1
	}
	used := make([]bool, N)
	filledSlots := make([]int, 0, N)
	coveredSet := make([]bool, s.numPairs)
	copy(coveredSet, covered)

	order := make([]int, N)
	for i := 0; i < N; i++ {
		order[i] = i
	}
	rng.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })

	// For the last arrangement (level == K-2), enumerate slots differently:
	// Start with slot 19 (the 2-edge slot) first, then the rest
	isLastLevel := (level == K-2)

	slotOrder := make([]int, N)
	if isLastLevel {
		slotOrder[0] = specialSlot
		idx := 1
		for i := 0; i < N; i++ {
			if i != specialSlot {
				slotOrder[idx] = i
				idx++
			}
		}
	} else {
		for i := 0; i < N; i++ {
			slotOrder[i] = i
		}
	}

	var enumerate func(depth, overlap, localCovered int)
	enumerate = func(depth, overlap, localCovered int) {
		if atomic.LoadInt32(&s.found) != 0 {
			return
		}

		if depth == N {
			arrCopy := make([]int, N)
			copy(arrCopy, arr)
			coveredCopy := make([]bool, s.numPairs)
			copy(coveredCopy, coveredSet)

			newParentArrs := append(parentArrs, arrCopy)

			count := atomic.AddInt32(&s.printedLevel[level], 1)
			if count <= 10 {
				newEdges := localCovered - coveredCount
				fmt.Printf("[%s] Valid arr%d #%d: %v (overlap=%d, new=%d, covered=%d/%d)\n",
					time.Now().Format("15:04:05.000"), level+1, count, arrCopy, s.numEdges-newEdges, newEdges, localCovered, s.numPairs)
			}

			if level == K-2 {
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

		slot := slotOrder[depth]

		// Determine which items to try for this slot
		var candidates []int
		if isLastLevel && depth == 0 {
			// First slot at last level is specialSlot - only try items needing ≤2 partners
			for _, item := range order {
				if used[item] {
					continue
				}
				needed := s.countNeededPartners(item, coveredSet)
				if needed <= specialSlotDegree {
					candidates = append(candidates, item)
				}
			}
		} else {
			for _, item := range order {
				if !used[item] {
					candidates = append(candidates, item)
				}
			}
		}

		for _, item := range candidates {
			if atomic.LoadInt32(&s.found) != 0 {
				return
			}

			// Calculate overlap and new pairs from edges to already-filled slots
			newOverlap := 0
			var newPairs []int
			for _, adjSlot := range s.slotAdj[slot] {
				if arr[adjSlot] == -1 {
					continue // adjacent slot not filled yet
				}
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

			// Doomed pair check for last arrangement
			if remaining == 1 {
				doomed := false
				for _, filledSlot := range filledSlots {
					other := arr[filledSlot]
					pi := s.pairIndex(item, other)
					if coveredSet[pi] {
						continue
					}
					// Check if this pair can still be covered
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
			filledSlots = append(filledSlots, slot)
			for _, pi := range newPairs {
				coveredSet[pi] = true
			}

			enumerate(depth+1, overlap+newOverlap, localCovered+len(newPairs))

			arr[slot] = -1
			used[item] = false
			filledSlots = filledSlots[:len(filledSlots)-1]
			for _, pi := range newPairs {
				coveredSet[pi] = false
			}
		}
	}

	enumerate(0, 0, coveredCount)
}

func (s *Solver) Solve(numWorkers int) bool {
	arr0 := make([]int, N)
	for i := 0; i < N; i++ {
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

func parseOverlapLimits(s string) ([]int, error) {
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	limits := make([]int, len(parts))
	for i, p := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return nil, fmt.Errorf("invalid overlap limit %q: %v", p, err)
		}
		limits[i] = v
	}
	return limits, nil
}

func main() {
	workers := flag.Int("workers", 8, "Number of parallel workers")
	maxOverlap := flag.String("max-overlap", "", "Comma-separated max overlap per level (e.g., '0,0,10,10')")
	flag.Parse()

	fmt.Printf("Searching for %d arrangements of %d items\n", K, N)

	solver := NewSolver()

	overlapLimits, err := parseOverlapLimits(*maxOverlap)
	if err != nil {
		fmt.Printf("Error parsing max-overlap: %v\n", err)
		return
	}
	if overlapLimits != nil {
		solver.SetMaxOverlap(overlapLimits)
		fmt.Printf("Max overlap limits: %v\n", overlapLimits)
	}

	fmt.Printf("Edges per arrangement: %d, Total pairs: %d\n", solver.numEdges, solver.numPairs)
	fmt.Printf("Lower bound: ceil(%d/%d) = %d arrangements\n",
		solver.numPairs, solver.numEdges, (solver.numPairs+solver.numEdges-1)/solver.numEdges)
	fmt.Printf("Special slot %d has degree %d (filled first at last level)\n", specialSlot, specialSlotDegree)
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
