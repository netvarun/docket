![](http://static2.stuff.co.nz/1318390536/513/5775513_600x400.jpg)

# Docket

DOCKEr + torrenT = Docket

Docket is a custom docker registry that allows for deployments through bittorrent.

This was design and built in 48 hours as part of the Gopher Gala Golang 48 hour hackathon.
Hence kindly forgive me for the hackish code, and lack of tests.

[On hind sight - container flood or docker rush would have been a better name.]

[Screencast](https://asciinema.org/a/15752)

## Problem Statement

In 2015, git deploys are soon going to replace docker [or rocket or lxd] deploys.
Large scale deploys are going to choke your docker registry.

## Solution

The solution is BitTorrent. Any technology good for distributing music and movies, is usually good for doing server deploys.

## Features

- Written in Golang :)
- Very easy to use
- Works along side your private docker registry

## Concepts

Docket constitutes of 3 components:

1. Docket Registry

A REST service that acts as a registry. It receives docker image tarballs from the client, stores metadata into a database,
creates torrents out of them and seeds them.

2. Docket Client

The client itnerface in which the end user will be interacting. Can view available images in the registry, push an image to the registry and pull an image (which triggers a bittorrent deploy) from the registry.

3. Bittorrent Tracker

Docket allows you to BYOT (bring your own tracker), but I suggest installing opentracker

Docket uses ctorrent for the actual downloading of torrents.

## Installation

### Step 0: You need to already have Docker installed

### Step 1: Install ctorrent and opentracker

```bash
$ sudo apt-get update
$ sudo apt-get install zlib1g-dev make g++ ctorrent
$ git clone https://github.com/gebi/libowfat
$ make
$ git clone git://erdgeist.org/opentracker
$ cd opentracker
$ make
```

### Step 2: Download the Docket binaries

```bash
$ wget http://storage.googleapis.com/docket/docket.zip
$ unzip docket.zip
```

### Step 3: That's all


## Getting Started

### Step 1: Fire up the Tracker

```bash
$ cd opentracker
$ ./opentracker 8940
```

### Step 2: Fire up the Registry

```bash
$ cd registry
# Note you need to put an ip address in which other machines can contact the tracker
# You cannot put in localhost or 127.0.0.1
$ sudo ./registry --tracker "10.240.101.85:8940"
```

## Usage

### Push an image to the Registry

```bash
$ cd client
#Note: You will have to to explicitly mention the tag ":latest"
$ sudo ./client -h "http://10.240.101.85" push netvarun/test:latest
Found image:  netvarun/test:latest
ID:  353b94eb357ddb343ebe054ccc80b49bb6d0828522e9f2eff313406363449d17
RepoTags:  [netvarun/test:latest]
Created:  1422145581
Size:  0
VirtualSize:  188305056
ParentId:  d7d8be71d422a83c97849c4a8e124fcbe42170d5ce508f339ce52be9954dc3b4
Exporting image to tarball...
Successively exported tarball...
key =  image  val =  netvarun/test:latest
key =  id  val =  353b94eb357ddb343ebe054ccc80b49bb6d0828522e9f2eff313406363449d17
key =  created  val =  1422145581
Successfully uploaded image:  netvarun/test:latest  to the Docket registry.
```

### List all images available in the Registry

```bash
$ cd client
$ sudo ./client -h "http://10.240.101.85" images
netvarun/test:latest
perl:latest

```

### Do a deploy

```bash
$ cd client
$ sudo ./client -h "http://10.240.101.85" pull perl:latest
Downloading the torrent file for image:  perl:latest
Downloading from the torrent file...
META INFO
Announce: http://10.240.101.85:8940/announce
Piece length: 524288
Created with: docket-registry
FILES INFO
<1> 14f61693dd2db6380755a662d6e4e3583b5214fad9032bd983ce6c70df2144bc_perl_latest.tar [838467072]
Total: 799 MB
Found bit field file; verifying previous state.
Listening on 0.0.0.0:2706
Press 'h' or '?' for help (display/control client options).
Checking completed.
- 0/0/1 [0/1600/0] 0MB,0MB | 0,0K/s | 0,0K E:0,0 Connecting
End of input reached.
Input channel is now off
\ 1/0/2 [1572/1600/1600] 785MB,0MB | 36422,0K/s | 38896,0K E:0,1
Download complete.
Total time used: 0 minutes.
Seed for others 0 hours
| 0/0/2 [1600/1600/1600] 799MB,0MB | 0,0K/s | 14320,0K E:0,1 Connecting

Tarball path =  /tmp/docket/14f61693dd2db6380755a662d6e4e3583b5214fad9032bd983ce6c70df2144bc_perl_latest.tar
Exporting image to tarball...
```

## Reference (Docket Registry)

```bash
./registry --help

usage: registry [<flags>]

Docket Registry

Flags:
  --help             Show help.
  -t, --tracker="10.240.101.85:8940"  
                     Set host and port of bittorrent tracker. Example: -host 10.240.101.85:8940 Note: This cannot be set to localhost, since this is the tracker in which all the torrents will be created with. They have to be some accessible ip
                     address from outside
  -p, --port="8000"  Set port of docket registry.
  -l, --location="/var/local/docket"  
                     Set location to save torrents and docker images.
```

## Reference (Docket Client)

```bash
./client 

usage: client [<flags>] <command> [<flags>] [<args> ...]

Docket Client

Flags:
  --help             Show help.
  -h, --host="http://127.0.0.1"  
                     Set host of docket registry.
  -p, --port="8000"  Set port of docket registry.
  -l, --location="/tmp/docket"  
                     Set location to store torrents and tarballs.

Commands:
  help [<command>]
    Show help for a command.

  push <push>
    Push to the docket registry.

  pull <pull>
    pull to the docket registry.

  images [<flags>]
    display images in the docket registry.

```

## Author 

Sivamani Varun (varun@semantics3.com)


## License

MIT License

