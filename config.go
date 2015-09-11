package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
)

type Config struct {
	Token         string `json:"token"`
	CacheFileName string `json:"cache_file_name"`
	CacheDuration int    `json:"cache_duration"`

	DefaultUser        string `json:"default_user"`
	DefaultKeyFileName string `json:"default_key_file_name"`
}

const CONFIG_FILE_NAME = ".dropletconn.conf"

const DEFAULT_CACHE_FILE_NAME = ".dropletconn.cache"
const DEFAULT_CACHE_DURATION = 60
const DEFAULT_USER = ""
const DEFAULT_KEY_FILE_NAME = ""

var config *Config

func getConfig() (*Config, error) {
	if config != nil {
		return config, nil
	}

	config = new(Config)

	configFile, err := getConfigFh()
	if err != nil {
		return nil, err
	}
	defer func() {
		configFile.Close()
	}()

	configFileBytes, err := ioutil.ReadAll(configFile)
	if err != nil {
		fmt.Printf("Unable to read config file. Error: %s\n", err.Error())
		return nil, err
	}

	err = json.Unmarshal(configFileBytes, config)
	if err != nil {
		fmt.Printf("Unable to decode JSON from config file. Error: %s\n", err.Error())
		return nil, err
	}

	return config, nil
}

func (c *Config) getDropletsCacheFileName() (string, error) {
	return getAbsoluteFilePath(c.CacheFileName)
}

func (c *Config) getAuthToken() (string, error) {
	token := c.Token
	token = strings.TrimSpace(token)

	if token == "" {
		fmt.Println("Token is empty. Please update config file.")
		return "", errors.New("Empty token in config file")
	}
	return token, nil
}

func getConfigParamValueWithDefault(prompt, defaultValue string) (string, error) {
	inputReader := bufio.NewReader(os.Stdin)

	fmt.Printf("%s [%s]: ", prompt, defaultValue)
	input, err := inputReader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error while reading user input. Error: %s\n", err.Error())
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue, nil
	} else {
		return input, nil
	}
}

func createConfig() error {
	token, err := getConfigParamValueWithDefault("Token", "")
	if err != nil {
		return err
	}

	cacheFileName, err := getConfigParamValueWithDefault("Cache file name", DEFAULT_CACHE_FILE_NAME)
	if err != nil {
		return err
	}

	cacheDuration, err := getConfigParamValueWithDefault("Cache duration (in minutes)", strconv.Itoa(DEFAULT_CACHE_DURATION))
	if err != nil {
		return err
	}

	cacheDurationInt, err := strconv.Atoi(cacheDuration)
	if err != nil {
		fmt.Printf("Unable to conver input \"%s\" into integer. Error: %s\n", cacheDuration, err.Error())
		return err
	}

	defaultUser, err := getConfigParamValueWithDefault("Default user (optional)", "")
	if err != nil {
		return err
	}

	defaultKeyFileName, err := getConfigParamValueWithDefault("Default SSH key file (optional)", "")
	if err != nil {
		return err
	}

	newConf := Config{
		Token:              token,
		CacheFileName:      cacheFileName,
		CacheDuration:      cacheDurationInt,
		DefaultUser:        defaultUser,
		DefaultKeyFileName: defaultKeyFileName,
	}

	configFilePath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	configFile, err := os.Create(configFilePath)
	if err != nil {
		fmt.Printf("Unable to create config file. Error: %s\n", err.Error())
		return err
	}
	defer func() {
		configFile.Close()
	}()

	configJson, err := json.Marshal(newConf)
	if err != nil {
		fmt.Printf("Unable to encode JSON. Error: %s\n", err.Error())
		return err
	}

	if _, err = configFile.Write(configJson); err != nil {
		fmt.Printf("Unable to write to config file. Error: %s\n", err.Error())
		return err
	}

	return nil
}

func getConfigFilePath() (string, error) {
	userInfo, err := user.Current()
	if err != nil {
		return "", err
	}

	configFilePath := path.Join(userInfo.HomeDir, CONFIG_FILE_NAME)
	return configFilePath, nil
}

func getConfigFh() (*os.File, error) {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}

	configFile, err := os.Open(configFilePath)
	if err != nil {
		configFile.Close()

		fmt.Printf("Unable to open config file at \"%s\".\nIf the file does not exist, you can create is using the \"config\" command.\nError: %s\n", err.Error())
		return nil, err
	}

	return configFile, nil
}

func getAbsoluteFilePath(givenFilePath string) (string, error) {
	if !path.IsAbs(givenFilePath) {
		userInfo, err := user.Current()
		if err != nil {
			return "", err
		}

		return path.Join(userInfo.HomeDir, givenFilePath), nil
	} else {
		return givenFilePath, nil
	}
}
