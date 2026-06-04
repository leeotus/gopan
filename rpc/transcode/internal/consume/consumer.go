// package consume 实现 Kafka Consumer，消费转码任务并执行 FFmpeg 1080p HLS 转码。
package consume

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopan/common/kafka"
	"gopan/common/storage"
	"gopan/rpc/transcode/internal/svc"
	"gopan/rpc/video/videoclient"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/core/logx"
)

// StartConsumer 启动 Kafka 消费者（阻塞运行）。
func StartConsumer(ctx context.Context, svcCtx *svc.ServiceContext) {
	reader := kafka.NewConsumer(svcCtx.Config.Kafka.Brokers, svcCtx.Config.Kafka.TranscodeTopic, "gopan-transcode-worker")
	defer reader.Close()

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if err == context.Canceled {
				return
			}
			logx.Errorf("kafka fetch error: %v", err)
			continue
		}

		var task kafka.TranscodeTask
		if err := json.Unmarshal(msg.Value, &task); err != nil {
			logx.Errorf("kafka unmarshal task error: %v", err)
			reader.CommitMessages(ctx, msg)
			continue
		}

		logx.Infof("kafka consumed transcode task: video_id=%d, key=%s", task.VideoId, string(msg.Key))

		if err := processTranscode(ctx, svcCtx, &task); err != nil {
			logx.Errorf("transcode failed: video_id=%d, err=%v", task.VideoId, err)
			svcCtx.VideoClient.TranscodeCallback(ctx, &videoclient.TranscodeCallbackReq{
				VideoId: task.VideoId,
				Status:  3,
			})
		}

		reader.CommitMessages(ctx, msg)
	}
}

// processTranscode 执行单条转码任务：
//  1. 从 MinIO 下载源文件到工作目录
//  2. FFmpeg 1080p HLS 转码（libx264 + AAC，hls_time=10s）
//  3. 上传 HLS 切片 + index.m3u8 到 MinIO
//  4. 回调 video-svc.TranscodeCallback（status=2 成功，3 失败）
func processTranscode(ctx context.Context, svcCtx *svc.ServiceContext, task *kafka.TranscodeTask) error {
	workDir := svcCtx.Config.WorkDir
	if workDir == "" {
		workDir = "/tmp/gopan-transcode"
	}
	taskDir := filepath.Join(workDir, fmt.Sprintf("video-%d", task.VideoId))
	os.MkdirAll(taskDir, 0755)
	defer os.RemoveAll(taskDir)

	// 1. 下载源文件
	inputPath := filepath.Join(taskDir, "input.mp4")
	if err := downloadFromMinio(ctx, svcCtx.MinioClient, task.ObjectKey, inputPath); err != nil {
		return fmt.Errorf("download source: %w", err)
	}

	// 2. FFmpeg 1080p HLS 转码
	hlsDir := filepath.Join(taskDir, "1080p")
	os.MkdirAll(hlsDir, 0755)
	ffmpegPath := svcCtx.Config.FFmpeg.Path
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}

	cmd := exec.CommandContext(ctx, ffmpegPath,
		"-i", inputPath,
		"-c:v", "libx264", "-preset", "fast",
		"-b:v", "5000k", "-maxrate", "5000k", "-bufsize", "10000k",
		"-s", "1920x1080",
		"-c:a", "aac", "-b:a", "128k",
		"-hls_time", "10",
		"-hls_list_size", "0",
		"-hls_segment_filename", filepath.Join(hlsDir, "segment_%03d.ts"),
		"-f", "hls", filepath.Join(hlsDir, "index.m3u8"),
	)
	cmd.Dir = taskDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		logx.Errorf("ffmpeg failed: %v, output: %s", err, string(output))
		return fmt.Errorf("ffmpeg transcode: %w", err)
	}
	logx.Infof("ffmpeg done: video_id=%d", task.VideoId)

	// 3. 上传 HLS 文件到 MinIO
	prefix := fmt.Sprintf("videos/%d/1080p", task.VideoId)
	entries, _ := os.ReadDir(hlsDir)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if err := uploadToMinio(ctx, svcCtx.MinioClient, filepath.Join(hlsDir, e.Name()), prefix+"/"+e.Name()); err != nil {
			return fmt.Errorf("upload %s: %w", e.Name(), err)
		}
	}

	// 4. 回调 video-svc
	_, err = svcCtx.VideoClient.TranscodeCallback(ctx, &videoclient.TranscodeCallbackReq{
		VideoId: task.VideoId,
		Status:  2,
		Transcodes: []*videoclient.TranscodeInfo{
			{Resolution: "1080p", M3U8Url: svcCtx.MinioClient.ObjectURL(prefix + "/index.m3u8"), Bitrate: 5000},
		},
	})
	if err != nil {
		return fmt.Errorf("callback video-svc: %w", err)
	}
	logx.Infof("transcode completed: video_id=%d", task.VideoId)
	return nil
}

func downloadFromMinio(ctx context.Context, client *storage.MinioClient, key, destPath string) error {
	reader, err := client.GetObject(ctx, key)
	if err != nil {
		return err
	}
	defer reader.Close()
	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, reader)
	return err
}

func uploadToMinio(ctx context.Context, client *storage.MinioClient, localPath, minioKey string) error {
	f, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	contentType := "application/octet-stream"
	if strings.HasSuffix(minioKey, ".m3u8") {
		contentType = "application/vnd.apple.mpegurl"
	} else if strings.HasSuffix(minioKey, ".ts") {
		contentType = "video/mp2t"
	}
	return client.PutObject(ctx, minioKey, f, fi.Size(), contentType)
}

// StartMergeConsumer 消费合并任务——下载 chunks → 合并 → 上传 → 回写状态 → 发转码任务
func StartMergeConsumer(ctx context.Context, svcCtx *svc.ServiceContext) {
	if svcCtx.Config.Kafka.MergeTopic == "" {
		return
	}
	reader := kafka.NewConsumer(svcCtx.Config.Kafka.Brokers, svcCtx.Config.Kafka.MergeTopic, "gopan-merge-worker")
	defer reader.Close()

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if err == context.Canceled {
				return
			}
			logx.Errorf("merge kafka fetch error: %v", err)
			continue
		}

		var task kafka.MergeTask
		if err := json.Unmarshal(msg.Value, &task); err != nil {
			reader.CommitMessages(ctx, msg)
			continue
		}

		logx.Infof("merge consumer: video_id=%d chunks=%d", task.VideoId, len(task.ChunkKeys))
		if err := processMerge(ctx, svcCtx, &task); err != nil {
			logx.Errorf("merge failed: video_id=%d err=%v", task.VideoId, err)
		}
		reader.CommitMessages(ctx, msg)
	}
}

func processMerge(ctx context.Context, svcCtx *svc.ServiceContext, task *kafka.MergeTask) error {
	var buf bytes.Buffer
	for _, key := range task.ChunkKeys {
		reader, err := svcCtx.MinioClient.GetObject(ctx, key)
		if err != nil {
			return fmt.Errorf("get chunk %s: %w", key, err)
		}
		if _, err := io.Copy(&buf, reader); err != nil {
			reader.Close()
			return err
		}
		reader.Close()
	}

	destKey := fmt.Sprintf("videos/%d/source.mp4", task.VideoId)
	if err := svcCtx.MinioClient.PutObject(ctx, destKey, bytes.NewReader(buf.Bytes()), int64(buf.Len()), "video/mp4"); err != nil {
		return fmt.Errorf("put merged: %w", err)
	}

	logx.Infof("merge async done: video_id=%d", task.VideoId)

	// 发转码任务
	transcodeTask := kafka.TranscodeTask{VideoId: task.VideoId, ObjectKey: destKey}
	taskBody, _ := json.Marshal(transcodeTask)
	if svcCtx.KafkaWriter != nil {
		if err := svcCtx.KafkaWriter.WriteMessages(ctx, kafkago.Message{
			Key:   []byte(fmt.Sprintf("transcode-video-%d", task.VideoId)),
			Value: taskBody,
		}); err != nil {
			logx.Errorf("kafka write transcode task after merge error: %v", err)
		} else {
			logx.Infof("transcode task sent after merge: video_id=%d", task.VideoId)
		}
	}

	return nil
}
