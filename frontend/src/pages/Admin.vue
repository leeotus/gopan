<template>
  <div class="page-container">
    <div class="page-content" style="padding:16px">
      <!-- Not logged in -->
      <div v-if="!isAdmin" class="card" style="padding:40px 28px;text-align:center">
        <div class="admin-icon">⬡</div>
        <h2 class="admin-title">Admin Panel</h2>
        <form @submit.prevent="handleAdminLogin" style="margin-top:24px">
          <input class="input-field" v-model="form.username" placeholder="Admin username" style="margin-bottom:12px" />
          <input class="input-field" v-model="form.password" type="password" placeholder="Admin password" style="margin-bottom:20px" />
          <button class="btn-primary btn-primary--solid" type="submit" style="width:100%">Sign In</button>
        </form>
      </div>

      <!-- Admin panel -->
      <template v-else>
        <div class="section-title text-muted" style="margin-bottom:12px">VIDEO MANAGEMENT</div>
        <div class="filter-bar">
          <button v-for="(tab, i) in tabs" :key="i"
            :class="['filter-chip', { active: filterStatus === i }]"
            @click="filterStatus = i; fetchVideos()"
          >{{ tab }}</button>
        </div>

        <div v-for="v in videos" :key="v.id" class="admin-card card">
          <div class="admin-row">
            <div class="admin-cover">
              <img :src="v.cover_url" v-if="v.cover_url" />
              <div class="cover-empty" v-else><div class="cover-shimmer"></div></div>
            </div>
            <div class="admin-body">
              <div class="admin-title-text">{{ v.title }}</div>
              <div class="admin-meta">{{ v.username || 'Unknown' }} · {{ formatCount(v.play_count) }} plays</div>
              <div class="admin-actions">
                <span class="tag" :class="statusTag(v.status)">{{ statusLabel(v.status) }}</span>
                <button v-if="v.status === 3" class="btn-secondary" @click="approve(v.id)">✓ Approve</button>
                <button v-if="v.status === 2 || v.status === 3" class="btn-secondary" @click="reject(v.id)">✕ Reject</button>
                <button class="btn-danger" @click="del(v.id)">Delete</button>
              </div>
            </div>
          </div>
        </div>

        <div v-if="videos.length === 0" class="empty-state">
          <p class="text-muted">No videos</p>
        </div>
      </template>
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
const filterStatus = ref(0);
const tabs = ["All", "Review", "Published", "Removed"];

async function handleAdminLogin() {
  try {
    const res = await axios.post("/api/admin/login", form.value);
    adminToken.value = res.data?.token || res.data?.data?.token;
    if (adminToken.value) { isAdmin.value = true; fetchVideos(); }
    else showToast("Login failed");
  } catch { showToast("Invalid credentials"); }
}

async function fetchVideos() {
  const statusMap = [-1, 3, 2, 4];
  const status = statusMap[filterStatus.value];
  try {
    const res = await axios.get("/api/admin/videos", {
      params: { status, limit: 50 },
      headers: { Authorization: "Bearer " + adminToken.value },
    });
    videos.value = res.data?.videos || [];
  } catch {}
}

async function approve(id) {
  await axios.post("/api/admin/approve", null, {
    params: { video_id: id, admin_id: 1 },
    headers: { Authorization: "Bearer " + adminToken.value },
  });
  showToast("Approved"); fetchVideos();
}
async function reject(id) {
  await axios.post("/api/admin/reject", null, {
    params: { video_id: id, admin_id: 1 },
    headers: { Authorization: "Bearer " + adminToken.value },
  });
  showToast("Rejected"); fetchVideos();
}
async function del(id) {
  await axios.delete("/api/admin/video", {
    params: { video_id: id, admin_id: 1 },
    headers: { Authorization: "Bearer " + adminToken.value },
  });
  showToast("Deleted"); fetchVideos();
}

function statusLabel(s) { return {0:"Uploading",1:"Transcoding",2:"Published",3:"Review",4:"Removed"}[s] || "Unknown"; }
function statusTag(s) { return {0:"tag-amber",1:"tag-amber",2:"tag-green",3:"tag-purple",4:"tag-red"}[s] || ""; }
function formatCount(n) { return n >= 10000 ? (n/10000).toFixed(1)+"w" : String(n); }
</script>

<style scoped>
.admin-icon { font-size: 40px; color: var(--cyan); filter: drop-shadow(0 0 16px var(--cyan-glow)); }
.admin-title { font-family: var(--font-display); font-size: 20px; letter-spacing: 3px; margin-top: 8px; color: var(--text-primary); }
.section-title { font-family: var(--font-display); font-size: 11px; letter-spacing: 2px; }
.filter-bar { display: flex; gap: 6px; margin-bottom: 16px; overflow-x: auto; }
.filter-chip {
  padding: 5px 16px;
  border: 1px solid var(--border);
  border-radius: 20px;
  background: transparent;
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
  transition: all var(--duration);
  flex-shrink: 0;
}
.filter-chip.active { background: rgba(0,240,255,0.08); border-color: var(--cyan-dim); color: var(--cyan); }

.admin-card { margin-bottom: 8px; }
.admin-row { display: flex; gap: 12px; padding: 12px; align-items: flex-start; }
.admin-cover {
  width: 100px; height: 60px;
  border-radius: var(--radius-sm);
  overflow: hidden;
  flex-shrink: 0;
  background: var(--bg-secondary);
}
.admin-cover img { width: 100%; height: 100%; object-fit: cover; }
.cover-empty { width: 100%; height: 100%; background: linear-gradient(135deg, rgba(0,240,255,0.04), rgba(179,71,234,0.04)); }
.admin-body { flex: 1; min-width: 0; }
.admin-title-text { font-size: 13px; font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.admin-meta { font-size: 11px; color: var(--text-muted); margin-top: 4px; }
.admin-actions { display: flex; gap: 6px; align-items: center; margin-top: 8px; flex-wrap: wrap; }
.empty-state { text-align: center; padding: 60px; }
</style>
