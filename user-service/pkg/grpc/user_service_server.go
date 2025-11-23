package grpc

import (
	"context"
	"fmt"
	"log"

	pb "go-vedio-1/proto/user"
	"user-service/ddd/application/app"
	"user-service/pkg/errno"
)

// UserServiceServer gRPC服务实现
type UserServiceServer struct {
	pb.UnimplementedUserServiceServer
	userApp app.UserApp
}

// NewUserServiceServer 创建gRPC服务实例
func NewUserServiceServer(userApp app.UserApp) *UserServiceServer {
	return &UserServiceServer{
		userApp: userApp,
	}
}

// GetUserByUUID 根据用户UUID获取用户信息
func (s *UserServiceServer) GetUserByUUID(ctx context.Context, req *pb.GetUserByUUIDRequest) (*pb.GetUserByUUIDResponse, error) {
	log.Printf("gRPC GetUserByUUID called with UUID: %s", req.UserUuid)

	// 调用应用层服务
	userInfo, err := s.userApp.GetUserInfo(ctx, req.UserUuid)
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		return &pb.GetUserByUUIDResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get user info: %v", err),
		}, nil
	}

	// 转换为gRPC响应
	pbUser := &pb.UserInfo{
		UserUuid:  userInfo.UserUUID,
		UserId:    0,
		Account:   userInfo.Account,
		Email:     "",
		Nickname:  "",
		AvatarUrl: userInfo.AvatarUrl,
		CreatedAt: 0,
		UpdatedAt: 0,
	}

	return &pb.GetUserByUUIDResponse{
		Success: true,
		Message: "User found successfully",
		User:    pbUser,
	}, nil
}

// ValidateUser 验证用户是否存在
func (s *UserServiceServer) ValidateUser(ctx context.Context, req *pb.ValidateUserRequest) (*pb.ValidateUserResponse, error) {
	log.Printf("gRPC ValidateUser called with UUID: %s", req.UserUuid)

	// 调用应用层服务
	_, err := s.userApp.GetUserInfo(ctx, req.UserUuid)
	if err != nil {
		// 检查是否是用户不存在的错误
		if err == errno.ErrUserNotFound {
			return &pb.ValidateUserResponse{
				Success: true,
				Message: "User validation completed",
				Exists:  false,
			}, nil
		}
		// 其他错误
		return &pb.ValidateUserResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to validate user: %v", err),
			Exists:  false,
		}, nil
	}

	return &pb.ValidateUserResponse{
		Success: true,
		Message: "User validation completed",
		Exists:  true,
	}, nil
}

// GetUsersByUUIDs 批量获取用户信息
func (s *UserServiceServer) GetUsersByUUIDs(ctx context.Context, req *pb.GetUsersByUUIDsRequest) (*pb.GetUsersByUUIDsResponse, error) {
	log.Printf("gRPC GetUsersByUUIDs called with %d UUIDs", len(req.UserUuids))

	var users []*pb.UserInfo
	var failedUUIDs []string

	// 批量查询用户信息
	for _, uuid := range req.UserUuids {
		userInfo, err := s.userApp.GetUserInfo(ctx, uuid)
		if err != nil {
			log.Printf("Failed to get user info for UUID %s: %v", uuid, err)
			failedUUIDs = append(failedUUIDs, uuid)
			continue
		}

		// 转换为gRPC响应
		pbUser := &pb.UserInfo{
			UserUuid:  userInfo.UserUUID,
			UserId:    0,
			Account:   userInfo.Account,
			Email:     "",
			Nickname:  "",
			AvatarUrl: userInfo.AvatarUrl,
			CreatedAt: 0,
			UpdatedAt: 0,
		}
		users = append(users, pbUser)
	}

	// 构建响应消息
	message := fmt.Sprintf("Found %d users", len(users))
	if len(failedUUIDs) > 0 {
		message += fmt.Sprintf(", failed to find %d users", len(failedUUIDs))
	}

	return &pb.GetUsersByUUIDsResponse{
		Success: true,
		Message: message,
		Users:   users,
	}, nil
}