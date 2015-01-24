//Docket Client
//Author: Sivamani Varun
//Gopher Gala
package main

//push
//pull
//images
//-h[ost]
//-p[ort]

import (
	"fmt"
	"github.com/alecthomas/kingpin"
)

var (
	host = kingpin.Flag("host", "Set host of docket registry.").Short('h').Default("127.0.0.1").IP()
	port = kingpin.Flag("port", "Set port of docket registry.").Short('p').Default("9090").Int()

	push      = kingpin.Command("push", "Push to the docket registry.")
	pushFlag  = push.Flag("test", "Push flag").Bool()
	pushImage = push.Arg("push", "Image to push.").Required().String()
)

func applyPush(image string) error {
	fmt.Println("image = ", image)
	fmt.Println("host = ", host)
	fmt.Println("port = ", port)
	return nil
}

func main() {
	kingpin.CommandLine.Help = "Docket Client"
	switch kingpin.Parse() {
	case "push":
		kingpin.FatalIfError(applyPush((*pushImage)), "Pushing of image failed")
	}
}
