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

/*
Configuration structure that holds all configration needed to run any of the commands. Sort of a wrapper on top of the various other configuration structs
so we don't have to pass those individually to each command function
*/
type runningConf struct {
	listPublicIps bool
	command       string

	conf *Config

	api *digitalOceanAPI
}

func main() {
	var forceUpdate bool
	var listPublicIps bool

	flag.BoolVar(&forceUpdate, "force-update", false, "Force update the cache file")
	flag.BoolVar(&listPublicIps, "list-public-ip", false, "If the list command is used, only print out list of public IPs")
	flag.Parse()

	nArgs := flag.NArg()
	if nArgs < 1 {
		fmt.Println("Not enough arguments")
		printUsage()
		return
	}

	command := flag.Arg(0)

	// We treat the config command differently because it is the only command that doesn't use the API
	if command == "config" {
		if err := createConfig(); err != nil {
			fmt.Printf("Error while creating config. Error: %s\n", err.Error())
		}

		return
	}

	userConf, err := getConfig()
	if err != nil {
		fmt.Printf("Unable to read configuration. Error: %s\n", err.Error())
		return
	}

	api, err := newApiFromConfig(userConf)
	if err != nil {
		fmt.Printf("Error while creating API connector. Error: %s\n", err.Error())
		return
	}

	if forceUpdate {
		if err = api.forceUpdateCache(); err != nil {
			fmt.Printf("Error trying to force update cache. Error: %s\n", err.Error())
			return
		}
	}

	runningConf := runningConf{listPublicIps, command, userConf, api}

	switch command {
	case "c", "connect":
		connectToDroplet(runningConf)
	case "l", "list":
		filterExpressions := make([]string, 0)
		if nArgs > 1 {
			filterExpressions = flag.Args()[1:]
		}
		listDropletsInfo(runningConf, filterExpressions)
	case "r", "run":
		var filterExpression, cmd string

		if nArgs != 3 {
			fmt.Println("Not enough arguments for the run command")
			break
		}

		filterExpression = flag.Arg(1)
		cmd = flag.Arg(2)
		runCommandOnDroplets(runningConf, filterExpression, cmd)
	case "completion":
		printCompletions(runningConf)
	default:
		fmt.Println("Unknown command")
	}
}

func printUsage() {
	fmt.Print(`Usage:
	dropletconn <COMMAND> <ARGS>.
	Available commands:
		- config
			Create a new configuration file by prompting for configuration options
		- connect/c <DROPLET NAME>
			Connect to the named droplet
		- list/l [FILTER EXPRESSION]...
			List all droplets that match ANY of the given filter expressions. Matching means that the filter expression is a substring of the droplet name
		- run/r <FILTER EXPRESSION> CMD
			Run the given command on all matched droplets one by one`)
}

func connectToDroplet(rConf runningConf) {
	nArgs := flag.NArg()
	if nArgs < 2 {
		fmt.Println("No droplet name given")
		return
	}

	inputDropletName := flag.Arg(1)
	var extraCmdOptions []string
	if nArgs > 2 {
		extraCmdOptions = flag.Args()[2:]
	}

	droplets, err := rConf.api.getDroplets()

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

	cmdOptions := []string{firstPublicIpAddress}
	if config.DefaultUser != "" {
		cmdOptions = append(cmdOptions, "-l")
		cmdOptions = append(cmdOptions, rConf.conf.DefaultUser)
	}
	if rConf.conf.DefaultKeyFileName != "" {
		keyFileName := rConf.conf.DefaultKeyFileName
		keyFilePath, err := getAbsoluteFilePath(keyFileName)
		if err != nil {
			fmt.Printf("Unable to get absolute ssh key file path. Error: %s\n", err.Error())
			return
		}

		cmdOptions = append(cmdOptions, "-i")
		cmdOptions = append(cmdOptions, keyFilePath)
	}

	for _, extraOpt := range extraCmdOptions {
		cmdOptions = append(cmdOptions, extraOpt)
	}

	fmt.Printf("Connecting to \"%s\"\n", inputDropletName)
	cmd := exec.Command("ssh", cmdOptions...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error while trying to run ssh. Error: %s\n", err.Error())
	}
}

func listDropletsInfo(rConf runningConf, filterExpresions []string) {
	droplets := rConf.api.getFilteredDroplets(rConf, filterExpresions)

	// Only list public Ips
	if rConf.listPublicIps {
		for _, di := range droplets {
			netAdd := di.getInterfaceAddresses()
			publicIpAddressesString := strings.Join(netAdd.publicIps, ", ")

			fmt.Println(publicIpAddressesString)
		}
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Id", "Name", "Public IP", "Private IP"})

		for _, di := range droplets {
			netAdd := di.getInterfaceAddresses()
			publicIpAddressesString := strings.Join(netAdd.publicIps, ", ")
			privateIpAddressesString := strings.Join(netAdd.privateIps, ", ")

			table.Append([]string{strconv.Itoa(di.Id), di.Name, publicIpAddressesString, privateIpAddressesString})
		}

		table.Render()
	}
	fmt.Printf("Total droplets: %d\n", len(droplets))
}

func runCommandOnDroplets(rConf runningConf, filterExpression string, command string) {
	droplets := rConf.api.getFilteredDroplets(rConf, []string{filterExpression})

	fmt.Printf("'%s' will be run on:\n", command)
	for _, di := range droplets {
		fmt.Println(di.Name)
	}
}

func printCompletions(rConf runningConf) {
	if flag.NArg() != 2 {
		fmt.Println("Not enough arguments")
		return
	}

	completionFilter := flag.Arg(1)
	droplets, err := rConf.api.getDroplets()
	if err != nil {
		return
	}

	compFilterLower := strings.ToLower(completionFilter)
	for _, di := range droplets {
		dropletName := di.Name
		dropletNameLower := strings.ToLower(dropletName)

		if strings.Contains(dropletNameLower, compFilterLower) {
			fmt.Println(dropletName)
		}
	}
}
