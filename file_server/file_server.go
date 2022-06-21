package file_server

import (
	"os"
	"path/filepath"
)

type ServerNode struct {
	NodeName string `json:"node_name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DirSize  int64  `json:"dir_size"`
}

//todo 迁移走

func DirSizeB(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}
