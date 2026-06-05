import io
import torch
from fastapi import FastAPI, UploadFile, File, HTTPException
from pydantic import BaseModel, Field
from PIL import Image
from transformers import ChineseCLIPProcessor, ChineseCLIPModel

app = FastAPI(
    title="GoPan AI Vectorization Service - CPU Edition",
    description="本地离线多模态大模型向量化提取接口 (强制 CPU 版、深度兼容)"
)

device = "cpu"
print(f"[Device Audit] Running AI Inference on: {device}")

MODEL_NAME = "OFA-Sys/chinese-clip-vit-base-patch16"
model = None
processor = None

@app.on_event("startup")
def load_clip_model():
    global model, processor
    try:
        print("[CPU STARTUP] Loading CLIP model to system memory ...")
        model = ChineseCLIPModel.from_pretrained(MODEL_NAME).to(device)
        processor = ChineseCLIPProcessor.from_pretrained(MODEL_NAME)
        model.eval()
        print("[Model Load] Model loaded successfully on SYSTEM CPU!")
    except Exception as e:
        print(f"[Fatal] Failed to load local CLIP model: {e}")
        model = None

class TextEmbeddingRequest(BaseModel):
    text: str = Field(..., description="待翻译向量化的中文文本句子")

@app.get("/health")
def health_check():
    if model is None:
        return {"status": "unhealthy", "message": "Model not loaded", "device": device}
    return {
        "status": "healthy", 
        "device": device,
        "cuda_available": torch.cuda.is_available()
    }

@app.post("/embed/text", summary="中文文本 -> 512 维归一化向量")
def embed_text(req: TextEmbeddingRequest):
    if model is None:
        raise HTTPException(status_code=503, detail="AI Model not ready yet.")
    if not req.text.strip():
        raise HTTPException(status_code=400, detail="Text string cannot be empty")
    
    try:
        with torch.no_grad():
            inputs = processor(text=[req.text], return_tensors="pt", padding=True).to(device)
            result = model.get_text_features(**inputs)
            
            # 【智高防护分支 - Shape-Aware Fallback Engine】
            features = None
            
            # 1. 尝试直接验证 tensor 类别
            if isinstance(result, torch.Tensor):
                features = result
                print("[Match Direct] Result is already a standard torch.Tensor")
            else:
                # 2. 扫描复合对象中的所有成员，动态搜寻已被内置 projected 的 [1, 512] Tensor
                for attr_name in dir(result):
                    try:
                        val = getattr(result, attr_name)
                        if isinstance(val, torch.Tensor) and list(val.shape) == [1, 512]:
                            features = val
                            print(f"[Match Dynamic] Found 512-dim tensor in result.{attr_name}")
                            break
                    except Exception:
                        continue
                
                # 3. 若无匹配，手动降级安全提取
                if features is None:
                    print("[Match Manual] No direct 512-dim tensor found. Extracting manually.")
                    pooled = result.pooler_output if hasattr(result, "pooler_output") else result[1]
                    # 精准验证底层尺寸，隐式做矩阵安全投影乘法
                    if pooled.shape[-1] == 768:
                        features = model.text_projection(pooled)
                    else:
                        features = pooled
            
            # 100% 成功脱离，执行标准的 L2 规范化
            features /= features.norm(dim=-1, keepdim=True)
            
            vector = features[0].cpu().numpy().tolist()
            return {"dimension": len(vector), "vector": vector}
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Text Inference Error: {str(e)}")

@app.post("/embed/image", summary="视频帧图片 -> 512 维归一化向量")
async def embed_image(file: UploadFile = File(..., description="视频帧截图文件")):
    if model is None:
        raise HTTPException(status_code=503, detail="AI Model not ready yet.")
    
    try:
        image_bytes = await file.read()
        image = Image.open(io.BytesIO(image_bytes)).convert("RGB")
    except Exception:
        raise HTTPException(status_code=400, detail="Uploaded file is not a valid image format")

    try:
        with torch.no_grad():
            inputs = processor(images=image, return_tensors="pt").to(device)
            result = model.get_image_features(**inputs)
            
            # 【智高防护二代】视觉端防漏：动态自适应特征抽取
            features = None
            if isinstance(result, torch.Tensor):
                features = result
            else:
                for attr_name in dir(result):
                    try:
                        val = getattr(result, attr_name)
                        if isinstance(val, torch.Tensor) and list(val.shape) == [1, 512]:
                            features = val
                            break
                    except Exception:
                        continue
                
                if features is None:
                    pooled = result.pooler_output if hasattr(result, "pooler_output") else result[1]
                    if pooled.shape[-1] == 768:
                        features = model.visual_projection(pooled)
                    else:
                        features = pooled
            
            features /= features.norm(dim=-1, keepdim=True)
            
            vector = features[0].cpu().numpy().tolist()
            return {"dimension": len(vector), "vector": vector}
    except Exception as e:
         raise HTTPException(status_code=500, detail=f"Image Inference Error: {str(e)}")

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=9900)
