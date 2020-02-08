package main

import (
	// "bytes"
	"fmt"
	"math/rand"
	"time"

	"bufio"
	"encoding/csv"
	"io"
	"os"
	"strconv"
)

var cost_matrix = [][]float64{}
var flow_matrix = [][]float64{}
var total_flow float64
var alpha = 0.2
var no_hubs = 3
var no_routines = 4

// MutationRate is the rate of mutation
var MutationRate = 0.01

// PopSize is the size of the population
var PopSize = 5000

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

func calcTotalFlow(flow_matrix [][]float64) float64 {
	total_flow := 0.0
	for _, c := range flow_matrix {
		for _, e := range c {
			total_flow += e
		}
	}
	return total_flow
}

func isInSlice(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func main() {
	var err error

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

	start := time.Now()
	rand.Seed(time.Now().UTC().UnixNano())

	// target := []byte("To be or not to be")
	population := createPopulation(cost_matrix, flow_matrix, alpha, no_hubs)

	generation := 0
	bestOrganism := Organism{}
	for i := 0; i < 300; i++ {
		generation++
		bestOrganism = getBest(population)

		// if bestOrganism.Fitness >= 0.00130 {
		// 	break
		// }
		fmt.Printf("\r generation: %d | fitness: %2f | Normalized Cost %2f\n", generation, bestOrganism.Fitness, bestOrganism.DNA.Cost/total_flow)
		// bestOrganism.DNA.Print()

		maxFitness := bestOrganism.Fitness
		pool := createPool(population, maxFitness)
		population = naturalSelection(pool, population)

		elapsed := time.Since(start)
		fmt.Printf("\nTime taken: %s\n", elapsed)
	}
	// fmt.Printf("\r generation: %d | fitness: %2f | Normalized Cost %2f\n", generation, bestOrganism.Fitness, bestOrganism.DNA.Cost/total_flow)

}

//DNA
type SolutionDNA struct {
	Solution []int
	Hubs     []int
	Cost     float64 // total cost
}

func (c SolutionDNA) Print() {
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
	fmt.Printf("Normalized Cost: %+v\n", c.Cost/total_flow)

}

// Organism for this genetic algorithm
type Organism struct {
	DNA     *SolutionDNA
	Fitness float64 // normalized cost
}

// creates a Organism
func createOrganism(cost_matrix, flow_matrix [][]float64, alpha float64, number_of_hubs int) (organism Organism) {

	organism = Organism{}
	organism.DNA = &SolutionDNA{}
	organism.DNA.Hubs = make([]int, number_of_hubs)
	organism.DNA.Solution = make([]int, len(cost_matrix))

	// randomly select certain number of hubs
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < number_of_hubs; i++ {
		random_number := rand.Intn(len(cost_matrix))
		if !isInSlice(random_number, organism.DNA.Hubs) {
			organism.DNA.Hubs[i] = random_number
		}
	}

	// allocate nodes to their nearest organism.DNA.Hubs
	for i, _ := range cost_matrix {
		target_hub := organism.DNA.Hubs[0]
		for _, hub := range organism.DNA.Hubs {
			if cost_matrix[i][hub] < cost_matrix[i][target_hub] {
				target_hub = hub
			}
		}
		organism.DNA.Solution[i] = target_hub
	}

	organism.calcFitness()

	return organism
}

// creates the initial population
func createPopulation(cost_matrix, flow_matrix [][]float64, alpha float64, number_of_hubs int) (population []Organism) {
	population = make([]Organism, PopSize)
	for i := 0; i < PopSize; i++ {
		population[i] = createOrganism(cost_matrix, flow_matrix, alpha, number_of_hubs)
	}
	return
}

// calculates the fitness of the Organism
func (d *Organism) calcFitness() {
	var total_cost float64

	solution := d.DNA.Solution

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

	d.Fitness = (total_cost / total_flow)
	d.DNA.Cost = total_cost
	return
}

// create the breeding pool that creates the next generation
func createPool(population []Organism, maxFitness float64) (pool []Organism) {
	pool = make([]Organism, 0)
	// create a pool for next generation
	for i := 0; i < len(population); i++ {
		population[i].calcFitness()
		num := int((population[i].Fitness / maxFitness) * 100)
		for n := 0; n < num; n++ {
			pool = append(pool, population[i])
		}
	}
	return
}

// perform natural selection to create the next generation
func naturalSelection(pool []Organism, population []Organism) []Organism {
	next := make([]Organism, len(population))
	for i := 0; i < len(population); i++ {
		r1, r2 := rand.Intn(len(pool)), rand.Intn(len(pool))
		a := pool[r1]
		b := pool[r2]

		child := crossover(a, b)
		child.mutate()
		child.calcFitness()

		next[i] = child
	}
	return next
}

// crosses over 2 Organisms
func crossover(d1 Organism, d2 Organism) Organism {
	dna := SolutionDNA{}

	dna.Hubs = make([]int, len(d1.DNA.Hubs))
	dna.Solution = make([]int, len(d1.DNA.Solution))

	child := Organism{
		DNA:     &dna,
		Fitness: 0,
	}
	mid := rand.Intn(len(d1.DNA.Hubs))
	for i := 0; i < len(d1.DNA.Hubs); i++ {
		if i > mid {
			child.DNA.Hubs[i] = d1.DNA.Hubs[i]
		} else {
			child.DNA.Hubs[i] = d2.DNA.Hubs[i]
		}
	}

	// allocate nodes to their nearest organism.DNA.Hubs
	for i, _ := range cost_matrix {
		target_hub := child.DNA.Hubs[0]
		for _, hub := range child.DNA.Hubs {
			if cost_matrix[i][hub] < cost_matrix[i][target_hub] {
				target_hub = hub
			}
		}
		child.DNA.Solution[i] = target_hub
	}

	return child
}

// mutate the Organism
func (d *Organism) mutate() {
	for i := 0; i < len(d.DNA.Solution); i++ {
		if rand.Float64() < MutationRate {
			d.DNA.Solution[i] = d.DNA.Hubs[rand.Intn(len(d.DNA.Hubs))]
		}
	}
}

// Get the best organism
func getBest(population []Organism) Organism {
	best := 0.0
	index := 0
	for i := 0; i < len(population); i++ {
		if population[i].Fitness < best {
			index = i
			best = population[i].Fitness
		}
	}
	return population[index]
}
