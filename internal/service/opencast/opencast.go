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

	"github.com/zanzhit/opencast_recorder/internal/domain/models"
)

type OpencastService struct {
	acl        []byte
	processing []byte
	videosPath string
	url        string
	login      string
	password   string
}

const fileExtension = 3

func NewOpencastService(acl, processing []byte, videosPath, url, login, password string) *OpencastService {
	return &OpencastService{
		acl:        acl,
		processing: processing,
		videosPath: videosPath,
		url:        url,
		login:      login,
		password:   password,
	}
}

func (o *OpencastService) Move(rec models.Recording) (*http.Response, error) {
	videoFile, err := os.ReadFile(rec.FilePath)
	if err != nil {
		return nil, err
	}

	duration := rec.StopTime.Sub(rec.StartTime)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60
	formattedDuration := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)

	md := []Metadata{
		{
			Flavor: "dublincore/episode",
			Fields: []Field{
				{
					ID:    "title",
					Value: rec.CameraIP,
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
					Value: formattedDuration,
				},
				{
					ID:    "location",
					Value: rec.CameraIP,
				},
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
		"acl":        o.acl,
		"processing": o.processing,
	}

	body := &bytes.Buffer{}
	contentType, err := o.createForm(data, body, rec)
	if err != nil {
		return nil, err
	}

	opencastVideos := fmt.Sprintf("%s/api/events", o.url)
	req, err := http.NewRequest("POST", opencastVideos, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	req.SetBasicAuth(o.login, o.password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (o *OpencastService) createForm(data map[string][]byte, body *bytes.Buffer, rec models.Recording) (string, error) {
	writer := multipart.NewWriter(body)
	defer writer.Close()

	for fieldName, fieldData := range data {
		if fieldName == "presenter" {
			part, err := writer.CreateFormFile(fieldName, fmt.Sprintf("%s.%s", fieldName, rec.FilePath[len(rec.FilePath)-fileExtension:]))
			if err != nil {
				return "", err
			}

			_, err = io.Copy(part, bytes.NewReader(fieldData))
			if err != nil {
				return "", err
			}

			continue
		}

		part, err := writer.CreateFormField(fieldName)
		if err != nil {
			return "", err
		}
		part.Write(fieldData)
	}

	return writer.FormDataContentType(), nil
}
