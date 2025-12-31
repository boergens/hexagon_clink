package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
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

// toGraph6 converts a graph to graph6 format (used by nauty)
func (g Graph) toGraph6() string {
	// Graph6 format:
	// 1. n encoded as single char (for n < 63: char = n + 63)
	// 2. Upper triangle of adjacency matrix, 6 bits per char

	result := []byte{byte(n + 63)}

	// Build upper triangle bits
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

	// Pad to multiple of 6
	for len(bits)%6 != 0 {
		bits = append(bits, 0)
	}

	// Convert to characters
	for i := 0; i < len(bits); i += 6 {
		val := bits[i]<<5 | bits[i+1]<<4 | bits[i+2]<<3 | bits[i+3]<<2 | bits[i+4]<<1 | bits[i+5]
		result = append(result, byte(val+63))
	}

	return string(result)
}

// toDIMACS converts a graph to DIMACS format (used by bliss)
func (g Graph) toDIMACS() string {
	edgeCount := 0
	for idx := 0; idx < numEdges; idx++ {
		if g&(1<<idx) != 0 {
			edgeCount++
		}
	}

	result := fmt.Sprintf("p edge %d %d\n", n, edgeCount)
	for idx := 0; idx < numEdges; idx++ {
		if g&(1<<idx) != 0 {
			i, j := edgePairs[idx][0], edgePairs[idx][1]
			result += fmt.Sprintf("e %d %d\n", i+1, j+1) // DIMACS is 1-indexed
		}
	}
	return result
}

func main() {
	if len(os.Args) < 5 {
		fmt.Println("Usage: convert <input.bin> <output> <n> <input-format> [output-format]")
		fmt.Println("  input.bin: binary file with graphs")
		fmt.Println("  output: output file")
		fmt.Println("  n: number of vertices")
		fmt.Println("  input-format: 'raw' or 'grouped'")
		fmt.Println("  output-format: 'g6' (default), 'dimacs', or 'dimacs-dir'")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]
	vertices, _ := strconv.Atoi(os.Args[3])
	inputFormat := os.Args[4]
	format := "g6"
	if len(os.Args) > 5 {
		format = os.Args[5]
	}

	initEdges(vertices)

	bytesPerGraph := 4
	if numEdges > 32 {
		bytesPerGraph = 8
	}

	f, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	reader := bufio.NewReader(f)

	var graphs []Graph

	if inputFormat == "raw" {
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
	} else if inputFormat == "grouped" {
		var numGroups uint32
		binary.Read(reader, binary.LittleEndian, &numGroups)
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
	} else {
		fmt.Printf("Unknown input format: %s (use 'raw' or 'grouped')\n", inputFormat)
		os.Exit(1)
	}

	fmt.Printf("Read %d graphs\n", len(graphs))

	switch format {
	case "g6":
		out, _ := os.Create(outputFile)
		defer out.Close()
		for _, g := range graphs {
			fmt.Fprintln(out, g.toGraph6())
		}
		fmt.Printf("Wrote %d graphs to %s in graph6 format\n", len(graphs), outputFile)

	case "dimacs":
		out, _ := os.Create(outputFile)
		defer out.Close()
		for i, g := range graphs {
			fmt.Fprintf(out, "c graph %d\n", i)
			fmt.Fprint(out, g.toDIMACS())
		}
		fmt.Printf("Wrote %d graphs to %s in DIMACS format\n", len(graphs), outputFile)

	case "dimacs-dir":
		os.MkdirAll(outputFile, 0755)
		for i, g := range graphs {
			fname := fmt.Sprintf("%s/graph_%06d.dimacs", outputFile, i)
			out, _ := os.Create(fname)
			fmt.Fprint(out, g.toDIMACS())
			out.Close()
		}
		fmt.Printf("Wrote %d graphs to %s/ in DIMACS format\n", len(graphs), outputFile)
	}
}
