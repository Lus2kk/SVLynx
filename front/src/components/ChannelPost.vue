<template>
  <article class="channel-post" :class="{ 'theme-light': isLight, pinned: post.pinned }"
    @mouseenter="showActions = true" @mouseleave="showActions = false">
    <div v-if="post.pinned" class="pin-badge">
      <svg viewBox="0 0 24 24" width="10" height="10" fill="currentColor"><path d="M16 12V4h1V2H7v2h1v8l-2 2v2h5.2v6h1.6v-6H18v-2l-2-2z"/></svg>
      Pinned
    </div>
    <div class="post-body">
      <p v-if="post.content" class="post-text">{{ post.content }}</p>
      <img v-if="post.media_url && post.media_type === 'image'" :src="post.media_url" class="post-media-img" />
      <video v-else-if="post.media_url && post.media_type === 'video'" :src="post.media_url" class="post-media-video" controls />
      <a v-else-if="post.media_url" :href="post.media_url" class="post-file" target="_blank">
        📎 {{ post.file_name || 'File' }}
      </a>
    </div>
    <div class="post-footer">
      <span class="post-time">{{ timeText }}</span>
      <span v-if="post.view_count" class="post-views">
        <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="1.8">
          <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/>
        </svg>
        {{ post.view_count }}
      </span>
      <div v-if="canEdit && showActions" class="post-actions">
        <button v-if="isAdmin" class="post-action-btn" @click="$emit('pin', post)" :title="post.pinned ? 'Unpin' : 'Pin'">
          <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor"><path d="M16 12V4h1V2H7v2h1v8l-2 2v2h5.2v6h1.6v-6H18v-2l-2-2z"/></svg>
        </button>
        <button v-if="isAuthor || isAdmin" class="post-action-btn" @click="$emit('edit', post)" title="Edit">
          <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="1.8">
            <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
            <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
          </svg>
        </button>
        <button class="post-action-btn danger" @click="$emit('delete', post.id)" title="Delete">
          <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="1.8">
            <path d="M3 6h18"/>
            <path d="M8 6V4.8c0-.99.81-1.8 1.8-1.8h4.4c.99 0 1.8.81 1.8 1.8V6"/>
            <path d="M18.2 6l-.72 11.02A2 2 0 0 1 15.48 19H8.52a2 2 0 0 1-1.99-1.98L5.8 6"/>
          </svg>
        </button>
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
    isAuthor() { return String(this.post.author_id) === String(this.currentUserId) },
    canEdit()  { return this.isAuthor || this.isAdmin },
    timeText() {
      if (!this.post.created_at) return ''
      const d = new Date(this.post.created_at)
      return d.toLocaleString('ru-RU', { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' })
    }
  }
}
</script>

<style scoped>
.channel-post {
  padding: 14px 16px; border-radius: 14px;
  background: rgba(20,26,50,0.7);
  border: 1px solid rgba(255,255,255,0.06);
  position: relative; transition: border-color 0.15s;
}
.channel-post:hover { border-color: rgba(110,121,255,0.2); }
.channel-post.pinned { border-color: rgba(110,121,255,0.25); background: rgba(110,121,255,0.06); }
.channel-post.theme-light { background: #fff; border-color: #e4e6f0; }
.channel-post.theme-light.pinned { background: rgba(91,106,255,0.04); border-color: rgba(91,106,255,0.2); }

.pin-badge {
  display: flex; align-items: center; gap: 4px;
  font-size: 10px; font-weight: 700; color: #6e79ff;
  margin-bottom: 8px; text-transform: uppercase; letter-spacing: 0.05em;
}

.post-body { margin-bottom: 8px; }
.post-text {
  color: #dde2f8; font-size: 14px; line-height: 1.6; font-weight: 500;
  white-space: pre-wrap; word-break: break-word;
}
.theme-light .post-text { color: #1a1d2e; }

.post-media-img { max-width: 100%; border-radius: 10px; margin-top: 8px; display: block; }
.post-media-video { max-width: 100%; border-radius: 10px; margin-top: 8px; display: block; }
.post-file {
  display: inline-flex; align-items: center; gap: 6px; margin-top: 8px;
  color: #6e79ff; font-size: 13px; text-decoration: none;
  padding: 6px 12px; border-radius: 8px;
  background: rgba(110,121,255,0.1); border: 1px solid rgba(110,121,255,0.2);
}

.post-footer { display: flex; align-items: center; gap: 10px; }
.post-time { color: #5d6888; font-size: 11px; font-weight: 500; flex: 1; }
.post-views { display: flex; align-items: center; gap: 3px; color: #5d6888; font-size: 11px; }

.post-actions { display: flex; gap: 4px; }
.post-action-btn {
  width: 26px; height: 26px; border-radius: 8px;
  display: grid; place-items: center;
  color: #a6afd4; background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.06); cursor: pointer; transition: all 0.15s;
}
.post-action-btn:hover { background: rgba(110,121,255,0.15); color: #6e79ff; border-color: rgba(110,121,255,0.3); }
.post-action-btn.danger:hover { background: rgba(255,77,109,0.12); color: #ff4d6d; border-color: rgba(255,77,109,0.25); }
</style>