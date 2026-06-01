<template>
  <div class="page-container">
    <div class="page-content" style="padding:14px">
      <div v-if="!auth.isLoggedIn" class="card" style="text-align:center;padding:60px 20px">
        <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="#5a5a7a" stroke-width="1.2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
        <p style="margin:16px 0;color:var(--text-muted)">登录后查看个人中心</p>
        <button class="btn-primary" @click="$router.push('/login')">登录 / 注册</button>
      </div>

      <template v-else>
        <div class="card" style="display:flex;align-items:center;gap:14px;padding:20px;margin-bottom:20px">
          <div class="user-avatar">{{ auth.user?.username?.[0]?.toUpperCase() || "U" }}</div>
          <div style="flex:1">
            <div style="font-size:17px;font-weight:700">{{ auth.user?.username }}</div>
            <div style="font-size:13px;color:var(--text-muted);margin-top:4px">{{ auth.user?.signature || "这个人很懒，什么都没写" }}</div>
          </div>
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#5a5a7a" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
        </div>

        <div class="section-title">我的视频 ({{ myVideos.length }})</div>
        <div v-if="myVideos.length" class="video-list">
          <div v-for="v in myVideos" :key="v.id" class="card card-clickable" style="display:flex;gap:12px;overflow:hidden;margin-bottom:12px" @click="$router.push(`/video/${v.id}`)">
            <div style="position:relative;width:130px;height:76px;flex-shrink:0">
              <img :src="v.cover_url" style="width:100%;height:100%;object-fit:cover" />
              <span style="position:absolute;bottom:4px;right:4px;background:rgba(0,0,0,0.7);color:#fff;font-size:10px;padding:1px 6px;border-radius:3px">{{ formatDuration(v.duration) }}</span>
            </div>
            <div style="padding:10px 10px 10px 0;flex:1;display:flex;flex-direction:column;justify-content:center;gap:6px">
              <div style="font-size:14px;font-weight:600;display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden">{{ v.title }}</div>
              <div style="font-size:11px;color:var(--text-muted)">{{ formatCount(v.play_count) }} 播放 · {{ formatTime(v.created_at) }}</div>
            </div>
          </div>
        </div>
        <van-empty v-else description="还没有上传视频" />

        <button class="btn-outline" style="display:block;margin:24px auto" @click="handleLogout">退出登录</button>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { showToast } from "vant";
import { useAuthStore } from "../stores/auth";
import { useVideoStore } from "../stores/video";
import { formatDuration, formatCount, formatTime } from "../composables/utils";

const router = useRouter();
const auth = useAuthStore();
const videoStore = useVideoStore();
const myVideos = ref([]);

onMounted(async () => {
  if (auth.isLoggedIn) { await videoStore.fetchMyVideos(); myVideos.value = videoStore.myVideos; }
});
function handleLogout() { auth.logout(); showToast("已退出"); router.push("/"); }
</script>

<style scoped>
.user-avatar {
  width: 52px; height: 52px; border-radius: 50%;
  background: linear-gradient(135deg, var(--accent), #7c3aed);
  display: flex; align-items: center; justify-content: center;
  font-size: 22px; font-weight: 800; color: #fff;
}
.section-title { font-size: 15px; font-weight: 700; margin-bottom: 12px; }
</style>
