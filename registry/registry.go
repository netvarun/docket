package main

import (
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/codegangsta/martini"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

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

	file, header, err := r.FormFile("file")
	defer file.Close()

	if err != nil {
		return 500, "failed"
	}

	out, err := os.Create("/tmp/file")
	if err != nil {
		return 500, "also failed"
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		fmt.Fprintln(w, err)
	}

	// the header contains useful info, like the original file name
	//fmt.Fprintf(w, "File %s uploaded successfully.", header.Filename)
	fmt.Println("header = ", header.Filename)

	return http.StatusOK, "success"
}

func doTest(params martini.Params, w http.ResponseWriter) (int, string) {
	resource := strings.ToLower(params["resource"])
	w.Header().Set("Content-Type", "application/json")

	return http.StatusOK, resource
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
