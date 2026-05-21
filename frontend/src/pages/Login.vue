<template>
  <div class="page-container auth-page">
    <div class="auth-card card">
      <div class="auth-logo">✦ GoPan</div>
      <div class="auth-subtitle">欢迎回来</div>
      <van-form @submit="handleLogin">
        <van-field v-model="form.username" placeholder="用户名" :style="fieldStyle" :rules="[{ required: true }]" />
        <van-field v-model="form.password" type="password" placeholder="密码" :style="fieldStyle" :rules="[{ required: true }]" />
        <button class="btn-primary" type="submit" :disabled="loading" style="width:100%;margin-top:20px">
          {{ loading ? "登录中..." : "登 录" }}
        </button>
      </van-form>
      <p class="auth-link">没有账号？<router-link to="/register">立即注册</router-link></p>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { showToast } from "vant";
import { useAuthStore } from "../stores/auth";

const router = useRouter();
const auth = useAuthStore();
const loading = ref(false);
const form = ref({ username: "", password: "" });
const fieldStyle = { '--van-field-background': '#1e1e32', '--van-field-input-text-color': '#e8e6f0', '--van-field-placeholder-text-color': '#5a5a7a' };

async function handleLogin() {
  loading.value = true;
  try { await auth.login(form.value.username, form.value.password); showToast("登录成功"); router.replace("/"); }
  catch (e) { showToast(e.message || "登录失败"); }
  finally { loading.value = false; }
}
</script>

<style scoped>
.auth-page { display: flex; align-items: center; justify-content: center; min-height: 100vh; padding: 20px; }
.auth-card { padding: 36px 28px; width: 100%; max-width: 360px; text-align: center; }
.auth-logo { font-size: 32px; font-weight: 900; background: linear-gradient(135deg, var(--accent), #c084fc); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
.auth-subtitle { color: var(--text-secondary); font-size: 14px; margin: 8px 0 24px; }
.auth-link { margin-top: 20px; font-size: 13px; color: var(--text-muted); }
.auth-link a { color: var(--accent); font-weight: 500; }
</style>
