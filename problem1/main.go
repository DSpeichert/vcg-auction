package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

// Agent's bid (mapping of allocation => utility)
// index is a binary "flag", in which:
// right-most bit is item 0, second from the right is item 1 and so on
type Bid map[int64]float64

// Contains bids for all agents (1..n)
type BidSet []Bid

func (bs BidSet) CopyExcludingAgent(agent int) (new_bs BidSet) {
	new_bs = make(BidSet, len(bs)-1)
	for a, bid := range bs {
		if a < agent {
			new_bs[a] = make(Bid)
			for k, v := range bid {
				new_bs[a][k] = v
			}
		} else if a > agent {
			new_bs[a-1] = make(Bid)
			for k, v := range bid {
				new_bs[a-1][k] = v
			}
		}
	}
	return
}

// Allocation: Agent x Item = Bool
// Agent 0 is "nobody"
type Allocation map[int]map[int]bool

func (a Allocation) FindTotalUtility(bs BidSet) (u float64) {
	for agent, items := range a {
		var flags int64
		for item, _ := range items {
			flags = flags | 1<<uint(item)
		}
		if agent > 0 {
			var value float64
			switch agent {
			case 1:
				if v, ok := bs[agent][flags]; ok {
					value = v
				} else {
					value = 5
				}
			case 2:
				if v, ok := bs[agent][flags]; ok {
					value = v
				} else {
					value = 4
				}
			case 3:
				if v, ok := bs[agent][flags]; ok {
					value = v
				} else {
					value = 3
				}
			case 4:
				if v, ok := bs[agent][flags]; ok {
					value = v
				} else {
					value = 3
				}
			}
			u += value
		}
	}
	return
}

func (a Allocation) FindTotalUtilityExceptAgent(bs BidSet, excluded_agent int) (u float64) {
	for agent, items := range a {
		var flags int64
		for item, _ := range items {
			flags = flags | 1<<uint(item)
		}
		if agent > 0 && agent != excluded_agent {
			u += bs[agent][flags]
		}
	}
	return
}

func (a Allocation) Copy() (c Allocation) {
	c = make(Allocation)
	for k, v := range a {
		c[k] = make(map[int]bool)
		for k2, v2 := range v {
			c[k][k2] = v2
		}
	}
	return
}

type Solution struct {
	Allocation    Allocation
	TotalUtility  float64
	PricePerAgent []float64
}

func (s *Solution) CalculatePrices(bs BidSet, n, m int) {
	s.PricePerAgent = make([]float64, len(s.Allocation))
	for agent, _ := range s.Allocation {
		if agent > 0 {
			new_bs := bs.CopyExcludingAgent(agent)
			alternative_solution := solveAllocation(new_bs, n-1, m)
			s.PricePerAgent[agent] = alternative_solution.TotalUtility - s.Allocation.FindTotalUtilityExceptAgent(bs, agent)
		}
	}
}

func main() {
	n := 4
	m := 4
	fmt.Printf("Using n = %d agents and m = %d items\n", n, m)

	bs := make(BidSet, 5)
	for k, _ := range bs {
		bs[k] = make(Bid)
	}

	bs[1][0] = 0
	bs[1][1<<0] = 1 // agent 1, item 1
	bs[1][1<<1] = 2
	bs[1][1<<2] = 2
	bs[1][1<<3] = 4
	bs[1][1|1<<1|1<<2|1<<3] = 4 // agent 1, items {a, b, c, d}

	bs[2][0] = 0
	bs[2][1<<0] = 1   // agent 2, item a
	bs[2][1<<1] = 1   // agent 2, item b
	bs[2][1<<2] = 1   // agent 2, item c
	bs[2][1<<3] = 1   // agent 2, item d
	bs[2][1|1<<1] = 5 // agent 2, item {a, b}

	bs[3][0] = 0
	bs[3][1<<0] = 1      // agent 3, item a
	bs[3][1<<1] = 2      // agent 3, item b
	bs[3][1<<2] = 4      // agent 3, item c
	bs[3][1<<3] = 1      // agent 3, item d
	bs[3][1<<1|1<<2] = 7 // agent 2, item {b, c}

	bs[4][0] = 0
	bs[4][1<<0] = 1 // agent 4, item a
	bs[4][1<<1] = 1 // agent 4, item b
	bs[4][1<<2] = 1 // agent 4, item c
	bs[4][1<<3] = 3 // agent 4, item d

	for agent, bid := range bs {
		if agent != 0 { // agent 0 is nobody!
			fmt.Printf("Bids for Agent %d\n", agent)
			for items, utility := range bid {
				fmt.Printf("  %0"+strconv.Itoa(m)+"b => %f\n", items, utility)
			}
		}
	}

	// start looking for solutions
	start := time.Now()
	solution := solveAllocation(bs, n, m)
	solution.CalculatePrices(bs, n, m)
	elapsed := time.Since(start)
	fmt.Printf("%+v\n", solution)
	fmt.Printf("Finding solution took %s\n", elapsed)
}

// this is not parallel - no need to synchronize map writes
func randomizeBidSet(n, m int) (bs BidSet) {
	bs = make(BidSet, n+1)
	for a := 1; a <= n; a++ {
		bs[a] = getRandomBid(m)
	}
	return
}

func getRandomBid(m int) (b Bid) {
	b = make(Bid)
	recursiveRandomBidGenerator(b, 0, 0, 1, m)
	return
}

func recursiveRandomBidGenerator(b Bid, carry int64, previous_sum int, current_bit, bits int) {
	new_carry := carry                                    // prepending 0
	b[new_carry] = float64(previous_sum) * rand.Float64() // no utility for no items (sum == 0)
	if current_bit < bits {
		recursiveRandomBidGenerator(b, new_carry, previous_sum, current_bit+1, bits)
	}

	new_carry = carry | 1<<uint(current_bit-1) // prepending 1 but current_bit = 1 is actually "array index 0"
	b[new_carry] = float64(previous_sum+1) * rand.Float64()
	if current_bit < bits {
		recursiveRandomBidGenerator(b, new_carry, previous_sum+1, current_bit+1, bits)
	}
}

func solveAllocation(bs BidSet, n, m int) (s Solution) {
	allocation := make(Allocation)
	for a := 0; a <= n; a++ {
		allocation[a] = make(map[int]bool)
	}
	recursiveAllocationGenerator(&s, bs, allocation, 0, m)
	return
}

func recursiveAllocationGenerator(s *Solution, bs BidSet, a Allocation, current_item, items int) {
	for agent := 0; agent < len(a); agent++ {
		fmt.Printf("agent: %d, current_item: %d\n", agent, current_item)
		a[agent][current_item] = true

		if current_item < items-1 {
			recursiveAllocationGenerator(s, bs, a, current_item+1, items)
			delete(a[agent], current_item)
		} else {
			fmt.Printf("Considering allocation: %+v\n", a)
			total_utility := a.FindTotalUtility(bs)
			fmt.Printf("Total utility: %f\n", total_utility)

			if s.TotalUtility < total_utility {
				s.Allocation = a.Copy()
				s.TotalUtility = total_utility
			}

			// cleanup for backtrack
			delete(a[agent], current_item)
		}
	}
}
