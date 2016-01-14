package main

import (
	"fmt"
	"strings"
)

func getFilteredDroplets(filterExpressions []string, runningConf CommandConfig) []DropletInfo {
	fmt.Printf("Filtering on: %#v\n", filterExpressions)
	dropletsInfo, err := getDropletsFromApi(runningConf.forceUpdate)
	if err != nil {
		fmt.Printf("Unable to get droplets. Error: %s\n", err.Error())
		return nil
	} else if len(dropletsInfo) == 0 {
		fmt.Println("No droplets found in account")
		return nil
	}

	// Convert all filter expression string to lower
	for i := range filterExpressions {
		filterExpressions[i] = strings.ToLower(filterExpressions[i])
	}

	matchedDroplets := make([]DropletInfo, 0)
	for _, di := range dropletsInfo {
		if di.matchesFilterExpressions(filterExpressions) {
			matchedDroplets = append(matchedDroplets, di)
		}
	}

	return matchedDroplets
}
