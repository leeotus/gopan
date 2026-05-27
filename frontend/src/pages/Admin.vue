<template>
  <div class="page-container">
    <div class="page-content" style="padding:14px">
      <!-- 未登录 -->
      <div v-if="!isAdmin" class="card" style="padding:40px;text-align:center">
        <div style="font-size:32px;margin-bottom:16px">🔐</div>
        <van-form @submit="handleAdminLogin">
          <van-field v-model="form.username" placeholder="管理员账号" />
          <van-field v-model="form.password" type="password" placeholder="管理员密码" />
          <button class="btn-primary" type="submit" style="width:100%;margin-top:16px">管理员登录</button>
        </van-form>
      </div>

      <!-- 已登录管理面板 -->
      <div v-else>
        <div class="card" style="padding:16px;margin-bottom:14px">
          <div style="font-size:20px;font-weight:700;margin-bottom:8px">📺 视频管理</div>
          <van-tabs v-model:active="filterStatus" @change="fetchVideos">
            <van-tab title="全部" />
            <van-tab title="待审核" />
            <van-tab title="正常" />
            <van-tab title="已下架" />
          </van-tabs>
        </div>

        <div v-for="v in videos" :key="v.id" class="card" style="display:flex;gap:12px;padding:12px;margin-bottom:8px;overflow:hidden">
          <img :src="v.cover_url" style="width:100px;height:60px;object-fit:cover;border-radius:6px;flex-shrink:0" />
          <div style="flex:1;min-width:0">
            <div style="font-size:13px;font-weight:600;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">{{ v.title }}</div>
            <div style="font-size:11px;color:var(--text-muted);margin-top:2px">{{ v.username }} · {{ formatCount(v.play_count) }} 播放</div>
            <div style="margin-top:6px;display:flex;gap:6px">
              <span class="tag" :class="statusClass(v.status)">{{ statusText(v.status) }}</span>
              <button v-if="v.status === 3" class="btn-small" @click="approve(v.id)">✓ 通过</button>
              <button v-if="v.status === 2 || v.status === 3" class="btn-small btn-danger" @click="reject(v.id)">✕ 下架</button>
            </div>
          </div>
        </div>

        <div v-if="videos.length === 0" style="text-align:center;padding:40px;color:var(--text-muted)">暂无视频</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { showToast } from "vant";
import axios from "axios";

const isAdmin = ref(false);
const adminToken = ref("");
const form = ref({ username: "", password: "" });
const videos = ref([]);
const filterStatus = ref(-1);

async function handleAdminLogin() {
  try {
    const res = await axios.post("/api/admin/login", {
      username: form.value.username,
      password: form.value.password,
    });
    adminToken.value = res.data.token;
    isAdmin.value = true;
    await fetchVideos();
  } catch {
    showToast("登录失败，请确认是管理员账号");
  }
}

async function fetchVideos() {
  try {
    const status = filterStatus.value === 0 ? -1 : filterStatus.value;
    const res = await axios.get("/api/admin/videos", {
      params: { status, limit: 50 },
      headers: { Authorization: "Bearer " + adminToken.value },
    });
    videos.value = res.data?.videos || [];
  } catch {}
}

async function approve(videoId) {
  await axios.put(`/api/admin/video/${videoId}/approve`, {}, { headers: { Authorization: "Bearer " + adminToken.value } });
  showToast("已通过");
  fetchVideos();
}

async function reject(videoId) {
  await axios.put(`/api/admin/video/${videoId}/reject`, {}, { headers: { Authorization: "Bearer " + adminToken.value } });
  showToast("已下架");
  fetchVideos();
}

function statusText(s) { return { 0: "上传中", 1: "转码中", 2: "正常", 3: "待审核", 4: "已下架" }[s] || "未知"; }
function statusClass(s) { return { 0: "tag-yellow", 1: "tag-yellow", 2: "tag-green", 3: "tag-orange", 4: "tag-red" }[s] || ""; }
function formatCount(n) { return n >= 10000 ? (n/10000).toFixed(1)+"w" : String(n); }
</script>

<style scoped>
.btn-primary { padding: 12px; border: none; border-radius: var(--radius); background: linear-gradient(135deg, var(--accent), #7c3aed); color: #fff; font-size: 15px; font-weight: 600; cursor: pointer; }
.btn-small { padding: 2px 10px; border: none; border-radius: 4px; background: var(--accent); color: #fff; font-size: 11px; cursor: pointer; }
.btn-danger { background: var(--danger); }
.tag { padding: 2px 8px; border-radius: 4px; font-size: 10px; }
.tag-green { background: rgba(34,197,94,0.2); color: #22c55e; }
.tag-yellow { background: rgba(234,179,8,0.2); color: #eab308; }
.tag-orange { background: rgba(249,115,22,0.2); color: #f97316; }
.tag-red { background: rgba(239,68,68,0.2); color: #ef4444; }
</style>
