package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"time"

	// "hash/fnv"
	"sort"
	// "sync"
)

var cost_matrix = [][]float64{}
var flow_matrix = [][]float64{}
var total_flow float64

var no_hubs = 3
var alpha = 0.8
var no_routines = 10

// Configurations
var iterations = 10
var maxCandidates = 60
var tabuSizeDivider = 5
var maxCandidatesMultiplier = 5
var aspiration = 4

func read_matrix(location string, no_nodes int) (matrix [][]float64, err error) {
	f, _ := os.Open(location)

	// Create a new reader.
	r := csv.NewReader(bufio.NewReader(f))

	row := []float64{}

	for {
		record, err := r.Read()

		// Stop at EOF.
		if err == io.EOF {
			break
		}

		for i := 0; i < no_nodes; i++ {
			value, err := strconv.ParseFloat(record[i], 10)
			if err != nil {
				return matrix, err
			} else {
				row = append(row, value)
			}
		}
		matrix = append(matrix, row)
		row = []float64{}
	}

	return matrix, err
}

func init() {
}

func isInSlice(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

type Candidate struct {
	Solution       []int
	Cost           float64
	Hubs           []int
	NormalizedCost float64
	SwappedNode    int
	ElapsedTime    time.Duration
	Iteration      int
}

type CandidateVector []Candidate

func (c CandidateVector) Len() int {
	return len(c)
}

func (c CandidateVector) Less(i, j int) bool {
	return c[i].Cost < c[j].Cost
}

func (c CandidateVector) Swap(i, j int) {
	c[j], c[i] = c[i], c[j]
}

func (c Candidate) Print() {
	// fmt.Fprintln(os.Stderr, "")
	fmt.Printf("Hubs: %+v\n", c.Hubs)
	fmt.Printf("Nodes:  \t")
	for i, _ := range c.Solution {
		fmt.Printf("%-2d\t", i+1)
	}
	fmt.Printf("\n")
	fmt.Printf("Solution:\t")
	for _, n := range c.Solution {
		fmt.Printf("%-2d\t", n+1)
	}
	fmt.Printf("\nTotal Cost: %+v\n", c.Cost)
	fmt.Printf("Normalized Cost: %+v\n", c.NormalizedCost)

}

func (c Candidate) PrintTable() {
	// fmt.Fprintln(os.Stderr, "")
	fmt.Printf("Hubs Locations: %+v\n", c.Hubs)
	fmt.Printf("Total Cost: %+v\n", c.Cost)
	fmt.Printf("Normalized Cost: %+v\n", c.NormalizedCost)
	fmt.Printf("Time: %s\n", c.ElapsedTime)

}

func (c *Candidate) calcCost(alpha float64) {
	c.Cost = calcTotalCost(cost_matrix, flow_matrix, alpha, c.Solution)
	c.NormalizedCost = c.Cost / total_flow
}

func get_initial_solution(cost_matrix, flow_matrix [][]float64, alpha float64, number_of_hubs int) Candidate {
	candidate := Candidate{}

	// randomly select certain number of hubs
	rand.Seed(time.Now().UnixNano())
	for len(candidate.Hubs) < number_of_hubs {
		random_number := rand.Intn(len(cost_matrix))
		if !isInSlice(random_number, candidate.Hubs) {
			candidate.Hubs = append(candidate.Hubs, random_number)
		}
	}

	// allocate nodes to their nearest candidate.Hubs
	for i, _ := range cost_matrix {
		target_hub := candidate.Hubs[0]
		for _, hub := range candidate.Hubs {
			if cost_matrix[i][hub] < cost_matrix[i][target_hub] {
				target_hub = hub
			}
		}
		candidate.Solution = append(candidate.Solution, target_hub)
	}

	return candidate
}

//  calculate total_cost follow Spoke-Hub-Hub-spoke strategy
func calcTotalCost(cost_matrix, flow_matrix [][]float64, alpha float64, solution []int) float64 {
	var total_cost float64
	total_cost = 0.0
	for i, _ := range flow_matrix {
		for j, _ := range flow_matrix {
			collection_cost := flow_matrix[i][j] * cost_matrix[i][solution[i]]
			transportation_cost := flow_matrix[i][j] * cost_matrix[solution[i]][solution[j]] * alpha
			distribution_cost := flow_matrix[i][j] * cost_matrix[solution[j]][j]
			cost := collection_cost + transportation_cost + distribution_cost
			total_cost += cost
		}
	}
	return total_cost
}

func calcTotalFlow(flow_matrix [][]float64) float64 {
	total_flow := 0.0
	for _, c := range flow_matrix {
		for _, e := range c {
			total_flow += e
		}
	}
	return total_flow
}

func isInTabuList(node int, tabuList []int) bool {
	for _, t := range tabuList {
		if node == t {
			return true
		}
	}
	return false
}

func selectRandomNodeAndHub(best Candidate) (int, int) {

	// fmt.Printf(best.Solution)
	selected_node := rand.Intn(len(best.Solution))
	for isInSlice(selected_node, best.Hubs) {
		selected_node = rand.Intn(len(best.Solution))
	}

	// select another random hub to assign to
	selected_hub := best.Hubs[rand.Intn(len(best.Hubs))]
	for selected_hub == best.Solution[selected_node] {
		selected_hub = best.Hubs[rand.Intn(len(best.Hubs))]
	}

	return selected_node, selected_hub
}

func updateTabuList(node int, tabuList *[]int, tabuSize int) {
	if len(*tabuList) == tabuSize {
		*tabuList = (*tabuList)[1:]
	}
	*tabuList = append(*tabuList, node)
}

// swap a node with its hub
func generateCandidateTypeA(current_solution Candidate) (c Candidate, swapped_node int) {
	neighbor := Candidate{}

	neighbor.Solution = make([]int, len(current_solution.Solution))
	neighbor.Hubs = make([]int, len(current_solution.Hubs))
	copy(neighbor.Solution, current_solution.Solution)
	copy(neighbor.Hubs, current_solution.Hubs)

	random_node, _ := selectRandomNodeAndHub(neighbor)

	hub_to_switch := neighbor.Solution[random_node]

	for i, hub := range neighbor.Solution {
		if hub == hub_to_switch {
			neighbor.Solution[i] = random_node
		}
	}

	for i, hub := range neighbor.Hubs {
		if hub == hub_to_switch {
			neighbor.Hubs[i] = random_node
		}
	}

	return neighbor, random_node
}

// swap two non hub nodes
func generateCandidateTypeB(current_solution Candidate) (c Candidate, swapped_node int) {
	neighbor := Candidate{}

	neighbor.Solution = make([]int, len(current_solution.Solution))
	neighbor.Hubs = make([]int, len(current_solution.Hubs))
	copy(neighbor.Solution, current_solution.Solution)
	copy(neighbor.Hubs, current_solution.Hubs)

	random_node_1, _ := selectRandomNodeAndHub(neighbor)
	random_node_2, _ := selectRandomNodeAndHub(neighbor)

	hub_node_1 := neighbor.Solution[random_node_1]
	neighbor.Solution[random_node_1] = neighbor.Solution[random_node_2]
	neighbor.Solution[random_node_2] = hub_node_1

	return neighbor, random_node_1
}

// reallocate a random node to a new hub
func generateCandidateTypeC(current_solution Candidate) (c Candidate, swapped_node int) {
	neighbor := Candidate{}

	neighbor.Solution = make([]int, len(current_solution.Solution))
	neighbor.Hubs = make([]int, len(current_solution.Hubs))
	copy(neighbor.Solution, current_solution.Solution)
	copy(neighbor.Hubs, current_solution.Hubs)

	random_node, random_hub := selectRandomNodeAndHub(neighbor)

	neighbor.Solution[random_node] = random_hub

	return neighbor, random_node
}

func TabuSearch(initial_solution Candidate, cost_matrix, flow_matrix [][]float64, tabuSize, maxCandidates, iterations int, alpha float64) (best Candidate) {
	current := initial_solution
	best = current

	tabuList := make([]int, tabuSize)

	for i := 0; i < iterations; i++ {

		if i-best.Iteration > 10000 {
			break
		}

		var candidates []Candidate
		for j := 0; j < maxCandidates; j++ {
			neighbor, swapped_node := generateCandidateTypeA(current)

			neighbor.SwappedNode = swapped_node
			neighbor.calcCost(alpha)
			candidates = append(candidates, neighbor)
		}

		sort.Sort(CandidateVector(candidates))
		bestCandidate := candidates[0]

		if !isInSlice(bestCandidate.SwappedNode, tabuList) || (i-best.Iteration) > aspiration {
			current = bestCandidate
			updateTabuList(bestCandidate.SwappedNode, &tabuList, tabuSize)
			if bestCandidate.Cost < best.Cost {
				best = bestCandidate
				best.Iteration = i
			}
		} else {
			c := 0
			for isInSlice(bestCandidate.SwappedNode, tabuList) && c < len(candidates) {
				bestCandidate = candidates[c]
				c++
			}

			current = bestCandidate
			updateTabuList(bestCandidate.SwappedNode, &tabuList, tabuSize)
			if bestCandidate.Cost < best.Cost {
				best = bestCandidate
				best.Iteration = i
			}
		}

	}

	return best
}

func main() {

	var err error

	data_sets_flow := []string{
		"Flow_matrix10.csv",
		"Flow_matrix15.csv",
		"Flow_matrix20.csv",
		"Flow_matrix25.csv",
		"postal_office_network_flow_25.csv",
		"postal_office_network_flow_55.csv",
	}
	data_sets_cost := []string{
		"Cost_matrix10.csv",
		"Cost_matrix15.csv",
		"Cost_matrix20.csv",
		"Cost_matrix25.csv",
		"postal_office_network_distance_25.csv",
		"postal_office_network_distance_55.csv",
	}
	sizes := []int{
		10,
		15,
		20,
		25,
		25,
		55,
	}

	start := time.Now()

	alphas := []float64{
		0.2,
		0.4,
		0.8,
	}
	hubs := []int{
		3,
		4,
	}

	fmt.Printf("Confirguration: Iterations[%d]\tMax Candidates Multiplier[%d]\tTabu Size Divider[%d]\tAspiration[%d]\n", iterations, maxCandidatesMultiplier, tabuSizeDivider, aspiration)
	fmt.Printf("%-40s\t%-10s\t%-10s\t%-20s\t%-20s\t%-20s\t%-20s\t%-20s\t%-20s\n", "Datset", "No Hubs", "Alpha", "Hub Locations", "TNC", "Avg TNC", "Time Per Run", "Total Time", "Iterations")
	for i, _ := range data_sets_flow {

		// dynamic configurations
		tabuSize := sizes[i] / tabuSizeDivider
		maxCandidates = sizes[i] * maxCandidatesMultiplier

		// read input data
		cost_matrix, err = read_matrix(data_sets_cost[i], sizes[i])
		if err != nil {
			fmt.Printf("Error", err.Error())
			return
		}

		flow_matrix, err = read_matrix(data_sets_flow[i], sizes[i])
		if err != nil {
			fmt.Printf("Error", err.Error())
			return
		}

		// calc total flow
		total_flow = calcTotalFlow(flow_matrix)

		for _, hub := range hubs {
			no_hubs = hub
			for _, alfa := range alphas {

				// update the global variable used
				alpha = alfa
				var best []Candidate

				fmt.Printf("%-40s\t%-10d\t%-10f\t", data_sets_cost[i], no_hubs, alpha)
				primary_start_time := time.Now()
				for k := 0; k < no_routines; k++ {
					start = time.Now()
					init_solution := get_initial_solution(cost_matrix, flow_matrix, alpha, no_hubs)
					init_solution.calcCost(alpha)
					c := TabuSearch(init_solution, cost_matrix, flow_matrix, tabuSize, maxCandidates, iterations, alpha)
					elapsed := time.Since(start)
					c.ElapsedTime = elapsed
					best = append(best, c)
				}

				sort.Sort(CandidateVector(best))

				// average TNC
				average_tnc := 0.0
				for _, c := range best {
					average_tnc += c.NormalizedCost
				}
				average_tnc = average_tnc / float64(len(best))
				fmt.Printf("%-v\t%-20f\t%-20f\t%-20s\t%-20s\t%-20d\n", best[0].Hubs, best[0].NormalizedCost, average_tnc, best[0].ElapsedTime, time.Since(primary_start_time), best[0].Iteration)
			}
		}
	}
}
