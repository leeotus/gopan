package logic

import (
	"context"

	"gopan/rpc/user/internal/svc"
	"gopan/rpc/user/model"
	"gopan/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UpdateProfileLogic 更新用户个人信息（头像、签名、邮箱）。
type UpdateProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProfileLogic {
	return &UpdateProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateProfileLogic) UpdateProfile(in *user.UpdateProfileReq) (*user.UpdateProfileResp, error) {
	u := &model.User{
		Id:        in.UserId,
		Email:     in.Email,
		Avatar:    in.Avatar,
		Signature: in.Signature,
	}
	if err := l.svcCtx.UserStore.Update(l.ctx, u); err != nil {
		l.Logger.Errorf("update user error: %v", err)
		return nil, status.Error(codes.Internal, "更新失败")
	}
	return &user.UpdateProfileResp{}, nil
}
