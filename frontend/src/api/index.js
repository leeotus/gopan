import request from "./request";

// 用户 API
export const userApi = {
  register: (data) => request.post("/user/register", data),
  login: (data) => request.post("/user/login", data),
  getProfile: () => request.get("/user/profile"),
  updateProfile: (data) => request.put("/user/profile", data),
};

// 视频 API
export const videoApi = {
  list: (params) => request.get("/video/list", { params }),
  detail: (params) => request.get("/video/detail", { params }),
  getPlayUrl: (params) => request.get("/video/play-url", { params }),
  initUpload: (data) => request.post("/video/init-upload", data),
  // 分片上传用 FormData
  uploadChunk: (formData) =>
    request.post("/video/upload-chunk", formData, {
      headers: { "Content-Type": "multipart/form-data" },
    }),
  mergeChunks: (data) => request.post("/video/merge-chunks", data),
};

// 互动 API
export const interactApi = {
  like: (params) => request.post("/video/like", null, { params }),
  unlike: (params) => request.delete("/video/like", { params }),
  favorite: (params) => request.post("/video/favorite", null, { params }),
  unfavorite: (params) => request.delete("/video/favorite", { params }),
  postComment: (data) => request.post("/video/comment", data),
  listComments: (params) => request.get("/video/comments", { params }),
  deleteComment: (params) => request.delete("/video/comment", { params }),
  sendDanmaku: (data) => request.post("/video/danmaku", data),
};

// 搜索 API
export const searchApi = {
  search: (params) => request.get("/search/videos", { params }),
};
