<template>
  <article
    class="channel-post"
    :class="{ 'theme-light': isLight, pinned: post.pinned }"
    @mouseenter="showActions = true"
    @mouseleave="showActions = false"
  >
    <!-- Pinned badge -->
    <div v-if="post.pinned" class="pin-badge">
      <svg viewBox="0 0 24 24" width="10" height="10" fill="currentColor">
        <path d="M16 12V4h1V2H7v2h1v8l-2 2v2h5.2v6h1.6v-6H18v-2l-2-2z"/>
      </svg>
      Закреплено
    </div>

    <!-- VOICE -->
    <div v-if="postType === 'voice'" class="voice-player">
      <button type="button" class="voice-play-btn" @click="togglePlay">
        <svg v-if="!isPlaying" viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
          <path d="M5 3l14 9-14 9V3z"/>
        </svg>
        <svg v-else viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
          <path d="M6 4h4v16H6zM14 4h4v16h-4z"/>
        </svg>
      </button>
      <div class="voice-waveform" @click="seek">
        <div class="voice-wave-bars">
          <div
            v-for="(bar, i) in waveformBars"
            :key="i"
            class="wave-bar"
            :class="{ active: i < activeBars }"
            :style="{ height: bar + 'px' }"
          ></div>
        </div>
        <div class="voice-times">
          <span>{{ currentTimeText }}</span>
          <span>{{ durationText }}</span>
        </div>
      </div>
      <button class="voice-speed-btn" @click="cycleSpeed">{{ playbackSpeed }}x</button>
      <audio
        ref="audio"
        :src="mediaUrl"
        @timeupdate="onTimeUpdate"
        @ended="onEnded"
        @loadedmetadata="onMeta"
      ></audio>
    </div>

    <!-- IMAGE -->
    <template v-else-if="postType === 'image'">
      <p v-if="post.content && !isUrl(post.content)" class="post-text">{{ post.content }}</p>
      <img :src="mediaUrl" class="post-media-img" @click="lightboxUrl = mediaUrl" />
    </template>

    <!-- VIDEO -->
    <template v-else-if="postType === 'video'">
      <p v-if="post.content && !isUrl(post.content)" class="post-text">{{ post.content }}</p>
      <video :src="mediaUrl" class="post-media-video" controls></video>
    </template>

    <!-- AUDIO -->
    <template v-else-if="postType === 'audio'">
      <p v-if="post.content && !isUrl(post.content)" class="post-text">{{ post.content }}</p>
      <audio :src="mediaUrl" controls class="post-media-audio"></audio>
    </template>

    <!-- FILE -->
    <template v-else-if="postType === 'file'">
      <p v-if="post.content && !isUrl(post.content)" class="post-text">{{ post.content }}</p>
      <a :href="mediaUrl" class="post-file" target="_blank" rel="noopener">
        <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.8">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
          <path d="M14 2v6h6"/>
        </svg>
        {{ post.file_name || 'Файл' }}
      </a>
    </template>

    <!-- TEXT (default) -->
    <template v-else>
      <p class="post-text">{{ post.content }}</p>
    </template>

    <!-- Footer -->
    <div class="post-footer">
      <span class="post-time">{{ timeText }}</span>
      <div class="post-right">
        <span v-if="post.view_count" class="post-views">
          <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="1.8">
            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
            <circle cx="12" cy="12" r="3"/>
          </svg>
          {{ post.view_count }}
        </span>
        <div v-if="canEdit" class="post-actions" :class="{ visible: showActions }">
          <button v-if="isAdmin" class="post-action-btn" @click="$emit('pin', post)" :title="post.pinned ? 'Открепить' : 'Закрепить'">
            <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor">
              <path d="M16 12V4h1V2H7v2h1v8l-2 2v2h5.2v6h1.6v-6H18v-2l-2-2z"/>
            </svg>
          </button>
          <button v-if="isAuthor || isAdmin" class="post-action-btn" @click="$emit('edit', post)" title="Редактировать">
            <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="1.8">
              <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
              <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
            </svg>
          </button>
          <button class="post-action-btn danger" @click="$emit('delete', post.id)" title="Удалить">
            <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="1.8">
              <path d="M3 6h18"/>
              <path d="M8 6V4.8c0-.99.81-1.8 1.8-1.8h4.4c.99 0 1.8.81 1.8 1.8V6"/>
              <path d="M18.2 6l-.72 11.02A2 2 0 0 1 15.48 19H8.52a2 2 0 0 1-1.99-1.98L5.8 6"/>
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- Lightbox -->
    <teleport to="body">
      <div v-if="lightboxUrl" class="lightbox" @click="lightboxUrl = null">
        <img :src="lightboxUrl" class="lightbox-img" @click.stop />
        <button class="lightbox-close" @click="lightboxUrl = null">✕</button>
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
      showActions: false,
      isPlaying: false,
      currentTime: 0,
      duration: 0,
      playbackSpeed: 1,
      lightboxUrl: null,
    }
  },

  computed: {
    isAuthor() { return String(this.post.author_id) === String(this.currentUserId) },
    canEdit()  { return this.isAuthor || this.isAdmin },

    postType() {
      // Явный тип из поля type
      const t = this.post.type || ''
      if (t === 'voice') return 'voice'
      if (t === 'image') return 'image'
      if (t === 'video') return 'video'
      if (t === 'audio') return 'audio'
      if (t === 'file')  return 'file'

      // По media_url + media_type
      if (this.post.media_url) {
        if (this.post.media_type === 'image') return 'image'
        if (this.post.media_type === 'video') return 'video'
        if (this.post.media_type === 'audio') return 'audio'
        return 'file'
      }

      // По content — если это URL, определяем по паттерну
      const c = this.post.content || ''
      if (!c.startsWith('http')) return 'text'

      if (
        c.includes('/voice/') ||
        c.includes('/uploads/voice/') ||
        c.match(/\.(webm|ogg|mp3|opus)(\?|$)/i)
      ) return 'voice'

      if (
        c.includes('/media/images/') ||
        c.includes('/uploads/media/images/') ||
        c.match(/\.(png|jpg|jpeg|gif|webp|avif)(\?|$)/i)
      ) return 'image'

      if (
        c.includes('/media/videos/') ||
        c.includes('/uploads/media/videos/') ||
        c.match(/\.(mp4|mov|avi|mkv)(\?|$)/i)
      ) return 'video'

      if (c.includes('/media/audio/')) return 'audio'
      if (c.includes('/media/files/')) return 'file'

      return 'text'
    },

    mediaUrl() {
      return this.post.media_url || this.post.content || ''
    },

    waveformBars() {
      const waveform = this.post.waveform || []
      if (waveform.length) {
        return waveform.slice(0, 44).map(v => Math.max(3, Math.round(v * 32)))
      }
      return Array.from({ length: 44 }, (_, i) => {
        const v = Math.sin(i * 0.45) * 0.38 + 0.52 + Math.sin(i * 1.1) * 0.1
        return Math.max(3, Math.round(v * 28))
      })
    },

    activeBars() {
      if (!this.duration) return 0
      return Math.round((this.currentTime / this.duration) * this.waveformBars.length)
    },

    currentTimeText() { return this.formatTime(this.currentTime) },
    durationText()    { return this.formatTime(this.duration || this.post.duration || 0) },

    timeText() {
      if (!this.post.created_at) return ''
      const d = new Date(this.post.created_at)
      return d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
    }
  },

  methods: {
    isUrl(str) {
      return typeof str === 'string' && /^https?:\/\//i.test(str.trim())
    },

    formatTime(s) {
      const sec = Math.floor(s || 0)
      return `${Math.floor(sec / 60).toString().padStart(2, '0')}:${(sec % 60).toString().padStart(2, '0')}`
    },

    togglePlay() {
      const audio = this.$refs.audio
      if (!audio) return
      if (this.isPlaying) { audio.pause() } else { audio.play() }
      this.isPlaying = !this.isPlaying
    },

    cycleSpeed() {
      const speeds = [1, 1.5, 2]
      const idx = speeds.indexOf(this.playbackSpeed)
      this.playbackSpeed = speeds[(idx + 1) % speeds.length]
      const audio = this.$refs.audio
      if (audio) audio.playbackRate = this.playbackSpeed
    },

    seek(e) {
      const audio = this.$refs.audio
      if (!audio) return
      const rect = e.currentTarget.getBoundingClientRect()
      const ratio = Math.max(0, Math.min(1, (e.clientX - rect.left) / rect.width))
      audio.currentTime = ratio * (audio.duration || 0)
    },

    onTimeUpdate() {
      const audio = this.$refs.audio
      if (audio) this.currentTime = audio.currentTime
    },

    onEnded() { this.isPlaying = false; this.currentTime = 0 },

    onMeta() {
      const audio = this.$refs.audio
      if (!audio) return
      if (audio.duration && isFinite(audio.duration)) { this.duration = audio.duration; return }
      audio.currentTime = 1e101
      audio.addEventListener('timeupdate', () => {
        if (isFinite(audio.duration)) { this.duration = audio.duration; audio.currentTime = 0 }
      }, { once: true })
    },
  }
}
</script>

<style scoped>
.channel-post {
  padding: 14px 16px;
  border-radius: 16px;
  background: rgba(10, 14, 32, 0.95);
  border: 1px solid rgba(255, 255, 255, 0.06);
  position: relative;
  transition: border-color 0.15s;
  font-family: 'Inter', 'Satoshi', sans-serif;
}
.channel-post:hover { border-color: rgba(99, 102, 241, 0.22); }
.channel-post.pinned { border-color: rgba(99, 102, 241, 0.35); background: rgba(99, 102, 241, 0.07); }
.channel-post.theme-light { background: #ffffff; border-color: #e8eaf0; }
.channel-post.theme-light:hover { border-color: rgba(99, 102, 241, 0.25); }
.channel-post.theme-light.pinned { background: rgba(99, 102, 241, 0.04); border-color: rgba(99, 102, 241, 0.2); }

.pin-badge {
  display: inline-flex; align-items: center; gap: 5px;
  font-size: 10px; font-weight: 700; color: #6366f1;
  letter-spacing: 0.05em; text-transform: uppercase;
  background: rgba(99, 102, 241, 0.1); border: 1px solid rgba(99, 102, 241, 0.2);
  padding: 3px 8px; border-radius: 999px; margin-bottom: 10px;
}

.post-text {
  color: #c8d0f0; font-size: 14.5px; line-height: 1.65; font-weight: 400;
  white-space: pre-wrap; word-break: break-word; margin: 0 0 10px;
}
.theme-light .post-text { color: #1a1d2e; }

.post-media-img {
  width: 100%; max-height: 420px; object-fit: cover;
  display: block; cursor: zoom-in; border-radius: 12px; margin-bottom: 8px;
}
.post-media-video { width: 100%; border-radius: 12px; display: block; margin-bottom: 8px; max-height: 380px; }
.post-media-audio { width: 100%; margin-bottom: 8px; }

.post-file {
  display: inline-flex; align-items: center; gap: 8px; margin-bottom: 10px;
  color: #6366f1; font-size: 13px; text-decoration: none;
  padding: 8px 14px; border-radius: 10px;
  background: rgba(99, 102, 241, 0.1); border: 1px solid rgba(99, 102, 241, 0.2);
  transition: background 0.15s;
}
.post-file:hover { background: rgba(99, 102, 241, 0.18); }

/* Voice */
.voice-player {
  display: flex; align-items: center; gap: 12px; margin-bottom: 10px; padding: 2px 0;
}
.voice-play-btn {
  width: 44px; height: 44px; border-radius: 50%; flex-shrink: 0;
  display: grid; place-items: center;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  border: none; cursor: pointer; color: #fff;
  box-shadow: 0 6px 14px rgba(99, 102, 241, 0.35);
  transition: transform 0.15s;
}
.voice-play-btn:hover { transform: scale(1.06); }
.voice-waveform { flex: 1; cursor: pointer; min-width: 0; }
.voice-wave-bars { display: flex; align-items: center; gap: 2px; height: 36px; margin-bottom: 3px; }
.wave-bar {
  width: 3px; border-radius: 2px; flex-shrink: 0;
  background: rgba(99, 102, 241, 0.22); min-height: 3px; transition: background 0.08s;
}
.wave-bar.active { background: #6366f1; }
.voice-times { display: flex; justify-content: space-between; font-size: 11px; color: #5d6888; font-weight: 500; }
.voice-speed-btn {
  flex-shrink: 0; font-size: 11px; font-weight: 700; color: #6366f1;
  background: rgba(99, 102, 241, 0.1); border: 1px solid rgba(99, 102, 241, 0.2);
  padding: 4px 8px; border-radius: 6px; cursor: pointer; transition: all 0.15s; font-family: inherit;
}
.voice-speed-btn:hover { background: rgba(99, 102, 241, 0.2); }

/* Footer */
.post-footer { display: flex; align-items: center; justify-content: space-between; gap: 10px; margin-top: 6px; }
.post-time { color: #5d6888; font-size: 11px; font-weight: 500; }
.post-right { display: flex; align-items: center; gap: 8px; }
.post-views { display: flex; align-items: center; gap: 4px; color: #5d6888; font-size: 11px; font-weight: 500; }
.post-actions { display: flex; gap: 3px; opacity: 0; transition: opacity 0.15s; }
.post-actions.visible { opacity: 1; }
.post-action-btn {
  width: 26px; height: 26px; border-radius: 8px; display: grid; place-items: center;
  color: #a6afd4; background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.06); cursor: pointer; transition: all 0.15s;
}
.post-action-btn:hover { background: rgba(99, 102, 241, 0.15); color: #6366f1; border-color: rgba(99, 102, 241, 0.3); }
.post-action-btn.danger:hover { background: rgba(255, 77, 109, 0.12); color: #ff4d6d; border-color: rgba(255, 77, 109, 0.25); }

/* Lightbox */
.lightbox { position: fixed; inset: 0; z-index: 9999; background: rgba(0,0,0,0.92); display: flex; align-items: center; justify-content: center; cursor: zoom-out; }
.lightbox-img { max-width: 90vw; max-height: 90vh; border-radius: 8px; object-fit: contain; cursor: default; }
.lightbox-close { position: absolute; top: 20px; right: 20px; background: rgba(255,255,255,0.1); border: none; color: white; font-size: 20px; width: 40px; height: 40px; border-radius: 50%; cursor: pointer; }

@media (max-width: 760px) {
  .channel-post { padding: 12px 14px; border-radius: 14px; }
  .voice-play-btn { width: 38px; height: 38px; }
  .post-text { font-size: 14px; }
}
</style>