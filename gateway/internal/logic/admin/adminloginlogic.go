// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package admin

import (
	"context"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
	"gopan/rpc/admin/adminclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminLoginLogic {
	return &AdminLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminLoginLogic) AdminLogin(req *types.LoginReq) (resp *types.LoginResp, err error) {
	r, rpcErr := l.svcCtx.AdminClient.AdminLogin(l.ctx, &adminclient.AdminLoginReq{
		Username: req.Username,
		Password: req.Password,
	})
	if rpcErr != nil {
		return nil, rpcErr
	}
	return &types.LoginResp{
		Token:    r.Token,
		UserId:   r.UserId,
		Username: r.Username,
	}, nil
}
