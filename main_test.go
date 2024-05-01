package main

import (
	"context"
	"errors"
	//"github.com/DATA-DOG/go-sqlmock"
	"google.golang.org/grpc"
	"net/http"
	"net/http/httptest"
	"testing"

	pb "task2/grpc"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// Mock gRPC client
type MockGRPCClient struct {
	ctrl *gomock.Controller
	err  error
}

func NewMockGRPCClient(ctrl *gomock.Controller, err error) *MockGRPCClient {
	return &MockGRPCClient{ctrl: ctrl, err: err}
}

func (m *MockGRPCClient) GetSuspiciousUrl(ctx context.Context, in *pb.GetUrlRequest, opts ...grpc.CallOption) (*pb.GetUrlResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	// Mock response
	return &pb.GetUrlResponse{Url: "http://example.com"}, nil
}

func TestGetSuspiciousUrl(t *testing.T) {
	// Create a new instance of the gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock gRPC client
	mockClient := NewMockGRPCClient(ctrl, nil)

	// Replace the global grpcClient variable with the mock client
	grpcClient = mockClient

	// Create a new HTTP request
	req, err := http.NewRequest("GET", "/suspicious", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new HTTP response recorder
	recorder := httptest.NewRecorder()

	// Call the handler function
	getSuspiciousUrl(recorder, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, recorder.Code)

	// Check the response body
	expectedResponseBody := "http://example.com"
	assert.Equal(t, expectedResponseBody, recorder.Body.String())
}
func TestGetSuspiciousUrl_ServerNotRunning(t *testing.T) {
	// Create a new instance of the gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock gRPC client that returns an error
	mockClient := NewMockGRPCClient(ctrl, errors.New("server not running"))

	// Replace the global grpcClient variable with the mock client
	grpcClient = mockClient

	// Create a new HTTP request
	req, err := http.NewRequest("GET", "/suspicious", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new HTTP response recorder
	recorder := httptest.NewRecorder()

	// Call the handler function
	getSuspiciousUrl(recorder, req)

	// Check the status code
	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)

	// Check the response body
	expectedResponseBody := "Service is not available now\n"
	assert.Equal(t, expectedResponseBody, recorder.Body.String())
}

func TestGetHandler(t *testing.T) {
	//req, _ := http.NewRequest("GET", "/get", strings.NewReader("domain=example.com"))
	badReq, _ := http.NewRequest("GET", "/get", nil)

	recorder := httptest.NewRecorder()
	getHandler(recorder, badReq)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}
