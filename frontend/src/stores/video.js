import { defineStore } from "pinia";
import { ref } from "vue";
import { videoApi, interactApi, searchApi } from "../api";

export const useVideoStore = defineStore("video", () => {
  const videos = ref([]);
  const currentVideo = ref(null);
  const hasMore = ref(true);
  const nextCursor = ref(0);
  const loading = ref(false);

  // 列表
  async function fetchVideos(params = {}) {
    if (loading.value) return;
    loading.value = true;
    try {
      const res = await videoApi.list(params);
      const data = res.data;
      if (params.cursor === 0) videos.value = [];
      videos.value.push(...(data.videos || []));
      hasMore.value = data.has_more;
      nextCursor.value = data.next_cursor;
    } catch {
      // mock 兜底
      if (videos.value.length === 0) {
        videos.value = mockVideos;
        hasMore.value = false;
      }
    } finally {
      loading.value = false;
    }
  }

  // 详情
  async function fetchDetail(videoId) {
    try {
      const res = await videoApi.detail({ video_id: videoId });
      currentVideo.value = res.data.video;
    } catch {
      currentVideo.value = mockVideos.find((v) => v.id === videoId);
    }
  }

  // 搜索
  async function search(keyword, page = 1) {
    try {
      return await searchApi.search({ keyword, page, size: 20 });
    } catch {
      return { data: { videos: [], total: 0 } };
    }
  }

  // 点赞
  async function toggleLike(videoId, liked) {
    try {
      if (liked) {
        await interactApi.unlike({ video_id: videoId });
      } else {
        await interactApi.like({ video_id: videoId });
      }
      // 本地更新
      const v = videos.value.find((v) => v.id === videoId);
      if (v) {
        v.liked = !liked;
        v.like_count += liked ? -1 : 1;
      }
    } catch {
      // ignore
    }
  }

  // 收藏
  async function toggleFavorite(videoId, favorited) {
    try {
      if (favorited) {
        await interactApi.unfavorite({ video_id: videoId });
      } else {
        await interactApi.favorite({ video_id: videoId });
      }
    } catch {
      // ignore
    }
  }

  return {
    videos,
    currentVideo,
    hasMore,
    nextCursor,
    loading,
    fetchVideos,
    fetchDetail,
    search,
    toggleLike,
    toggleFavorite,
  };
});

// mock 数据 —— 后端未实现封面截取时的占位数据
const mockVideos = [
  {
    id: 1,
    title: "Go微服务实战 - 第一课",
    cover_url: "https://picsum.photos/seed/video1/360/200",
    description: "从零开始搭建微服务架构",
    user_id: 1,
    username: "讲师A",
    play_count: 12500,
    like_count: 834,
    duration: 1847,
    status: 2,
    category: "技术",
    created_at: 1716153600,
    liked: false,
    favorited: false,
    transcodes: [
      { resolution: "360p", bitrate: 500 },
      { resolution: "720p", bitrate: 2500 },
    ],
  },
  {
    id: 2,
    title: "Vue3 组件化开发指南",
    cover_url: "https://picsum.photos/seed/video2/360/200",
    description: "深入理解 Vue3 组合式 API",
    user_id: 2,
    username: "前端达人",
    play_count: 8900,
    like_count: 621,
    duration: 2230,
    status: 2,
    category: "前端",
    created_at: 1716067200,
    liked: true,
    favorited: false,
    transcodes: [
      { resolution: "480p", bitrate: 1000 },
      { resolution: "1080p", bitrate: 5000 },
    ],
  },
  {
    id: 3,
    title: "Docker & K8s 入门到精通",
    cover_url: "https://picsum.photos/seed/video3/360/200",
    description: "容器编排与云原生最佳实践",
    user_id: 1,
    username: "讲师A",
    play_count: 34500,
    like_count: 2100,
    duration: 3200,
    status: 2,
    category: "技术",
    created_at: 1715980800,
    liked: false,
    favorited: true,
    transcodes: [
      { resolution: "360p", bitrate: 500 },
      { resolution: "720p", bitrate: 2500 },
      { resolution: "1080p", bitrate: 5000 },
    ],
  },
  {
    id: 4,
    title: "Python 数据分析实战",
    cover_url: "https://picsum.photos/seed/video4/360/200",
    description: "Pandas + Matplotlib 实战案例",
    user_id: 3,
    username: "数据科学家",
    play_count: 6700,
    like_count: 445,
    duration: 2800,
    status: 2,
    category: "数据",
    created_at: 1715894400,
    liked: false,
    favorited: false,
    transcodes: [
      { resolution: "480p", bitrate: 1000 },
      { resolution: "720p", bitrate: 2500 },
    ],
  },
  {
    id: 5,
    title: "计算机网络 TCP/IP 详解",
    cover_url: "https://picsum.photos/seed/video5/360/200",
    description: "从物理层到应用层完整梳理",
    user_id: 4,
    username: "网络专家",
    play_count: 15400,
    like_count: 987,
    duration: 4500,
    status: 2,
    category: "基础",
    created_at: 1715808000,
    liked: false,
    favorited: false,
    transcodes: [
      { resolution: "360p", bitrate: 500 },
      { resolution: "480p", bitrate: 1000 },
    ],
  },
  {
    id: 6,
    title: "Rust 系统编程入门",
    cover_url: "https://picsum.photos/seed/video6/360/200",
    description: "所有权、生命周期、并发编程",
    user_id: 5,
    username: "Rustacean",
    play_count: 4200,
    like_count: 310,
    duration: 3600,
    status: 2,
    category: "技术",
    created_at: 1715721600,
    liked: true,
    favorited: true,
    transcodes: [
      { resolution: "720p", bitrate: 2500 },
      { resolution: "1080p", bitrate: 5000 },
    ],
  },
];
