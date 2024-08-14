package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/canhlinh/hlsdl"
)

const destination = "result"

func titleAsm3u(title string) string {
	return title + ".m3u8"
}

func titleAsmp4(title string) string {
	return title + ".mp4"
}

func main() {
	entries := getEntries()
	downloadChannel := make(chan VideoEntry)
	convertedChannel := make(chan string)
	cleanupChannel := make(chan VideoEntry)

	checkDestinationDir()
	fmt.Printf("Found %d entries to download\n", len(entries))
	filesToDownload := getNewFiles(entries)
	if len(filesToDownload) > 0 {
		go func() {
			getFiles(filesToDownload, downloadChannel)
		}()
	}

	go func() {
		convertFile(downloadChannel, convertedChannel, cleanupChannel)
	}()

	go func() {
		cleanUpFiles(cleanupChannel)
	}()

	for msg := range convertedChannel {
		fmt.Println(msg)
	}
	cleanUp()
}

func getNewFiles(entries []VideoEntry) []VideoEntry {
	newFiles := []VideoEntry{}
	for _, entry := range entries {
		if !checkResultExists(titleAsm3u(entry.Title)) && !checkResultExists(titleAsmp4(entry.Title)) {
			newFiles = append(newFiles, entry)
		}
	}
	return newFiles
}

func getDownloadedVideos(entries []VideoEntry) []VideoEntry {
	newFiles := []VideoEntry{}
	for _, entry := range entries {
		if !checkResultExists(titleAsmp4(entry.Title)) {
			newFiles = append(newFiles, entry)
		}
	}
	return newFiles
}

func getFiles(files []VideoEntry, downChannel chan<- VideoEntry) {
	for _, file := range files {
		header := make(map[string]string)
		header["User-Agent"] = "Chrome/83.0.4103.61 Safari/537.36"

		fmt.Printf("start downloading: %s\n", file.Title)
		target := titleAsm3u(file.Title)
		downloader := hlsdl.New(file.Link, header, destination, target, 64, true)
		filepath, err := downloader.Download()
		if err != nil {
			log.Fatalf("problem downloading %s : %s\n", file.Title, err)
		}
		fmt.Println("Finished: ", filepath)
		downChannel <- file
	}
	close(downChannel)
}

func getNumberOfExistingMp4(entries []VideoEntry) int {
	var result []string
	for _, entry := range entries {
		mp4 := titleAsmp4(entry.Title)
		if checkResultExists(mp4) {
			result = append(result, entry.Title)
		}
	}
	return len(result)
}

func checkDestinationDir() {
	res, err := os.Stat(destination)
	if err != nil {
		os.Mkdir(destination, 0755)
		return
	}
	if !res.IsDir() {
		log.Fatalf("Destination cannot be created, pls create a %q dir for results", destination)
	}
}

type VideoEntry struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

func getEntries() []VideoEntry {
	filename := "./entries.json"
	file, err := os.Stat(filename)
	if err != nil {
		log.Fatal("Cannot get filestuff", err)
	}
	if !file.IsDir() {
		content, err := os.ReadFile(filename)
		validJson := json.Valid(content)
		if err != nil && validJson {
			log.Fatal("Reading imposible", err)
		}
		var entries []VideoEntry
		err = json.Unmarshal(content, &entries)
		if err != nil {
			log.Fatal("Cannot unmarshal content", err)
		}
		return entries
	}
	return nil
}

func checkResultExists(file string) bool {
	filePath := path.Join(destination, file)
	_, err := os.Stat(filePath)
	return err == nil
}

func cleanUpFiles(cleanupChannel <-chan VideoEntry) {
	for file := range cleanupChannel {
		fileWithPath := path.Join(destination, titleAsm3u(file.Title))
		_, err := os.Stat(fileWithPath)
		if err != nil {
			log.Fatal(err)
			return
		}
		os.Remove(fileWithPath)
	}
}

func cleanUp() {
	fmt.Println("Clean old file")
	dirEntries, err := os.ReadDir(destination)
	if err != nil {
		fmt.Println("Error reading directory", err)
	}

	fmt.Printf("Found %d files\n", len(dirEntries))
	for _, entry := range dirEntries {
		if !entry.IsDir() {
			file := entry.Name()
			fileExt := filepath.Ext(file)
			if fileExt == ".mp4" {
				continue
			}
			name := strings.Replace(filepath.Base(file), fileExt, "", 1)
			foundConverted := false
			for _, checkfile := range dirEntries {
				converted := fmt.Sprintf("%s%s", name, ".mp4")
				if checkfile.Name() == converted {
					foundConverted = true
					break
				}
			}
			if foundConverted {
				toRemove := path.Join(destination, file)
				fmt.Println("Removing file: " + toRemove)
				err := os.Remove(toRemove)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}

func convertFile(downloadChannel <-chan VideoEntry, convertedChannel chan<- string, cleanupChannel chan<- VideoEntry) {
	for file := range downloadChannel {
		mp4 := path.Join(destination, titleAsmp4(file.Title))
		m3u := path.Join(destination, titleAsm3u(file.Title))

		cmd := exec.Command("ffmpeg", "-i", m3u, "-c", "copy", mp4)
		fmt.Println()
		fmt.Println()

		// cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		convertedChannel <- fmt.Sprintf("converting \n%s \n%s\n\n", m3u, mp4)
		err := cmd.Run()
		convertedChannel <- fmt.Sprintln("== end of conversion")

		if err != nil {
			log.Fatalf("Error converting file %q", err)
		}

		cleanupChannel <- file
	}
	close(convertedChannel)
	close(cleanupChannel)
}
