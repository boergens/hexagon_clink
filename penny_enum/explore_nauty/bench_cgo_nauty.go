package main

/*
#cgo CFLAGS: -I/opt/homebrew/include
#cgo LDFLAGS: -L/opt/homebrew/lib -lnauty

#include <nauty.h>
#include <naututil.h>

// Wrapper to canonicalize a graph
// adj is the adjacency matrix as a flat array (row-major)
// n is the number of vertices
// Returns the canonical labeling hash
unsigned long canonical_hash(int *adj, int n) {
    DYNALLSTAT(int, lab, lab_sz);
    DYNALLSTAT(int, ptn, ptn_sz);
    DYNALLSTAT(int, orbits, orbits_sz);
    DYNALLSTAT(graph, g, g_sz);
    DYNALLSTAT(graph, cg, cg_sz);

    static DEFAULTOPTIONS_GRAPH(options);
    statsblk stats;

    int m = SETWORDSNEEDED(n);
    nauty_check(WORDSIZE, m, n, NAUTYVERSIONID);

    DYNALLOC1(int, lab, lab_sz, n, "malloc");
    DYNALLOC1(int, ptn, ptn_sz, n, "malloc");
    DYNALLOC1(int, orbits, orbits_sz, n, "malloc");
    DYNALLOC2(graph, g, g_sz, n, m, "malloc");
    DYNALLOC2(graph, cg, cg_sz, n, m, "malloc");

    EMPTYGRAPH(g, m, n);

    // Build graph from adjacency matrix
    for (int i = 0; i < n; i++) {
        for (int j = i + 1; j < n; j++) {
            if (adj[i * n + j]) {
                ADDONEEDGE(g, i, j, m);
            }
        }
    }

    options.getcanon = TRUE;
    options.defaultptn = TRUE;

    densenauty(g, lab, ptn, orbits, &options, &stats, m, n, cg);

    // Hash the canonical graph
    unsigned long hash = 0;
    for (int i = 0; i < n * m; i++) {
        hash = hash * 31 + cg[i];
    }

    DYNFREE(lab, lab_sz);
    DYNFREE(ptn, ptn_sz);
    DYNFREE(orbits, orbits_sz);
    DYNFREE(g, g_sz);
    DYNFREE(cg, cg_sz);

    return hash;
}
*/
import "C"

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"time"
	"unsafe"
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

func (g Graph) toAdjMatrix() []C.int {
	adj := make([]C.int, n*n)
	for idx := 0; idx < numEdges; idx++ {
		if g&(1<<idx) != 0 {
			i, j := edgePairs[idx][0], edgePairs[idx][1]
			adj[i*n+j] = 1
			adj[j*n+i] = 1
		}
	}
	return adj
}

func (g Graph) canonicalHash() uint64 {
	adj := g.toAdjMatrix()
	hash := C.canonical_hash((*C.int)(unsafe.Pointer(&adj[0])), C.int(n))
	return uint64(hash)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: bench_cgo_nauty <input.bin> <n>")
		fmt.Println("  Benchmarks nauty via CGO on binary graph file")
		fmt.Println("")
		fmt.Println("Requires nauty library: brew install nauty")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	vertices, _ := strconv.Atoi(os.Args[2])
	initEdges(vertices)

	bytesPerGraph := 4
	if numEdges > 32 {
		bytesPerGraph = 8
	}

	// Read graphs
	f, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
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

	fmt.Printf("Read %d graphs (n=%d)\n", len(graphs), n)

	fmt.Println("\n=== nauty via CGO ===")
	unique := make(map[uint64]bool)
	start := time.Now()

	for i, g := range graphs {
		hash := g.canonicalHash()
		unique[hash] = true

		if (i+1)%50000 == 0 {
			elapsed := time.Since(start)
			fmt.Printf("  %d/%d graphs (%.0f/sec)\n", i+1, len(graphs), float64(i+1)/elapsed.Seconds())
		}
	}

	elapsed := time.Since(start)
	fmt.Printf("\nTime: %v\n", elapsed)
	fmt.Printf("Graphs/sec: %.0f\n", float64(len(graphs))/elapsed.Seconds())
	fmt.Printf("Unique canonical forms: %d\n", len(unique))
}
