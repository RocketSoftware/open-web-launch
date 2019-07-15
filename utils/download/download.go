package download

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/rocketsoftware/open-web-launch/utils/log"
)

func ToMemory(url string) ([]byte, error) {
	var buffer bytes.Buffer
	if err := download(url, &buffer); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// ToFile downloads url and saves it under the directory dir.
// allowCached indicates whether is allowed to use an existing file
// if the resource on the network is not available.
func ToFile(url string, dir string, allowCached bool) (string, error) {
	isFileExist := true
	filename := filepath.Join(dir, path.Base(url))
	stat, err := os.Stat(filename)
	if os.IsNotExist(err) {
		isFileExist = false
	} else if err != nil {
		return "", err
	}
	if isFileExist {
		lastModifiedTime, err := GetLastModifiedTime(url)
		if err != nil {
			if allowCached {
				log.Printf("warning: unable to update %s because %v, cached version will be used", url, err)
				return filename, nil
			}
			return "", err
		}
		if lastModifiedTime.Before(stat.ModTime()) {
			log.Printf("no newer version for %s found on the network, cached version will be used", url)
			return filename, nil
		}
	}
	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	if err := download(url, file); err != nil {
		return "", err
	}
	return filename, nil
}

func download(url string, writer io.Writer) (err error) {
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "downloading %s", url)
		}
	}()
	response, err := http.Get(url)
	if err != nil {
		return
	}
	body := response.Body
	defer body.Close()
	if response.StatusCode != 200 {
		return fmt.Errorf("HTTP %s", response.Status)
	}
	_, err = io.Copy(writer, body)
	return
}

func GetLastModifiedTime(url string) (time.Time, error) {
	response, err := http.Head(url)
	if err != nil {
		return time.Time{}, err
	}
	var lastModifiedTime time.Time
	if response.StatusCode != 200 {
		return time.Time{}, fmt.Errorf("HTTP %s", response.Status)
	}
	lastModifiedHeader := response.Header.Get("Last-Modified")
	if lastModifiedHeader == "" {
		return time.Now(), nil
	}
	if lastModifiedTime, err = time.Parse(http.TimeFormat, lastModifiedHeader); err != nil {
		return time.Time{}, err
	}
	return lastModifiedTime, nil
}

func init() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}
