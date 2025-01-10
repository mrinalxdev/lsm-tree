package store

import (
    "bytes"
    "encoding/binary"
    "encoding/json"
    "fmt"
    // "io"
    // "path/filepath"
    "sort"
    "time"

    "github.com/dgraph-io/badger/v3"
)

type SSTableMetadata struct {
    Level     int       `json:"level"`
    Timestamp time.Time `json:"timestamp"`
    MinKey    string    `json:"min_key"`
    MaxKey    string    `json:"max_key"`
    Size      int       `json:"size"`
}

type SSTable struct {
    db       *badger.DB
    level    int
    metadata SSTableMetadata
}

func NewSSTable(db *badger.DB, level int) *SSTable {
    return &SSTable{
        db:    db,
        level: level,
        metadata: SSTableMetadata{
            Level:     level,
            Timestamp: time.Now(),
        },
    }
}

func (sst *SSTable) Write(entries map[string]string) error {
    if len(entries) == 0 {
        return nil
    }

    keys := make([]string, 0, len(entries))
    for k := range entries {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    sst.metadata.MinKey = keys[0]
    sst.metadata.MaxKey = keys[len(keys)-1]
    sst.metadata.Size = len(entries)

    // starting the db transaction
    txn := sst.db.NewTransaction(true)
    defer txn.Discard()

    for _, key := range keys {
        value := entries[key]
        compositeKey := makeCompositeKey(sst.level, key)

        if err := txn.Set(compositeKey, []byte(value)); err != nil {
            return fmt.Errorf("failed to write entry: %v", err)
        }
    }

    // writing the metadata
    metadataKey := makeMetadataKey(sst.level, sst.metadata.Timestamp)
    metadataBytes, err := json.Marshal(sst.metadata)
    if err != nil {
        return fmt.Errorf("failed to marshal metadata: %v", err)
    }

    if err := txn.Set(metadataKey, metadataBytes); err != nil {
        return fmt.Errorf("failed to write metadata: %v", err)
    }

    // commiting the transaction and a null check
    if err := txn.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %v", err)
    }

    return nil
}

func (sst *SSTable) Compact(other *SSTable) (*SSTable, error) {
    newLevel := max(sst.level, other.level) + 1
    newSST := NewSSTable(sst.db, newLevel)

    entries := make(map[string]string)

    readEntries := func(level int) error {
        return sst.db.View(func(txn *badger.Txn) error {
            prefix := makeMetadataKey(level, time.Time{})
            it := txn.NewIterator(badger.DefaultIteratorOptions)
            defer it.Close()

            for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
                item := it.Item()
                key := getKeyFromComposite(item.Key())

                err := item.Value(func(val []byte) error {
                    entries[key] = string(val)
                    return nil
                })
                if err != nil {
                    return err
                }
            }
            return nil
        })
    }

    if err := readEntries(sst.level); err != nil {
        return nil, err
    }
    if err := readEntries(other.level); err != nil {
        return nil, err
    }

    if err := newSST.Write(entries); err != nil {
        return nil, err
    }

    if err := sst.delete(); err != nil {
        return nil, err
    }
    if err := other.delete(); err != nil {
        return nil, err
    }

    return newSST, nil
}

func (sst *SSTable) delete() error {
    prefix := makeMetadataKey(sst.level, time.Time{})

    return sst.db.Update(func(txn *badger.Txn) error {
        it := txn.NewIterator(badger.DefaultIteratorOptions)
        defer it.Close()

        for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
            if err := txn.Delete(it.Item().Key()); err != nil {
                return err
            }
        }
        return nil
    })
}

func makeCompositeKey(level int, key string) []byte {
    var buf bytes.Buffer
    binary.Write(&buf, binary.BigEndian, uint32(level))
    buf.WriteString(key)
    return buf.Bytes()
}

func makeMetadataKey(level int, timestamp time.Time) []byte {
    var buf bytes.Buffer
    binary.Write(&buf, binary.BigEndian, uint32(level))
    binary.Write(&buf, binary.BigEndian, timestamp.UnixNano())
    return buf.Bytes()
}

func getKeyFromComposite(compositeKey []byte) string {
    return string(compositeKey[4:])
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}
