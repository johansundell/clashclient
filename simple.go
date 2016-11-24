package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type page struct {
	Title       string
	Name        string
	Description string
	MembersJson template.JS
	Image       string
}

type Test []Player

func init() {
	routes = append(routes, Route{"default", "GET", "/", handleDefault})
	routes = append(routes, Route{"clan-info", "GET", "/tmpl/clan-info.html", handleClanInfo})
}

func handleDefault(w http.ResponseWriter, req *http.Request) {
	data, err := Asset("pages/index.html")
	if err != nil {
		log.Println(err)
	}
	t, err := template.New("index.html").Delims("*{{", "}}*").Parse(string(data))
	if err != nil {
		log.Println(err)
	}

	p := page{}
	p.Title = "Clash Clan Viewer"

	err = t.Execute(w, p)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		fmt.Println(err)
	}
}

func handleClanInfo(w http.ResponseWriter, req *http.Request) {
	data, err := Asset("tmpl/clan-info.html")
	if err != nil {
		log.Println(err)
	}
	t, err := template.New("clan-info.html").Delims("*{{", "}}*").Parse(string(data))
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		fmt.Println(err)
		return
	}

	p := page{}
	//members := getMembersFromDb()
	members := getSmallMembersFromDb()
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		fmt.Println(err)
		return
	}
	//fmt.Println(Test(members))

	b, err := json.Marshal(members)
	//fmt.Println(string(b))
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		fmt.Println(err)
		return
	}
	//p.Name = clan.Name
	//p.Description = clan.Description

	p.MembersJson = template.JS(string(b))
	//p.Image = clan.BadgeUrls.Small

	err = t.Execute(w, p)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		fmt.Println(err)
	}
}
