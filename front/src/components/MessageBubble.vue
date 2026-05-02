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
        <div v-if="message.type !== 'voice'" class="message-text">{{ message.content }}</div>

        <div v-else class="voice-player">
          <button type="button" class="play-btn" @click="togglePlay">
            <svg v-if="!isPlaying" viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
              <path d="M5 3l14 9-14 9V3z"/>
            </svg>
            <svg v-else viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
              <path d="M6 4h4v16H6zM14 4h4v16h-4z"/>
            </svg>
          </button>
          <div class="voice-progress" @click="seek">
            <div class="voice-bar">
              <div class="voice-fill" :style="{ width: progress + '%' }"></div>
            </div>
          </div>
          <span class="voice-duration">{{ durationText }}</span>
          <audio ref="audio" :src="message.content" @timeupdate="onTimeUpdate" @ended="onEnded" @loadedmetadata="onMeta"></audio>
        </div>

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
  computed: {
    durationText() {
      const t = this.isPlaying
        ? (this.$refs.audio?.currentTime || 0)
        : this.duration
      const m = Math.floor(t / 60).toString().padStart(2, '0')
      const s = Math.floor(t % 60).toString().padStart(2, '0')
      return `${m}:${s}`
    }
  },

  data() {
    return { 
    showActions: false,
    isPlaying: false,
    progress: 0,
    duration: 0
     }
  },
  methods: {
    togglePlay() {
    const audio = this.$refs.audio
    if (!audio) return
    if (this.isPlaying) {
      audio.pause()
      this.isPlaying = false
    } else {
      audio.play()
      this.isPlaying = true
    }
  },

  onTimeUpdate() {
    const audio = this.$refs.audio
    if (!audio || !audio.duration) return
    this.progress = (audio.currentTime / audio.duration) * 100
  },

  onEnded() {
    this.isPlaying = false
    this.progress = 0
  },

  onMeta() {
    this.duration = this.$refs.audio?.duration || 0
  },

  seek(e) {
    const audio = this.$refs.audio
    if (!audio) return
    const bar = e.currentTarget
    const rect = bar.getBoundingClientRect()
    const ratio = (e.clientX - rect.left) / rect.width
    audio.currentTime = ratio * audio.duration
  },
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
  word-break: break-word;
  overflow-wrap: anywhere;
  min-width: 0;
}

@media (max-width: 760px) {
  .message-bubble-wrapper {
    max-width: calc(100% - 44px);
  }
  .message-row.theirs .message-bubble-wrapper {
    max-width: 85%;
  }
}

.message-bubble.theirs {
  background: rgba(255, 255, 255, 0.075);
  border: 1px solid rgba(255, 255, 255, 0.035);
  color: #eef1fb;
  border-bottom-left-radius: 8px;
}

.message-bubble.mine {
  background: linear-gradient(180deg, rgba(108, 118, 255, 0.9), rgba(93, 104, 240, 0.92));
  color: #f9fbff;
  border-bottom-right-radius: 8px;
  box-shadow: 0 10px 22px rgba(70, 80, 210, 0.16);
}

.theme-light .message-bubble.theirs {
  background: #ffffff;
  border-color: #e4e6f0;
  color: #1a1d2e;
  box-shadow: 0 2px 8px rgba(91, 106, 200, 0.06);
}

.theme-light .message-bubble.mine {
  background: linear-gradient(180deg, #5b6aff, #6e79ff);
  color: #ffffff;
  box-shadow: 0 8px 20px rgba(91, 106, 255, 0.25);
}

.message-text { font-size: 14px; line-height: 1.6; font-weight: 500; white-space: pre-wrap; word-break: break-word; overflow-wrap: anywhere; }

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
.message-status.failed { color: #ff4d6d; opacity: 1; }

.message-row {
  animation: msgFade 0.2s ease-out both;
}

.message-row.mine {
  animation: msgFade 0.2s ease-out both;
}

.message-row.theirs {
  animation: msgFade 0.2s ease-out both;
}

@keyframes msgFade {
  from { opacity: 0; }
  to { opacity: 1; }
}
</style>