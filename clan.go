package main

import (
	"log"
	"math"
	"sort"
	"time"

	"github.com/johansundell/cocapi"
)

type Player struct {
	cocapi.Player
	Active           bool      `json:"active"`
	Created          time.Time `json:"created"`
	LastUpdated      time.Time `json:"lastUpdated"`
	Left             time.Time `json:"left"`
	ClanRank         int       `json:"clanRank"`
	PreviousClanRank int       `json:"previousClanRank"`
}

type Clan struct {
	cocapi.ClanInfo
	MemberList string `json:"memberList"`
}

type SmallPlayer struct {
	cocapi.Member
	WarStars   int  `json:"warStars"`
	Active     bool `json:"active"`
	AttackWins int  `json:"attackWins"`
}

type SortRank []SmallPlayer

func (c SortRank) Len() int {
	return len(c)
}

func (c SortRank) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c SortRank) Less(i, j int) bool {
	return c[i].ClanRank < c[j].ClanRank
}

type SortDonationRatio []SmallPlayer

func (c SortDonationRatio) Len() int {
	return len(c)
}

func (c SortDonationRatio) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c SortDonationRatio) Less(i, j int) bool {
	var d = make([]float64, 4)
	d[0], d[1], d[2], d[3] = float64(c[i].Donations), float64(c[i].DonationsReceived), float64(c[j].Donations), float64(c[j].DonationsReceived)
	for k, n := range d {
		if n == 0 {
			d[k] = math.SmallestNonzeroFloat64
		}
	}
	if (d[0] / d[1]) == (d[2] / d[3]) {
		return d[0] > d[2]
	}
	return (d[0] / d[1]) > (d[2] / d[3])
}

type SortRoles []SmallPlayer

func (c SortRoles) Len() int {
	return len(c)
}

func (c SortRoles) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c SortRoles) Less(i, j int) bool {
	if c[i].Role == c[j].Role {
		return c[i].ExpLevel > c[j].ExpLevel
	}
	roles := map[string]int{
		"leader":   4,
		"coLeader": 3,
		"admin":    2,
		"member":   1,
	}
	return roles[c[i].Role] > roles[c[j].Role]
}

type SortWarStars []SmallPlayer

func (c SortWarStars) Len() int {
	return len(c)
}

func (c SortWarStars) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c SortWarStars) Less(i, j int) bool {
	return c[i].WarStars > c[j].WarStars
}

type SortAttackWins []SmallPlayer

func (c SortAttackWins) Len() int {
	return len(c)
}

func (c SortAttackWins) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c SortAttackWins) Less(i, j int) bool {
	return c[i].AttackWins > c[j].AttackWins
}

func getMembers(clanTag, sortDir string) []SmallPlayer {
	members := getSmallMembersFromDb(clanTag)

	switch sortDir {
	case "rate":
		sort.Sort(SortDonationRatio(members))
		break
	case "role":
		sort.Sort(SortRoles(members))
		break
	case "warstars":
		sort.Sort(SortWarStars(members))
		break
	case "attackwins":
		sort.Sort(SortAttackWins(members))
		break
	default:
		sort.Sort(SortRank(members))
		break
	}

	return members
}

func updateClan() error {
	client := cocapi.NewClient(mySettings.apikey)
	clan, err := client.GetClanInfo(mySettings.clan)
	if err != nil {
		return err
	}
	myClan := Clan{clan, ""}
	saveClan(myClan)
	currentMembers := make(map[string]Player)
	for _, row := range clan.MemberList {
		p, err := client.GetPlayerInfo(row.Tag)
		if err != nil {
			log.Println(err)
			continue
		}
		member, err := getMemberFromDb(row.Tag, mySettings.clan)
		switch t := err.(type) {
		case *dbError:
			if t.errorType == NotFound {
				member = Player{p, true, time.Now(), time.Now(), time.Time{}, row.ClanRank, row.PreviousClanRank}
			}
			break
		default:
			member = Player{p, true, member.Created, time.Now(), time.Time{}, row.ClanRank, row.PreviousClanRank}
			break
		}

		//fmt.Println(member)
		if err := saveMember(member, mySettings.clan); err != nil {
			log.Println(err)
		}
		currentMembers[member.Tag] = member
		//time.Sleep(250 * time.Millisecond)
	}
	log.Println("Saved current members to db")
	oldMembers := getMembersFromDb(mySettings.clan)
	for _, row := range oldMembers {
		if m, ok := currentMembers[row.Tag]; !ok {
			m.Active = false
			m.Left = time.Now()
			if err := saveMember(m, mySettings.clan); err != nil {
				log.Println(err)
			}
		}
	}

	return nil
}
