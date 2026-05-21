export function formatDuration(seconds) {
  if (!seconds || seconds < 0) return "00:00";
  const h = Math.floor(seconds / 3600), m = Math.floor((seconds % 3600) / 60), s = seconds % 60;
  const pad = (n) => String(n).padStart(2, "0");
  return h > 0 ? `${pad(h)}:${pad(m)}:${pad(s)}` : `${pad(m)}:${pad(s)}`;
}

export function formatCount(num) {
  if (!num) return "0";
  return num >= 10000 ? (num / 10000).toFixed(1) + "w" : String(num);
}

export function formatTime(timestamp) {
  const now = Date.now() / 1000, diff = now - timestamp;
  if (diff < 60) return "刚刚";
  if (diff < 3600) return Math.floor(diff / 60) + "分钟前";
  if (diff < 86400) return Math.floor(diff / 3600) + "小时前";
  if (diff < 2592000) return Math.floor(diff / 86400) + "天前";
  return new Date(timestamp * 1000).toLocaleDateString("zh-CN");
}
