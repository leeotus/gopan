<template>
  <div class="page-container">
    <div class="player-area">
      <video ref="videoEl" controls autoplay style="width:100%;height:100%;object-fit:contain;background:#000" />
      <canvas ref="danmakuCanvas" class="danmaku-canvas" />
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
import { ref, onMounted, onUnmounted, nextTick } from "vue";
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
const danmakuCanvas = ref(null);
const video = ref(null);
const comments = ref([]);
const commentText = ref("");
const commentsEl = ref(null);
const fieldStyle = { '--van-field-background': '#1e1e32', '--van-field-input-text-color': '#e8e6f0', '--van-field-placeholder-text-color': '#5a5a7a' };

// ─────────── 弹幕系统 ───────────

// 弹幕池: Map<秒区间key, 弹幕数组>
const danmakuPool = new Map();
// 已加载的区间集合，避免重复请求
const loadedSegments = new Set();
// 当前预加载窗口大小（秒）
const SEGMENT_SIZE = 10;
// WebSocket 实例
let ws = null;
// Canvas 渲染上下文
let canvasCtx = null;
// 动画帧 ID
let renderRaf = null;
// 正在飞行的弹幕
const flyingDanmakus = [];
// 预加载锁
let isLoadingDanmaku = false;

function getSegmentKey(time) {
  return Math.floor(time / SEGMENT_SIZE);
}

// 从后端加载一段弹幕
async function loadDanmakus(videoId, time) {
  const key = getSegmentKey(time);
  if (loadedSegments.has(key) || isLoadingDanmaku) return;
  isLoadingDanmaku = true;
  try {
    const res = await axios.get("/api/video/danmakus", {
      params: { video_id: videoId, time },
      headers: auth.token ? { Authorization: "Bearer " + auth.token } : {},
    });
    const list = res.data?.danmakus || res.data?.data?.danmakus || [];
    list.forEach((d) => {
      const seg = getSegmentKey(d.time);
      if (!danmakuPool.has(seg)) danmakuPool.set(seg, []);
      // 去重
      const exists = danmakuPool.get(seg).some((e) => e.content === d.content && Math.abs(e.time - d.time) < 0.1);
      if (!exists) danmakuPool.get(seg).push(d);
    });
    loadedSegments.add(key);
  } catch {}
  isLoadingDanmaku = false;
}

// 从弹幕池取当前时间附近的弹幕
function getNearbyDanmakus(currentTime) {
  const seg = getSegmentKey(currentTime);
  const result = [];
  for (const s of [seg - 1, seg, seg + 1]) {
    const list = danmakuPool.get(s);
    if (list) {
      for (const d of list) {
        if (d.time >= currentTime - 2 && d.time <= currentTime + 10) {
          result.push(d);
        }
      }
    }
  }
  return result;
}

// 清理过期弹幕池
function cleanOldSegments(currentTime) {
  const seg = getSegmentKey(currentTime);
  for (const key of danmakuPool.keys()) {
    if (key < seg - 2) {
      danmakuPool.delete(key);
      loadedSegments.delete(key);
    }
  }
}

// Canvas 弹幕渲染
function initDanmakuCanvas() {
  if (!danmakuCanvas.value) return;
  const parent = danmakuCanvas.value.parentElement;
  danmakuCanvas.value.width = parent.clientWidth;
  danmakuCanvas.value.height = parent.clientHeight;
  canvasCtx = danmakuCanvas.value.getContext("2d");
}

// 向飞行列表添加弹幕
function emitDanmaku(d) {
  if (!canvasCtx) return;
  const canvas = danmakuCanvas.value;
  const y = 20 + Math.random() * (canvas.height - 80);
  const speed = 1.5 + Math.random() * 1;
  const fontSize = 18 + Math.random() * 4;
  flyingDanmakus.push({
    x: canvas.width,
    y,
    speed,
    text: d.content,
    color: d.color || "#ffffff",
    fontSize,
    opacity: 0.9,
    time: d.time,
  });
}

// 渲染飞行中的弹幕
function renderDanmakuFrame() {
  if (!canvasCtx || !danmakuCanvas.value) return;
  const ctx = canvasCtx;
  const canvas = danmakuCanvas.value;
  ctx.clearRect(0, 0, canvas.width, canvas.height);

  for (let i = flyingDanmakus.length - 1; i >= 0; i--) {
    const d = flyingDanmakus[i];
    d.x -= d.speed;
    ctx.font = `bold ${d.fontSize}px "PingFang SC","Microsoft YaHei",sans-serif`;
    ctx.fillStyle = d.color;
    ctx.globalAlpha = d.opacity;
    // 描边增强可读性
    ctx.strokeStyle = "rgba(0,0,0,0.5)";
    ctx.lineWidth = 3;
    ctx.strokeText(d.text, d.x, d.y);
    ctx.fillText(d.text, d.x, d.y);
    ctx.globalAlpha = 1;

    if (d.x < -ctx.measureText(d.text).width - 50) {
      flyingDanmakus.splice(i, 1);
    }
  }
  renderRaf = requestAnimationFrame(renderDanmakuFrame);
}

// 监听播放进度，从弹幕池取弹幕并发射
let lastSegmentKey = -1;
function onVideoTimeUpdate() {
  if (!videoEl.value) return;
  const t = videoEl.value.currentTime;
  const key = getSegmentKey(t);

  // 预加载下一个区间
  if (key !== lastSegmentKey) {
    loadDanmakus(video.value?.id, (key + 1) * SEGMENT_SIZE);
    cleanOldSegments(t);
    lastSegmentKey = key;
  }

  // 从弹幕池取当前时间附近的弹幕
  const nearby = getNearbyDanmakus(t);
  for (const d of nearby) {
    // 避免重复发射（检查是否已在飞行中）
    const flying = flyingDanmakus.find((f) => f.text === d.content && Math.abs(f.time - d.time) < 0.5);
    if (!flying) {
      emitDanmaku(d);
    }
  }
}

// 用户 Seek 时清空当前飞行弹幕并重新加载
function onVideoSeeked() {
  flyingDanmakus.length = 0;
  if (!videoEl.value || !video.value) return;
  const t = videoEl.value.currentTime;
  const seg = getSegmentKey(t);
  // 清空当前段标记，强制重新加载
  loadedSegments.delete(seg);
  loadedSegments.delete(seg + 1);
  lastSegmentKey = -1;
  loadDanmakus(video.value.id, t);
  loadDanmakus(video.value.id, (seg + 1) * SEGMENT_SIZE);
}

// WebSocket 实时弹幕
function connectDanmakuWS(videoId) {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const host = window.location.host;
  const tokenParam = auth.token ? `&token=${encodeURIComponent(auth.token)}` : "";
  const wsUrl = `${protocol}//${host}/ws/danmaku?video_id=${videoId}${tokenParam}`;

  ws = new WebSocket(wsUrl);
  ws.onmessage = (event) => {
    try {
      const d = JSON.parse(event.data);
      const seg = getSegmentKey(d.time);
      // 加入弹幕池
      if (!danmakuPool.has(seg)) danmakuPool.set(seg, []);
      const exists = danmakuPool.get(seg).some((e) => e.content === d.content && Math.abs(e.time - d.time) < 0.1);
      if (!exists) danmakuPool.get(seg).push(d);

      // 如果弹幕时间在可视窗口内，立即发射
      if (videoEl.value) {
        const t = videoEl.value.currentTime;
        if (d.time >= t - 2 && d.time <= t + 10) {
          emitDanmaku(d);
        }
      }
    } catch {}
  };
  ws.onerror = () => {};
  ws.onclose = () => {
    // 断开后不再重连，如需自动重连可在此处添加
  };
}

// ─────────── 生命周期 ───────────

onMounted(async () => {
  await videoStore.fetchDetail(Number(route.params.id));
  video.value = videoStore.currentVideo;

  // 播放器初始化
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

  // 播放进度恢复
  if (auth.isLoggedIn && video.value) {
    try {
      const res = await axios.get("/api/video/play-progress", {
        params: { video_id: video.value.id },
        headers: { Authorization: "Bearer " + auth.token },
      });
      const pos = parseFloat(res.data?.message || "0");
      if (pos > 0 && videoEl.value) {
        videoEl.value.currentTime = pos;
        showToast(`从 ${pos.toFixed(0)} 秒处继续播放`);
      }
    } catch {}
  }

  // 弹幕系统初始化
  await nextTick();
  initDanmakuCanvas();
  renderDanmakuFrame();

  // 监听播放事件
  if (videoEl.value) {
    videoEl.value.addEventListener("timeupdate", onVideoTimeUpdate);
    videoEl.value.addEventListener("seeked", onVideoSeeked);
  }

  // 加载首段弹幕 + 预加载下一段 + 连接 WebSocket
  if (video.value) {
    loadDanmakus(video.value.id, 0);
    loadDanmakus(video.value.id, SEGMENT_SIZE);
    connectDanmakuWS(video.value.id);
  }

  // 评论
  comments.value = [
    { id: 1, username: "张三", content: "讲得太好了！收获很多", created_at: Math.floor(Date.now()/1000)-3600 },
  ];
  fetchComments();
});

onUnmounted(() => {
  if (ws) ws.close();
  if (renderRaf) cancelAnimationFrame(renderRaf);
  if (videoEl.value) {
    videoEl.value.removeEventListener("timeupdate", onVideoTimeUpdate);
    videoEl.value.removeEventListener("seeked", onVideoSeeked);
  }
  danmakuPool.clear();
  loadedSegments.clear();
  flyingDanmakus.length = 0;
});

// ─────────── 互动操作 ───────────

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
.danmaku-canvas { position: absolute; top: 0; left: 0; width: 100%; height: 100%; pointer-events: none; z-index: 1; }
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
