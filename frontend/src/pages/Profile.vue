<template>
  <div class="page-container">
    <van-nav-bar title="个人中心" />

    <div class="page-content">
      <!-- 未登录 -->
      <div v-if="!authStore.isLoggedIn" class="login-card">
        <van-icon name="user-circle-o" size="60" color="#ccc" />
        <p>登录后享受更多功能</p>
        <van-button type="primary" round block to="/login">登录 / 注册</van-button>
      </div>

      <!-- 已登录 -->
      <template v-else>
        <div class="user-card">
          <img :src="authStore.user?.avatar || defaultAvatar" class="user-avatar" />
          <div class="user-info">
            <div class="user-name">{{ authStore.user?.username }}</div>
            <div class="user-signature">{{ authStore.user?.signature || "这个人很懒，什么都没写" }}</div>
          </div>
          <van-icon name="edit" size="18" color="#999" />
        </div>

        <van-cell-group inset style="margin-top: 12px">
          <van-cell title="我的视频" icon="video-o" is-link to="/" />
          <van-cell title="我的收藏" icon="star-o" is-link />
          <van-cell title="播放历史" icon="clock-o" is-link />
        </van-cell-group>

        <van-cell-group inset style="margin-top: 12px">
          <van-cell title="设置" icon="setting-o" is-link />
          <van-cell title="关于" icon="info-o" is-link value="v1.0.0" />
        </van-cell-group>

        <div style="margin: 24px 16px">
          <van-button block round type="danger" @click="handleLogout">退出登录</van-button>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from "vue-router";
import { showToast } from "vant";
import { useAuthStore } from "../stores/auth";

const router = useRouter();
const authStore = useAuthStore();

const defaultAvatar = "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'%3E%3Ccircle cx='50' cy='50' r='50' fill='%23ddd'/%3E%3Ccircle cx='50' cy='40' r='16' fill='%23aaa'/%3E%3Cellipse cx='50' cy='78' rx='30' ry='18' fill='%23aaa'/%3E%3C/svg%3E";

function handleLogout() {
  authStore.logout();
  showToast("已退出");
  router.push("/");
}
</script>

<style scoped>
.login-card {
  text-align: center;
  padding: 60px 20px;
  background: #fff;
  border-radius: 8px;
}

.login-card p {
  margin: 12px 0 20px;
  color: var(--gopan-text-secondary);
  font-size: 14px;
}

.user-card {
  display: flex;
  align-items: center;
  gap: 12px;
  background: #fff;
  padding: 20px 16px;
  border-radius: 8px;
}

.user-avatar {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: #eee;
}

.user-info {
  flex: 1;
}

.user-name {
  font-size: 17px;
  font-weight: 600;
}

.user-signature {
  font-size: 13px;
  color: var(--gopan-text-secondary);
  margin-top: 4px;
}
</style>
