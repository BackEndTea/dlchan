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

	for _, post := range posts {
		downloadFile(board, post.getFileName(), saveLocation)
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
		fmt.Println("Found a post with no file, skipping")
		return
	}

	url := "https://i.4cdn.org/" + board + "/" + filename
	fmt.Println("Downloading from " + url)
	filename = saveLocation + "/" + board + "/" + filename

	// If the file already exists, then we don't need to download it
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		fmt.Println("File already exists, not downloading")
		return
	}
	fmt.Println("Saving file to " + filename)

	// Create the file
	out, err := os.Create(filename)
	if err != nil {
		fmt.Println("Unable to create the file")
		panic(err)
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Unable to download the file")
		return
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println("Unable to write to the file")
		return
	}
	fmt.Println("Successfully downloaded file")
}
