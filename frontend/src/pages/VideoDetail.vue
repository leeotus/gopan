<template>
  <div class="page-container">
    <van-nav-bar fixed placeholder>
      <template #left>
        <van-icon name="arrow-left" size="20" @click="$router.back()" />
      </template>
      <template #title>
        <span class="nav-title">{{ video?.title?.slice(0, 15) || "视频详情" }}</span>
      </template>
    </van-nav-bar>

    <!-- 播放器占位 -->
    <div class="player-area">
      <img v-if="video?.cover_url" :src="video.cover_url" class="player-cover" />
      <div v-else class="player-placeholder">
        <van-icon name="play-circle-o" size="48" color="#fff" />
      </div>
      <div class="player-controls">
        <van-icon name="play-circle-o" size="52" color="rgba(255,255,255,0.9)" @click="handlePlay" />
      </div>
    </div>

    <!-- 视频信息 -->
    <div class="info-section" v-if="video">
      <h2 class="video-title">{{ video.title }}</h2>

      <div class="info-row">
        <span class="info-text">{{ formatCount(video.play_count) }} 次播放</span>
        <span class="info-divider">·</span>
        <span class="info-text">{{ formatTime(video.created_at) }}</span>
        <span class="info-divider">·</span>
        <span class="info-text info-category">{{ video.category }}</span>
      </div>

      <div class="video-desc" v-if="video.description">
        {{ video.description }}
      </div>

      <div class="transcode-tags">
        <van-tag
          v-for="t in video.transcodes"
          :key="t.resolution"
          type="primary"
          size="medium"
          style="margin-right: 6px"
        >
          {{ t.resolution }}
        </van-tag>
      </div>
    </div>

    <!-- 互动按钮栏 -->
    <div class="action-bar" v-if="video">
      <div class="action-item" @click="handleLike">
        <van-icon :name="video.liked ? 'like' : 'like-o'" :color="video.liked ? '#ee0a24' : '#666'" size="22" />
        <span :class="{ active: video.liked }">{{ formatCount(video.like_count) }}</span>
      </div>
      <div class="action-item" @click="handleFavorite">
        <van-icon :name="video.favorited ? 'star' : 'star-o'" :color="video.favorited ? '#ff976a' : '#666'" size="22" />
        <span :class="{ active: video.favorited }">{{ video.favorited ? '已收藏' : '收藏' }}</span>
      </div>
      <div class="action-item" @click="scrollToComments">
        <van-icon name="comment-o" size="22" />
        <span>评论</span>
      </div>
    </div>

    <!-- 评论区域 -->
    <div class="comment-section" ref="commentSection">
      <div class="section-title">评论</div>

      <div class="comment-input" v-if="authStore.isLoggedIn">
        <van-field
          v-model="commentText"
          rows="1"
          type="textarea"
          placeholder="发条友善的评论..."
          autosize
        >
          <template #button>
            <van-button size="small" type="primary" @click="handlePostComment" :loading="commenting">
              发表
            </van-button>
          </template>
        </van-field>
      </div>
      <div class="comment-input" v-else>
        <van-button block plain type="primary" size="small" to="/login">登录后发表评论</van-button>
      </div>

      <div v-for="c in comments" :key="c.id" class="comment-item">
        <div class="comment-header">
          <img :src="c.avatar || defaultAvatar" class="comment-avatar" />
          <span class="comment-username">{{ c.username }}</span>
          <span class="comment-time">{{ formatTime(c.created_at) }}</span>
        </div>
        <div class="comment-content">{{ c.content }}</div>
      </div>

      <van-empty v-if="comments.length === 0" description="暂无评论，快来抢沙发吧~" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useRoute } from "vue-router";
import { showToast } from "vant";
import { useVideoStore } from "../stores/video";
import { useAuthStore } from "../stores/auth";
import { formatCount, formatTime } from "../composables/utils";

const route = useRoute();
const videoStore = useVideoStore();
const authStore = useAuthStore();

const video = ref(null);
const comments = ref([]);
const commentText = ref("");
const commenting = ref(false);
const commentSection = ref(null);

const defaultAvatar = "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'%3E%3Ccircle cx='50' cy='50' r='50' fill='%23ddd'/%3E%3Ccircle cx='50' cy='40' r='16' fill='%23aaa'/%3E%3Cellipse cx='50' cy='78' rx='30' ry='18' fill='%23aaa'/%3E%3C/svg%3E";

onMounted(async () => {
  const id = Number(route.params.id);
  await videoStore.fetchDetail(id);
  video.value = videoStore.currentVideo;
  // mock comments
  comments.value = [
    { id: 1, username: "张三", avatar: "", content: "讲得真好！收获很多。", created_at: Math.floor(Date.now() / 1000) - 3600 },
    { id: 2, username: "李四", avatar: "", content: "请问有配套代码吗？", created_at: Math.floor(Date.now() / 1000) - 7200, },
    { id: 3, username: "王五", avatar: "", content: "已收藏，等更新",created_at: Math.floor(Date.now() / 1000) - 10800,},
  ];
});

function handlePlay() {
  showToast("播放功能需接入 HLS 播放器 (后续添加)");
}

async function handleLike() {
  if (!authStore.isLoggedIn) {
    showToast("请先登录");
    return;
  }
  await videoStore.toggleLike(video.value.id, video.value.liked);
}

async function handleFavorite() {
  if (!authStore.isLoggedIn) {
    showToast("请先登录");
    return;
  }
  await videoStore.toggleFavorite(video.value.id, video.value.favorited);
  video.value.favorited = !video.value.favorited;
  showToast(video.value.favorited ? "已收藏" : "已取消收藏");
}

async function handlePostComment() {
  if (!commentText.value.trim()) return;
  commenting.value = true;
  try {
    // 本地 mock 添加
    comments.value.unshift({
      id: Date.now(),
      username: authStore.user?.username || "我",
      avatar: "",
      content: commentText.value,
      created_at: Math.floor(Date.now() / 1000),
    });
    commentText.value = "";
    showToast("评论成功");
  } catch {
    showToast("评论失败");
  } finally {
    commenting.value = false;
  }
}

function scrollToComments() {
  commentSection.value?.scrollIntoView({ behavior: "smooth" });
}
</script>

<style scoped>
.nav-title {
  font-size: 15px;
}

.player-area {
  position: relative;
  width: 100%;
  aspect-ratio: 16 / 9;
  background: #000;
  overflow: hidden;
}

.player-cover {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.player-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.player-controls {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}

.info-section {
  padding: 14px 16px;
  background: #fff;
  margin-bottom: 8px;
}

.video-title {
  font-size: 17px;
  font-weight: 600;
  line-height: 1.5;
  margin-bottom: 8px;
}

.info-row {
  font-size: 13px;
  color: var(--gopan-text-secondary);
  margin-bottom: 10px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.info-divider {
  color: #ccc;
}

.info-category {
  color: var(--gopan-primary);
}

.video-desc {
  font-size: 14px;
  color: var(--gopan-text);
  line-height: 1.6;
  margin-bottom: 10px;
}

.transcode-tags {
  margin-top: 8px;
}

/* 互动栏 */
.action-bar {
  display: flex;
  justify-content: space-around;
  padding: 12px 0;
  background: #fff;
  margin-bottom: 8px;
  border-top: 1px solid #f0f0f0;
}

.action-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #666;
  cursor: pointer;
}

.action-item span.active {
  color: #ee0a24;
}

/* 评论区 */
.comment-section {
  background: #fff;
  padding: 0 16px 80px;
}

.comment-input {
  margin-bottom: 16px;
}

.comment-item {
  padding: 12px 0;
  border-bottom: 1px solid #f5f5f5;
}

.comment-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.comment-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #eee;
}

.comment-username {
  font-size: 13px;
  font-weight: 500;
}

.comment-time {
  font-size: 11px;
  color: var(--gopan-text-secondary);
  margin-left: auto;
}

.comment-content {
  font-size: 14px;
  line-height: 1.6;
  padding-left: 36px;
}
</style>
