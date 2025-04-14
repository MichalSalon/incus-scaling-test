package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
)

func main() {
	http.HandleFunc("/cpuinfo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		cpus, err := getCpuIds()
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting CPUs: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		_, _ = fmt.Fprint(w, strings.Join(cpus, ","))
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func getCpuIds() ([]string, error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return nil, fmt.Errorf("failed to open /proc/stat: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	cores := make([]string, 0, 64) // can't use runtime.NumCPU()
	for scanner.Scan() {
		cols := strings.Fields(scanner.Text())
		if len(cols) == 0 || cols[0] == "cpu" || !strings.HasPrefix(cols[0], "cpu") {
			continue
		}
		cores = append(cores, strings.TrimPrefix(cols[0], "cpu"))
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading /proc/stat: %w", err)
	}

	slices.SortFunc(cores, func(a, b string) int {
		intA, _ := strconv.Atoi(a)
		intB, _ := strconv.Atoi(b)
		return intA - intB
	})
	return cores, nil
}
