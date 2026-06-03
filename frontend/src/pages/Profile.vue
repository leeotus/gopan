<template>
  <div class="page-container">
    <div class="page-content" style="padding: 16px">
      <!-- Not logged in -->
      <div v-if="!auth.isLoggedIn" class="card" style="text-align:center;padding:60px 20px">
        <div class="empty-icon">🔐</div>
        <p class="empty-text">Sign in to view your profile</p>
        <router-link to="/login"><button class="btn-primary btn-primary--solid" style="margin-top:20px">Sign In</button></router-link>
      </div>

      <!-- Profile -->
      <template v-else>
        <div class="profile-header card">
          <div class="avatar-large">
            <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="var(--cyan)" stroke-width="1.5"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
          </div>
          <div class="profile-info">
            <h2 class="profile-name">{{ auth.user?.username || 'User' }}</h2>
            <p class="profile-email">{{ auth.user?.email || '' }}</p>
          </div>
        </div>

        <div class="section-title text-muted" style="margin:24px 0 12px">MY VIDEOS</div>
        <div class="my-videos">
          <div v-for="v in myVideos" :key="v.id" class="video-row card card-clickable" @click="$router.push(`/video/${v.id}`)">
              <div class="row-cover">
                <img :src="v.cover_url || `/covers/${v.id}.jpg`" @error="(e) => { e.target.style.display = 'none'; e.target.nextElementSibling.style.display = 'flex'; }" />
                <div class="row-placeholder" style="display:none">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="rgba(0,240,255,0.2)"><polygon points="5,3 19,12 5,21"/></svg>
                </div>
              </div>
            <div class="row-body">
              <div class="row-title">{{ v.title }}</div>
              <div class="row-meta">{{ formatCount(v.play_count) }} plays · {{ formatTime(v.created_at) }}</div>
            </div>
          </div>
        </div>
        <div v-if="myVideos.length === 0" class="text-muted" style="text-align:center;padding:40px">
          No videos uploaded yet
        </div>

        <button class="btn-danger" @click="handleLogout" style="width:100%;margin-top:30px;padding:12px">Sign Out</button>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useAuthStore } from "../stores/auth";
import { useVideoStore } from "../stores/video";
import { formatCount, formatTime } from "../composables/utils";

const auth = useAuthStore();
const videoStore = useVideoStore();
const myVideos = ref([]);

onMounted(async () => {
  if (auth.isLoggedIn) {
    await videoStore.fetchMyVideos({ user_id: auth.user?.id || 1 });
    myVideos.value = videoStore.myVideos;
  }
});

function handleLogout() {
  localStorage.removeItem("token");
  auth.token = "";
  auth.user = null;
  myVideos.value = [];
}
</script>

<style scoped>
.profile-header {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
}
.avatar-large {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: var(--bg-elevated);
  border: 2px solid var(--cyan-dim);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.profile-name { font-size: 20px; font-weight: 700; }
.profile-email { font-size: 12px; color: var(--text-muted); margin-top: 4px; }
.section-title { font-family: var(--font-display); font-size: 11px; letter-spacing: 2px; }

.my-videos { display: flex; flex-direction: column; gap: 8px; }
.video-row { display: flex; gap: 12px; padding: 10px; align-items: center; }
.row-cover {
  width: 100px;
  height: 60px;
  border-radius: var(--radius-sm);
  overflow: hidden;
  flex-shrink: 0;
  background: var(--bg-secondary);
}
.row-cover img { width: 100%; height: 100%; object-fit: cover; }
.row-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, rgba(0,240,255,0.05), rgba(179,71,234,0.05));
}
.row-body { flex: 1; min-width: 0; }
.row-title { font-size: 14px; font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.row-meta { font-size: 11px; color: var(--text-muted); margin-top: 4px; }
</style>
