package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	recorder "github.com/zanzhit/opencast_recorder"
	"github.com/zanzhit/opencast_recorder/pkg/errs"
	"github.com/zanzhit/opencast_recorder/pkg/repository"
)

type VideoService interface {
	Move(recorder.Recording) ([]byte, error)
}

type RecordingService struct {
	repo     repository.Recording
	cfg      Config
	commands map[string]*exec.Cmd
}

const fileExtension = 3

func NewRecordingService(repo repository.Recording, cfg Config) *RecordingService {
	return &RecordingService{
		repo:     repo,
		cfg:      cfg,
		commands: make(map[string]*exec.Cmd),
	}
}

func (s *RecordingService) Start(rec []recorder.Recording) error {
	parametres, err := recordingMode(rec, s.cfg.VideosPath)
	if err != nil {
		return err
	}

	cmd := exec.Command(parametres[0], parametres[1:]...)
	if err := cmd.Start(); err != nil {
		return err
	}

	s.commands[rec[len(rec)-1].CameraIP] = cmd

	if err := s.repo.Start(rec[len(rec)-1]); err != nil {
		return err
	}

	return nil
}

func (s *RecordingService) Stop(cameraIP string) error {
	cmd, ok := s.commands[cameraIP]
	if !ok {
		return &errs.ErrNoRecording{}
	}

	if err := cmd.Process.Kill(); err != nil {
		return err
	}

	delete(s.commands, cameraIP)

	if err := s.repo.Stop(cameraIP); err != nil {
		return err
	}

	return nil
}

//РЕАЛИЗОВАТЬ АБСТРАКЦИЮ НА СЛУЧАЙ ДРУГОГО СЕРВИСА ВИДЕО

func (s *RecordingService) Move(cameraIP string) ([]byte, error) {
	rec, err := s.repo.LastRecording(cameraIP)
	if err != nil {
		return nil, err
	}

	if rec.IsMoved {
		return nil, &errs.BadRequst{Message: "recording has already been moved"}
	}

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
		"acl":        s.cfg.ACL,
		"processing": s.cfg.Processing,
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

	opencastVideos := fmt.Sprintf("%s/api/events", s.cfg.VideoService)
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println(respBody)
	fmt.Println(resp.Status)

	return respBody, nil
}

func (s *RecordingService) Schedule(rec recorder.RecordingSchedule) error {
	errChan := make(chan error)
	go func() {
		delay := time.Until(rec.ScheduleStartTime)
		if delay < 0 {
			errChan <- &errs.BadRequst{Message: "wrong time"}
			return
		}

		time.Sleep(delay)

		err := s.Start(rec.Recordings)
		if err != nil {
			errChan <- err
			return
		}

		time.Sleep(rec.Duration)

		selectedRecord := rec.Recordings[len(rec.Recordings)-1].CameraIP
		err = s.Stop(selectedRecord)
		if err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-time.After(time.Until(rec.ScheduleStartTime.Add(rec.Duration))):
		return nil
	}
}

func (s *RecordingService) Stats(cameraIP string) (recorder.Recording, error) {
	rec, err := s.repo.LastRecording(cameraIP)
	if err != nil {
		return rec, err
	}

	return rec, nil
}

func recordingMode(rec []recorder.Recording, videosPath string) ([]string, error) {
	selectedRecord := len(rec) - 1
	fileName := strings.ReplaceAll(rec[selectedRecord].CameraIP, ":", "..")

	rec[selectedRecord].FilePath = fmt.Sprintf("%s/%s.mkv", videosPath, fileName)

	var parametres string
	switch len(rec) {
	case 1:
		parametres = fmt.Sprintf("gst-launch-1.0 rtspsrc location=%s ! rtph264depay ! h264parse ! matroskamux ! filesink location=%s",
			rec[0].RTSP, rec[0].FilePath)
	case 2:
		parametres = fmt.Sprintf("gst-launch-1.0 -e videomixer name=mix sink_0::xpos=0 sink_1::xpos=640 ! videoconvert ! x264enc ! matroskamux ! filesink location=%s uridecodebin uri=%s ! videoconvert ! videoscale ! video/x-raw,width=640,height=480 ! mix.sink_0 uridecodebin uri=%s ! videoconvert ! videoscale ! video/x-raw,width=640,height=480 ! mix.sink_1 uridecodebin uri=%s ! audioconvert ! vorbisenc ! matroskamux ! filesink location=%s",
			rec[1].FilePath, rec[0].RTSP, rec[1].RTSP, rec[0].RTSP, rec[1].FilePath)
	default:

		return nil, fmt.Errorf("too many arguments")
	}

	fmt.Println(parametres)

	return strings.Split(parametres, " "), nil
}
