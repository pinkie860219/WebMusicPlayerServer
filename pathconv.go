package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/fsnotify/fsnotify"
//	"io/ioutil"
)
type DirItem struct {
	Name string
	HashedCode string
}

type DirInfo struct {
	DirArray []*DirItem
	DirStr string
}
func NewDirInfo(dArray []*DirItem, dStr string) *DirInfo{
	di := new(DirInfo)
	di.DirArray = dArray
	di.DirStr = dStr
	return di
} 

type PathConv struct {
	table           map[string] *DirInfo /* hashed (12bit) -> original path*/
	isBuildingTable bool
	watcher         *fsnotify.Watcher
}

func NewPathConv() *PathConv {
	pcv := new(PathConv)
	pcv.table = make(map[string] *DirInfo)
	pcv.isBuildingTable = false
	pcv.watcher = nil
	return pcv
}

func PathConvHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", sum)[:11]
}

func (pcv *PathConv) AddHash(pre_hashed string, file_name string) string{
	preDirInfo := pcv.Query(pre_hashed)
	prefix := ""
	if(preDirInfo != nil){
		prefix = preDirInfo.DirStr
	}
	hashed := PathConvHash(prefix+"/"+file_name)

	dirItem := new(DirItem)
	dirItem.Name = file_name
	dirItem.HashedCode = hashed
	
	var newArray []*DirItem
	if(preDirInfo != nil){
		newArray = append(newArray, preDirInfo.DirArray...)
	}
	newArray = append(newArray, dirItem)
	dirInfo := NewDirInfo(newArray, prefix+"/"+file_name)
	pcv.table[hashed] = dirInfo
	return hashed
}

/**
 * This function implments relative path (local filesystem) to record path (url path).
 *
 * E.g.
 *    "/var/log/data/cat" (local filesystem path)
 * => "179c2d3a892b"      (hashed from local filesystem path)
 * => "./files/cat"       (path for url)
 */

//func (pcv *PathConv) buildFromImpl(walkPath string, recordPath string, prefix string) {
//	files, err := ioutil.ReadDir(walkPath + prefix)
//	if err != nil {
//		panic(err)
//	}
//
//	for _, f := range files {
//		if f.IsDir() {
//			pcv.buildFromImpl(walkPath, recordPath, prefix+f.Name()+"/")
//			pcv.AddHash(recordPath + prefix + f.Name() + "/")
//			pcv.watcher.Add(walkPath + prefix + f.Name())
//		} else {
//			pcv.AddHash(recordPath + prefix + f.Name())
//		}
//	}
//}
//
//func (pcv *PathConv) BuildFrom(dir string, recordPath string) {
//	if pcv.isBuildingTable {
//		fmt.Println("Building in process. Skip.")
//		return
//	}
//
//	fmt.Println("Pathconv building...")
//	if dir[len(dir)-1] != []byte("/")[0] {
//		dir += "/"
//	}
//	pcv.isBuildingTable = true
//	defer func() { pcv.isBuildingTable = false }()
//	pcv.buildFromImpl(dir, recordPath, "")
//	fmt.Println("Pathconv built")
//}

func (pcv *PathConv) Query(hashed string) *DirInfo {
	return pcv.table[hashed]
}

func (pcv *PathConv) Show() {
	for i, v := range pcv.table {
		fmt.Println(i, "=>", v)
	}
}

//func (pcv *PathConv) StartWatching(dir string, recordPath string) {
//	watcher, err := fsnotify.NewWatcher()
//	if err != nil {
//		panic(err)
//	}
//	defer watcher.Close()
//	pcv.watcher = watcher
//	// build for first time
//	pcv.BuildFrom(dir, recordPath)
//
//	done := make(chan bool)
//	go func() {
//		for {
//			select {
//			case event := <-watcher.Events:
//				fmt.Println("event:", event)
//				// clear pcv.table
//				pcv.table = make(map[string]string)
//				pcv.BuildFrom(dir, recordPath)
//			}
//		}
//	}()
//
//	err = watcher.Add(dir)
//	if err != nil {
//		panic(err)
//	}
//	<-done
//}
