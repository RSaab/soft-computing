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
	"strings"

	"sync"
)

var cost_matrix = [][]float64{}
var flow_matrix = [][]float64{}
var total_flow float64

var alpha = 0.2
var no_hubs = 3
var no_routines = 4

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

func (c *Candidate) calcCost(alpha float64) {
	c.Cost = calcTotalCost(cost_matrix, flow_matrix, alpha, c.Solution)
	c.NormalizedCost = c.Cost / total_flow
}

func get_initial_solution(cost_matrix, flow_matrix [][]float64, alpha float64, number_of_hubs int) Candidate {
	// use a map to generate this so it can scale
	// but first check if this is really an issue cuz how many hubs will ever allocate at max
	candidate := Candidate{}

	// randomly select certain number of hubs
	rand.Seed(time.Now().UnixNano())
	for len(candidate.Hubs) < number_of_hubs {
		random_number := rand.Intn(len(cost_matrix))
		if !isInSlice(random_number, candidate.Hubs) {
			candidate.Hubs = append(candidate.Hubs, random_number)
		}
	}

	//The optimal solution
	// candidate.Hubs = append(candidate.Hubs, 16)
	// candidate.Hubs = append(candidate.Hubs, 11)
	// candidate.Hubs = append(candidate.Hubs, 3)

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

	// candidate.Cost = calcTotalCost(cost_matrix, flow_matrix, alpha, candidate.Solution)

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

func arrayToString(a []int, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
	//return strings.Trim(strings.Join(strings.Split(fmt.Sprint(a), " "), delim), "[]")
	//return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(a)), delim), "[]")
}

// func hash(s string) int {
// 	h := fnv.New32a()
// 	h.Write([]byte(s))
// 	return h.Sum32()
// }

// func isInTabuList(c Candidate, tabuList []int) bool {

// 	hash := hash(arrayToString(c.Solution, ""))
// 	for _, t := range tabuList {
// 		if hash == t {
// 			return true
// 		}
// 	}

// 	return false
// }

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

func generateCandidate(sol, hubs []int, current Candidate, tabuList *[]int, tabuSize int, cost_matrix, flow_matrix [][]float64, alpha float64) Candidate {

	/**
	 *  swap a random node from one hub to another
	 * 	in a smart way:
	 *		find the next nearest hub
	 */

	best := Candidate{}

	best.Solution = make([]int, len(sol))
	best.Hubs = make([]int, len(hubs))
	copy(best.Solution, sol)
	copy(best.Hubs, hubs)

	found_tabu := true
	var selected_node, selected_hub int
	for found_tabu {
		selected_node, selected_hub = selectRandomNodeAndHub(best)
		if !isInTabuList(selected_node, *tabuList) {
			found_tabu = false
		}
	}

	best.Solution[selected_node] = selected_hub
	best.Cost = calcTotalCost(cost_matrix, flow_matrix, alpha, best.Solution)

	return best

	// randomly switch

	// candidate := Candidate{}

	// candidate.Solution = make([]int, len(best.Solution))
	// candidate.Hubs = make([]int, len(best.Hubs))
	// copy(candidate.Solution, best.Solution)
	// copy(candidate.Hubs, best.Hubs)

	// random_node := rand.Intn(len(candidate.Solution))
	// for isInSlice(random_node, candidate.Hubs) {
	// 	random_node = rand.Intn(len(candidate.Solution))
	// }

	// random_hub := rand.Intn(len(candidate.Hubs))

	// candidate.Solution[random_node] = candidate.Hubs[random_hub]

	// candidate.Cost = calcTotalCost(cost_matrix, flow_matrix, alpha, candidate.Solution)

	// fmt.Printf("%+v\n", candidate.Solution)

	// return candidate

	// return get_initial_solution(cost_matrix, flow_matrix, alpha, 3)

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

/**
If i am randomly selecting a node to swap then it is possible to miss one
of the neighboring solutions

IS THIS CORRECT??
*/
func TabuSearch(initial_solution Candidate, cost_matrix, flow_matrix [][]float64, tabuSize, maxCandidates, iterations int, alpha float64) (best Candidate) {
	current := initial_solution
	best = current

	tabuList := make([]int, tabuSize)

	// current_best_solution_iteration := 0
	for i := 0; i < iterations; i++ {
		// if i-current_best_solution_iteration > 10000 {
		// 	fmt.Printf("Aspiration condition met. Stopping...")
		// 	break
		// }

		if i%10000 == 0 {
			// fmt.Printf("Iteration: %d\n", i)
			// fmt.Printf(current.Hubs)
			fmt.Fprintf(os.Stderr, "\rIteration: %d", i)

		}
		var candidates []Candidate
		for j := 0; j < maxCandidates; j++ {
			// candidates[j] = generateCandidate(current.Solution, current.Hubs, current, &tabuList, tabuSize, cost_matrix, flow_matrix, alpha)
			neighbor, swapped_node := generateCandidateTypeA(current)
			// neighbor, swapped_node := generateCandidateTypeB(current)
			// neighbor, swapped_node := generateCandidateTypeC(current)
			if isInSlice(swapped_node, tabuList) {
				// fmt.Printf("solution is tabu")
				// fmt.Printf("swapped node %d - ", swapped_node)
				// fmt.Printf("tabu list: %+v\n", tabuList)
				continue
			}

			neighbor.SwappedNode = swapped_node
			neighbor.calcCost(alpha)
			candidates = append(candidates, neighbor)
		}

		sort.Sort(CandidateVector(candidates))

		if len(candidates) <= 0 {
			fmt.Printf("No candidates generated!\n")
			continue
		}

		bestCandidate := candidates[0]
		updateTabuList(bestCandidate.SwappedNode, &tabuList, tabuSize)

		if bestCandidate.Cost < current.Cost {
			current = bestCandidate
			if bestCandidate.Cost < best.Cost {
				fmt.Printf("found better solution!\n")
				// current_best_solution_iteration = i
				best = bestCandidate
				best.Print()
			}
		}

	}

	return best
}

/*
Tabu Solution: best so far for postal office network

Setup:
iterations := 1000000
tabuSize := 20
maxCandidates := 35
alpha := 0.2
no_hubs := 3

Elapsed Time 13m52.429978999s
Hubs: [3 11 54]
Nodes:  	1 	2 	3 	4 	5 	6 	7 	8 	9 	10	11	12	13	14	15	16	17	18	19	20	21	22	23	24	25	26	27	28	29	30	31	32	33	34	35	36	37	38	39	40	41	42	43	44	45	46	47	48	49	50	51	52	53	54	55
Solution:	55	55	55	4 	12	4 	12	12	55	55	12	12	12	4 	55	12	55	12	55	4 	4 	55	55	12	55	12	12	4 	4 	12	4 	12	12	55	55	55	4 	55	4 	12	55	55	4 	4 	4 	4 	4 	55	4 	55	55	12	4 	55	55
Total Cost: 2.5081982534400032e+10
Normalized Cost: 615.4721173882486


Setup:
iterations := 1000 //100
tabuSize := 20  //15
maxCandidates := 35

alpha := 0.2
no_hubs := 3
no_routines := 4
Elapsed Time 823.945195ms
Tabu Solution:
Hubs: [29 44 1]
Nodes:  	1 	2 	3 	4 	5 	6 	7 	8 	9 	10	11	12	13	14	15	16	17	18	19	20	21	22	23	24	25	26	27	28	29	30	31	32	33	34	35	36	37	38	39	40	41	42	43	44	45	46	47	48	49	50	51	52	53	54	55
Solution:	45	2 	2 	45	45	2 	30	30	2 	2 	45	30	30	45	2 	30	2 	30	2 	45	45	2 	2 	45	45	30	30	45	45	30	45	30	30	2 	2 	2 	45	2 	30	30	2 	2 	45	30	45	2 	45	2 	30	2 	2 	30	30	2 	2
Total Cost: 2.4847582590400024e+10
Normalized Cost: 609.7203140907418

*/

func main() {

	var err error

	iterations := 100000 //100
	tabuSize := 10       //15
	maxCandidates := 60

	// cost_matrix, err = read_matrix("postal_office_network_distance_55.csv", 55)
	cost_matrix, err = read_matrix("./Cost_matrix25.csv", 25)
	if err != nil {
		fmt.Printf("Error", err.Error())
		return
	}

	// flow_matrix, err = read_matrix("postal_office_network_flow_55.csv", 55)
	flow_matrix, err = read_matrix("./Flow_matrix25.csv", 25)
	if err != nil {
		fmt.Printf("Error", err.Error())
		return
	}

	total_flow = calcTotalFlow(flow_matrix)
	fmt.Printf("Total Flow %+v\n", total_flow)

	// felipes_solution := []int{3, 16, 16, 3, 3, 3, 3, 3, 3, 3, 3, 11, 3, 16, 3, 3, 16, 16, 11, 16, 3, 11, 11, 3, 16}
	// fmt.Printf("Felipe's code: %+v\n", calcTotalCost(cost_matrix, flow_matrix, alpha, felipes_solution))
	// os.Exit(0)

	start := time.Now()
	init_solution := get_initial_solution(cost_matrix, flow_matrix, alpha, no_hubs)

	init_solution.calcCost(alpha)
	elapsed := time.Since(start)
	fmt.Printf("Elapsed Time %s\n", elapsed)

	init_solution.Print()

	start = time.Now()
	// best := TabuSearch(init_solution, cost_matrix, flow_matrix, tabuSize, maxCandidates, iterations, alpha)
	// elapsed = time.Since(start)

	var wg sync.WaitGroup

	var best []Candidate

	for i := 0; i < no_routines; i++ {
		wg.Add(1)
		go func() {
			start = time.Now()
			init_solution = get_initial_solution(cost_matrix, flow_matrix, alpha, no_hubs)
			init_solution.calcCost(alpha)
			c := TabuSearch(init_solution, cost_matrix, flow_matrix, tabuSize, maxCandidates, iterations, alpha)
			elapsed = time.Since(start)
			fmt.Printf("Elapsed Time %s\n", elapsed)
			best = append(best, c)
			wg.Done()
		}()

	}

	wg.Wait()

	sort.Sort(CandidateVector(best))

	fmt.Printf("\n")
	fmt.Printf("Elapsed Time %s\n", elapsed)

	fmt.Printf("\n\nTabu Search Solution:\n")

	best[0].Print()

	fmt.Printf("\n\nSimulate Annealing\n")
	SA()

}
