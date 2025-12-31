package main

import "math"

var hexDirs = [6][2]float64{
	{1.5, 0},
	{0.75, 1.3},
	{-0.75, 1.3},
	{-1.5, 0},
	{-0.75, -1.3},
	{0.75, -1.3},
}

type Vec2 struct {
	x, y float64
}

type Edge struct {
	a, b int
}

func vecClose(a, b Vec2) bool {
	return math.Abs(a.x-b.x) < 0.1 && math.Abs(a.y-b.y) < 0.1
}

func vecAdd(a Vec2, dx, dy float64) Vec2 {
	return Vec2{a.x + dx, a.y + dy}
}

func positionOccupied(pos Vec2, positions []Vec2, count int) bool {
	for i := 0; i < count; i++ {
		if vecClose(pos, positions[i]) {
			return true
		}
	}
	return false
}

func buildSpiral(n int) ([]Edge, int) {
	positions := make([]Vec2, n)
	edges := make([]Edge, 0, n*3)

	if n < 1 {
		return edges, 0
	}

	positions[0] = Vec2{0, 0}
	if n == 1 {
		return edges, 0
	}

	for node := 1; node < n; node++ {
		prevPos := positions[node-1]
		var bestPos Vec2
		bestContacts := -1
		bestDist := 1e9

		for d := 0; d < 6; d++ {
			cand := vecAdd(prevPos, hexDirs[d][0], hexDirs[d][1])
			if positionOccupied(cand, positions, node) {
				continue
			}

			contacts := 0
			for i := 0; i < node; i++ {
				for dd := 0; dd < 6; dd++ {
					neighbor := vecAdd(positions[i], hexDirs[dd][0], hexDirs[dd][1])
					if vecClose(cand, neighbor) {
						contacts++
						break
					}
				}
			}

			dist := cand.x*cand.x + cand.y*cand.y
			if contacts > bestContacts || (contacts == bestContacts && dist < bestDist) {
				bestPos = cand
				bestContacts = contacts
				bestDist = dist
			}
		}

		positions[node] = bestPos

		for i := 0; i < node; i++ {
			for d := 0; d < 6; d++ {
				neighbor := vecAdd(positions[i], hexDirs[d][0], hexDirs[d][1])
				if vecClose(bestPos, neighbor) {
					edges = append(edges, Edge{i, node})
					break
				}
			}
		}
	}
	return edges, len(edges)
}
