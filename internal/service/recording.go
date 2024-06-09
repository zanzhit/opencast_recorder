package service

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/zanzhit/opencast_recorder/internal/domain/models"
	"github.com/zanzhit/opencast_recorder/internal/errs"
	"github.com/zanzhit/opencast_recorder/internal/repository"
)

type VideoService interface {
	Move(models.Recording) (*http.Response, error)
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

func (s *RecordingService) Start(rec []models.Recording) error {
	if available := isCameraAvailable(rec[len(rec)-1].CameraIP); !available {
		return &errs.BadRequst{Message: "camera is not available"}
	}

	parametres, err := recordingMode(rec, s.videosPath)
	if err != nil {
		return err
	}

	if _, ok := s.commands[rec[len(rec)-1].CameraIP]; ok {
		err = s.Stop(rec[len(rec)-1].CameraIP)
		if err != nil {
			return err
		}
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

	if err := os.Remove(rec.FilePath); err != nil {
		return nil, err
	}

	return response, nil
}

func (s *RecordingService) DeleteLocal(filePath string) error {
	if err := os.Remove(filePath); err != nil {
		return err
	}

	return nil
}

func (s *RecordingService) Schedule(rec models.RecordingSchedule) error {
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

func (s *RecordingService) Stats(cameraIP string) (models.Recording, error) {
	rec, err := s.repo.LastRecording(cameraIP)
	if err != nil {
		return rec, err
	}

	return rec, nil
}

func recordingMode(rec []models.Recording, videosPath string) ([]string, error) {
	selectedRecord := len(rec) - 1

	fileName := strings.ReplaceAll(rec[selectedRecord].CameraIP, ":", "..")

	rec[selectedRecord].StartTime = time.Now()
	rec[selectedRecord].FilePath = fmt.Sprintf("%s/%s_%s.mkv", videosPath, fileName, rec[selectedRecord].StartTime.Format("2006-01-02_15-04-05"))

	var parametres string
	switch len(rec) {
	case 1:
		parametres = fmt.Sprintf("gst-launch-1.0 rtspsrc location=%s ! rtph264depay ! h264parse ! matroskamux ! filesink location=%s",
			rec[0].RTSP, rec[0].FilePath)
	case 2:
		parametres = fmt.Sprintf("gst-launch-1.0 -e videomixer name=mix sink_0::xpos=0 sink_1::xpos=640 ! videoconvert ! x264enc ! queue ! mux. uridecodebin uri=%s ! videoconvert ! videoscale ! video/x-raw,width=640,height=480 ! mix.sink_0 uridecodebin uri=%s ! videoconvert ! videoscale ! video/x-raw,width=640,height=480 ! mix.sink_1 uridecodebin uri=%s ! audioconvert ! vorbisenc ! queue ! mux. matroskamux name=mux ! filesink location=%s",
			rec[0].RTSP, rec[1].RTSP, rec[0].RTSP, rec[1].FilePath)
	default:
		return nil, fmt.Errorf("too many arguments")
	}

	return strings.Split(parametres, " "), nil
}

func isCameraAvailable(cameraIP string) bool {
	conn, err := net.DialTimeout("tcp", cameraIP, 3*time.Second)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return false
	}
	conn.Close()
	return true
}
