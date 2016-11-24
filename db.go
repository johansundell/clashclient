package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

type ErrorType int

const (
	NotFound ErrorType = iota // 0
	Unknown                   // 1
)

type dbError struct {
	msg       string
	errorType ErrorType
}

func (e *dbError) Error() string {
	return e.msg
}

func initDb(tag string) {
	err := db.Update(func(tx *bolt.Tx) error {
		//tx.DeleteBucket([]byte(mySettings.clan))
		b, err := tx.CreateBucketIfNotExists([]byte(tag))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		_, err = b.CreateBucketIfNotExists([]byte("members"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func saveClan(clan Clan) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(clan.Tag))
		buf, err := json.Marshal(clan)
		if err != nil {
			return err
		}
		return b.Put([]byte(clan.Tag), buf)
	})
}

func getClan(tag string) (Clan, error) {
	var clan Clan
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(tag))
		v := b.Get([]byte(tag))
		if v == nil {
			//fmt.Println("not found")
			return &dbError{"Not found", NotFound}
		}
		json.Unmarshal(v, &clan)
		return nil
	})
	return clan, err
}

func saveMember(member Player, tag string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(tag))
		b = b.Bucket([]byte("members"))
		buf, err := json.Marshal(member)
		if err != nil {
			return err
		}
		return b.Put([]byte(member.Tag), buf)
	})
}

func getMemberFromDb(tag, clan string) (Player, error) {
	var member Player
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(clan))
		b = b.Bucket([]byte("members"))
		v := b.Get([]byte(tag))
		if v == nil {
			//fmt.Println("not found")
			return &dbError{"Not found", NotFound}
		}
		json.Unmarshal(v, &member)
		return nil
	})
	return member, err
}

func getMembersFromDb(clan string) []Player {
	players := make([]Player, 0)
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(clan))
		b = b.Bucket([]byte("members"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			//fmt.Printf("key=%s, value=%s\n", k, v)
			var m Player
			json.Unmarshal(v, &m)
			if m.Active {
				players = append(players, m)
			}
		}

		return nil
	})
	//fmt.Println(players)
	return players
}

func getSmallMembersFromDb(clan string) []SmallPlayer {
	members := make([]SmallPlayer, 0)
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(clan))
		b = b.Bucket([]byte("members"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			//fmt.Printf("key=%s, value=%s\n", k, v)
			var m SmallPlayer
			json.Unmarshal(v, &m)
			if m.Active {
				members = append(members, m)
			}
		}

		return nil
	})
	//fmt.Println(players)
	return members
}

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
