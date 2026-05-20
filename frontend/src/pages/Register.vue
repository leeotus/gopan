<template>
  <div class="page-container">
    <van-nav-bar title="注册" left-arrow @click-left="$router.back()" />

    <div class="form-wrapper">
      <van-form @submit="handleRegister">
        <van-cell-group inset>
          <van-field
            v-model="form.username"
            name="username"
            label="用户名"
            placeholder="4-20位字母数字"
            :rules="[{ required: true, message: '请输入用户名' }]"
          />
          <van-field
            v-model="form.password"
            type="password"
            name="password"
            label="密码"
            placeholder="至少6位"
            :rules="[{ required: true, message: '请输入密码' }]"
          />
          <van-field
            v-model="form.email"
            name="email"
            label="邮箱"
            placeholder="请输入邮箱"
            :rules="[{ required: true, pattern: /^.+@.+$/, message: '邮箱格式不正确' }]"
          />
        </van-cell-group>

        <div style="margin: 16px">
          <van-button round block type="primary" native-type="submit" :loading="loading">
            注册
          </van-button>
        </div>

        <div class="link-row">
          <router-link to="/login">已有账号？去登录</router-link>
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
  email: "",
});

async function handleRegister() {
  loading.value = true;
  try {
    await authStore.register(form.value.username, form.value.password, form.value.email);
    showToast("注册成功，请登录");
    router.replace("/login");
  } catch (e) {
    showToast(e.message || "注册失败");
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
