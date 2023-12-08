package usecase

import (
	"bytes"
	"fmt"
	"log/slog"
	"reflect"
	"testing"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/comments/mocks"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/pkg/models"
	"github.com/golang/mock/gomock"
)

func TestAddComment(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockObj := mocks.NewMockICommentRepo(mockCtrl)
	mockObj.EXPECT().HasUsersComment(uint64(1), uint64(1)).Return(false, nil)
	mockObj.EXPECT().HasUsersComment(uint64(0), uint64(1)).Return(false, fmt.Errorf("repo_error"))
	mockObj.EXPECT().HasUsersComment(uint64(2), uint64(1)).Return(true, nil)
	mockObj.EXPECT().HasUsersComment(uint64(2), uint64(2)).Return(false, nil)
	mockObj.EXPECT().AddComment(uint64(1), uint64(1), uint16(1), string("t")).Return(nil)
	mockObj.EXPECT().AddComment(uint64(2), uint64(2), uint16(1), string("t")).Return(fmt.Errorf("repo_error"))

	var buff bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buff, nil))
	core := Core{comments: mockObj, lg: logger}

	found, err := core.AddComment(1, 1, 1, "t")
	if err != nil {
		t.Errorf("waited no errors")
		return
	}
	if found {
		t.Errorf("waited not found")
		return
	}

	found, err = core.AddComment(1, 0, 1, "t")
	if err == nil {
		t.Errorf("waited error")
		return
	}
	if found {
		t.Errorf("waited not found")
		return
	}

	found, err = core.AddComment(1, 2, 1, "t")
	if err != nil {
		t.Errorf("waited no errors")
		return
	}
	if !found {
		t.Errorf("waited to find")
		return
	}

	found, err = core.AddComment(2, 2, 1, "t")
	if err == nil {
		t.Errorf("waited find error")
		return
	}
	if found {
		t.Errorf("waited not found")
		return
	}
}

func TestGetFilmComments(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	expected := []models.CommentItem{}
	expected = append(expected, models.CommentItem{Username: "u"})

	mockObj := mocks.NewMockICommentRepo(mockCtrl)
	firstCall := mockObj.EXPECT().GetFilmComments(uint64(1), uint64(1), uint64(1)).Return(expected, nil)
	mockObj.EXPECT().GetFilmComments(uint64(1), uint64(0), uint64(1)).After(firstCall).Return(nil, fmt.Errorf("repo_error"))

	var buff bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buff, nil))
	core := Core{comments: mockObj, lg: logger}

	got, err := core.GetFilmComments(1, 1, 1)
	if err != nil {
		t.Errorf("waited no errors")
		return
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("waited %v, got %v", expected, got)
		return
	}

	got, err = core.GetFilmComments(1, 0, 1)
	if err == nil {
		t.Errorf("waited error")
		return
	}
	if got != nil {
		t.Errorf("waited nil")
		return
	}
}
