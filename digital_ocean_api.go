package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"io/ioutil"
	"net/http"
	"os"
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

func getDroplets(authToken string) *ApiResponseDroplets {
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

		return nil
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Expected 200, got %d as response status\n", resp.StatusCode)
		return nil
	}

	apiResponse := &ApiResponseDroplets{}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		fmt.Printf("Unable to decode json. Error: %s\n", err.Error())
		return nil
	}

	return apiResponse
}

/*
Get info about all the droplets from the DO API and store it in the given file as Json
*/
func updateDropletsInfoCacheFile(filename string) error {
	authToken, err := getAuthTokenFromConfigFile()
	if err != nil {
		return err
	}

	var droplets *ApiResponseDroplets = getDroplets(authToken)
	if droplets == nil {
		return errors.New("Unable to get droplets")
	}

	jsonStr, err := json.MarshalIndent(droplets, "", "  ")
	if err != nil {
		fmt.Printf("Unable to convert DO Api response to JSON for caching. Error: %s\n", err.Error())
		return err
	}

	cacheFile, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Unable to create cache file. Error: %s\n", err.Error())
		return err
	}

	_, err = cacheFile.Write(jsonStr)
	if err != nil {
		fmt.Printf("Error while writing cache file. Error: %s\n", err.Error())
		return err
	}

	return nil
}

func readDropletsInfoCacheFile(fh *os.File) *ApiResponseDroplets {
	fileBody, err := ioutil.ReadAll(fh)
	if err != nil {
		fmt.Printf("Unable to read from cache file. Error: %s\n", err.Error())
		return nil
	}

	apiResponse := &ApiResponseDroplets{}
	if err := json.Unmarshal(fileBody, &apiResponse); err != nil {
		fmt.Printf("Unable to decode json. Error: %s\n", err.Error())
		return nil
	}

	return apiResponse
}
