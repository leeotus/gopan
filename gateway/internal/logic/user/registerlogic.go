// RegisterLogic 注册编排: 接收 HTTP 请求 → 调用 user-svc gRPC → 返回结果。
package user

import (
	"context"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
	"gopan/rpc/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.BaseResp, err error) {
	_, err = l.svcCtx.UserClient.Register(l.ctx, &userclient.RegisterReq{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
	})
	if err != nil {
		return &types.BaseResp{Message: err.Error()}, nil
	}
	return &types.BaseResp{Message: "注册成功"}, nil
}
