<template>
  <div class="page-container">
    <div class="category-bar">
      <span v-for="cat in categories" :key="cat.value"
        :class="['cat-chip', { active: activeCategory === cat.value }]"
        @click="switchCategory(cat.value)">{{ cat.label }}</span>
    </div>

    <div class="page-content" style="padding:14px">
      <div class="video-grid">
        <div v-for="(v, i) in videoStore.videos" :key="v.id"
          class="card anim-fade-up" :style="{ animationDelay: i * 0.04 + 's' }"
          @click="$router.push(`/video/${v.id}`)">
          <div class="card-cover">
            <img :src="v.cover_url" :alt="v.title" loading="lazy" />
            <div class="cover-badge">
              <span class="badge-duration">{{ formatDuration(v.duration) }}</span>
            </div>
            <div class="cover-overlay">
              <svg width="36" height="36" viewBox="0 0 24 24" fill="rgba(255,255,255,0.9)"><polygon points="5,3 19,12 5,21"/></svg>
            </div>
          </div>
          <div class="card-body">
            <div class="card-title">{{ v.title }}</div>
            <div class="card-footer">
              <span class="card-username">{{ v.username }}</span>
              <span class="card-views">{{ formatCount(v.play_count) }} 播放</span>
            </div>
          </div>
        </div>
      </div>

      <div class="load-more" v-if="videoStore.hasMore">
        <button class="btn-primary" @click="loadMore" :disabled="videoStore.loading" style="width:100%">
          {{ videoStore.loading ? "加载中..." : "加载更多" }}
        </button>
      </div>

      <van-empty v-if="!videoStore.loading && videoStore.videos.length === 0" description="暂无视频" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from "vue";
import { useVideoStore } from "../stores/video";
import { formatDuration, formatCount } from "../composables/utils";

const videoStore = useVideoStore();
const activeCategory = ref("all");
const categories = [
  { label: "全部", value: "all" }, { label: "技术", value: "技术" },
  { label: "前端", value: "前端" }, { label: "数据", value: "数据" }, { label: "基础", value: "基础" },
];

onMounted(() => videoStore.fetchVideos({ cursor: 0, sort: "newest" }));
watch(activeCategory, (v) => videoStore.fetchVideos({ cursor: 0, sort: "newest", category: v === "all" ? "" : v }));
function switchCategory(v) { activeCategory.value = v; }
function loadMore() { videoStore.fetchVideos({ cursor: videoStore.nextCursor, sort: "newest", category: activeCategory.value === "all" ? "" : activeCategory.value }); }
</script>

<style scoped>
.category-bar { display: flex; gap: 8px; padding: 10px 16px; overflow-x: auto; white-space: nowrap; }
.cat-chip {
  padding: 6px 16px; border-radius: 20px; font-size: 13px; font-weight: 500;
  background: var(--bg-card); border: 1px solid var(--border); color: var(--text-secondary);
  cursor: pointer; transition: all var(--transition);
}
.cat-chip:active { transform: scale(0.95); }
.cat-chip.active { background: var(--accent); border-color: var(--accent); color: #fff; box-shadow: 0 4px 12px var(--accent-glow); }

.video-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 14px; }

.card { animation-fill-mode: both; }
.card-cover {
  position: relative; width: 100%; aspect-ratio: 16/10; background: #0a0a10; overflow: hidden;
}
.card-cover img { width: 100%; height: 100%; object-fit: cover; transition: transform 0.3s; }
.card:hover .card-cover img { transform: scale(1.05); }
.cover-badge { position: absolute; bottom: 8px; right: 8px; }
.badge-duration { background: rgba(0,0,0,0.75); backdrop-filter: blur(4px); color: #fff; font-size: 11px; padding: 2px 7px; border-radius: 4px; }
.cover-overlay {
  position: absolute; top: 0; left: 0; right: 0; bottom: 0;
  display: flex; align-items: center; justify-content: center;
  background: rgba(0,0,0,0.2); opacity: 0; transition: opacity 0.2s;
}
.card:hover .cover-overlay { opacity: 1; }

.card-body { padding: 10px; }
.card-title { font-size: 13px; font-weight: 600; line-height: 1.4; color: var(--text-primary); display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.card-footer { display: flex; justify-content: space-between; align-items: center; margin-top: 6px; }
.card-username { font-size: 11px; color: var(--text-muted); }
.card-views { font-size: 11px; color: var(--accent-light); }

.load-more { margin-top: 20px; }
</style>
