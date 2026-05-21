import { createRouter, createWebHistory } from "vue-router";

const routes = [
  { path: "/", name: "Home", component: () => import("../pages/Home.vue"), meta: { title: "首页" } },
  { path: "/login", name: "Login", component: () => import("../pages/Login.vue"), meta: { title: "登录" } },
  { path: "/register", name: "Register", component: () => import("../pages/Register.vue"), meta: { title: "注册" } },
  { path: "/video/:id", name: "VideoDetail", component: () => import("../pages/VideoDetail.vue"), meta: { title: "视频详情" } },
  { path: "/search", name: "Search", component: () => import("../pages/Search.vue"), meta: { title: "搜索" } },
  { path: "/profile", name: "Profile", component: () => import("../pages/Profile.vue"), meta: { title: "我的" } },
  { path: "/upload", name: "Upload", component: () => import("../pages/Upload.vue"), meta: { title: "上传", auth: true } },
];

const router = createRouter({ history: createWebHistory(), routes });

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem("token");
  if (to.meta.auth && !token) next("/login");
  else next();
});

export default router;
