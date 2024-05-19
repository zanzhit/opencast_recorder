package opencast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	recorder "github.com/zanzhit/opencast_recorder"
)

type OpencastService struct {
	ACL        []byte
	Processing []byte
	VideosPath string
	URL        string
	Login      string
	Password   string
}

const fileExtension = 3

func (o *OpencastService) Move(rec recorder.Recording) (io.ReadCloser, error) {
	videoFile, err := os.ReadFile(rec.FilePath)
	if err != nil {
		return nil, err
	}

	md := Metadata{
		Flavor: "dublincore/episode",
		Fields: []Field{
			{
				ID:    "title",
				Value: "title",
			},
			{
				ID:    "startDate",
				Value: rec.StartTime.Format(time.DateOnly),
			},
			{
				ID:    "startTime",
				Value: rec.StartTime.Format(time.TimeOnly),
			},
			{
				ID:    "duration",
				Value: "00:00:01",
			},
			{
				ID:    "location",
				Value: "title",
			},
		},
	}

	metadata, err := json.Marshal(md)
	if err != nil {
		return nil, err
	}

	data := map[string][]byte{
		"presenter":  videoFile,
		"metadata":   metadata,
		"acl":        o.ACL,
		"processing": o.Processing,
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	defer writer.Close()

	for fieldName, fieldData := range data {
		if fieldName == "presenter" {
			part, err := writer.CreateFormFile(fieldName, fmt.Sprintf("%s.%s", fieldName, rec.FilePath[len(rec.FilePath)-fileExtension:]))
			if err != nil {
				return nil, err
			}

			_, err = io.Copy(part, bytes.NewReader(fieldData))
			if err != nil {
				return nil, err
			}

			continue
		}

		part, err := writer.CreateFormField(fieldName)
		if err != nil {
			return nil, err
		}
		part.Write(fieldData)
	}

	opencastVideos := fmt.Sprintf("%s/api/events", o.URL)
	req, err := http.NewRequest("POST", opencastVideos, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetBasicAuth("test", "test")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp.Body, nil
}
