package model

import (
	"github.com/coreos/bbolt"
	"errors"
	"encoding/json"
	"encoding/binary"
	"time"
)

var (
	DB        *bolt.DB
	JOBS      = []byte("jobs")
	CHECKSUMS = []byte("checksums")
	PHOTOS    = []byte("photos")
)

type Photo struct {
	ID          string   `json:"id"`
	Checksum    string   `json:"checksum"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

func createBuckets(tx *bolt.Tx) error {
	buckets := [][]byte{
		JOBS,
		CHECKSUMS,
		PHOTOS,
	}
	for _, name := range buckets {
		_, err := tx.CreateBucketIfNotExists(name)
		if err != nil {
			return err
		}
	}
	return nil
}

func InitDB() error {
	var dbErr error
	DB, dbErr = bolt.Open("store.db", 0600, nil)
	if dbErr != nil {
		return dbErr
	}
	// Create buckets
	return DB.Update(createBuckets)
}

func AddPhoto(photo *Photo) error {
	err := DB.Update(func(tx *bolt.Tx) error {
		c := tx.Bucket(CHECKSUMS)
		err := c.Put([]byte(photo.Checksum), []byte(photo.ID))
		p := tx.Bucket(PHOTOS)
		serialised, jsonErr := json.Marshal(*photo)
		if jsonErr != nil {
			return jsonErr
		}
		p.Put([]byte(photo.ID), serialised)
		return err
	})

	if err != nil {
		return errors.New("failed Adding photo: " + err.Error())
	}
	return nil
}

func AddJob(what string, when int64) error {
	err := DB.Update(func(tx *bolt.Tx) error {
		j := tx.Bucket(JOBS)
		// convert int64 to []byte
		buf := make([]byte, binary.MaxVarintLen64)
		n := binary.PutVarint(buf, when)
		b := buf[:n]

		var photoIDs []byte
		var jsonErr error
		var ids []string
		existing := j.Get(b)
		// If the id is NOT nil, the key exists, so we add an id to the array,
		if existing != nil {
			jsonErr = json.Unmarshal(existing, &ids)
			if jsonErr != nil {
				return jsonErr
			}
			ids = append(ids, what)
		} else {
			ids = []string{what}
		}
		photoIDs, jsonErr = json.Marshal(ids)
		if jsonErr != nil {
			return jsonErr
		}

		err := j.Put([]byte(b), photoIDs)

		return err
	})

	if err != nil {
		return errors.New("failed Adding photo: " + err.Error())
	}
	return nil
}

func CheckPhotoExists(checksum string) bool {
	exists := false
	DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(CHECKSUMS)
		id := b.Get([]byte(checksum))
		// If the id is NOT nil, the key exists, so we set the variable to true
		if id != nil {
			exists = true
		}
		return nil
	})
	return exists
}

func GetJobsDue() [][]byte {
	due := make([][]byte, 0)
	DB.View(func(tx *bolt.Tx) error {
		// Assume our events bucket exists and has RFC3339 encoded time keys.
		c := tx.Bucket(JOBS).Cursor()

		now := time.Now().Unix()
		buf := make([]byte, binary.MaxVarintLen64)
		n := binary.PutVarint(buf, now)
		b := buf[:n]
		// Iterate over the 90's.
		for k, v := c.Seek(b); k != nil; k, v = c.Prev() {
			due = append(due, v)
		}

		return nil
	})
	return due
}

func GetJob(id []byte) *Photo {
	var ph Photo
	DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(PHOTOS)
		jobStr := b.Get(id)
		// If the id is nil, the photo doesn't exist
		if jobStr == nil {
			// Early return because not found
			return nil
		}
		json.Unmarshal(jobStr, &ph)
		return nil
	})
	return &ph
}
