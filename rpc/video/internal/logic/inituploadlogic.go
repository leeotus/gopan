// InitUploadLogic 初始化分片上传，创建视频记录并返回 upload_id。
package logic

import (
	"context"
	"fmt"
	"time"

	"gopan/rpc/video/internal/svc"
	"gopan/rpc/video/model"
	"gopan/rpc/video/video"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InitUploadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInitUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitUploadLogic {
	return &InitUploadLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *InitUploadLogic) InitUpload(in *video.InitUploadReq) (*video.InitUploadResp, error) {
	uploadId := fmt.Sprintf("upload_%d_%d", time.Now().UnixMilli(), time.Now().Nanosecond()%10000)

	v := &model.Video{
		Title:       in.Title,
		UserId:      in.UserId,
		FileSize:    in.FileSize,
		TotalChunks: in.TotalChunks,
		UploadId:    uploadId,
		Status:      0, // 上传中
		ObjectKey:   fmt.Sprintf("videos/%d/source.mp4", 0), // temp，真正 key 在 merge 时确定
	}

	result, err := l.svcCtx.VideoStore.Insert(l.ctx, v)
	if err != nil {
		l.Logger.Errorf("insert video error: %v", err)
		return nil, status.Error(codes.Internal, "创建视频记录失败")
	}

	videoId, _ := result.LastInsertId()
	return &video.InitUploadResp{VideoId: videoId, UploadId: uploadId}, nil
}
