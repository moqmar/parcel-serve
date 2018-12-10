package parcelServe

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// Serve adds the files from the `prefix` directory to a Gin instance, either by starting parcel as a proxy, or by using static data supplied by go-bindata.
func Serve(prefix string, r *gin.Engine, assetNames []string, MustAsset func(string) []byte) {
	if gin.Mode() != "release" {

		parcel := exec.Command("parcel", "index.html")
		parcel.Stdout = os.Stdout
		parcel.Stderr = os.Stderr
		parcel.Dir, _ = filepath.Abs(prefix + "/")
		err := parcel.Start()
		if err == nil {
			r.Use(parcelProxy("127.0.0.1:1234"))
			return
		}

	}

	serveAsset := parcelAssetHandler(prefix+"/dist", MustAsset)
	for _, asset := range assetNames {
		r.GET(strings.TrimSuffix(strings.TrimPrefix(asset, prefix+"/dist"), "index.html"), serveAsset)
	}

	// Fall back to last index.html if not found
	r.Use(func(c *gin.Context) {
		path := strings.Trim(c.Request.URL.Path, "/")
		segments := strings.SplitAfter(path, "/")
		if path == "" {
			c.Next()
			return
		}
		c.Request.URL.Path = "/" + strings.Join(segments[:len(segments)-1], "")
		r.HandleContext(c)
	})
}

func parcelAssetHandler(prefix string, MustAsset func(string) []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := c.Request.URL.Path
		if strings.HasSuffix(p, "/") {
			p = p + "index.html"
		}
		a := MustAsset(prefix + p)
		var contentType = mime.TypeByExtension(filepath.Ext(p))

		c.DataFromReader(http.StatusOK, int64(len(a)), contentType, bytes.NewReader(a), map[string]string{})
	}
}

func parcelProxy(host string) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			return
		}
		c.Request.Body = ioutil.NopCloser(bytes.NewReader(body))
		url := fmt.Sprintf("%s://%s%s", "http", host, c.Request.RequestURI)

		proxyReq, err := http.NewRequest(c.Request.Method, url, bytes.NewReader(body))

		proxyReq.Header = make(http.Header)
		for h, val := range c.Request.Header {
			proxyReq.Header[h] = val
		}

		client := &http.Client{}
		resp, err := client.Do(proxyReq)
		if err != nil {
			http.Error(c.Writer, err.Error(), http.StatusBadGateway)
			return
		}

		c.Status(resp.StatusCode)
		for h, val := range resp.Header {
			c.Writer.Header()[h] = val
		}

		defer resp.Body.Close()
		bodyContent, _ := ioutil.ReadAll(resp.Body)
		c.Writer.Write(bodyContent)

		c.Abort()
	}
}
