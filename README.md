# LSM-Tree Implementation in Go with Real-Time Visualization 🚀

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/doc/devel/release.html#go1.21)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/yourusername/lsm-tree)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![WebSocket](https://img.shields.io/badge/WebSocket-Enabled-4e9bcd.svg)](https://developer.mozilla.org/en-US/docs/Web/API/WebSocket)
[![BadgerDB](https://img.shields.io/badge/BadgerDB-v3.2103.5-7B42BC)](https://dgraph.io/docs/badger/)

A real-time visualization of a Log-Structured Merge Tree (LSM-Tree) implementation in Go, featuring an interactive web interface and WebSocket communication.

## 🌟 Features

- Log-Structured Merge Tree implementation in Go
- Real-time visualization using P5.js
- WebSocket-based communication
- Interactive web interface with TailwindCSS
- Docker support
- Persistent storage using BadgerDB
- Automatic compaction process

## 🔧 Prerequisites

- Go 1.21 or higher
- Docker (optional)
- Git

## 🚀 Quick Start

### Local Development

1. Clone the repository:
```bash
git clone https://github.com/yourusername/lsm-tree.git
cd lsm-tree
```

2. Install dependencies:
```bash
go mod download
```

3. Run the server:
```bash
go run cmd/server/main.go
```

### Docker Deployment

1. Build the Docker image:
```bash
docker build -t lsm-store .
```

2. Run the container:
```bash
docker run -p 8080:8080 -v $(pwd)/data:/app/data lsm-store
```

Visit `http://localhost:8080` in your browser to see the visualization.

## 🏗️ Architecture

### Core Components

- **LSM Tree**: Main data structure implementation
  - MemTable: In-memory buffer for recent writes
  - SSTable: Sorted String Table for persistent storage
  - Compaction: Background process for merging SSTables

- **Visualization**
  - Real-time WebSocket communication
  - Interactive web interface
  - Visual representation of data structures

## 🎯 API Usage

### WebSocket Messages

```javascript
// Set a key-value pair
{
    "type": "set",
    "key": "myKey",
    "value": "myValue",
    "timestamp": "2024-01-10T12:00:00Z"
}

// Get a value by key
{
    "type": "get",
    "key": "myKey"
}

// Delete a key
{
    "type": "delete",
    "key": "myKey"
}
```

## 🛠️ Development

### Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── store/
│   │   ├── lsm.go
│   │   ├── memtable.go
│   │   └── sstable.go
│   └── visualization/
│       └── websocket.go
├── web/
│   ├── static/
│   │   └── js/
│   └── templates/
├── Dockerfile
└── go.mod
```

### Running Tests

```bash
go test ./...
```

## 📝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [BadgerDB](https://github.com/dgraph-io/badger) for the underlying storage engine
- [P5.js](https://p5js.org/) for visualization capabilities
- [TailwindCSS](https://tailwindcss.com/) for the UI styling

## 🤝 Support

For support, email your-email@example.com or open an issue in the GitHub repository.
