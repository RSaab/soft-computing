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

func get_initial_solution(cost_matrix, flow_matrix [][]float64, alpha float64, number_of_hubs int) ([]int, []int, float64) {
	var hubs []int
	var solution []int
	total_cost := 0.0

	// randomly select certain number of hubs
	rand.Seed(time.Now().UnixNano())
	for len(hubs) < number_of_hubs {
		random_number := rand.Intn(10)
		if !isInSlice(random_number, hubs) {
			hubs = append(hubs, random_number)
		}
	}

	// allocate nodes to their nearest hubs
	for i, _ := range cost_matrix {
		target_hub := hubs[0]
		for _, hub := range hubs {
			if cost_matrix[i][hub] < cost_matrix[i][target_hub] {
				target_hub = hub
			}
		}
		solution = append(solution, target_hub)
	}

	//  calculate total_cost follow Spoke-Hub-Hub-spoke strategy
	for i, _ := range flow_matrix {
		for j, _ := range flow_matrix {
			collection_cost := flow_matrix[i][j] * cost_matrix[i][solution[i]]
			transportation_cost := flow_matrix[i][j] * cost_matrix[solution[i]][solution[j]] * alpha
			distribution_cost := flow_matrix[i][j] * cost_matrix[solution[j]][j]
			cost := collection_cost + transportation_cost + distribution_cost
			total_cost += cost
		}
	}

	return hubs, solution, total_cost
}

func main() {
	var err error
	cost_matrix, err = read_matrix("./Cost_matrix10.csv", 10)
	if err != nil {
		fmt.Println("Error", err.Error())
		return
	}

	flow_matrix, err = read_matrix("./Flow_matrix10.csv", 10)
	if err != nil {
		fmt.Println("Error", err.Error())
		return
	}

	start := time.Now()
	hubs, solution, totoal_cost := get_initial_solution(cost_matrix, flow_matrix, 0.2, 3)
	elapsed := time.Since(start)
	fmt.Printf("Elapsed Time %s\n", elapsed)

	fmt.Printf("Hubs: %+v\n", hubs)
	fmt.Printf("Solution: %+v\n", solution)
	fmt.Printf("Total Cost: %+v\n", totoal_cost)

}
