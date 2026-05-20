<template>
  <div id="app-root">
    <router-view v-slot="{ Component }">
      <transition name="fade" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>

    <!-- 底部导航栏 -->
    <van-tabbar v-if="showTabbar" v-model="activeTab" route>
      <van-tabbar-item icon="home-o" to="/">首页</van-tabbar-item>
      <van-tabbar-item icon="search" to="/search">搜索</van-tabbar-item>
      <van-tabbar-item icon="plus" to="/upload">上传</van-tabbar-item>
      <van-tabbar-item icon="user-o" to="/profile">我的</van-tabbar-item>
    </van-tabbar>
  </div>
</template>

<script setup>
import { computed } from "vue";
import { useRoute } from "vue-router";

const route = useRoute();

// 只在首页、搜索、上传、个人中心显示底部导航
const showTabbar = computed(() => {
  const paths = ["/", "/search", "/upload", "/profile"];
  return paths.includes(route.path);
});

const activeTab = computed(() => {
  const map = { "/": 0, "/search": 1, "/upload": 2, "/profile": 3 };
  return map[route.path] ?? 0;
});
</script>

<style>
#app-root {
  min-height: 100vh;
  background: var(--gopan-bg);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
