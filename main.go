package main

import (
	"fmt"
	"time"

	"github.com/ryanuber/go-filecache"
)

func main() {
	fc := filecache.New("droplets.cache", 5*time.Second, updateDropletsInfoCacheFile)
	fh, err := fc.Get()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	dropletsInfo := readDropletsInfoCacheFile(fh)
	if dropletsInfo == nil {
		return
	}

	fmt.Printf("%#v\n", dropletsInfo)
}
