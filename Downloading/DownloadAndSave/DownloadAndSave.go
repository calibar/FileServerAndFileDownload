package DownloadAndSave

import (
	"os"
	"net/http"
	"fmt"
	"io"
)

func DownloadAndSave(url string,savePath string,filename string)(err error){
	err = os.MkdirAll(savePath,0755)
	if err!=nil{
		return  err
	} //Create save path if it is not existed.
	outputFile,err:= os.Create(savePath+"\\"+filename) //Create a file in the save path with a filename
	resp,err:=http.Get(url) //Download the file
	if err!=nil{
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode!=http.StatusOK{
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	_,err = io.Copy(outputFile,resp.Body) //copy the content to the created file
	if err!=nil{
		return err
	}else {
		fmt.Println(savePath+"\\"+filename+" has downloaded")
	}
	return nil
}