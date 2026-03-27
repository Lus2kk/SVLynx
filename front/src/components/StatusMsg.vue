<template>
  <div v-if="type" :class="['status', type]">
    <div class="status-dot"></div>
    <span>{{ message }}</span>
  </div>
</template>

<script>
export default {
  props: {
    type: String,
    message: String
  }
}
</script>

<style scoped>
.status {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
  border-radius: 12px;
  font-size: 13.5px;
  margin-bottom: 16px;
}
.status.success {
  background: rgba(52,211,153,0.1);
  border: 1px solid rgba(52,211,153,0.2);
  color: #34d399;
}
.status.error {
  background: rgba(248,113,113,0.1);
  border: 1px solid rgba(248,113,113,0.2);
  color: #f87171;
}
.status-dot {
  width: 8px; height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
  animation: pulse 1.5s ease-in-out infinite;
}
.status.success .status-dot { background: #34d399; }
.status.error .status-dot   { background: #f87171; }

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50%       { opacity: 0.5; transform: scale(0.8); }
}
</style>