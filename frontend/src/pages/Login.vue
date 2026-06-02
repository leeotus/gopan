<template>
  <div class="page-container flex-center" style="min-height:100vh">
    <div class="auth-card anim-fade-up">
      <div class="auth-icon">⬡</div>
      <h1 class="auth-title">GoPan</h1>
      <p class="auth-subtitle">{{ isLogin ? 'Welcome back' : 'Create your account' }}</p>

      <form @submit.prevent="handleSubmit" class="auth-form">
        <input class="input-field" v-model="form.username" placeholder="Username" autocomplete="username" required />
        <input class="input-field" v-model="form.password" type="password" placeholder="Password" autocomplete="current-password" required />
        <input class="input-field" v-if="!isLogin" v-model="form.email" type="email" placeholder="Email" autocomplete="email" />

        <button class="btn-primary btn-primary--solid" type="submit" style="width:100%">
          {{ isLogin ? 'Sign In' : 'Create Account' }}
        </button>
      </form>

      <p class="auth-switch" v-if="isLogin">
        Don't have an account? <router-link to="/register" class="text-cyan">Sign up</router-link>
      </p>
      <p class="auth-switch" v-else>
        Already have an account? <router-link to="/login" class="text-cyan">Sign in</router-link>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { showToast } from "vant";
import { userApi } from "../api";
import { useAuthStore } from "../stores/auth";

const props = defineProps({ isLogin: { type: Boolean, default: true } });
const router = useRouter();
const auth = useAuthStore();
const form = ref({ username: "", password: "", email: "" });

async function handleSubmit() {
  try {
    if (props.isLogin) {
      const res = await userApi.login({ username: form.value.username, password: form.value.password });
      localStorage.setItem("token", res.token);
      auth.token = res.token;
      router.push("/");
    } else {
      await userApi.register({ username: form.value.username, password: form.value.password, email: form.value.email });
      showToast("Account created! Please sign in.");
      router.push("/login");
    }
  } catch (e) {
    showToast(e?.response?.data?.message || e.message || "Error");
  }
}
</script>

<style scoped>
.auth-card {
  width: 100%;
  max-width: 360px;
  padding: 40px 28px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl);
  text-align: center;
}
.auth-icon {
  font-size: 36px;
  color: var(--cyan);
  filter: drop-shadow(0 0 12px var(--cyan-glow));
  margin-bottom: 12px;
}
.auth-title {
  font-family: var(--font-display);
  font-size: 24px;
  font-weight: 900;
  letter-spacing: 3px;
  background: linear-gradient(135deg, var(--cyan), var(--purple));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  margin-bottom: 4px;
}
.auth-subtitle { font-size: 13px; color: var(--text-muted); margin-bottom: 28px; }
.auth-form { display: flex; flex-direction: column; gap: 12px; }
.auth-switch { margin-top: 20px; font-size: 13px; color: var(--text-muted); }
</style>
