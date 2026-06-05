// UpdateProfileLogic 更新个人信息编排。
package user

import (
	"context"

	"gopan/gateway/internal/middleware"
	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
	"gopan/rpc/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProfileLogic {
	return &UpdateProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateProfileLogic) UpdateProfile(req *types.UpdateProfileReq) (resp *types.BaseResp, err error) {
	userId := middleware.GetUserIdFromContext(l.ctx)
	_, err = l.svcCtx.UserClient.UpdateProfile(l.ctx, &userclient.UpdateProfileReq{
		UserId:    userId,
		Avatar:    req.Avatar,
		Signature: req.Signature,
		Email:     req.Email,
	})
	if err != nil {
		return &types.BaseResp{Message: err.Error()}, nil
	}
	return &types.BaseResp{Message: "更新成功"}, nil
}
