<template>
  <div class="page-container auth-page">
    <div class="auth-card card">
      <div class="auth-logo">✦ GoPan</div>
      <div class="auth-subtitle">创建账号</div>
      <van-form @submit="handleRegister">
        <van-field v-model="form.username" placeholder="用户名" :style="fieldStyle" :rules="[{ required: true }]" />
        <van-field v-model="form.password" type="password" placeholder="密码（至少6位）" :style="fieldStyle" :rules="[{ required: true }]" />
        <van-field v-model="form.email" placeholder="邮箱" :style="fieldStyle" :rules="[{ required: true, pattern: /^.+@.+$/ }]" />
        <button class="btn-primary" type="submit" :disabled="loading" style="width:100%;margin-top:20px">
          {{ loading ? "注册中..." : "注 册" }}
        </button>
      </van-form>
      <p class="auth-link">已有账号？<router-link to="/login">去登录</router-link></p>
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
const form = ref({ username: "", password: "", email: "" });
const fieldStyle = { '--van-field-background': '#1e1e32', '--van-field-input-text-color': '#e8e6f0', '--van-field-placeholder-text-color': '#5a5a7a' };

async function handleRegister() {
  loading.value = true;
  try { await auth.register(form.value.username, form.value.password, form.value.email); showToast("注册成功"); router.replace("/login"); }
  catch (e) { showToast(e.message || "注册失败"); }
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
