import axios from "axios";

const request = axios.create({
  baseURL: "/api",
  timeout: 8000,
});

request.interceptors.request.use((config) => {
  const token = localStorage.getItem("token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

request.interceptors.response.use(
  (res) => {
    const data = res.data;
    if (data.code !== 0) {
      return Promise.reject(new Error(data.message || "请求失败"));
    }
    return data;
  },
  (err) => {
    return Promise.reject(err);
  }
);

export default request;
