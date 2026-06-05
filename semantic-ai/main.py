import io
import torch
from fastapi import FastAPI, UploadFile, File, HTTPException
from pydantic import BaseModel, Field
from PIL import Image
from transformers import ChineseCLIPProcessor, ChineseCLIPModel

app = FastAPI(
    title="GoPan AI Vectorization Service",
    description="本地离线多模态大模型向量化提取接口"
)

# 1. 检查 CPU/GPU。GTX 1650 具有 4GB 显存，RTX 3060 具有 12GB 显存，均完美支持 CUDA 加速。
device = "cuda" if torch.cuda.is_available() else "cpu"
print(f"[Device Audit] Running AI Inference on: {device}")

# 2. 空白预热，让 OFA-Sys 模型热注册
MODEL_NAME = "OFA-Sys/chinese-clip-vit-base-patch16"
model = None
processor = None

@app.on_event("startup")
def load_clip_model():
    """服务启动时，预先将模型装载进 GPU 显存，减少后续请求耗时"""
    global model, processor
    try:
        print("[V4 PATCHED STARTUP] Loading CLIP model to GPU/CPU ...")
        model = ChineseCLIPModel.from_pretrained(MODEL_NAME).to(device)
        processor = ChineseCLIPProcessor.from_pretrained(MODEL_NAME)
        model.eval()  # 设置为推理（评估）模式，停用 Dropout 等机制
        print("[Model Load] Model loaded successfully on GPU/CPU!")
    except Exception as e:
        print(f"[Fatal] Failed to load local CLIP model: {e}")
        model = None

# 文本请求结构
class TextEmbeddingRequest(BaseModel):
    text: str = Field(..., description="待翻译向量化的中文文本句子")

@app.get("/health")
def health_check():
    """节点健康检测"""
    if model is None:
        return {"status": "unhealthy", "message": "Model not loaded", "device": device}
    return {
        "status": "healthy", 
        "device": device,
        "cuda_available": torch.cuda.is_available(),
        "gpu_name": torch.cuda.get_device_name(0) if torch.cuda.is_available() else "N/A"
    }

@app.post("/embed/text", summary="中文文本 -> 512 维归一化向量")
def embed_text(req: TextEmbeddingRequest):
    if model is None:
        raise HTTPException(status_code=503, detail="AI Model not ready yet.")
    if not req.text.strip():
        raise HTTPException(status_code=400, detail="Text string cannot be empty")
    
    try:
        with torch.no_grad(): # 不计算梯度，节省显存与算力
            # 对中文分词（Tokenize）
            inputs = processor(text=[req.text], return_tensors="pt", padding=True).to(device)

            # 使用官方提供的 get_text_features 辅助函数，直接拿 (B, 512) 文本投影向量。
            # 之前用 model.forward + outputs.text_embeds 在 transformers 新版本中会得到 None，
            # 因为不同时传入 image 时 text_embeds 不会被填充。
            text_features = model.get_text_features(
                input_ids=inputs["input_ids"],
                attention_mask=inputs["attention_mask"],
            )

            # 做 L2 规范化，直接用于余弦相似度计算
            text_features = text_features / text_features.norm(dim=-1, keepdim=True)

            # 转为普通的 float array 吐出给 GoPan 控制端
            vector = text_features[0].cpu().numpy().tolist()
            return {"dimension": len(vector), "vector": vector}
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Text Inference Error: {str(e)}")

@app.post("/embed/image", summary="视频帧图片 -> 512 维归一化向量")
async def embed_image(file: UploadFile = File(..., description="视频帧截图文件")):
    if model is None:
        raise HTTPException(status_code=503, detail="AI Model not ready yet.")
    
    try:
        # 读取字节流转换为 Pillow 格式
        image_bytes = await file.read()
        image = Image.open(io.BytesIO(image_bytes)).convert("RGB")
    except Exception:
        raise HTTPException(status_code=400, detail="Uploaded file is not a valid image format")

    try:
        with torch.no_grad():
            # 缩放至 CLIP 尺寸并进行基础视觉归一化
            inputs = processor(images=image, return_tensors="pt").to(device)

            # 同上：使用官方 get_image_features 拿 (B, 512) 图像投影向量
            image_features = model.get_image_features(pixel_values=inputs["pixel_values"])

            # L2 规范化
            image_features = image_features / image_features.norm(dim=-1, keepdim=True)

            vector = image_features[0].cpu().numpy().tolist()
            return {"dimension": len(vector), "vector": vector}
    except Exception as e:
         raise HTTPException(status_code=500, detail=f"Image Inference Error: {str(e)}")

if __name__ == "__main__":
    import uvicorn
    # 启动命令：python main.py
    uvicorn.run(app, host="0.0.0.0", port=9900)
