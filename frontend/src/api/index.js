import request from "./request";
import axios from "axios";

// direct 直连后端绕过 Vite 代理（multipart 请求必须直连，代理会截断 body）
const direct = axios.create({ timeout: 60000 });

export const userApi = {
  register: (data) => request.post("/user/register", data),
  login: (data) => request.post("/user/login", data),
  getProfile: () => request.get("/user/profile"),
  updateProfile: (data) => request.put("/user/profile", data),
};

export const videoApi = {
  list: (params) => request.get("/video/list", { params }),
  detail: (params) => request.get("/video/detail", { params }),
  getPlayUrl: (params) => request.get("/video/play-url", { params }),
  initUpload: (data) => request.post("/video/init-upload", data),
  uploadChunk: (formData) => direct.post("http://localhost:8888/api/video/upload-chunk", formData),
  uploadChunkRaw: (data) => request.post("/video/upload-chunk", data),  // base64 JSON 方式
  uploadStatus: (params) => request.get("/video/upload-status", { params }),
  mergeChunks: (data) => request.post("/video/merge-chunks", data),
  update: (data) => request.put("/video/update", data),
  delete: (params) => request.delete("/video/delete", { params }),
  like: () => request.post("/video/like"),
  unlike: () => request.delete("/video/like"),
  favorite: () => request.post("/video/favorite"),
  unfavorite: () => request.delete("/video/favorite"),
  postComment: (data) => request.post("/video/comment", data),
  listComments: (params) => request.get("/video/comments", { params }),
  deleteComment: (params) => request.delete("/video/comment", { params }),
  sendDanmaku: (data) => request.post("/video/danmaku", data),
  getDanmakus: (params) => request.get("/video/danmakus", { params }),
};

export const searchApi = {
  search: (params) => request.get("/search/videos", { params }),
};
