import { defineStore } from "pinia";
import { ref } from "vue";
import { videoApi } from "../api";

export const useVideoStore = defineStore("video", () => {
  const videos = ref([]);
  const myVideos = ref([]);
  const currentVideo = ref(null);
  const hasMore = ref(true);
  const nextCursor = ref(0);
  const loading = ref(false);

  async function fetchVideos(params = {}) {
    if (loading.value) return;
    loading.value = true;
    try {
      const res = await videoApi.list(params);
      if (params.cursor === 0) videos.value = [];
      videos.value.push(...(res.data.videos || []));
      hasMore.value = res.data.has_more;
      nextCursor.value = res.data.next_cursor;
    } catch (e) {
      if (videos.value.length === 0) videos.value = mockVideos;
    } finally {
      loading.value = false;
    }
  }

  async function fetchMyVideos(params = {}) {
    try {
      const res = await videoApi.list(params);
      myVideos.value = res.data?.videos || [];
    } catch {
      myVideos.value = mockVideos.slice(0, 3);
    }
  }

  async function fetchDetail(videoId) {
    try {
      const res = await videoApi.detail({ video_id: videoId });
      currentVideo.value = res.data.video;
    } catch {
      currentVideo.value = mockVideos.find((v) => v.id === videoId);
    }
  }

  async function toggleLike(videoId, liked) {
    try {
      liked ? await videoApi.unlike() : await videoApi.like();
      const v = videos.value.find((v) => v.id === videoId) || currentVideo.value;
      if (v) { v.liked = !liked; v.like_count += liked ? -1 : 1; }
    } catch { /* ignore */ }
  }

  async function toggleFavorite(videoId) {
    try {
      await videoApi.favorite();
    } catch { /* ignore */ }
  }

  return { videos, myVideos, currentVideo, hasMore, nextCursor, loading, fetchVideos, fetchMyVideos, fetchDetail, toggleLike, toggleFavorite };
});

const mockVideos = [
  { id: 1, title: "Go微服务实战 - 第一课", cover_url: "https://picsum.photos/seed/v1/360/200", user_id: 1, username: "讲师A", play_count: 12500, like_count: 834, duration: 1847, status: 2, category: "技术", created_at: 1716153600, liked: false, favorited: false, transcodes: [{ resolution: "1080p", bitrate: 5000 }] },
  { id: 2, title: "Vue3 组合式API详解", cover_url: "https://picsum.photos/seed/v2/360/200", user_id: 2, username: "前端达人", play_count: 8900, like_count: 621, duration: 2230, status: 2, category: "前端", created_at: 1716067200, liked: true, favorited: false, transcodes: [{ resolution: "1080p", bitrate: 5000 }] },
  { id: 3, title: "Docker & K8s 入门", cover_url: "https://picsum.photos/seed/v3/360/200", user_id: 1, username: "讲师A", play_count: 34500, like_count: 2100, duration: 3200, status: 2, category: "技术", created_at: 1715980800, liked: false, favorited: true, transcodes: [{ resolution: "1080p", bitrate: 5000 }] },
  { id: 4, title: "Python 数据分析实战", cover_url: "https://picsum.photos/seed/v4/360/200", user_id: 3, username: "数据科学家", play_count: 6700, like_count: 445, duration: 2800, status: 2, category: "数据", created_at: 1715894400, liked: false, favorited: false, transcodes: [{ resolution: "1080p", bitrate: 5000 }] },
  { id: 5, title: "TCP/IP 协议详解", cover_url: "https://picsum.photos/seed/v5/360/200", user_id: 4, username: "网络专家", play_count: 15400, like_count: 987, duration: 4500, status: 2, category: "基础", created_at: 1715808000, liked: false, favorited: false, transcodes: [{ resolution: "1080p", bitrate: 5000 }] },
  { id: 6, title: "Rust 系统编程入门", cover_url: "https://picsum.photos/seed/v6/360/200", user_id: 5, username: "Rustacean", play_count: 4200, like_count: 310, duration: 3600, status: 2, category: "技术", created_at: 1715721600, liked: true, favorited: true, transcodes: [{ resolution: "1080p", bitrate: 5000 }] },
];
