package DownloadTiles
import (
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"os"
	"strings"
	"strconv"
	"log"
	"Downloading/DownloadAndSave"
)

type RequestConfiguration struct {
	name string
	bounds string
	format string
	gridSetId string
	OperationType string
	threadCount string
	zoomStart string
	zoomStop string
	url string
}
type tileInfo struct {
	Path string //relatively filepath
	Layername string //layername
	GridSetId string //projection system identifer
	Z int //zoomlevel
	X int //tileX
	Y int //tileY
}

func setConfiguration(config RequestConfiguration)(configOut RequestConfiguration){
	/*fmt.Println("layername:")
	fmt.Scan(&config.name)
	fmt.Println("bounds:")
	fmt.Scan(&config.bounds)
	fmt.Println("gridsetid:")
	fmt.Scan(&config.gridSetId)
	fmt.Println("zoomStart:")
	fmt.Scan(&config.zoomStart)
	fmt.Println("zoomStop")
	fmt.Scan(&config.zoomStop)*/
	config.name="nurc_Arc_Sample"
	config.bounds="-180A-85A180A85"
	config.gridSetId="4326"
	config.zoomStart="0"
	config.zoomStop="2"
	config.url="http://localhost:8050/d"
	return config
}
func getTilesDir(config RequestConfiguration)(err error,tileinfo []tileInfo){
	if err!=nil {
		return err,tileinfo
	}
	req,err := http.NewRequest("Get",config.url,nil)
	q:=req.URL.Query()
	q.Add("name",config.name)
	q.Add("bounds",config.bounds)
	q.Add("gridSetId",config.gridSetId)
	q.Add("zoomStart",config.zoomStart)
	q.Add("zoomStop",config.zoomStop)
	req.URL.RawQuery = q.Encode()
	queryUrl:=req.URL.String()
	fmt.Println(req.URL.String())
	resp,err:= http.Get(queryUrl)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err!=nil{
		return err,tileinfo
	} else {
		json.Unmarshal(body,&tileinfo)
	}
	return nil,tileinfo
}

func downloadtiles(tiles []tileInfo)(err error){
	for _,t := range tiles{
		fmt.Println("Downloading:"+t.Path)
		url:=strings.Replace("http://localhost:8050/s/"+t.Path,"\\","/",3)
		localPath:= t.Layername+"\\"+"EPSG_"+t.GridSetId+"\\"+strconv.Itoa(t.Z)+"\\"
		err =  os.MkdirAll(localPath,0755)
		filename := strconv.Itoa(t.X)+"_"+strconv.Itoa(t.Y)+".png"
		DownloadAndSave.DownloadAndSave(url,localPath,filename)
	}
	if err != nil  {
		return err
	}


	return nil
}
func DownloadrequiredTiles(config RequestConfiguration)(error){
	config=setConfiguration(config)
	fmt.Printf("%+v",config)
	err,tiles:=getTilesDir(config)
	if err!=nil {
		log.Fatal(err)
	}
	downloadtiles(tiles)
	return nil
}