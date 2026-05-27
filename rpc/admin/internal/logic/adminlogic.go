package logic

import (
	"context"
	"database/sql"

	"gopan/rpc/admin/internal/svc"
	"gopan/rpc/admin/admin"
	"gopan/rpc/admin/store"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	// JWT
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type AdminLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminLoginLogic {
	return &AdminLoginLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *AdminLoginLogic) AdminLogin(in *admin.AdminLoginReq) (*admin.AdminLoginResp, error) {
	u, err := l.svcCtx.AdminStore.FindAdminByUsername(l.ctx, in.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.PermissionDenied, "非管理员账号")
		}
		return nil, status.Error(codes.Internal, "查询失败")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(in.Password)); err != nil {
		return nil, status.Error(codes.PermissionDenied, "密码错误")
	}

	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  u.Id,
		"username": u.Username,
		"role":     u.Role,
		"exp":      time.Now().Add(time.Duration(l.svcCtx.Config.JWT.AccessExpire) * time.Second).Unix(),
	}).SignedString([]byte(l.svcCtx.Config.JWT.AccessSecret))

	return &admin.AdminLoginResp{Token: token, UserId: u.Id, Username: u.Username}, nil
}

type ListVideosLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListVideosLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListVideosLogic {
	return &ListVideosLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ListVideosLogic) ListVideos(in *admin.AdminListVideosReq) (*admin.AdminListVideosResp, error) {
	rows, err := l.svcCtx.AdminStore.ListVideos(l.ctx, in.Cursor, in.Limit, in.Status)
	if err != nil {
		return nil, status.Error(codes.Internal, "查询失败")
	}

	hasMore := len(rows) > int(in.Limit)
	if hasMore {
		rows = rows[:in.Limit]
	}

	videos := make([]*admin.AdminVideoInfo, len(rows))
	for i, r := range rows {
		videos[i] = &admin.AdminVideoInfo{
			Id: r.Id, Title: r.Title, CoverUrl: r.CoverUrl,
			UserId: r.UserId, Status: r.Status, PlayCount: r.PlayCount, CreatedAt: r.CreatedAt,
		}
	}

	resp := &admin.AdminListVideosResp{HasMore: hasMore, Videos: videos}
	if !hasMore && len(rows) > 0 {
		resp.NextCursor = rows[len(rows)-1].Id
	}
	return resp, nil
}

type ApproveVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewApproveVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApproveVideoLogic {
	return &ApproveVideoLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ApproveVideoLogic) ApproveVideo(in *admin.ApproveVideoReq) (*admin.ApproveVideoResp, error) {
	if err := l.svcCtx.AdminStore.UpdateVideoStatus(l.ctx, in.VideoId, 2); err != nil {
		return nil, status.Error(codes.Internal, "审核失败")
	}
	return &admin.ApproveVideoResp{}, nil
}

type RejectVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRejectVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RejectVideoLogic {
	return &RejectVideoLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *RejectVideoLogic) RejectVideo(in *admin.RejectVideoReq) (*admin.RejectVideoResp, error) {
	if err := l.svcCtx.AdminStore.UpdateVideoStatus(l.ctx, in.VideoId, 4); err != nil {
		return nil, status.Error(codes.Internal, "审核失败")
	}
	return &admin.RejectVideoResp{}, nil
}

type AdminDeleteVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminDeleteVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminDeleteVideoLogic {
	return &AdminDeleteVideoLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *AdminDeleteVideoLogic) AdminDeleteVideo(in *admin.AdminDeleteVideoReq) (*admin.AdminDeleteVideoResp, error) {
	if err := l.svcCtx.AdminStore.UpdateVideoStatus(l.ctx, in.VideoId, 3); err != nil {
		return nil, status.Error(codes.Internal, "删除失败")
	}
	return &admin.AdminDeleteVideoResp{}, nil
}

var _ store.AdminStore
