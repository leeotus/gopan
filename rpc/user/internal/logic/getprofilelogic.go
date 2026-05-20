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

// GetProfileLogic 获取用户个人信息。
type GetProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProfileLogic {
	return &GetProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetProfileLogic) GetProfile(in *user.GetProfileReq) (*user.GetProfileResp, error) {
	u, err := l.svcCtx.UserStore.FindById(l.ctx, in.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "用户不存在")
		}
		return nil, status.Error(codes.Internal, "内部错误")
	}

	return &user.GetProfileResp{
		UserId:    u.Id,
		Username:  u.Username,
		Email:     u.Email,
		Avatar:    u.Avatar,
		Signature: u.Signature,
		CreatedAt: u.CreatedAt.Unix(),
	}, nil
}
