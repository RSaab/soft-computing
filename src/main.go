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
	Solution []int
	Cost     float64
	Hubs     []int
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

	candidate.Cost = calcTotalCost(cost_matrix, flow_matrix, alpha, candidate.Solution)

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

	if len(*tabuList) == tabuSize {
		*tabuList = (*tabuList)[1:]
	}
	*tabuList = append(*tabuList, selected_node)

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

func TabuSearch(initial_solution Candidate, cost_matrix, flow_matrix [][]float64, tabuSize, maxCandidates, iterations int, alpha float64) (best Candidate) {
	current := initial_solution
	best = current

	tabuList := make([]int, tabuSize)

	for i := 0; i < iterations; i++ {
		if i%10000 == 0 {
			fmt.Printf("Iteration: %d\n", i)
			current = get_initial_solution(cost_matrix, flow_matrix, alpha, 3)
			fmt.Println(current.Hubs)
			// fmt.Printf("Tabu list: %+v\n", tabuList)
		}
		candidates := make([]Candidate, maxCandidates)
		for j, _ := range candidates {
			candidates[j] = generateCandidate(current.Solution, current.Hubs, current, &tabuList, tabuSize, cost_matrix, flow_matrix, alpha)
		}

		sort.Sort(CandidateVector(candidates))
		bestCandidate := candidates[0]
		if bestCandidate.Cost < current.Cost {
			current = bestCandidate
			if bestCandidate.Cost < best.Cost {
				fmt.Println("found better solution!")
				best = bestCandidate
				best.Print()
				// // HASH the solution and add to tabu list
				// hash := hash(arrayToString(best.Solution, ""))
				// if len(tabuList) == tabuSize {
				// 	tabuList = tabuList[1:]
				// }
				// tabuList = append(tabuList, hash)

			}
		}

	}

	return best
}

func main() {

	var err error

	alpha := 0.2

	// cost_matrix, err = read_matrix("postal_office_network_distance_55.csv", 55)
	cost_matrix, err = read_matrix("./Cost_matrix25.csv", 25)
	if err != nil {
		fmt.Println("Error", err.Error())
		return
	}

	// flow_matrix, err = read_matrix("postal_office_network_flow_55.csv", 55)
	flow_matrix, err = read_matrix("./Flow_matrix25.csv", 25)
	if err != nil {
		fmt.Println("Error", err.Error())
		return
	}

	total_flow := calcTotalFlow(flow_matrix)
	fmt.Printf("Total Flow %+v\n", total_flow)

	// felipes_solution := []int{3, 16, 16, 3, 3, 3, 3, 3, 3, 3, 3, 11, 3, 16, 3, 3, 16, 16, 11, 16, 3, 11, 11, 3, 16}
	// fmt.Printf("Felipe's code: %+v\n", calcTotalCost(cost_matrix, flow_matrix, alpha, felipes_solution))
	// os.Exit(0)

	start := time.Now()
	init_solution := get_initial_solution(cost_matrix, flow_matrix, alpha, 3)
	elapsed := time.Since(start)
	fmt.Printf("Elapsed Time %s\n", elapsed)

	init_solution.Print()
	fmt.Println(init_solution.Solution)

	iterations := 1000000 //100
	tabuSize := 20        //15
	maxCandidates := 30
	start = time.Now()
	// best := TabuSearch(init_solution, cost_matrix, flow_matrix, tabuSize, maxCandidates, iterations, alpha)
	elapsed = time.Since(start)

	var wg sync.WaitGroup
	wg.Add(4)
	var best1 Candidate
	var best2 Candidate
	var best3 Candidate
	var best4 Candidate

	go func() {
		start = time.Now()
		best1 = TabuSearch(init_solution, cost_matrix, flow_matrix, tabuSize, maxCandidates, iterations, alpha)
		elapsed = time.Since(start)
		fmt.Printf("Elapsed Time %s\n", elapsed)
		wg.Done()
	}()

	go func() {
		start = time.Now()
		best2 = TabuSearch(init_solution, cost_matrix, flow_matrix, tabuSize, maxCandidates, iterations, alpha)
		elapsed = time.Since(start)
		fmt.Printf("Elapsed Time %s\n", elapsed)
		wg.Done()
	}()

	go func() {
		start = time.Now()
		best3 = TabuSearch(init_solution, cost_matrix, flow_matrix, tabuSize, maxCandidates, iterations, alpha)
		elapsed = time.Since(start)
		fmt.Printf("Elapsed Time %s\n", elapsed)
		wg.Done()
	}()

	go func() {
		start = time.Now()
		best4 = TabuSearch(init_solution, cost_matrix, flow_matrix, tabuSize, maxCandidates, iterations, alpha)
		elapsed = time.Since(start)
		fmt.Printf("Elapsed Time %s\n", elapsed)
		wg.Done()
	}()
	wg.Wait()

	best := []Candidate{best1, best2, best3, best4}
	sort.Sort(CandidateVector(best))

	fmt.Println()
	fmt.Println("Tabu Solution:")
	fmt.Printf("Elapsed Time %s\n", elapsed)

	best[0].Print()
	fmt.Printf("Normalized Cost: %+v\n", best[0].Cost/total_flow)

}
