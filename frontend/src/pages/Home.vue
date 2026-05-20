<template>
  <div class="page-container">
    <!-- 顶部 -->
    <van-nav-bar title="GoPan" fixed placeholder>
      <template #right>
        <van-icon name="search" size="20" @click="$router.push('/search')" />
      </template>
    </van-nav-bar>

    <!-- 分类 Tab -->
    <van-tabs v-model:active="activeCategory" sticky offset-top="46">
      <van-tab v-for="cat in categories" :key="cat.value" :title="cat.label" :name="cat.value" />
    </van-tabs>

    <!-- 视频网格 -->
    <div class="page-content">
      <div class="video-grid">
        <div
          v-for="video in videoStore.videos"
          :key="video.id"
          class="video-card"
          @click="$router.push(`/video/${video.id}`)"
        >
          <div class="card-cover">
            <img :src="video.cover_url" :alt="video.title" />
            <span class="card-duration">{{ formatDuration(video.duration) }}</span>
            <span class="card-views">{{ formatCount(video.play_count) }} 播放</span>
          </div>
          <div class="card-title">{{ video.title }}</div>
          <div class="card-meta">
            <van-icon name="user-o" size="12" />
            <span>{{ video.username }}</span>
          </div>
        </div>
      </div>

      <!-- 加载更多 -->
      <div class="load-more" v-if="videoStore.hasMore">
        <van-button :loading="videoStore.loading" loading-text="加载中..." size="small" block @click="loadMore">
          加载更多
        </van-button>
      </div>

      <van-empty v-if="!videoStore.loading && videoStore.videos.length === 0" description="暂无视频" />
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onMounted } from "vue";
import { useVideoStore } from "../stores/video";
import { formatDuration, formatCount } from "../composables/utils";

const videoStore = useVideoStore();
const activeCategory = ref("all");

const categories = [
  { label: "全部", value: "all" },
  { label: "技术", value: "技术" },
  { label: "前端", value: "前端" },
  { label: "数据", value: "数据" },
  { label: "基础", value: "基础" },
];

onMounted(() => {
  videoStore.fetchVideos({ cursor: 0, sort: "newest" });
});

watch(activeCategory, (val) => {
  videoStore.fetchVideos({
    cursor: 0,
    sort: "newest",
    category: val === "all" ? "" : val,
  });
});

function loadMore() {
  videoStore.fetchVideos({
    cursor: videoStore.nextCursor,
    sort: "newest",
    category: activeCategory.value === "all" ? "" : activeCategory.value,
  });
}
</script>

<style scoped>
.video-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.video-card {
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
}

.card-cover {
  position: relative;
  width: 100%;
  aspect-ratio: 16 / 9;
  background: #eee;
  overflow: hidden;
}

.card-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.card-duration {
  position: absolute;
  bottom: 6px;
  right: 6px;
  background: rgba(0, 0, 0, 0.7);
  color: #fff;
  font-size: 11px;
  padding: 1px 5px;
  border-radius: 3px;
}

.card-views {
  position: absolute;
  bottom: 6px;
  left: 6px;
  background: rgba(0, 0, 0, 0.7);
  color: #fff;
  font-size: 11px;
  padding: 1px 5px;
  border-radius: 3px;
}

.card-title {
  padding: 8px 8px 4px;
  font-size: 13px;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  color: var(--gopan-text);
}

.card-meta {
  padding: 0 8px 8px;
  font-size: 11px;
  color: var(--gopan-text-secondary);
  display: flex;
  align-items: center;
  gap: 4px;
}

.load-more {
  margin-top: 16px;
}
</style>
