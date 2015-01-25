package main

import (
	"encoding/json"
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/codegangsta/martini"
	"github.com/jackpal/Taipei-Torrent/torrent"
	//"github.com/steveyen/gkvlite"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

var (
	host     = kingpin.Flag("host", "Set host of docket registry.").Short('h').Default("127.0.0.1").IP()
	port     = kingpin.Flag("port", "Set port of docket registry.").Short('p').Default("9090").Int()
	location = kingpin.Flag("location", "Set location to save torrents and docker images.").Short('l').Default("/tmp/dlds").String()
)

// The one and only martini instance.
var m *martini.Martini

func init() {
	m = martini.New()
	// Setup routes
	r := martini.NewRouter()
	r.Post(`/images`, postImage)
	r.Get(`/torrents`, getTorrent)
	//r.Get(`/images`, getImages)
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

	err = createTorrentFile(torrentPath, filePath, "10.240.101.85:8940")
	if err != nil {
		return 500, "torrent creation failed"
	}

	//Seed the torrent
	fmt.Println("Seeding torrent in the background...")
	os.Chdir(loc)
	importCmd := fmt.Sprintf("ctorrent -d -e 9999 %s", torrentFile)
	_, err1 := exec.Command("sh", "-c", importCmd).Output()
	if err1 != nil {
		fmt.Printf("Failed to seed torrent..")
		fmt.Println(err1)
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

	image := queryObj["image"]
	fmt.Println("image = ", image)
	//TODO:
	//Query db and find if image exists. If not throw error
	//If exists, find location to torrent
	//Check if file exists

	filepath := "/tmp/dlds/353b94eb357ddb343ebe054ccc80b49bb6d0828522e9f2eff313406363449d17_netvarun_test_latest.tar.torrent"
	file, err := ioutil.ReadFile(filepath)
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

func createTorrentFile(torrentFileName, root, announcePath string) (err error) {
	var metaInfo *torrent.MetaInfo
	metaInfo, err = torrent.CreateMetaInfoFromFileSystem(nil, root, 0, false)
	if err != nil {
		return
	}
	metaInfo.Announce = "http://10.240.101.85:8940/announce"
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

/*
- POST /images
receives data and writes to file
generates torrent and saves to file
- GET /torrents?q={"image":}
Retrieve the torrent file
- GET /images
List out all images, metadata and torrent file
*/

func main() {
	kingpin.CommandLine.Help = "Docket Registry"
	kingpin.Parse()

	if err := http.ListenAndServe(":8000", m); err != nil {
		log.Fatal(err)
	}
}
