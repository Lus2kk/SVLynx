<template>
  <div
    class="message-row"
    :class="{ mine: isMine, theirs: !isMine, 'theme-light': isLight }"
    @mouseenter="showActions = true"
    @mouseleave="showActions = false"
  >
    <div class="message-bubble-wrapper">
      <button
        v-if="isMine"
        class="message-action delete-btn"
        :class="{ visible: showActions }"
        @click="$emit('delete', message.id)"
        title="Delete message"
        type="button"
      >
        <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="1.8">
          <path d="M3 6h18"></path>
          <path d="M8 6V4.8c0-.99.81-1.8 1.8-1.8h4.4c.99 0 1.8.81 1.8 1.8V6"></path>
          <path d="M18.2 6l-.72 11.02A2 2 0 0 1 15.48 19H8.52a2 2 0 0 1-1.99-1.98L5.8 6"></path>
          <path d="M10 10.5v4.5"></path>
          <path d="M14 10.5v4.5"></path>
        </svg>
      </button>

      <div class="message-bubble" :class="{ mine: isMine, theirs: !isMine }">
        <div class="message-text">{{ message.content }}</div>

        <div class="message-meta">
          <span class="message-time">
            {{ formatTime(message.created_at || message.createdat) }}
          </span>

          <span
            v-if="isMine"
            class="message-status"
            :class="{ read: message.status === 'read' }"
          >
            <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M4 12l4 4 7-7"></path>
              <path d="M10 12l4 4 6-6"></path>
            </svg>
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'MessageBubble',
  props: {
    message: { type: Object, required: true },
    isMine: { type: Boolean, required: true },
    isLight: { type: Boolean, default: false }
  },
  emits: ['delete'],
  data() {
    return { showActions: false }
  },
  methods: {
    formatTime(dateStr) {
      if (!dateStr) return ''
      const d = new Date(dateStr)
      if (Number.isNaN(d.getTime())) return ''
      return d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
    }
  }
}
</script>

<style scoped>
.message-row { display: flex; margin-bottom: 14px; }
.message-row.mine { justify-content: flex-end; }
.message-row.theirs { justify-content: flex-start; }

.message-bubble-wrapper {
  position: relative; display: flex; align-items: center;
  max-width: min(420px, 72%);
}

.message-bubble {
  padding: 14px 16px 12px;
  border-radius: 14px; position: relative; width: 100%;
}

/* Dark - theirs */
.message-bubble.theirs {
  background: rgba(255, 255, 255, 0.075);
  border: 1px solid rgba(255, 255, 255, 0.035);
  color: #eef1fb;
  border-bottom-left-radius: 8px;
}
/* Dark - mine */
.message-bubble.mine {
  background: linear-gradient(180deg, rgba(108, 118, 255, 0.9), rgba(93, 104, 240, 0.92));
  color: #f9fbff;
  border-bottom-right-radius: 8px;
  box-shadow: 0 10px 22px rgba(70, 80, 210, 0.16);
}

/* Light - theirs */
.theme-light .message-bubble.theirs {
  background: #ffffff;
  border-color: #e4e6f0;
  color: #1a1d2e;
  box-shadow: 0 2px 8px rgba(91, 106, 200, 0.06);
}
/* Light - mine */
.theme-light .message-bubble.mine {
  background: linear-gradient(180deg, #5b6aff, #6e79ff);
  color: #ffffff;
  box-shadow: 0 8px 20px rgba(91, 106, 255, 0.25);
}

.message-text { font-size: 14px; line-height: 1.6; font-weight: 500; white-space: pre-wrap; }

.message-meta {
  margin-top: 8px; display: flex; justify-content: flex-end; align-items: center; gap: 6px;
}
.message-time { font-size: 11px; opacity: 0.72; }
.message-status { display: inline-flex; align-items: center; color: rgba(255, 255, 255, 0.72); opacity: 0.95; }
.message-status.read { color: #4f8ef7; opacity: 1; }
.theme-light .message-status { color: rgba(255, 255, 255, 0.7); }
.theme-light .message-status.read { color: #93c5fd; }

.message-action {
  position: absolute; left: -36px;
  width: 28px; height: 28px; border-radius: 10px;
  display: grid; place-items: center;
  opacity: 0; visibility: hidden; transition: all 0.2s ease; cursor: pointer;
}
.message-action.visible { opacity: 1; visibility: visible; }

.delete-btn {
  color: #aeb7dc;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.04);
}
.theme-light .delete-btn { color: #9098b8; background: #f0f1f8; border-color: #e4e6f0; }
.delete-btn:hover { color: #ff4d6d; background: rgba(255, 77, 109, 0.1); border-color: rgba(255, 77, 109, 0.2); }
</style>