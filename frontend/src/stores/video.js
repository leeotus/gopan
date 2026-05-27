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

// 不再使用 mock 数据，全部从后端获取
const mockVideos = [];
