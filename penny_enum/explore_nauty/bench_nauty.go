package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// Benchmark nauty's labelg tool for canonical labeling
// labelg reads graph6 format and outputs canonical graph6

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: bench_nauty <input.g6>")
		fmt.Println("  Benchmarks nauty's labelg on graph6 file")
		fmt.Println("")
		fmt.Println("Install nauty: brew install nauty")
		os.Exit(1)
	}

	inputFile := os.Args[1]

	// Count graphs
	f, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		count++
	}
	f.Close()
	fmt.Printf("Input: %d graphs\n", count)

	// Check if labelg exists
	_, err = exec.LookPath("labelg")
	if err != nil {
		fmt.Println("Error: labelg not found. Install with: brew install nauty")
		os.Exit(1)
	}

	// Run labelg (canonical labeling)
	fmt.Println("\n=== nauty labelg (canonical labeling) ===")
	start := time.Now()
	cmd := exec.Command("labelg", "-q", inputFile)
	output, err := cmd.Output()
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("Error running labelg: %v\n", err)
		os.Exit(1)
	}

	// Count unique canonical forms
	unique := make(map[string]bool)
	scanner = bufio.NewScanner(bufio.NewReader(
		&struct{ b []byte }{output},
	))
	// Actually parse the output properly
	lines := 0
	for i := 0; i < len(output); i++ {
		if output[i] == '\n' {
			lines++
		}
	}

	// Re-run to get unique count
	cmd = exec.Command("labelg", "-q", inputFile)
	cmd.Stdout = nil
	outPipe, _ := cmd.StdoutPipe()
	cmd.Start()
	scanner = bufio.NewScanner(outPipe)
	for scanner.Scan() {
		unique[scanner.Text()] = true
	}
	cmd.Wait()

	fmt.Printf("Time: %v\n", elapsed)
	fmt.Printf("Graphs/sec: %.0f\n", float64(count)/elapsed.Seconds())
	fmt.Printf("Unique canonical forms: %d\n", len(unique))

	// Also try shortg (removes isomorphic duplicates)
	fmt.Println("\n=== nauty shortg (deduplicate) ===")
	start = time.Now()
	cmd = exec.Command("shortg", "-q", inputFile)
	output, err = cmd.Output()
	elapsed = time.Since(start)

	if err != nil {
		fmt.Printf("Error running shortg: %v\n", err)
	} else {
		// Count output lines
		outCount := 0
		for _, b := range output {
			if b == '\n' {
				outCount++
			}
		}
		fmt.Printf("Time: %v\n", elapsed)
		fmt.Printf("Graphs/sec: %.0f\n", float64(count)/elapsed.Seconds())
		fmt.Printf("Unique graphs: %d\n", outCount)
	}
}

type byteReader struct {
	b []byte
	i int
}

func (r *byteReader) Read(p []byte) (n int, err error) {
	if r.i >= len(r.b) {
		return 0, os.ErrClosed
	}
	n = copy(p, r.b[r.i:])
	r.i += n
	return n, nil
}
