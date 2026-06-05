<template>
  <div id="app-root">
    <!-- Scan line overlay -->
    <div class="scan-line"></div>

    <!-- Top bar -->
    <header class="top-bar" v-if="showTopBar">
      <div class="top-logo" @click="$router.push('/')">
        <span class="logo-icon">⬡</span>
        <span class="logo-text">GoPan</span>
      </div>

      <div class="search-box" v-if="showSearchBar">
        <svg class="search-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
        <input
          v-model="searchKeyword"
          class="search-input"
          placeholder="Search videos..."
          @keyup.enter="doSearch"
        />
      </div>

      <div class="top-actions">
        <button class="live-tag" @click="$router.push('/live')">
          <span class="live-dot"></span>
          LIVE
        </button>
        <button class="top-action-btn" @click="$router.push('/upload')" title="Upload">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
        </button>
        <div class="avatar-circle" @click="$router.push('/profile')">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
        </div>
      </div>
    </header>

    <!-- Simple back bar -->
    <header class="top-bar-simple" v-else>
      <button class="back-btn" @click="$router.back()">
        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="m15 18-6-6 6-6"/></svg>
      </button>
      <span class="top-title">{{ pageTitle }}</span>
      <div style="width:22px"></div>
    </header>

    <!-- Page content -->
    <router-view v-slot="{ Component }">
      <transition name="page" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>

    <!-- Bottom nav -->
    <nav class="bottom-nav" v-if="showTabbar">
      <router-link to="/" class="nav-item" :class="{ active: $route.path === '/' }">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>
        <span>Home</span>
      </router-link>
      <router-link to="/upload" class="nav-item nav-item--center" :class="{ active: $route.path === '/upload' }">
        <div class="nav-upload-btn">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
        </div>
      </router-link>
      <router-link to="/profile" class="nav-item" :class="{ active: $route.path === '/profile' }">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
        <span>Me</span>
      </router-link>
    </nav>
  </div>
</template>

<script setup>
import { computed, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

const route = useRoute();
const router = useRouter();
const searchKeyword = ref("");

const pageTitles = {
  "/": "Home", "/login": "Sign In", "/register": "Sign Up",
  "/search": "Search", "/profile": "Profile", "/upload": "Upload",
  "/admin": "Admin",
};
const pageTitle = computed(() => pageTitles[route.path] || "");

const showTabbar = computed(() => ["/", "/profile"].includes(route.path));
const showTopBar = computed(() => !["/login", "/register"].includes(route.path));
const showSearchBar = computed(() => ["/", "/search"].includes(route.path));

// 监听路由 Q参数，自动同步输入框检索词
watch(() => route.query.q, (newQ) => {
  searchKeyword.value = newQ || "";
}, { immediate: true });

function doSearch() {
  const kw = searchKeyword.value.trim();
  if (kw) {
    router.push({ path: "/search", query: { q: kw } });
  } else if (route.path === "/search") {
    // 搜索词清空时，自动返回不带参的搜索首页
    router.push({ path: "/search" });
  }
}
</script>

<style scoped>
#app-root { min-height: 100vh; position: relative; background: var(--bg-deep); }

/* ── Scan line ── */
.scan-line {
  position: fixed;
  inset: 0;
  z-index: 10000;
  pointer-events: none;
  background: linear-gradient(to bottom, transparent 50%, rgba(0,240,255,0.015) 50%);
  background-size: 100% 4px;
  animation: scan-line 8s linear infinite;
}

/* ── Top bar ── */
.top-bar {
  position: sticky;
  top: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  background: rgba(10,10,20,0.85);
  backdrop-filter: blur(20px);
  border-bottom: 1px solid var(--border);
}
.top-logo {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  flex-shrink: 0;
}
.logo-icon {
  font-size: 22px;
  color: var(--cyan);
  filter: drop-shadow(0 0 6px var(--cyan-glow));
}
.logo-text {
  font-family: var(--font-display);
  font-size: 16px;
  font-weight: 700;
  letter-spacing: 2px;
  background: linear-gradient(135deg, var(--cyan), var(--purple));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}
.search-box {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 14px;
  background: var(--bg-input);
  border: 1px solid var(--border);
  border-radius: 24px;
  transition: border-color var(--duration) var(--ease-out), box-shadow var(--duration) var(--ease-out);
}
.search-box:focus-within {
  border-color: var(--cyan-dim);
  box-shadow: 0 0 12px var(--cyan-glow);
}
.search-icon { color: var(--text-muted); flex-shrink: 0; }
.search-input {
  flex: 1;
  border: none;
  background: transparent;
  color: var(--text-primary);
  font-family: var(--font-body);
  font-size: 13px;
  outline: none;
}
.search-input::placeholder { color: var(--text-muted); }
.top-actions { display: flex; align-items: center; gap: 10px; flex-shrink: 0; }
.top-action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: 1px solid var(--border);
  border-radius: 50%;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all var(--duration) var(--ease-out);
}
.top-action-btn:hover { border-color: var(--cyan-dim); color: var(--cyan); }
.live-tag {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 5px 12px;
  border: 1px solid rgba(255,45,149,0.4);
  border-radius: 20px;
  background: rgba(255,45,149,0.08);
  color: var(--pink);
  font-family: var(--font-display);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 1px;
  cursor: pointer;
  transition: all var(--duration) var(--ease-out);
  animation: pulse-glow-live 2s infinite;
}
.live-tag:hover {
  background: rgba(255,45,149,0.18);
  box-shadow: 0 0 16px var(--pink-glow);
}
.live-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--pink);
  box-shadow: 0 0 6px var(--pink-glow);
}
@keyframes pulse-glow-live {
  0%, 100% { box-shadow: 0 0 6px var(--pink-glow); }
  50%      { box-shadow: 0 0 16px var(--pink-glow); }
}
.avatar-circle {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all var(--duration) var(--ease-out);
}
.avatar-circle:hover { border-color: var(--purple-dim); color: var(--purple); }

/* Back bar */
.top-bar-simple {
  position: sticky;
  top: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  background: rgba(10,10,20,0.85);
  backdrop-filter: blur(20px);
  border-bottom: 1px solid var(--border);
}
.back-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  background: none;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 4px;
  border-radius: 6px;
  transition: color var(--duration);
}
.back-btn:hover { color: var(--cyan); }
.top-title {
  font-family: var(--font-display);
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 2px;
  color: var(--text-secondary);
  text-transform: uppercase;
}

/* ── Bottom nav ── */
.bottom-nav {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 100;
  display: flex;
  justify-content: space-around;
  align-items: center;
  padding: 6px 0 10px;
  background: rgba(10,10,20,0.9);
  backdrop-filter: blur(20px);
  border-top: 1px solid var(--border);
}
.nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 4px 10px;
  color: var(--text-muted);
  font-size: 10px;
  font-weight: 500;
  transition: color var(--duration);
  text-decoration: none;
}
.nav-item.active { color: var(--cyan); }
.nav-item svg { transition: filter var(--duration); }
.nav-item.active svg { filter: drop-shadow(0 0 4px var(--cyan-glow)); }
.nav-item--center { margin-top: -16px; }
.nav-upload-btn {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--cyan), var(--purple));
  display: flex;
  align-items: center;
  justify-content: center;
  color: #000;
  box-shadow: 0 4px 20px var(--cyan-glow-intense);
  transition: transform var(--duration) var(--ease-spring);
}
.nav-item--center.active .nav-upload-btn { transform: scale(1.1); }

/* ── Page transitions ── */
.page-enter-active { animation: fadeInUp 0.35s var(--ease-out) both; }
.page-leave-active { animation: fadeIn 0.18s var(--ease-out) both reverse; }
</style>
