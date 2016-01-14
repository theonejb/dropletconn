package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/ryanuber/go-filecache"
)

type digitalOceanAPI struct {
	authToken string
	cache     *filecache.Cache
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
	fileCache := filecache.New(cacheFilePath, time.Duration(cacheDuration)*time.Minute, updateDropletsInfoCacheFile)

	return &digitalOceanAPI{authToken, fileCache}
}

func (api *digitalOceanAPI) forceUpdateCache() error {
	return api.cache.Update()
}

func (api *digitalOceanAPI) getDroplets() ([]DropletInfo, error) {
	fh, err := api.cache.Get()
	if err != nil {
		fmt.Printf("Unable to read cache file. Error: %s\n", err.Error())
		return nil, err
	}

	dropletsInfo := readDropletsInfoCacheFile(fh)
	return dropletsInfo.Droplets, nil
}

func (api *digitalOceanAPI) getFilteredDroplets(rConf runningConf, filterExpressions []string) []DropletInfo {
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

	matchedDroplets := make([]DropletInfo, 0)
	for _, di := range dropletsInfo {
		if di.matchesFilterExpressions(filterExpressions) {
			matchedDroplets = append(matchedDroplets, di)
		}
	}

	return matchedDroplets
}
