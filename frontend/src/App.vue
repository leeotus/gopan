<template>
  <div id="app-root">
    <!-- 顶部搜索栏 -->
    <div class="top-bar" v-if="showSearchBar">
      <div class="top-logo" @click="$router.push('/')">✦ GoPan</div>
      <div class="search-box-inline">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#8b8baa" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
        <input
          v-model="searchKeyword"
          class="search-input"
          placeholder="搜索视频..."
          @keyup.enter="doSearch"
        />
      </div>
      <div class="top-avatar" @click="$router.push('/profile')">
        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#e8e6f0" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
      </div>
    </div>

    <!-- 普通页面顶部 -->
    <div class="top-bar-simple" v-else-if="showSimpleBar">
      <svg class="back-icon" @click="$router.back()" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#e8e6f0" stroke-width="2"><path d="m15 18-6-6 6-6"/></svg>
      <span class="top-title">{{ pageTitle }}</span>
      <div style="width:22px" />
    </div>

    <router-view v-slot="{ Component }">
      <transition name="fade" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>

    <!-- 底部导航 -->
    <van-tabbar v-if="showTabbar" v-model="activeTab" route :fixed="true" :border="false"
      active-color="#8b5cf6" inactive-color="#5a5a7a"
      :style="{ background: 'linear-gradient(180deg, #12121a 0%, #0f0f18 100%)', borderTop: '1px solid #232340' }">
      <van-tabbar-item to="/">
        <template #icon><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg></template>
        首页
      </van-tabbar-item>
      <van-tabbar-item to="/upload">
        <template #icon><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="16"/><line x1="8" y1="12" x2="16" y2="12"/></svg></template>
        上传
      </van-tabbar-item>
      <van-tabbar-item to="/profile">
        <template #icon><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg></template>
        我的
      </van-tabbar-item>
    </van-tabbar>
  </div>
</template>

<script setup>
import { computed, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

const route = useRoute();
const router = useRouter();
const searchKeyword = ref("");

function doSearch() {
  const kw = searchKeyword.value.trim();
  if (kw) {
    router.push({ path: "/search", query: { q: kw } });
  }
}

const showSearchBar = computed(() => ["/", "/search"].includes(route.path));
const showSimpleBar = computed(() => !showSearchBar.value && !["/login", "/register"].includes(route.path));
const showTabbar = computed(() => ["/", "/upload", "/profile"].includes(route.path));
const activeTab = computed(() => ({ "/": 0, "/upload": 1, "/profile": 2 }[route.path] ?? 0));
const pageTitle = computed(() => route.meta?.title || "");
</script>

<style scoped>
.top-bar {
  position: sticky; top: 0; z-index: 100;
  display: flex; align-items: center; gap: 12px;
  padding: 10px 16px; backdrop-filter: blur(20px);
  background: rgba(10, 10, 15, 0.85);
  border-bottom: 1px solid var(--border);
}
.top-logo { font-size: 18px; font-weight: 800; background: linear-gradient(135deg, var(--accent), #c084fc); -webkit-background-clip: text; -webkit-text-fill-color: transparent; cursor: pointer; }
.search-box-inline { flex: 1; display: flex; align-items: center; gap: 8px; background: var(--bg-input); border: 1px solid var(--border); border-radius: 24px; padding: 6px 16px; transition: border-color var(--transition); }
.search-box-inline:focus-within { border-color: var(--accent); }
.search-input { flex: 1; border: none; background: transparent; color: var(--text-primary); font-size: 13px; outline: none; }
.search-input::placeholder { color: var(--text-muted); }
.top-avatar { cursor: pointer; padding: 4px; border-radius: 50%; transition: background var(--transition); }
.top-avatar:active { background: var(--bg-card-hover); }

.top-bar-simple {
  position: sticky; top: 0; z-index: 100;
  display: flex; align-items: center; justify-content: space-between;
  padding: 10px 16px; backdrop-filter: blur(20px);
  background: rgba(10, 10, 15, 0.85);
  border-bottom: 1px solid var(--border);
}
.back-icon { cursor: pointer; padding: 4px; border-radius: 50%; transition: background var(--transition); }
.back-icon:active { background: var(--bg-card-hover); }
.top-title { font-size: 16px; font-weight: 700; }
</style>
