<template>
  <div
    class="message-row"
    :class="{ mine: isMine, theirs: !isMine, 'theme-light': isLight, 'is-selected': isSelected, 'is-selecting': isSelecting }"
    @click="isSelecting ? $emit('select', message) : null"
  >
    <!-- Lightbox -->
    <teleport to="body">
      <div v-if="lightboxUrl" class="lightbox" @click="lightboxUrl = null">
        <img :src="lightboxUrl" class="lightbox-img" @click.stop />
        <button class="lightbox-close" @click="lightboxUrl = null">✕</button>
      </div>
    </teleport>

    <!-- Чекбокс -->
    <transition name="cb-fade">
      <div v-if="isSelecting" class="select-checkbox" :class="{ checked: isSelected }">
        <svg v-if="isSelected" viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="white" stroke-width="3">
          <path d="M5 12l5 5L19 7"/>
        </svg>
      </div>
    </transition>

    <div class="message-bubble-wrapper">
      <div
        class="message-bubble"
        :class="{ mine: isMine, theirs: !isMine, highlight: highlight, 'highlight-active': highlightActive }"
        @touchstart="onTouchStart"
         @touchend.prevent="onTouchEnd"
        @touchend.stop="onTouchEnd"
        @touchmove="onTouchCancel"
        @touchcancel="onTouchCancel"
        @contextmenu.prevent="openMenu"
      >
        <!-- Цитата -->
        <div v-if="message.reply_to" class="reply-quote">
          <div class="reply-quote-bar"></div>
          <div class="reply-quote-content">
            <span class="reply-quote-name">{{ message.reply_to.is_mine ? 'Вы' : 'Собеседник' }}</span>
            <span class="reply-quote-text">{{ message.reply_to.type === 'voice' ? '🎤 Голосовое' : message.reply_to.type === 'image' ? '📷 Фото' : message.reply_to.content }}</span>
          </div>
        </div>

        <!-- Текст -->
        <div v-if="message.type === 'text' || !message.type" class="message-text">
          <template v-if="isUrl(message.content)">
            <a :href="message.content" target="_blank" class="message-link">{{ message.content }}</a>
          </template>
          <template v-else>{{ message.content }}</template>
        </div>

        <!-- Фото -->
        <div v-else-if="message.type === 'image'" class="media-image-wrap">
          <img :src="message.content" class="media-image" @click="lightboxUrl = message.content" />
          <div class="image-meta">
            <span class="message-time">{{ formatTime(message.created_at || message.createdat) }}</span>
            <span v-if="isMine" class="message-status" :class="{ read: message.status === 'read' }">
              <svg viewBox="0 0 16 12" width="16" height="12" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
  <path d="M2 6 L5 9 L12 2"></path>
  <circle v-if="message.status === 'read'" cx="14" cy="9" r="1.8" fill="currentColor" stroke="none"></circle>
</svg>
            </span>
          </div>
        </div>

        <!-- Видео -->
        <div v-else-if="message.type === 'video'" class="media-video-wrap">
          <video :src="message.content" class="media-video" controls></video>
        </div>

        <!-- Аудио -->
        <div v-else-if="message.type === 'audio'" class="media-audio-wrap">
          <audio :src="message.content" controls class="media-audio"></audio>
          <span v-if="message.file_name" class="media-filename">{{ message.file_name }}</span>
        </div>

        <!-- Файл -->
        <div v-else-if="message.type === 'file'" class="media-file-wrap">
          <a :href="message.content" target="_blank" class="media-file-link">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.8">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
              <path d="M14 2v6h6"/>
            </svg>
            <div class="media-file-info">
              <span class="media-filename">{{ message.file_name || 'File' }}</span>
              <span class="media-filesize">{{ formatSize(message.file_size) }}</span>
            </div>
          </a>
        </div>

        <!-- Голосовое -->
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

        <div class="message-meta" v-if="message.type !== 'image'">
          <span class="message-time">{{ formatTime(message.created_at || message.createdat) }}</span>
          <span v-if="isMine" class="message-status" :class="{ read: message.status === 'read' }">
            <svg viewBox="0 0 16 12" width="16" height="12" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
  <path d="M2 6 L5 9 L12 2"/>
  <circle v-if="message.status === 'read'" cx="14" cy="9" r="1.8" fill="currentColor" stroke="none"/>
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
          <button class="ctx-item" @click="onSelect">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.8">
              <circle cx="12" cy="12" r="9"/>
              <path d="M8 12l3 3 5-5"/>
            </svg>
            Выбрать
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
    highlight:      { type: Boolean, default: false },
    highlightActive:{ type: Boolean, default: false },
    message:        { type: Object,  required: true },
    isMine:         { type: Boolean, required: true },
    isLight:        { type: Boolean, default: false },
    isSelecting:    { type: Boolean, default: false },
    isSelected:     { type: Boolean, default: false },
  },
  emits: ['delete', 'reply', 'select'],

  data() {
    return {
      isPlaying:   false,
      progress:    0,
      duration:    0,
      lightboxUrl: null,
      menuOpen:    false,
      menuStyle:   {},
      pressTimer:  null,
    }
  },

  methods: {
    isUrl(str) { return str && (str.startsWith('http://') || str.startsWith('https://')) },

    formatSize(bytes) {
      if (!bytes) return ''
      if (bytes < 1024)    return bytes + ' B'
      if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB'
      return (bytes / 1048576).toFixed(1) + ' MB'
    },

    openMenu(e) {
      if (this.isSelecting) { this.$emit('select', this.message); return }
      const x = e.clientX ?? e.touches?.[0]?.clientX ?? window.innerWidth / 2
      const y = e.clientY ?? e.touches?.[0]?.clientY ?? window.innerHeight / 2
      const menuW = 200, menuH = this.isMine ? 150 : 100
      const left = Math.min(x, window.innerWidth - menuW - 12)
      const top  = Math.min(y, window.innerHeight - menuH - 12)
      this.menuStyle = { left: left + 'px', top: top + 'px' }
      this.menuOpen = true
    },

    onReply() { this.$emit('reply', this.message); this.closeMenu() },
    onDelete() { this.$emit('delete', this.message.id); this.closeMenu() },
    onSelect() { this.$emit('select', this.message); this.closeMenu() },

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
  this._justOpened = false
  this.menuOpen = false
},
    onTouchCancel() { clearTimeout(this.pressTimer) },

    getDurationText() {
      const audio = this.$refs.audio
      if (this.isPlaying && audio) {
        const t = audio.currentTime || 0
        return `${Math.floor(t/60).toString().padStart(2,'0')}:${Math.floor(t%60).toString().padStart(2,'0')}`
      }
      const t = (this.duration && isFinite(this.duration) && this.duration > 0.5) ? this.duration : (this.message.duration || 0)
      return `${Math.floor(t/60).toString().padStart(2,'0')}:${Math.floor(t%60).toString().padStart(2,'0')}`
    },

    togglePlay() {
      const audio = this.$refs.audio
      if (!audio) return
      this.isPlaying ? audio.pause() : audio.play()
      this.isPlaying = !this.isPlaying
    },
    onTimeUpdate() {
      const audio = this.$refs.audio
      if (!audio || !audio.duration) return
      this.progress = (audio.currentTime / audio.duration) * 100
    },
    onEnded() { this.isPlaying = false; this.progress = 0 },
    onMeta() {
      const audio = this.$refs.audio
      if (!audio) return
      if (audio.duration && isFinite(audio.duration)) { this.duration = audio.duration; return }
      audio.currentTime = 1e101
      audio.addEventListener('timeupdate', () => {
        if (isFinite(audio.duration)) { this.duration = audio.duration; audio.currentTime = 0 }
      }, { once: true })
    },
    seek(e) {
      const audio = this.$refs.audio
      if (!audio) return
      const rect = e.currentTarget.getBoundingClientRect()
      audio.currentTime = ((e.clientX - rect.left) / rect.width) * audio.duration
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
.message-row {
  display: flex; margin-bottom: 0px; align-items: center; gap: 8px;
  padding: 1px 0; transition: background 0.15s;
  border-radius: 10px;
}
.message-row.mine   { justify-content: flex-end; }
.message-row.theirs { justify-content: flex-start; }
.message-row.is-selected { background: rgba(110,121,255,0.08); }
.message-row.is-selecting { cursor: pointer; }

/* чекбокс */
.select-checkbox {
  width: 24px; height: 24px; border-radius: 50%; flex-shrink: 0;
  border: 2px solid rgba(110,121,255,0.4);
  display: grid; place-items: center;
  background: transparent;
  transition: all 0.2s cubic-bezier(0.34,1.56,0.64,1);
}
.select-checkbox.checked {
  background: #6e79ff; border-color: #6e79ff;
  transform: scale(1.1);
}
.message-row.mine .select-checkbox { order: 2; }

.cb-fade-enter-active, .cb-fade-leave-active { transition: all 0.2s; }
.cb-fade-enter-from, .cb-fade-leave-to { opacity: 0; transform: scale(0.6); }

.message-bubble-wrapper {
  position: relative; display: flex; align-items: center;
  max-width: min(420px, 72%);
}

.message-bubble {
  padding: 6px 10px; border-radius: 14px; position: relative;
  width: 100%; word-break: break-word; overflow-wrap: anywhere; min-width: 0;
  cursor: default; user-select: none; -webkit-user-select: none;
  transition: transform 0.1s;
}
.is-selecting .message-bubble { cursor: pointer; }
.message-bubble:has(.media-image-wrap) { padding: 0; overflow: hidden; }

@media (max-width: 760px) {
  .message-bubble-wrapper { max-width: calc(100% - 44px); }
  .message-row.theirs .message-bubble-wrapper { max-width: 85%; }
}

.message-bubble.theirs { background: rgba(30,35,60,0.95); border: 1px solid rgba(255,255,255,0.08); color: #eef1fb; border-radius: 16px 16px 16px 4px; }
.message-bubble.mine   { background: linear-gradient(180deg,rgba(108,118,255,0.95),rgba(93,104,240,0.97)); color: #fff; border-radius: 16px 16px 4px 16px; box-shadow: 0 10px 22px rgba(70,80,210,0.16); }.theme-light .message-bubble.theirs { background: #fff; border-color: #e4e6f0; color: #1a1d2e; }
.theme-light .message-bubble.mine   { background: linear-gradient(180deg,#5b6aff,#6e79ff); color: #fff; }

.reply-quote { display:flex; gap:8px; margin-bottom:6px; padding:6px 8px; border-radius:8px; background:rgba(0,0,0,0.15); cursor:pointer; }
.reply-quote-bar { width:3px; border-radius:2px; background:#6e79ff; flex-shrink:0; }
.reply-quote-content { display:flex; flex-direction:column; gap:2px; min-width:0; }
.reply-quote-name { font-size:12px; font-weight:700; color:#6e79ff; }
.reply-quote-text { font-size:12px; opacity:0.8; white-space:nowrap; overflow:hidden; text-overflow:ellipsis; }

.message-text { font-size: 14px; line-height: 1.5; font-weight: 500; white-space: pre-wrap; word-break: break-word; overflow-wrap: anywhere; }
.message-meta { display: flex; justify-content: flex-end; align-items: center; gap: 3px; margin-top: 2px; }
.message-time { font-size: 11px; opacity: 0.85; color: rgba(255,255,255,0.85); white-space: nowrap; }
.message-status { display: inline-flex; align-items: center; color: rgba(255,255,255,0.7); transition: none; }
.message-status.read { color: rgba(255,255,255,0.7); filter: none; }
.theme-light .message-status { color: rgba(255,255,255,0.7); }
.theme-light .message-status.read { color: rgba(255,255,255,0.7); }

.media-image-wrap { border-radius: 10px; overflow: hidden; max-width: 260px; position: relative; }
.media-image { width: 100%; display: block; cursor: pointer; border-radius: 10px; }
.media-video-wrap { border-radius: 10px; overflow: hidden; max-width: 260px; }
.media-video { width: 100%; display: block; border-radius: 10px; }
.media-audio-wrap { display: flex; flex-direction: column; gap: 4px; }
.media-audio { width: 200px; }
.media-file-wrap { padding: 2px 0; }
.media-file-link { display: flex; align-items: center; gap: 10px; padding: 8px 10px; border-radius: 10px; background: rgba(255,255,255,0.08); text-decoration: none; color: inherit; }
.media-file-link:hover { background: rgba(255,255,255,0.14); }
.media-file-info { display: flex; flex-direction: column; gap: 2px; }
.media-filename { font-size: 13px; font-weight: 600; }
.media-filesize { font-size: 11px; opacity: 0.6; }
.image-meta { position: absolute; bottom: 6px; right: 8px; display: flex; align-items: center; gap: 3px; background: rgba(0,0,0,0.4); padding: 2px 5px; border-radius: 6px; }
.image-meta .message-time { color: #fff; font-size: 11px; opacity: 0.9; }
.image-meta .message-status { color: rgba(255,255,255,0.7); }
.image-meta .message-status.read { color: #fff; }

.lightbox { position: fixed; inset: 0; z-index: 9999; background: rgba(0,0,0,0.92); display: flex; align-items: center; justify-content: center; cursor: zoom-out; }
.lightbox-img { max-width: 90vw; max-height: 90vh; border-radius: 8px; object-fit: contain; cursor: default; }
.lightbox-close { position: absolute; top: 20px; right: 20px; background: rgba(255,255,255,0.1); border: none; color: white; font-size: 20px; width: 40px; height: 40px; border-radius: 50%; cursor: pointer; }

.voice-player { display: flex; align-items: center; gap: 8px; min-width: 160px; max-width: 220px; padding: 2px 0; }
.play-btn { width: 28px; height: 28px; border-radius: 50%; display: grid; place-items: center; flex-shrink: 0; background: rgba(255,255,255,0.2); border: none; cursor: pointer; color: inherit; transition: background 0.2s; }
.play-btn:hover { background: rgba(255,255,255,0.3); }
.voice-progress { flex: 1; cursor: pointer; padding: 6px 0; }
.voice-bar { height: 3px; border-radius: 999px; background: rgba(255,255,255,0.25); }
.voice-fill { height: 100%; border-radius: 999px; background: currentColor; transition: width 0.1s linear; }
.voice-duration { font-size: 11px; opacity: 0.8; flex-shrink: 0; min-width: 36px; text-align: right; }

.ctx-overlay { position: fixed; inset: 0; z-index: 1000; background: rgba(0,0,0,0.3); backdrop-filter: blur(2px); animation: ctxFadeIn 0.15s ease; }
.ctx-menu { position: fixed; background: rgba(22,26,46,0.97); border: 1px solid rgba(255,255,255,0.08); border-radius: 14px; padding: 6px; min-width: 180px; box-shadow: 0 20px 50px rgba(0,0,0,0.5); animation: ctxSlideIn 0.2s cubic-bezier(0.16,1,0.3,1); }
.ctx-item { width: 100%; display: flex; align-items: center; gap: 10px; padding: 10px 12px; border-radius: 10px; background: none; border: none; cursor: pointer; color: #eef2ff; font-size: 14px; font-weight: 500; font-family: inherit; text-align: left; transition: background 0.15s; }
.ctx-item:hover { background: rgba(255,255,255,0.06); }
.ctx-delete { color: #ff4d6d; }
.ctx-delete:hover { background: rgba(255,77,109,0.1); }
.ctx-divider { height: 1px; background: rgba(255,255,255,0.06); margin: 4px 0; }

@keyframes ctxFadeIn { from { opacity: 0; } to { opacity: 1; } }
@keyframes ctxSlideIn { from { opacity: 0; transform: scale(0.95) translateY(-4px); } to { opacity: 1; transform: scale(1) translateY(0); } }
@keyframes msgFade { from { opacity: 0; } to { opacity: 1; } }
.message-row { animation: msgFade 0.2s ease-out both; }
.message-bubble.highlight { outline: 2px solid rgba(110,121,255,0.5); outline-offset: 2px; }
.message-bubble.highlight-active { outline: 2px solid #6e79ff; outline-offset: 2px; box-shadow: 0 0 0 4px rgba(110,121,255,0.15); }
.message-bubble {
  user-select: none;
  -webkit-user-select: none;
  -webkit-touch-callout: none;
}
</style>