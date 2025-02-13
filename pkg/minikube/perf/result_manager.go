/*
Copyright 2020 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package perf

import (
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
)

type resultManager struct {
	results map[*Binary]resultWrapper
}

func newResultManager() *resultManager {
	return &resultManager{
		results: map[*Binary]resultWrapper{},
	}
}

func (rm *resultManager) addResult(binary *Binary, test string, r result) {
	a, ok := rm.results[binary]
	if !ok {
		r := map[string][]*result{test: {&r}}
		rm.results[binary] = resultWrapper{r}
		return
	}
	b, ok := a.results[test]
	if !ok {
		a.results[test] = []*result{&r}
		return
	}
	a.results[test] = append(b, &r)
}

func (rm *resultManager) totalTimes(binary *Binary, t string) []float64 {
	var totals []float64
	results, ok := rm.results[binary].results[t]
	if !ok {
		return nil
	}
	for _, r := range results {
		total := 0.0
		for _, t := range r.timedLogs {
			total += t
		}
		totals = append(totals, total)
	}
	return totals
}

func (rm *resultManager) summarizeResults(binaries []*Binary) {
	// print total and average times
	for _, b := range binaries {
		for t := range rm.results[b].results {
			fmt.Printf("Times for %s %s: ", b.Name(), t)
			totalTimes := rm.totalTimes(b, t)
			for _, tt := range totalTimes {
				fmt.Printf("%.1fs ", tt)
			}
			fmt.Println()
			fmt.Printf("Average time for %s %s: %.1fs\n\n", b.Name(), t, average(totalTimes))
		}
	}

	// print out summary per log
	rm.summarizeTimesPerLog(binaries)
}

func (rm *resultManager) summarizeTimesPerLog(binaries []*Binary) {
	results := rm.results[binaries[0]]
	logs := results.results["start"][0].logs

	table := make([][]string, len(logs))
	for i := range table {
		table[i] = make([]string, len(binaries)+1)
	}

	for i, l := range logs {
		table[i][0] = l
	}

	for i, b := range binaries {
		results := rm.results[b]
		averageTimeForLog := averageTimePerLog(results.results["start"])
		for log, time := range averageTimeForLog {
			index := indexForLog(logs, log)
			if index == -1 {
				continue
			}
			table[index][i+1] = fmt.Sprintf("%.1fs", time)
		}
	}

	t := tablewriter.NewWriter(os.Stdout)
	t.SetHeader([]string{"Log", binaries[0].Name(), binaries[1].Name()})

	for _, v := range table {
		t.Append(v)
	}
	fmt.Println("Averages Time Per Log")
	fmt.Println("<details>")
	fmt.Println()
	fmt.Println("```")
	t.Render() // Send output
	fmt.Println("```")
	fmt.Println()
	fmt.Println("</details>")
}

func indexForLog(logs []string, log string) int {
	for i, l := range logs {
		if strings.Contains(log, l) {
			return i
		}
	}
	return -1
}

func averageTimePerLog(results []*result) map[string]float64 {
	collection := map[string][]float64{}
	for _, r := range results {
		for log, time := range r.timedLogs {
			if _, ok := collection[log]; !ok {
				collection[log] = []float64{time}
			} else {
				collection[log] = append(collection[log], time)
			}
		}
	}
	avgs := map[string]float64{}
	for log, times := range collection {
		avgs[log] = average(times)
	}
	return avgs
}
