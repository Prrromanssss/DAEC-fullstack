package logcleaner

import (
	"bufio"
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
	timer := time.NewTimer(timeBetweenCleanning)
	defer timer.Stop()
	<-timer.C

	ticker := time.NewTicker(timeBetweenCleanning)
	for ; ; <-ticker.C {
		file, err := os.Open(filePath)
		if err != nil {
			log.Fatalf("Log file is not found in environment: %v", err)
		}
		defer file.Close()

		tmpFilename := filePath + ".tmp"
		tmpFile, err := os.Create(tmpFilename)
		if err != nil {
			log.Fatalf("Can't create temporary file: %v", err)
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
				log.Fatalf("Can't write to temporary file: %v", err)
				return
			}
		}

		if err := scanner.Err(); err != nil {
			log.Fatalf("Can't read source file: %v", err)
			return
		}
		if err := os.Rename(tmpFilename, filePath); err != nil {
			log.Fatalf("Can't rename temporary file: %v", err)
			return
		}

		log.Println("First", linesToRemove, "lines was successfully removed from: ", filePath)
	}
}
