package video

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/zeromicro/go-zero/core/logx"
)

type UploadCoverLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUploadCoverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadCoverLogic {
	return &UploadCoverLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// UploadCover 上传视频封面前端接口。
func (l *UploadCoverLogic) UploadCover(r *http.Request) (resp *types.BaseResp, err error) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return &types.BaseResp{Message: "解析失败"}, nil
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		return &types.BaseResp{Message: "缺少封面文件"}, nil
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return &types.BaseResp{Message: "读取文件失败"}, nil
	}

	videoIdStr := r.FormValue("video_id")
	videoId, _ := strconv.ParseInt(videoIdStr, 10, 64)
	if videoId == 0 {
		return &types.BaseResp{Message: "缺少 video_id"}, nil
	}

	// 上传到 MinIO（gateway 直接连 MinIO）
	client, err := minio.New("minio:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("minioadmin", "minioadmin", ""),
		Secure: false,
	})
	if err != nil {
		l.Logger.Errorf("minio connect error: %v", err)
		return &types.BaseResp{Message: "存储服务不可用"}, nil
	}

	coverKey := "covers/" + strconv.FormatInt(videoId, 10) + ".jpg"
	_, err = client.PutObject(l.ctx, "gopan-videos", coverKey, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{ContentType: "image/jpeg"})
	if err != nil {
		l.Logger.Errorf("upload cover error: %v", err)
		return &types.BaseResp{Message: "封面上传失败"}, nil
	}

	coverUrl := "/videos/../covers/" + strconv.FormatInt(videoId, 10) + ".jpg"
	return &types.BaseResp{Message: coverUrl}, nil
}
