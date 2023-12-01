package usecase

import (
	"context"
	"fmt"
	"log/slog"

	auth "github.com/go-park-mail-ru/2023_2_Vkladyshi/authorization/proto"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/comments/repository/comment"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/configs"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/pkg/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//go:generate mockgen -source=core.go -destination=../mocks/core_mock.go -package=mocks

type ICore interface {
	GetFilmComments(filmId uint64, first uint64, limit uint64) ([]models.CommentItem, error)
	AddComment(filmId uint64, userId uint64, rating uint16, text string) (bool, error)
	GetUserId(ctx context.Context, sid string) (uint64, error)
}

type Core struct {
	lg       *slog.Logger
	comments comment.ICommentRepo
}

func GetCore(cfg_sql *configs.CommentCfg, lg *slog.Logger, comments comment.ICommentRepo) *Core {
	core := Core{
		lg:       lg.With("module", "core"),
		comments: comments,
	}
	return &core
}

func (core *Core) GetFilmComments(filmId uint64, first uint64, limit uint64) ([]models.CommentItem, error) {
	comments, err := core.comments.GetFilmComments(filmId, first, limit)
	if err != nil {
		core.lg.Error("Get Film Comments error", "err", err.Error())
		return nil, fmt.Errorf("GetFilmComments err: %w", err)
	}

	return comments, nil
}

func (core *Core) AddComment(filmId uint64, userId uint64, rating uint16, text string) (bool, error) {
	found, err := core.comments.HasUsersComment(userId, filmId)
	if err != nil {
		core.lg.Error("find users comment error", "err", err.Error())
		return false, fmt.Errorf("find users comment error: %w", err)
	}
	if found {
		return found, nil
	}

	err = core.comments.AddComment(filmId, userId, rating, text)
	if err != nil {
		core.lg.Error("Add Comment error", "err", err.Error())
		return false, fmt.Errorf("GetActorsCareer err: %w", err)
	}

	return false, nil
}

func (core *Core) GetUserId(ctx context.Context, sid string) (uint64, error) {
	conn, err := grpc.Dial(":8081", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		core.lg.Error("grpc connect error", "err", err.Error())
		return 0, fmt.Errorf("grpc connect err: %w", err)
	}
	client := auth.NewAuthorizationClient(conn)

	request := auth.FindIdRequest{Sid: sid}

	response, err := client.GetId(ctx, &request)
	if err != nil {
		core.lg.Error("get user id error", "err", err.Error())
		return 0, fmt.Errorf("get user id err: %w", err)
	}
	return uint64(response.Value), nil
}
