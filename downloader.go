package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type thread struct {
	No int
}

type page struct {
	page    int
	Threads []thread
}

// 4chan JSON structure is somewhat 'special'
type posts struct {
	Posts []post
}

/*
 * The only types we care about, together they create the filename that we need
 * The resulting url would  be Images: https://i.4cdn.org/<board>/<tim><ext>
 * Note that the . is already part of the Ext, so don't add that again
 */
type post struct {
	Tim int
	Ext string
}

const concurrencyCount int = 100

var downloadedCount int
var alreadyDownloadedCount int

func (p post) getFileName() string {
	if p.Tim == 0 || p.Ext == "" {
		return ""
	}
	return strconv.Itoa(p.Tim) + p.Ext
}

func main() {
	board, saveLocation, thread := getFlags()

	if board == "" {
		fmt.Println("Please specify the board flag")
		os.Exit(1)
	}
	fmt.Println("Downloading from board", board)
	if thread != "" {
		fmt.Println("Downloading thread: ", thread)
	}
	fmt.Println("Saving to loaction", saveLocation)

	posts := getPosts(board, saveLocation, thread)

	createDirIfNotExist(saveLocation + "/" + board)
	downloadImageFromPosts(posts, board, saveLocation)
	fmt.Println("Downloaded", downloadedCount, "new images.")
	fmt.Println(alreadyDownloadedCount, "images were previously downloaded")
}

func downloadImageFromPosts(ps []post, board, saveLocation string) {
	sem := make(chan bool, concurrencyCount)
	for _, p := range ps {
		sem <- true
		go func(p post, board, saveLocation string) {
			defer func() { <-sem }()
			downloadFile(board, p.getFileName(), saveLocation)

		}(p, board, saveLocation)
	}
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
}

func getFlags() (string, string, string) {
	var board string
	flag.StringVar(&board, "board", "", "Board to download all images from")
	var saveLocation string
	flag.StringVar(&saveLocation, "out", ".", "Output directory to save images. A child directory with the board name will contain the images")

	var thread string
	flag.StringVar(&thread, "thread", "", "Specific thread to download (optional)")
	flag.Parse()

	saveLocation = strings.TrimSuffix(saveLocation, "/")

	return board, saveLocation, thread
}

func getPosts(board, saveLocation, thread string) []post {
	if thread != "" {
		return getThreadContent(board, thread)
	}
	threads := getThreads(board)
	var posts []post
	for _, thread := range threads {
		posts = append(posts, getThreadContent(board, strconv.Itoa(thread.No))...)
	}
	return posts
}
func getThreads(board string) []thread {
	url := "https://a.4cdn.org/" + board + "/threads.json"
	threadList := []byte(readURLl(url))

	var keys []page
	json.Unmarshal(threadList, &keys)

	var threads []thread
	for _, page := range keys {
		threads = append(threads, page.Threads...)
	}
	return threads
}

func getThreadContent(board string, t string) []post {
	fmt.Println("Gathering content from ", t, " thread")
	url := "https://a.4cdn.org/" + board + "/thread/" + t + ".json"
	body := []byte(readURLl(url))

	var key posts
	json.Unmarshal(body, &key)

	return key.Posts
}

func createDirIfNotExist(direcotry string) {
	if _, err := os.Stat(direcotry); !os.IsExist(err) {
		os.MkdirAll(direcotry, 0777)
	}
}

func readURLl(url string) string {
	resp, _ := http.Get(url)
	bytes, _ := ioutil.ReadAll(resp.Body)
	stringBody := string(bytes)
	defer resp.Body.Close()
	return stringBody
}

// downloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func downloadFile(board, filename, saveLocation string) {
	if filename == "" {
		return
	}

	url := "https://i.4cdn.org/" + board + "/" + filename
	fmt.Println("Downloading from " + url)
	filename = saveLocation + "/" + board + "/" + filename

	// If the file already exists, then we don't need to download it
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		fmt.Println(filename, " already exists, not downloading")
		alreadyDownloadedCount++
		return
	}

	// Create the file
	out, err := os.Create(filename)
	if err != nil {
		fmt.Println("Unable to create ", filename)
		panic(err)
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Unable to download from ", url)
		return
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println("Unable to write to ", filename)
		return
	}
	fmt.Println("Successfully downloaded ", filename)
	downloadedCount++
	return
}
