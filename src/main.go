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

	"hash/fnv"
	"sort"
	"strings"
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
	// candidate.Hubs = append(candidate.Hubs, 7)
	// candidate.Hubs = append(candidate.Hubs, 15)
	// candidate.Hubs = append(candidate.Hubs, 5)

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

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func isInTabuList(c Candidate, tabuList []uint32) bool {

	hash := hash(arrayToString(c.Solution, ""))
	for _, t := range tabuList {
		if hash == t {
			return true
		}
	}

	return false
}

func generateCandidate(best Candidate, tabuList []uint32, cost_matrix, flow_matrix [][]float64, alpha float64) Candidate {
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
	return get_initial_solution(cost_matrix, flow_matrix, alpha, 3)

	// return candidate

}

func TabuSearch(initial_solution Candidate, cost_matrix, flow_matrix [][]float64, tabuSize, maxCandidates, iterations int, alpha float64) (best Candidate) {
	current := initial_solution
	best = current

	tabuList := make([]uint32, tabuSize)

	for i := 0; i < iterations; i++ {
		if i%10000 == 0 {
			fmt.Printf("Iteration: %d\n", i)
		}
		candidates := make([]Candidate, maxCandidates)
		for j, _ := range candidates {
			candidates[j] = generateCandidate(current, tabuList, cost_matrix, flow_matrix, alpha)
			found_tabu := true

			for found_tabu {
				if !isInTabuList(candidates[j], tabuList) {
					found_tabu = false
				} else {
					candidates[j] = generateCandidate(current, tabuList, cost_matrix, flow_matrix, alpha)
				}
			}
			found_tabu = true
		}

		sort.Sort(CandidateVector(candidates))
		bestCandidate := candidates[0]
		if bestCandidate.Cost < current.Cost {
			current = bestCandidate
			if bestCandidate.Cost < best.Cost {
				fmt.Println("found better solution!")
				best = bestCandidate
				best.Print()
				// HASH the solution and add to tabu list
				hash := hash(arrayToString(best.Solution, ""))
				if len(tabuList) == tabuSize {
					tabuList = tabuList[1:]
				}
				tabuList = append(tabuList, hash)

			}
		}

	}

	return best
}

func main() {

	var err error

	alpha := 0.8

	cost_matrix, err = read_matrix("postal_office_network_distance_55.csv", 55)
	// cost_matrix, err = read_matrix("./Cost_matrix10.csv", 10)
	if err != nil {
		fmt.Println("Error", err.Error())
		return
	}

	flow_matrix, err = read_matrix("postal_office_network_flow_55.csv", 55)
	// flow_matrix, err = read_matrix("./Flow_matrix10.csv", 10)
	if err != nil {
		fmt.Println("Error", err.Error())
		return
	}

	total_flow := calcTotalFlow(flow_matrix)
	fmt.Printf("Total Flow %+v", total_flow)

	// felipes_solution := []int{3, 2, 2, 3, 3, 8, 3, 3, 8, 3, 3, 3, 3, 3, 3, 3, 2, 2, 3, 2, 3, 3, 3, 3, 2}
	// fmt.Printf("Felipe's code: %+v\n", calcTotalCost(cost_matrix, flow_matrix, alpha, felipes_solution))
	// os.Exit(0)

	start := time.Now()
	init_solution := get_initial_solution(cost_matrix, flow_matrix, alpha, 3)
	elapsed := time.Since(start)
	fmt.Printf("Elapsed Time %s\n", elapsed)

	init_solution.Print()

	iterations := 20000 //100
	tabuSize := 50      //15
	maxCandidates := 30
	start = time.Now()
	best := TabuSearch(init_solution, cost_matrix, flow_matrix, tabuSize, maxCandidates, iterations, alpha)
	elapsed = time.Since(start)

	fmt.Println()
	fmt.Println("Tabu Solution:")
	fmt.Printf("Elapsed Time %s\n", elapsed)

	best.Print()
	fmt.Printf("Normalized Cost: %+v\n", best.Cost/total_flow)

	// go func() {
	// 	start = time.Now()
	// 	TabuMain()
	// 	elapsed = time.Since(start)
	// 	fmt.Printf("Elapsed Time %s\n", elapsed)
	// }()

}
