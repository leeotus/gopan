<template>
  <div class="page-container">
    <!-- Player area -->
    <div class="player-area">
      <video ref="videoEl" controls autoplay playsinline style="width:100%;height:100%;object-fit:contain;background:#000" />
      <canvas ref="danmakuCanvas" class="danmaku-canvas" />
    </div>

    <!-- Video info -->
    <div class="info-section card" v-if="video">
      <h1 class="video-title">{{ video.title }}</h1>
      <div class="info-row">
        <span>{{ formatCount(video.play_count) }} plays</span>
        <span class="dot">·</span>
        <span>{{ formatTime(video.created_at) }}</span>
        <span v-if="video.category" class="tag tag-cyan">{{ video.category }}</span>
      </div>
      <p class="video-desc" v-if="video.description">{{ video.description }}</p>
    </div>

    <!-- Action bar -->
    <div class="action-bar card" v-if="video">
      <button class="action-btn" @click="handleLike" :class="{ active: video.liked }">
        <svg width="20" height="20" viewBox="0 0 24 24" :fill="video.liked ? 'var(--pink)' : 'none'" :stroke="video.liked ? 'var(--pink)' : 'var(--text-muted)'" stroke-width="1.8"><path d="M14 9V5a3 3 0 0 0-3-3l-4 9v11h11.28a2 2 0 0 0 2-1.7l1.38-9a2 2 0 0 0-2-2.3zM7 22H4a2 2 0 0 1-2-2v-7a2 2 0 0 1 2-2h3"/></svg>
        <span>{{ formatCount(video.like_count) }}</span>
      </button>
      <button class="action-btn" @click="handleFavorite" :class="{ active: video.favorited }">
        <svg width="20" height="20" viewBox="0 0 24 24" :fill="video.favorited ? 'var(--amber)' : 'none'" :stroke="video.favorited ? 'var(--amber)' : 'var(--text-muted)'" stroke-width="1.8"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
        <span>{{ video.favorited ? 'Saved' : 'Save' }}</span>
      </button>
      <button class="action-btn" @click="scrollToComments">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="var(--text-muted)" stroke-width="1.8"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
        <span>{{ comments.length }}</span>
      </button>
    </div>

    <!-- Danmaku input -->
    <div class="danmaku-bar" v-if="auth.isLoggedIn">
      <input v-model="danmakuText" class="danmaku-input" placeholder="Send a danmaku..." @keyup.enter="sendDanmakuNow" />
      <button class="btn-primary" style="padding:6px 16px;font-size:12px" @click="sendDanmakuNow">Send</button>
    </div>

    <!-- Comments -->
    <div class="comment-section card" ref="commentsEl">
      <div class="section-title text-muted" style="margin-bottom:12px">COMMENTS {{ comments.length }}</div>
      <div v-if="auth.isLoggedIn" class="comment-input-row">
        <input v-model="commentText" class="input-field" placeholder="Write a comment..." @keyup.enter="postComment" />
        <button class="btn-primary" style="padding:6px 16px;font-size:12px;flex-shrink:0" @click="postComment">Post</button>
      </div>
      <div v-for="c in comments" :key="c.id" class="comment-item">
        <div class="comment-header">
          <span class="comment-user">{{ c.username }}</span>
          <span class="comment-time">{{ formatTime(c.created_at) }}</span>
        </div>
        <div class="comment-body">{{ c.content }}</div>
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
const danmakuText = ref("");
const commentsEl = ref(null);

// ── Danmaku system ──
const danmakuPool = new Map();
const loadedSegments = new Set();
const SEGMENT_SIZE = 10;
let ws = null, canvasCtx = null, renderRaf = null;
const flyingDanmakus = [];
let isLoadingDanmaku = false, lastSegmentKey = -1;
let saveProgressTimer = null;

function getSegmentKey(t) { return Math.floor(t / SEGMENT_SIZE); }

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
    list.forEach(d => {
      const seg = getSegmentKey(d.time);
      if (!danmakuPool.has(seg)) danmakuPool.set(seg, []);
      if (!danmakuPool.get(seg).some(e => e.content === d.content && Math.abs(e.time - d.time) < 0.1))
        danmakuPool.get(seg).push(d);
    });
    loadedSegments.add(key);
  } catch {} finally { isLoadingDanmaku = false; }
}

function getNearbyDanmakus(t) {
  const seg = getSegmentKey(t);
  const r = [];
  for (const s of [seg-1, seg, seg+1]) {
    const list = danmakuPool.get(s);
    if (list) for (const d of list) { if (d.time >= t-2 && d.time <= t+10) r.push(d); }
  }
  return r;
}
function cleanOldSegments(t) { const seg = getSegmentKey(t); for (const k of danmakuPool.keys()) { if (k < seg-2) { danmakuPool.delete(k); loadedSegments.delete(k); }}}

function initDanmakuCanvas() {
  if (!danmakuCanvas.value) return;
  danmakuCanvas.value.width = danmakuCanvas.value.parentElement.clientWidth;
  danmakuCanvas.value.height = danmakuCanvas.value.parentElement.clientHeight;
  canvasCtx = danmakuCanvas.value.getContext("2d");
}
function emitDanmaku(d) {
  if (!canvasCtx) return;
  const canvas = danmakuCanvas.value;
  flyingDanmakus.push({ x: canvas.width, y: 20 + Math.random() * (canvas.height - 80), speed: 1.5 + Math.random(), text: d.content, color: d.color || "#fff", fontSize: 18 + Math.random() * 4, time: d.time });
}
function renderDanmakuFrame() {
  if (!canvasCtx || !danmakuCanvas.value) return;
  const ctx = canvasCtx, canvas = danmakuCanvas.value;
  ctx.clearRect(0, 0, canvas.width, canvas.height);
  for (let i = flyingDanmakus.length-1; i >= 0; i--) {
    const d = flyingDanmakus[i]; d.x -= d.speed;
    ctx.font = `bold ${d.fontSize}px "Plus Jakarta Sans", sans-serif`;
    ctx.strokeStyle = "rgba(0,0,0,0.6)"; ctx.lineWidth = 3;
    ctx.strokeText(d.text, d.x, d.y);
    ctx.fillStyle = d.color; ctx.fillText(d.text, d.x, d.y);
    if (d.x < -ctx.measureText(d.text).width - 50) flyingDanmakus.splice(i, 1);
  }
  renderRaf = requestAnimationFrame(renderDanmakuFrame);
}
function onVideoTimeUpdate() {
  if (!videoEl.value) return;
  const t = videoEl.value.currentTime, key = getSegmentKey(t);
  if (key !== lastSegmentKey) { loadDanmakus(video.value?.id, (key+1)*SEGMENT_SIZE); cleanOldSegments(t); lastSegmentKey = key; }
  for (const d of getNearbyDanmakus(t)) {
    if (!flyingDanmakus.find(f => f.text === d.content && Math.abs(f.time-d.time) < 0.5)) {
      emitDanmaku(d);
      const seg = getSegmentKey(d.time), list = danmakuPool.get(seg);
      if (list) { const idx = list.findIndex(e => e === d); if (idx !== -1) list.splice(idx, 1); }
    }
  }
}
function onVideoSeeked() {
  flyingDanmakus.length = 0; danmakuPool.clear(); loadedSegments.clear();
  if (!videoEl.value || !video.value) return;
  const t = videoEl.value.currentTime, seg = getSegmentKey(t);
  lastSegmentKey = -1; loadDanmakus(video.value.id, t); loadDanmakus(video.value.id, (seg+1)*SEGMENT_SIZE);
}

// ── Playback progress ──
async function savePlayProgress() {
  if (!videoEl.value || !auth.isLoggedIn || !video.value) return;
  const t = videoEl.value.currentTime; if (t <= 0) return;
  const token = auth.token || localStorage.getItem("token");
  navigator.sendBeacon("/api/video/play-progress", new Blob([JSON.stringify({ video_id: video.value.id, position: t, token: token })], { type: "application/json" }));
}
async function clearPlayProgress() {
  if (!video.value || !auth.isLoggedIn) return;
  const token = auth.token || localStorage.getItem("token");
  try { await axios.post("/api/video/play-progress", { video_id: video.value.id, position: 0, token: token }); } catch {}
}

// ── WebSocket ──
function connectDanmakuWS(videoId) {
  const proto = location.protocol === "https:" ? "wss:" : "ws:";
  ws = new WebSocket(`${proto}//${location.host}/ws/danmaku?video_id=${videoId}&token=${encodeURIComponent(auth.token || "")}`);
  ws.onmessage = e => {
    try {
      const d = JSON.parse(e.data);
      const seg = getSegmentKey(d.time);
      if (!danmakuPool.has(seg)) danmakuPool.set(seg, []);
      if (!danmakuPool.get(seg).some(x => x.content === d.content && Math.abs(x.time - d.time) < 0.1))
        danmakuPool.get(seg).push(d);
      if (videoEl.value) { const t = videoEl.value.currentTime; if (d.time >= t-2 && d.time <= t+10) emitDanmaku(d); }
    } catch {}
  };
}

// ── Lifecycle ──
onMounted(async () => {
  await videoStore.fetchDetail(Number(route.params.id));
  video.value = videoStore.currentVideo;
  if (video.value && videoEl.value) {
    const transcodes = video.value.transcodes || [];
    if (transcodes.length) {
      const raw = transcodes[0].m3u8_url || transcodes[0].M3U8Url;
      const url = raw?.replace(/^https?:\/\/minio:\d+\/gopan-videos/, "").replace(/^https?:\/\/\d+\.\d+\.\d+\.\d+(:\d+)?\/gopan-videos/, "") || "";
      if (url) {
        if (Hls.isSupported()) { const h = new Hls(); h.loadSource(url); h.attachMedia(videoEl.value); }
        else if (videoEl.value.canPlayType("application/vnd.apple.mpegurl")) videoEl.value.src = url;
      }
    }
  }
  if (auth.isLoggedIn && video.value) {
    try {
      const token = auth.token || localStorage.getItem("token");
      const res = await axios.get("/api/video/play-progress", { params: { video_id: video.value.id }, headers: { Authorization: "Bearer " + token } });
      const pos = parseFloat(res.data?.message || "0");
      if (pos > 0 && videoEl.value) { videoEl.value.currentTime = pos; showToast(`Resuming from ${pos.toFixed(0)}s`); }
    } catch {}
  }
  await nextTick(); initDanmakuCanvas(); renderDanmakuFrame();
  if (videoEl.value) {
    videoEl.value.addEventListener("timeupdate", onVideoTimeUpdate);
    videoEl.value.addEventListener("seeked", onVideoSeeked);
    videoEl.value.addEventListener("ended", () => { clearPlayProgress(); showToast("Playback complete"); });
    saveProgressTimer = setInterval(savePlayProgress, 10000);
  }
  if (video.value) { loadDanmakus(video.value.id, 0); loadDanmakus(video.value.id, SEGMENT_SIZE); connectDanmakuWS(video.value.id); }
  fetchComments();
});

onUnmounted(() => {
  if (saveProgressTimer) clearInterval(saveProgressTimer); savePlayProgress();
  if (ws) ws.close(); if (renderRaf) cancelAnimationFrame(renderRaf);
  if (videoEl.value) { videoEl.value.removeEventListener("timeupdate", onVideoTimeUpdate); videoEl.value.removeEventListener("seeked", onVideoSeeked); }
  danmakuPool.clear(); loadedSegments.clear(); flyingDanmakus.length = 0;
});

// ── Actions ──
async function handleLike() {
  if (!auth.isLoggedIn) { showToast("Please login"); return; }
  try {
    if (video.value.liked) await axios.delete("/api/video/like", { params: { video_id: video.value.id }, headers: { Authorization: "Bearer " + auth.token } });
    else await axios.post("/api/video/like", null, { params: { video_id: video.value.id }, headers: { Authorization: "Bearer " + auth.token } });
    video.value.liked = !video.value.liked; video.value.like_count += video.value.liked ? 1 : -1;
  } catch {}
}
async function handleFavorite() {
  if (!auth.isLoggedIn) { showToast("Please login"); return; }
  try {
    if (video.value.favorited) await axios.delete("/api/video/favorite", { params: { video_id: video.value.id }, headers: { Authorization: "Bearer " + auth.token } });
    else await axios.post("/api/video/favorite", null, { params: { video_id: video.value.id }, headers: { Authorization: "Bearer " + auth.token } });
    video.value.favorited = !video.value.favorited;
  } catch {}
}
function postComment() {
  if (!commentText.value.trim()) return;
  axios.post("/api/video/comment", { content: commentText.value }, { params: { video_id: video.value.id }, headers: { Authorization: "Bearer " + auth.token } })
    .then(() => { comments.value.unshift({ id: Date.now(), username: "Me", content: commentText.value, created_at: Math.floor(Date.now()/1000) }); commentText.value = ""; }).catch(() => {});
}
async function sendDanmakuNow() {
  const t = danmakuText.value.trim(); if (!t || !videoEl.value) return;
  try { await axios.post("/api/video/danmaku", { content: t, time: videoEl.value.currentTime, color: "#00f0ff", mode: 1 }, { params: { video_id: video.value.id }, headers: { Authorization: "Bearer " + auth.token } }); danmakuText.value = ""; } catch {}
}
async function fetchComments() {
  try {
    const res = await axios.get("/api/video/comments", { params: { video_id: video.value.id }, headers: { Authorization: "Bearer " + auth.token } });
    comments.value = (res.data?.comments || res.data?.data?.comments || []).map(c => ({ ...c, username: c.username || "Anonymous" }));
  } catch {}
}
function scrollToComments() { commentsEl.value?.scrollIntoView({ behavior: "smooth" }); }
</script>

<style scoped>
.player-area { position: relative; width: 100%; aspect-ratio: 16/9; background: #000; overflow: hidden; }
.danmaku-canvas { position: absolute; inset: 0; pointer-events: none; z-index: 1; }
.info-section { padding: 16px; margin: 12px 14px 0; }
.video-title { font-size: 18px; font-weight: 700; line-height: 1.4; margin-bottom: 10px; }
.info-row { font-size: 12px; color: var(--text-muted); display: flex; align-items: center; gap: 8px; flex-wrap: wrap; margin-bottom: 8px; }
.dot { color: var(--border); }
.video-desc { font-size: 13px; color: var(--text-secondary); line-height: 1.6; }
.action-bar { margin: 10px 14px; display: flex; justify-content: space-around; padding: 12px 0; }
.action-btn { display: flex; flex-direction: column; align-items: center; gap: 4px; font-size: 11px; background: none; border: none; color: var(--text-muted); cursor: pointer; transition: color var(--duration); }
.action-btn.active { color: var(--pink); }
.danmaku-bar { display: flex; gap: 8px; align-items: center; margin: 0 14px 12px; }
.danmaku-input { flex: 1; padding: 8px 14px; background: var(--bg-input); border: 1px solid var(--border); border-radius: 20px; color: var(--text-primary); font-size: 12px; outline: none; }
.danmaku-input:focus { border-color: var(--cyan-dim); }
.danmaku-input::placeholder { color: var(--text-muted); }
.comment-section { margin: 0 14px 80px; padding: 16px; }
.section-title { font-family: var(--font-display); font-size: 11px; letter-spacing: 2px; }
.comment-input-row { display: flex; gap: 8px; margin-bottom: 16px; }
.comment-item { padding: 12px 0; border-bottom: 1px solid var(--border); }
.comment-header { display: flex; gap: 8px; align-items: center; margin-bottom: 4px; }
.comment-user { font-size: 13px; font-weight: 600; }
.comment-time { font-size: 10px; color: var(--text-muted); }
.comment-body { font-size: 14px; color: var(--text-secondary); line-height: 1.5; }
</style>
