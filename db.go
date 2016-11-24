package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
	"github.com/johansundell/cocapi"
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

func initDb() {
	err := db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket([]byte(mySettings.clan))
		_, err := tx.CreateBucketIfNotExists([]byte(mySettings.clan))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func saveMember(member Player) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(mySettings.clan))
		buf, err := json.Marshal(member)
		if err != nil {
			return err
		}
		return b.Put([]byte(member.Tag), buf)
	})
}

func getMember(tag string) (Player, error) {
	var member Player
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(mySettings.clan))
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

func getMembersFromDb() []Player {
	players := make([]Player, 0)
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(mySettings.clan))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			//fmt.Printf("key=%s, value=%s\n", k, v)
			var m Player
			json.Unmarshal(v, &m)
			players = append(players, m)
		}

		return nil
	})
	//fmt.Println(players)
	return players
}

func getSmallMembersFromDb() []cocapi.Member {
	members := make([]cocapi.Member, 0)
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(mySettings.clan))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			//fmt.Printf("key=%s, value=%s\n", k, v)
			var m cocapi.Member
			json.Unmarshal(v, &m)
			members = append(members, m)
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
