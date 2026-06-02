<template>
  <div class="page-container">
    <div class="page-content" style="padding:14px">
      <van-search
        v-model="keyword"
        placeholder="搜索视频"
        show-action
        @search="doSearch"
        @cancel="keyword='';results=[]"
        shape="round"
        :style="{ '--van-search-background': '#12121a', '--van-search-action-text-color': '#8b5cf6' }"
      />
      <div v-if="keyword && results.length === 0 && !loading" style="text-align:center;padding:40px;color:var(--text-muted)">暂无结果</div>
      <div v-if="results.length" class="result-list">
        <div v-for="item in results" :key="item.id"
          class="card card-clickable" style="display:flex;gap:12px;padding:0;margin-bottom:12px;overflow:hidden;animation:fadeInUp 0.3s ease-out"
          @click="$router.push(`/video/${item.id}`)">
          <img :src="item.cover_url" style="width:140px;height:80px;object-fit:cover;flex-shrink:0" />
          <div style="padding:10px 10px 10px 0;flex:1;display:flex;flex-direction:column;justify-content:space-between">
            <div style="font-size:14px;font-weight:600;line-height:1.4;display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden">{{ item.title }}</div>
            <div class="card-meta">{{ formatCount(item.play_count) }} 播放 · {{ item.username }}</div>
          </div>
        </div>
      </div>
      <div v-else-if="!keyword" style="text-align:center;padding:40px;color:var(--text-muted)">输入关键词搜索视频</div>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { searchApi } from "../api";
import { useVideoStore } from "../stores/video";
import { formatCount } from "../composables/utils";

const videoStore = useVideoStore();
const keyword = ref("");
const results = ref([]);
const loading = ref(false);

async function doSearch() {
  if (!keyword.value.trim()) return;
  loading.value = true;
  try {
    const r = await searchApi.search({ keyword: keyword.value, page: 1, size: 20 });
    results.value = r?.videos || r?.data?.videos || [];
  } catch {
    results.value = videoStore.videos.filter(v => v.title.includes(keyword.value));
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.card-meta { font-size: 11px; color: var(--text-muted); }
</style>
