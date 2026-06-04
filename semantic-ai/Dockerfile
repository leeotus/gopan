# semantic-ai：基于 ChineseCLIP 的文本/图片向量化服务（用于 ES KNN 语义搜索）。
# 镜像默认走 CPU-only torch 以减小体积；如需 GPU 请把 torch 行换成官方 CUDA 版即可。
FROM docker.m.daocloud.io/python:3.11-slim

WORKDIR /app

# 系统依赖：libgomp（torch 依赖）
RUN apt-get update && apt-get install -y --no-install-recommends libgomp1 \
    && rm -rf /var/lib/apt/lists/*

# CPU-only PyTorch（~200MB 替代 2GB CUDA 版本）。如有 GPU 节点请单独构建。
RUN pip install --no-cache-dir torch --index-url https://download.pytorch.org/whl/cpu

COPY requirements.txt .
# torch 已装，requirements 里那行 transformers 等会拉，跳 torch
RUN grep -v '^torch' requirements.txt > /tmp/req.txt && \
    pip install --no-cache-dir -r /tmp/req.txt

COPY main.py main_cpu.py ./

# HuggingFace 模型会下到 /root/.cache/huggingface，建议挂卷持久化（compose 已配）
ENV HF_HOME=/root/.cache/huggingface

EXPOSE 9900

# 使用 main_cpu.py：它带 Shape-Aware fallback，兼容 transformers 新版返回 BaseModelOutputWithPooling 的情况。
CMD ["python", "main_cpu.py"]
