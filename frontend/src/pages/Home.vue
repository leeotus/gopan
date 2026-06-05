<template>
  <div class="page-container">
    <!-- Category chips -->
    <div class="category-strip">
      <button
        v-for="cat in categories" :key="cat.value"
        :class="['cat-chip', { active: activeCategory === cat.value }]"
        @click="switchCategory(cat.value)"
      >{{ cat.label }}</button>
    </div>

    <div class="page-content">
      <!-- Video grid -->
      <div class="video-grid">
        <div
          v-for="(v, i) in videoStore.videos" :key="v.id"
          class="video-card anim-fade-up"
          :style="{ animationDelay: i * 0.05 + 's' }"
          @click="$router.push(`/video/${v.id}`)"
        >
          <div class="card-cover">
            <img
              :src="v.cover_url || `/covers/${v.id}.jpg`"
              :alt="v.title"
              loading="lazy"
              @error="(e) => { e.target.style.display = 'none'; e.target.nextElementSibling.style.display = 'flex'; }"
            />
            <div class="cover-placeholder" style="display:none">
              <div class="cover-shimmer"></div>
              <svg width="40" height="40" viewBox="0 0 24 24" fill="rgba(0,240,255,0.3)"><polygon points="5,3 19,12 5,21"/></svg>
            </div>
            <div class="cover-overlay"></div>
            <div class="cover-duration" v-if="v.duration">{{ formatDuration(v.duration) }}</div>
          </div>
          <div class="card-body">
            <div class="card-title">{{ v.title || 'Untitled' }}</div>
            <div class="card-footer">
              <span class="card-user">{{ v.username || 'Anonymous' }}</span>
              <span class="card-sep">·</span>
              <span class="card-views">{{ formatCount(v.play_count) }} plays</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Load more -->
      <div class="load-more" v-if="videoStore.hasMore">
        <button class="btn-primary" @click="loadMore" :disabled="videoStore.loading">
          {{ videoStore.loading ? 'Loading...' : 'Load More' }}
        </button>
      </div>

      <!-- Empty -->
      <div class="empty-state" v-if="!videoStore.loading && videoStore.videos.length === 0">
        <div class="empty-icon">📺</div>
        <p class="empty-text">No videos yet</p>
        <p class="empty-sub">Be the first to upload!</p>
      </div>
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
  { label: "All", value: "all" },
  { label: "Tech", value: "tech" },
  { label: "Music", value: "music" },
  { label: "Gaming", value: "gaming" },
  { label: "Sports", value: "sports" },
];

onMounted(() => videoStore.fetchVideos({ cursor: 0, sort: "newest" }));

watch(activeCategory, (v) =>
  videoStore.fetchVideos({ cursor: 0, sort: "newest", category: v === "all" ? "" : v })
);

function switchCategory(v) { activeCategory.value = v; }
function loadMore() {
  videoStore.fetchVideos({
    cursor: videoStore.nextCursor,
    sort: "newest",
    category: activeCategory.value === "all" ? "" : activeCategory.value,
  });
}
</script>

<style scoped>
/* ── Category strip ── */
.category-strip {
  display: flex;
  gap: 8px;
  padding: 12px 16px;
  overflow-x: auto;
  white-space: nowrap;
  scrollbar-width: none;
}
.category-strip::-webkit-scrollbar { display: none; }
.cat-chip {
  padding: 6px 18px;
  border: 1px solid var(--border);
  border-radius: 20px;
  background: transparent;
  color: var(--text-secondary);
  font-family: var(--font-body);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--duration) var(--ease-out);
  flex-shrink: 0;
}
.cat-chip.active {
  background: rgba(0,240,255,0.08);
  border-color: var(--cyan-dim);
  color: var(--cyan);
  box-shadow: 0 0 12px var(--cyan-glow);
}
.cat-chip:hover:not(.active) { border-color: var(--border-glow); color: var(--text-primary); }

/* ── Video grid ── */
.video-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
  padding: 0 12px;
}
@media (min-width: 600px) { .video-grid { grid-template-columns: repeat(3, 1fr); } }

.video-card {
  cursor: pointer;
  border-radius: var(--radius);
  overflow: hidden;
  background: var(--bg-card);
  border: 1px solid var(--border);
  transition: all var(--duration) var(--ease-out);
}
.video-card:hover {
  border-color: var(--cyan-dim);
  transform: translateY(-2px);
  box-shadow: var(--shadow-card-hover);
}

/* Cover */
.card-cover {
  position: relative;
  aspect-ratio: 16/10;
  background: var(--bg-secondary);
  overflow: hidden;
}
.card-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.6s var(--ease-out);
}
.video-card:hover .card-cover img { transform: scale(1.05); }
.cover-placeholder {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}
.cover-shimmer {
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, rgba(0,240,255,0.04), rgba(179,71,234,0.04));
}
.cover-overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(to top, rgba(5,5,10,0.8), transparent);
}
.cover-duration {
  position: absolute;
  bottom: 6px;
  right: 6px;
  padding: 2px 8px;
  background: rgba(0,0,0,0.7);
  border-radius: 4px;
  font-size: 10px;
  font-weight: 600;
  color: #fff;
}

/* Body */
.card-body { padding: 10px 10px 12px; }
.card-title {
  font-size: 13px;
  font-weight: 600;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  color: var(--text-primary);
  margin-bottom: 6px;
}
.card-footer {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: var(--text-muted);
}
.card-sep { color: var(--border); }

/* Load more */
.load-more { padding: 20px 16px 30px; text-align: center; }

/* Empty */
.empty-state { text-align: center; padding: 80px 20px; }
.empty-icon { font-size: 48px; margin-bottom: 16px; opacity: 0.6; }
.empty-text { font-size: 18px; font-weight: 600; color: var(--text-secondary); }
.empty-sub { font-size: 13px; color: var(--text-muted); margin-top: 6px; }
</style>
