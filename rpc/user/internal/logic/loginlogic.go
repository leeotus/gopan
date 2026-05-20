package logic

import (
	"context"
	"database/sql"

	"gopan/rpc/user/internal/svc"
	"gopan/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoginLogic 登录逻辑。
type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Login 用户登录。
// 流程: 通过用户名查用户 → bcrypt 对比密码 → 生成 JWT → 返回 token。
func (l *LoginLogic) Login(in *user.LoginReq) (*user.LoginResp, error) {
	// 1. 查询用户
	u, err := l.svcCtx.UserStore.FindByUsername(l.ctx, in.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "用户名或密码错误")
		}
		l.Logger.Errorf("find user error: %v", err)
		return nil, status.Error(codes.Internal, "内部错误")
	}

	// 2. 验证密码（bcrypt 比对哈希）
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(in.Password)); err != nil {
		return nil, status.Error(codes.Unauthenticated, "用户名或密码错误")
	}

	// 3. 生成 JWT token
	token, err := GenerateToken(u.Id, u.Username, l.svcCtx.Config.JWT.AccessSecret, l.svcCtx.Config.JWT.AccessExpire)
	if err != nil {
		l.Logger.Errorf("generate token error: %v", err)
		return nil, status.Error(codes.Internal, "生成token失败")
	}

	return &user.LoginResp{
		Token:    token,
		UserId:   u.Id,
		Username: u.Username,
		Avatar:   u.Avatar,
	}, nil
}
