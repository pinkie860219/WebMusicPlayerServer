package main

import (
	"github.com/BurntSushi/toml"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"fmt"
//	"os"
//	"path/filepath"
)

var conf Config
var pconv = NewPathConv()
var ltb = NewListTable()

func main() {
	tomlData, err := ioutil.ReadFile("./config.toml")
	if err != nil {
		panic("can't read config.toml")
	}
	if _, err := toml.Decode(string(tomlData), &conf); err != nil {
		panic("config parse error")
	}


//	go pconv.StartWatching(conf.Server.Root, "/file/")
	//build the map
	pconv.BuildMap("", "", true)
	pconv.SaveMapToDB()
	//pconv.ReadMapFromDB()
	
	
	log.Println("start")
	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "DELETE"}
	router.Use(cors.New(config))

	//serve for music fils.
	//router.Static(conf.Server.UrlPrefix+"/file/", conf.Server.Root)
	router.GET(conf.Server.UrlPrefix+"/file", serveFileHandler)
	router.GET(conf.Server.UrlPrefix+"/songName", getSongNameHandler)
	
	//serve for files list.
	router.GET(conf.Server.UrlPrefix+"/dir", directoryHandler)

	/////MONGOBD
	router.GET(conf.Server.UrlPrefix+"/songlist", showSongListHandler)
	router.GET(conf.Server.UrlPrefix+"/songlist/:listname", singleSongListHandler)
	router.POST(conf.Server.UrlPrefix+"/songlist", addToSongListHandler)
	router.GET(conf.Server.UrlPrefix+"/songquery", songQueryHandler)
	router.DELETE(conf.Server.UrlPrefix+"/songlist", deleteSongHandler)
	/////

	router.Run(":"+conf.Server.Port)
	log.Println("Serveing on "+conf.Server.Port)
}


func songQueryHandler(c *gin.Context) {
	hashed := c.Query("h")
	log.Println(hashed)
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs: conf.DB.Host,
	})
	if err != nil {
		panic(err)
	}
	defer session.Close()

	//SongLists in DB.
	songListNames, err := session.DB(conf.DB.Name[0]).CollectionNames()
	if err != nil {
		panic(err)
	}

	songListOutput := []string{}
	var songs []Item
	for _, songListName := range songListNames {
		err = session.DB(conf.DB.Name[0]).C(songListName).Find(bson.M{"HashedCode":hashed}).All(&songs)
		if err != nil {
			panic(err)
		}
		if len(songs) != 0 {
			songListOutput = append(songListOutput, songListName)
		}
	}
	//Make the list of json for output.
	list := []string{}
	for _, v := range songListOutput{
		if !strings.Contains(v, "system."){
			list = append(list, v)
		}
	}
	c.JSON(http.StatusOK, list)

}

func showSongListHandler(c *gin.Context) {
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs: conf.DB.Host,
	})
	if err != nil {
		panic(err)
	}
	defer session.Close()

	
	c.JSON(http.StatusOK, ltb.Items())

}

func singleSongListHandler(c *gin.Context) {
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs: conf.DB.Host,
	})
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Collection
	listHashed := c.Param("listname")
	listName := ltb.Query(listHashed)
	log.Println(listName)
	collection := session.DB(conf.DB.Name[0]).C(listName)

	// Find All
	var response []Item
	err = collection.Find(nil).All(&response)
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, response)

}

func addToSongListHandler(c *gin.Context) {
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs: conf.DB.Host,
	})
	if err != nil {
		panic(err)
	}
	defer session.Close()

	songList := c.PostForm("songlist")
	hashed := c.PostForm("hashed")

	// log.Println(fmt.Sprintf("songList:%s, hashed:%s", songList, hashed))
	// Collection
	collection := session.DB(conf.DB.Name[0]).C(songList)

	
	// Insert
	if pconv.Query(hashed) != nil{
		if err := collection.Insert(
			Item{
				HashedCode:hashed,
				Name:pconv.QueryItem(hashed).Name,
				IsDir:false,
			});err != nil {
				panic(err)
			}
	}
	// else {
	// 	c.String(http.StatusNotFound, "Unknown Song")
	// }
	
	c.JSON(http.StatusOK, ltb.Items())
}

func deleteSongHandler(c *gin.Context) {
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs: conf.DB.Host,
	})
	if err != nil {
		panic(err)
	}
	defer session.Close()

	songList := c.PostForm("songlist")
	hashed := c.PostForm("hashed")

	// Collection
	collection := session.DB(conf.DB.Name[0]).C(songList)

	// delete
	if pconv.Query(hashed) != nil{
		if _, err := collection.RemoveAll(bson.M{"HashedCode":hashed});
		err != nil {
			panic(err)
		}
	}
	
	c.JSON(http.StatusOK, ltb.Items())
}


func directoryHandler(c *gin.Context) {
	query_dir := c.Query("dir")
	dirInfo := pconv.Query(query_dir)//type: DirInfo
	
	dirRes := new(DirResponse)
	dirRes.DirArray = dirInfo.DirArray
	dirRes.DirFiles = dirInfo.ItemArray
	
	c.JSON(http.StatusOK, dirRes)
}

func serveFileHandler(c *gin.Context) {
	query_file := c.Query("m")
	real_file := pconv.Query(query_file)
	file_name := ""
	if(real_file != nil){
		file_name = real_file.DirStr
	}
	c.File(conf.Server.Root+file_name)
}

func getSongNameHandler(c *gin.Context) {
	query_file := c.Query("m")
	file_item := pconv.QueryItem(query_file)
	file_name := file_item.Name
	
	c.String(http.StatusOK, file_name)
}

func isAudioExt(val string) bool {
	for i := range conf.Server.AudioExt {
		if conf.Server.AudioExt[i] == val {
			return true
		}
	}
	return false
}

// global config type
type Config struct {
	Server serverConfig `toml:"server"`
	DB     dbConfig     `toml:"database"`
}

type serverConfig struct {
	Root      string   `toml:"root"`
	AudioExt  []string `toml:"audioExt"`
	UrlPrefix string   `toml:"urlPrefix"`
	Port string        `toml:"port"`
}

type dbConfig struct {
	Host []string `toml:"host"`
	Name []string `toml:"dbName"`
}

//datatype of file or folder
type Item struct {
	ID bson.ObjectId  `bson:"_id,omitempty"`
	HashedCode string `bson:"HashedCode"`
	Name   string     `bson:"Name"`
	IsDir  bool       `bson:"IsDir"`
}
func (item Item) String() string {
	return fmt.Sprintf("{Name: %s, HashedCode: %s, IsDir:%t}", item.Name, item.HashedCode, item.IsDir)
}
type DirResponse struct {
	DirArray []*DirItem
	DirFiles []Item
}


