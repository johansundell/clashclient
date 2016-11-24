package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
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

	routes = append(routes, Route{"members", "GET", "/members", handleGetMembers})
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
	sortDir := req.FormValue("sort")
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		log.Println(err)
		return
	}
	t, err := template.New("clan-info.html").Delims("*{{", "}}*").Parse(string(data))
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		log.Println(err)
		return
	}

	p := page{}
	clan, err := getClan(mySettings.clan)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		log.Println(err)
		return
	}
	//members := getMembersFromDb()
	//members := getSmallMembersFromDb(mySettings.clan)
	members := getMembers(mySettings.clan, sortDir)
	/*if err != nil {
		http.Error(w, http.StatusText(500), 500)
		log.Println(err)
		return
	}*/
	//fmt.Println(Test(members))

	b, err := json.Marshal(members)
	//fmt.Println(string(b))
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		log.Println(err)
		return
	}
	p.Name = clan.Name
	p.Description = clan.Description

	p.MembersJson = template.JS(string(b))
	p.Image = clan.BadgeUrls.Small

	err = t.Execute(w, p)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		log.Println(err)
	}
}

func getPublicIP() []byte {
	resp, _ := getUrl("http://kontoret.pixpro.net/ip")
	//fmt.Println(string(resp))
	return resp
}

func getUrl(url string) (b []byte, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		b = []byte{}
		err = errors.New("Error from server: " + strconv.Itoa(resp.StatusCode))
	}
	return
}

func handleGetMembers(w http.ResponseWriter, req *http.Request) {
	sortDir := req.FormValue("sort")
	members := getMembers(mySettings.clan, sortDir)

	b, err := json.Marshal(members)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		fmt.Println(err)
		return
	}
	fmt.Fprint(w, string(b))
}
