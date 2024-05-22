package service

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	recorder "github.com/zanzhit/opencast_recorder"
	"github.com/zanzhit/opencast_recorder/pkg/errs"
	"github.com/zanzhit/opencast_recorder/pkg/repository"
)

type VideoService interface {
	Move(recorder.Recording) (*http.Response, error)
}

type RecordingService struct {
	repo       repository.Recording
	video      VideoService
	commands   map[string]*exec.Cmd
	videosPath string
}

func NewRecordingService(repo repository.Recording, video VideoService, videosPath string) *RecordingService {
	return &RecordingService{
		repo:       repo,
		video:      video,
		commands:   make(map[string]*exec.Cmd),
		videosPath: videosPath,
	}
}

func (s *RecordingService) Start(rec []recorder.Recording) error {
	parametres, err := recordingMode(rec, s.videosPath)
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

func (s *RecordingService) Move(cameraIP string) (*http.Response, error) {
	rec, err := s.repo.LastRecording(cameraIP)
	if err != nil {
		return nil, err
	}

	response, err := s.video.Move(rec)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *RecordingService) Schedule(rec recorder.RecordingSchedule) error {
	delay := time.Until(rec.ScheduleStartTime)
	if delay < 0 {
		return &errs.BadRequst{Message: "wrong time"}
	}

	time.Sleep(delay)

	err := s.Start(rec.Recordings)
	if err != nil {
		return err
	}

	time.Sleep(rec.Duration)

	selectedRecord := rec.Recordings[len(rec.Recordings)-1].CameraIP
	if err = s.Stop(selectedRecord); err != nil {
		return err
	}

	return nil
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

	rec[selectedRecord].StartTime = time.Now()
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

	return strings.Split(parametres, " "), nil
}
