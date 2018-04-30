package main

import (
"net/http"
"fmt"
	"path/filepath"
	"os"
	"io/ioutil"
	"strconv"
	"strings"
	"math"
	"encoding/json"
)
type Dir struct {
//httpRequestHandler
}
type Tile struct {
	Z    int //zoomlevel
	X    int //tileX coordinator
	Y    int //tileY coordinator
	Lat  float64 // latitude
	Long float64 // longitude
}
type tileInfo struct {
	Path string //relatively filepath
	Layername string //layername
	GridSetId string //projection system identifer
	Z int //zoomlevel
	X int //tileX
	Y int //tileY
} //Information of tile to be sent
type Conversion interface {
	deg2num(t *Tile) (x int, y int) //convert lat,lon to tileX, tileY
	num2deg(t *Tile) (lat float64, long float64) //convert X,Y to lat,lon
}
func (*Tile) Deg2num4326(t Tile) (x int, y int) {
	/*x = int(math.Floor((t.Long + 180.0) / 360.0 * (math.Exp2(float64(t.Z)))))
	y = int(math.Floor((1.0 - math.Log(math.Tan(t.Lat*math.Pi/180.0)+1.0/math.Cos(t.Lat*math.Pi/180.0))/math.Pi) / 2.0 * (math.Exp2(float64(t.Z)))))*/
	x = int(math.Floor((t.Long + 180.0) / 180.0 * (math.Exp2(float64(t.Z)))))
	y= int((math.Exp2(float64(t.Z))))-1- int(math.Floor((1.0 - math.Log(math.Tan(t.Lat*math.Pi/180.0)+1.0/math.Cos(t.Lat*math.Pi/180.0))/math.Pi) / 2.0 * (math.Exp2(float64(t.Z)))))
	return
}
func (*Tile) Deg2num900913(t Tile) (x int, y int) {
	x = int(math.Floor((t.Long + 180.0) / 360.0 * (math.Exp2(float64(t.Z)))))
	/*y = int(math.Floor((1.0 - math.Log(math.Tan(t.Lat*math.Pi/180.0)+1.0/math.Cos(t.Lat*math.Pi/180.0))/math.Pi) / 2.0 * (math.Exp2(float64(t.Z)))))
	x = int(math.Floor((t.Long + 180.0) / 180.0 * (math.Exp2(float64(t.Z)))))*/
	y= int((math.Exp2(float64(t.Z))))-1- int(math.Floor((1.0 - math.Log(math.Tan(t.Lat*math.Pi/180.0)+1.0/math.Cos(t.Lat*math.Pi/180.0))/math.Pi) / 2.0 * (math.Exp2(float64(t.Z)))))
	return
}
func (*Tile) Num2deg(t Tile) (lat float64, long float64) {
	n := math.Pi - 2.0*math.Pi*float64(t.Y)/math.Exp2(float64(t.Z))
	lat = 180.0 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
	long = float64(t.X)/math.Exp2(float64(t.Z))*360.0 - 180.0
	return lat, long
}
type layerSeedConfig struct {
	name string
	bounds []string
	gridSetId string
	zoomStart string
	zoomStop string
	format string
	operationType string
	threadCount string
} // Request's Parameters to get specific tile files
func LatLonToPixels(lat float64,lon float64,zoomlevel int)(px float64,py float64){
	res1:=180.0 / 200.0
	res:=res1/math.Exp2(float64(zoomlevel))
	px = (180 + lat) / res
	py = (90 + lon) / res
	return px, py
}
func PixelsToTile(px float64,py float64)(tx int,ty int){
	tx = int( math.Ceil( px / 200) - 1 )
	ty = int( math.Ceil( py / 200) - 1 )
	return tx,ty
}
func LatLonToTile(lat float64,lon float64,zoomlevel int)(tx int,ty int){
	px,py:=LatLonToPixels(lat,lon,zoomlevel)
	tx,ty=PixelsToTile(px,py)
	return tx,ty
}
func FilePathWalkDir(root string) ([]string,error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files,err
}// get all the filepath under a directory
func IOReadDir(root string) ([]string, error) {
	var files []string
	fileInfo, err := ioutil.ReadDir(root)
	if err != nil {
		return files, err
	}

	for _, file := range fileInfo {
		files = append(files, file.Name())
	}
	return files, nil
} // get all the filepath under a directory
func OSReadDir(root string) ([]string, error) {
	var files []string
	f, err := os.Open(root)
	if err != nil {
		return files, err
	}
	fileInfo, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return files, err
	}

	for _, file := range fileInfo {
		files = append(files, file.Name())
	}
	return files, nil
} // get all the filepat under a directory
func addquote(in string)(out string){
	out=`"`+in+`"`
	return out
} //add quote to a string
func (layer layerSeedConfig)ChangetoString()(str string){
	str=`{"seedRequest":{
	"name":`+addquote(layer.name)+`,
	"bounds":{"coords":{ "double":[`+addquote(layer.bounds[0])+`,`+addquote(layer.bounds[1])+`,`+addquote(layer.bounds[2])+`,`+addquote(layer.bounds[3])+`]}},
	"srs":{"number":`+addquote(layer.gridSetId)+`},
	"zoomStart":`+addquote(layer.zoomStart)+`,
	"zoomStop":`+addquote(layer.zoomStop)+`,
	"format":`+addquote(layer.format)+`,
	"type":`+addquote(layer.operationType)+`,
        "threadCount":`+addquote(layer.threadCount)+`
	}
}}

`
return str
} // translate the request to a Json format geowebcahche use
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
} // Find if there is a file in the path
func setLim(layer layerSeedConfig,zoomlevel int)(Xmin int, Ymin int, Xmax int, Ymax int)  {
	var Tmin,Tmax Tile
	Tmin.Z=zoomlevel
	Tmax.Z=zoomlevel
	Tmin.Long,_=strconv.ParseFloat(layer.bounds[0],64)
	Tmin.Lat,_=strconv.ParseFloat(layer.bounds[1],64)
	Tmax.Long,_ =strconv.ParseFloat(layer.bounds[2],64)
	Tmax.Lat,_=strconv.ParseFloat(layer.bounds[3],64)
	if layer.gridSetId=="4326" {
		Tmin.X,Tmin.Y=Tmin.Deg2num4326(Tmin)
		Tmax.X,Tmax.Y=Tmax.Deg2num4326(Tmax)
	}else if layer.gridSetId=="900913" {
		Tmin.X,Tmin.Y=Tmin.Deg2num900913(Tmin)
		Tmax.X,Tmax.Y=Tmax.Deg2num900913(Tmax)
	}else {
		return -1,-1,-1,-1
	}
	return Tmin.X, Tmin.Y,Tmax.X,Tmax.Y
} // translate request parameters to the tileX, tileY range
func getTilename(path string)(x int,y int){
	str1:=strings.SplitN(path,"\\",4)
	str2:=strings.Split(str1[3],".")
	str3:=strings.Split(str2[0],"_")
	x,_=strconv.Atoi(str3[0])
	y,_=strconv.Atoi(str3[1])
	return
} // Get the tileX.tileY from tile path
func (d Dir) ServeHTTP(w http.ResponseWriter,req *http.Request)  {
	method:=req.Method
	if(method=="GET"){
		/*files,_ := ioutil.ReadDir("./")*/
		/*for _,f:= range files{
			d1:=string(f.Name())
			f1,_:=ioutil.ReadDir("./"+d1)
			for _,f2:=range f1{
				fmt.Println(f2.Name())
			}
		}*/
		fmt.Println("Get a post request")
		var layerReq layerSeedConfig
		layerReq.name= req.FormValue("name")
		boundspre:= req.FormValue("bounds")
		layerReq.format=req.FormValue("format")
		layerReq.gridSetId = req.FormValue("gridSetId")
		layerReq.operationType=req.FormValue("type")
		layerReq.threadCount=req.FormValue("threadCount")
		layerReq.zoomStart=req.FormValue("zoomStart")
		layerReq.zoomStop=req.FormValue("zoomStop")
		layerReq.bounds = strings.SplitN(boundspre,"A",4) // Get the parameters
		fmt.Println(layerReq)
		/*path:="./"+layerReq.name
		files,_:=FilePathWalkDir(path)*/
		var files []string
		var tilesPath []string
		var tilesinfo []tileInfo
		jsonContent := layerReq.ChangetoString()
		fmt.Println(jsonContent)
		zoomStart,_:=strconv.Atoi(layerReq.zoomStart)
		zoomStop,_:=strconv.Atoi(layerReq.zoomStop)
		for zoomlevel:=zoomStart;zoomlevel<=zoomStop ;zoomlevel++  {
			xmin,ymin,xmax,ymax:=setLim(layerReq,zoomlevel)
			fmt.Println(zoomlevel)
			fmt.Println(xmin,ymin,xmax,ymax)
			var path string
			if zoomlevel>=10{
				path=layerReq.name+"/EPSG_"+layerReq.gridSetId+"_"+strconv.Itoa(zoomlevel)
			}else {
				path=layerReq.name+"/EPSG_"+layerReq.gridSetId+"_0"+strconv.Itoa(zoomlevel)
			}

			flag:=Exists(path)
			if flag==true{
				path="./"+path
				files,_=FilePathWalkDir(path)
				for _,f :=range files{
					x,y:=getTilename(f)
					if x>=xmin&&x<=xmax&&y>=ymin&&y<=ymax {
						tilesPath=append(tilesPath, f)
						var tileinfo tileInfo
						tileinfo.Layername= layerReq.name
						tileinfo.GridSetId= layerReq.gridSetId
						tileinfo.Z=zoomlevel
						tileinfo.X=x
						tileinfo.Y=y
						tileinfo.Path=f
						tilesinfo=append(tilesinfo,tileinfo)
						fmt.Printf("%+v\n", tileinfo)
					}
				}
				fmt.Println(files)
			}else {
				fmt.Println("no file")
			}


		} // Get all the path of the required tiles
		/*files1,_:= IOReadDir("./")
		files2,_:= OSReadDir("./")*/

		/*fmt.Println(files)
		fmt.Println(files1)
		fmt.Println(files2)*/
		/*strBody,_:=json.Marshal(jsonContent)
		reqBody:=strings.NewReader(string(strBody))
		resp,_:=http.Post("http://localhost:8080/geowebcache/rest/seed/"+layerReq.name+".json","application/json",reqBody)
		defer resp.Body.Close()
		body,_:=ioutil.ReadAll(resp.Body)
		fmt.Println(body)*/
		tilesinfoJson,_:=json.Marshal(tilesinfo)
		fmt.Println(tilesinfoJson)
		if files!=nil{
			/*for _,f := range tilesPath{
				str:="http://localhost:8050/s/"+string(f)+"\n"
				fmt.Println(str)
				w.Write([]byte(string(str)))
			}*/
			w.Write(tilesinfoJson)
			fmt.Println("done")
		}else {
			w.Write([]byte("No Results"))
		} // send the path of required tiles as response


	}

}
func main()  {
	var a Dir
	http.Handle("/s/", http.StripPrefix("/s/", http.FileServer(http.Dir("./"))))
	http.Handle("/d",a)
	err := http.ListenAndServe(":8050", nil)
	if err != nil {
		fmt.Println(err)
	}
}

