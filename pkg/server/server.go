package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
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
	return params["filename"], err
}
func getExtractor(header http.Header) (extractor, error) {
	n, err := getFilename(header)
	if err != nil {
		return nil, err
	}
	filename := strings.Split(n, ".")
	if len(filename) < 2 {
		return nil, fmt.Errorf("invalid filename")
	}
	switch filename[len(filename)] {
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

	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{})
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	n, err := new(XZExtractor).Extract(w, f.Body)
	if err != nil {
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	w.Header().Add("Content-Type", "application/octet-stream")
	w.Header().Add("Content-Disposition", "attachment; filename=metal-amd64.raw")
	w.Header().Add("Content-Length", fmt.Sprintf("%d", n))
	logrus.WithField("size", n).Info("file send to client")
}
