//Docket Client
//Author: Sivamani Varun
//Gopher Gala
package main

//push
//pull
//-h[ost]
//-p[ort]

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/fsouza/go-dockerclient"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	host     = kingpin.Flag("host", "Set host of docket registry.").Short('h').Default("http://127.0.0.1").String()
	port     = kingpin.Flag("port", "Set port of docket registry.").Short('p').Default("8000").String()
	location = kingpin.Flag("location", "Set location to store torrents and tarballs.").Short('l').Default("/tmp/docket").String()

	push      = kingpin.Command("push", "Push to the docket registry.")
	pushImage = push.Arg("push", "Image to push.").Required().String()

	pull      = kingpin.Command("pull", "pull to the docket registry.")
	pullImage = pull.Arg("pull", "Image to pull.").Required().String()

	imagesCmd = kingpin.Command("images", "display images in the docket registry.")
	imageFlag = imagesCmd.Flag("images", "display images in the docket registry.").Bool()
)

// Creates a new tarball upload http request to the Docket registry
func uploadFile(params map[string]string, paramName, path string) (*http.Request, error) {
	uri := *host + ":" + *port + "/images"
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())
	for key, val := range params {
		fmt.Println("key = ", key, " val = ", val)
		request.Header.Add(key, val)
	}
	return request, nil
}

func applyPush(image string) error {
	reg, err := regexp.Compile("[^A-Za-z0-9]+")
	if err != nil {
		return err
	}

	loc := *location
	if _, err := os.Stat(loc); os.IsNotExist(err) {
		os.Mkdir(loc, 0644)
	}

	endpoint := "unix:///var/run/docker.sock"
	client, _ := docker.NewClient(endpoint)
	imgs, _ := client.ListImages(docker.ListImagesOptions{All: false})

	found := false
	imageId := ""
	filePath := ""
	created := ""

	for _, img := range imgs {
		tags := img.RepoTags
		for _, tag := range tags {
			if tag == image {
				found = true
				imageId = img.ID
				created = strconv.FormatInt(img.Created, 10)
				fmt.Println("Found image: ", image)
				fmt.Println("ID: ", img.ID)
				fmt.Println("RepoTags: ", img.RepoTags)
				fmt.Println("Created: ", img.Created)
				fmt.Println("Size: ", img.Size)
				fmt.Println("VirtualSize: ", img.VirtualSize)
				fmt.Println("ParentId: ", img.ParentID)
				safeImageName := reg.ReplaceAllString(image, "_")
				s := []string{loc, "/", imageId, "_", safeImageName, ".tar"}
				filePath = strings.Join(s, "")
				break
			}
		}
	}
	if !found {
		return errors.New("Sorry the image could not be found.")
	}

	//Run export command
	//command invocation
	//run docker command save to tar ball in location
	fmt.Println("Exporting image to tarball...")
	cmd := fmt.Sprintf("docker save %s > %s", image, filePath)
	_, err1 := exec.Command("sh", "-c", cmd).Output()
	if err1 != nil {
		return err1
	}

	fmt.Println("Successively exported tarball...")
	//make post request with contents of tarball to docket registry

	imageParams := map[string]string{
		"image":   image,
		"id":      imageId,
		"created": created,
	}

	//Adapted from http://matt.aimonetti.net/posts/2013/07/01/golang-multipart-file-upload-example/ (C) Matt Aimonetti
	request, err := uploadFile(imageParams, "file", filePath)
	if err != nil {
		log.Fatal(err)
	}
	uploadClient := &http.Client{}
	resp, err := uploadClient.Do(request)
	if err != nil {
		log.Fatal(err)
	} else {
		body := &bytes.Buffer{}
		_, err := body.ReadFrom(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != 200 {
			return errors.New("Failed to push image...")
		}
	}
	fmt.Println("Successfully uploaded image: ", image, " to the Docket registry.")
	return nil
}

//Adapted from https://github.com/thbar/golang-playground/blob/master/download-files.go
func downloadFromUrl(url string, fileName string) (err error) {
	output, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error while creating", fileName, "-", err)
		return err
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return err
	}
	if response.StatusCode != 200 {
		fmt.Println("Failed to pull image")
		return errors.New("Failed to pull image...")
	}
	defer response.Body.Close()

	n, err := io.Copy(output, response.Body)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return err
	}

	//fmt.Println(n, "bytes downloaded.")
	//Hack: trivial check to ensure if file downloaded is not too small
	if n < 100 {
		return errors.New("Failed to pull image...")
	}
	return nil
}

func applyPull(image string) error {
	reg, err := regexp.Compile("[^A-Za-z0-9]+")
	if err != nil {
		return err
	}

	loc := *location
	if _, err := os.Stat(loc); os.IsNotExist(err) {
		os.Mkdir(loc, 0644)
	}

	safeImageName := reg.ReplaceAllString(image, "_")
	filePath := loc + "/"
	fileName := filePath + safeImageName + ".torrent"

	//Download torrent file
	queryParam := map[string]string{
		"image": image,
	}
	queryParamJson, _ := json.Marshal(queryParam)

	metaUrl := *host + ":" + *port + "/images?q=" + url.QueryEscape(string(queryParamJson))
	response, err3 := http.Get(metaUrl)
	if err3 != nil {
		fmt.Println("Failed to query image metadata endpoint")
		return err3
	}
	if response.StatusCode != 200 {
		fmt.Println("Failed to get image metadata")
		return errors.New("Failed to get images metadata...")
	}
	defer response.Body.Close()
	metaJson, err4 := ioutil.ReadAll(response.Body)
	if err4 != nil {
		fmt.Println("Failed to get image metadata json")
		return errors.New("Failed to get image metadata json")
	}

	var queryObj map[string]interface{}
	if err := json.Unmarshal([]byte(metaJson), &queryObj); err != nil {
		return errors.New("Failed to decode images metadata json...")
	}

	tarballNameInterface := queryObj["fileName"]
	tarballName := tarballNameInterface.(string)

	fmt.Println("Downloading the torrent file for image: ", image)

	url := *host + ":" + *port + "/torrents?q=" + url.QueryEscape(string(queryParamJson))
	err1 := downloadFromUrl(url, fileName)
	if err1 != nil {
		fmt.Println("Failed to pull image")
		return err
	}

	fmt.Println("Downloading from the torrent file...")
	ctorrentCmd := fmt.Sprintf("cd %s && ctorrent -e 0 %s", filePath, fileName)
	cmd := exec.Command("bash", "-c", ctorrentCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	//TODO:Replace filename with that from metadata
	tarballPath := filePath + tarballName
	fmt.Println("Tarball path = ", tarballPath)

	//Load the downloaded tarball
	fmt.Println("Exporting image to tarball...")
	importCmd := fmt.Sprintf("docker load -i %s", tarballPath)
	_, err2 := exec.Command("sh", "-c", importCmd).Output()
	if err2 != nil {
		fmt.Printf("Failed to load image into docker!")
		return err2
	}
	fmt.Printf("Successfively pulled image: ", image)
	return nil
}

func applyImages() error {
	imagesUrl := *host + ":" + *port + "/images/all"
	//TODO:Get metadata GET /images?q={"image":}
	response, err3 := http.Get(imagesUrl)
	if err3 != nil {
		fmt.Println("Failed to query images list endpoint")
		return err3
	}
	if response.StatusCode != 200 {
		fmt.Println("Failed to get images list")
		return errors.New("Failed to get images list...")
	}
	defer response.Body.Close()
	imagesList, err4 := ioutil.ReadAll(response.Body)
	if err4 != nil {
		fmt.Println("Failed to get images list")
		return errors.New("Failed to get images list")
	}

	fmt.Println(string(imagesList))

	return nil
}

func main() {
	kingpin.CommandLine.Help = "Docket Client"

	switch kingpin.Parse() {
	case "push":
		kingpin.FatalIfError(applyPush(*pushImage), "Pushing of image failed")
	case "pull":
		kingpin.FatalIfError(applyPull((*pullImage)), "Pulling of image failed")
	case "images":
		kingpin.FatalIfError(applyImages(), "Listing of images failed")
	}
}
