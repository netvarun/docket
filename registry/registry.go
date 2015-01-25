package main

import (
	"encoding/json"
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/codegangsta/martini"
	"github.com/jackpal/Taipei-Torrent/torrent"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

var (
	tracker  = kingpin.Flag("tracker", "Set host and port of bittorrent tracker. Example: -host 10.240.101.85:8940 Note: This cannot be set to localhost, since this is the tracker in which all the torrents will be created with. They have to be some accessible ip address from outside").Short('t').Default("10.240.101.85:8940").String()
	port     = kingpin.Flag("port", "Set port of docket registry.").Short('p').Default("8000").String()
	location = kingpin.Flag("location", "Set location to save torrents and docker images.").Short('l').Default("/var/local/docket").String()
)

// The one and only martini instance.
var store *Store
var m *martini.Martini

func init() {
	m = martini.New()
	// Setup routes
	r := martini.NewRouter()
	r.Post(`/images`, postImage)
	r.Get(`/torrents`, getTorrent)
	r.Get(`/images/all`, getImagesList)
	r.Get(`/images`, getImages)
	// Add the router action
	m.Action(r.Handle)
}

func postImage(w http.ResponseWriter, r *http.Request) (int, string) {
	w.Header().Set("Content-Type", "application/json")

	loc := *location
	fmt.Println("location, ", loc)

	// the FormFile function takes in the POST input id file
	file, header, err := r.FormFile("file")

	if err != nil {
		fmt.Println(err)
		return 500, "bad"
	}

	defer file.Close()

	//Get metadata
	image := r.Header.Get("image")
	id := r.Header.Get("id")
	created := r.Header.Get("created")
	fileName := header.Filename

	fmt.Println("Got image: ", image, " id = ", id, " created = ", created, " filename = ", fileName)

	s := []string{loc, "/", fileName}
	t := []string{loc, "/", fileName, ".torrent"}
	filePath := strings.Join(s, "")
	torrentFile := fileName + ".torrent"
	torrentPath := strings.Join(t, "")

	//JSON string of metadata
	imageMeta := map[string]string{
		"image":    image,
		"id":       id,
		"created":  created,
		"fileName": fileName,
	}
	imageMetaJson, _ := json.Marshal(imageMeta)

	out, err := os.Create(filePath)
	if err != nil {
		fmt.Println(err)
		return 500, "bad"
	}

	defer out.Close()

	// write the content from POST to the file
	_, err = io.Copy(out, file)
	if err != nil {
		fmt.Println(err)
		return 500, "bad"
	}

	fmt.Println("File uploaded successfully")

	btHost := *tracker
	fmt.Println("btHost = ", btHost)
	err = createTorrentFile(torrentPath, filePath, btHost)
	if err != nil {
		return 500, "torrent creation failed"
	}

	//Write to datastore
	err = writeToStore(store, "docket", image, string(imageMetaJson))
	if err != nil {
		fmt.Println("Error writing result: ", err)
	}

	//Seed the torrent
	fmt.Println("Seeding torrent in the background...")
	os.Chdir(loc)
	importCmd := fmt.Sprintf("ctorrent -d -e 9999 %s", torrentFile)
	_, err2 := exec.Command("sh", "-c", importCmd).Output()
	if err2 != nil {
		fmt.Printf("Failed to seed torrent..")
		fmt.Println(err2)
		return 500, "bad"
	}

	return http.StatusOK, "{\"status\":\"OK\"}"
}

func getTorrent(w http.ResponseWriter, r *http.Request) int {
	query := r.URL.Query()
	queryJson := query.Get("q")

	var queryObj map[string]interface{}
	if err := json.Unmarshal([]byte(queryJson), &queryObj); err != nil {
		return 500
	}

	imageInterface := queryObj["image"]
	image := imageInterface.(string)

	//Query db and find if image exists. If not throw error (done)
	jsonVal, err := getFromStore(store, "docket", image)
	if err != nil {
		fmt.Println("Error reading from file : %v\n", err)
		return 500
	}

	if jsonVal == "" {
		fmt.Println("Invalid image requested")
		return 500
	}

	//Unmarshall
	var imageObj map[string]interface{}
	if err := json.Unmarshal([]byte(jsonVal), &imageObj); err != nil {
		return 500
	}

	//find location to torrent
	torrentFileInterface := imageObj["fileName"]
	torrentFile := torrentFileInterface.(string) + ".torrent"

	torrentPath := *location + "/" + torrentFile
	//Check if file exists
	if _, err := os.Stat(torrentPath); os.IsNotExist(err) {
		fmt.Println("no such file or directory: %s", torrentPath)
		return 500
	}

	//set filepath to that
	file, err := ioutil.ReadFile(torrentPath)
	if err != nil {
		return 500
	}

	w.Header().Set("Content-Type", "application/x-bittorrent")
	if file != nil {
		w.Write(file)
		return http.StatusOK
	}

	return 500
}

func getImages(w http.ResponseWriter, r *http.Request) (int, string) {
	query := r.URL.Query()
	queryJson := query.Get("q")

	var queryObj map[string]interface{}
	if err := json.Unmarshal([]byte(queryJson), &queryObj); err != nil {
		return 500, ""
	}

	imageInterface := queryObj["image"]
	image := imageInterface.(string)

	fmt.Println("image = ", image)

	//Query db and find if image exists. If not throw error (done)
	jsonVal, err := getFromStore(store, "docket", image)
	if err != nil {
		fmt.Println("Error reading from file : %v\n", err)
		return 500, ""
	}

	if jsonVal == "" {
		fmt.Println("Invalid image requested")
		return 500, ""
	}

	w.Header().Set("Content-Type", "application/json")
	return http.StatusOK, jsonVal
}

func getImagesList(w http.ResponseWriter, r *http.Request) (int, string) {
	//Query db and find if image exists. If not throw error (done)
	keys, err := iterateStore(store, "docket")
	if err != nil {
		fmt.Println("Error reading from file : %v\n", err)
		return 500, ""
	}

	if keys == "" {
		fmt.Println("Invalid image requested")
		return 500, ""
	}

	w.Header().Set("Content-Type", "text/plain")
	return http.StatusOK, keys
}

func createTorrentFile(torrentFileName, root, announcePath string) (err error) {
	var metaInfo *torrent.MetaInfo
	metaInfo, err = torrent.CreateMetaInfoFromFileSystem(nil, root, 0, false)
	if err != nil {
		return
	}
	btHost := *tracker
	metaInfo.Announce = "http://" + btHost + "/announce"
	metaInfo.CreatedBy = "docket-registry"
	var torrentFile *os.File
	torrentFile, err = os.Create(torrentFileName)
	if err != nil {
		return
	}
	defer torrentFile.Close()
	err = metaInfo.Bencode(torrentFile)
	if err != nil {
		return
	}
	return
}

func main() {
	kingpin.CommandLine.Help = "Docket Registry"
	kingpin.Parse()

	loc := *location
	if _, err := os.Stat(loc); os.IsNotExist(err) {
		os.Mkdir(loc, 0644)
	}

	var storeErr error

	store, storeErr = openStore()
	if storeErr != nil {
		log.Fatal("Failed to open data store: %v", storeErr)
	}
	deferCloseStore(store)

	pString := ":" + *port

	if err := http.ListenAndServe(pString, m); err != nil {
		log.Fatal(err)
	}
}
