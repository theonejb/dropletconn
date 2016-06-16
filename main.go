package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/fatih/color"
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
	api.loadDroplets(forceUpdate)

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
	case "t", "test":
		testSshConnectionToDroplets(runningConf)
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

	dropletName := flag.Arg(1)
	var extraCmdOptions []string
	if nArgs > 2 {
		extraCmdOptions = flag.Args()[2:]
	}

	droplet := rConf.api.getDropletByName(dropletName)
	if droplet == nil {
		fmt.Println("No droplet with maching name found")
		return
	}

	cmdOptions := droplet.getArgsForSsh(rConf.conf)
	if cmdOptions == nil {
		fmt.Println("Unable to create SSH connection")
		return
	}

	for _, extraOpt := range extraCmdOptions {
		cmdOptions = append(cmdOptions, extraOpt)
	}

	fmt.Printf("Connecting to \"%s\"\n", dropletName)
	if err := runSsh(cmdOptions); err != nil {
		fmt.Printf("Error while trying to run ssh. Error: %s\n", err.Error())
	}
}

func listDropletsInfo(rConf runningConf, filterExpresions []string) {
	droplets := rConf.api.getFilteredDroplets(filterExpresions)

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
	droplets := rConf.api.getFilteredDroplets([]string{filterExpression})

	fmt.Printf("'%s' will be run on:\n", command)
	for _, di := range droplets {
		fmt.Printf(" - %s\n", di.Name)
	}

	fmt.Println("Enter 'Y' to continue: ")
	var confirmation string
	if n, err := fmt.Scanln(&confirmation); err != nil || n != 1 {
		fmt.Printf("Error while trying to get confirmation from user. Error: %s\n", err.Error())
		return
	}

	if len(confirmation) != 1 {
		fmt.Println("Only 1 character answers are accepted. Cancelling operation")
		return
	}
	if confirmation[0] != 'Y' && confirmation[0] != 'y' {
		fmt.Println("Cancelling operation")
		return
	}

	for _, droplet := range droplets {
		fmt.Printf("********** %s **********\n", droplet.Name)
		sshArgs := droplet.getArgsForSsh(rConf.conf)
		sshArgs = append(sshArgs, command)
		if err := runSsh(sshArgs); err != nil {
			fmt.Printf("Error running ssh command: Error: %s\n", err.Error())
		}
		fmt.Println("***********************")
	}
}

func testSshConnectionToDroplets(rConf runningConf) {
	droplets, err := rConf.api.getDroplets()
	if err != nil {
		return
	}

	dropletStatus := make([]bool, len(droplets))

	for i, droplet := range droplets {
		fmt.Printf("Running test on %s\n", droplet.Name)

		sshArgs := droplet.getArgsForSsh(rConf.conf)
		sshArgs = append(sshArgs, "-o", "StrictHostKeyChecking=no", "uptime")

		if err := exec.Command("ssh", sshArgs...).Run(); err != nil {
			dropletStatus[i] = false
		} else {
			dropletStatus[i] = true
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Status", "Id", "Name", "Public IP", "Private IP"})

	errorPrintFunc := color.New(color.FgRed).SprintFunc()

	for i, droplet := range droplets {
		var status string
		if dropletStatus[i] {
			status = "\u2713"
		} else {
			status = errorPrintFunc("\u2717")
		}

		netAdd := droplet.getInterfaceAddresses()
		publicIpAddressesString := strings.Join(netAdd.publicIps, ", ")
		privateIpAddressesString := strings.Join(netAdd.privateIps, ", ")

		table.Append([]string{status, strconv.Itoa(droplet.Id), droplet.Name, publicIpAddressesString, privateIpAddressesString})
	}

	table.Render()
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

func runSsh(args []string) error {
	cmd := exec.Command("ssh", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
