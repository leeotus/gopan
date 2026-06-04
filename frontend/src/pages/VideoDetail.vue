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

      <!-- 1. 原创视频简介（永远处于顶端保全层） -->
      <p class="video-desc" :class="{ 'desc-empty': !video.description }">
        {{ video.description || "🚀 该视频目前暂无作者简介。你可以试用顶端的 GoPan AI 智能语义检索，输入任何关于视频画面的长句（如“一只小狗在沙滩上冲浪”），AI 即可帮你直接跨越障碍搜索定位到该视频！" }}
      </p>

      <!-- 2. AI 智能听译总结区域（始终保持常驻显示，内嵌平滑状态切换，保障零布局偏置抖动） -->
      <div class="ai-summary-box">
        <div class="ai-summary-header">
          <span class="ai-badge-neon">🤖 AI 语音听译大纲摘要</span>
          <!-- 动感闪烁信标：在听译计算中处于红蓝交替闪烁，计算完化身常规信号灯 -->
          <div class="pulse-spark" :class="{ 'spark-active': aiSummaryStatus === 1 || aiSummaryStatus === 0 }"></div>
        </div>

        <!-- 子状态 A：AI 仍在分析音频中（status=1 生成中 或 0 等待入队） -->
        <div v-if="aiSummaryStatus === 1 || aiSummaryStatus === 0" class="ai-loading-container">
          <div class="skeleton-bar bar-long"></div>
          <div class="skeleton-bar bar-medium"></div>
          <div class="skeleton-bar bar-short"></div>
          <p class="loading-subtext">
            🤖 AI 正在后台听译音轨并生成摘要，请稍候...（一般需要几十秒到几分钟，完成后会自动刷新）
          </p>
        </div>

        <!-- 子状态 B：摘要已就绪（status=2），流式打字机输出 -->
        <p v-else-if="aiSummaryStatus === 2" class="ai-summary-text">
          {{ typewriterText }}
        </p>

        <!-- 子状态 C：摘要生成失败（status=3） -->
        <div v-else-if="aiSummaryStatus === 3" class="ai-failed-container">
          <p class="ai-failed-text">❌ AI 摘要生成失败。可能是 Whisper 服务暂时不可用，请稍后刷新页面重试。</p>
        </div>
      </div>
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

// AI 智能听译总结相关状态
// aiSummaryStatus: 0=未生成 1=生成中 2=已完成 3=失败
const aiSummaryStatus = ref(0);
const aiSummary = ref("");
const typewriterText = ref("");
let typewriterTimer = null;
let aiPollTimer = null;
const AI_POLL_INTERVAL = 5000; // 5 秒轮询一次
const AI_POLL_MAX = 120;       // 最多轮询 10 分钟（120 * 5s），避免无限请求
let aiPollCount = 0;

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
  
  if (video.value) {
    initAISummary(video.value);
  }

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
  if (typewriterTimer) clearInterval(typewriterTimer);
  if (aiPollTimer) clearInterval(aiPollTimer);
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

// AI 智能转译大纲：
// 不再"一进页面就同步等待 Whisper 跑完"，而是按视频自带的 ai_summary_status 字段进行分支处理。
// status=2：直接展示已有摘要，跑打字机。
// status=0/1：等待中，启动 5s 轮询，直到完成或失败为止。
// status=3：展示失败提示。
function initAISummary(v) {
  // 同步当前状态到响应式变量
  aiSummaryStatus.value = Number(v.ai_summary_status ?? 0);
  aiSummary.value = v.ai_summary || "";

  if (aiSummaryStatus.value === 2 && aiSummary.value) {
    startTypewriter(aiSummary.value);
    return;
  }
  if (aiSummaryStatus.value === 0 || aiSummaryStatus.value === 1) {
    startAISummaryPoll(v.id);
  }
  // status === 3 直接由模板渲染失败 UI，无需轮询
}

function startAISummaryPoll(videoId) {
  if (aiPollTimer) clearInterval(aiPollTimer);
  aiPollCount = 0;
  aiPollTimer = setInterval(async () => {
    aiPollCount++;
    if (aiPollCount > AI_POLL_MAX) {
      clearInterval(aiPollTimer);
      aiPollTimer = null;
      console.warn("[AI Analyze] polling timeout, give up");
      return;
    }
    try {
      const token = auth.token || localStorage.getItem("token");
      const res = await axios.post("/api/video/ai-analyze", { video_id: videoId }, {
        headers: token ? { Authorization: "Bearer " + token } : {},
      });
      const data = res.data?.data || res.data || {};
      const status = Number(data.status ?? 0);
      const summary = data.summary || "";

      aiSummaryStatus.value = status;

      if (status === 2 && summary) {
        aiSummary.value = summary;
        // 同步回 video 对象，方便其他地方使用
        if (video.value) {
          video.value.ai_summary = summary;
          video.value.ai_summary_status = 2;
        }
        startTypewriter(summary);
        clearInterval(aiPollTimer);
        aiPollTimer = null;
      } else if (status === 3) {
        // 失败：停止轮询，UI 渲染失败状态
        clearInterval(aiPollTimer);
        aiPollTimer = null;
      }
      // status 0 / 1：继续轮询
    } catch (err) {
      console.error("[AI Analyze] poll error:", err);
      // 单次失败不中断轮询，等下次再试
    }
  }, AI_POLL_INTERVAL);
}

function startTypewriter(fullText) {
  if (typewriterTimer) clearInterval(typewriterTimer);
  typewriterText.value = "";
  let idx = 0;
  typewriterTimer = setInterval(() => {
    if (idx < fullText.length) {
      typewriterText.value += fullText[idx];
      idx++;
    } else {
      clearInterval(typewriterTimer);
      typewriterTimer = null;
    }
  }, 35); // 逐字流式打字输出
}
</script>

<style scoped>
.player-area { position: relative; width: 100%; aspect-ratio: 16/9; background: #000; overflow: hidden; }
.danmaku-canvas { position: absolute; inset: 0; pointer-events: none; z-index: 1; }
.info-section { padding: 16px; margin: 12px 14px 0; }
.video-title { font-size: 18px; font-weight: 700; line-height: 1.4; margin-bottom: 10px; }
.info-row { font-size: 12px; color: var(--text-muted); display: flex; align-items: center; gap: 8px; flex-wrap: wrap; margin-bottom: 8px; }
.dot { color: var(--border); }
.video-desc { font-size: 13px; color: var(--text-secondary); line-line-height: 1.6; }
.video-desc.desc-empty { font-style: italic; color: var(--text-muted); opacity: 0.8; font-size: 11.5px; border-left: 2px solid var(--border); padding-left: 10px; margin-top: 10px; line-height: 1.5; }

/* AI Summary Display Box below author description */
.ai-summary-box {
  margin-top: 14px;
  background: rgba(15, 23, 42, 0.3);
  border: 1px dashed rgba(0, 240, 255, 0.25);
  border-radius: 8px;
  padding: 12px 14px;
  box-shadow: 0 0 10px rgba(0, 240, 255, 0.02);
}
.ai-summary-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}
.ai-badge-neon {
  font-family: var(--font-display);
  font-size: 9.5px;
  font-weight: bold;
  color: #00f0ff;
  background: rgba(0, 240, 255, 0.08);
  padding: 2px 6px;
  border-radius: 4px;
  box-shadow: 0 0 6px rgba(0, 240, 255, 0.1);
}
.ai-summary-text {
  font-size: 12.5px;
  color: var(--text-secondary);
  line-height: 1.6;
}
.ai-loading-container {
  margin-top: 10px;
}
.loading-subtext {
  font-size: 11px;
  color: var(--text-muted);
  font-style: italic;
  margin-top: 8px;
  animation: spark-blink 1.2s infinite alternate;
}
.pulse-spark.spark-active {
  animation: spark-blink 0.8s infinite alternate;
}
@keyframes ai-pulse-anim {
  0% { text-shadow: 0 0 5px rgba(0,240,255,0.05); opacity: 0.7; }
  50% { text-shadow: 0 0 12px rgba(0,240,255,0.45); opacity: 1.0; color: var(--cyan); }
  100% { text-shadow: 0 0 5px rgba(0,240,255,0.05); opacity: 0.7; }
}
.ai-pulse {
  animation: ai-pulse-anim 1.6s infinite ease-in-out;
  font-weight: 500;
}
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

/* AI Premium Neon Skeleton Card */
.ai-skeleton-card {
  margin-top: 12px;
  background: rgba(15, 23, 42, 0.45);
  border: 1px solid rgba(0, 240, 255, 0.15);
  border-radius: 8px;
  padding: 14px;
  box-shadow: 0 0 15px rgba(0, 240, 255, 0.05);
  position: relative;
  overflow: hidden;
}
.ai-skeleton-card::after {
  content: "";
  position: absolute;
  top: 0; right: 0; bottom: 0; left: 0;
  background: linear-gradient(90deg, transparent, rgba(157, 78, 221, 0.2), transparent);
  transform: translateX(-100%);
  animation: shimmer-anim 1.8s infinite;
}
@keyframes shimmer-anim {
  100% { transform: translateX(100%); }
}
.ai-skeleton-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.ai-chip-groping {
  font-family: var(--font-display);
  font-size: 10px;
  font-weight: bold;
  letter-spacing: 1px;
  text-transform: uppercase;
  color: #00f0ff;
  background: rgba(0, 240, 255, 0.15);
  padding: 2px 8px;
  border-radius: 4px;
  box-shadow: 0 0 8px rgba(0, 240, 255, 0.2);
}
.pulse-spark {
  width: 6px; height: 6px;
  border-radius: 50%;
  background: #9d4edd;
  box-shadow: 0 0 8px #9d4edd;
  animation: spark-blink 1s infinite alternate;
}
@keyframes spark-blink {
  0% { transform: scale(0.8); background: #9d4edd; opacity: 0.5; }
  100% { transform: scale(1.3); background: #00f0ff; box-shadow: 0 0 12px #00f0ff; opacity: 1; }
}
.skeleton-bar {
  height: 8px;
  background: rgba(148, 163, 184, 0.15);
  border-radius: 4px;
  margin-bottom: 8px;
}
.skeleton-bar.bar-long { width: 90%; }
.skeleton-bar.bar-medium { width: 75%; }
.skeleton-bar.bar-short { width: 40%; margin-bottom: 0; }
.ai-failed-container {
  margin-top: 6px;
}
.ai-failed-text {
  font-size: 12px;
  color: #ff6b9d;
  line-height: 1.5;
  font-style: italic;
}
</style>
