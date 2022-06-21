package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func FileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func isDir(path string) bool {
	f, err := os.Stat(path)
	if err != nil {
		return false
	}
	return f.IsDir()
}

// 最终方案-全兼容
func GetCurrentAbPath() string {
	dir := getCurrentAbPathByExecutable()
	tmpDir, _ := filepath.EvalSymlinks(os.TempDir())
	if strings.Contains(dir, tmpDir) {
		return getCurrentAbPathByCaller()
	}
	return dir
}

func GetWorkdir() string {
	return getCurrentAbPathByCaller()
}

// 获取当前执行文件绝对路径
func getCurrentAbPathByExecutable() string {
	exePath, _ := os.Executable()
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}

// 获取当前执行文件绝对路径（go run）
func getCurrentAbPathByCaller() string {
	dir, _ := os.Getwd()
	return dir
}
