package main

import "Downloading/DownloadTiles"

func main()  {
	var config DownloadTiles.RequestConfiguration
	DownloadTiles.DownloadrequiredTiles(config)
}