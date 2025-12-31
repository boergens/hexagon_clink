package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"sort"
	"sync"
)

// Vertex in triangular lattice (a, b) coordinates
type Vertex struct {
	A, B int
}

// Triangle represented by 3 vertices (sorted for canonical form)
type Triangle [3]Vertex

// Polyiamond is a set of triangles
type Polyiamond struct {
	Triangles []Triangle
}

func makeTriangle(v1, v2, v3 Vertex) Triangle {
	verts := []Vertex{v1, v2, v3}
	sort.Slice(verts, func(i, j int) bool {
		if verts[i].A != verts[j].A {
			return verts[i].A < verts[j].A
		}
		return verts[i].B < verts[j].B
	})
	return Triangle{verts[0], verts[1], verts[2]}
}

func rotateVertex60(v Vertex) Vertex {
	return Vertex{-v.B, v.A + v.B}
}

func reflectVertex(v Vertex) Vertex {
	return Vertex{v.A + v.B, -v.B}
}

func transformTriangle(t Triangle, rotate int, reflect bool) Triangle {
	verts := []Vertex{t[0], t[1], t[2]}

	if reflect {
		for i := range verts {
			verts[i] = reflectVertex(verts[i])
		}
	}

	for r := 0; r < rotate%6; r++ {
		for i := range verts {
			verts[i] = rotateVertex60(verts[i])
		}
	}

	return makeTriangle(verts[0], verts[1], verts[2])
}

func normalizePolyiamond(p Polyiamond) Polyiamond {
	if len(p.Triangles) == 0 {
		return p
	}

	// Find min a, b across all vertices
	minA, minB := 1000000, 1000000
	for _, t := range p.Triangles {
		for _, v := range t {
			if v.A < minA {
				minA = v.A
			}
			if v.B < minB {
				minB = v.B
			}
		}
	}

	// Translate
	result := Polyiamond{Triangles: make([]Triangle, len(p.Triangles))}
	for i, t := range p.Triangles {
		result.Triangles[i] = makeTriangle(
			Vertex{t[0].A - minA, t[0].B - minB},
			Vertex{t[1].A - minA, t[1].B - minB},
			Vertex{t[2].A - minA, t[2].B - minB},
		)
	}

	// Sort triangles for canonical ordering
	sort.Slice(result.Triangles, func(i, j int) bool {
		for k := 0; k < 3; k++ {
			if result.Triangles[i][k].A != result.Triangles[j][k].A {
				return result.Triangles[i][k].A < result.Triangles[j][k].A
			}
			if result.Triangles[i][k].B != result.Triangles[j][k].B {
				return result.Triangles[i][k].B < result.Triangles[j][k].B
			}
		}
		return false
	})

	return result
}

func transformPolyiamond(p Polyiamond, rotate int, reflect bool) Polyiamond {
	result := Polyiamond{Triangles: make([]Triangle, len(p.Triangles))}
	for i, t := range p.Triangles {
		result.Triangles[i] = transformTriangle(t, rotate, reflect)
	}
	return normalizePolyiamond(result)
}

func polyiamondKey(p Polyiamond) string {
	var key string
	for _, t := range p.Triangles {
		key += fmt.Sprintf("%d,%d,%d,%d,%d,%d;", t[0].A, t[0].B, t[1].A, t[1].B, t[2].A, t[2].B)
	}
	return key
}

func comparePolyiamonds(a, b Polyiamond) int {
	for i := 0; i < len(a.Triangles) && i < len(b.Triangles); i++ {
		for k := 0; k < 3; k++ {
			if a.Triangles[i][k].A != b.Triangles[i][k].A {
				return a.Triangles[i][k].A - b.Triangles[i][k].A
			}
			if a.Triangles[i][k].B != b.Triangles[i][k].B {
				return a.Triangles[i][k].B - b.Triangles[i][k].B
			}
		}
	}
	return len(a.Triangles) - len(b.Triangles)
}

func canonicalize(p Polyiamond) Polyiamond {
	best := normalizePolyiamond(p)

	for rot := 0; rot < 6; rot++ {
		for _, refl := range []bool{false, true} {
			candidate := transformPolyiamond(p, rot, refl)
			if comparePolyiamonds(candidate, best) < 0 {
				best = candidate
			}
		}
	}

	return best
}

func getAdjacentTriangles(t Triangle) []Triangle {
	verts := []Vertex{t[0], t[1], t[2]}
	neighbors := make([]Triangle, 0, 3)

	for i := 0; i < 3; i++ {
		v1, v2 := verts[i], verts[(i+1)%3]
		v3 := verts[(i+2)%3]

		// Fourth vertex completing parallelogram
		v4 := Vertex{v1.A + v2.A - v3.A, v1.B + v2.B - v3.B}
		neighbors = append(neighbors, makeTriangle(v1, v2, v4))
	}

	return neighbors
}

func polyiamondContains(p Polyiamond, t Triangle) bool {
	for _, tri := range p.Triangles {
		if tri == t {
			return true
		}
	}
	return false
}

func getBoundary(p Polyiamond) []Triangle {
	seen := make(map[Triangle]bool)
	for _, t := range p.Triangles {
		for _, neighbor := range getAdjacentTriangles(t) {
			if !polyiamondContains(p, neighbor) {
				seen[neighbor] = true
			}
		}
	}

	result := make([]Triangle, 0, len(seen))
	for t := range seen {
		result = append(result, t)
	}
	return result
}

func addTriangle(p Polyiamond, t Triangle) Polyiamond {
	newTris := make([]Triangle, len(p.Triangles)+1)
	copy(newTris, p.Triangles)
	newTris[len(p.Triangles)] = t
	return Polyiamond{Triangles: newTris}
}

func enumeratePolyiamonds(n int, workers int) []Polyiamond {
	if n < 1 {
		return nil
	}

	// Initial triangle
	initial := Polyiamond{
		Triangles: []Triangle{
			makeTriangle(Vertex{0, 0}, Vertex{1, 0}, Vertex{0, 1}),
		},
	}

	if n == 1 {
		return []Polyiamond{canonicalize(initial)}
	}

	current := map[string]Polyiamond{
		polyiamondKey(canonicalize(initial)): canonicalize(initial),
	}

	for size := 2; size <= n; size++ {
		fmt.Printf("  Size %d: processing %d shapes...\n", size, len(current))

		shapes := make([]Polyiamond, 0, len(current))
		for _, p := range current {
			shapes = append(shapes, p)
		}

		// Parallel processing
		var mu sync.Mutex
		next := make(map[string]Polyiamond)

		var wg sync.WaitGroup
		chunkSize := (len(shapes) + workers - 1) / workers

		for w := 0; w < workers; w++ {
			start := w * chunkSize
			end := start + chunkSize
			if end > len(shapes) {
				end = len(shapes)
			}
			if start >= len(shapes) {
				break
			}

			wg.Add(1)
			go func(chunk []Polyiamond) {
				defer wg.Done()
				localNext := make(map[string]Polyiamond)

				for _, shape := range chunk {
					for _, newTri := range getBoundary(shape) {
						newShape := addTriangle(shape, newTri)
						canon := canonicalize(newShape)
						key := polyiamondKey(canon)
						localNext[key] = canon
					}
				}

				mu.Lock()
				for k, v := range localNext {
					next[k] = v
				}
				mu.Unlock()
			}(shapes[start:end])
		}

		wg.Wait()
		current = next
	}

	result := make([]Polyiamond, 0, len(current))
	for _, p := range current {
		result = append(result, p)
	}
	return result
}

func polyiamondToGraph(p Polyiamond) (int, int) {
	vertices := make(map[Vertex]bool)
	edges := make(map[[2]Vertex]bool)

	for _, t := range p.Triangles {
		for _, v := range t {
			vertices[v] = true
		}
		for i := 0; i < 3; i++ {
			v1, v2 := t[i], t[(i+1)%3]
			if v1.A > v2.A || (v1.A == v2.A && v1.B > v2.B) {
				v1, v2 = v2, v1
			}
			edges[[2]Vertex{v1, v2}] = true
		}
	}

	return len(vertices), len(edges)
}

func polyiamondToCoords(p Polyiamond) ([]Vertex, [][2]int) {
	// Collect vertices and edges
	vertexSet := make(map[Vertex]bool)
	edgeSet := make(map[[2]Vertex]bool)

	for _, t := range p.Triangles {
		for _, v := range t {
			vertexSet[v] = true
		}
		for i := 0; i < 3; i++ {
			v1, v2 := t[i], t[(i+1)%3]
			if v1.A > v2.A || (v1.A == v2.A && v1.B > v2.B) {
				v1, v2 = v2, v1
			}
			edgeSet[[2]Vertex{v1, v2}] = true
		}
	}

	// Create sorted vertex list
	vertices := make([]Vertex, 0, len(vertexSet))
	for v := range vertexSet {
		vertices = append(vertices, v)
	}
	sort.Slice(vertices, func(i, j int) bool {
		if vertices[i].A != vertices[j].A {
			return vertices[i].A < vertices[j].A
		}
		return vertices[i].B < vertices[j].B
	})

	// Map vertices to indices
	vertexIdx := make(map[Vertex]int)
	for i, v := range vertices {
		vertexIdx[v] = i
	}

	// Build edge list with indices
	edges := make([][2]int, 0, len(edgeSet))
	for e := range edgeSet {
		edges = append(edges, [2]int{vertexIdx[e[0]], vertexIdx[e[1]]})
	}

	return vertices, edges
}

func polyiamondToGraph6(p Polyiamond) string {
	// Collect vertices and edges
	vertexSet := make(map[Vertex]bool)
	edgeSet := make(map[[2]Vertex]bool)

	for _, t := range p.Triangles {
		for _, v := range t {
			vertexSet[v] = true
		}
		for i := 0; i < 3; i++ {
			v1, v2 := t[i], t[(i+1)%3]
			if v1.A > v2.A || (v1.A == v2.A && v1.B > v2.B) {
				v1, v2 = v2, v1
			}
			edgeSet[[2]Vertex{v1, v2}] = true
		}
	}

	// Create sorted vertex list for consistent indexing
	vertices := make([]Vertex, 0, len(vertexSet))
	for v := range vertexSet {
		vertices = append(vertices, v)
	}
	sort.Slice(vertices, func(i, j int) bool {
		if vertices[i].A != vertices[j].A {
			return vertices[i].A < vertices[j].A
		}
		return vertices[i].B < vertices[j].B
	})

	// Map vertices to indices
	vertexIdx := make(map[Vertex]int)
	for i, v := range vertices {
		vertexIdx[v] = i
	}

	n := len(vertices)

	// Build adjacency matrix (upper triangle)
	adj := make([][]bool, n)
	for i := range adj {
		adj[i] = make([]bool, n)
	}
	for e := range edgeSet {
		i, j := vertexIdx[e[0]], vertexIdx[e[1]]
		if i > j {
			i, j = j, i
		}
		adj[i][j] = true
	}

	// Encode in graph6 format
	var result []byte

	// Encode n (assuming n <= 62)
	if n <= 62 {
		result = append(result, byte(n+63))
	} else {
		// For larger n, use extended encoding
		result = append(result, 126)
		result = append(result, byte((n>>12)+63))
		result = append(result, byte(((n>>6)&63)+63))
		result = append(result, byte((n&63)+63))
	}

	// Encode upper triangle bits
	var bits []bool
	for j := 1; j < n; j++ {
		for i := 0; i < j; i++ {
			bits = append(bits, adj[i][j])
		}
	}

	// Pad to multiple of 6
	for len(bits)%6 != 0 {
		bits = append(bits, false)
	}

	// Convert to characters
	for i := 0; i < len(bits); i += 6 {
		var val byte = 0
		for j := 0; j < 6; j++ {
			if bits[i+j] {
				val |= 1 << (5 - j)
			}
		}
		result = append(result, val+63)
	}

	return string(result)
}

func printPolyiamond(p Polyiamond, idx int, nTri int) {
	fmt.Printf("--- Polyiamond %d (%d triangles) ---\n", idx, nTri)

	// Find bounds
	minA, maxA, minB, maxB := 1000000, -1000000, 1000000, -1000000
	for _, t := range p.Triangles {
		for _, v := range t {
			if v.A < minA {
				minA = v.A
			}
			if v.A > maxA {
				maxA = v.A
			}
			if v.B < minB {
				minB = v.B
			}
			if v.B > maxB {
				maxB = v.B
			}
		}
	}

	// Determine triangle orientation and position
	type TriPos struct {
		q, r   int
		isUp   bool
	}

	triPositions := make([]TriPos, 0, len(p.Triangles))
	for _, t := range p.Triangles {
		bVals := []int{t[0].B, t[1].B, t[2].B}
		sort.Ints(bVals)
		isUp := bVals[0] == bVals[1] // Two lower vertices = up pointing

		var q, r int
		if isUp {
			// Find bottom-left vertex
			minV := t[0]
			for _, v := range t {
				if v.B < minV.B || (v.B == minV.B && v.A < minV.A) {
					minV = v
				}
			}
			q, r = minV.A, minV.B
		} else {
			// Find top-left vertex
			maxB := t[0].B
			for _, v := range t {
				if v.B > maxB {
					maxB = v.B
				}
			}
			minA := 1000000
			for _, v := range t {
				if v.B == maxB && v.A < minA {
					minA = v.A
				}
			}
			q, r = minA, maxB-1
		}
		triPositions = append(triPositions, TriPos{q - minA, r - minB, isUp})
	}

	// Find grid bounds
	maxQ, maxR := 0, 0
	for _, tp := range triPositions {
		if tp.q > maxQ {
			maxQ = tp.q
		}
		if tp.r > maxR {
			maxR = tp.r
		}
	}

	// Create grid
	height := maxR + 1
	width := (maxQ+1)*2 + maxR + 2
	grid := make([][]rune, height)
	for i := range grid {
		grid[i] = make([]rune, width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	// Place triangles
	for _, tp := range triPositions {
		row := maxR - tp.r
		col := tp.q*2 + tp.r
		if tp.isUp {
			grid[row][col] = '△'
		} else {
			grid[row][col+1] = '▽'
		}
	}

	// Print
	for _, row := range grid {
		line := string(row)
		// Trim trailing spaces
		for len(line) > 0 && line[len(line)-1] == ' ' {
			line = line[:len(line)-1]
		}
		fmt.Println(line)
	}
	fmt.Println()
}

func main() {
	minTri := flag.Int("min", 6, "Minimum triangles")
	maxTri := flag.Int("max", 15, "Maximum triangles")
	targetV := flag.Int("v", 13, "Target vertices")
	targetE := flag.Int("e", 26, "Target edges")
	workers := flag.Int("w", 0, "Number of workers (0 = num CPUs)")
	showShapes := flag.Bool("show", false, "Show matching shapes")
	g6Output := flag.String("g6", "", "Output matching graphs to this .g6 file")
	coordOutput := flag.String("coords", "", "Output vertex coordinates to this file (for plotting)")
	flag.Parse()

	if *workers == 0 {
		*workers = runtime.NumCPU()
	}

	fmt.Printf("Searching for polyiamonds with %d vertices and %d edges\n", *targetV, *targetE)
	fmt.Printf("Triangle range: %d to %d, workers: %d\n\n", *minTri, *maxTri, *workers)

	total := 0
	var allMatches []struct {
		p    Polyiamond
		nTri int
	}

	for nTri := *minTri; nTri <= *maxTri; nTri++ {
		fmt.Printf("n=%d triangles:\n", nTri)
		shapes := enumeratePolyiamonds(nTri, *workers)
		fmt.Printf("  Found %d polyiamonds\n", len(shapes))

		count := 0
		for _, p := range shapes {
			v, e := polyiamondToGraph(p)
			if v == *targetV && e == *targetE {
				count++
				if *showShapes || *g6Output != "" || *coordOutput != "" {
					allMatches = append(allMatches, struct {
						p    Polyiamond
						nTri int
					}{p, nTri})
				}
			}
		}

		fmt.Printf("  Matches (%d vertices, %d edges): %d\n\n", *targetV, *targetE, count)
		total += count
	}

	fmt.Printf("Total: %d\n", total)

	if *showShapes && len(allMatches) > 0 {
		fmt.Printf("\n=== Matching shapes ===\n\n")
		for i, m := range allMatches {
			printPolyiamond(m.p, i+1, m.nTri)
		}
	}

	if *g6Output != "" && len(allMatches) > 0 {
		f, err := os.Create(*g6Output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()

		for _, m := range allMatches {
			g6 := polyiamondToGraph6(m.p)
			fmt.Fprintln(f, g6)
		}
		fmt.Printf("\nWrote %d graphs to %s\n", len(allMatches), *g6Output)
	}

	if *coordOutput != "" && len(allMatches) > 0 {
		f, err := os.Create(*coordOutput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()

		// Deduplicate by edge signature
		seen := make(map[string]bool)
		graphIdx := 0

		for _, m := range allMatches {
			verts, edges := polyiamondToCoords(m.p)

			// Create signature for dedup
			sig := fmt.Sprintf("%v", edges)
			if seen[sig] {
				continue
			}
			seen[sig] = true

			graphIdx++
			fmt.Fprintf(f, "GRAPH %d\n", graphIdx)
			fmt.Fprintf(f, "VERTICES %d\n", len(verts))
			for _, v := range verts {
				fmt.Fprintf(f, "%d %d\n", v.A, v.B)
			}
			fmt.Fprintf(f, "EDGES %d\n", len(edges))
			for _, e := range edges {
				fmt.Fprintf(f, "%d %d\n", e[0], e[1])
			}
		}
		fmt.Printf("Wrote %d unique graphs to %s\n", graphIdx, *coordOutput)
	}
}
