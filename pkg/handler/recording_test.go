package handler

// import (
// 	"testing"

// 	recorder "github.com/zanzhit/opencast_recorder"
// 	mock_service "github.com/zanzhit/opencast_recorder/pkg/service/mocks"
// )

// func TestHandler_stats(t *testing.T) {
// 	type mockBehavior func(s *mock_service.MockRecording, inputParam string)

// 	testTable := []struct {
// 		name                string
// 		inputParam          string
// 		mockBehavior        mockBehavior
// 		expectedStatusCode  int
// 		expectedRequestBody string
// 	}{
// 		{
// 			name:       "OK",
// 			inputParam: "12345",
// 			mockBehavior: func(s *mock_service.MockRecording, inputParam string) {
// 				s.EXPECT().Stats(inputParam).Return(recorder.Recording{CameraIP: inputParam}, nil)
// 			},
// 			expectedStatusCode:  200,
// 			expectedRequestBody: `{"camera_ip:"1","start_time:0"}`,
// 		},
// 	}

// }
