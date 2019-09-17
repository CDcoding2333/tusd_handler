package local

// fs是配合tusd使用的static服务

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tus/tusd"
)

// FileSystem ...
type FileSystem struct {
	http.FileSystem
	root    string
	indexes bool
}

// NewFileSystem ...
func NewFileSystem(root string, indexes bool) *FileSystem {
	return &FileSystem{
		FileSystem: gin.Dir(root, indexes),
		root:       root,
		indexes:    indexes,
	}
}

// Exists override http.FileSystem.Exist
func (l *FileSystem) Exists(prefix string, filepath string) bool {
	if p := strings.TrimPrefix(filepath, prefix); len(p) < len(filepath) {
		name := path.Join(l.root, p)
		stats, err := os.Stat(name)
		if err != nil {
			return false
		}
		if !l.indexes && stats.IsDir() {
			return false
		}
		return true
	}
	return false
}

// FileInfo ...
func (l *FileSystem) FileInfo(id string) (*tusd.FileInfo, error) {

	name := path.Join(l.root, id)
	dat, err := ioutil.ReadFile(name + ".info")
	if err != nil {
		return nil, err
	}
	info := &tusd.FileInfo{}
	if err = json.Unmarshal(dat, info); err != nil {
		return nil, err
	}

	return info, nil
}
