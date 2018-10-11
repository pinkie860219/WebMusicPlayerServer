package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io/ioutil"
	"path/filepath"
	"log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)
type DirItem struct {
	Name string
	HashedCode string
}
func (diri DirItem) String() string {
	return fmt.Sprintf("{Name: %s, HashedCode: %s}", diri.Name, diri.HashedCode)
}
type DirInfo struct {
	DirArray []*DirItem
	DirStr string
	ItemArray []Item
}
func (diri DirInfo)String()string{
	return fmt.Sprintf("DirArray: %s\nDirStr: %s\nItemArray: %s\n\n", diri.DirArray, diri.DirStr, diri.ItemArray)	
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
  
func (pcv *PathConv) Query(hashed string) *DirInfo {
	dirInfo := pcv.table[hashed]
	if dirInfo == nil{
		dirInfo = pconv.Root()
	}
	return dirInfo
}
func (pcv *PathConv) Root() *DirInfo {
	for _, v := range pcv.table {
		if v.DirStr == ""{
			return v
		}
	}
	return nil
}
func (pcv *PathConv) Show() {
	for i, v := range pcv.table {
		fmt.Println(i, "=>", v)
	}
}

func (pcv *PathConv) BuildMap(pre_hashed string, file_name string, isDir bool) string{
	//  1. 根據 hashed 找到前一路徑
	preDirInfo := pcv.Query(pre_hashed)
	prefix := ""
	if preDirInfo != nil {
		prefix = preDirInfo.DirStr
	}
	//  2. 計算自身的hashed
	convName := ""
	if file_name != "" {
		convName = "/"+file_name
	}
	curDir := prefix+convName
	hashed := PathConvHash(curDir)

	//  6. 將自身紀錄到map中
	dirItem := new(DirItem)
	dirItem.Name = file_name
	dirItem.HashedCode = hashed
	
	var breadArray []*DirItem
	if(preDirInfo != nil){
		breadArray = append(breadArray, preDirInfo.DirArray...)
	}
	if curDir != ""{
		breadArray = append(breadArray, dirItem)
	}
	
	newDirInfo := NewDirInfo(breadArray, curDir)
	pcv.table[hashed] = newDirInfo
	
	//  3. 前一路徑加上本身名稱， 讀取目前路徑中的檔案
	folderArray := make([]Item, 0)
	songArray := make([]Item, 0)
	if isDir {
		openDir := conf.Server.Root + curDir
		//log.Println(fmt.Sprintf("Now opening:[%s]:'%s'",hashed,openDir))
		files, err := ioutil.ReadDir(openDir)
		if err != nil {
			log.Println("---ERROR opening:'" + openDir +"'")
			pcv.Show()
			panic(err)
		}
		//  4. 把自身的hashed和子檔案傳給下一個遞迴
		//  5. 紀錄子檔案和其hashed陣列
		//hashed := pcv.AddHash(pre_hashed, file_name)
		
		for _, f := range files {
			ext := filepath.Ext(f.Name())
			if (f.Name()[0] != []byte(".")[0]) && (f.IsDir() || isAudioExt(ext)) {
				item := Item{
					Name: f.Name(),
					IsDir: f.IsDir(),
					HashedCode:pcv.BuildMap(hashed, f.Name(), f.IsDir()),
				}
				if f.IsDir() {
					folderArray = append(folderArray,item)
				} else {
					songArray = append(songArray,item)
				}
			}
		}
	}
	pcv.table[hashed].ItemArray = append(folderArray, songArray...)
	//  7. 回傳 hashed
	return hashed
}

func (pcv *PathConv) SaveMapToDB(){
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs: conf.DB.Host,
	})
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Collection
	collection := session.DB(conf.DB.Name[1]).C("table")
	// Insert
	if err := collection.Insert(pcv.table); err != nil {
		panic(err)
	}

	
	
}
func (pcv *PathConv) ReadMapFromDB(){
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs: conf.DB.Host,
	})
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Collection
	collection := session.DB(conf.DB.Name[1]).C("table")
	// Find
	collection.Find(bson.M{}).One(&pcv.table)
	
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
