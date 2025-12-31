package main

import (
	"bufio"
	"fmt"
	"os"
)

func decodeGraph6(s string) (int, [][2]int) {
	n := int(s[0]) - 63

	// Decode bits from remaining characters
	var bits []int
	for i := 1; i < len(s); i++ {
		val := int(s[i]) - 63
		for b := 5; b >= 0; b-- {
			bits = append(bits, (val>>b)&1)
		}
	}

	// Read upper triangle
	var edges [][2]int
	idx := 0
	for j := 1; j < n; j++ {
		for i := 0; i < j; i++ {
			if idx < len(bits) && bits[idx] == 1 {
				edges = append(edges, [2]int{i, j})
			}
			idx++
		}
	}

	return n, edges
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	graphNum := 1

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		n, edges := decodeGraph6(line)

		fmt.Printf("(* Graph %d: %d vertices, %d edges *)\n", graphNum, n, len(edges))
		fmt.Printf("graph%dEdges = {\n", graphNum)
		for i, e := range edges {
			if i < len(edges)-1 {
				fmt.Printf("  {%d, %d},\n", e[0], e[1])
			} else {
				fmt.Printf("  {%d, %d}\n", e[0], e[1])
			}
		}
		fmt.Printf("};\n\n")
		graphNum++
	}
}
