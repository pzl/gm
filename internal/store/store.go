package store

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	badger "github.com/dgraph-io/badger/v2"
	jsoniter "github.com/json-iterator/go"
)

var jsCfg = jsoniter.Config{
	EscapeHTML:                    true,
	SortMapKeys:                   false,
	ValidateJsonRawMessage:        false,
	MarshalFloatWith6Digits:       true,
	ObjectFieldMustBeSimpleString: false,
}.Froze()

type StorageKey byte

const (
	SvcKey StorageKey = 's'
)

// composes the indexing Key
func MakeKey(sk StorageKey, id string) []byte {
	key := []byte(id)
	key = append([]byte{byte(sk)}, key...)
	return key
}

// UserMeta flags

type UMField byte

const (
//FIELDNAME UMField = 1 << iota
// ...
)

func (u UMField) Set(flag UMField) UMField    { return u | flag }
func (u UMField) Clear(flag UMField) UMField  { return u &^ flag }
func (u UMField) Toggle(flag UMField) UMField { return u ^ flag }
func (u UMField) Has(flag UMField) bool       { return u&flag != 0 }

// fetches all entries for the storage type
func GetAllForType(db *badger.DB, sk StorageKey) ([][]byte, error) {
	total := make([][]byte, 0, 20)

	pfx := []byte{byte(sk)}
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = true
	opts.Prefix = pfx
	err := db.View(func(tx *badger.Txn) error {
		it := tx.NewIterator(opts)
		defer it.Close()
		for it.Seek(pfx); it.ValidForPrefix(pfx); it.Next() {
			if err := it.Item().Value(func(v []byte) error {
				buf := make([]byte, len(v))
				copy(buf, v)
				total = append(total, buf)
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	})

	return total, err
}

func GetAllForPrefix(db *badger.DB, sk StorageKey, pfx string) (map[string][]byte, error) {
	total := make(map[string][]byte, 20)

	pfxKey := MakeKey(sk, pfx)
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = true
	opts.Prefix = pfxKey
	err := db.View(func(tx *badger.Txn) error {
		it := tx.NewIterator(opts)
		defer it.Close()
		for it.Seek(pfxKey); it.ValidForPrefix(pfxKey); it.Next() {
			if err := it.Item().Value(func(v []byte) error {
				keybuf := it.Item().Key()
				k := string(keybuf[len(pfxKey):])
				buf := make([]byte, len(v))
				copy(buf, v)
				total[k] = buf
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	})

	return total, err
}

// lists sub-entries found (as above) without getting the contents
func ListAllForPrefix(db *badger.DB, sk StorageKey, id string) ([]string, error) {
	total := make([]string, 0, 20)

	pfx := MakeKey(sk, id)
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	opts.Prefix = pfx
	err := db.View(func(tx *badger.Txn) error {
		it := tx.NewIterator(opts)
		defer it.Close()
		for it.Seek(pfx); it.ValidForPrefix(pfx); it.Next() {
			keybuf := it.Item().Key()
			k := string(keybuf[len(pfx):])
			total = append(total, string(k))
		}
		return nil
	})

	return total, err
}
func WriteType(db *badger.DB, sk StorageKey, id string, item interface{}, u UMField, ttl int64) error {
	buf, err := jsCfg.Marshal(item)
	if err != nil {
		return err
	}
	return WriteBytes(db, sk, id, buf, u, ttl)
}

func WriteBytes(db *badger.DB, sk StorageKey, id string, buf []byte, u UMField, ttl int64) error {
	key := MakeKey(sk, id)
	entry := badger.NewEntry(key, buf).WithMeta(byte(u))
	if ttl > 0 {
		entry = entry.WithTTL(time.Duration(ttl) * time.Second)
	}
	return db.Update(func(tx *badger.Txn) error {
		return tx.SetEntry(entry)
	})
}

func BatchWrite(wb *badger.WriteBatch, sk StorageKey, id string, buf []byte, u UMField, ttl int64) error {
	key := MakeKey(sk, id)
	entry := badger.NewEntry(key, buf).WithMeta(byte(u))
	if ttl > 0 {
		entry = entry.WithTTL(time.Duration(ttl) * time.Second)
	}
	return wb.SetEntry(entry)
}

func _getOne(db *badger.DB, sk StorageKey, id string, cb func([]byte) error) error {
	key := MakeKey(sk, id)

	err := db.View(func(tx *badger.Txn) error {
		item, err := tx.Get(key)
		if err != nil {
			return err
		}
		err = item.Value(cb)
		if err != nil {
			return err
		}
		return nil
	})

	return err
}

func GetOneBytes(db *badger.DB, sk StorageKey, id string) ([]byte, error) {
	var buf []byte
	err := _getOne(db, sk, id, func(b []byte) error {
		buf = make([]byte, len(b))
		copy(buf, b)
		return nil
	})
	return buf, err
}

func GetOne(db *badger.DB, sk StorageKey, id string, t interface{}) error {
	return _getOne(db, sk, id, func(b []byte) error {
		return jsCfg.Unmarshal(b, t)
	})
}

func GetMeta(db *badger.DB, sk StorageKey, id string) (UMField, error) {
	var u UMField
	key := MakeKey(sk, id)
	return u, db.View(func(tx *badger.Txn) error {
		item, err := tx.Get(key)
		if err != nil {
			return err
		}
		u = UMField(item.UserMeta())
		return nil
	})
}

// deletes a single record
func DeleteRecord(db *badger.DB, sk StorageKey, id string) error {
	key := MakeKey(sk, id)
	return db.Update(func(tx *badger.Txn) error {
		return tx.Delete(key)
	})
}

type Info struct {
	ID   string
	Meta UMField
}

func ListInfo(db *badger.DB, sk StorageKey) ([]Info, error) {
	total := make([]Info, 0, 20)

	pfx := []byte{byte(sk)}
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	opts.Prefix = pfx
	err := db.View(func(tx *badger.Txn) error {
		it := tx.NewIterator(opts)
		defer it.Close()
		for it.Seek(pfx); it.ValidForPrefix(pfx); it.Next() {
			total = append(total, Info{
				ID:   string(it.Item().KeyCopy(nil)[1:]),
				Meta: UMField(it.Item().UserMeta()),
			})
		}
		return nil
	})

	return total, err
}

func WriteBatch(d *badger.DB) *badger.WriteBatch {
	return d.NewWriteBatch()
}

func NewID() string {
	buf := make([]byte, 4)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
