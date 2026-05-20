<template>
  <div class="page-container">
    <van-nav-bar title="搜索" fixed placeholder />

    <div class="search-bar">
      <van-search
        v-model="keyword"
        placeholder="搜索视频"
        show-action
        @search="handleSearch"
        @cancel="keyword = ''"
      />
    </div>

    <div class="page-content">
      <div v-if="searchResults.length > 0">
        <div
          v-for="item in searchResults"
          :key="item.id"
          class="search-item"
          @click="$router.push(`/video/${item.id}`)"
        >
          <img :src="item.cover_url" class="search-thumb" />
          <div class="search-info">
            <div class="search-title">{{ item.title }}</div>
            <div class="search-meta">
              <span>{{ formatCount(item.play_count) }} 播放</span>
              <span class="meta-divider">·</span>
              <span>{{ item.username }}</span>
              <span class="meta-divider">·</span>
              <span>{{ formatTime(item.created_at) }}</span>
            </div>
          </div>
        </div>
      </div>

      <van-empty v-else description="输入关键词搜索视频" />
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { useVideoStore } from "../stores/video";
import { formatCount, formatTime } from "../composables/utils";

const videoStore = useVideoStore();
const keyword = ref("");
const searchResults = ref([]);

async function handleSearch() {
  if (!keyword.value.trim()) return;
  try {
    const res = await videoStore.search(keyword.value);
    searchResults.value = res.data?.videos || [];
  } catch {
    // 本地过滤 mock 数据
    const kw = keyword.value.toLowerCase();
    searchResults.value = videoStore.videos.filter(
      (v) => v.title.toLowerCase().includes(kw) || v.username.toLowerCase().includes(kw)
    );
  }
}
</script>

<style scoped>
.search-bar {
  background: #fff;
}

.search-item {
  display: flex;
  gap: 12px;
  padding: 10px 0;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
}

.search-thumb {
  width: 140px;
  height: 78px;
  border-radius: 6px;
  object-fit: cover;
  flex-shrink: 0;
  background: #eee;
}

.search-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.search-title {
  font-size: 14px;
  font-weight: 500;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.search-meta {
  font-size: 12px;
  color: var(--gopan-text-secondary);
}

.meta-divider {
  margin: 0 4px;
}
</style>
