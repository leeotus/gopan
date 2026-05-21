import axios from "axios";

const request = axios.create({ baseURL: "/api", timeout: 10000 });

request.interceptors.request.use((config) => {
  const token = localStorage.getItem("token");
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

request.interceptors.response.use(
  (res) => {
    const body = res.data;
    // 网关未统一包装 code/message，业务数据直接返回
    // 有 code 字段时才做判断，否则直接通过
    if (body && typeof body.code === "number" && body.code !== 0) {
      return Promise.reject(new Error(body.message || "请求失败"));
    }
    return body;
  },
  (err) => Promise.reject(err)
);

export default request;
