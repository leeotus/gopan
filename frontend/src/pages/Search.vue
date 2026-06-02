<template>
  <div class="page-container">
    <div class="page-content" style="padding:16px">
      <!-- Search bar -->
      <div class="search-hero">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--cyan-dim)" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
        <input
          v-model="keyword"
          class="search-hero-input"
          placeholder="Search videos, creators, categories..."
          @keyup.enter="doSearch"
          ref="searchInput"
          autofocus
        />
      </div>

      <!-- Loading -->
      <div v-if="loading" class="loading-state">
        <div class="loading-dot"></div>
        <div class="loading-dot"></div>
        <div class="loading-dot"></div>
      </div>

      <!-- No results -->
      <div v-if="!loading && keyword && results.length === 0" class="empty-state">
        <div class="empty-icon">🔍</div>
        <p class="empty-text">No results for "{{ keyword }}"</p>
      </div>

      <!-- Results -->
      <div v-if="results.length" class="result-list">
        <div
          v-for="(item, i) in results" :key="item.id"
          class="result-card card card-clickable anim-fade-up"
          :style="{ animationDelay: i * 0.04 + 's' }"
          @click="$router.push(`/video/${item.id}`)"
        >
          <div class="result-cover">
            <img :src="item.cover_url" v-if="item.cover_url" />
            <div class="cover-placeholder" v-else>
              <svg width="28" height="28" viewBox="0 0 24 24" fill="rgba(0,240,255,0.15)"><polygon points="5,3 19,12 5,21"/></svg>
            </div>
          </div>
          <div class="result-body">
            <div class="result-title">{{ item.title }}</div>
            <div class="result-meta">
              <span>{{ item.username }}</span>
              <span class="dot">·</span>
              <span>{{ formatCount(item.play_count) }} plays</span>
            </div>
            <div class="result-desc" v-if="item.description">{{ item.description }}</div>
          </div>
        </div>
      </div>

      <!-- Initial state -->
      <div v-if="!keyword && results.length === 0" class="empty-state">
        <div class="empty-icon">🎬</div>
        <p class="empty-text">Discover amazing content</p>
        <p class="empty-sub">Type a keyword to search</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useRoute } from "vue-router";
import { searchApi } from "../api";
import { useVideoStore } from "../stores/video";
import { formatCount } from "../composables/utils";

const route = useRoute();
const videoStore = useVideoStore();
const keyword = ref("");
const results = ref([]);
const loading = ref(false);
const searchInput = ref(null);

onMounted(() => {
  const q = route.query.q;
  if (q) { keyword.value = q; doSearch(); }
});

async function doSearch() {
  if (!keyword.value.trim()) return;
  loading.value = true;
  try {
    const r = await searchApi.search({ keyword: keyword.value, page: 1, size: 20 });
    results.value = r?.videos || r?.data?.videos || [];
  } catch {
    results.value = videoStore.videos.filter(v => v.title?.includes(keyword.value));
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.search-hero {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 18px;
  background: var(--bg-input);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  margin-bottom: 20px;
  transition: border-color var(--duration), box-shadow var(--duration);
}
.search-hero:focus-within {
  border-color: var(--cyan-dim);
  box-shadow: 0 0 20px var(--cyan-glow);
}
.search-hero-input {
  flex: 1;
  border: none;
  background: transparent;
  color: var(--text-primary);
  font-family: var(--font-body);
  font-size: 14px;
  outline: none;
}
.search-hero-input::placeholder { color: var(--text-muted); }

.result-list { display: flex; flex-direction: column; gap: 10px; }
.result-card { display: flex; gap: 14px; padding: 0; }
.result-cover {
  width: 140px;
  height: 84px;
  flex-shrink: 0;
  background: var(--bg-secondary);
}
.result-cover img { width: 100%; height: 100%; object-fit: cover; }
.cover-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, rgba(0,240,255,0.04), rgba(179,71,234,0.04));
}
.result-body { padding: 12px 12px 12px 0; flex: 1; display: flex; flex-direction: column; justify-content: center; }
.result-title { font-size: 14px; font-weight: 600; line-height: 1.4; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.result-meta { font-size: 11px; color: var(--text-muted); margin-top: 6px; display: flex; gap: 4px; align-items: center; }
.dot { color: var(--border); }
.result-desc { font-size: 11px; color: var(--text-secondary); margin-top: 4px; display: -webkit-box; -webkit-line-clamp: 1; -webkit-box-orient: vertical; overflow: hidden; }

.loading-state { display: flex; justify-content: center; gap: 8px; padding: 40px; }
.loading-dot {
  width: 8px; height: 8px;
  border-radius: 50%;
  background: var(--cyan);
  animation: pulse 0.8s var(--ease-spring) infinite alternate;
}
.loading-dot:nth-child(2) { animation-delay: 0.2s; background: var(--purple); }
.loading-dot:nth-child(3) { animation-delay: 0.4s; background: var(--pink); }
@keyframes pulse { to { transform: scale(1.5); opacity: 0.5; } }

.empty-state { text-align: center; padding: 80px 20px; }
.empty-icon { font-size: 48px; margin-bottom: 16px; opacity: 0.6; }
.empty-text { font-size: 18px; font-weight: 600; color: var(--text-secondary); }
.empty-sub { font-size: 13px; color: var(--text-muted); margin-top: 6px; }
</style>
