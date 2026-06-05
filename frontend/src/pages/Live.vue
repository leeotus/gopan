<template>
  <div class="page-container flex-center" style="min-height:70vh;flex-direction:column">
    <div class="empty-state">
      <div class="live-icon">📡</div>
      <h1 class="live-title">Live Streaming</h1>
      <p class="live-subtitle">Coming soon</p>
      <p class="live-hint">We're building something amazing.<br/>Live streaming will be available in the next update.</p>
    </div>
  </div>
</template>

<style scoped>
.empty-state { text-align: center; padding: 60px 20px; }
.live-icon { font-size: 64px; margin-bottom: 20px; animation: pulse-glow 2s var(--ease-spring) infinite; }
.live-title {
  font-family: var(--font-display);
  font-size: 24px;
  font-weight: 900;
  letter-spacing: 3px;
  background: linear-gradient(135deg, var(--pink), var(--purple));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  margin-bottom: 8px;
}
.live-subtitle { font-size: 14px; color: var(--text-muted); margin-bottom: 24px; }
.live-hint { font-size: 13px; color: var(--text-secondary); line-height: 1.8; }
@keyframes pulse-glow { 0%,100% { opacity: 0.5; transform: scale(1); } 50% { opacity: 1; transform: scale(1.05); } }
</style>
