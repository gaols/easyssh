package easyssh

import (
	"fmt"
	"crypto/sha1"
	"os"
	"strings"
)

func Sha1(input string) string {
	hash := sha1.New()
	hash.Write([]byte(input))
	hashed := hash.Sum(nil)
	return fmt.Sprintf("%x", hashed)
}

func IsFileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	if err == nil {
		return true
	}
	return os.IsExist(err)
}

func RemoveTrailingSlash(path string) string {
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		return path[:len(path)-1]
	}
	return path
}

func IsDir(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		panic("error: " + err.Error())
	}

	mode := stat.Mode()
	return mode.IsDir()
}

func IsRegular(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		panic("error: " + err.Error())
	}

	mode := stat.Mode()
	return mode.IsRegular()
}
