package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
)

// State is an interface of a state of a problem.
// These three methods will handle the state.
type State interface {
	Copy() interface{} // Returns an address of an exact copy of the current state
	Move()             // Move to a different state
	Energy() float64   // Return the energy of the current state
}

// Annealer performs simulated annealing by calling functions to calculate
// energy and make moves on a state.  The temperature schedule for
// annealing may be provided manually or estimated automatically.
type Annealer struct {
	// parameters
	Tmax    float64 // max temperature
	Tmin    float64 // minimum temperature
	Steps   int
	Updates int

	// user settings
	CopyStrategy string
	UserExit     bool

	// placeholders
	State      State
	bestState  State
	bestEnergy float64
	startTime  float64
}

// NewAnnealer initializes an Annealer struct
func NewAnnealer(initialState State) *Annealer {
	a := new(Annealer)
	a.State = initialState
	a.Tmax = 25000.0
	a.Tmin = 2.5
	a.Steps = 50000
	a.Updates = 100
	return a
}

// Outputs to stderr.
// Prints the current temperature, energy, acceptance rate, improvement rate, elapsed time, and remaining time.
// The acceptance rate indicates the percentage of moves since the last update
// that were accepted by the Metropolis algorithm.
// It includes moves that decreased the energy, moves that left the energy unchanged,
// and moves that increased the energy yet were reached by thermal excitation.
// The improvement rate indicates the percentage of moves since the last update that strictly decreased the energy.
// At high temperatures it will include both moves that improved the overall state and
// moves that simply undid previously accepted moves that increased the energy by thermal excititation.
// At low temperatures it will tend toward zero as the moves that can decrease the energy are exhausted and
// moves that would increase the energy are no longer thermally accessible.
func (a *Annealer) update(step int, T float64, E float64, acceptance float64, improvement float64) {
	elapsed := now() - a.startTime
	if step == 0 {
		fmt.Fprintln(os.Stderr, " Temperature        Energy    Accept   Improve     Elapsed   Remaining")
		fmt.Fprintf(os.Stderr, "\r%12.5f  %12.2f                      %s            ", T, E, timeString(elapsed))
	} else {
		remain := float64(a.Steps-step) * (elapsed / float64(step))
		fmt.Fprintf(os.Stderr, "\r%12.5f  %12.2f  %7.2f%%  %7.2f%%  %s  %s",
			T, E, 100.0*acceptance, 100.0*improvement, timeString(elapsed), timeString(remain))
	}
}

// Anneal minimizes the energy of a system by simulated annealing.
// Parameters
// state : an initial arrangement of the system
// Returns
// (state, energy): the best state and energy found.
func (a *Annealer) Anneal() (interface{}, float64) {
	step := 0
	a.startTime = now()

	// Precompute factor for exponential cooling from Tmax to Tmin
	if a.Tmin <= 0.0 {
		panic("Exponential cooling requires a minimum temperature greater than zero.")
	}
	Tfactor := -math.Log(a.Tmax / a.Tmin)

	// Note initial state
	T := a.Tmax
	E := a.State.Energy()
	prevState := a.State.Copy().(State)
	prevEnergy := E
	a.bestState = a.State.Copy().(State)
	a.bestEnergy = E
	trials, accepts, improves := 0, 0.0, 0.0
	var updateWavelength float64
	if a.Updates > 0 {
		updateWavelength = float64(a.Steps) / float64(a.Updates)
		a.update(step, T, E, 0.0, 0.0)
	}

	// Attempt moves to new states
	for step < a.Steps && !a.UserExit {
		step++
		T = a.Tmax * math.Exp(Tfactor*float64(step)/float64(a.Steps))
		a.State.Move()
		E := a.State.Energy()
		dE := E - prevEnergy
		trials++
		if dE > 0.0 && math.Exp(-dE/T) < rand.Float64() {
			// Restore previous state
			a.State = prevState.Copy().(State)
			E = prevEnergy
		} else {
			// Accept new state and compare to best state
			accepts += 1.0
			if dE < 0.0 {
				improves += 1.0
			}
			prevState = a.State.Copy().(State)
			prevEnergy = E
			if E < a.bestEnergy {
				a.bestState = a.State.Copy().(State)
				a.bestEnergy = E
			}
		}
		if a.Updates > 1 {
			if (step / int(updateWavelength)) > ((step - 1) / int(updateWavelength)) {
				a.update(step, T, E, accepts/float64(trials), improves/float64(trials))
				trials, accepts, improves = 0, 0.0, 0.0
			}
		}
	}
	fmt.Fprintln(os.Stderr, "")

	// Return best state and energy
	return a.bestState, a.bestEnergy
}

type CandidateState struct {
	state Candidate
}

// Returns an address of an exact copy of the receiver's state
func (ts *CandidateState) Copy() interface{} {
	cs := CandidateState{}

	cp := cs.state

	cp.Solution = make([]int, len(ts.state.Solution))
	cp.Hubs = make([]int, len(ts.state.Hubs))
	copy(cp.Solution, ts.state.Solution)
	copy(cp.Hubs, ts.state.Hubs)
	cs.state = cp

	return &cs

}

// Swaps two cities in the route.
func (ts *CandidateState) Move() {

	// ts.state.Print()
	ts.state, _ = generateCandidateTypeA(ts.state)
	ts.state.calcCost(alpha)
	// a := rand.Intn(len(ts.state))
	// b := rand.Intn(len(ts.state))
	// ts.state[a], ts.state[b] = ts.state[b], ts.state[a]
}

// Calculates the length of the route.
func (ts *CandidateState) Energy() float64 {
	ts.state.calcCost(alpha)
	return ts.state.NormalizedCost
}

func SA() {
	// initial state, a randomly-ordered itinerary
	initial_solution := get_initial_solution(cost_matrix, flow_matrix, alpha, no_hubs)
	init_state := CandidateState{state: initial_solution}
	tsp := NewAnnealer(&init_state)
	tsp.Steps = 10000000
	tsp.Tmax = 10000

	state, _ := tsp.Anneal()
	ts := state.(*CandidateState)

	ts.state.calcCost(alpha)

	ts.state.Print()
}
