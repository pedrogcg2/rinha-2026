package index

import (
	"log"
	"slices"
)

const (
	maxSize = 3_000_000
	Scale   = 1_000
)

type VpTreeNode struct {
	Left      uint32
	Right     uint32
	Vector    [14]int16
	Label     bool
	Threshold uint16
}

type VpTree struct {
	Nodes [maxSize]VpTreeNode
}

type queryCandidate struct {
	vp *VpTreeNode
	d  uint16
}

func (h *VpTree) Query(vector [14]int16) float64 {
	candidates := make([]queryCandidate, 0, 5)
	farthestCandidate := 0
	var idx uint32
	for idx < maxSize {
		current := &h.Nodes[idx]
		d := calculateDistance(vector, current.Vector)
		qc := queryCandidate{vp: current, d: d}
		if len(candidates) < 5 {
			candidates = append(candidates, queryCandidate{vp: current, d: d})
			if candidates[farthestCandidate].d < qc.d {
				farthestCandidate = len(candidates) - 1
			}
		} else if qc.d < candidates[farthestCandidate].d {
			candidates[farthestCandidate] = qc
			farthestCandidate = recalculateFarthestCandidate(candidates)
		}
		if d > current.Threshold {
			idx = current.Right
			continue
		}
		idx = current.Left
	}

	var sum float64
	for _, c := range candidates {
		if !c.vp.Label {
			sum += 1.0
		}
	}

	return sum / 5
}

func recalculateFarthestCandidate(candidates []queryCandidate) int {
	idx := 0
	farthestDistance := candidates[0].d
	for i, c := range candidates {
		if c.d > farthestDistance {
			farthestDistance = c.d
			idx = i
		}
	}
	return idx
}

// TODO: NAO TA PREENCEHNDO TODOS OS INDICIES
// TEM ALGUMA COISA ERRADA
func BuildVpTree(t []*QuantizeTransaction) *VpTree {
	tree := [maxSize]VpTreeNode{}
	var c uint32
	build(t, &tree, &c)
	log.Println(c)
	return &VpTree{Nodes: tree}
}

func build(points []*QuantizeTransaction, nodes *[maxSize]VpTreeNode, next *uint32) uint32 {
	if len(points) == 0 || *next >= maxSize {
		return maxSize
	}
	idx := *next
	*next++
	vp := points[0]
	if len(points) == 1 {
		nodes[idx] = VpTreeNode{
			Left:      maxSize + 1,
			Right:     maxSize + 1,
			Vector:    [14]int16(vp.Vector),
			Label:     vp.Legit,
			Threshold: 0,
		}
		return idx
	}

	distances := make([]VectorDistance, len(points)-1)
	for i := range points {
		if i == 0 {
			continue
		}
		d := calculateDistance(vp.Vector, points[i].Vector)

		distances[i-1] = VectorDistance{
			idx: i,
			d:   d,
		}
	}

	slices.SortFunc(distances, func(a, b VectorDistance) int {
		if a.d > b.d {
			return 1
		}
		if a.d == b.d {
			return 0
		}
		return -1
	})

	median := len(distances) / 2
	threshold := distances[median].d
	left := make([]*QuantizeTransaction, median)
	for i, l := range distances[:median] {
		left[i] = points[l.idx]
	}

	right := make([]*QuantizeTransaction, len(distances)-median)
	for i, r := range distances[median:] {
		right[i] = points[r.idx]
	}
	nodes[idx] = VpTreeNode{
		Vector:    vp.Vector,
		Label:     vp.Legit,
		Threshold: threshold,
		Left:      build(left, nodes, next),
		Right:     build(right, nodes, next),
	}

	log.Printf("index: %d - Vector: v%", idx, vp.Vector)
	return idx
}

func calculateDistance(v1, v2 [14]int16) uint16 {
	var sum uint16
	for i := range len(v1) {
		d := v1[i] - v2[i]
		sum += uint16(d * d)
	}
	return sum
}

type VectorDistance struct {
	idx int
	d   uint16
}
