package logcleaner

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"
)

// CleanLog every timeout clean log file
func CleanLog(
	timeBetweenCleanning time.Duration,
	filePath string,
	linesToRemove int,
) {
	timer := time.NewTimer(10 * time.Minute)
	defer timer.Stop()
	<-timer.C

	ticker := time.NewTicker(timeBetweenCleanning)
	for ; ; <-ticker.C {
		file, err := os.Open(filePath)
		if err != nil {
			log.Fatal("Log file is not found in environment")
		}
		defer file.Close()

		tmpFilename := filePath + ".tmp"
		tmpFile, err := os.Create(tmpFilename)
		if err != nil {
			log.Println("Can't create temporary file:", err)
			return
		}
		defer os.Remove(tmpFilename)
		defer tmpFile.Close()

		scanner := bufio.NewScanner(file)
		for i := 0; i < linesToRemove; i++ {
			if !scanner.Scan() {
				break
			}
		}
		for scanner.Scan() {
			if _, err := tmpFile.WriteString(scanner.Text() + "\n"); err != nil {
				fmt.Println("Can't write to temporary file:", err)
				return
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Can't read  source file:", err)
			return
		}
		if err := os.Rename(tmpFilename, filePath); err != nil {
			fmt.Println("Can't rename temporary file:", err)
			return
		}

		log.Println("First", linesToRemove, "lines was successfully removed from: ", filePath)
	}
}
