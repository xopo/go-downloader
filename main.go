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

func main() {
	header := make(map[string]string)
	header["User-Agent"] = "Chrome/83.0.4103.61 Safari/537.36"

	entries := getEntries()
	fmt.Printf("Found %d entries to download\n", len(entries))
	for _, link := range entries {
		if !checkResultExists(link.Title) {
			m3u8file := strings.Replace(link.Title, "mp4", "m3u8", 1)
			if checkResultExists(m3u8file) {
				fmt.Println("Converting --- " + m3u8file)
				convertFile(m3u8file)
			} else {
				fmt.Println("Downloading --- " + link.Title)
				downloader := hlsdl.New(link.Link, header, destination, m3u8file, 64, true)
				fmt.Printf("start downloading: %s\n", link.Title)
				filepath, err := downloader.Download()
				if err != nil {
					fmt.Printf("problem downloading %s : %s\n", link.Title, err)
				}
				fmt.Println("Finished: ", filepath)
				convertFile(m3u8file)
			}
		}
	}

	fmt.Println()
	fmt.Println()
	fmt.Println()
	cleanUp()
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
		if err != nil {
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

func convertFile(file string) {
	convertedfile := path.Join(destination, strings.Replace(file, "m3u8", "mp4", 1))
	filePath := path.Join(destination, file)

	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", convertedfile)
	fmt.Println()
	fmt.Println()

	// cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("converting \n%s \n%s\n\n", filePath, convertedfile)
	err := cmd.Run()
	fmt.Println("== end of conversion")
	fmt.Println()
	fmt.Println()

	if err != nil {
		log.Fatal("Error converting file", err)
	}
	if checkResultExists(convertedfile) {
		err := os.Remove(filePath)
		if err != nil {
			log.Fatalf("Cannot remove %s: %s ", filePath, err)
		}
	} else {
		fmt.Println("Cannot find convertedfile", filePath)
	}
}
