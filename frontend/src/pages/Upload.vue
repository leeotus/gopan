<template>
  <div class="page-container">
    <div class="page-content" style="padding:14px" v-if="auth.isLoggedIn">
      <div class="card" style="padding:16px;margin-bottom:14px">
        <van-field v-model="form.title" placeholder="视频标题" :rules="[{ required: true }]" />
        <van-field v-model="form.description" placeholder="视频简介（选填）" type="textarea" rows="2" />
        <van-field v-model="form.category" placeholder="分类（选填）" />
      </div>
      <div class="card" style="padding:16px;margin-bottom:20px">
        <div class="card-label">视频文件</div>
        <div v-if="!selectedFile" class="upload-btn" @click="triggerFileInput">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="#5a5a7a" stroke-width="1.5"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="16"/><line x1="8" y1="12" x2="16" y2="12"/></svg>
          <span>选择视频</span>
        </div>
        <div v-else class="file-info">
          <span>{{ selectedFile.name }}</span>
          <button class="btn-clear" @click="selectedFile = null">✕</button>
        </div>
        <input ref="fileInput" type="file" accept="video/*" style="display:none" @change="onFileChange" />
      </div>
      <div v-if="uploading" class="card" style="padding:16px;margin-bottom:14px">
        <div class="progress-out">
          <div class="progress-in" :style="{width: percent + '%'}"></div>
        </div>
        <div class="progress-label">{{ doneChunks }} / {{ total }} 分片 ({{ percent }}%)</div>
      </div>
      <button class="btn-primary" @click="start" :disabled="uploading || !selectedFile" style="width:100%">
        {{ uploading ? "上传中..." : "开始上传" }}
      </button>
    </div>
    <div v-else class="page-content" style="padding:14px;text-align:center">
      <p style="color:var(--text-muted);margin-bottom:16px">登录后上传</p>
      <button class="btn-primary" @click="$router.push('/login')">去登录</button>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { showToast } from "vant";
import { useAuthStore } from "../stores/auth";
import axios from "axios";

const CHUNK = 3 * 1024 * 1024; // 3MB，低于 gRPC 默认 4MB 限制
const auth = useAuthStore();
const selectedFile = ref(null);
const fileInput = ref(null);
const form = ref({ title: "", description: "", category: "" });
const uploading = ref(false);
const total = ref(0);
const doneChunks = ref(0);
const percent = ref(0);

function triggerFileInput() { fileInput.value?.click(); }
function onFileChange(e) { selectedFile.value = e.target.files?.[0] || null; }

async function start() {
  if (!selectedFile.value) return;
  uploading.value = true;
  const file = selectedFile.value;
  total.value = Math.ceil(file.size / CHUNK);
  doneChunks.value = 0;
  percent.value = 0;

  try {
    // 1. init
    const tok = localStorage.getItem("token");
    const hdr = { Authorization: "Bearer " + tok, "Content-Type": "application/json" };
    const init = await axios.post("/api/video/init-upload", {
      filename: file.name, title: form.value.title, file_size: file.size, total_chunks: total.value,
    }, { headers: hdr });
    const vid = init.data.video_id;
    const uid = init.data.upload_id;

    // 2. upload chunks concurrently (batch of 3)
    const MAX_CONCURRENT = 3;
    const uploadOne = async (i) => {
      const start = i * CHUNK;
      const end = Math.min(start + CHUNK, file.size);
      const fd = new FormData();
      fd.append("video_id", vid);
      fd.append("upload_id", uid);
      fd.append("chunk_index", i);
      fd.append("file_size", end - start);
      fd.append("file", file.slice(start, end), "chunk_" + i);

      for (let retry = 0; retry < 3; retry++) {
        try {
          await axios.post("/api/video/upload-chunk", fd, {
            headers: { Authorization: "Bearer " + tok },
            timeout: 60000,
          });
          return;
        } catch {
          if (retry < 2) await new Promise(r => setTimeout(r, 1000 * (retry + 1)));
        }
      }
    };

    for (let batch = 0; batch < total.value; batch += MAX_CONCURRENT) {
      const tasks = [];
      for (let i = batch; i < Math.min(batch + MAX_CONCURRENT, total.value); i++) {
        tasks.push(uploadOne(i));
      }
      await Promise.all(tasks);
      doneChunks.value = Math.min(batch + MAX_CONCURRENT, total.value);
      percent.value = Math.round((doneChunks.value / total.value) * 100);
    }

    // 3. upload-status confirm
    const st = await axios.get("/api/video/upload-status?upload_id=" + uid, { headers: { Authorization: "Bearer " + tok } });
    const rcvd = st.data.received_chunks || st.data.data?.received_chunks || [];
    const all = Array.from({length: total.value}, (_, i) => i);
    const miss = all.filter(i => !rcvd.includes(i));
    if (miss.length > 0) { showToast("缺失 " + miss.length + " 个分片，请重试"); return; }

    // 4. merge
    await axios.post("/api/video/merge-chunks", { video_id: vid, upload_id: uid }, { headers: hdr });
    showToast("上传成功！");
    selectedFile.value = null;
    form.value = { title: "", description: "", category: "" };
  } catch (e) {
    showToast(e?.response?.data?.message || e.message || "上传失败");
  } finally {
    uploading.value = false;
  }
}
</script>

<style scoped>
.card { background: var(--bg-card); border-radius: var(--radius); border: 1px solid var(--border); }
.card-label { font-size: 13px; font-weight: 600; margin-bottom: 10px; color: var(--text-secondary); }
.btn-primary { display: block; padding: 14px; border: none; border-radius: var(--radius); background: linear-gradient(135deg, var(--accent), #7c3aed); color: #fff; font-size: 15px; font-weight: 600; cursor: pointer; box-shadow: 0 4px 16px var(--accent-glow); }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.upload-btn { width: 100%; height: 110px; border: 2px dashed var(--border); border-radius: var(--radius); display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 8px; color: var(--text-muted); font-size: 13px; cursor: pointer; }
.upload-btn:active { border-color: var(--accent); }
.file-info { display: flex; align-items: center; gap: 8px; padding: 10px; background: var(--bg-input); border-radius: var(--radius-sm); font-size: 13px; }
.btn-clear { background: none; border: none; color: var(--danger); cursor: pointer; font-size: 16px; }
.progress-out { width: 100%; height: 8px; background: var(--bg-primary); border-radius: 4px; overflow: hidden; margin-bottom: 8px; }
.progress-in { height: 100%; background: linear-gradient(90deg, var(--accent), #c084fc); transition: width 0.3s; }
.progress-label { font-size: 12px; color: var(--text-muted); text-align: right; }
</style>
