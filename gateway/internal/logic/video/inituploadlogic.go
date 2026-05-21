package video

import (
	"context"
	"fmt"
	"time"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type InitUploadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInitUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitUploadLogic {
	return &InitUploadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// InitUpload 初始化上传，生成唯一的 upload_id 并返回。
// 真正的视频记录创建在 mergeChunks 时由 video-svc 完成。
func (l *InitUploadLogic) InitUpload(req *types.InitUploadReq) (resp *types.InitUploadResp, err error) {
	uploadId := fmt.Sprintf("upload_%d_%d", time.Now().UnixMilli(), time.Now().Nanosecond()%10000)
	return &types.InitUploadResp{
		VideoId:  0, // merge 后由 video-svc 分配真正的 ID
		UploadId: uploadId,
	}, nil
}
