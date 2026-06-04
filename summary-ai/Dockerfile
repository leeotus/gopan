# summary-ai：基于 Whisper + LLM 的视频音轨听译 + 摘要生成服务。
# 需要 ffmpeg 提取音频。镜像默认 CPU-only torch；Whisper tiny 模型首次启动自动下载。
FROM docker.m.daocloud.io/python:3.11-slim

WORKDIR /app

# 系统依赖：ffmpeg（音频抽取）、libgomp（torch 依赖）
RUN apt-get update && apt-get install -y --no-install-recommends ffmpeg libgomp1 \
    && rm -rf /var/lib/apt/lists/*

# CPU-only PyTorch
RUN pip install --no-cache-dir torch --index-url https://download.pytorch.org/whl/cpu

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY main.py .

# Whisper 模型缓存目录，建议挂卷持久化
ENV XDG_CACHE_HOME=/root/.cache

EXPOSE 9920

CMD ["python", "main.py"]
