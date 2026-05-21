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
          <div v-else class="file-selected" style="display:flex;align-items:center;gap:10px;padding:12px;background:var(--bg-input);border-radius:var(--radius-sm)">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#22c55e" stroke-width="2"><polygon points="23 7 16 12 23 17 23 7"/><rect x="1" y="5" width="15" height="14" rx="2" ry="2"/></svg>
            <span style="flex:1;font-size:13px">{{ selectedFile.name }}</span>
            <svg @click="clearFile" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#ef4444" stroke-width="2" style="cursor:pointer"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
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

const auth = useAuthStore();
const uploading = ref(false);
const selectedFile = ref(null);
const coverList = ref([]);
const coverFile = ref(null);
const fileInput = ref(null);
const form = ref({ title: "", description: "", category: "" });
const fieldStyle = { '--van-field-background': 'transparent', '--van-field-input-text-color': '#e8e6f0', '--van-field-placeholder-text-color': '#5a5a7a' };

function triggerFileInput() { fileInput.value?.click(); }
function onFileChange(e) { selectedFile.value = e.target.files?.[0] || null; }
function onCoverRead(file) { coverFile.value = file.file; }
function clearFile() { selectedFile.value = null; }

async function handleUpload() {
  if (!selectedFile.value) return;
  showToast("上传功能开发中，敬请期待 🔧");
  // TODO: 分片上传服务端实现后启用以下逻辑
  // uploading.value = true;
  // try { ... } catch { ... } finally { uploading.value = false; }
}
</script>

<style scoped>
.label { font-size: 13px; font-weight: 600; margin-bottom: 10px; color: var(--text-secondary); }
.upload-btn {
  width: 100%; height: 110px; border: 2px dashed var(--border); border-radius: var(--radius);
  display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 8px;
  color: var(--text-muted); font-size: 13px; cursor: pointer; transition: all var(--transition);
}
.upload-btn:active { border-color: var(--accent); background: rgba(139,92,246,0.05); }
</style>
