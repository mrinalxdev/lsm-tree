package store

import (
    // "encoding/json"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/dgraph-io/badger/v3"
)

type Entry struct {
    Key       string    `json:"key"`
    Value     string    `json:"value"`
    Timestamp time.Time `json:"timestamp"`
}

type LSMTree struct {
    memTable    *MemTable
    db          *badger.DB
    mu          sync.RWMutex
    subscribers []chan<- Entry
    sstables []*SSTable
    maxLevel int
}

func NewLSMTree(dataDir string) (*LSMTree, error) {
    opts := badger.DefaultOptions(dataDir)
    db, err := badger.Open(opts)
    if err != nil {
        return nil, fmt.Errorf("failed to open badger: %v", err)
    }

    return &LSMTree{
        memTable: NewMemTable(),
        db:       db,
    }, nil
}

func (l *LSMTree) Set(key, value string) error {
    l.mu.Lock()
    defer l.mu.Unlock()

    entry := Entry{
        Key:       key,
        Value:     value,
        Timestamp: time.Now(),
    }

    if err := l.memTable.Set(key, value); err != nil {
        return err
    }
    l.notifySubscribers(entry)

    // flushing to BadgerDB if memTable is full
    if l.memTable.Size() >= MemTableMaxSize {
        if err := l.flush(); err != nil {
            return err
        }
    }

    return nil
}

func (l *LSMTree) Get(key string) (string, error) {
    l.mu.RLock()
    defer l.mu.RUnlock()

    // Check memTable first
    if value, exists := l.memTable.Get(key); exists {
        return value, nil
    }

    // Check SSTables in reverse level order
    var result string
    err := l.db.View(func(txn *badger.Txn) error {
        for level := l.maxLevel; level >= 0; level-- {
            prefix := makeCompositeKey(level, key)
            item, err := txn.Get(prefix)
            if err == nil {
                return item.Value(func(val []byte) error {
                    result = string(val)
                    return nil
                })
            }
        }
        return badger.ErrKeyNotFound
    })

    if err == badger.ErrKeyNotFound {
        return "", fmt.Errorf("key not found")
    }
    if err != nil {
        return "", err
    }
    return result, nil
}

func (l *LSMTree) Delete(key string) error {
    l.mu.Lock()
    defer l.mu.Unlock()
    l.memTable.Delete(key)
    return l.db.Update(func(txn *badger.Txn) error {
        return txn.Delete([]byte(key))
    })
}

func (l *LSMTree) flush() error {
    l.mu.Lock()
    defer l.mu.Unlock()

    sst := NewSSTable(l.db, 0)
    if err := sst.Write(l.memTable.Entries()); err != nil {
        return fmt.Errorf("failed to write SSTable: %v", err)
    }

    l.sstables = append(l.sstables, sst)
    l.memTable = NewMemTable()

    go l.triggerCompaction()

    return nil
}

func (l *LSMTree) triggerCompaction() {
    l.mu.Lock()
    defer l.mu.Unlock()

    levelMap := make(map[int] []*SSTable)
    // grouping the sstables by level
    for _, sst := range l.sstables {
        levelMap[sst.level] = append(levelMap[sst.level], sst)
    }

    for level := 0; level < l.maxLevel; level ++ {
        tables := levelMap[level]
        if len(tables) >= 2 {
            sst1, sst2 := tables[0], tables[1]
            go func(s1, s2 *SSTable){
                newSST, err := s1.Compact(s2)
                if err != nil {
                    log.Printf("Compaction failed : %v", err)
                    return
                }
                l.mu.Lock()
                var newSSTables []*SSTable
                // removing the old sstables and adding new one
                for _, sst := range l.sstables {
                    if sst != s1 && sst != s2 {
                        newSSTables = append(newSSTables, sst)
                    }
                }
                newSSTables = append(newSSTables, newSST)
                l.sstables = newSSTables
                l.mu.Unlock()

                l.notifySubscribers(Entry{
                    Key: "compaction",
                    Value: fmt.Sprintf("Compacted level %d", level),
                    Timestamp: time.Now(),
                })
            }(sst1, sst2)
        }
    }
}

func (l *LSMTree) Subscribe(ch chan<- Entry) {
    l.mu.Lock()
    defer l.mu.Unlock()
    l.subscribers = append(l.subscribers, ch)
}

func (l *LSMTree) Unsubscribe(ch chan<- Entry) {
    l.mu.Lock()
    defer l.mu.Unlock()
    for i, sub := range l.subscribers {
        if sub == ch {
            l.subscribers = append(l.subscribers[:i], l.subscribers[i+1:]...)
            return
        }
    }
}

func (l *LSMTree) notifySubscribers(entry Entry) {
    for _, ch := range l.subscribers {
        select {
        case ch <- entry:
        default:
            log.Println("Subscriber channel is full")
        }
    }
}

func (l *LSMTree) Close() error {
    return l.db.Close()
}
