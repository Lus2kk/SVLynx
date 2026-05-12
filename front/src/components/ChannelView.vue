<template>
  <section class="channel-view" :class="{ 'theme-light': isLight }">
    <!-- Header -->
    <header class="channel-header">
      <div class="channel-info">
        <button v-if="showBackButton" class="back-btn" type="button" @click="$emit('back')">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.2">
            <path d="M15 18l-6-6 6-6"/>
          </svg>
        </button>
        <div class="channel-avatar" :style="!channel.avatar_url ? { background: channel.avatar_color || 'linear-gradient(135deg, #6572ff, #8a67ff)' } : {}">
          <img v-if="channel.avatar_url" :src="channel.avatar_url" alt="" />
          <span v-else>{{ channel.name?.[0]?.toUpperCase() || '#' }}</span>
        </div>
        <div class="channel-meta">
          <div class="channel-name">{{ channel.name }}</div>
          <div class="channel-handle">@{{ channel.handle }} · {{ channel.member_count }} subscribers</div>
        </div>
      </div>
      <div class="channel-actions">
        <!-- Кнопка подписки — показывается только не-членам -->
        <button
          v-if="!isSubscribed && !isEditor"
          class="subscribe-btn"
          :class="{ loading: subscribing }"
          type="button"
          @click="subscribeToChannel"
          :disabled="subscribing"
        >
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/>
            <path d="M13.73 21a2 2 0 0 1-3.46 0"/>
          </svg>
          Subscribe
        </button>

        <button v-if="isAdmin" class="icon-btn" title="Invite link" @click="openInvite">
          <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="1.8">
            <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/>
            <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>
          </svg>
        </button>
        <button class="icon-btn" :class="{ active: searchOpen }" title="Search" @click="searchOpen = !searchOpen">
          <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="1.8">
            <circle cx="11" cy="11" r="7"/><path d="M20 20l-3.5-3.5"/>
          </svg>
        </button>
        <button v-if="isAdmin" class="icon-btn" title="Channel settings" @click="settingsOpen = true">
          <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="1.8">
            <circle cx="12" cy="12" r="3"/>
            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>
          </svg>
        </button>
      </div>

      <transition name="search-slide">
        <div v-if="searchOpen" class="search-bar">
          <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="1.8" class="search-icon">
            <circle cx="11" cy="11" r="7"/><path d="M20 20l-3.5-3.5"/>
          </svg>
          <input ref="searchInput" v-model="searchQuery" type="text" class="search-input"
            placeholder="Search posts..." @input="onSearchInput" @keydown.escape="searchOpen = false" />
          <button class="search-close-btn" @click="searchOpen = false; searchQuery = ''; searchResults = []">
            <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 6L6 18M6 6l12 12"/>
            </svg>
          </button>
        </div>
      </transition>
    </header>

    <!-- Posts -->
    <div class="posts-area" ref="postsArea" @scroll="onScroll">
      <div v-if="loading" class="feed-state"><div class="spinner"></div></div>

      <template v-else>
        <div v-if="pinnedPosts.length" class="pinned-section">
          <div class="section-label">
            <svg viewBox="0 0 24 24" width="12" height="12" fill="currentColor"><path d="M16 12V4h1V2H7v2h1v8l-2 2v2h5.2v6h1.6v-6H18v-2l-2-2z"/></svg>
            Pinned
          </div>
          <ChannelPost
            v-for="post in pinnedPosts" :key="'pinned-' + post.id"
            :post="post" :isAdmin="isAdmin" :isLight="isLight" :currentUserId="currentUserId"
            @delete="deletePost" @pin="togglePin" @edit="startEdit"
          />
        </div>

        <template v-if="displayPosts.length">
          <ChannelPost
            v-for="post in displayPosts" :key="post.id"
            :post="post" :isAdmin="isAdmin" :isLight="isLight" :currentUserId="currentUserId"
            @delete="deletePost" @pin="togglePin" @edit="startEdit"
          />
          <div v-if="loadingMore" class="feed-state"><div class="spinner"></div></div>
        </template>

        <div v-else-if="!pinnedPosts.length" class="feed-state empty">
          <svg viewBox="0 0 24 24" width="32" height="32" fill="none" stroke="currentColor" stroke-width="1.4">
            <path d="M12 20h9"/><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"/>
          </svg>
          <span>No posts yet</span>
        </div>
      </template>
    </div>

    <!-- Composer (owner/admin/editor only) -->
    <div v-if="isEditor" class="composer-wrap">
      <div v-if="editingPost" class="edit-banner">
        <span>Editing post</span>
        <button @click="cancelEdit">✕</button>
      </div>
      <form class="composer" @submit.prevent="submitPost">
        <input
          v-model="newPostContent"
          type="text"
          class="post-input"
          placeholder="Write a post..."
          ref="postInput"
        />

        <button v-show="!voiceMode" type="button" class="send-btn" @click="onSendClick">
          <svg viewBox="0 0 24 24" width="17" height="17" fill="currentColor">
            <path d="M21.8 2.2a1 1 0 0 0-1.04-.23L2.76 8.97a1 1 0 0 0 .08 1.89l7.14 2.38 2.38 7.14a1 1 0 0 0 .91.68h.05a1 1 0 0 0 .9-.59l7-18a1 1 0 0 0-.22-1.03z"/>
          </svg>
        </button>

        <button
          v-show="voiceMode"
          ref="voiceBtn"
          type="button"
          class="send-btn"
          :class="{ recording: isRecordingVoice }"
          title="Hold to record"
        >
          <svg v-if="!isRecordingVoice" viewBox="0 0 24 24" width="17" height="17" fill="currentColor">
            <path d="M12 1a4 4 0 0 1 4 4v6a4 4 0 0 1-8 0V5a4 4 0 0 1 4-4z"/>
            <path d="M19 10a7 7 0 0 1-14 0H3a9 9 0 0 0 18 0h-2z"/>
            <line x1="12" y1="19" x2="12" y2="23" stroke="white" stroke-width="2"/>
            <line x1="8" y1="23" x2="16" y2="23" stroke="white" stroke-width="2"/>
          </svg>
          <svg v-else viewBox="0 0 24 24" width="17" height="17" fill="currentColor">
            <rect x="6" y="6" width="12" height="12" rx="2"/>
          </svg>
          <span v-if="isRecordingVoice" class="rec-timer">{{ voiceTimerText }}</span>
        </button>
      </form>
    </div>

    <!-- Settings modal -->
    <div v-if="settingsOpen" class="modal-overlay" @click.self="settingsOpen = false">
      <div class="settings-modal">
        <h3>Channel Settings</h3>
        <div class="settings-field"><label>Name</label><input v-model="editName" type="text" /></div>
        <div class="settings-field"><label>Handle</label><input v-model="editHandle" type="text" /></div>
        <div class="settings-field"><label>Description</label><textarea v-model="editDescription" rows="3"></textarea></div>
        <div class="settings-field">
          <label>Type</label>
          <select v-model="editType">
            <option value="public">Public</option>
            <option value="private">Private</option>
          </select>
        </div>
        <div class="modal-actions">
          <button class="btn-cancel" @click="settingsOpen = false">Cancel</button>
          <button class="btn-save" @click="saveSettings">Save</button>
        </div>
      </div>
    </div>

    <!-- Invite modal -->
    <div v-if="inviteOpen" class="modal-overlay" @click.self="inviteOpen = false">
      <div class="invite-modal">
        <div class="invite-modal-header">
          <h3>Invite Link</h3>
          <button class="invite-close" @click="inviteOpen = false">
            <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 6L6 18M6 6l12 12"/>
            </svg>
          </button>
        </div>
        <p class="invite-desc">Anyone with this link can join the channel.</p>
        <div v-if="inviteLoading" class="invite-loading"><div class="spinner"></div></div>
        <div v-else class="invite-link-wrap">
          <span class="invite-link-text">{{ inviteLink }}</span>
          <button class="invite-copy-btn" :class="{ copied: inviteCopied }" @click="copyInviteLink">
            <svg v-if="!inviteCopied" viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="1.8">
              <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
              <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
            </svg>
            <svg v-else viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="20 6 9 17 4 12"/>
            </svg>
            {{ inviteCopied ? 'Copied!' : 'Copy' }}
          </button>
        </div>
      </div>
    </div>
  </section>
</template>

<script>
import { apiFetch, getCookie } from '../api.js'
import ChannelPost from './ChannelPost.vue'

const BASE       = import.meta.env.VITE_API_URL       || 'http://localhost:8080'
const VOICE_BASE = import.meta.env.VITE_VOICE_API_URL || 'http://localhost:9090'

export default {
  name: 'ChannelView',
  components: { ChannelPost },

  props: {
    channel:       { type: Object,  required: true },
    currentUserId: { type: String,  default: null },
    userRole:      { type: String,  default: 'member' },
    isLight:       { type: Boolean, default: false },
    showBackButton:{ type: Boolean, default: false }
  },

  emits: ['back', 'channel-updated', 'post-created', 'post-deleted', 'post-pinned', 'post-edited', 'subscribe', 'unsubscribe'],

  data() {
    return {
      posts: [], pinnedPosts: [],
      loading: false, loadingMore: false, hasMore: true,
      newPostContent: '', editingPost: null,
      searchOpen: false, searchQuery: '', searchResults: [], searchDebounce: null,
      settingsOpen: false, editName: '', editHandle: '', editDescription: '', editType: 'public',
      inviteOpen: false, inviteLink: '', inviteLoading: false, inviteCopied: false,
      subscribing: false,
      // voice
      voiceMode: false,
      isRecordingVoice: false,
      voiceMediaRecorder: null,
      voiceChunks: [],
      voiceTimerSeconds: 0,
      voiceTimerInterval: null,
      waveformData: [],
      waveformInterval: null,
      audioContext: null,
      analyser: null,
    }
  },

  computed: {
    isAdmin()      { return ['owner', 'admin'].includes(this.userRole) },
    isEditor()     { return ['owner', 'admin', 'editor'].includes(this.userRole) },
    // Считается подписанным если userRole не пустой (member / editor / admin / owner)
    isSubscribed() { return !!this.userRole && this.userRole !== '' },
    displayPosts() {
      if (this.searchQuery.trim() && this.searchResults.length) return this.searchResults
      return this.posts
    },
    voiceTimerText() {
      const m = Math.floor(this.voiceTimerSeconds / 60).toString().padStart(2, '0')
      const s = (this.voiceTimerSeconds % 60).toString().padStart(2, '0')
      return `${m}:${s}`
    }
  },

  async mounted() {
    await this.loadPosts()
    await this.loadPinnedPosts()
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
      stream.getTracks().forEach(t => t.stop())
    } catch {}
  },

  beforeUnmount() {
    this._removeBtnListeners()
  },

  watch: {
    channel(val) {
      if (val) {
        this.posts = []
        this.pinnedPosts = []
        this.loadPosts()
        this.loadPinnedPosts()
      }
    },
    settingsOpen(val) { if (val) this.initEditFields() },
    voiceMode(val) {
      this._removeBtnListeners()
      if (val) this.$nextTick(() => this._addBtnListeners())
    }
  },

  methods: {
    initEditFields() {
      this.editName = this.channel.name || ''
      this.editHandle = this.channel.handle || ''
      this.editDescription = this.channel.description || ''
      this.editType = this.channel.type || 'public'
    },

    // ─── Subscribe ───────────────────────────────────────────────────
    async subscribeToChannel() {
      this.subscribing = true
      try {
        const res = await apiFetch(`${BASE}/channels/${this.channel.id}/subscribe`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ user_id: this.currentUserId })
        })
        if (!res.ok) return
        // Запросить разрешение на push-уведомления
        if ('Notification' in window && Notification.permission === 'default') {
          await Notification.requestPermission()
        }
        this.$emit('subscribe', this.channel)
      } catch (e) {
        console.error('subscribe error', e)
      } finally {
        this.subscribing = false
      }
    },

    // ─── Load posts ───────────────────────────────────────────────────
    async loadPosts() {
      if (!this.channel?.id) return
      this.loading = true
      try {
        const url = new URL(`${BASE}/channels/${this.channel.id}/posts`)
        url.searchParams.set('user_id', this.currentUserId)
        url.searchParams.set('limit', '50')
        const res = await apiFetch(url.toString())
        if (!res.ok) return
        const data = await res.json()
        this.posts = (data.posts || []).reverse()
        this.hasMore = (data.posts || []).length === 50
      } catch (e) { console.error('loadPosts error', e) }
      finally { this.loading = false }
    },

    async loadPinnedPosts() {
      if (!this.channel?.id) return
      try {
        const url = new URL(`${BASE}/channels/${this.channel.id}/posts/pinned`)
        url.searchParams.set('user_id', this.currentUserId)
        const res = await apiFetch(url.toString())
        if (!res.ok) return
        const data = await res.json()
        this.pinnedPosts = data.posts || []
      } catch (e) { console.error('loadPinnedPosts error', e) }
    },

    async loadMorePosts() {
      if (!this.hasMore || this.loadingMore || !this.posts.length) return
      this.loadingMore = true
      try {
        const oldest = this.posts[0]?.created_at
        const url = new URL(`${BASE}/channels/${this.channel.id}/posts`)
        url.searchParams.set('user_id', this.currentUserId)
        url.searchParams.set('before', oldest)
        url.searchParams.set('limit', '50')
        const res = await apiFetch(url.toString())
        if (!res.ok) return
        const data = await res.json()
        const older = (data.posts || []).reverse()
        if (!older.length) { this.hasMore = false; return }
        this.posts = [...older, ...this.posts]
      } catch (e) { console.error('loadMorePosts error', e) }
      finally { this.loadingMore = false }
    },

    onScroll() {
      const el = this.$refs.postsArea
      if (!el) return
      if (el.scrollTop < 100 && this.hasMore && !this.loadingMore) this.loadMorePosts()
    },

    // ─── Post actions ─────────────────────────────────────────────────
    onSendClick() {
      if (this.newPostContent.trim()) { this.submitPost() }
      else { this.voiceMode = !this.voiceMode }
    },

    async submitPost() {
      const content = this.newPostContent.trim()
      if (!content) return
      if (this.editingPost) { await this.updatePost(content) }
      else { await this.createPost(content) }
      this.newPostContent = ''
      this.editingPost = null
    },

    async createPost(content) {
      try {
        const res = await apiFetch(`${BASE}/channels/${this.channel.id}/posts`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ author_id: this.currentUserId, content })
        })
        if (!res.ok) return
        const data = await res.json()
        this.posts.push(data.post)
        this.$nextTick(() => {
          const el = this.$refs.postsArea
          if (el) el.scrollTop = el.scrollHeight
        })
        // Уведомляем ChatLayout — он рассылает по WS
        this.$emit('post-created', data.post)
      } catch (e) { console.error('createPost error', e) }
    },

    async updatePost(content) {
      try {
        const res = await apiFetch(`${BASE}/channels/${this.channel.id}/posts/${this.editingPost.id}`, {
          method: 'PATCH',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ editor_id: this.currentUserId, content })
        })
        if (!res.ok) return
        const data = await res.json()
        const idx = this.posts.findIndex(p => p.id === data.post.id)
        if (idx !== -1) this.posts[idx] = data.post
        this.$emit('post-edited', data.post)
      } catch (e) { console.error('updatePost error', e) }
    },

    async deletePost(postId) {
      try {
        const url = new URL(`${BASE}/channels/${this.channel.id}/posts/${postId}`)
        url.searchParams.set('user_id', this.currentUserId)
        const res = await apiFetch(url.toString(), { method: 'DELETE' })
        if (!res.ok) return
        this.posts = this.posts.filter(p => p.id !== postId)
        this.pinnedPosts = this.pinnedPosts.filter(p => p.id !== postId)
        this.$emit('post-deleted', postId)
      } catch (e) { console.error('deletePost error', e) }
    },

    async togglePin(post) {
      try {
        const res = await apiFetch(`${BASE}/channels/${this.channel.id}/posts/${post.id}/pin`, {
          method: 'PATCH',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ user_id: this.currentUserId, pinned: !post.pinned })
        })
        if (!res.ok) return
        const idx = this.posts.findIndex(p => p.id === post.id)
        if (idx !== -1) this.posts[idx] = { ...this.posts[idx], pinned: !post.pinned }
        await this.loadPinnedPosts()
        this.$emit('post-pinned', { ...post, pinned: !post.pinned })
      } catch (e) { console.error('togglePin error', e) }
    },

    startEdit(post) {
      this.editingPost = post
      this.newPostContent = post.content
      this.$nextTick(() => this.$refs.postInput?.focus())
    },

    cancelEdit() { this.editingPost = null; this.newPostContent = '' },

    // ─── Search ───────────────────────────────────────────────────────
    onSearchInput() {
      clearTimeout(this.searchDebounce)
      if (!this.searchQuery.trim()) { this.searchResults = []; return }
      this.searchDebounce = setTimeout(() => this.doSearch(), 400)
    },

    async doSearch() {
      try {
        const url = new URL(`${BASE}/channels/${this.channel.id}/posts/search`)
        url.searchParams.set('user_id', this.currentUserId)
        url.searchParams.set('q', this.searchQuery.trim())
        const res = await apiFetch(url.toString())
        if (!res.ok) return
        const data = await res.json()
        this.searchResults = (data.posts || []).reverse()
      } catch (e) { console.error('search error', e) }
    },

    // ─── Settings ─────────────────────────────────────────────────────
    async saveSettings() {
      try {
        const url = new URL(`${BASE}/channels/${this.channel.id}`)
        url.searchParams.set('user_id', this.currentUserId)
        const res = await apiFetch(url.toString(), {
          method: 'PATCH',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ name: this.editName, handle: this.editHandle, description: this.editDescription, type: this.editType })
        })
        if (!res.ok) return
        const data = await res.json()
        this.$emit('channel-updated', data.channel)
        this.settingsOpen = false
      } catch (e) { console.error('saveSettings error', e) }
    },

    // ─── Invite ───────────────────────────────────────────────────────
    async openInvite() {
      this.inviteOpen = true; this.inviteLink = ''; this.inviteCopied = false; this.inviteLoading = true
      try {
        const res = await apiFetch(`${BASE}/channels/${this.channel.id}/invites`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ created_by: this.currentUserId })
        })
        if (!res.ok) return
        const data = await res.json()
        const token = data.invite?.token || ''
        this.inviteLink = `${import.meta.env.VITE_APP_URL || window.location.origin}/invite/${token}`
      } catch (e) { console.error('openInvite error', e) }
      finally { this.inviteLoading = false }
    },

    async copyInviteLink() {
      if (!this.inviteLink) return
      try {
        await navigator.clipboard.writeText(this.inviteLink)
        this.inviteCopied = true
        setTimeout(() => { this.inviteCopied = false }, 2000)
      } catch {}
    },

    // ─── WS handlers (called from ChatLayout via ref) ────────────────
    handleNewPost(post) {
      if (this.posts.find(p => p.id === post.id)) return
      this.posts.push(post)
      this.$nextTick(() => { const el = this.$refs.postsArea; if (el) el.scrollTop = el.scrollHeight })
    },
    handleDeletePost(postId) {
      this.posts = this.posts.filter(p => String(p.id) !== String(postId))
      this.pinnedPosts = this.pinnedPosts.filter(p => String(p.id) !== String(postId))
    },
    handleUpdatePost(post) {
      const idx = this.posts.findIndex(p => String(p.id) === String(post.id))
      if (idx !== -1) this.posts[idx] = { ...this.posts[idx], ...post }
    },
    handlePinPost(post) {
      const idx = this.posts.findIndex(p => String(p.id) === String(post.id))
      if (idx !== -1) this.posts[idx] = { ...this.posts[idx], pinned: post.pinned }
      // Обновляем pinnedPosts
      if (post.pinned) {
        if (!this.pinnedPosts.find(p => String(p.id) === String(post.id))) {
          this.pinnedPosts.push(post)
        }
      } else {
        this.pinnedPosts = this.pinnedPosts.filter(p => String(p.id) !== String(post.id))
      }
    },

    // ─── Voice ───────────────────────────────────────────────────────
    _addBtnListeners() {
      const btn = this.$refs.voiceBtn
      if (!btn) return
      this._onPressDown = () => {
        this._isLongPress = false; clearTimeout(this._pressTimer)
        this._pressTimer = setTimeout(() => { this._isLongPress = true; this.startVoice() }, 300)
      }
      this._onPressUp = () => {
        clearTimeout(this._pressTimer)
        if (this.isRecordingVoice) { this.stopVoice() }
        else if (!this._isLongPress) { this.voiceMode = false }
        this._isLongPress = false
      }
      this._onTouchStart = (e) => {
        e.preventDefault(); this._isLongPress = false; clearTimeout(this._pressTimer)
        this._pressTimer = setTimeout(() => { this._isLongPress = true; this.startVoice() }, 300)
      }
      this._onTouchEnd = (e) => {
        e.preventDefault(); clearTimeout(this._pressTimer)
        if (this.isRecordingVoice) { this.stopVoice() }
        else if (!this._isLongPress) { this.voiceMode = false }
        this._isLongPress = false
      }
      this._onMouseLeave = () => { clearTimeout(this._pressTimer); if (this.isRecordingVoice) this.stopVoice(); this._isLongPress = false }
      btn.addEventListener('mousedown', this._onPressDown)
      btn.addEventListener('mouseup', this._onPressUp)
      btn.addEventListener('mouseleave', this._onMouseLeave)
      btn.addEventListener('touchstart', this._onTouchStart, { passive: false })
      btn.addEventListener('touchend', this._onTouchEnd, { passive: false })
      btn.addEventListener('touchcancel', this._onTouchEnd, { passive: false })
    },

    _removeBtnListeners() {
      const btn = this.$refs.voiceBtn
      if (!btn) return
      if (this._onPressDown)  btn.removeEventListener('mousedown',  this._onPressDown)
      if (this._onPressUp)    btn.removeEventListener('mouseup',    this._onPressUp)
      if (this._onMouseLeave) btn.removeEventListener('mouseleave', this._onMouseLeave)
      if (this._onTouchStart) btn.removeEventListener('touchstart', this._onTouchStart)
      if (this._onTouchEnd)   { btn.removeEventListener('touchend', this._onTouchEnd); btn.removeEventListener('touchcancel', this._onTouchEnd) }
    },

    async startVoice() {
      if (this.isRecordingVoice) return
      try {
        const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
        this.voiceChunks = []; this.waveformData = []
        this.audioContext = new AudioContext()
        this.analyser = this.audioContext.createAnalyser()
        this.analyser.fftSize = 256
        this.audioContext.createMediaStreamSource(stream).connect(this.analyser)
        const mimeType = MediaRecorder.isTypeSupported('audio/webm;codecs=opus') ? 'audio/webm;codecs=opus'
          : MediaRecorder.isTypeSupported('audio/mp4') ? 'audio/mp4' : 'audio/webm'
        this.voiceMediaRecorder = new MediaRecorder(stream, { mimeType })
        this.voiceMediaRecorder.ondataavailable = e => { if (e.data.size > 0) this.voiceChunks.push(e.data) }
        this.voiceMediaRecorder.onstop = this.handleVoiceStop
        this.voiceMediaRecorder.start()
        this.isRecordingVoice = true
        this.voiceTimerSeconds = 0
        this.voiceTimerInterval = setInterval(() => this.voiceTimerSeconds++, 1000)
        this.waveformInterval = setInterval(() => {
          const arr = new Uint8Array(this.analyser.frequencyBinCount)
          this.analyser.getByteFrequencyData(arr)
          this.waveformData.push(arr.reduce((a, b) => a + b, 0) / arr.length / 255)
        }, 100)
      } catch (e) { console.error('Microphone error', e) }
    },

    stopVoice() {
      if (!this.voiceMediaRecorder || !this.isRecordingVoice) return
      this.isRecordingVoice = false
      clearInterval(this.voiceTimerInterval); clearInterval(this.waveformInterval)
      if (this.audioContext) { this.audioContext.close(); this.audioContext = null }
      this.voiceMediaRecorder.requestData()
      this.voiceMediaRecorder.stream.getTracks().forEach(t => t.stop())
      this.voiceMediaRecorder.stop()
    },

    async handleVoiceStop() {
      if (!this.voiceChunks.length) return
      const mimeType = MediaRecorder.isTypeSupported('audio/webm;codecs=opus') ? 'audio/webm;codecs=opus'
        : MediaRecorder.isTypeSupported('audio/mp4') ? 'audio/mp4' : 'audio/webm'
      const blob = new Blob(this.voiceChunks, { type: mimeType })
      const ext  = mimeType.includes('mp4') ? '.mp4' : '.webm'
      const form = new FormData()
      form.append('file', blob, `voice_${Date.now()}${ext}`)
      form.append('chat_id', String(this.channel.id))
      form.append('sender_id', String(this.currentUserId))
      form.append('recipient_id', String(this.currentUserId))
      form.append('waveform', JSON.stringify(this.waveformData))
      form.append('duration', String(this.voiceTimerSeconds))
      try {
        const res = await fetch(`${VOICE_BASE}/voice/upload`, {
          method: 'POST',
          headers: { Authorization: `Bearer ${getCookie('access_token') || ''}` },
          body: form
        })
        if (!res.ok) return
        const data = await res.json()
        if (data.message?.content) {
          await this.createPost(data.message.content)
        }
      } catch (e) { console.error('Voice upload error', e) }
      finally { this.voiceMode = false }
    }
  }
}
</script>

<style scoped>
.channel-view {
  height: 100%; display: flex; flex-direction: column;
  min-width: 0; overflow: hidden;
  background: linear-gradient(180deg, rgba(10,13,28,0.72), rgba(7,10,22,0.82));
}
.channel-view.theme-light { background: #f5f6fc; }

.channel-header {
  flex-shrink: 0; height: 78px; min-height: 78px; padding: 14px 20px;
  display: flex; align-items: center; justify-content: space-between;
  border-bottom: 1px solid rgba(255,255,255,0.05);
  background: rgba(255,255,255,0.015); position: relative;
}
.theme-light .channel-header { background: #fff; border-bottom-color: #e4e6f0; }

.channel-info { display: flex; align-items: center; gap: 12px; min-width: 0; }
.channel-avatar {
  width: 38px; height: 38px; border-radius: 12px; flex-shrink: 0;
  display: grid; place-items: center; overflow: hidden;
  color: #fff; font-weight: 700; font-size: 14px;
}
.channel-avatar img { width: 100%; height: 100%; object-fit: cover; }
.channel-meta { min-width: 0; }
.channel-name { color: #f2f4ff; font-size: 14px; font-weight: 700; }
.theme-light .channel-name { color: #1a1d2e; }
.channel-handle { color: #7d87ab; font-size: 11px; font-weight: 500; margin-top: 2px; }

.channel-actions { display: flex; gap: 8px; flex-shrink: 0; align-items: center; }
.icon-btn {
  width: 32px; height: 32px; border-radius: 10px;
  display: grid; place-items: center;
  color: #95a0c8; background: transparent; border: 1px solid transparent; cursor: pointer; transition: all 0.15s;
}
.icon-btn:hover, .icon-btn.active { background: rgba(110,121,255,0.15); border-color: rgba(110,121,255,0.3); color: #6e79ff; }
.back-btn {
  width: 34px; height: 34px; border-radius: 11px; display: grid; place-items: center; flex-shrink: 0;
  color: #a6afd4; background: rgba(255,255,255,0.04); border: 1px solid rgba(255,255,255,0.06); cursor: pointer;
}

/* Subscribe button */
.subscribe-btn {
  display: flex; align-items: center; gap: 6px;
  padding: 7px 14px; border-radius: 10px; border: none; cursor: pointer;
  font-size: 12px; font-weight: 700; font-family: inherit;
  color: #fff; background: linear-gradient(135deg, #6e79ff, #8669ff);
  box-shadow: 0 6px 16px rgba(94,102,255,0.3);
  transition: all 0.2s; white-space: nowrap;
}
.subscribe-btn:hover { transform: translateY(-1px); box-shadow: 0 8px 20px rgba(94,102,255,0.4); }
.subscribe-btn:disabled { opacity: 0.6; cursor: not-allowed; transform: none; }
.subscribe-btn.loading { opacity: 0.7; }

/* Search bar */
.search-bar {
  position: absolute; top: 0; left: 0; right: 0; bottom: 0;
  display: flex; align-items: center; gap: 8px; padding: 0 16px;
  background: rgba(13,17,32,0.98); backdrop-filter: blur(12px); z-index: 20;
}
.theme-light .search-bar { background: #fff; }
.search-icon { color: #6e79ff; flex-shrink: 0; }
.search-input { flex: 1; background: transparent; border: none; outline: none; color: #eef2ff; font-size: 14px; font-weight: 500; }
.theme-light .search-input { color: #1a1d2e; }
.search-input::placeholder { color: #4a5270; }
.search-close-btn { width: 26px; height: 26px; border-radius: 8px; display: grid; place-items: center; color: #a6afd4; background: transparent; border: none; cursor: pointer; }
.search-close-btn:hover { color: #ff4d6d; }
.search-slide-enter-active, .search-slide-leave-active { transition: opacity 0.2s, transform 0.2s; }
.search-slide-enter-from, .search-slide-leave-to { opacity: 0; transform: translateY(-6px); }

/* Posts area */
.posts-area {
  flex: 1; min-height: 0; overflow-y: auto; overflow-x: hidden;
  padding: 16px 28px; display: flex; flex-direction: column; gap: 4px;
  -webkit-overflow-scrolling: touch; overscroll-behavior: contain;
  background-image: linear-gradient(rgba(255,255,255,0.03) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.02) 1px, transparent 1px);
  background-size: 48px 48px;
}
.posts-area::-webkit-scrollbar { width: 6px; }
.posts-area::-webkit-scrollbar-thumb { background: rgba(148,159,212,0.16); border-radius: 999px; }

.pinned-section { display: flex; flex-direction: column; gap: 4px; margin-bottom: 8px; padding-bottom: 8px; border-bottom: 1px solid rgba(255,255,255,0.05); }
.section-label { display: flex; align-items: center; gap: 5px; color: #7d87ab; font-size: 10px; font-weight: 700; letter-spacing: 0.06em; text-transform: uppercase; margin-bottom: 4px; }
.feed-state { display: flex; align-items: center; justify-content: center; flex-direction: column; gap: 10px; padding: 40px 0; color: #7d87ab; font-size: 13px; }
.spinner { width: 22px; height: 22px; border-radius: 50%; border: 2px solid rgba(110,121,255,0.2); border-top-color: #6e79ff; animation: spin 0.7s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

/* Composer */
.composer-wrap {
  flex-shrink: 0; padding: 12px 28px 16px;
  background: linear-gradient(180deg, rgba(8,12,24,0.18), rgba(8,12,24,0.32));
}
.theme-light .composer-wrap { background: #f5f6fc; }
.edit-banner {
  display: flex; align-items: center; justify-content: space-between;
  padding: 6px 12px; margin-bottom: 6px; border-radius: 8px;
  background: rgba(110,121,255,0.1); border-left: 3px solid #6e79ff;
  color: #6e79ff; font-size: 12px; font-weight: 600;
}
.edit-banner button { background: none; border: none; color: #a6afd4; cursor: pointer; font-size: 14px; }
.composer {
  height: 56px; display: flex; align-items: center; gap: 12px;
  padding: 0 12px 0 16px; border-radius: 18px;
  border: 1px solid rgba(110,123,255,0.18);
  background: linear-gradient(180deg, rgba(25,30,58,0.68), rgba(18,23,46,0.78));
  box-shadow: 0 10px 30px rgba(0,0,0,0.18);
}
.theme-light .composer { background: #fff; border-color: rgba(91,106,255,0.2); }
.post-input {
  flex: 1; background: transparent; border: none; outline: none;
  color: #eef2ff; font-size: 16px; font-weight: 500;
}
.theme-light .post-input { color: #1a1d2e; }
.post-input::placeholder { color: #747ea2; }
.send-btn {
  width: 34px; height: 34px; border-radius: 11px; flex-shrink: 0;
  display: grid; place-items: center; cursor: pointer; border: none; position: relative;
  color: #fff; background: linear-gradient(135deg, #6e79ff, #8669ff);
  box-shadow: 0 8px 18px rgba(94,102,255,0.28);
  user-select: none; -webkit-user-select: none; touch-action: none;
}
.send-btn.recording { background: linear-gradient(135deg, #ff4d6d, #d93856); box-shadow: 0 0 0 4px rgba(255,77,109,0.2); animation: pulse 1s infinite; }
@keyframes pulse { 0%,100% { box-shadow: 0 0 0 4px rgba(255,77,109,0.2); } 50% { box-shadow: 0 0 0 8px rgba(255,77,109,0.1); } }
.rec-timer { position: absolute; top: -22px; left: 50%; transform: translateX(-50%); font-size: 10px; font-weight: 700; color: #ff4d6d; white-space: nowrap; background: rgba(0,0,0,0.5); padding: 2px 6px; border-radius: 6px; }

/* Modals */
.modal-overlay { position: fixed; inset: 0; background: rgba(5,8,18,0.6); backdrop-filter: blur(4px); display: grid; place-items: center; z-index: 100; }
.settings-modal { background: linear-gradient(180deg, rgba(22,28,52,0.97), rgba(16,20,38,0.99)); border: 1px solid rgba(132,144,224,0.15); border-radius: 18px; padding: 28px; width: 360px; box-shadow: 0 24px 48px rgba(0,0,0,0.4); }
.settings-modal h3 { color: #eef2ff; font-size: 16px; font-weight: 700; margin-bottom: 20px; }
.settings-field { margin-bottom: 14px; }
.settings-field label { display: block; color: #7d87ab; font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 6px; }
.settings-field input, .settings-field textarea, .settings-field select { width: 100%; background: rgba(255,255,255,0.04); border: 1px solid rgba(255,255,255,0.08); border-radius: 10px; padding: 9px 12px; color: #eef2ff; font-size: 13px; outline: none; font-family: inherit; box-sizing: border-box; resize: vertical; }
.settings-field textarea { min-height: 70px; }
.modal-actions { display: flex; gap: 10px; justify-content: flex-end; margin-top: 20px; }
.btn-cancel { padding: 9px 18px; border-radius: 10px; background: rgba(255,255,255,0.05); border: 1px solid rgba(255,255,255,0.06); color: #a6afd4; font-size: 13px; cursor: pointer; }
.btn-save { padding: 9px 18px; border-radius: 10px; background: linear-gradient(135deg, #6e79ff, #8669ff); border: none; color: #fff; font-size: 13px; font-weight: 600; cursor: pointer; }

.invite-modal { background: linear-gradient(180deg, rgba(22,28,52,0.97), rgba(16,20,38,0.99)); border: 1px solid rgba(132,144,224,0.15); border-radius: 18px; padding: 24px; width: 380px; box-shadow: 0 24px 48px rgba(0,0,0,0.4); }
.theme-light .invite-modal { background: #fff; border-color: #dde1f0; }
.invite-modal-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; }
.invite-modal-header h3 { color: #eef2ff; font-size: 15px; font-weight: 700; }
.theme-light .invite-modal-header h3 { color: #1a1d2e; }
.invite-close { width: 26px; height: 26px; border-radius: 8px; display: grid; place-items: center; color: #7d87ab; background: transparent; border: none; cursor: pointer; }
.invite-close:hover { color: #ff4d6d; }
.invite-desc { color: #7d87ab; font-size: 12px; margin-bottom: 16px; }
.invite-loading { display: flex; justify-content: center; padding: 16px 0; }
.invite-link-wrap { display: flex; align-items: center; gap: 8px; background: rgba(255,255,255,0.04); border: 1px solid rgba(255,255,255,0.08); border-radius: 11px; padding: 10px 12px; }
.theme-light .invite-link-wrap { background: #f5f6fc; border-color: #e4e6f0; }
.invite-link-text { flex: 1; min-width: 0; color: #6e79ff; font-size: 12px; font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.invite-copy-btn { display: flex; align-items: center; gap: 5px; flex-shrink: 0; padding: 6px 12px; border-radius: 8px; border: none; cursor: pointer; font-size: 12px; font-weight: 600; color: #6e79ff; background: rgba(110,121,255,0.12); transition: all 0.15s; }
.invite-copy-btn:hover { background: rgba(110,121,255,0.2); }
.invite-copy-btn.copied { color: #22c55e; background: rgba(34,197,94,0.12); }

@media (max-width: 760px) {
  .channel-header { height: 64px; min-height: 64px; padding: 10px 14px; }
  .posts-area { padding: 12px 14px; }
  .composer-wrap { padding: 8px 14px calc(8px + env(safe-area-inset-bottom)); }
  .composer { height: 50px; border-radius: 16px; }
  .settings-modal, .invite-modal { width: calc(100vw - 32px); }
}
</style>