package easyssh

import (
	"crypto/sha1"
	"fmt"
	"os"
	"strings"
)

// Sha1 is a helper method of sha1 algo.
func Sha1(input string) string {
	hash := sha1.New()
	hash.Write([]byte(input))
	hashed := hash.Sum(nil)
	return fmt.Sprintf("%x", hashed)
}

// IsFileExists tests whether specified filepath is exists.
func IsFileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	if err == nil {
		return true
	}
	return os.IsExist(err)
}

// RemoveTrailingSlash removes trailing slash from a path.
func RemoveTrailingSlash(path string) string {
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		return path[:len(path)-1]
	}
	return path
}

// IsDir tests whether a path is a directory.
func IsDir(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		panic("error: " + err.Error())
	}

	mode := stat.Mode()
	return mode.IsDir()
}

// IsRegular is a helper method to test whether a path is a regular file.
func IsRegular(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		panic("error: " + err.Error())
	}

	mode := stat.Mode()
	return mode.IsRegular()
}
