<template>
  <div
    class="message-row"
    :class="{ mine: isMine, theirs: !isMine, 'theme-light': isLight }"
  >
    <div class="message-bubble-wrapper">
      <div
        class="message-bubble"
        :class="{ mine: isMine, theirs: !isMine, highlight: highlight, 'highlight-active': highlightActive }"
        @touchstart="onTouchStart"
        @touchend.stop="onTouchEnd"
        @touchmove="onTouchCancel"
        @touchcancel="onTouchCancel"
        @contextmenu.prevent="openMenu"
      >
        <div v-if="message.type !== 'voice'" class="message-text">{{ message.content }}</div>

        <div v-else class="voice-player">
          <button type="button" class="play-btn" @click="togglePlay">
            <svg v-if="!isPlaying" viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
              <path d="M5 3l14 9-14 9V3z"/>
            </svg>
            <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
              <path d="M6 4h4v16H6zM14 4h4v16h-4z"/>
            </svg>
          </button>

          <div class="voice-progress" @click="seek">
            <div class="voice-bar">
              <div class="voice-fill" :style="{ width: progress + '%' }"></div>
            </div>
          </div>

          <span class="voice-duration">{{ getDurationText() }}</span>
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
            <svg viewBox="0 0 22 12" width="20" height="10" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M1 6l3 3 5-6"></path>
              <path d="M9 6l3 3 5-6"></path>
            </svg>
          </span>
        </div>
      </div>
    </div>

    <teleport to="body">
      <div v-if="menuOpen" class="ctx-overlay" @click="closeMenu" @contextmenu.prevent="closeMenu" @touchend.stop>
        <div class="ctx-menu" :style="menuStyle" @click.stop>
          <button class="ctx-item" @click="onReply">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.8">
              <polyline points="9 17 4 12 9 7"/>
              <path d="M20 18v-2a4 4 0 0 0-4-4H4"/>
            </svg>
            Ответить
          </button>
          <div class="ctx-divider" v-if="isMine"></div>
          <button v-if="isMine" class="ctx-item ctx-delete" @click="onDelete">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.8">
              <path d="M3 6h18"/>
              <path d="M8 6V4.8c0-.99.81-1.8 1.8-1.8h4.4c.99 0 1.8.81 1.8 1.8V6"/>
              <path d="M18.2 6l-.72 11.02A2 2 0 0 1 15.48 19H8.52a2 2 0 0 1-1.99-1.98L5.8 6"/>
            </svg>
            Удалить
          </button>
        </div>
      </div>
    </teleport>
  </div>
</template>

<script>
export default {
  name: 'MessageBubble',
  props: {
    highlight: { type: Boolean, default: false },
    highlightActive: { type: Boolean, default: false },
    message: { type: Object, required: true },
    isMine: { type: Boolean, required: true },
    isLight: { type: Boolean, default: false }
  },
  emits: ['delete', 'reply'],

  data() {
    return {
      showActions: false,
      isPlaying: false,
      progress: 0,
      duration: 0,
      menuOpen: false,
      menuStyle: {},
      pressTimer: null,
    }
  },

  methods: {
    openMenu(e) {
      const x = e.clientX ?? e.touches?.[0]?.clientX ?? window.innerWidth / 2
      const y = e.clientY ?? e.touches?.[0]?.clientY ?? window.innerHeight / 2

      const menuW = 200
      const menuH = this.isMine ? 108 : 56
      const left = Math.min(x, window.innerWidth - menuW - 12)
      const top = Math.min(y, window.innerHeight - menuH - 12)

      this.menuStyle = { left: left + 'px', top: top + 'px' }
      this.menuOpen = true
    },

    onReply() {
      this.$emit('reply', this.message)
      this.closeMenu()
    },

    onDelete() {
      this.$emit('delete', this.message.id)
      this.closeMenu()
    },

    onTouchStart(e) {
      this.pressTimer = setTimeout(() => {
        this.openMenu(e.touches[0])
        this._justOpened = true
      }, 500)
    },

    onTouchEnd() {
      clearTimeout(this.pressTimer)
    },

    closeMenu() {
      if (this._justOpened) {
        this._justOpened = false
        return
      }
      this.menuOpen = false
    },
    onTouchCancel() { clearTimeout(this.pressTimer) },

    getDurationText() {
      const audio = this.$refs.audio
      if (this.isPlaying && audio) {
        const t = audio.currentTime || 0
        const m = Math.floor(t / 60).toString().padStart(2, '0')
        const s = Math.floor(t % 60).toString().padStart(2, '0')
        return `${m}:${s}`
      }
      const t = (this.duration && isFinite(this.duration) && this.duration > 0.5)
        ? this.duration
        : (this.message.duration || 0)
      const m = Math.floor(t / 60).toString().padStart(2, '0')
      const s = Math.floor(t % 60).toString().padStart(2, '0')
      return `${m}:${s}`
    },

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
      const audio = this.$refs.audio
      if (!audio) return
      if (audio.duration && isFinite(audio.duration)) {
        this.duration = audio.duration
      } else {
        audio.currentTime = 1e101
        audio.addEventListener('timeupdate', () => {
          if (isFinite(audio.duration)) {
            this.duration = audio.duration
            audio.currentTime = 0
          }
        }, { once: true })
      }
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
.message-row { display: flex; margin-bottom: 2px; }
.message-row.mine { justify-content: flex-end; }
.message-row.theirs { justify-content: flex-start; }

.message-bubble-wrapper {
  position: relative; display: flex; align-items: center;
  max-width: min(420px, 72%);
}

.message-bubble {
  padding: 6px 10px 6px 10px;
  border-radius: 14px; position: relative; width: 100%;
  word-break: break-word;
  overflow-wrap: anywhere;
  min-width: 0;
  cursor: default;
  user-select: none;
  -webkit-user-select: none;
}

@media (max-width: 760px) {
  .message-bubble-wrapper { max-width: calc(100% - 44px); }
  .message-row.theirs .message-bubble-wrapper { max-width: 85%; }
}

.message-bubble.theirs {
  background: rgba(30, 35, 60, 0.95);
  border: 1px solid rgba(255, 255, 255, 0.08);
  color: #eef1fb;
  border-bottom-left-radius: 8px;
}

.message-bubble.mine {
  background: linear-gradient(180deg, rgba(108, 118, 255, 0.95), rgba(93, 104, 240, 0.97));
  color: #ffffff;
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

.message-text {
  font-size: 14px; line-height: 1.5; font-weight: 500;
  white-space: pre-wrap; word-break: break-word; overflow-wrap: anywhere;
}

.message-meta {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 3px;
  margin-top: 2px;
}

.message-time {
  font-size: 11px;
  opacity: 0.85;
  color: rgba(255, 255, 255, 0.85);
  white-space: nowrap;
}

.message-status {
  display: inline-flex;
  align-items: center;
  color: rgba(255, 255, 255, 0.5);
  opacity: 1;
  transition: all 0.3s ease;
}
.message-status.read {
  color: #ffffff;
  filter: drop-shadow(0 0 3px rgba(255, 255, 255, 0.8));
}
.theme-light .message-status { color: rgba(255, 255, 255, 0.7); }
.theme-light .message-status.read { color: #93c5fd; }
.message-status.failed { color: #ff4d6d; opacity: 1; }

.message-row { animation: msgFade 0.2s ease-out both; }

.voice-player {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 160px;
  max-width: 220px;
  padding: 2px 0;
}

.play-btn {
  width: 28px; height: 28px;
  border-radius: 50%;
  display: grid; place-items: center; flex-shrink: 0;
  background: rgba(255,255,255,0.2);
  border: none; cursor: pointer;
  color: inherit;
  transition: background 0.2s;
}
.play-btn:hover { background: rgba(255,255,255,0.3); }

.voice-progress {
  flex: 1; cursor: pointer; padding: 6px 0;
}
.voice-bar {
  height: 3px; border-radius: 999px;
  background: rgba(255,255,255,0.25);
}
.voice-fill {
  height: 100%; border-radius: 999px;
  background: currentColor;
  transition: width 0.1s linear;
}
.voice-duration {
  font-size: 11px;
  opacity: 0.8;
  flex-shrink: 0;
  min-width: 36px;
  text-align: right;
}

/* Context menu */
.ctx-overlay {
  position: fixed; inset: 0; z-index: 1000;
  background: rgba(0,0,0,0.3);
  backdrop-filter: blur(2px);
  animation: ctxFadeIn 0.15s ease;
}

.ctx-menu {
  position: fixed;
  background: rgba(22, 26, 46, 0.97);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 14px;
  padding: 6px;
  min-width: 180px;
  box-shadow: 0 20px 50px rgba(0,0,0,0.5);
  animation: ctxSlideIn 0.2s cubic-bezier(0.16,1,0.3,1);
}

.ctx-item {
  width: 100%; display: flex; align-items: center; gap: 10px;
  padding: 10px 12px; border-radius: 10px;
  background: none; border: none; cursor: pointer;
  color: #eef2ff; font-size: 14px; font-weight: 500;
  font-family: inherit; text-align: left;
  transition: background 0.15s;
}
.ctx-item:hover { background: rgba(255,255,255,0.06); }

.ctx-delete { color: #ff4d6d; }
.ctx-delete:hover { background: rgba(255,77,109,0.1); }

.ctx-divider {
  height: 1px; background: rgba(255,255,255,0.06);
  margin: 4px 0;
}

@keyframes ctxFadeIn {
  from { opacity: 0; } to { opacity: 1; }
}
@keyframes ctxSlideIn {
  from { opacity: 0; transform: scale(0.95) translateY(-4px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}

@keyframes msgFade {
  from { opacity: 0; }
  to { opacity: 1; }
}

.message-bubble.highlight {
  outline: 2px solid rgba(110, 121, 255, 0.5);
  outline-offset: 2px;
}
.message-bubble.highlight-active {
  outline: 2px solid #6e79ff;
  outline-offset: 2px;
  box-shadow: 0 0 0 4px rgba(110, 121, 255, 0.15);
}
</style>