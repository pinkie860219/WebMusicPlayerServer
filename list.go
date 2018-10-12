package main

import (
	"gopkg.in/mgo.v2"
	"strings"
)

type ListTable struct {
	table  map[string] string
}
func NewListTable() *ListTable{
	lt := new(ListTable)
	lt.table = make(map[string] string)
	return lt
}
func (lt *ListTable)Add(listname string)string{
	hashed := PathConvHash(listname)
	lt.table[hashed] = listname
	return hashed
}
func (lt *ListTable)Query(hashed string)string{
	return lt.table[hashed]
}
func (lt *ListTable)QueryItem(hashed string)DirItem{
	return DirItem{Name:lt.table[hashed], HashedCode:hashed}
}

func (lt *ListTable)Update(){
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs: conf.DB.Host,
	})
	if err != nil {
		panic(err)
	}
	defer session.Close()
	songListNames, err := session.DB(conf.DB.Name[0]).CollectionNames()
	if err != nil {
		panic(err)
	}
	for _, v := range songListNames{
		if !strings.Contains(v, "system."){
			ltb.Add(v)
		}
	} 	
}
func (lt *ListTable)Items()[]DirItem{
	lt.Update()
	ia := make([]DirItem,0)
	for i,_ := range lt.table{
		ia = append(ia, lt.QueryItem(i))
	}
	return ia
}
