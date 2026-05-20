// LoginLogic 登录编排: 调用 user-svc.Login → 返回 token + 用户信息。
package user

import (
	"context"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
	"gopan/rpc/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	r, err := l.svcCtx.UserClient.Login(l.ctx, &userclient.LoginReq{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &types.LoginResp{
		Token:    r.Token,
		UserId:   r.UserId,
		Username: r.Username,
		Avatar:   r.Avatar,
	}, nil
}
