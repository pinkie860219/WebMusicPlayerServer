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
//	"os"
//	"path/filepath"
)

var conf Config
var pconv = NewPathConv()

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
	router.POST(conf.Server.UrlPrefix+"/songquery", songQueryHandler)
	router.DELETE(conf.Server.UrlPrefix+"/songlist", deleteSongHandler)
	/////

	router.Run(":"+conf.Server.Port)
	log.Println("Serveing on "+conf.Server.Port)
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
	songname := c.PostForm("name")
	songurl := c.PostForm("url")
	// Collection
	collection := session.DB(conf.DB.Name).C(songList)

	deleteSong := Song{
		Name: songname,
		Url:  songurl,
	}

	// delete
	if _, err := collection.RemoveAll(deleteSong); err != nil {
		panic(err)
	}

	songListNames, err := session.DB(conf.DB.Name).CollectionNames()
	if err != nil {
		panic(err)
	}

	//Make the list of json for output.
	list := make([]SongListAll, 0)

	list = append(list, SongListAll{
		SongListNames: songListNames,
	})

	c.JSON(http.StatusOK, list)

}

func songQueryHandler(c *gin.Context) {
	songurl := c.PostForm("url")
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs: conf.DB.Host,
	})
	if err != nil {
		panic(err)
	}
	defer session.Close()

	//SongLists in DB.
	songListNames, err := session.DB(conf.DB.Name).CollectionNames()
	if err != nil {
		panic(err)
	}

	songListOutput := []string{}
	var songs []Song
	for _, songListName := range songListNames {
		err = session.DB(conf.DB.Name).C(songListName).Find(bson.M{"url": songurl}).All(&songs)
		if err != nil {
			panic(err)
		}

		if len(songs) != 0 {
			songListOutput = append(songListOutput, songListName)
			log.Println("findin: " + songListName)
		}
	}

	//Make the list of json for output.
	list := make([]SongListAll, 0)

	list = append(list, SongListAll{
		SongListNames: songListOutput,
	})

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

	//SongLists in DB.
	songListNames, err := session.DB(conf.DB.Name).CollectionNames()
	if err != nil {
		panic(err)
	}

	//Make the list of json for output.
	list := make([]SongListAll, 0)

	list = append(list, SongListAll{
		SongListNames: songListNames,
	})

	c.JSON(http.StatusOK, list)

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
	listName := c.Param("listname")

	collection := session.DB(conf.DB.Name).C(listName)

	// Find All
	var songs []Song
	err = collection.Find(nil).All(&songs)
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, songs)

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
	songname := c.PostForm("name")
	songurl := c.PostForm("url")
	// Collection
	collection := session.DB(conf.DB.Name).C(songList)

	insertSong := Song{
		Name: songname,
		Url:  songurl,
	}
	log.Println(insertSong)
	// Insert
	if err := collection.Insert(insertSong); err != nil {
		panic(err)
	}

	//SongLists in DB.
	songListNames, err := session.DB(conf.DB.Name).CollectionNames()
	if err != nil {
		panic(err)
	}

	//Make the list of json for output.
	list := make([]SongListAll, 0)

	list = append(list, SongListAll{
		SongListNames: songListNames,
	})

	c.JSON(http.StatusOK, list)

}

func directoryHandler(c *gin.Context) {
	query_dir := c.Query("dir")
	dirInfo := pconv.Query(query_dir)//type: DirInfo

	if dirInfo == nil{
		dirInfo = pconv.Root()
	}
	
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
	real_file := pconv.Query(query_file)
	file_name := ""
	if(real_file != nil){
		file_name = real_file.DirArray[len(real_file.DirArray)-1].Name
	}
	log.Println(file_name)
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
	Name string   `toml:"dbName"`
}

//datatype of file or folder
type Item struct {
	HashedCode string
	Name   string
	IsDir  bool
}
type DirResponse struct {
	DirArray []*DirItem
	DirFiles []Item
}
//datatype of song in db.
type Song struct {
	Name string
	Url  string
}
type SongInDB struct {
	ID   bson.ObjectId
	Name string
	Url  string
}
type SongUrl struct {
	Url string
}

type SongListAll struct {
	SongListNames []string
}
