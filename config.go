package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
)

func getAuthTokenFromConfigFile() (string, error) {
	TOKEN_FILE_NAME := ".dropletconn.token"

	userInfo, err := user.Current()
	if err != nil {
		return "", err
	}

	tokenFilePath := path.Join(userInfo.HomeDir, TOKEN_FILE_NAME)
	authTokenFile, err := os.Open(tokenFilePath)
	if err != nil {
		fmt.Printf("Unable to read DigitalOcean API token from \"%s\". Are you sure that file exists?\n", TOKEN_FILE_NAME)
		return "", err
	}
	authTokenBytes, err := ioutil.ReadAll(authTokenFile)
	if err != nil {
		fmt.Printf("Unable to read from token file \"%s\"\n", TOKEN_FILE_NAME)
		return "", err
	}

	return string(authTokenBytes), nil
}

func getDropletsCacheFileName() (string, error) {
	CACHE_FILE_NAME := ".dropletconn.cache"

	userInfo, err := user.Current()
	if err != nil {
		return "", err
	}

	tokenFilePath := path.Join(userInfo.HomeDir, CACHE_FILE_NAME)
	return tokenFilePath, nil
}
