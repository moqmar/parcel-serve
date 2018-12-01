# parcel-serve - an easy way to add parcel to your gin webserver

Compile your [gin](https://gin-gonic.github.io/gin/)-based web app into a single file using [parcel](https://parceljs.org/) and [go-bindata](https://github.com/go-bindata/go-bindata).

The examples use the `ui` folder for the frontend files, and `ui/index.html` for the main HTML file.

## Install prerequisites

You need [go](https://golang.org/) and [node.js](https://nodejs.org/) to get started.

```bash
go get -u github.com/go-bindata/go-bindata
npm install -g parcel-bundler
```

## Add a Makefile to your project

```makefile
bindata:
	cd ui && parcel build index.html
	find ui/dist -type d -exec go-bindata -pkg main {} +
```

You can now build a new version of your static files using the `make` command. You should either do this before every commit, *or* add `bindata.go` and `ui/dist` to your `.gitignore` and run `make` in your CI.

## Use parcel-serve in your app

```go
package main

import (
    "github.com/moqmar/parcel-serve"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    parcelServe.Serve("ui", r, AssetNames(), MustAsset)
    r.Run()
}
```

## Add it to your CI
If you're using [Drone CI](https://drone.io/), a build configuration could look like this:

```yaml
kind: pipeline
name: default

steps:
- name: frontend
  image: node
  commands:
  - npm install -g parcel
  - cd ui && parcel build index.html
- name: backend
  image: golang
  commands:
  - go get -u github.com/go-bindata/go-bindata
  - find ui/dist -type d -exec go-bindata -pkg main {} +
  
  - CGO_ENABLED=0 go build -a -ldflags '-extldflags -static -s -w' -v -o my-application .
  # - or - (if CGO is required)
  - CGO_ENABLED=1 go build -a -ldflags '-extldflags -static -s -w' -installsuffix cgo -v -o my-application .
```
