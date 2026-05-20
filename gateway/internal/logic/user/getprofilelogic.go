// GetProfileLogic 获取个人信息编排。
package user

import (
	"context"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
	"gopan/rpc/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProfileLogic {
	return &GetProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProfileLogic) GetProfile() (resp *types.UserProfileResp, err error) {
	userId := int64(1) // TODO: 从 context 中提取 user_id

	r, err := l.svcCtx.UserClient.GetProfile(l.ctx, &userclient.GetProfileReq{UserId: userId})
	if err != nil {
		return nil, err
	}
	return &types.UserProfileResp{
		UserId:    r.UserId,
		Username:  r.Username,
		Email:     r.Email,
		Avatar:    r.Avatar,
		Signature: r.Signature,
		CreatedAt: r.CreatedAt,
	}, nil
}
