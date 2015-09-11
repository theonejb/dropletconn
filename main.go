package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func main() {
	flag.Parse()
	command := flag.Arg(0)
	if flag.NArg() < 1 {
		fmt.Println("Not enough arguments")
		return
	}

	switch command {
	case "connect":
		connectToDroplet()
	case "list":
		listDropletsInfo()
	default:
	}
}

func connectToDroplet() {
	inputDropletName := flag.Arg(1)
	droplets, err := getDropletsFromApi()

	if err != nil {
		fmt.Printf("Unable to get droplets. Error: %s\n", err.Error())
		return
	} else if len(droplets) == 0 {
		fmt.Println("No droplets found in account")
		return
	}

	var matchedDropletInfo *DropletInfo
	inputDropletNameLower := strings.ToLower(inputDropletName)
	for _, dropletInfo := range droplets {
		dropletNameLower := strings.ToLower(dropletInfo.Name)
		if inputDropletNameLower == dropletNameLower {
			matchedDropletInfo = &dropletInfo
			break
		}
	}

	if matchedDropletInfo == nil {
		fmt.Printf("No match found for \"%s\"\n", inputDropletName)
		return
	}

	firstPublicIpAddress := ""
	for _, interfaceInfo := range matchedDropletInfo.Networks.V4 {
		if interfaceInfo.Type == "public" {
			firstPublicIpAddress = interfaceInfo.IpAddress
			break
		}

	}
	if firstPublicIpAddress == "" {
		fmt.Printf("No public IP address found for droplet \"%s\"\n", matchedDropletInfo.Name)
		return
	}

	fmt.Printf("Connecting to \"%s\"\n", inputDropletName)
	cmd := exec.Command("ssh", firstPublicIpAddress, "-l", "root", "-i", "/Users/asadjb/.ssh/jrd")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error while trying to run ssh. Error: %s\n", err.Error())
	}
}

func listDropletsInfo() {
	dropletsInfo, err := getDropletsFromApi()
	if err != nil {
		fmt.Printf("Unable to get droplets. Error: %s\n", err.Error())
		return
	} else if len(dropletsInfo) == 0 {
		fmt.Println("No droplets found in account")
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Name", "Public IP", "Private IP"})

	for _, di := range dropletsInfo {
		netAdd := di.getInterfaceAddresses()
		table.Append([]string{strconv.Itoa(di.Id), di.Name, strings.Join(netAdd.publicIps, ", "), strings.Join(netAdd.privateIps, ", ")})
	}

	table.Render()
}
