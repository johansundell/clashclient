package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/boltdb/bolt"
	"github.com/kardianos/osext"
)

type settings struct {
	port   string
	clan   string
	apikey string
}

var mySettings settings

//var cocClient cocapi.Client

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes Routes

func init() {
	routes = make([]Route, 0)
	mySettings = settings{}
	mySettings.port = os.Getenv("COC_PORT")
	mySettings.clan = os.Getenv("COC_CLANTAG")
	mySettings.apikey = os.Getenv("COC_KEY")
	if mySettings.port == "" {
		mySettings.port = "8080"
	}
}

var db *bolt.DB

func main() {
	flag.StringVar(&mySettings.port, "port", mySettings.port, "Port to run service on")
	flag.StringVar(&mySettings.clan, "clan", mySettings.clan, "Clan tag to view")
	flag.StringVar(&mySettings.apikey, "apikey", mySettings.apikey, "API key to use")
	flag.Parse()
	fmt.Println("Starting", mySettings)
	myPath, _ := osext.ExecutableFolder()
	var err error
	db, err = bolt.Open(myPath+"my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if mySettings.apikey == "" && mySettings.clan == "" {
		resp := getPublicIP()
		type ip struct {
			Ip string `json:"ip"`
		}
		i := ip{}
		err := json.Unmarshal(resp, &i)
		if err != nil {
			log.Println(err)
		}
		log.Println("API key not set or clan tag not set, please see clashclient -h")
		log.Println("Your public ip adress is", i.Ip)
		log.Println("Create an account on https://developer.clashofclans.com to get an API key")
		return
	}
	initDb(mySettings.clan)
	updateClan()

	//return

	router := NewRouter()
	go func() {
		log.Fatal(http.ListenAndServe(":"+mySettings.port, router))
	}()
	log.Println("Webserver started")

	quit := make(chan struct{})
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				updateClan()
			}
			select {
			case <-quit:
				return
			}
		}
	}()

	// Wait for SIGINT and SIGTERM (HIT CTRL-C)
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	/*
		log.Println(<-ch)
		return
	*/

	var startErr error
	switch runtime.GOOS {
	case "linux":
		startErr = exec.Command("xdg-open", "http://localhost:"+mySettings.port).Start()
	case "windows", "darwin":
		startErr = exec.Command("open", "http://localhost:"+mySettings.port).Start()
	default:
		startErr = fmt.Errorf("unsupported platform")
	}
	if startErr != nil {
		log.Println(startErr)
	}

	log.Println(<-ch)

	close(quit)
	log.Println("Bye ;)")
}
