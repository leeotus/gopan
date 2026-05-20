import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { userApi } from "../api";

export const useAuthStore = defineStore("auth", () => {
  const token = ref(localStorage.getItem("token") || "");
  const user = ref(JSON.parse(localStorage.getItem("user") || "null"));

  const isLoggedIn = computed(() => !!token.value);

  async function login(username, password) {
    const res = await userApi.login({ username, password });
    token.value = res.data.token;
    user.value = {
      userId: res.data.user_id,
      username: res.data.username,
      avatar: res.data.avatar,
    };
    localStorage.setItem("token", res.data.token);
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
      user.value = {
        userId: res.data.user_id,
        username: res.data.username,
        email: res.data.email,
        avatar: res.data.avatar,
        signature: res.data.signature,
      };
      localStorage.setItem("user", JSON.stringify(user.value));
    } catch {
      // ignore
    }
  }

  return { token, user, isLoggedIn, login, register, logout, fetchProfile };
});
