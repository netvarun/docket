package main

import (
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/codegangsta/martini"
	"github.com/jackpal/Taipei-Torrent/torrent"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

//TODO:
//Store image metadata in some db

var (
	host = kingpin.Flag("host", "Set host of docket registry.").Short('h').Default("127.0.0.1").IP()
	port = kingpin.Flag("port", "Set port of docket registry.").Short('p').Default("9090").Int()
)

// The one and only martini instance.
var m *martini.Martini

func init() {
	m = martini.New()
	// Setup routes
	r := martini.NewRouter()
	r.Post(`/images`, postImage)
	r.Get(`/test/:resource`, doTest)
	//r.Post(`/torrents/:image`, getTorrent)
	//r.Get(`/images`, getImages)
	// Add the router action
	m.Action(r.Handle)
}

func postImage(w http.ResponseWriter, r *http.Request) (int, string) {
	w.Header().Set("Content-Type", "application/json")

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

	s := []string{"/tmp/dlds/", fileName}
	t := []string{"/tmp/dlds/", fileName, ".torrent"}
	filePath := strings.Join(s, "")
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

	return http.StatusOK, "success"
}

func doTest(params martini.Params, w http.ResponseWriter) (int, string) {
	resource := strings.ToLower(params["resource"])
	w.Header().Set("Content-Type", "application/json")

	return http.StatusOK, resource
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
	fmt.Println("Docket Registry")

	if err := http.ListenAndServe(":8000", m); err != nil {
		log.Fatal(err)
	}
}
