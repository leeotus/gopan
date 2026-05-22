<template>
  <div class="page-container">
    <div class="page-content" style="padding:14px" v-if="auth.isLoggedIn">
      <van-form @submit="handleUpload">
        <div class="card" style="padding:16px;margin-bottom:14px">
          <van-field v-model="form.title" placeholder="视频标题" :rules="[{ required: true }]" :style="fieldStyle" />
          <van-field v-model="form.description" placeholder="视频简介（选填）" type="textarea" rows="2" :style="fieldStyle" />
          <van-field v-model="form.category" placeholder="分类（选填）" :style="fieldStyle" />
        </div>

        <div class="card" style="padding:16px;margin-bottom:14px">
          <div class="label">封面图片</div>
          <van-uploader v-model="coverList" :max-count="1" accept="image/*" :after-read="onCoverRead" />
        </div>

        <div class="card" style="padding:16px;margin-bottom:20px">
          <div class="label">视频文件</div>
          <div class="upload-btn" v-if="!selectedFile" @click="triggerFileInput">
            <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="#5a5a7a" stroke-width="1.5"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="16"/><line x1="8" y1="12" x2="16" y2="12"/></svg>
            <span>选择视频</span>
          </div>
          <div v-else class="file-selected" style="padding:12px;background:var(--bg-input);border-radius:var(--radius-sm)">
            <div style="display:flex;align-items:center;gap:10px">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#22c55e" stroke-width="2"><polygon points="23 7 16 12 23 17 23 7"/><rect x="1" y="5" width="15" height="14" rx="2" ry="2"/></svg>
              <span style="flex:1;font-size:13px">{{ selectedFile.name }}</span>
              <svg @click="clearFile" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#ef4444" stroke-width="2" style="cursor:pointer"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            </div>
            <!-- 进度条 -->
            <div v-if="uploading" style="margin-top:12px">
              <div class="progress-bar-bg">
                <div class="progress-bar-fill" :style="{ width: progress + '%' }" />
              </div>
              <div class="progress-text">{{ uploadedChunks }}/{{ totalChunks }} 分片 · {{ Math.round(progress) }}%</div>
            </div>
          </div>
          <input ref="fileInput" type="file" accept="video/*" style="display:none" @change="onFileChange" />
        </div>

        <button class="btn-primary" type="submit" :disabled="uploading || !selectedFile" style="width:100%">
          {{ uploading ? "上传中..." : "开始上传" }}
        </button>
      </van-form>
    </div>
    <div v-else class="page-content" style="padding:14px">
      <van-empty description="登录后上传"><button class="btn-primary" @click="$router.push('/login')">去登录</button></van-empty>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { showToast } from "vant";
import { useAuthStore } from "../stores/auth";
import { videoApi } from "../api";

const CHUNK_SIZE = 5 * 1024 * 1024; // 5MB
const MAX_CONCURRENT = 3;

const auth = useAuthStore();
const uploading = ref(false);
const selectedFile = ref(null);
const coverList = ref([]);
const coverFile = ref(null);
const fileInput = ref(null);
const form = ref({ title: "", description: "", category: "" });
const fieldStyle = { '--van-field-background': 'transparent', '--van-field-input-text-color': '#e8e6f0', '--van-field-placeholder-text-color': '#5a5a7a' };

// 进度
const totalChunks = ref(0);
const uploadedChunks = ref(0);
const progress = ref(0);

function triggerFileInput() { fileInput.value?.click(); }
function onFileChange(e) { selectedFile.value = e.target.files?.[0] || null; }
function onCoverRead(file) { coverFile.value = file.file; }
function clearFile() { selectedFile.value = null; totalChunks.value = 0; uploadedChunks.value = 0; progress.value = 0; }

async function handleUpload() {
  if (!selectedFile.value) return;
  uploading.value = true;

  try {
    const file = selectedFile.value;
    totalChunks.value = Math.ceil(file.size / CHUNK_SIZE);
    uploadedChunks.value = 0;
    progress.value = 0;

    // 1. InitUpload
    const initRes = await videoApi.initUpload({
      filename: file.name,
      title: form.value.title,
      file_size: file.size,
      total_chunks: totalChunks.value,
    });
    const { video_id, upload_id } = initRes;

    // 2. 并发上传所有分片
    const chunks = sliceFile(file);
    await uploadChunksWithConcurrency(chunks, video_id, upload_id);

    // 3. upload-status 确认完整性
    const statusRes = await videoApi.uploadStatus({ upload_id });
    const received = statusRes.received_chunks || [];
    const allIndexes = Array.from({ length: totalChunks.value }, (_, i) => i);
    const missing = allIndexes.filter(i => !received.includes(i));

    if (missing.length > 0) {
      showToast(`补传 ${missing.length} 个缺失分片...`);
      const missingChunks = missing.map(i => chunks[i]);
      await uploadChunksWithConcurrency(missingChunks, video_id, upload_id);
    }

    // 4. MergeChunks
    const mergeRes = await videoApi.mergeChunks({ video_id, upload_id });
    if (mergeRes.status === "incomplete") {
      showToast(`合并失败，缺失分片: ${mergeRes.missing_chunks?.join(", ")}`);
      return;
    }

    showToast("上传成功！转码完成后可播放");
    form.value = { title: "", description: "", category: "" };
    selectedFile.value = null;
    coverList.value = [];
    coverFile.value = null;
    totalChunks.value = 0;
    uploadedChunks.value = 0;
    progress.value = 0;
  } catch (e) {
    showToast(e.message || "上传失败，请稍后再试");
  } finally {
    uploading.value = false;
  }
}

// 切分文件
function sliceFile(file) {
  const chunks = [];
  for (let offset = 0; offset < file.size; offset += CHUNK_SIZE) {
    chunks.push({
      index: Math.floor(offset / CHUNK_SIZE),
      blob: file.slice(offset, Math.min(offset + CHUNK_SIZE, file.size)),
      size: Math.min(CHUNK_SIZE, file.size - offset),
    });
  }
  return chunks;
}

// 并发上传
async function uploadChunksWithConcurrency(chunks, videoId, uploadId) {
  const queue = [...chunks];
  const workers = [];

  async function worker() {
    while (queue.length > 0) {
      const chunk = queue.shift();

      // 最多重试 3 次
      for (let retry = 0; retry < 3; retry++) {
        try {
          const formData = new FormData();
          formData.append("video_id", videoId);
          formData.append("upload_id", uploadId);
          formData.append("chunk_index", chunk.index);
          formData.append("file_size", chunk.size);
          formData.append("file", chunk.blob);
          await videoApi.uploadChunk(formData);
          uploadedChunks.value++;
          progress.value = (uploadedChunks.value / totalChunks.value) * 100;
          break; // 成功，跳出重试循环
        } catch {
          if (retry === 2) queue.push(chunk); // 3 次都失败，放回队列
          await sleep(1000 * (retry + 1));   // 等 1s / 2s / 3s 后重试
        }
      }
    }
  }

  for (let i = 0; i < MAX_CONCURRENT; i++) workers.push(worker());
  await Promise.all(workers);
}

function sleep(ms) { return new Promise(r => setTimeout(r, ms)); }
</script>

<style scoped>
.label { font-size: 13px; font-weight: 600; margin-bottom: 10px; color: var(--text-secondary); }
.upload-btn {
  width: 100%; height: 110px; border: 2px dashed var(--border); border-radius: var(--radius);
  display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 8px;
  color: var(--text-muted); font-size: 13px; cursor: pointer; transition: all var(--transition);
}
.upload-btn:active { border-color: var(--accent); background: rgba(139,92,246,0.05); }
.progress-bar-bg { width: 100%; height: 6px; background: var(--bg-card); border-radius: 3px; overflow: hidden; }
.progress-bar-fill { height: 100%; background: linear-gradient(90deg, var(--accent), #c084fc); transition: width 0.3s; }
.progress-text { font-size: 11px; color: var(--text-muted); margin-top: 4px; text-align: right; }
</style>
