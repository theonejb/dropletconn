package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/parnurzeal/gorequest"
	"github.com/ryanuber/go-filecache"
)

type digitalOceanAPI struct {
	authToken string
	droplets  []Droplet

	cache *filecache.Cache
}

func newApiFromConfig(config *Config) (*digitalOceanAPI, error) {
	authToken, err := config.getAuthToken()
	cacheFilePath, err := config.getDropletsCacheFileName()
	cacheDuration := config.CacheDuration

	if err != nil {
		return nil, err
	}

	return newApi(authToken, cacheFilePath, cacheDuration), nil
}

func newApi(authToken, cacheFilePath string, cacheDuration int) *digitalOceanAPI {
	api := &digitalOceanAPI{authToken: authToken}
	api.cache = filecache.New(cacheFilePath, time.Duration(cacheDuration)*time.Minute, api.updateDropletsInfoCacheFile)

	return api
}

func (api *digitalOceanAPI) forceUpdateCache() error {
	return api.cache.Update()
}

func (api *digitalOceanAPI) loadDroplets(forceUpdate bool) error {
	if forceUpdate {
		if err := api.cache.Update(); err != nil {
			return err
		}
	}

	_, err := api.getDroplets()
	return err
}

func (api *digitalOceanAPI) getDroplets() ([]Droplet, error) {
	if api.droplets != nil {
		return api.droplets, nil
	}

	// Also calls the update function if the file doesn't exist or is stale
	fh, err := api.cache.Get()
	if err != nil {
		fmt.Printf("Unable to read cache file. Error: %s\n", err.Error())
		return nil, err
	}

	api.droplets = getDropletsFromCacheFile(fh)
	return api.droplets, nil
}

func (api *digitalOceanAPI) updateDropletsInfoCacheFile(filename string) error {
	droplets := api.getDropletsFromApi()
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

func (api *digitalOceanAPI) getDropletsFromApi() []Droplet {
	req := gorequest.New()
	resp, body, errs := req.Get("https://api.digitalocean.com/v2/droplets?page=1&per_page=1000").
		Set("Content-Type", "application/json").
		Set("Authorization", fmt.Sprintf("Bearer %s", api.authToken)).
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

	apiResponse := &digitalOceanApiResponse{}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		fmt.Printf("Unable to decode json. Error: %s\n", err.Error())
		return nil
	}

	return apiResponse.Droplets
}

func getDropletsFromCacheFile(fh *os.File) []Droplet {
	fileBody, err := ioutil.ReadAll(fh)
	if err != nil {
		fmt.Printf("Unable to read from cache file. Error: %s\n", err.Error())
		return nil
	}

	var droplets []Droplet
	if err := json.Unmarshal(fileBody, &droplets); err != nil {
		fmt.Printf("Unable to decode json. Error: %s\n", err.Error())
		return nil
	}

	return droplets
}

func (api *digitalOceanAPI) getFilteredDroplets(filterExpressions []string) []Droplet {
	dropletsInfo, err := api.getDroplets()
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

	matchedDroplets := make([]Droplet, 0)
	for _, di := range dropletsInfo {
		if di.matchesFilterExpressions(filterExpressions) {
			matchedDroplets = append(matchedDroplets, di)
		}
	}

	return matchedDroplets
}

func (api *digitalOceanAPI) getDropletByName(dropletName string) *Droplet {
	inputDropletNameLower := strings.ToLower(dropletName)

	for _, droplet := range api.droplets {
		dropletNameLower := strings.ToLower(droplet.Name)
		if inputDropletNameLower == dropletNameLower {
			return &droplet
		}
	}

	return nil
}
