# Overview 
An implementation of the Genetic Algorithm and Tabu Search algorithms in Golang. Adapted to the CAB and TR datasets.

# How to Run
Run the makefile using the `make` command

This will generate three binaries for each algorithm (a binary for each of Windows, Mac OS and your current OS)

Run the compiled binaries from the terminal to run the algorithms

# Report
You can find a detailed report in this repository (report.pdf)


# Sample Outputs:

## Genetic Algorithm
```
Confirguration: Mutataion Rate[0.050]   Population Size[300]    Generations[200]    Aspiration[300]
Datset                                      No Hubs     Alpha       Hub Locations           TNC                     Avg TNC                 Time Per Run            Total Time              Avg Generations     
Cost_matrix10.csv                           3           0.200000    [3 5 6]                 491.934331              492.323457              342.477838ms            3.62031503s             36                  
Cost_matrix10.csv                           3           0.400000    [5 3 6]                 567.912798              567.912798              468.811565ms            4.171716559s            54                  
Cost_matrix10.csv                           3           0.800000    [8 3 6]                 717.397641              719.375314              574.691221ms            4.745320088s            59                  
Cost_matrix10.csv                           4           0.200000    [5 2 6 3]               395.130366              404.618114              686.397325ms            5.001303788s            43                  
Cost_matrix10.csv                           4           0.400000    [6 5 3 7]               493.793763              497.363323              435.591191ms            5.456565509s            72                  
```

## Tabu Search

```
Confirguration: Iterations[10]  Max Candidates Multiplier[5]    Tabu Size Divider[5]    Aspiration[4]
Datset                                      No Hubs     Alpha       Hub Locations   TNC                     Avg TNC                 Time Per Run            Total Time              Iterations          
Cost_matrix10.csv                           3           0.200000    [3 6 2]         495.825585              556.275403              455.019µs               4.953548ms              2                   
Cost_matrix10.csv                           3           0.400000    [5 6 3]         567.912798              677.708844              419.786µs               4.859982ms              0                   
Cost_matrix10.csv                           3           0.800000    [8 6 3]         721.604618              799.661906              949.638µs               5.941901ms              0                   
Cost_matrix10.csv                           4           0.200000    [8 3 2 6]       409.691932              475.798844              367.945µs               3.84893ms               2                   
Cost_matrix10.csv                           4           0.400000    [2 7 6 3]       515.555264              561.904177              378.977µs               4.356536ms              0                   
```
