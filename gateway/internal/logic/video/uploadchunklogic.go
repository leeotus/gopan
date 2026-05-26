package video

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	"gopan/gateway/internal/svc"
	"gopan/gateway/internal/types"
	"gopan/rpc/video/videoclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadChunkLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewUploadChunkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadChunkLogic {
	return &UploadChunkLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *UploadChunkLogic) SetRequest(r *http.Request) { l.r = r }

func (l *UploadChunkLogic) UploadChunk() (resp *types.BaseResp, err error) {
	if l.r == nil {
		return &types.BaseResp{Message: "no request"}, nil
	}

	if err := l.r.ParseMultipartForm(100 << 20); err != nil {
		return &types.BaseResp{Message: "解析上传数据失败"}, nil
	}

	uploadId := l.r.FormValue("upload_id")
	videoIdStr := l.r.FormValue("video_id")
	chunkIndexStr := l.r.FormValue("chunk_index")

	file, _, err := l.r.FormFile("file")
	if err != nil {
		return &types.BaseResp{Message: "未找到上传文件"}, nil
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return &types.BaseResp{Message: "读取文件失败"}, nil
	}

	ctx, cancel := context.WithTimeout(l.ctx, 30*time.Second)
	defer cancel()

	start := time.Now()
	gRPCResp, gRPCErr := l.svcCtx.VideoClient.UploadChunk(ctx, &videoclient.UploadChunkReq{
		UploadId:   uploadId,
		VideoId:    parseInt64(videoIdStr),
		ChunkIndex: parseInt32(chunkIndexStr),
		FileSize:   int32(len(data)),
		Data:       data,
	})
	elapsed := time.Since(start)

	if gRPCErr != nil {
		l.Logger.Errorf("upload-chunk rpc FAIL: video=%s chunk=%s size=%d elapsed=%v err=%v", videoIdStr, chunkIndexStr, len(data), elapsed, gRPCErr)
		return &types.BaseResp{Message: gRPCErr.Error()}, nil
	}

	_ = gRPCResp
	l.Logger.Infof("upload-chunk rpc OK: video=%s chunk=%s size=%d elapsed=%v", videoIdStr, chunkIndexStr, len(data), elapsed)
	return &types.BaseResp{Message: "ok"}, nil
}

func parseInt64(s string) int64 { v, _ := strconv.ParseInt(s, 10, 64); return v }
func parseInt32(s string) int32 { v, _ := strconv.ParseInt(s, 10, 32); return int32(v) }
