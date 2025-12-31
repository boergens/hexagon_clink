package main

import (
	"flag"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// All 4 maximal penny graphs on 13 vertices (26 edges each)
var allGraphs = [][][2]int{
	// Graph A
	{
		{2, 3}, {1, 4}, {0, 6}, {5, 6}, {3, 7}, {5, 7},
		{2, 8}, {4, 8}, {0, 9}, {1, 9}, {6, 9}, {1, 10},
		{4, 10}, {8, 10}, {9, 10}, {5, 11}, {6, 11}, {7, 11},
		{9, 11}, {10, 11}, {2, 12}, {3, 12}, {7, 12}, {8, 12},
		{10, 12}, {11, 12},
	},
	// Graph B
	{
		{1, 3}, {2, 4}, {0, 5}, {0, 6}, {5, 6}, {3, 7},
		{6, 7}, {4, 8}, {5, 8}, {1, 9}, {2, 9}, {5, 10},
		{6, 10}, {7, 10}, {8, 10}, {1, 11}, {3, 11}, {7, 11},
		{9, 11}, {10, 11}, {2, 12}, {4, 12}, {8, 12}, {9, 12},
		{10, 12}, {11, 12},
	},
	// Graph C
	{
		{0, 3}, {1, 4}, {3, 5}, {4, 6}, {2, 7}, {5, 7},
		{2, 8}, {6, 8}, {0, 9}, {1, 9}, {0, 10}, {3, 10},
		{5, 10}, {7, 10}, {9, 10}, {1, 11}, {4, 11}, {6, 11},
		{8, 11}, {9, 11}, {2, 12}, {7, 12}, {8, 12}, {9, 12},
		{10, 12}, {11, 12},
	},
	// Graph D
	{
		{0, 2}, {1, 3}, {1, 4}, {0, 5}, {2, 6}, {4, 7},
		{3, 8}, {6, 8}, {5, 9}, {7, 9}, {0, 10}, {2, 10},
		{5, 10}, {6, 10}, {9, 10}, {1, 11}, {3, 11}, {4, 11},
		{7, 11}, {8, 11}, {6, 12}, {7, 12}, {8, 12}, {9, 12},
		{10, 12}, {11, 12},
	},
}

const numItems = 13
const numEdges = 26

var allNeighbors [4][][]int
var allPairs [][2]int

func init() {
	// Precompute neighbor lists
	for i, g := range allGraphs {
		neighbors := make([][]int, numItems)
		for j := range neighbors {
			neighbors[j] = []int{}
		}
		for _, e := range g {
			neighbors[e[0]] = append(neighbors[e[0]], e[1])
			neighbors[e[1]] = append(neighbors[e[1]], e[0])
		}
		allNeighbors[i] = neighbors
	}

	// All pairs
	for i := 0; i < numItems-1; i++ {
		for j := i + 1; j < numItems; j++ {
			allPairs = append(allPairs, [2]int{i, j})
		}
	}
}

func buildPairsTable(shapeIdx int, arr []int) [numItems][numItems]bool {
	var table [numItems][numItems]bool
	for _, e := range allGraphs[shapeIdx] {
		item1 := arr[e[0]]
		item2 := arr[e[1]]
		table[item1][item2] = true
		table[item2][item1] = true
	}
	return table
}

type Solution struct {
	shape0, shape1, shape2 int
	arr1, arr2             [numItems]int
}

// Search for arr2 that covers exactly the needed pairs
func searchArr2(shape2 int, neededTable *[numItems][numItems]bool, neededCount int, found *atomic.Bool) (bool, [numItems]int) {
	neighbors2 := allNeighbors[shape2]
	var arr2 [numItems]int
	var used2 [numItems]bool
	pairsCovered := 0
	var result [numItems]int
	success := false

	var search func(pos int)
	search = func(pos int) {
		if success || found.Load() {
			return
		}

		if pos == numItems {
			if pairsCovered == neededCount {
				success = true
				result = arr2
			}
			return
		}

		for item := 0; item < numItems; item++ {
			if used2[item] {
				continue
			}

			arr2[pos] = item
			used2[item] = true

			isWaste := false
			newPairs := 0

			for _, nPos := range neighbors2[pos] {
				if nPos < pos {
					nItem := arr2[nPos]
					if neededTable[item][nItem] {
						newPairs++
					} else {
						isWaste = true
						break
					}
				}
			}

			if !isWaste {
				pairsCovered += newPairs
				search(pos + 1)
				pairsCovered -= newPairs
			}

			arr2[pos] = 0
			used2[item] = false
		}
	}

	search(0)
	return success, result
}

// Search for arr1 starting with firstItem at position 0
func searchArr1Worker(shape0, shape1, firstItem int, pairs0Table *[numItems][numItems]bool,
	found *atomic.Bool, resultChan chan<- Solution, countChan chan<- int64) {

	neighbors1 := allNeighbors[shape1]
	var arr1 [numItems]int
	var used1 [numItems]bool
	var localCount int64

	arr1[0] = firstItem
	used1[firstItem] = true

	var search func(pos int)
	search = func(pos int) {
		if found.Load() {
			return
		}

		if pos == numItems {
			localCount++
			// Complete arr1 found, compute needed pairs and search arr2
			pairs1Table := buildPairsTable(shape1, arr1[:])

			var neededTable [numItems][numItems]bool
			neededCount := 0
			for _, p := range allPairs {
				if !pairs0Table[p[0]][p[1]] && !pairs1Table[p[0]][p[1]] {
					neededTable[p[0]][p[1]] = true
					neededTable[p[1]][p[0]] = true
					neededCount++
				}
			}

			// Try each shape2 >= shape1
			for shape2 := shape1; shape2 < 4 && !found.Load(); shape2++ {
				success, arr2 := searchArr2(shape2, &neededTable, neededCount, found)
				if success && found.CompareAndSwap(false, true) {
					resultChan <- Solution{shape0, shape1, shape2, arr1, arr2}
					return
				}
			}
			return
		}

		for item := 0; item < numItems; item++ {
			if used1[item] {
				continue
			}

			arr1[pos] = item
			used1[item] = true

			hasOverlap := false
			for _, nPos := range neighbors1[pos] {
				if nPos < pos {
					nItem := arr1[nPos]
					if pairs0Table[item][nItem] {
						hasOverlap = true
						break
					}
				}
			}

			if !hasOverlap {
				search(pos + 1)
			}

			arr1[pos] = 0
			used1[item] = false
		}
	}

	search(1) // Start from position 1 since position 0 is fixed
	countChan <- localCount
}

func main() {
	workers := flag.Int("w", 13, "number of workers per shape pair")
	flag.Parse()

	start := time.Now()

	fmt.Println("============================================")
	fmt.Println("SOLVER: n=13, testing if 3 arrangements suffice")
	fmt.Println("============================================")
	fmt.Printf("Workers: %d\n\n", *workers)

	var identity [numItems]int
	for i := 0; i < numItems; i++ {
		identity[i] = i
	}

	found := &atomic.Bool{}
	resultChan := make(chan Solution, 1)

	// shape0 <= shape1 <= shape2 (symmetry breaking)
	for shape0 := 0; shape0 < 4 && !found.Load(); shape0++ {
		pairs0Table := buildPairsTable(shape0, identity[:])

		for shape1 := shape0; shape1 < 4 && !found.Load(); shape1++ {
			label := string(rune('A'+shape0)) + string(rune('A'+shape1)) + "*"
			fmt.Printf("Testing %s: ", label)

			var wg sync.WaitGroup
			countChan := make(chan int64, numItems)

			// Launch workers for each first digit
			numWorkers := *workers
			if numWorkers > numItems {
				numWorkers = numItems
			}

			for firstItem := 0; firstItem < numItems; firstItem++ {
				wg.Add(1)
				go func(fi int) {
					defer wg.Done()
					searchArr1Worker(shape0, shape1, fi, &pairs0Table, found, resultChan, countChan)
				}(firstItem)
			}

			// Wait for all workers
			wg.Wait()
			close(countChan)

			var totalArr1 int64
			for c := range countChan {
				totalArr1 += c
			}

			fmt.Printf("%d arr1 checked (elapsed: %v)\n", totalArr1, time.Since(start))

			if found.Load() {
				break
			}
		}
	}

	fmt.Println()
	fmt.Println("============================================")
	fmt.Println("RESULT")
	fmt.Println("============================================")
	fmt.Println()

	if found.Load() {
		sol := <-resultChan
		fmt.Println("*** FOUND A SOLUTION! ***")
		fmt.Printf("Shapes: %c%c%c\n", 'A'+sol.shape0, 'A'+sol.shape1, 'A'+sol.shape2)
		fmt.Printf("arr0 = %v\n", identity)
		fmt.Printf("arr1 = %v\n", sol.arr1)
		fmt.Printf("arr2 = %v\n", sol.arr2)
	} else {
		fmt.Println("No solution found.")
		fmt.Println("3 arrangements are NOT sufficient for n=13.")
		fmt.Println("CONCLUSION: n=13 requires at least 4 arrangements.")
	}

	fmt.Printf("\nTotal time: %v\n", time.Since(start))
}
