package main

import (
	"fmt"
	"strings"
)

type InterfaceInfo struct {
	IpAddress string `json:"ip_address"`
	Type      string `json:"type"`
}

type NetworkInfo struct {
	V4 []InterfaceInfo `json:"v4"`
}

type Droplet struct {
	Id       int         `json:"id"`
	Name     string      `json:"name"`
	Networks NetworkInfo `json:"networks"`
}

type digitalOceanApiResponse struct {
	Droplets []Droplet `json:"droplets"`
}

type interfaceAddresses struct {
	publicIps  []string
	privateIps []string
}

func (droplet *Droplet) getInterfaceAddresses() interfaceAddresses {
	var ia interfaceAddresses

	for _, interfaceInfo := range droplet.Networks.V4 {
		if interfaceInfo.Type == "public" {
			ia.publicIps = append(ia.publicIps, interfaceInfo.IpAddress)
		} else if interfaceInfo.Type == "private" {
			ia.privateIps = append(ia.privateIps, interfaceInfo.IpAddress)
		}
	}

	return ia
}

func (droplet *Droplet) matchesFilterExpressions(filterExpressions []string) bool {
	if len(filterExpressions) == 0 {
		return true
	}

	dropletNameLower := strings.ToLower(droplet.Name)

	netAdd := droplet.getInterfaceAddresses()
	publicIpAddressesString := strings.Join(netAdd.publicIps, ", ")
	privateIpAddressesString := strings.Join(netAdd.privateIps, ", ")

	for _, fe := range filterExpressions {
		if strings.Contains(dropletNameLower, fe) || strings.Contains(publicIpAddressesString, fe) || strings.Contains(privateIpAddressesString, fe) {
			return true
		}
	}

	return false
}

func (droplet *Droplet) getArgsForSsh(userConf *Config) []string {
	interfaceAddresses := droplet.getInterfaceAddresses()
	if len(interfaceAddresses.publicIps) == 0 {
		fmt.Printf("No public IP address found for droplet \"%s\"\n", droplet.Name)
		return nil
	}

	cmdOptions := make([]string, 0, 5)
	cmdOptions = append(cmdOptions, interfaceAddresses.publicIps[0])

	if userConf.DefaultUser != "" {
		cmdOptions = append(cmdOptions, "-l")
		cmdOptions = append(cmdOptions, userConf.DefaultUser)
	}
	if userConf.DefaultKeyFileName != "" {
		keyFileName := userConf.DefaultKeyFileName
		keyFilePath, err := getAbsoluteFilePath(keyFileName)
		if err != nil {
			fmt.Printf("Unable to get absolute ssh key file path. Error: %s\n", err.Error())
			return nil
		}

		cmdOptions = append(cmdOptions, "-i")
		cmdOptions = append(cmdOptions, keyFilePath)
	}

	return cmdOptions
}
