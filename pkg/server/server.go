package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

func Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/", extractXZ)
	return m
}

func getDL(url string) (*http.Response, error) {
	client := new(http.Client)

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return errors.New("Redirect")
	}

	response, err := client.Get(url)
	if response != nil && response.StatusCode == http.StatusFound { //status code 302
		l, _ := response.Location()
		logrus.WithField("to", l.String()).Info("rediret")
		return getDL(l.String())
	}
	return response, err
}

func getFilename(header http.Header) (string, error) {
	_, params, err := mime.ParseMediaType(header.Get("Content-Disposition"))
	logrus.WithFields(logrus.Fields{
		"disposition": header.Get("Content-Disposition"),
		"params":      params,
		"error":       err,
	}).Debug("reading content-disposition")
	return params["filename"], err
}
func getExtractor(f string) (extractor, error) {
	filename := strings.Split(f, ".")
	if len(filename) < 2 {
		return nil, fmt.Errorf("invalid filename")
	}
	switch filename[len(filename)-1] {
	case "xz":
		return &XZExtractor{
			filename: strings.Join(filename[:len(filename)-1], "."),
		}, nil
	case "tar":
		return &TarExtractor{
			filename: strings.Join(filename[:len(filename)-1], "."),
		}, nil
	}
	return nil, fmt.Errorf("unknown archive")
}

func getFilenameFromURL(u string) (string, error) {
	uri, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	p := strings.Split(uri.Path, "/")
	return p[len(p)], nil
}
func extractXZ(w http.ResponseWriter, r *http.Request) {
	dl := r.URL.Query().Get("url")

	logrus.WithField("url", dl).Info("start download")

	f, err := getDL(dl)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"msg": fmt.Sprintf("failed downloading %s", dl),
			"err": err,
		})
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer f.Body.Close()
	logrus.WithField("status", f.Status).Info("read request")

	fn, err := getFilename(f.Header)
	if err != nil {
		fn, err = getFilenameFromURL(dl)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"err": err,
				"msg": "can't get filename",
			})
			logrus.WithField("error", err).Warn("failed reading archive meta")
			return
		}
	}
	ex, err := getExtractor(fn)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"err": err,
		})
		logrus.WithField("error", err).Warn("failed reading archive meta")
		return
	}

	w.Header().Add("Content-Type", "application/octet-stream")
	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s", ex.Filename()))
	n, err := ex.Extract(w, f.Body)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"err": err,
		})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Length", fmt.Sprintf("%d", n))
	logrus.WithField("size", n).Info("file send to client")
}
