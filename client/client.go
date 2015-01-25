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
	"errors"
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/fsouza/go-dockerclient"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	host = kingpin.Flag("host", "Set host of docket registry.").Short('h').Default("http://127.0.0.1").String()
	port = kingpin.Flag("port", "Set port of docket registry.").Short('p').Default("8000").String()

	push      = kingpin.Command("push", "Push to the docket registry.")
	pushImage = push.Arg("push", "Image to push.").Required().String()

	pull      = kingpin.Command("pull", "pull to the docket registry.")
	pullImage = pull.Arg("pull", "Image to pull.").Required().String()
)

// Creates a new tarball upload http request to the Docket registry
func newfileUploadRequest(params map[string]string, paramName, path string) (*http.Request, error) {
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
				s := []string{"/tmp/", imageId, "_", safeImageName, ".tar"}
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
	//run docker command save to tar ball in /tmp
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

	request, err := newfileUploadRequest(imageParams, "file", filePath)
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

func applyPull(image string) error {
	fmt.Println("image = ", image)
	fmt.Println("host = ", host)
	fmt.Println("port = ", port)
	return nil
}

func main() {
	kingpin.CommandLine.Help = "Docket Client"

	switch kingpin.Parse() {
	case "push":
		kingpin.FatalIfError(applyPush(*pushImage), "Pushing of image failed")
	case "pull":
		kingpin.FatalIfError(applyPush((*pushImage)), "Pushing of image failed")
	}
}
