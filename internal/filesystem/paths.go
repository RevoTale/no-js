package filesystem

import "os"
func PathExists(target string) bool {
	_, err := os.Stat(target)
	return err == nil
} 