package index

type PriorityQueue struct {
	candidates        [5]QueryCandidate
	count             int8
	farthestCandidate int8
}

func (pq *PriorityQueue) Add(qc QueryCandidate) {
	if pq.count < 5 {
		pq.candidates[pq.count] = qc
		if pq.count == 0 || qc.d > pq.candidates[pq.farthestCandidate].d {
			pq.farthestCandidate = pq.count
		}
		pq.count++
		return
	}

	if qc.d > pq.candidates[pq.farthestCandidate].d {
		return
	}

	pq.candidates[pq.farthestCandidate] = qc
	max := pq.candidates[0].d
	var maxIdx int8
	for i := int8(1); i < 5; i++ {
		if pq.candidates[i].d > max {
			max = pq.candidates[i].d
			maxIdx = i
		}
	}
	pq.farthestCandidate = maxIdx
}
