package index

import (
	"log"
	"math"
	"slices"
)

const (
	maxSize = 3_000_000
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

func (h *VpTree) Query(vector [14]int16) int8 {
	pq := &PriorityQueue{candidates: [5]QueryCandidate{}, count: 0, farthestCandidate: 0}
	current := h.Nodes[0]
	tau := math.MaxInt
	h.search(&vector, &current, &tau, pq)
	var sum int8
	for _, c := range pq.candidates {
		if !c.vp.Label {
			sum += 1
		}
	}

	return sum
}

func (h *VpTree) search(vector *[14]int16, current *VpTreeNode, tau *int, pq *PriorityQueue) {
	d := calculateDistance(*vector, current.Vector)
	if int(d) < *tau || pq.count < 5 {
		qc := QueryCandidate{vp: current, d: d}
		pq.Add(qc)
		*tau = int(pq.candidates[pq.farthestCandidate].d)
	}

	if current.Left >= maxSize && current.Right >= maxSize {
		return
	}

	t := int(current.Threshold)
	if d < current.Threshold {
		if current.Left < maxSize {
			h.search(vector, &h.Nodes[current.Left], tau, pq)
		}
		if int(d)+*tau >= t && current.Right < maxSize {
			h.search(vector, &h.Nodes[current.Right], tau, pq)
		}
	} else {
		if current.Right < maxSize {
			h.search(vector, &h.Nodes[current.Right], tau, pq)
		}
		if int(d)-*tau <= t && current.Left < maxSize {
			h.search(vector, &h.Nodes[current.Left], tau, pq)
		}
	}
}

func BuildVpTree(t []*QuantizeTransaction) *VpTree {
	tree := [maxSize]VpTreeNode{}
	var c uint32
	build(t, &tree, &c)
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

	return idx
}

func calculateDistance(v1, v2 [14]int16) uint16 {
	var sum int64
	for i := range len(v1) {
		d := int64(v1[i]) - int64(v2[i])
		sum += d * d
	}
	r := int64(math.Sqrt(float64(sum)))
	if r*r < sum {
		r++
	}
	if r > math.MaxUint16 {
		log.Println("Overflow")
	}
	return uint16(r)
}

type VectorDistance struct {
	idx int
	d   uint16
}
