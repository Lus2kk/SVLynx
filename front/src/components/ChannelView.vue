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
        <button class="icon-btn" title="Search" @click="searchOpen = !searchOpen" :class="{ active: searchOpen }">
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

      <!-- Search bar -->
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

    <!-- Posts feed -->
    <div class="posts-area" ref="postsArea" @scroll="onScroll">
      <div v-if="pinnedPosts.length" class="pinned-section">
        <div class="section-label">
          <svg viewBox="0 0 24 24" width="12" height="12" fill="currentColor"><path d="M16 12V4h1V2H7v2h1v8l-2 2v2h5.2v6h1.6v-6H18v-2l-2-2z"/></svg>
          Pinned
        </div>
        <ChannelPost
          v-for="post in pinnedPosts" :key="'pinned-' + post.id"
          :post="post" :isAdmin="isAdmin" :isLight="isLight"
          :currentUserId="currentUserId"
          @delete="deletePost" @pin="togglePin" @edit="startEdit"
        />
      </div>

      <div v-if="loading" class="feed-state">
        <div class="spinner"></div>
      </div>

      <template v-else-if="displayPosts.length">
        <ChannelPost
          v-for="post in displayPosts" :key="post.id"
          :post="post" :isAdmin="isAdmin" :isLight="isLight"
          :currentUserId="currentUserId"
          @delete="deletePost" @pin="togglePin" @edit="startEdit"
        />
        <div v-if="loadingMore" class="feed-state"><div class="spinner"></div></div>
      </template>

      <div v-else class="feed-state empty">
        <svg viewBox="0 0 24 24" width="32" height="32" fill="none" stroke="currentColor" stroke-width="1.4">
          <path d="M12 20h9"/><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"/>
        </svg>
        <span>No posts yet</span>
      </div>
    </div>

    <!-- Composer (only for admin/editor) -->
    <div v-if="isEditor" class="composer-wrap">
      <div v-if="editingPost" class="edit-banner">
        <span>Editing post</span>
        <button @click="cancelEdit">✕</button>
      </div>
      <form class="composer" @submit.prevent="submitPost">
        <textarea
          v-model="newPostContent"
          class="post-input"
          placeholder="Write a post..."
          rows="1"
          @input="autoResize"
          ref="postInput"
        ></textarea>
        <div class="composer-actions">
          <button type="submit" class="send-btn" :disabled="!newPostContent.trim()">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
              <path d="M21.8 2.2a1 1 0 0 0-1.04-.23L2.76 8.97a1 1 0 0 0 .08 1.89l7.14 2.38 2.38 7.14a1 1 0 0 0 .91.68h.05a1 1 0 0 0 .9-.59l7-18a1 1 0 0 0-.22-1.03z"/>
            </svg>
          </button>
        </div>
      </form>
    </div>

    <!-- Settings modal -->
    <div v-if="settingsOpen" class="modal-overlay" @click.self="settingsOpen = false">
      <div class="settings-modal">
        <h3>Channel Settings</h3>
        <div class="settings-field">
          <label>Name</label>
          <input v-model="editName" type="text" />
        </div>
        <div class="settings-field">
          <label>Handle</label>
          <input v-model="editHandle" type="text" />
        </div>
        <div class="settings-field">
          <label>Description</label>
          <textarea v-model="editDescription" rows="3"></textarea>
        </div>
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
  </section>
</template>

<script>
import { apiFetch, getCookie } from '../api.js'

const BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

// Inline ChannelPost component
const ChannelPost = {
  name: 'ChannelPost',
  props: {
    post: { type: Object, required: true },
    isAdmin: { type: Boolean, default: false },
    isLight: { type: Boolean, default: false },
    currentUserId: { type: String, default: null }
  },
  emits: ['delete', 'pin', 'edit'],
  data() { return { showActions: false } },
  computed: {
    isAuthor() { return String(this.post.author_id) === String(this.currentUserId) },
    canEdit() { return this.isAuthor || this.isAdmin },
    timeText() {
      if (!this.post.created_at) return ''
      const d = new Date(this.post.created_at)
      return d.toLocaleString('ru-RU', { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' })
    }
  },
  template: `
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
              <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
            </svg>
          </button>
          <button class="post-action-btn danger" @click="$emit('delete', post.id)" title="Delete">
            <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="1.8">
              <path d="M3 6h18"/><path d="M8 6V4.8c0-.99.81-1.8 1.8-1.8h4.4c.99 0 1.8.81 1.8 1.8V6"/>
              <path d="M18.2 6l-.72 11.02A2 2 0 0 1 15.48 19H8.52a2 2 0 0 1-1.99-1.98L5.8 6"/>
            </svg>
          </button>
        </div>
      </div>
    </article>
  `
}

export default {
  name: 'ChannelView',
  components: { ChannelPost },

  props: {
    channel: { type: Object, required: true },
    currentUserId: { type: String, default: null },
    userRole: { type: String, default: 'member' },
    isLight: { type: Boolean, default: false },
    showBackButton: { type: Boolean, default: false }
  },

  emits: ['back', 'channel-updated'],

  data() {
    return {
      posts: [],
      pinnedPosts: [],
      loading: false,
      loadingMore: false,
      hasMore: true,
      newPostContent: '',
      editingPost: null,
      searchOpen: false,
      searchQuery: '',
      searchResults: [],
      searchDebounce: null,
      settingsOpen: false,
      editName: '',
      editHandle: '',
      editDescription: '',
      editType: 'public',
    }
  },

  computed: {
    isAdmin() { return ['owner', 'admin'].includes(this.userRole) },
    isEditor() { return ['owner', 'admin', 'editor'].includes(this.userRole) },
    displayPosts() {
      if (this.searchQuery.trim() && this.searchResults.length) return this.searchResults
      return this.posts
    }
  },

  async mounted() {
    await this.loadPosts()
    await this.loadPinnedPosts()
    if (this.settingsOpen) this.initEditFields()
  },

  watch: {
    channel: {
      immediate: true,
      handler(val) {
        if (val) { this.loadPosts(); this.loadPinnedPosts() }
      }
    },
    settingsOpen(val) {
      if (val) this.initEditFields()
    }
  },

  methods: {
    initEditFields() {
      this.editName = this.channel.name || ''
      this.editHandle = this.channel.handle || ''
      this.editDescription = this.channel.description || ''
      this.editType = this.channel.type || 'public'
    },

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
      } catch (e) {
        console.error('loadPosts error', e)
      } finally {
        this.loading = false
      }
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
      } catch (e) {
        console.error('loadPinnedPosts error', e)
      }
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
      } catch (e) {
        console.error('loadMorePosts error', e)
      } finally {
        this.loadingMore = false
      }
    },

    onScroll() {
      const el = this.$refs.postsArea
      if (!el) return
      if (el.scrollTop < 100 && this.hasMore && !this.loadingMore) this.loadMorePosts()
    },

    async submitPost() {
      const content = this.newPostContent.trim()
      if (!content) return

      if (this.editingPost) {
        await this.updatePost(content)
      } else {
        await this.createPost(content)
      }
      this.newPostContent = ''
      this.editingPost = null
      this.$nextTick(() => { if (this.$refs.postInput) this.$refs.postInput.style.height = 'auto' })
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
      } catch (e) {
        console.error('createPost error', e)
      }
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
      } catch (e) {
        console.error('updatePost error', e)
      }
    },

    async deletePost(postId) {
      try {
        const url = new URL(`${BASE}/channels/${this.channel.id}/posts/${postId}`)
        url.searchParams.set('user_id', this.currentUserId)
        const res = await apiFetch(url.toString(), { method: 'DELETE' })
        if (!res.ok) return
        this.posts = this.posts.filter(p => p.id !== postId)
        this.pinnedPosts = this.pinnedPosts.filter(p => p.id !== postId)
      } catch (e) {
        console.error('deletePost error', e)
      }
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
      } catch (e) {
        console.error('togglePin error', e)
      }
    },

    startEdit(post) {
      this.editingPost = post
      this.newPostContent = post.content
      this.$nextTick(() => { this.$refs.postInput?.focus(); this.autoResize() })
    },

    cancelEdit() {
      this.editingPost = null
      this.newPostContent = ''
    },

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
      } catch (e) {
        console.error('search error', e)
      }
    },

    async saveSettings() {
      try {
        const url = new URL(`${BASE}/channels/${this.channel.id}`)
        url.searchParams.set('user_id', this.currentUserId)
        const res = await apiFetch(url.toString(), {
          method: 'PATCH',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            name: this.editName,
            handle: this.editHandle,
            description: this.editDescription,
            type: this.editType
          })
        })
        if (!res.ok) return
        const data = await res.json()
        this.$emit('channel-updated', data.channel)
        this.settingsOpen = false
      } catch (e) {
        console.error('saveSettings error', e)
      }
    },

    autoResize() {
      const el = this.$refs.postInput
      if (!el) return
      el.style.height = 'auto'
      el.style.height = Math.min(el.scrollHeight, 120) + 'px'
    },

    // Called from parent when WS event arrives
    handleNewPost(post) {
      if (this.posts.find(p => p.id === post.id)) return
      this.posts.push(post)
      this.$nextTick(() => {
        const el = this.$refs.postsArea
        if (el) el.scrollTop = el.scrollHeight
      })
    },

    handleDeletePost(postId) {
      this.posts = this.posts.filter(p => p.id !== postId)
      this.pinnedPosts = this.pinnedPosts.filter(p => p.id !== postId)
    },

    handleUpdatePost(post) {
      const idx = this.posts.findIndex(p => p.id === post.id)
      if (idx !== -1) this.posts[idx] = { ...this.posts[idx], ...post }
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
  flex-shrink: 0; height: 78px; min-height: 78px;
  padding: 14px 20px;
  display: flex; align-items: center; justify-content: space-between;
  border-bottom: 1px solid rgba(255,255,255,0.05);
  background: rgba(255,255,255,0.015);
  position: relative;
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

.channel-actions { display: flex; gap: 8px; flex-shrink: 0; }
.icon-btn {
  width: 32px; height: 32px; border-radius: 10px;
  display: grid; place-items: center;
  color: #95a0c8; background: transparent; border: 1px solid transparent; cursor: pointer;
  transition: all 0.15s;
}
.icon-btn:hover, .icon-btn.active { background: rgba(110,121,255,0.15); border-color: rgba(110,121,255,0.3); color: #6e79ff; }

.back-btn {
  width: 34px; height: 34px; border-radius: 11px;
  display: grid; place-items: center; flex-shrink: 0;
  color: #a6afd4; background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.06); cursor: pointer; transition: all 0.2s;
}

.search-bar {
  position: absolute; top: 0; left: 0; right: 0; bottom: 0;
  display: flex; align-items: center; gap: 8px; padding: 0 16px;
  background: rgba(13,17,32,0.98); backdrop-filter: blur(12px); z-index: 20;
}
.theme-light .search-bar { background: #fff; }
.search-icon { color: #6e79ff; flex-shrink: 0; }
.search-input {
  flex: 1; background: transparent; border: none; outline: none;
  color: #eef2ff; font-size: 14px; font-weight: 500;
}
.theme-light .search-input { color: #1a1d2e; }
.search-input::placeholder { color: #4a5270; }
.search-close-btn {
  width: 26px; height: 26px; border-radius: 8px;
  display: grid; place-items: center;
  color: #a6afd4; background: transparent; border: none; cursor: pointer;
}
.search-close-btn:hover { color: #ff4d6d; }
.search-slide-enter-active, .search-slide-leave-active { transition: opacity 0.2s, transform 0.2s; }
.search-slide-enter-from, .search-slide-leave-to { opacity: 0; transform: translateY(-6px); }

.posts-area {
  flex: 1; min-height: 0; overflow-y: auto; overflow-x: hidden;
  padding: 16px 24px; display: flex; flex-direction: column; gap: 12px;
  -webkit-overflow-scrolling: touch; overscroll-behavior: contain;
}
.posts-area::-webkit-scrollbar { width: 6px; }
.posts-area::-webkit-scrollbar-thumb { background: rgba(148,159,212,0.16); border-radius: 999px; }

.pinned-section { display: flex; flex-direction: column; gap: 8px; margin-bottom: 4px; }
.section-label {
  display: flex; align-items: center; gap: 5px;
  color: #7d87ab; font-size: 10px; font-weight: 700; letter-spacing: 0.06em; text-transform: uppercase;
}

.feed-state {
  display: flex; align-items: center; justify-content: center;
  flex-direction: column; gap: 10px; padding: 40px 0;
  color: #7d87ab; font-size: 13px;
}
.spinner {
  width: 22px; height: 22px; border-radius: 50%;
  border: 2px solid rgba(110,121,255,0.2); border-top-color: #6e79ff;
  animation: spin 0.7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* Post styles (used by inline ChannelPost) */
:deep(.channel-post) {
  padding: 14px 16px; border-radius: 14px;
  background: rgba(20,26,50,0.7);
  border: 1px solid rgba(255,255,255,0.06);
  position: relative; transition: border-color 0.15s;
}
:deep(.channel-post:hover) { border-color: rgba(110,121,255,0.2); }
:deep(.channel-post.pinned) { border-color: rgba(110,121,255,0.25); background: rgba(110,121,255,0.06); }
:deep(.channel-post.theme-light) { background: #fff; border-color: #e4e6f0; }
:deep(.channel-post.theme-light.pinned) { background: rgba(91,106,255,0.04); border-color: rgba(91,106,255,0.2); }

:deep(.pin-badge) {
  display: flex; align-items: center; gap: 4px;
  font-size: 10px; font-weight: 700; color: #6e79ff;
  margin-bottom: 8px; text-transform: uppercase; letter-spacing: 0.05em;
}

:deep(.post-body) { margin-bottom: 8px; }
:deep(.post-text) {
  color: #dde2f8; font-size: 14px; line-height: 1.6; font-weight: 500;
  white-space: pre-wrap; word-break: break-word;
}
:deep(.theme-light .post-text) { color: #1a1d2e; }

:deep(.post-media-img) {
  max-width: 100%; border-radius: 10px; margin-top: 8px; display: block;
}
:deep(.post-media-video) {
  max-width: 100%; border-radius: 10px; margin-top: 8px; display: block;
}
:deep(.post-file) {
  display: inline-flex; align-items: center; gap: 6px; margin-top: 8px;
  color: #6e79ff; font-size: 13px; text-decoration: none;
  padding: 6px 12px; border-radius: 8px; background: rgba(110,121,255,0.1);
  border: 1px solid rgba(110,121,255,0.2);
}

:deep(.post-footer) {
  display: flex; align-items: center; gap: 10px;
}
:deep(.post-time) { color: #5d6888; font-size: 11px; font-weight: 500; flex: 1; }
:deep(.post-views) {
  display: flex; align-items: center; gap: 3px;
  color: #5d6888; font-size: 11px;
}
:deep(.post-actions) { display: flex; gap: 4px; }
:deep(.post-action-btn) {
  width: 26px; height: 26px; border-radius: 8px;
  display: grid; place-items: center;
  color: #a6afd4; background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.06); cursor: pointer; transition: all 0.15s;
}
:deep(.post-action-btn:hover) { background: rgba(110,121,255,0.15); color: #6e79ff; border-color: rgba(110,121,255,0.3); }
:deep(.post-action-btn.danger:hover) { background: rgba(255,77,109,0.12); color: #ff4d6d; border-color: rgba(255,77,109,0.25); }

.composer-wrap {
  flex-shrink: 0; padding: 12px 24px 16px;
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
  display: flex; align-items: flex-end; gap: 10px;
  padding: 10px 14px; border-radius: 18px;
  border: 1px solid rgba(110,123,255,0.18);
  background: linear-gradient(180deg, rgba(25,30,58,0.68), rgba(18,23,46,0.78));
}
.theme-light .composer { background: #fff; border-color: rgba(91,106,255,0.2); }

.post-input {
  flex: 1; background: transparent; border: none; outline: none;
  color: #eef2ff; font-size: 14px; font-weight: 500; line-height: 1.5;
  resize: none; min-height: 22px; max-height: 120px; overflow-y: auto;
  font-family: inherit;
}
.theme-light .post-input { color: #1a1d2e; }
.post-input::placeholder { color: #747ea2; }

.send-btn {
  width: 34px; height: 34px; border-radius: 11px; flex-shrink: 0;
  display: grid; place-items: center; cursor: pointer; border: none;
  color: #fff; background: linear-gradient(135deg, #6e79ff, #8669ff);
  box-shadow: 0 8px 18px rgba(94,102,255,0.28); transition: opacity 0.15s;
}
.send-btn:disabled { opacity: 0.4; cursor: not-allowed; }

.modal-overlay {
  position: fixed; inset: 0;
  background: rgba(5,8,18,0.6); backdrop-filter: blur(4px);
  display: grid; place-items: center; z-index: 100;
}
.settings-modal {
  background: linear-gradient(180deg, rgba(22,28,52,0.97), rgba(16,20,38,0.99));
  border: 1px solid rgba(132,144,224,0.15);
  border-radius: 18px; padding: 28px; width: 360px;
  box-shadow: 0 24px 48px rgba(0,0,0,0.4);
}
.settings-modal h3 { color: #eef2ff; font-size: 16px; font-weight: 700; margin-bottom: 20px; }
.settings-field { margin-bottom: 14px; }
.settings-field label { display: block; color: #7d87ab; font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 6px; }
.settings-field input, .settings-field textarea, .settings-field select {
  width: 100%; background: rgba(255,255,255,0.04); border: 1px solid rgba(255,255,255,0.08);
  border-radius: 10px; padding: 9px 12px; color: #eef2ff; font-size: 13px; outline: none;
  font-family: inherit; box-sizing: border-box; resize: vertical;
}
.settings-field textarea { min-height: 70px; }
.modal-actions { display: flex; gap: 10px; justify-content: flex-end; margin-top: 20px; }
.btn-cancel {
  padding: 9px 18px; border-radius: 10px;
  background: rgba(255,255,255,0.05); border: 1px solid rgba(255,255,255,0.06);
  color: #a6afd4; font-size: 13px; cursor: pointer;
}
.btn-save {
  padding: 9px 18px; border-radius: 10px;
  background: linear-gradient(135deg, #6e79ff, #8669ff); border: none;
  color: #fff; font-size: 13px; font-weight: 600; cursor: pointer;
}

@media (max-width: 760px) {
  .channel-header { height: 64px; min-height: 64px; padding: 10px 14px; }
  .posts-area { padding: 12px 14px; }
  .composer-wrap { padding: 8px 14px calc(8px + env(safe-area-inset-bottom)); }
  .settings-modal { width: calc(100vw - 32px); }
}
</style>