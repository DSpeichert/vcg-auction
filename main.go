package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
)

// Agent's bid (mapping of allocation => utility)
// index is a binary "flag", in which:
// right-most bit is item 0, second from the right is item 1 and so on
type Bid map[int64]float64

// Contains bids for all agents (1..n)
type BidSet []Bid

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
			u += bs[agent][flags]
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
	Allocation     Allocation
	TotalUtility   float64
	PricePerAgent  []float64
	BestSolution   *Solution
	SecondSolution *Solution
}

func (s *Solution) CalculatePrices(bs BidSet) {
	s.PricePerAgent = make([]float64, len(s.Allocation))
	for agent, _ := range s.Allocation {
		if agent > 0 {
			s.PricePerAgent[agent] = s.SecondSolution.Allocation.FindTotalUtility(bs) - s.Allocation.FindTotalUtilityExceptAgent(bs, agent)
		}
	}
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Pass n and m as arguments.")
		os.Exit(1)
	}
	n, _ := strconv.Atoi(os.Args[1])
	m, _ := strconv.Atoi(os.Args[2])
	fmt.Printf("Using n = %d agents and m = %d items\n", n, m)

	rand.Seed(time.Now().UnixNano())
	fmt.Println("Generating agent's utilities for all combinations of allocations to them.")
	start := time.Now()
	bs := randomizeBidSet(n, m)
	elapsed := time.Since(start)
	fmt.Printf("Randomizing input data took %s\n", elapsed)
	for agent, bid := range bs {
		if agent != 0 { // agent 0 is nobody!
			fmt.Printf("Bids for Agent %d\n", agent)
			for items, utility := range bid {
				fmt.Printf("  %0"+strconv.Itoa(m)+"b => %f\n", items, utility)
			}
		}
	}

	// start looking for solutions
	var solution Solution
	allocation := make(Allocation)
	for a := 0; a <= n; a++ {
		allocation[a] = make(map[int]bool)
	}
	start = time.Now()
	recursiveAllocationGenerator(&solution, bs, allocation, 0, m)
	solution.Allocation = solution.BestSolution.Allocation.Copy()
	solution.TotalUtility = solution.BestSolution.TotalUtility
	solution.CalculatePrices(bs)
	fmt.Printf("%+v\n", solution)
	elapsed = time.Since(start)
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

func recursiveAllocationGenerator(s *Solution, bs BidSet, a Allocation, current_item, items int) {
	for agent, al := range a {
		fmt.Printf("agent: %d, current_item: %d\n", agent, current_item)
		if agent != 0 {
			delete(a[agent-1], current_item)
		}
		al[current_item] = true

		if current_item < items-1 {
			recursiveAllocationGenerator(s, bs, a, current_item+1, items)
		} else {
			fmt.Printf("Considering allocation: %+v\n", a)
			total_utility := a.FindTotalUtility(bs)
			fmt.Printf("Total utility: %f\n", total_utility)

			if s.BestSolution == nil || s.BestSolution.TotalUtility < total_utility {
				s.SecondSolution = s.BestSolution
				s.BestSolution = &Solution{Allocation: a.Copy(), TotalUtility: total_utility}
			} else if s.SecondSolution == nil || s.SecondSolution.TotalUtility < total_utility {
				s.SecondSolution = &Solution{Allocation: a.Copy(), TotalUtility: total_utility}
			}

			// cleanup for backtrack
			delete(a[agent], current_item)
		}
	}
}
