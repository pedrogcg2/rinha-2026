package preprocessor

import (
	"log"
	"runtime"
	"slices"
	"time"

	"gonum.org/v1/gonum/mat"
)

type VpTreeNode struct {
	left      *VpTreeNode
	right     *VpTreeNode
	t         *transaction
	threshold float64
}

type queryCandidate struct {
	vp VpTreeNode
	d  float64
}

func (h *VpTreeNode) Query(vector *mat.VecDense) float64 {
	candidates := make([]queryCandidate, 0, 5)
	farthestCandidate := 0
	current := h
	for current != nil {
		d := calculateDistance(vector, current.t.embedding)
		qc := queryCandidate{vp: *current, d: d}
		if len(candidates) < 5 {
			candidates = append(candidates, queryCandidate{vp: *current, d: d})
			if candidates[farthestCandidate].d < qc.d {
				farthestCandidate = len(candidates) - 1
			}
		} else if qc.d < candidates[farthestCandidate].d {
			candidates[farthestCandidate] = qc
			farthestCandidate = recalculateFarthestCandidate(candidates)
		}
		if d > current.threshold {
			current = current.right
			continue
		}
		current = current.left

	}

	var sum float64
	for _, c := range candidates {
		if !c.vp.t.legit {
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

func buildVpTree(t []transaction) *VpTreeNode {
	root := build(t)
	log.Println("Calling GC AGAIN")
	time.Sleep(time.Second * 3)
	runtime.GC()
	return root
}

func build(points []transaction) *VpTreeNode {
	if len(points) == 0 {
		return nil
	}
	vp := points[0]
	if len(points) == 1 {
		return &VpTreeNode{t: &vp}
	}
	others := points[1:]
	distances := make([]VectorDistance, 0, len(others))
	for i := range others {
		d := calculateDistance(vp.embedding, others[i].embedding)
		distances = append(distances, VectorDistance{
			t:   &others[i],
			d:   d,
			idx: i,
		})
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

	var left []transaction
	for _, l := range distances[:median] {
		left = append(left, others[l.idx])
	}

	var right []transaction
	for _, r := range distances[median:] {
		right = append(right, others[r.idx])
	}
	return &VpTreeNode{
		t:         &vp,
		threshold: threshold,
		left:      build(left),
		right:     build(right),
	}
}

func calculateDistance(v1, v2 *mat.VecDense) float64 {
	result := mat.NewVecDense(14, nil)
	result.SubVec(v1, v2)
	result.MulElemVec(result, result)
	var sum float64
	for idx := range 13 {
		sum += result.At(idx, 0)
	}
	return sum
}

type VectorDistance struct {
	t   *transaction
	d   float64
	idx int
}
