<template>
  <div class="page-container">
    <van-nav-bar title="上传视频" />

    <div class="page-content" v-if="authStore.isLoggedIn">
      <van-form @submit="handleUpload">
        <van-cell-group inset>
          <van-field v-model="form.title" name="title" label="标题" placeholder="请输入视频标题" :rules="[{ required: true, message: '请输入标题' }]" />
          <van-field v-model="form.description" name="description" label="简介" type="textarea" rows="2" placeholder="视频简介（选填）" />
          <van-field v-model="form.category" name="category" label="分类" placeholder="如：技术、前端" />
        </van-cell-group>

        <div class="upload-area">
          <van-uploader
            v-model="fileList"
            :max-count="1"
            accept="video/*"
            :before-read="beforeRead"
            :after-read="afterRead"
          >
            <template #default>
              <div class="upload-btn">
                <van-icon name="plus" size="24" />
                <span>选择视频文件</span>
              </div>
            </template>
          </van-uploader>
          <p class="upload-hint">支持 mp4 / mov / avi</p>
        </div>

        <div style="margin: 16px">
          <van-button round block type="primary" native-type="submit" :loading="uploading" :disabled="!selectedFile">
            开始上传
          </van-button>
        </div>
      </van-form>
    </div>

    <!-- 未登录 -->
    <div v-else class="page-content">
      <van-empty description="登录后即可上传视频">
        <van-button type="primary" to="/login">去登录</van-button>
      </van-empty>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { showToast } from "vant";
import { useAuthStore } from "../stores/auth";
import { videoApi } from "../api";

const authStore = useAuthStore();
const uploading = ref(false);
const selectedFile = ref(null);
const fileList = ref([]);

const form = ref({
  title: "",
  description: "",
  category: "",
});

function beforeRead(file) {
  if (!file.type.startsWith("video/")) {
    showToast("请选择视频文件");
    return false;
  }
  return true;
}

function afterRead(file) {
  selectedFile.value = file.file;
  fileList.value = [file];
}

async function handleUpload() {
  if (!selectedFile.value) {
    showToast("请选择视频文件");
    return;
  }
  uploading.value = true;
  try {
    // 1. 初始化上传
    const initRes = await videoApi.initUpload({
      filename: selectedFile.value.name,
      file_size: selectedFile.value.size,
      title: form.value.title,
    });

    // 2. 模拟分片上传（完整实现需前端切片）
    const formData = new FormData();
    formData.append("file", selectedFile.value);
    formData.append("video_id", initRes.data.video_id);
    formData.append("upload_id", initRes.data.upload_id);

    await videoApi.uploadChunk(formData);

    // 3. 合并分片
    await videoApi.mergeChunks({
      video_id: initRes.data.video_id,
      upload_id: initRes.data.upload_id,
    });

    showToast("上传成功！等待转码完成后即可播放");
    form.value = { title: "", description: "", category: "" };
    selectedFile.value = null;
    fileList.value = [];
  } catch (e) {
    showToast(e.message || "上传失败，请稍后再试");
  } finally {
    uploading.value = false;
  }
}
</script>

<style scoped>
.upload-area {
  margin: 20px 16px;
  text-align: center;
}

.upload-btn {
  width: 100%;
  height: 120px;
  border: 2px dashed #ddd;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--gopan-text-secondary);
  font-size: 14px;
}

.upload-hint {
  font-size: 12px;
  color: #ccc;
  margin-top: 8px;
}
</style>
