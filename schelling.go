package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"sort"
	"sync"
	"syscall/js"
)

// Keep track of status at each iteration.
type status struct {
	xs, ys []float64
	ts     []bool
}

// Make status by deep copying.
func (s *status) deepCopy() *status {
	ns := status{}
	ns.xs = append(make([]float64, 0, len(s.xs)), s.xs...)
	ns.ys = append(make([]float64, 0, len(s.ys)), s.ys...)
	ns.ts = append(make([]bool, 0, len(s.ts)), s.ts...)
	return &ns
}

// Convert status to Json
func (s *status) json() []any {
	res := []any{}
	for i := 0; i < len(s.ts); i++ {
		res = append(res, []any{s.xs[i], s.ys[i], s.ts[i]})
	}
	return res
}

// Storing the distance and type of neighbours together so that when
// sorting neighbours by distance, we also keep track of their type
type neighbour struct {
	d float64
	t bool
}

// Calculate the euclidean distance between two points
// [x, y]
func distance(a, b [2]float64) float64 {
	x := (a[0] - b[0]) * (a[0] - b[0])
	y := (a[1] - b[1]) * (a[1] - b[1])
	return math.Sqrt(x + y)
}

// Determine if agent at `idx` is happy at the specific `loc` location.
// The agent is happy if at least `k` out of `n` of its neighbours are
// of the same type `t`.
// Neighbours are the 10 closest agents based on euclidean distance.
// The locations and types of other agents are in `all`.
func isAgentHappyAtLocation(all *status, idx int, loc [2]float64, t bool, n, k int) bool {
	var ns []neighbour
	for i := 0; i < len(all.ts); i++ {
		if idx == i {
			continue
		}
		o := [2]float64{all.xs[i], all.ys[i]} // other agent
		ns = append(ns, neighbour{d: distance(loc, o), t: all.ts[i]})
	}

	// now sort by distance and get the `n` neighbours
	sort.Slice(ns, func(i, j int) bool {
		return ns[i].d < ns[j].d
	})

	// check how many neighbours are of the same type
	st := 0 // total neighbours of same type
	for _, o := range ns[:n] {
		if o.t == t {
			st++
		}
	}

	// is happy if number of same type agents > `k`
	return st >= k
}

// Determine if agent at `idx` is happy within the status `all`.
func isAgentHappy(all *status, idx, n, k int) bool {
	a := [2]float64{all.xs[idx], all.ys[idx]} // the agent
	return isAgentHappyAtLocation(all, idx, a, all.ts[idx], n, k)
}

// Determine which agents are unhappy with their current location.
// Agents are happy if at least `nst` out of `nn` neighbours are of
// the same type. An agent's neighbours are the 10 closest agents in
// based on the euclidean distance.
// It returns the indices of unhappy agents in `all`.
func unhappyAgents(all *status, nn, nst int) []int {
	na := len(all.ts)

	// get the happiness status for all agents
	unhappy := make(chan int, na)
	var wg sync.WaitGroup

	for j := 0; j < na; j++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if !isAgentHappy(all, idx, nn, nst) {
				unhappy <- idx
			}
		}(j)
	}
	wg.Wait()
	close(unhappy)

	// copy channel into a slice and return
	var indices []int
	for i := range unhappy {
		indices = append(indices, i)
	}
	return indices
}

// Move unhappy agents to new locations until they are happy.
// `all` is the current status of all agents, and `unhappy` contains
// the indices of unhappy agents.
// To be happy an agent needs 'nst' out of 'nn' neighbours to
// be of the same time. All agents make decisions at the same time,
// so they can't see where others have moved concurrently.
func moveAgents(all *status, unhappy []int, nn, nst int) {
	var wg sync.WaitGroup
	newLocs := make(chan [2]float64, len(unhappy))
	for _, j := range unhappy {
		wg.Add(1)
		// keep looking until happy with new location
		t := all.ts[j]
		go func(idx int) {
			defer wg.Done()
			for {
				// move to a new random location
				newLoc := [2]float64{rand.Float64(), rand.Float64()}

				if isAgentHappyAtLocation(all, idx, newLoc, t, nn, nst) {
					newLocs <- newLoc
					break
				}
			}
		}(j)
	}
	wg.Wait()
	close(newLocs)

	// now update locations for all agents that moved
	k := 0
	for j := range newLocs {
		idx := unhappy[k]
		all.xs[idx], all.ys[idx] = j[0], j[1]
		k++
	}
}

func Simulate(agents, neighbours, sameType, iterations int) []*status {
	allStatus := []*status{} // log status over all iterations

	// create agents and assign them location and type
	locs := make([][]float64, 2)
	locs[0], locs[1] = make([]float64, agents), make([]float64, agents)
	ts := make([]bool, agents)

	for i := 0; i < agents; i++ {
		locs[0][i], locs[1][i] = rand.Float64(), rand.Float64()
		ts[i] = rand.IntN(2) == 1
	}

	// save initial state
	allStatus = append(allStatus, &status{xs: locs[0], ys: locs[1], ts: ts})

	// simulate
	for i := 0; i < iterations; i++ {
		cur := allStatus[i].deepCopy()

		// indentify agents that aren't happy with their current location
		unhappy := unhappyAgents(cur, neighbours, sameType)
		fmt.Println("iteration:", i, "unhappy agents:", len(unhappy))
		if len(unhappy) == 0 { // stop if all happy
			break
		}

		// allow unhappy agents to move until they've found a location where
		// they will be happy
		moveAgents(cur, unhappy, neighbours, sameType)

		// update status
		allStatus = append(allStatus, cur)
	}
	return allStatus
}

func jsonWrapper() js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 4 {
			return "Invalid no of arguments passed"
		}
		agents := args[0].Int()
		neighbours := args[1].Int()
		sameType := args[2].Int()
		runs := args[3].Int()

		status := Simulate(agents, neighbours, sameType, runs)
		json := []any{}
		for _, s := range status {
			json = append(json, s.json())
		}
		return json
	})
	return jsonFunc
}

func main() {
	js.Global().Set("simulate", jsonWrapper())
	<-make(chan struct{})
	// Simulate(1000, 10, 5, 10)
}
