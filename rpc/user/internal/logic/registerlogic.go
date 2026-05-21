// package logic 实现 user-svc 的所有 gRPC 方法。
// 每个 proto rpc 对应一个 Logic 结构体，方法签名 = proto 定义的请求/响应类型。
package logic

import (
	"context"
	"database/sql"
	"time"

	"gopan/rpc/user/internal/svc"
	"gopan/rpc/user/model"
	"gopan/rpc/user/user"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RegisterLogic 注册逻辑。
type RegisterLogic struct {
	ctx    context.Context     // 请求上下文
	svcCtx *svc.ServiceContext // 服务上下文，包含数据库连接池
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Register 用户注册。
// 流程: 检查用户名冲突 → bcrypt 哈希密码 → INSERT users → 返回 user_id。
func (l *RegisterLogic) Register(in *user.RegisterReq) (*user.RegisterResp, error) {
	// 1. 检查用户名是否已存在
	existing, err := l.svcCtx.UserStore.FindByUsername(l.ctx, in.Username)
	if err != nil && err != sql.ErrNoRows {
		l.Logger.Errorf("check username error: %v", err)
		return nil, status.Error(codes.Internal, "内部错误")
	}
	if existing != nil {
		return nil, status.Error(codes.AlreadyExists, "用户名已存在")
	}

	// 2. bcrypt 加密密码（cost=10，约 100ms）
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "密码加密失败")
	}

	// 3. 写入数据库
	u := &model.User{
		Username: in.Username,
		Password: string(hashedPassword),
		Email:    in.Email,
	}
	result, err := l.svcCtx.UserStore.Insert(l.ctx, u)
	if err != nil {
		l.Logger.Errorf("insert user error: %v", err)
		return nil, status.Error(codes.Internal, "注册失败")
	}

	userId, _ := result.LastInsertId()
	return &user.RegisterResp{UserId: userId}, nil
}

// GenerateToken 使用 HMAC-SHA256 生成 JWT token。
// 包含 claims: user_id, username, exp（过期时间）, iat（签发时间）。
func GenerateToken(userId int64, username, secret string, expire int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userId,
		"username": username,
		"exp":      time.Now().Add(time.Duration(expire) * time.Second).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
