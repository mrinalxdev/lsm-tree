package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/websocket"
	"github.com/mrinalxdev/lsm-tree/internal/store"
	"github.com/mrinalxdev/lsm-tree/internal/visualization"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // For development
    },
}

func main() {
    dataDir := flag.String("data-dir", "./data", "Directory for LSM tree data")
    flag.Parse()

    // Ensure data directory exists
    if err := os.MkdirAll(*dataDir, 0755); err != nil {
        log.Fatalf("Failed to create data directory: %v", err)
    }

    // Initialize LSM tree
    lsm, err := store.NewLSMTree(*dataDir)
    if err != nil {
        log.Fatalf("Failed to create LSM tree: %v", err)
    }
    defer lsm.Close()

    // Initialize visualization hub
    hub := visualization.NewHub(lsm)
    go hub.Run()

    // Static file server
    fs := http.FileServer(http.Dir("web/static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))

    // WebSocket handler
    http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        conn, err := upgrader.Upgrade(w, r, nil)
        if err != nil {
            log.Printf("Failed to upgrade connection: %v", err)
            return
        }
        visualization.ServeWs(hub, conn)
    })

    // Serve index.html
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, filepath.Join("web", "templates", "index.html"))
    })

    log.Println("Server starting on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}
