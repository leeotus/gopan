<template>
  <div class="page-container">
    <van-nav-bar title="登录" left-arrow @click-left="$router.back()" />

    <div class="form-wrapper">
      <van-form @submit="handleLogin">
        <van-cell-group inset>
          <van-field
            v-model="form.username"
            name="username"
            label="用户名"
            placeholder="请输入用户名"
            :rules="[{ required: true, message: '请输入用户名' }]"
          />
          <van-field
            v-model="form.password"
            type="password"
            name="password"
            label="密码"
            placeholder="请输入密码"
            :rules="[{ required: true, message: '请输入密码' }]"
          />
        </van-cell-group>

        <div style="margin: 16px">
          <van-button round block type="primary" native-type="submit" :loading="loading">
            登录
          </van-button>
        </div>

        <div class="link-row">
          <router-link to="/register">没有账号？立即注册</router-link>
        </div>
      </van-form>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { showToast } from "vant";
import { useAuthStore } from "../stores/auth";

const router = useRouter();
const authStore = useAuthStore();
const loading = ref(false);

const form = ref({
  username: "",
  password: "",
});

async function handleLogin() {
  loading.value = true;
  try {
    await authStore.login(form.value.username, form.value.password);
    showToast("登录成功");
    router.replace("/");
  } catch (e) {
    showToast(e.message || "登录失败");
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.form-wrapper {
  padding-top: 40px;
}

.link-row {
  text-align: center;
  font-size: 14px;
}

.link-row a {
  color: var(--gopan-primary);
}
</style>
