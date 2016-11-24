package main

import (
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
	"github.com/johansundell/cocapi"
	"github.com/kardianos/osext"
)

type settings struct {
	port   string
	clan   string
	apikey string
}

var mySettings settings

var cocClient cocapi.Client

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
}

var db *bolt.DB

type Player struct {
	cocapi.Player
	Active      bool      `json:"active"`
	Created     time.Time `json:"created"`
	LastUpdated time.Time `json:"lastUpdated"`
	Left        time.Time `json:"left"`
}

func main() {
	flag.StringVar(&mySettings.port, "port", mySettings.port, "Port to run service on")
	flag.Parse()
	fmt.Println("Starting", mySettings)
	myPath, _ := osext.ExecutableFolder()
	var err error
	db, err = bolt.Open(myPath+"my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	initDb()
	cocClient = cocapi.NewClient(mySettings.apikey)

	updateClan()

	return

	router := NewRouter()
	go func() {

		//http.Handle("/tmpl/", http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: ""}))
		//log.Fatal(http.ListenAndServe(":8080", nil))
		log.Fatal(http.ListenAndServe(":"+mySettings.port, router))
	}()

	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-quit:
				return
			}
		}
	}()

	// Wait for SIGINT and SIGTERM (HIT CTRL-C)
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)
	return

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

func updateClan() error {
	cocClient = cocapi.NewClient(mySettings.apikey)
	clan, err := cocClient.GetClanInfo(mySettings.clan)
	if err != nil {
		return err
	}
	currentMembers := make(map[string]Player)
	for _, row := range clan.MemberList {
		p, err := cocClient.GetPlayerInfo(row.Tag)
		if err != nil {
			log.Println(err)
			continue
		}
		member, err := getMember(row.Tag)
		switch t := err.(type) {
		case *dbError:
			if t.errorType == NotFound {
				member = Player{p, true, time.Now(), time.Now(), time.Time{}}
			}
			break
		default:
			member = Player{p, true, member.Created, time.Now(), time.Time{}}
			break
		}

		//fmt.Println(member)
		if err := saveMember(member); err != nil {
			log.Println(err)
		}
		currentMembers[member.Tag] = member
		//time.Sleep(250 * time.Millisecond)
	}
	log.Println("Saved current members to db")
	oldMembers := getMembersFromDb()
	for _, row := range oldMembers {
		if m, ok := currentMembers[row.Tag]; !ok {
			m.Active = false
			m.Left = time.Now()
			if err := saveMember(m); err != nil {
				log.Println(err)
			}
		}
	}

	return nil
}
