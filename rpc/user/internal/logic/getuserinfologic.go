package logic

import (
	"context"
	"database/sql"

	"gopan/rpc/user/internal/svc"
	"gopan/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetUserInfoLogic 获取用户简要信息（供其他服务内部调用）。
// 与 GetProfile 的区别：仅返回 3 个字段（id/username/avatar），不返回敏感信息。
type GetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserInfoLogic) GetUserInfo(in *user.GetUserInfoReq) (*user.GetUserInfoResp, error) {
	u, err := l.svcCtx.UserStore.FindById(l.ctx, in.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "用户不存在")
		}
		return nil, status.Error(codes.Internal, "内部错误")
	}

	return &user.GetUserInfoResp{
		UserId:   u.Id,
		Username: u.Username,
		Avatar:   u.Avatar,
	}, nil
}
