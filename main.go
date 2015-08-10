package main

import (
	"encoding/json"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path"
)

type InterfaceInfo struct {
	IpAddress string `json:"ip_address"`
	Type      string `json:"type"`
}

type NetworkInfo struct {
	V4 []InterfaceInfo `json:"v4"`
}

type DropletInfo struct {
	Id       int         `json:"id"`
	Name     string      `json:"name"`
	Networks NetworkInfo `json:"networks"`
}

type ApiResponseDroplets struct {
	Droplets []DropletInfo `json:"droplets"`
}

func main() {
	TOKEN_FILE_NAME := ".dropletconn.token"

	userInfo, err := user.Current()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	tokenFilePath := path.Join(userInfo.HomeDir, TOKEN_FILE_NAME)
	authTokenFile, err := os.Open(tokenFilePath)
	if err != nil {
		fmt.Printf("Unable to read DigitalOcean API token from \"%s\". Are you sure that file exists?\n", TOKEN_FILE_NAME)
		fmt.Printf("Error: %s\n", err.Error())
		return
	}
	authTokenBytes, err := ioutil.ReadAll(authTokenFile)
	if err != nil {
		fmt.Printf("Unable to read from token file \"%s\"\n", TOKEN_FILE_NAME)
		fmt.Printf("Error: %s\n", err.Error())
		return
	}
	var authToken string = string(authTokenBytes)

	req := gorequest.New()
	resp, body, errs := req.Get("https://api.digitalocean.com/v2/droplets").
		Set("Content-Type", "application/json").
		Set("Authorization", fmt.Sprintf("Bearer %s", authToken)).
		EndBytes()

	if len(errs) != 0 {
		fmt.Println("Encountered errors while trying to get a list of droplets")
		for _, err := range errs {
			fmt.Println(err.Error())
		}

		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Expected 200, got %d as response status\n", resp.StatusCode)
		return
	}

	api_response := ApiResponseDroplets{}
	if err := json.Unmarshal(body, &api_response); err != nil {
		fmt.Printf("Unable to decode json. Error: %s\n", err.Error())
		return
	}

	fmt.Printf("%#v\n", api_response)
}
