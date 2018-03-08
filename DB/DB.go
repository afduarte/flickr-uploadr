package DB

import (
	"github.com/coreos/bbolt"
	"errors"
)

var (
	DB        *bolt.DB
	JOBS      = []byte("jobs")
	CHECKSUMS = []byte("checksums")
)

func createBuckets(tx *bolt.Tx) error {
	buckets := [][]byte{
		JOBS,
		CHECKSUMS,
	}
	for _, name := range buckets {
		_, err := tx.CreateBucketIfNotExists(name)
		if err != nil {
			return err
		}
	}

	return nil
}

func Init() error {
	DB, dbErr := bolt.Open("store.db", 0600, nil)
	if dbErr != nil {
		return dbErr
	}
	// Create buckets
	return DB.Update(createBuckets)

}

func AddPhoto(id string, checksum string) error {
	err := DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(CHECKSUMS)
		err := b.Put([]byte(checksum), []byte(id))
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
