package delivery_auth_grpc

import (
	"context"
	"log/slog"
	"net"
        "fmt"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/authorization/repository/profile"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/authorization/repository/session"
	"google.golang.org/grpc"
        "github.com/go-park-mail-ru/2023_2_Vkladyshi/configs"

	pb "github.com/go-park-mail-ru/2023_2_Vkladyshi/authorization/proto"
)

type Config struct {
	Port           string `yaml:"port"`
	ConnectionType string `yaml:"connection_type"`
}

type server struct {
	pb.UnimplementedAuthorizationServer
	userRepo    *profile.RepoPostgre
	sessionRepo *session.SessionRepo
	lg          *slog.Logger
}

func (s *server) GetId(ctx context.Context, req *pb.FindIdRequest) (*pb.FindIdResponse, error) {
        fmt.Println(req.Sid)
	login, err := s.sessionRepo.GetUserLogin(ctx, req.Sid, s.lg)
	if err != nil {
		return nil, err
	}
        fmt.Println(login)
	id, err := s.userRepo.GetUserProfileId(login)
        fmt.Println(id)
	if err != nil {
		return nil, err
	}
	return &pb.FindIdResponse{
		Value: id,
	}, nil
}

func (s *server) GetIdsAndPaths(ctx context.Context, req *pb.IdsAndPathsListRequest) (*pb.IdsAndPathsResponse, error) {
	ids, paths, err := s.userRepo.GetIdsAndPaths()
	if err != nil {
		return nil, err
	}
	return &pb.IdsAndPathsResponse{
		Ids:   ids,
		Paths: paths,
	}, nil
}

func (s *server) GetAuthorizationStatus(ctx context.Context, req *pb.AuthorizationCheckRequest) (*pb.AuthorizationCheckResponse, error) {
	status, err := s.sessionRepo.CheckActiveSession(ctx, req.Sid, s.lg)
	if err != nil {
		return nil, err
	}
	return &pb.AuthorizationCheckResponse{
		Status: status,
	}, nil
}

func ListenAndServeGrpc(l *slog.Logger) {
        config, err := configs.ReadConfig()
        if err != nil {
                l.Error("read config error", "err", err.Error())
                return
        }

        configSession, err := configs.ReadSessionRedisConfig()
        if err != nil {
                l.Error("read config error", "err", err.Error())
                return
        }

        session, err := session.GetSessionRepo(*configSession, l)

        if err != nil {
                l.Error("Session repository is not responding")
                return
        }

        users, err := profile.GetUserRepo(config, l)
        if err != nil {
                l.Error("cant create repo")
                return
        }

        lis, err := net.Listen("tcp", ":50051")
        if err != nil {
                l.Error("failed to listen: %v", err)
                //fmt.Errorf("get film genres err: %w", err)
                return
        }

	s := grpc.NewServer()
	pb.RegisterAuthorizationServer(s, &server{
		lg: l,
                sessionRepo: session,
                userRepo: users,
	})
	if err := s.Serve(lis); err != nil {
		l.Error("failed to serve: %v", err)
                //fmt.Errorf("get film genres err: %w", err)
		return
	}
}
