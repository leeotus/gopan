<template>
  <div class="page-container">
    <div class="player-area">
      <video ref="videoEl" controls autoplay style="width:100%;height:100%;object-fit:contain;background:#000" />
    </div>

    <div class="info-section card" style="margin:12px 14px;border-radius:var(--radius)" v-if="video">
      <h2 class="video-title">{{ video.title }}</h2>
      <div class="info-row">
        <span>{{ formatCount(video.play_count) }} 播放</span>
        <span class="dot">·</span>
        <span>{{ formatTime(video.created_at) }}</span>
        <span v-if="video.category" class="tag">{{ video.category }}</span>
      </div>
      <p class="video-desc" v-if="video.description">{{ video.description }}</p>
    </div>

    <div class="action-bar card" style="margin:0 14px 12px;border-radius:var(--radius);display:flex;justify-content:space-around;padding:14px 0" v-if="video">
      <div class="action-item" @click="handleLike">
        <svg width="22" height="22" viewBox="0 0 24 24" :fill="video.liked ? '#ef4444' : 'none'" :stroke="video.liked ? '#ef4444' : '#8b8baa'" stroke-width="1.8"><path d="M14 9V5a3 3 0 0 0-3-3l-4 9v11h11.28a2 2 0 0 0 2-1.7l1.38-9a2 2 0 0 0-2-2.3zM7 22H4a2 2 0 0 1-2-2v-7a2 2 0 0 1 2-2h3"/></svg>
        <span :style="{color: video.liked ? '#ef4444' : '#8b8baa'}">{{ formatCount(video.like_count) }}</span>
      </div>
      <div class="action-item" @click="handleFavorite">
        <svg width="22" height="22" viewBox="0 0 24 24" :fill="video.favorited ? '#f59e0b' : 'none'" :stroke="video.favorited ? '#f59e0b' : '#8b8baa'" stroke-width="1.8"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
        <span :style="{color: video.favorited ? '#f59e0b' : '#8b8baa'}">{{ video.favorited ? '已收藏' : '收藏' }}</span>
      </div>
      <div class="action-item" @click="scrollToComments">
        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#8b8baa" stroke-width="1.8"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
        <span>评论</span>
      </div>
    </div>

    <div class="comment-section card" style="margin:0 14px 80px;border-radius:var(--radius);padding:16px" ref="commentsEl">
      <div class="section-title">评论 ({{ comments.length }})</div>
      <div class="comment-input" v-if="auth.isLoggedIn">
        <van-field v-model="commentText" rows="1" type="textarea" placeholder="说点什么..." autosize :style="fieldStyle">
          <template #button><button class="btn-primary" style="padding:6px 16px;font-size:12px" @click="postComment">发送</button></template>
        </van-field>
      </div>
      <div v-else class="login-hint"><router-link to="/login">登录后评论</router-link></div>
      <div v-for="c in comments" :key="c.id" class="comment-item">
        <div class="comment-header">
          <span class="comment-user">{{ c.username }}</span>
          <span class="comment-time">{{ formatTime(c.created_at) }}</span>
        </div>
        <div class="comment-content">{{ c.content }}</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useRoute } from "vue-router";
import { showToast } from "vant";
import axios from "axios";
import Hls from "hls.js";
import { useVideoStore } from "../stores/video";
import { useAuthStore } from "../stores/auth";
import { formatCount, formatTime } from "../composables/utils";

const route = useRoute();
const videoStore = useVideoStore();
const auth = useAuthStore();
const videoEl = ref(null);
const video = ref(null);
const comments = ref([]);
const commentText = ref("");
const commentsEl = ref(null);
const fieldStyle = { '--van-field-background': '#1e1e32', '--van-field-input-text-color': '#e8e6f0', '--van-field-placeholder-text-color': '#5a5a7a' };

onMounted(async () => {
  await videoStore.fetchDetail(Number(route.params.id));
  video.value = videoStore.currentVideo;

  // hls.js 播放器初始化
  if (video.value && videoEl.value) {
    const transcodes = video.value.transcodes || [];
    if (transcodes.length > 0) {
      const url = transcodes[0].m3u8_url || transcodes[0].M3U8Url;
      if (url) {
        if (Hls.isSupported()) {
          const hls = new Hls();
          hls.loadSource(url);
          hls.attachMedia(videoEl.value);
        } else if (videoEl.value.canPlayType("application/vnd.apple.mpegurl")) {
          videoEl.value.src = url;
        }
      }
    }
  }

  // 查询播放进度
  if (auth.isLoggedIn && video.value) {
    try {
      const res = await axios.get("/api/video/play-progress", {
        params: { video_id: video.value.id, user_id: auth.user?.userId },
        headers: { Authorization: "Bearer " + auth.token },
      });
      const pos = parseFloat(res.data?.message || "0");
      if (pos > 0) {
        showToast(`从 ${pos.toFixed(0)} 秒处继续播放`);
        // TODO: player.seekTo(pos)
      }
    } catch {}
  }
  comments.value = [
    { id: 1, username: "张三", content: "讲得太好了！收获很多", created_at: Math.floor(Date.now()/1000)-3600 },
  ];
  fetchComments();

  // WebSocket 弹幕
  if (video.value) {
    const wsUrl = `ws://localhost:8888/ws/danmaku?video_id=${video.value.id}`;
    const ws = new WebSocket(wsUrl);
    ws.onmessage = (event) => {
      try {
        const d = JSON.parse(event.data);
        // 渲染弹幕到屏幕（简化：加到评论区上方）
      } catch {}
    };
  }
});

function handlePlay() {
  showToast("播放器待集成 hls.js");
  // 模拟播放：每 5s 上报进度（实际集成后从播放器取 currentTime）
  if (auth.isLoggedIn && video.value) {
    let pos = 0;
    setInterval(async () => {
      pos += 5;
      try {
        await axios.post("/api/video/play-progress", {
          video_id: video.value.id,
          user_id: auth.user?.userId,
          position: pos,
        }, { headers: { Authorization: "Bearer " + auth.token } });
      } catch {}
    }, 5000);
  }
}
async function handleLike() {
  if (!auth.isLoggedIn) { showToast("请先登录"); return; }
  try {
    if (video.value.liked) {
      await axios.delete("/api/video/like", { params: { video_id: video.value.id }, headers: { Authorization: "Bearer " + auth.token } });
    } else {
      await axios.post("/api/video/like", null, { params: { video_id: video.value.id }, headers: { Authorization: "Bearer " + auth.token } });
    }
    video.value.liked = !video.value.liked;
    video.value.like_count += video.value.liked ? 1 : -1;
  } catch {}
}

async function handleFavorite() {
  if (!auth.isLoggedIn) { showToast("请先登录"); return; }
  try {
    if (video.value.favorited) {
      await axios.delete("/api/video/favorite", { params: { video_id: video.value.id }, headers: { Authorization: "Bearer " + auth.token } });
    } else {
      await axios.post("/api/video/favorite", null, { params: { video_id: video.value.id }, headers: { Authorization: "Bearer " + auth.token } });
    }
    video.value.favorited = !video.value.favorited;
  } catch {}
}

function postComment() {
  if (!commentText.value.trim()) return;
  axios.post("/api/video/comment", { video_id: video.value.id, content: commentText.value }, { headers: { Authorization: "Bearer " + auth.token } }).then(() => {
    comments.value.unshift({ id: Date.now(), username: auth.user?.username || "我", content: commentText.value, created_at: Math.floor(Date.now()/1000) });
    commentText.value = "";
  }).catch(() => {});
}

async function fetchComments() {
  try {
    const res = await axios.get("/api/video/comments", { params: { video_id: video.value.id }, headers: { Authorization: "Bearer " + auth.token } });
    comments.value = (res.data?.comments || res.data?.data?.comments || []).map(c => ({
      ...c, username: c.username || "匿名",
    }));
  } catch {}
}
function scrollToComments() { commentsEl.value?.scrollIntoView({ behavior: "smooth" }); }
</script>

<style scoped>
.player-area { position: relative; width: 100%; aspect-ratio: 16/9; background: #000; overflow: hidden; }
.player-cover { width: 100%; height: 100%; object-fit: contain; }
.play-btn { position: absolute; top: 50%; left: 50%; transform: translate(-50%,-50%); cursor: pointer; transition: transform var(--transition); }
.play-btn:active { transform: translate(-50%,-50%) scale(0.9); }

.info-section { padding: 16px; }
.video-title { font-size: 17px; font-weight: 700; line-height: 1.5; margin-bottom: 10px; }
.info-row { font-size: 12px; color: var(--text-secondary); display: flex; align-items: center; gap: 8px; flex-wrap: wrap; margin-bottom: 10px; }
.dot { color: #444; }
.video-desc { font-size: 13px; color: var(--text-secondary); line-height: 1.7; }

.action-item { display: flex; flex-direction: column; align-items: center; gap: 4px; font-size: 11px; cursor: pointer; transition: transform var(--transition); }
.action-item:active { transform: scale(0.9); }

.section-title { font-size: 14px; font-weight: 700; margin-bottom: 12px; }
.comment-item { padding: 12px 0; border-bottom: 1px solid var(--border); }
.comment-header { display: flex; gap: 8px; align-items: center; margin-bottom: 6px; }
.comment-user { font-size: 13px; font-weight: 600; }
.comment-time { font-size: 11px; color: var(--text-muted); }
.comment-content { font-size: 14px; color: var(--text-secondary); line-height: 1.6; }
.login-hint { text-align: center; padding: 12px; font-size: 13px; }
.login-hint a { color: var(--accent); }
</style>
