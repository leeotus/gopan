// 格式化时长（秒 → mm:ss 或 hh:mm:ss）
export function formatDuration(seconds) {
  if (!seconds || seconds < 0) return "00:00";
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = seconds % 60;
  const pad = (n) => String(n).padStart(2, "0");
  if (h > 0) return `${pad(h)}:${pad(m)}:${pad(s)}`;
  return `${pad(m)}:${pad(s)}`;
}

// 格式化播放数（超过1万显示为 x.xw）
export function formatCount(num) {
  if (!num) return "0";
  if (num >= 10000) return (num / 10000).toFixed(1) + "w";
  return String(num);
}

// 格式化时间戳为相对时间
export function formatTime(timestamp) {
  const now = Date.now() / 1000;
  const diff = now - timestamp;
  if (diff < 60) return "刚刚";
  if (diff < 3600) return Math.floor(diff / 60) + "分钟前";
  if (diff < 86400) return Math.floor(diff / 3600) + "小时前";
  if (diff < 2592000) return Math.floor(diff / 86400) + "天前";
  return new Date(timestamp * 1000).toLocaleDateString("zh-CN");
}
