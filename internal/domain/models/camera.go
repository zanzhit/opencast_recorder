package models

type Camera struct {
	CameraIP   string `json:"camera_ip"`
	RoomNumber string `json:"room_number"`
	HasAudio   bool   `json:"has_audio"`
}
