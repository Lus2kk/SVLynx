<template>
  <article
    class="channel-post"
    :class="{ 'theme-light': isLight, pinned: post.pinned, 'menu-open': isHighlighted }"
    @touchstart="onTouchStart"
    @touchend="onTouchEnd"
    @touchmove="onTouchMove"
    @touchcancel="onTouchCancel"
    @contextmenu.prevent="openMenu"
  >
    <div v-if="post.pinned" class="pin-label">
      <svg viewBox="0 0 24 24" width="10" height="10" fill="currentColor">
        <path d="M16 12V4h1V2H7v2h1v8l-2 2v2h5.2v6h1.6v-6H18v-2l-2-2z"/>
      </svg>
      Pinned
    </div>

    <div class="post-content">
      <div v-if="isVoice" class="post-voice">
        <div class="voice-icon">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
            <path d="M12 1a4 4 0 0 1 4 4v6a4 4 0 0 1-8 0V5a4 4 0 0 1 4-4z"/>
            <path d="M19 10a7 7 0 0 1-14 0H3a9 9 0 0 0 18 0h-2z"/>
          </svg>
        </div>
        <audio :src="post.content" controls class="post-audio" preload="metadata" />
      </div>
      <p v-else-if="post.content && !post.media_url" class="post-text">{{ post.content }}</p>
      <img v-if="post.media_url && post.media_type === 'image'" :src="post.media_url" class="post-img" loading="lazy" />
      <video v-else-if="post.media_url && post.media_type === 'video'" :src="post.media_url" class="post-video" controls />
      <a v-else-if="post.media_url" :href="post.media_url" class="post-file-link" target="_blank" rel="noopener">
        <div class="post-file-icon">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.8">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
            <polyline points="14 2 14 8 20 8"/>
          </svg>
        </div>
        <span class="post-file-name">{{ post.file_name || 'File' }}</span>
      </a>
    </div>

    <div class="post-footer">
      <span class="post-time">{{ timeText }}</span>
      <div class="post-footer-right">
        <span v-if="post.view_count" class="post-views">
          <svg viewBox="0 0 24 24" width="11" height="11" fill="none" stroke="currentColor" stroke-width="1.8">
            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
            <circle cx="12" cy="12" r="3"/>
          </svg>
          {{ post.view_count }}
        </span>
      </div>
    </div>

    <teleport to="body">
      <div v-if="menuOpen" class="ctx-overlay" @click="closeMenu" @contextmenu.prevent="closeMenu">
        <div class="ctx-clone" :style="cloneStyle">
          <div
            class="channel-post clone-post"
            :class="{ 'theme-light': isLight, pinned: post.pinned }"
            :style="{ width: cloneStyle.width, maxWidth: cloneStyle.width, minWidth: cloneStyle.width, boxSizing: 'border-box' }"
          >
            <div v-if="post.pinned" class="pin-label">
              <svg viewBox="0 0 24 24" width="10" height="10" fill="currentColor">
                <path d="M16 12V4h1V2H7v2h1v8l-2 2v2h5.2v6h1.6v-6H18v-2l-2-2z"/>
              </svg>
              Pinned
            </div>
            <div class="post-content">
              <div v-if="isVoice" class="post-voice">
                <div class="voice-icon">
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                    <path d="M12 1a4 4 0 0 1 4 4v6a4 4 0 0 1-8 0V5a4 4 0 0 1 4-4z"/>
                    <path d="M19 10a7 7 0 0 1-14 0H3a9 9 0 0 0 18 0h-2z"/>
                  </svg>
                </div>
                <audio :src="post.content" controls class="post-audio" preload="metadata" />
              </div>
              <p
                v-else-if="post.content && !post.media_url"
                class="post-text"
                style="white-space: pre-wrap; word-break: break-word; overflow-wrap: anywhere; overflow: hidden;"
              >{{ post.content }}</p>
              <img v-if="post.media_url && post.media_type === 'image'" :src="post.media_url" class="post-img" />
              <video v-else-if="post.media_url && post.media_type === 'video'" :src="post.media_url" class="post-video" controls />
              <a v-else-if="post.media_url" :href="post.media_url" class="post-file-link" target="_blank" rel="noopener">
                <div class="post-file-icon">
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.8">
                    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                    <polyline points="14 2 14 8 20 8"/>
                  </svg>
                </div>
                <span class="post-file-name">{{ post.file_name || 'File' }}</span>
              </a>
            </div>
            <div class="post-footer">
              <span class="post-time">{{ timeText }}</span>
            </div>
          </div>
        </div>

        <div class="ctx-menu" :style="menuStyle" @click.stop>
          <button v-if="isAdmin" class="ctx-item" @click="$emit('pin', post); closeMenu()">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
              <path d="M16 12V4h1V2H7v2h1v8l-2 2v2h5.2v6h1.6v-6H18v-2l-2-2z"/>
            </svg>
            {{ post.pinned ? 'Открепить' : 'Закрепить' }}
          </button>
          <button v-if="isAuthor || isAdmin" class="ctx-item" @click="$emit('edit', post); closeMenu()">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.8">
              <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
              <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
            </svg>
            Редактировать
          </button>
          <div v-if="isAdmin || isAuthor" class="ctx-divider"></div>
          <button class="ctx-item ctx-delete" @click="$emit('delete', post.id); closeMenu()">
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
  </article>
</template>

<script>
export default {
  name: 'ChannelPost',
  props: {
    post:          { type: Object,  required: true },
    isAdmin:       { type: Boolean, default: false },
    isLight:       { type: Boolean, default: false },
    currentUserId: { type: String,  default: null }
  },

  emits: ['delete', 'pin', 'edit'],

  data() {
    return {
      menuOpen: false,
      menuStyle: {},
      cloneStyle: {},
      isHighlighted: false,
      pressTimer: null,
    }
  },

  beforeUnmount() {
    clearTimeout(this.pressTimer)
  },

  computed: {
    isVoice() {
      const c = this.post.content || ''
      return c.startsWith('http') && (
        c.includes('/uploads/voice/') ||
        c.includes('.webm') ||
        c.includes('.ogg') ||
        (c.includes('.mp4') && c.includes('voice'))
      )
    },
    isAuthor() { return String(this.post.author_id) === String(this.currentUserId) },
    canEdit()  { return this.isAuthor || this.isAdmin },
    timeText() {
      if (!this.post.created_at) return ''
      const d = new Date(this.post.created_at)
      const now = new Date()
      const isToday = d.toDateString() === now.toDateString()
      if (isToday) return d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
      return d.toLocaleString('ru-RU', { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' })
    }
  },

  methods: {
    openMenu(e) {
      const rect = this.$el.getBoundingClientRect()
      const menuH = 160
      const menuW = 220
      let top = rect.bottom + 8
      if (top + menuH > window.innerHeight - 20) top = rect.top - menuH - 8
      if (top < 8) top = 8

      const isMobile = 'ontouchstart' in window
      if (isMobile) {
        this.menuStyle = { top: top + 'px', right: '12px', width: menuW + 'px' }
      } else {
        const x = e.clientX ?? rect.left
        const left = Math.min(x, window.innerWidth - menuW - 12)
        this.menuStyle = { top: top + 'px', left: left + 'px', width: menuW + 'px' }
      }

      this.cloneStyle = {
        position: 'fixed',
        top:      rect.top  + 'px',
        left:     rect.left + 'px',
        width:    rect.width + 'px',
        zIndex:   '2002',
        pointerEvents: 'none',
      }

      this.isHighlighted = true
      this.menuOpen = true
    },

    closeMenu() {
      this.menuOpen = false
      this.isHighlighted = false
    },

   onTouchStart(e) {
      this._touchMoved = false
      const touch = e.touches[0]
      this.pressTimer = setTimeout(() => {
        if (!this._touchMoved) this.openMenu({ clientX: touch.clientX, clientY: touch.clientY })
      }, 500)
    },
    onTouchEnd()    { clearTimeout(this.pressTimer) },
    onTouchCancel() { clearTimeout(this.pressTimer); this._touchMoved = true },
    onTouchMove()   { this._touchMoved = true; clearTimeout(this.pressTimer) },
  }
}
</script>

<style scoped>
.channel-post {
  padding: 6px 10px;
  border-radius: 16px 16px 16px 4px;
  background: rgba(30, 35, 60, 0.95);
  border: 1px solid rgba(255, 255, 255, 0.1);
  position: relative;
  transition: border-color 0.18s, background 0.18s;
  margin-bottom: 4px;
  min-width: 120px;
  max-width: 640px;
  width: fit-content;
  user-select: none;
  -webkit-user-select: none;
  -webkit-touch-callout: none;
}
.channel-post:hover {
  background: rgba(38, 44, 72, 0.95);
  border-color: rgba(255, 255, 255, 0.14);
}
.channel-post.pinned {
  background: rgba(110, 121, 255, 0.15);
  border-color: rgba(110, 121, 255, 0.22);
}
.channel-post.menu-open { opacity: 0.15; }

.theme-light .channel-post { background: #ffffff; border-color: #e4e6f0; }
.theme-light .channel-post:hover { background: #f5f6fc; border-color: rgba(91, 106, 255, 0.18); }
.theme-light .channel-post.pinned { background: rgba(91, 106, 255, 0.04); border-color: rgba(91, 106, 255, 0.18); }

.pin-label {
  display: inline-flex; align-items: center; gap: 4px;
  font-size: 10px; font-weight: 700; letter-spacing: 0.05em; text-transform: uppercase;
  color: #6e79ff; margin-bottom: 8px;
}

.post-content { margin-bottom: 10px; }
.post-text {
  color: #dde2f8; font-size: 14.5px; line-height: 1.65;
  font-weight: 450; white-space: pre-wrap; word-break: break-word;
  overflow-wrap: anywhere; margin: 0;
}
.theme-light .post-text { color: #1e2240; }

.post-img {
  display: block; max-width: 100%; border-radius: 12px;
  margin-top: 10px; object-fit: cover;
  border: 1px solid rgba(255, 255, 255, 0.06);
}
.theme-light .post-img { border-color: #eaecf4; }
.post-video { display: block; max-width: 100%; border-radius: 12px; margin-top: 10px; }

.post-file-link {
  display: inline-flex; align-items: center; gap: 10px; margin-top: 10px;
  padding: 10px 14px; border-radius: 12px; text-decoration: none;
  background: rgba(110, 121, 255, 0.08);
  border: 1px solid rgba(110, 121, 255, 0.18);
  transition: background 0.15s;
}
.post-file-link:hover { background: rgba(110, 121, 255, 0.14); }
.post-file-icon {
  width: 32px; height: 32px; border-radius: 9px; flex-shrink: 0;
  display: grid; place-items: center;
  color: #6e79ff; background: rgba(110, 121, 255, 0.12);
}
.post-file-name {
  flex: 1; color: #c8cef0; font-size: 13px; font-weight: 600;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.theme-light .post-file-name { color: #1e2240; }

.post-footer {
  display: flex; align-items: center; justify-content: space-between;
  gap: 10px; min-height: 22px;
}
.post-footer-right { display: flex; align-items: center; gap: 8px; }
.post-time { color: #4e577a; font-size: 11px; font-weight: 500; }
.theme-light .post-time { color: #9098b8; }
.post-views {
  display: flex; align-items: center; gap: 3px;
  color: #4e577a; font-size: 11px; font-weight: 500;
}
.theme-light .post-views { color: #9098b8; }

.post-voice {
  display: flex; align-items: center; gap: 10px;
  padding: 10px 14px; border-radius: 12px;
  background: rgba(110, 121, 255, 0.08);
  border: 1px solid rgba(110, 121, 255, 0.15);
  margin-top: 4px;
}
.voice-icon {
  width: 32px; height: 32px; border-radius: 50%; flex-shrink: 0;
  display: grid; place-items: center;
  color: #fff; background: linear-gradient(135deg, #6e79ff, #8669ff);
}
.post-audio { flex: 1; height: 32px; min-width: 0; accent-color: #6e79ff; }

/* Context menu */
.ctx-clone { position: fixed; z-index: 2002; pointer-events: none; }
.clone-post {
  transition: none;
  box-shadow: none !important;
  outline: none !important;
  opacity: 1 !important;
  overflow: hidden;
}

.ctx-overlay {
  position: fixed; inset: 0; z-index: 2000;
  background: rgba(0,0,0,0.72);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  animation: ctxFadeIn 0.15s ease;
}
.ctx-menu {
  position: fixed; z-index: 2001;
  background: rgba(22,26,46,0.97);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 14px; padding: 6px; min-width: 180px;
  box-shadow: 0 20px 50px rgba(0,0,0,0.5);
  animation: ctxSlideIn 0.2s cubic-bezier(0.16,1,0.3,1);
}
.ctx-item {
  width: 100%; display: flex; align-items: center; gap: 10px;
  padding: 10px 12px; border-radius: 10px;
  background: none; border: none; cursor: pointer;
  color: #eef2ff; font-size: 14px; font-weight: 500;
  font-family: inherit; text-align: left; transition: background 0.15s;
}
.ctx-item:hover { background: rgba(255,255,255,0.06); }
.ctx-delete { color: #ff4d6d; }
.ctx-delete:hover { background: rgba(255,77,109,0.1); }
.ctx-divider { height: 1px; background: rgba(255,255,255,0.06); margin: 4px 0; }

@media (max-width: 760px) {
  .ctx-menu { border-radius: 18px; padding: 4px; min-width: unset; }
  .ctx-item { padding: 12px 20px; font-size: 15px; }
}

@keyframes ctxFadeIn  { from { opacity: 0; } to { opacity: 1; } }
@keyframes ctxSlideIn { from { opacity: 0; transform: scale(0.95) translateY(-4px); } to { opacity: 1; transform: scale(1) translateY(0); } }
</style>