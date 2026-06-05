import os
import io
import time
import tempfile
import subprocess
import urllib.request
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
import whisper
import torch
from openai import OpenAI

app = FastAPI(
    title="GoPan AI Subtitle and Summarization Service",
    description="本地 Whisper 离线音频转文 + MiniMax/DeepSeek 智能摘要双引擎大网关"
)

# 1. 检查物理设备是否支持 GPU 加速
device = "cuda" if torch.cuda.is_available() else "cpu"
print(f"[Device Audit] Summary-AI executing on: {device}")

# 2. 预热模型：Whisper tiny 规格（仅约 70MB），CPU 推理秒速，1650 显卡无压力。
# 生产环境 3060 下可手动更改为 "base" 或 "small" 以提升中文听音精度。
WHISPER_MODEL_SIZE = "tiny"
whisper_model = None

@app.on_event("startup")
def load_whisper():
    global whisper_model
    try:
        print(f"[Whisper Loading] Loading model size: '{WHISPER_MODEL_SIZE}' in {device}...")
        whisper_model = whisper.load_model(WHISPER_MODEL_SIZE, device=device)
        print("[Whisper Loading] Whisper model loaded successfully!")
    except Exception as e:
        print(f"[Fatal] Failed to load Whisper: {e}")
        whisper_model = None

# 请求参数结构
class AnalyzeRequest(BaseModel):
    video_url: str = Field(..., description="MinIO 视频的公网/内网可下载播放 URL 链接")

# 格式化 VTT 时间轴
def format_vtt_time(seconds):
    hours = int(seconds // 3600)
    minutes = int((seconds % 3600) // 60)
    secs = int(seconds % 60)
    millis = int((seconds % 1) * 1000)
    return f"{hours:02d}:{minutes:02d}:{secs:02d}.{millis:03d}"

# 格式化 SRT 时间轴（以防万一需要 SRT）
def format_srt_time(seconds):
    hours = int(seconds // 3600)
    minutes = int((seconds % 3600) // 60)
    secs = int(seconds % 60)
    millis = int((seconds % 1) * 1000)
    return f"{hours:02d}:{minutes:02d}:{secs:02d},{millis:03d}"

# 调用 MiniMax 特有或兼容大模型获取摘要
def generate_llm_summary(full_text: str) -> str:
    minimax_key = os.getenv("MINIMAX_API_KEY")
    deepseek_key = os.getenv("DEEPSEEK_API_KEY")

    prompt = (
        f"你是一个专业的音视频内容理解摘要助手。这有一段视频音轨转换出的文本：\n"
        f"------\n{full_text}\n------\n"
        f"请提取核心，为本期视频生成一份 150 字以内极其精炼、条理清晰的剧情或内容大纲摘要。直接输出结论无需寒暄。"
    )

    # 1. 优先使用 MiniMax 大模型 (abab6.5g 生产级体验)
    if minimax_key:
        print("[LLM Summary] Calling MiniMax API...")
        try:
            client = OpenAI(
                api_key=minimax_key,
                base_url="https://api.minimax.chat/v1"
            )
            chat_completion = client.chat.completions.create(
                model="abab6.5g-chat",
                messages=[
                    {"role": "system", "content": "You are a helpful assistant."},
                    {"role": "user", "content": prompt}
                ],
                temperature=0.3
            )
            return chat_completion.choices[0].message.content.strip()
        except Exception as e:
            print(f"[LLM Warning] MiniMax request failed: {e}, attempting fallback.")

    # 2. 次选高性价比 DeepSeek 模型
    if deepseek_key:
        print("[LLM Summary] Calling DeepSeek API...")
        try:
            client = OpenAI(
                api_key=deepseek_key,
                base_url="https://api.deepseek.com"
            )
            chat_completion = client.chat.completions.create(
                model="deepseek-chat",
                messages=[
                    {"role": "system", "content": "你是一个专业的视频摘要生成器"},
                    {"role": "user", "content": prompt}
                ],
                temperature=0.2
            )
            return chat_completion.choices[0].message.content.strip()
        except Exception as e:
            print(f"[LLM Warning] DeepSeek request failed: {e}")

    # 3. 双方均无 Key 时的本地零费用沙盒智能兜底
    print("[LLM Summary] No LLM API Keys found in environments. Mocking responsive summary.")
    if len(full_text) > 40:
        return f"🚀 [AI 智能摘要（已本地生成）]: 视频旁白提到「{full_text[:40]}...」。本片由 GoPan 自动听译转写，音轨识别全文共 {len(full_text)} 字。想要体验真正大模型的极致分析，请在后台系统环境变量中配置 MINIMAX_API_KEY 或 DEEPSEEK_API_KEY。"
    return "🚀 [AI 智能摘要]: 音轨识别未发现明显对话语音，该视频可能是一个纯音乐、静态默片或纯粹的动作片段。"


@app.get("/health")
def health():
    return {
        "device": device,
        "whisper_ready": whisper_model is not None,
        "minimax_configured": os.getenv("MINIMAX_API_KEY") is not None,
        "deepseek_configured": os.getenv("DEEPSEEK_API_KEY") is not None
    }


@app.post("/analyze", summary="异步抽取视频音轨 -> Whisper 本地听译 -> MiniMax AI 快速大纲")
async def analyze_video(req: AnalyzeRequest):
    if whisper_model is None:
        raise HTTPException(status_code=503, detail="Whisper speech engine is not loaded yet")
    if not req.video_url.startswith("http"):
        raise HTTPException(status_code=400, detail="Invalid video_url. Must be a HTTP/HTTPS link")

    # 创建独立沙盒临时路径保护本地磁盘
    with tempfile.TemporaryDirectory() as temp_dir:
        video_path = os.path.join(temp_dir, "temp_video.mp4")
        audio_path = os.path.join(temp_dir, "extracted_audio.wav")

        print(f"[Step 1] Loading video file from MinIO: {req.video_url} ...")
        try:
            urllib.request.urlretrieve(req.video_url, video_path)
        except Exception as e:
            raise HTTPException(status_code=500, detail=f"Failed to download video from storage: {e}")

        # 预检：先用 ffprobe 探测是否含音轨。无音轨视频直接返回友好摘要，避免 ffmpeg 抽音失败。
        probe = subprocess.run(
            ["ffprobe", "-v", "error", "-select_streams", "a",
             "-show_entries", "stream=codec_type", "-of", "csv=p=0", video_path],
            stdout=subprocess.PIPE, stderr=subprocess.PIPE
        )
        if probe.returncode != 0 or not probe.stdout.strip():
            print("[Step 2-skip] Video has no audio stream, returning placeholder summary.")
            no_audio_summary = "🎬 该视频无音轨，无法进行语音听译。建议作者在简介中补充内容说明。"
            return {
                "elapsed_seconds": 0.0,
                "full_text": "",
                "summary": no_audio_summary,
                "vtt": "WEBVTT\n\n",
                "srt": "",
            }

        print("[Step 2] Extracting wav 16kHz mono audio track via FFmpeg ...")
        # 直接调用系统内置 FFMpeg（无需 Python 重依赖）
        cmd = [
            "ffmpeg", "-y", "-i", video_path,
            "-vn", "-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le",
            audio_path
        ]
        res = subprocess.run(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        if res.returncode != 0:
            err_log = res.stderr.decode("utf-8", errors="ignore")
            raise HTTPException(status_code=500, detail=f"FFmpeg extraction failed: {err_log}")

        print("[Step 3] Executing local Whisper automatic speech transcription...")
        start_time = time.time()
        # 听写音频。由于 Whisper tiny 极为紧凑，1.2GHz CPU 耗时通常只有音频时长 1/5。CUDA 更快。
        result = whisper_model.transcribe(audio_path, language="zh")
        elapsed = time.time() - start_time
        print(f"[Step 3] Transcription finished in {elapsed:.2f} seconds!")

        # 4. 转换拼接 WebVTT 格式和纯文本
        vtt_stream = io.StringIO()
        vtt_stream.write("WEBVTT\n\n")

        srt_stream = io.StringIO()

        full_sentences = []
        for index, seg in enumerate(result["segments"]):
            start = seg["start"]
            end = seg["end"]
            text = seg["text"].strip()
            if not text:
                continue

            full_sentences.append(text)

            # WebVTT
            vtt_stream.write(f"{format_vtt_time(start)} --> {format_vtt_time(end)}\n")
            vtt_stream.write(f"{text}\n\n")

            # SRT (Fallback)
            srt_stream.write(f"{index + 1}\n")
            srt_stream.write(f"{format_srt_time(start)} --> {format_srt_time(end)}\n")
            srt_stream.write(f"{text}\n\n")

        full_text = " ".join(full_sentences)
        print(f"[Speech Text Length]: {len(full_text)} characters.")

        # 5. 调用 MiniMax 或 DeepSeek 获得文本高阶理解摘要
        summary = generate_llm_summary(full_text)

        return {
            "elapsed_seconds": round(elapsed, 2),
            "full_text": full_text,
            "summary": summary,
            "vtt": vtt_stream.getvalue(),
            "srt": srt_stream.getvalue()
        }

if __name__ == "__main__":
    import uvicorn
    # 本期摘要子服务跑在 9920 端口
    uvicorn.run(app, host="0.0.0.0", port=9920)
