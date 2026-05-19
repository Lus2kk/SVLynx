<template>
  <article
    class="channel-post"
    :class="{ 'theme-light': isLight, pinned: post.pinned }"
    @mouseenter="showActions = true"
    @mouseleave="showActions = false"
  >
    <!-- Pin label -->
    <div v-if="post.pinned" class="pin-label">
      <svg viewBox="0 0 24 24" width="10" height="10" fill="currentColor">
        <path d="M16 12V4h1V2H7v2h1v8l-2 2v2h5.2v6h1.6v-6H18v-2l-2-2z"/>
      </svg>
      Pinned
    </div>

    <!-- Content -->
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

    <!-- Footer -->
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

        <!-- Actions — visible on hover for admins/authors -->
        <transition name="fade-actions">
          <div v-if="canEdit && showActions" class="post-actions">
            <button
              v-if="isAdmin"
              class="post-btn"
              :class="{ active: post.pinned }"
              :title="post.pinned ? 'Unpin' : 'Pin'"
              @click.stop="$emit('pin', post)"
            >
              <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor">
                <path d="M16 12V4h1V2H7v2h1v8l-2 2v2h5.2v6h1.6v-6H18v-2l-2-2z"/>
              </svg>
            </button>
            <button
              v-if="isAuthor || isAdmin"
              class="post-btn"
              title="Edit"
              @click.stop="$emit('edit', post)"
            >
              <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="1.8">
                <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
                <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
              </svg>
            </button>
            <button
              class="post-btn danger"
              title="Delete"
              @click.stop="$emit('delete', post.id)"
            >
              <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="1.8">
                <path d="M3 6h18"/>
                <path d="M8 6V4.8c0-.99.81-1.8 1.8-1.8h4.4c.99 0 1.8.81 1.8 1.8V6"/>
                <path d="M18.2 6l-.72 11.02A2 2 0 0 1 15.48 19H8.52a2 2 0 0 1-1.99-1.98L5.8 6"/>
              </svg>
            </button>
          </div>
        </transition>
      </div>
    </div>
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

  data() { return { showActions: false } },

  computed: {
    isVoice() {
    const c = this.post.content || ''
    return c.startsWith('http') && (
        c.includes('/uploads/voice/') ||
        c.includes('.webm') ||
        c.includes('.ogg') ||
        c.includes('.mp4') && c.includes('voice')
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
  }
}
</script>

<style scoped>

.channel-post {
  padding: 6px 10px;
  border-radius: 16px 16px 16px 4px;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.1);
  position: relative;
  transition: border-color 0.18s, background 0.18s;
  margin-bottom: 4px;
  display: inline-block;
  min-width: 120px;
  max-width: 640px;
  width: auto;
}
.channel-post:hover {
  background: rgba(255, 255, 255, 0.11);
  border-color: rgba(255, 255, 255, 0.14);
}
.channel-post.pinned {
  background: rgba(110, 121, 255, 0.06);
  border-color: rgba(110, 121, 255, 0.22);
}
.theme-light .channel-post {
  background: #ffffff;
  border-color: #e4e6f0;
  box-shadow: none;
}
.theme-light .channel-post:hover {
  background: #f5f6fc;
  border-color: rgba(91, 106, 255, 0.18);
}
.theme-light .channel-post.pinned {
  background: rgba(91, 106, 255, 0.04);
  border-color: rgba(91, 106, 255, 0.18);
}

/* Pin label */
.pin-label {
  display: inline-flex; align-items: center; gap: 4px;
  font-size: 10px; font-weight: 700; letter-spacing: 0.05em; text-transform: uppercase;
  color: #6e79ff; margin-bottom: 8px;
}

/* Content */
.post-content { margin-bottom: 10px; }
.post-text {
  color: #dde2f8;
  font-size: 14.5px;
  line-height: 1.65;
  font-weight: 450;
  white-space: pre-wrap;
  word-break: break-word;
  margin: 0;
}
.theme-light .post-text { color: #1e2240; }

.post-img {
  display: block; max-width: 100%; border-radius: 12px;
  margin-top: 10px; object-fit: cover;
  border: 1px solid rgba(255, 255, 255, 0.06);
}
.theme-light .post-img { border-color: #eaecf4; }

.post-video {
  display: block; max-width: 100%; border-radius: 12px; margin-top: 10px;
}

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
.post-file-download { color: #6e79ff; flex-shrink: 0; }

/* Footer */
.post-footer {
  display: flex; align-items: center; justify-content: space-between;
  gap: 10px; min-height: 22px;
}
.post-footer-right { display: flex; align-items: center; gap: 8px; }
.post-time {
  color: #4e577a; font-size: 11px; font-weight: 500;
}
.theme-light .post-time { color: #9098b8; }
.post-views {
  display: flex; align-items: center; gap: 3px;
  color: #4e577a; font-size: 11px; font-weight: 500;
}
.theme-light .post-views { color: #9098b8; }

/* Action buttons */
.post-actions { display: flex; align-items: center; gap: 3px; }
.post-btn {
  width: 28px; height: 28px; border-radius: 8px;
  display: grid; place-items: center;
  color: #7d87ab; background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.06);
  cursor: pointer; transition: all 0.15s;
}
.post-btn:hover {
  background: rgba(110, 121, 255, 0.14);
  color: #6e79ff;
  border-color: rgba(110, 121, 255, 0.28);
}
.post-btn.active {
  background: rgba(110, 121, 255, 0.18);
  color: #6e79ff;
  border-color: rgba(110, 121, 255, 0.35);
}
.post-btn.danger:hover {
  background: rgba(255, 77, 109, 0.12);
  color: #ff4d6d;
  border-color: rgba(255, 77, 109, 0.25);
}
.theme-light .post-btn {
  color: #9098b8; background: #f5f6fc; border-color: #e4e6f0;
}
.theme-light .post-btn:hover { background: rgba(91, 106, 255, 0.08); color: #5b6aff; }
.theme-light .post-btn.danger:hover { background: rgba(255, 60, 80, 0.08); color: #ff3c50; }

.fade-actions-enter-active, .fade-actions-leave-active { transition: opacity 0.15s; }
.fade-actions-enter-from, .fade-actions-leave-to { opacity: 0; }
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
.post-audio {
  flex: 1; height: 32px; min-width: 0;
  accent-color: #6e79ff;
}
</style>