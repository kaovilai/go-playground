package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func main() {
	// Generate timestamp for filenames
	timestamp := time.Now().Format("20060102-150405")
	
	// Create results directory if it doesn't exist
	resultsDir := "results"
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating results directory: %v\n", err)
		os.Exit(1)
	}
	
	// Create output file
	outputFile := filepath.Join(resultsDir, fmt.Sprintf("output-%s.txt", timestamp))
	outF, err := os.Create(outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outF.Close()
	
	// Create error file (will be used if there are errors)
	errorFile := filepath.Join(resultsDir, fmt.Sprintf("error-%s.txt", timestamp))
	errF, err := os.Create(errorFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating error file: %v\n", err)
		os.Exit(1)
	}
	defer errF.Close()
	
	// Write output to both stdout and output file
	output1 := "Hello from Go Playground!"
	output2 := "This is a simple Go program that runs to completion."
	
	fmt.Println(output1)
	fmt.Println(output2)
	
	fmt.Fprintln(outF, output1)
	fmt.Fprintln(outF, output2)
	
	// Error file is created but remains empty since this program runs successfully
	// If there were errors, they would be written to both stderr and the error file:
	// fmt.Fprintf(os.Stderr, "Error message\n")
	// fmt.Fprintf(errF, "Error message\n")
	
	// Exit cleanly
	os.Exit(0)
}