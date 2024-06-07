package opencastrecorder

import "time"

type Recording struct {
	CameraIP  string `json:"camera_ip"`
	RTSP      string
	StartTime time.Time `json:"start_time"`
	StopTime  time.Time `json:"stop_time"`
	FilePath  string    `json:"file_path"`
	IsMoved   bool      `json:"is_moved"`
}

type RecordingSchedule struct {
	ScheduleStartTime time.Time `json:"start_time"`
	DurationStr       string    `json:"duration"`
	Duration          time.Duration
	Recordings        []Recording
}

type Camera struct {
	CameraIP   string `json:"camera_ip"`
	RoomNumber string `json:"room_number"`
	HasAudio   bool   `json:"has_audio"`
}
