import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { userApi } from "../api";

export const useAuthStore = defineStore("auth", () => {
  const token = ref(localStorage.getItem("token") || "");
  const user = ref(JSON.parse(localStorage.getItem("user") || "null"));
  const isLoggedIn = computed(() => !!token.value);

  async function login(username, password) {
    const res = await userApi.login({ username, password });
    token.value = res.token;
    user.value = { userId: res.user_id, username: res.username, avatar: res.avatar };
    localStorage.setItem("token", res.token);
    localStorage.setItem("user", JSON.stringify(user.value));
    return res;
  }

  async function register(username, password, email) {
    return await userApi.register({ username, password, email });
  }

  function logout() {
    token.value = "";
    user.value = null;
    localStorage.removeItem("token");
    localStorage.removeItem("user");
  }

  async function fetchProfile() {
    try {
      const res = await userApi.getProfile();
      user.value = res.data;
      localStorage.setItem("user", JSON.stringify(res.data));
    } catch { /* ignore */ }
  }

  return { token, user, isLoggedIn, login, register, logout, fetchProfile };
});
