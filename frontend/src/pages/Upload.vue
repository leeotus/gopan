<template>
  <div class="page-container">
    <div class="page-content" style="padding:16px;max-width:500px;margin:0 auto">
      <!-- Video file selection -->
      <div class="card upload-card" style="margin-bottom:16px">
        <div class="upload-zone" @click="triggerFile" @dragover.prevent @drop.prevent="onDrop">
          <div class="upload-icon" v-if="!file">⬆</div>
          <div class="upload-icon" v-else>📹</div>
          <p class="upload-text" v-if="!file">Drop your video here or click to browse</p>
          <p class="upload-text" v-else>{{ file.name }}</p>
          <p class="upload-hint">MP4 · Max 500MB · {{ file ? formatSize(file.size) : '' }}</p>
          <input ref="fileInput" type="file" accept="video/mp4" style="display:none" @change="onFileChange" />
        </div>
      </div>

      <!-- Video metadata (always visible) -->
      <div class="card" style="padding:20px;margin-bottom:16px">
        <div class="section-title text-muted" style="margin-bottom:16px">VIDEO INFO</div>

        <!-- Cover image (required) -->
        <label class="field-label">Cover Image <span class="required">*</span></label>
        <div class="cover-upload" @click="triggerCoverInput">
          <img v-if="coverPreview" :src="coverPreview" class="cover-preview" />
          <div v-else class="cover-placeholder">
            <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="var(--text-muted)" stroke-width="1.5"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>
            <span class="cover-hint">Click to upload cover</span>
          </div>
          <input ref="coverInput" type="file" accept="image/*" style="display:none" @change="onCoverChange" />
        </div>

        <!-- Title (required) -->
        <label class="field-label" style="margin-top:16px">Title <span class="required">*</span></label>
        <input class="input-field" v-model="title" placeholder="Enter a catchy title..." maxlength="100" style="margin-bottom:16px" />

        <!-- Description (optional) -->
        <label class="field-label">Description</label>
        <textarea class="input-field textarea" v-model="description" placeholder="Describe your video..." maxlength="500" style="margin-bottom:16px"></textarea>

        <!-- Category -->
        <label class="field-label">Category</label>
        <div class="category-grid">
          <button v-for="cat in categories" :key="cat"
            :class="['cat-option', { active: selectedCategory === cat }]"
            @click="selectedCategory = cat"
          >{{ cat }}</button>
        </div>
      </div>

      <!-- Upload button -->
      <button
        v-if="file"
        class="btn-primary btn-primary--solid"
        @click="startUpload"
        :disabled="uploading || !canUpload"
        style="width:100%;padding:14px"
      >
        {{ uploading ? `Uploading ${progress}%` : 'Publish Video' }}
      </button>
      <p v-if="file && !canUpload" class="text-muted" style="text-align:center;margin-top:8px;font-size:12px">
        Please enter a video title
      </p>

      <!-- Progress -->
      <div v-if="uploading" class="progress-bar" style="margin-top:16px">
        <div class="progress-fill" :style="{ width: progress + '%' }"></div>
        <span class="progress-text">{{ progress }}%</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from "vue";
import { useRouter } from "vue-router";
import { showToast } from "vant";
import { videoApi } from "../api";
import request from "../api/request";
import { useAuthStore } from "../stores/auth";

const auth = useAuthStore();
const router = useRouter();

const file = ref(null);
const title = ref("");
const description = ref("");
const selectedCategory = ref("");
const coverFile = ref(null);
const coverPreview = ref("");
const uploading = ref(false);
const progress = ref(0);
const fileInput = ref(null);
const coverInput = ref(null);

const categories = ["Tech", "Music", "Gaming", "Sports", "Education", "Entertainment", "Other"];
const canUpload = computed(() => title.value.trim() && file.value);
const coverPath = computed(() => coverFile.value ? `covers/${Date.now()}_${coverFile.value.name}` : "");

function triggerFile() { fileInput.value?.click(); }
function triggerCoverInput() { coverInput.value?.click(); }
function onFileChange(e) { file.value = e.target.files[0]; }
function onDrop(e) { file.value = e.dataTransfer.files[0]; }
function formatSize(b) { return b > 1048576 ? (b/1048576).toFixed(1) + 'MB' : b < 1024 ? b + 'B' : (b/1024).toFixed(0) + 'KB'; }

function onCoverChange(e) {
  coverFile.value = e.target.files[0];
  if (coverFile.value) {
    const reader = new FileReader();
    reader.onload = ev => coverPreview.value = ev.target.result;
    reader.readAsDataURL(coverFile.value);
  }
}

async function startUpload() {
  if (!canUpload.value) { showToast("Please enter a title"); return; }
  uploading.value = true;

  try {
    const token = auth.token || localStorage.getItem("token");
    if (!token) { showToast("Please login first"); uploading.value = false; return; }

    // 1. Init upload
    const totalChunks = Math.ceil(file.value.size / (3 * 1024 * 1024));
    var init;
    try {
      init = await videoApi.initUpload({
        filename: file.value.name,
        file_size: file.value.size,
        total_chunks: totalChunks,
        title: title.value,
      });
    } catch (e) {
      showToast("Init failed: " + (e?.response?.data?.message || e.message));
      uploading.value = false; return;
    }
    const videoId = init.video_id || init.data?.video_id;
    const uploadId = init.upload_id || init.data?.upload_id;
    if (!videoId || !uploadId) { showToast("Invalid init response"); uploading.value = false; return; }

    // 2. Upload chunks
    for (let i = 0; i < totalChunks; i++) {
      const start = i * 3 * 1024 * 1024;
      const end = Math.min(start + 3 * 1024 * 1024, file.value.size);
      const chunk = file.value.slice(start, end);
      const fd = new FormData();
      fd.append("file", chunk, "chunk_" + i);
      fd.append("upload_id", uploadId);
      fd.append("video_id", String(videoId));
      fd.append("chunk_index", String(i));
      try {
        await videoApi.uploadChunk(fd);
      } catch (e) {
        showToast("Chunk " + i + " failed: " + (e?.response?.data?.message || e.message));
        uploading.value = false; return;
      }
      progress.value = Math.round((i + 1) / totalChunks * 50);
    }

    // 3. Merge
    try {
      await videoApi.mergeChunks({ video_id: videoId, upload_id: uploadId });
    } catch (e) {
      showToast("Merge failed: " + (e?.response?.data?.message || e.message));
      uploading.value = false; return;
    }
    progress.value = 60;

    // 4. Upload cover
    if (coverFile.value) {
      try {
        const coverFd = new FormData();
        coverFd.append("file", coverFile.value);
        coverFd.append("video_id", String(videoId));
        await videoApi.uploadCover(coverFd);
      } catch (e) {
        showToast("Cover upload failed: " + (e?.response?.data?.message || e.message));
        // 不阻断，继续
      }
    }
    progress.value = 80;

    // 5. Update video info
    try {
      await request.put("/video/update", {
        video_id: videoId,
        title: title.value,
        description: description.value,
        category: selectedCategory.value,
      }, { params: { video_id: videoId } });
    } catch (e) {
      showToast("Update failed: " + (e?.response?.data?.message || e.message));
    }

    progress.value = 100;
    uploading.value = false;
    showToast("Video published!");
    router.push("/");
  } catch (e) {
    uploading.value = false;
    showToast("Error: " + (e?.response?.data?.message || e.message));
  }
}
</script>

<style scoped>
.upload-card { transition: none; }
.upload-zone {
  padding: 50px 20px;
  text-align: center;
  cursor: pointer;
  transition: background var(--duration);
}
.upload-zone:hover { background: rgba(0,240,255,0.02); }
.upload-icon { font-size: 42px; margin-bottom: 12px; }
.upload-text { font-size: 15px; font-weight: 600; color: var(--text-primary); }
.upload-hint { font-size: 11px; color: var(--text-muted); margin-top: 6px; }

.section-title { font-family: var(--font-display); font-size: 11px; letter-spacing: 2px; }
.field-label { display: block; font-size: 12px; font-weight: 600; color: var(--text-secondary); margin-bottom: 6px; }
.required { color: var(--pink); }

.textarea {
  min-height: 80px;
  resize: vertical;
  font-family: var(--font-body);
}

.category-grid { display: flex; flex-wrap: wrap; gap: 8px; }
.cat-option {
  padding: 6px 14px;
  border: 1px solid var(--border);
  border-radius: 20px;
  background: transparent;
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
  transition: all var(--duration);
}
.cat-option.active {
  background: rgba(0,240,255,0.1);
  border-color: var(--cyan-dim);
  color: var(--cyan);
}
.cat-option:hover:not(.active) { border-color: var(--border-glow); }

.cover-upload {
  width: 100%;
  aspect-ratio: 16/9;
  border: 2px dashed var(--border);
  border-radius: var(--radius);
  overflow: hidden;
  cursor: pointer;
  transition: border-color var(--duration);
  background: var(--bg-secondary);
}
.cover-upload:hover { border-color: var(--cyan-dim); }
.cover-preview { width: 100%; height: 100%; object-fit: cover; }
.cover-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
}
.cover-hint { font-size: 12px; color: var(--text-muted); }

.progress-bar {
  position: relative;
  height: 6px;
  background: var(--border);
  border-radius: 3px;
  overflow: hidden;
}
.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--cyan), var(--purple));
  border-radius: 3px;
  transition: width 0.3s var(--ease-out);
}
.progress-text {
  position: absolute;
  top: -20px;
  right: 0;
  font-size: 11px;
  color: var(--cyan);
  font-weight: 600;
}
</style>
