<template>
  <aside class="chat-sidebar" :class="{ 'theme-light': isLight }">
    <div class="sidebar-shell">

      <!-- Header -->
      <header class="sidebar-header">
        <div class="brand">
          <div class="brand-mark">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.9">
              <path d="M4 6.5A2.5 2.5 0 0 1 6.5 4h11A2.5 2.5 0 0 1 20 6.5v7A2.5 2.5 0 0 1 17.5 16H11l-4.5 4V16A2.5 2.5 0 0 1 4 13.5v-7z" />
            </svg>
          </div>
          <div class="brand-text">
            <span class="brand-main">SV</span><span class="brand-accent">Lynx</span>
          </div>
        </div>

        <div class="header-actions">
          <div class="compose-wrap">
            <button class="header-btn" title="New" type="button" @click="showCompose = !showCompose">
              <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M12 5v14M5 12h14"/>
              </svg>
            </button>
            <transition name="dropdown">
              <div v-if="showCompose" class="compose-dropdown" v-click-outside="() => showCompose = false">
                <button class="compose-item" @click="showCompose = false; showCreateChannel = true">
                  <div class="compose-icon channel-icon">
                    <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                      <path d="M4 10h3l6-4v12l-6-4H4z" />
                      <path d="M7 14.5v3a2 2 0 0 0 2 2h1" />
                      <path d="M18 9a3 3 0 0 1 0 6" />
                      <path d="M20.5 7.5a5 5 0 0 1 0 9" />
                    </svg>
                  </div>
                  <div>
                    <div class="compose-label">New Channel</div>
                    <div class="compose-hint">Broadcast to subscribers</div>
                  </div>
                </button>
                <button class="compose-item" @click="showCompose = false">
                  <div class="compose-icon group-icon">
                    <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.8">
                      <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
                      <circle cx="9" cy="7" r="4"/>
                      <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
                      <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
                    </svg>
                  </div>
                  <div>
                    <div class="compose-label">New Group</div>
                    <div class="compose-hint">Coming soon</div>
                  </div>
                </button>
              </div>
            </transition>
          </div>

          <button
            class="header-btn"
            :class="{ 'delete-mode-active': deleteMode }"
            title="Manage chats"
            type="button"
            @click="deleteMode = !deleteMode"
          >
            <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="1.8">
              <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
              <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
            </svg>
          </button>
        </div>
      </header>

      <!-- Search -->
      <div class="search-wrap">
        <div class="search-box">
          <svg class="search-icon" viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="1.8">
            <circle cx="11" cy="11" r="7"/><path d="M20 20l-3.5-3.5"/>
          </svg>
          <input v-model="search" type="text" class="search-input" placeholder="Search..." />
          <button v-if="search" class="search-clear" @click="search = ''">
            <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 6L6 18M6 6l12 12"/>
            </svg>
          </button>
        </div>
      </div>

      <!-- Tabs -->
      <div class="sidebar-tabs">
        <button class="tab-btn" :class="{ active: activeTab === 'all' }"      type="button" @click="activeTab = 'all'">All</button>
        <button class="tab-btn" :class="{ active: activeTab === 'chats' }"    type="button" @click="activeTab = 'chats'">Chats</button>
        <button class="tab-btn" :class="{ active: activeTab === 'groups' }"   type="button" @click="activeTab = 'groups'">Groups</button>
        <button class="tab-btn" :class="{ active: activeTab === 'channels' }" type="button" @click="activeTab = 'channels'">Channels</button>
      </div>

      <!-- List -->
      <div class="sidebar-list">

        <!-- SEARCH MODE -->
        <template v-if="search.trim().length > 0">
          <div v-if="isSearching" class="list-state">Searching...</div>
          <template v-else>

            <!-- Search: users -->
            <template v-if="activeTab !== 'channels' && searchUsers.length">
              <div v-if="activeTab === 'all' && searchChannels.length" class="section-label">Chats</div>
              <div
                v-for="user in searchUsers"
                :key="'u-' + user.id"
                class="chat-item"
                @click="handleStartChat(user.id, user.nickname)"
              >
                <div class="chat-avatar" :style="avatarBg(null, user.avatar_color)">
                  <img v-if="user.photo_url" :src="user.photo_url" alt="" class="avatar-img" />
                  <span v-else>{{ (user.name || user.first_name || user.nickname)?.[0]?.toUpperCase() || '?' }}</span>
                </div>
                <div class="chat-body">
                  <div class="chat-topline">
                    <span class="chat-name">{{ user.name || (user.first_name ? user.first_name + (user.last_name ? ' ' + user.last_name : '') : null) || user.nickname || user.username || 'Unknown' }}</span>
                  </div>
                  <div class="chat-bottomline">
                    <span class="chat-preview">@{{ user.nickname || user.username || '' }}</span>
                  </div>
                </div>
              </div>
            </template>

            <!-- Search: channels -->
            <template v-if="activeTab !== 'chats' && activeTab !== 'groups' && searchChannels.length">
              <div v-if="activeTab === 'all' && searchUsers.length" class="section-label">Channels</div>
              <button
                v-for="ch in searchChannels"
                :key="'sch-' + ch.id"
                class="chat-item"
                :class="{ active: String(activeChannelId) === String(ch.id) }"
                type="button"
                @click="selectOrJoinChannel(ch)"
              >
                <div class="chat-avatar channel-avatar" :style="avatarBg(ch.avatar_url, ch.avatar_color)">
                  <img v-if="ch.avatar_url" :src="ch.avatar_url" alt="" class="avatar-img" />
                  <span v-else>{{ ch.name?.[0]?.toUpperCase() || '#' }}</span>
                </div>
                <div class="chat-body">
                  <div class="chat-topline">
                    <span class="chat-name">{{ ch.name }}</span>
                    <span class="chat-time">{{ ch.member_count }} 👥</span>
                  </div>
                  <div class="chat-bottomline">
                    <span class="chat-preview">@{{ ch.handle }}</span>
                    <span v-if="!isMyChannel(ch)" class="join-badge">Subscribe</span>
                    <span v-else class="joined-badge">Joined</span>
                  </div>
                </div>
              </button>
            </template>

            <div v-if="!searchUsers.length && !searchChannels.length" class="list-state">Nothing found</div>
          </template>
        </template>

        <!-- BROWSE MODE -->
        <template v-else>

          <!-- Tab: All — Chats section -->
          <template v-if="activeTab === 'all' || activeTab === 'chats'">
            <div v-if="activeTab === 'all' && channels.length" class="section-label">Chats</div>
            <div v-for="direct in directs" :key="direct.id" class="chat-item-wrap">
              <button
                class="chat-item"
                :class="{ active: String(activeId) === String(direct.id) }"
                type="button"
                @click="!deleteMode && $emit('select', { chatId: direct.id, recipientId: getRecipientId(direct) })"
              >
                <div class="chat-avatar" :style="avatarBg(getAvatarUrl(direct), direct.companion_avatar_color)">
                  <img v-if="getAvatarUrl(direct)" :src="getAvatarUrl(direct)" alt="" class="avatar-img" />
                  <span v-else>{{ getAvatarLetter(direct) }}</span>
                </div>
                <div class="chat-body">
                  <div class="chat-topline">
                    <!-- В режиме All — добавляем значок чата чтобы отличать от каналов -->
                    <span class="chat-name">
                      {{ getRecipientName(direct) }}
                    </span>
                    <span class="chat-time">{{ getChatTime(direct) }}</span>
                  </div>
                  <div class="chat-bottomline">
                    <span class="chat-preview">{{ getChatPreview(direct) }}</span>
                    <span v-if="getUnreadCount(direct) > 0" class="unread-badge">{{ getUnreadCount(direct) }}</span>
                  </div>
                </div>
              </button>
              <button v-if="deleteMode" class="delete-chat-btn" type="button" @click="askDeleteChat(direct)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.8">
                  <path d="M3 6h18"/>
                  <path d="M8 6V4.8c0-.99.81-1.8 1.8-1.8h4.4c.99 0 1.8.81 1.8 1.8V6"/>
                  <path d="M18.2 6l-.72 11.02A2 2 0 0 1 15.48 19H8.52a2 2 0 0 1-1.99-1.98L5.8 6"/>
                </svg>
              </button>
            </div>
            <div v-if="directs.length === 0 && activeTab === 'chats'" class="list-state">No chats yet. Search to start one.</div>
          </template>

          <!-- Tab: All + Channels — channel section -->
          <template v-if="activeTab === 'all' || activeTab === 'channels'">
            <div v-if="activeTab === 'all' && directs.length && channels.length" class="section-label">Channels</div>

            <!-- MY CHANNELS subsection (only in Channels tab) -->
            <template v-if="activeTab === 'channels'">
              <div v-if="myOwnedChannels.length" class="section-label sub-label">
                <svg viewBox="0 0 24 24" width="10" height="10" fill="currentColor" style="opacity:0.7"><path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/></svg>
                My Channels
              </div>
              <div v-for="ch in myOwnedChannels" :key="'own-' + ch.id" class="chat-item-wrap">
                <button
                  class="chat-item channel-item"
                  :class="{ active: String(activeChannelId) === String(ch.id) }"
                  type="button"
                  @click="$emit('select-channel', ch)"
                >
                  <div class="chat-avatar channel-avatar" :style="avatarBg(ch.avatar_url, ch.avatar_color)">
                    <img v-if="ch.avatar_url" :src="ch.avatar_url" alt="" class="avatar-img" />
                    <span v-else>{{ ch.name?.[0]?.toUpperCase() || '#' }}</span>
                  </div>
                  <div class="chat-body">
                    <div class="chat-topline">
                      <span class="chat-name">{{ ch.name }}</span>
                      <span class="chat-time">{{ getChannelTime(ch) }}</span>
                    </div>
                    <div class="chat-bottomline">
                      <span class="chat-preview">{{ ch.last_post_content || '@' + ch.handle }}</span>
                      <span class="channel-badge owner-badge">owner</span>
                    </div>
                  </div>
                </button>
                <button class="leave-btn" type="button" title="Leave channel" @click="confirmLeave(ch)">
                  <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="1.8">
                    <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
                    <polyline points="16 17 21 12 16 7"/>
                    <line x1="21" y1="12" x2="9" y2="12"/>
                  </svg>
                </button>
              </div>

              <div v-if="myOwnedChannels.length && subscribedChannels.length" class="section-label sub-label" style="margin-top:8px">
                <svg viewBox="0 0 24 24" width="10" height="10" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/></svg>
                Subscribed
              </div>
            </template>

            <!-- All channels (in All tab) or subscribed (in Channels tab) -->
            <div v-for="ch in (activeTab === 'channels' ? subscribedChannels : channels)" :key="'ch-' + ch.id" class="chat-item-wrap">
              <button
                class="chat-item channel-item"
                :class="{ active: String(activeChannelId) === String(ch.id) }"
                type="button"
                @click="$emit('select-channel', ch)"
              >
                <div class="chat-avatar channel-avatar" :style="avatarBg(ch.avatar_url, ch.avatar_color)">
                  <img v-if="ch.avatar_url" :src="ch.avatar_url" alt="" class="avatar-img" />
                  <span v-else>{{ ch.name?.[0]?.toUpperCase() || '#' }}</span>
                </div>
                <div class="chat-body">
                  <div class="chat-topline">
                    <!-- В режиме All — значок канала для отличия от чатов -->
                    <span class="chat-name">
                      {{ ch.name }}
                    </span>
                    <span class="chat-time">{{ getChannelTime(ch) }}</span>
                  </div>
                  <div class="chat-bottomline">
                    <span class="chat-preview">{{ ch.last_post_content || '@' + ch.handle }}</span>
                    <span class="channel-badge">{{ ch.member_count }}</span>
                  </div>
                </div>
              </button>
              <button class="leave-btn" type="button" title="Leave channel" @click="confirmLeave(ch)">
                <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="1.8">
                  <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
                  <polyline points="16 17 21 12 16 7"/>
                  <line x1="21" y1="12" x2="9" y2="12"/>
                </svg>
              </button>
            </div>

            <div v-if="channels.length === 0 && activeTab === 'channels'" class="list-state">No channels yet.</div>
          </template>

          <div v-if="activeTab === 'groups'" class="list-state">Groups coming soon.</div>

        </template>
      </div>

      <!-- Footer -->
      <footer class="sidebar-footer">
        <div class="footer-actions">
          <button class="footer-btn" title="My profile" type="button" @click="$emit('open-profile')">
            <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="1.8">
              <circle cx="12" cy="8" r="4"/>
              <path d="M4 20c0-4 3.6-7 8-7s8 3 8 7"/>
            </svg>
          </button>
          <button class="footer-btn" title="Toggle theme" type="button" @click="toggleTheme">
            <svg v-if="!isLight" viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="1.8">
              <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
            </svg>
            <svg v-else viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="1.8">
              <circle cx="12" cy="12" r="5"/>
              <line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/>
              <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/>
              <line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/>
              <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/>
            </svg>
          </button>
        </div>
      </footer>
    </div>

    <!-- Delete chat confirm -->
    <div v-if="chatToDelete" class="modal-overlay" @click.self="chatToDelete = null">
      <div class="confirm-modal">
        <h3>Delete chat?</h3>
        <p>This will permanently delete the chat and all messages for both users.</p>
        <div class="modal-actions">
          <button class="btn-cancel" type="button" @click="chatToDelete = null">Cancel</button>
          <button class="btn-delete" type="button" @click="confirmDeleteChat">Delete</button>
        </div>
      </div>
    </div>

    <!-- Leave channel confirm -->
    <div v-if="channelToLeave" class="modal-overlay" @click.self="channelToLeave = null">
      <div class="confirm-modal">
        <div class="leave-icon">
          <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="1.8">
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
            <polyline points="16 17 21 12 16 7"/>
            <line x1="21" y1="12" x2="9" y2="12"/>
          </svg>
        </div>
        <h3>Leave "{{ channelToLeave.name }}"?</h3>
        <p>You will stop receiving posts from this channel. You can re-subscribe at any time.</p>
        <div class="modal-actions">
          <button class="btn-cancel" type="button" @click="channelToLeave = null">Cancel</button>
          <button class="btn-leave" type="button" @click="doLeaveChannel">Leave</button>
        </div>
      </div>
    </div>

    <!-- Create channel modal -->
    <CreateChannelModal
      v-if="showCreateChannel"
      :currentUserId="currentUserId"
      :isLight="isLight"
      @close="showCreateChannel = false"
      @created="onChannelCreated"
    />
  </aside>
</template>

<script>
import { apiFetch } from '../api.js'
import CreateChannelModal from './CreateChannelModal.vue'

const BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

export default {
  name: 'ChatSidebar',
  components: { CreateChannelModal },

  directives: {
    'click-outside': {
      mounted(el, binding) {
        el._co = e => { if (!el.contains(e.target)) binding.value(e) }
        document.addEventListener('mousedown', el._co)
      },
      unmounted(el) { document.removeEventListener('mousedown', el._co) }
    }
  },

  props: {
    directs:         { type: Array,            default: () => [] },
    channels:        { type: Array,            default: () => [] },
    activeId:        { type: [String, Number], default: null },
    activeChannelId: { type: [String, Number], default: null },
    currentUserId:   { type: String,           default: null },
    isLight:         { type: Boolean,          default: false }
  },

  emits: ['select', 'select-channel', 'channel-created', 'start-chat', 'toggle-theme', 'chat-deleted', 'open-profile', 'leave-channel'],

  data() {
    return {
      search: '',
      searchUsers: [],
      searchChannels: [],
      isSearching: false,
      searchTimeout: null,
      activeTab: 'all',
      deleteMode: false,
      chatToDelete: null,
      channelToLeave: null,
      showCreateChannel: false,
      showCompose: false,
    }
  },

  computed: {
    // Каналы где юзер — owner/admin (Мои каналы)
    myOwnedChannels() {
      return this.channels.filter(c => ['owner', 'admin'].includes(c.user_role))
    },
    // Каналы где юзер подписан но не owner/admin
    subscribedChannels() {
      return this.channels.filter(c => !['owner', 'admin'].includes(c.user_role))
    }
  },

  watch: {
    search(value) {
      clearTimeout(this.searchTimeout)
      if (!value.trim()) {
        this.searchUsers = []
        this.searchChannels = []
        this.isSearching = false
        return
      }
      const inviteMatch = value.match(/\/invite\/([a-f0-9]{32})/)
      if (inviteMatch) { this.joinByInvite(inviteMatch[1]); return }
      this.isSearching = true
      this.searchTimeout = setTimeout(() => this.doSearch(value.trim()), 350)
    },
  },

  methods: {
    avatarBg(url, color) {
      if (url) return {}
      return { background: color || 'linear-gradient(135deg, #6572ff, #8a67ff)' }
    },

    async doSearch(query) {
      this.searchUsers = []
      this.searchChannels = []
      this.isSearching = true
      try {
        const tasks = []
        if (this.activeTab !== 'channels') {
          tasks.push(
            apiFetch(`${BASE}/users/search?q=${encodeURIComponent(query)}&user_id=${this.currentUserId || ''}`)
              .then(r => r.json()).then(d => { this.searchUsers = d.users || [] }).catch(() => {})
          )
        }
        if (this.activeTab !== 'chats' && this.activeTab !== 'groups') {
          tasks.push(
            apiFetch(`${BASE}/channels/search?q=${encodeURIComponent(query)}`)
              .then(r => r.json()).then(d => { this.searchChannels = d.channels || [] }).catch(() => {})
          )
        }
        await Promise.all(tasks)
      } finally {
        this.isSearching = false
      }
    },

    async joinByInvite(token) {
      this.isSearching = true
      try {
        const res = await apiFetch(`${BASE}/invites/${token}/join`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ user_id: this.currentUserId })
        })
        const data = await res.json()
        if (!res.ok) { this.searchUsers = []; this.searchChannels = []; return }
        const channel = data.channel
        this.$emit('channel-created', channel)
        this.$emit('select-channel', channel)
        this.search = ''
      } catch (e) {
        console.error('joinByInvite error', e)
      } finally {
        this.isSearching = false
      }
    },

    toggleTheme() { this.$emit('toggle-theme') },

    handleStartChat(userId, nickname) {
      this.search = ''
      this.$emit('start-chat', userId, nickname)
    },

    getRecipientId(direct) {
      const first  = direct.first_user_id  ?? direct.firstuserid
      const second = direct.second_user_id ?? direct.seconduserid
      return String(first) === String(this.currentUserId) ? second : first
    },

    getRecipientName(direct) {
      return direct.companion_name || direct.companion_nickname || direct.nickname || 'Unknown'
    },

    getAvatarLetter(direct) { return this.getRecipientName(direct)?.[0]?.toUpperCase() || '?' },

    getAvatarUrl(direct) { return direct.companion_photo_url || direct.photo_url || null },

    getChatPreview(direct) {
      const c = direct.last_message_content || ''
      if (c.startsWith('http') && (c.includes('/voice/') || c.includes('.webm') || c.includes('.ogg') || c.includes('.mp3'))) return '🎤 Голосовое сообщение'
      if (c.startsWith('http') && c.includes('/media/images/')) return '📷 Фото'
      if (c.startsWith('http') && c.includes('/media/videos/')) return '🎥 Видео'
      if (c.startsWith('http') && c.includes('/media/audio/'))  return '🎵 Аудио'
      if (c.startsWith('http') && c.includes('/media/files/'))  return '📎 Файл'
      return c
    },

    getUnreadCount(direct) { return Number(direct.unread_count || direct.unreadcount || 0) },

    getChatTime(direct) {
      const raw = direct.last_message_at || direct.updated_at || direct.created_at || direct.creation_time
      if (!raw) return ''
      const d = new Date(raw)
      if (isNaN(d.getTime()) || d.getFullYear() < 2000) return ''
      return d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
    },

    getChannelTime(ch) {
      const raw = ch.last_post_at || ch.created_at
      if (!raw) return ''
      const d = new Date(raw)
      if (isNaN(d.getTime())) return ''
      return d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
    },

    askDeleteChat(direct) { this.chatToDelete = direct },

    isMyChannel(ch) {
      return this.channels.some(c => String(c.id) === String(ch.id))
    },

    async selectOrJoinChannel(ch) {
      if (this.isMyChannel(ch)) {
        this.$emit('select-channel', ch)
        return
      }
      try {
        const res = await apiFetch(`${BASE}/channels/${ch.id}/subscribe`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ user_id: this.currentUserId })
        })
        if (!res.ok) return
        // Запросить разрешение на уведомления при подписке
        if ('Notification' in window && Notification.permission === 'default') {
          await Notification.requestPermission()
        }
        this.$emit('channel-created', { ...ch, user_role: 'member' })
        this.$emit('select-channel', { ...ch, user_role: 'member' })
        this.search = ''
      } catch (e) {
        console.error('subscribe channel error', e)
      }
    },

    onChannelCreated(ch) {
      this.$emit('channel-created', ch)
      this.showCreateChannel = false
    },

    confirmLeave(channel) {
      this.channelToLeave = channel
    },

    async doLeaveChannel() {
      if (!this.channelToLeave) return
      const ch = this.channelToLeave
      this.channelToLeave = null
      try {
        const res = await apiFetch(`${BASE}/channels/${ch.id}/leave`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ user_id: this.currentUserId })
        })
        if (!res.ok) return
        this.$emit('leave-channel', ch.id)
      } catch (e) {
        console.error('leave channel error', e)
      }
    },

    async confirmDeleteChat() {
      if (!this.chatToDelete) return
      const chatId = this.chatToDelete.id
      const recipientId = this.getRecipientId(this.chatToDelete)
      this.chatToDelete = null
      try {
        const url = new URL(`${BASE}/chat/direct/${chatId}`)
        url.searchParams.set('recipient_id', recipientId)
        const res = await apiFetch(url.toString(), { method: 'DELETE' })
        if (!res.ok) return
        this.$emit('chat-deleted', chatId)
        this.deleteMode = false
      } catch (e) {
        console.error('Delete chat error', e)
      }
    }
  }
}
</script>

<style scoped>
.chat-sidebar {
  width: 100%; height: 100%; overflow: hidden;
  background: transparent; font-family: 'Satoshi', sans-serif; position: relative;
}
.sidebar-shell {
  height: 100%; display: flex; flex-direction: column; overflow: hidden;
  background: linear-gradient(180deg, rgba(8,12,26,0.98), rgba(7,10,22,0.98));
}
.theme-light .sidebar-shell { background: #ffffff; }

/* Header */
.sidebar-header {
  height: 78px; padding: 18px 16px 14px;
  display: flex; align-items: center; justify-content: space-between;
}
.brand { display: flex; align-items: center; gap: 10px; }
.brand-mark {
  width: 31px; height: 31px; border-radius: 10px;
  display: grid; place-items: center; color: #fff;
  background: linear-gradient(135deg, #6675ff, #7b68ff);
  box-shadow: 0 8px 18px rgba(90,98,255,0.22);
}
.theme-light .brand-mark { background: linear-gradient(135deg, #5b6aff, #7b68ff); }
.brand-text { font-size: 18px; font-weight: 800; letter-spacing: -0.02em; }
.brand-main  { color: #f3f5ff; } .theme-light .brand-main  { color: #1a1d2e; }
.brand-accent{ color: #92a0ff; } .theme-light .brand-accent{ color: #5b6aff; }

.header-actions { display: flex; align-items: center; gap: 6px; }
.header-btn {
  width: 32px; height: 32px; border-radius: 10px;
  display: grid; place-items: center;
  color: #98a2ca; background: rgba(255,255,255,0.02); border: 1px solid rgba(255,255,255,0.04);
  cursor: pointer; transition: all 0.2s;
}
.header-btn:hover { background: rgba(255,255,255,0.06); }
.theme-light .header-btn { color: #7880a0; background: #f3f4f8; border-color: #e4e6f0; }
.header-btn.delete-mode-active { color: #ff4d6d; background: rgba(255,77,109,0.12); border-color: rgba(255,77,109,0.25); }

/* Compose dropdown */
.compose-wrap { position: relative; }
.compose-dropdown {
  position: absolute; top: calc(100% + 8px); right: 0; width: 220px; z-index: 50;
  background: linear-gradient(180deg, rgba(22,28,52,0.98), rgba(14,18,36,0.99));
  border: 1px solid rgba(132,144,224,0.15); border-radius: 14px; overflow: hidden;
  box-shadow: 0 16px 40px rgba(0,0,0,0.4);
}
.theme-light .compose-dropdown { background: #fff; border-color: #e4e6f0; box-shadow: 0 8px 24px rgba(91,106,200,0.15); }
.compose-item {
  width: 100%; display: flex; align-items: center; gap: 12px;
  padding: 12px 14px; background: transparent; border: none; cursor: pointer; text-align: left; transition: background 0.15s;
}
.compose-item:hover { background: rgba(255,255,255,0.05); }
.theme-light .compose-item:hover { background: #f5f6fc; }
.compose-icon {
  width: 32px; height: 32px; border-radius: 10px; flex-shrink: 0; display: grid; place-items: center;
}
.channel-icon { background: rgba(110,121,255,0.15); color: #6e79ff; border: 1px solid rgba(110,121,255,0.2); }
.group-icon   { background: rgba(34,197,94,0.12);  color: #22c55e; border: 1px solid rgba(34,197,94,0.2); }
.compose-label { color: #eef2ff; font-size: 13px; font-weight: 600; }
.theme-light .compose-label { color: #1a1d2e; }
.compose-hint  { color: #5d6888; font-size: 11px; margin-top: 1px; }

/* Search */
.search-wrap { padding: 0 16px 12px; }
.search-box {
  height: 40px; display: flex; align-items: center; gap: 9px; padding: 0 12px;
  border-radius: 13px; background: rgba(255,255,255,0.028); border: 1px solid rgba(255,255,255,0.035);
}
.theme-light .search-box { background: #f3f4f8; border-color: #e2e4ee; }
.search-icon { color: #7480a8; flex-shrink: 0; }
.theme-light .search-icon { color: #9098b8; }
.search-input {
  flex: 1; min-width: 0; background: transparent; border: none; outline: none;
  color: #eef1ff; font-size: 16px; font-weight: 500;
}
.theme-light .search-input { color: #1a1d2e; }
.search-input::placeholder { color: #6d7798; }
.theme-light .search-input::placeholder { color: #aab0cc; }
.search-clear {
  width: 18px; height: 18px; border-radius: 50%; flex-shrink: 0;
  display: grid; place-items: center; background: rgba(255,255,255,0.08);
  border: none; color: #7480a8; cursor: pointer;
}
.theme-light .search-clear { background: #e4e6f0; color: #9098b8; }

/* Tabs */
.sidebar-tabs { display: flex; gap: 6px; padding: 0 16px 14px; overflow-x: auto; }
.sidebar-tabs::-webkit-scrollbar { display: none; }
.tab-btn {
  height: 28px; padding: 0 11px; border-radius: 10px; white-space: nowrap;
  color: #7d87ab; font-size: 11px; font-weight: 700;
  background: transparent; border: 1px solid transparent; cursor: pointer; transition: all 0.2s;
}
.tab-btn.active { color: #f1f4ff; background: rgba(96,108,255,0.14); border-color: rgba(114,126,255,0.16); }
.theme-light .tab-btn { color: #8890b4; }
.theme-light .tab-btn.active { color: #1a1d2e; background: rgba(91,106,255,0.1); border-color: rgba(91,106,255,0.18); }

/* List */
.sidebar-list { flex: 1; overflow-y: auto; min-height: 0; padding: 2px 10px 12px; }
.sidebar-list::-webkit-scrollbar { width: 6px; }
.sidebar-list::-webkit-scrollbar-thumb { background: rgba(147,158,211,0.16); border-radius: 999px; }

.section-label {
  padding: 10px 4px 4px; color: #5d6888;
  font-size: 10px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.06em;
  display: flex; align-items: center; gap: 5px;
}
.sub-label { color: #7d87ab; }

/* Item wrap with leave button */
.chat-item-wrap { display: flex; align-items: center; gap: 6px; margin-bottom: 2px; }

.chat-item {
  flex: 1; min-width: 0; overflow: hidden;
  display: flex; align-items: center; gap: 11px;
  padding: 9px 10px; border-radius: 14px; text-align: left;
  background: transparent; border: 1px solid transparent;
  transition: all 0.15s; cursor: pointer;
}
.chat-item:hover { background: rgba(255,255,255,0.03); }
.theme-light .chat-item:hover { background: #f3f4f8; }

/* P2P chats — active: синяя оболочка */
.chat-item.active {
  background: linear-gradient(180deg, rgba(75,88,228,0.17), rgba(64,78,210,0.12));
  border-color: rgba(110,122,255,0.26);
  box-shadow: 0 0 0 1px rgba(98,112,255,0.08) inset, 0 8px 20px rgba(49,61,180,0.14);
}
.theme-light .chat-item.active {
  background: linear-gradient(180deg, rgba(91,106,255,0.1), rgba(91,106,255,0.07));
  border-color: rgba(91,106,255,0.22);
}

/* Channels — active: тоже синяя оболочка, идентично чатам */
.channel-item.active {
  background: linear-gradient(180deg, rgba(75,88,228,0.17), rgba(64,78,210,0.12));
  border-color: rgba(110,122,255,0.26);
  box-shadow: 0 0 0 1px rgba(98,112,255,0.08) inset, 0 8px 20px rgba(49,61,180,0.14);
}
.theme-light .channel-item.active {
  background: linear-gradient(180deg, rgba(91,106,255,0.1), rgba(91,106,255,0.07));
  border-color: rgba(91,106,255,0.22);
}

/* Delete chat button */
.delete-chat-btn {
  flex-shrink: 0; width: 30px; height: 30px; border-radius: 10px;
  display: grid; place-items: center;
  color: #ff4d6d; background: rgba(255,77,109,0.1); border: 1px solid rgba(255,77,109,0.2);
  cursor: pointer; transition: all 0.15s;
}
.delete-chat-btn:hover { background: rgba(255,77,109,0.2); }
.theme-light .delete-chat-btn { background: rgba(255,60,80,0.06); border-color: rgba(255,60,80,0.15); color: #ff3c50; }

/* Leave channel button */
.leave-btn {
  flex-shrink: 0; width: 28px; height: 28px; border-radius: 9px;
  display: grid; place-items: center;
  color: #7d87ab; background: rgba(255,255,255,0.03); border: 1px solid rgba(255,255,255,0.05);
  cursor: pointer; transition: all 0.15s; opacity: 0;
}
.chat-item-wrap:hover .leave-btn { opacity: 1; }
.leave-btn:hover { color: #ff4d6d; background: rgba(255,77,109,0.1); border-color: rgba(255,77,109,0.2); }
.theme-light .leave-btn { color: #9098b8; background: #f3f4f8; border-color: #e4e6f0; }
.theme-light .leave-btn:hover { color: #ff3c50; background: rgba(255,60,80,0.06); }

/* Avatar */
.chat-avatar {
  width: 42px; height: 42px; border-radius: 14px; flex-shrink: 0;
  display: grid; place-items: center; overflow: hidden;
  color: #fff; font-size: 13px; font-weight: 700;
  background: linear-gradient(135deg, #6572ff, #8a67ff);
}
.channel-avatar { border-radius: 12px; }
.avatar-img { width: 100%; height: 100%; object-fit: cover; }

.chat-body { flex: 1; min-width: 0; }
.chat-topline, .chat-bottomline { display: flex; align-items: center; justify-content: space-between; gap: 8px; overflow: hidden; }
.chat-topline { margin-bottom: 3px; }
.chat-name { color: #eef2ff; font-size: 13px; font-weight: 700; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; display: flex; align-items: center; gap: 3px; }
.theme-light .chat-name { color: #1a1d2e; }
.chat-time { flex-shrink: 0; color: #7580a6; font-size: 10.5px; font-weight: 600; }
.theme-light .chat-time { color: #9098b8; }
.chat-preview { color: #8590b4; font-size: 11.5px; font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; flex: 1; min-width: 0; }
.theme-light .chat-preview { color: #7880a0; }
.unread-badge {
  min-width: 18px; height: 18px; padding: 0 5px; border-radius: 999px;
  display: inline-grid; place-items: center; flex-shrink: 0;
  color: #fff; font-size: 10px; font-weight: 700;
  background: linear-gradient(135deg, #6a76ff, #8866ff);
}
.channel-badge {
  font-size: 10px; font-weight: 600; color: #7d87ab;
  background: rgba(255,255,255,0.04); padding: 2px 6px; border-radius: 999px; flex-shrink: 0;
}
.owner-badge {
  color: #f0b429; background: rgba(240,180,41,0.12); border: 1px solid rgba(240,180,41,0.2);
}
.theme-light .channel-badge { background: #f0f1f8; }

.join-badge {
  font-size: 10px; font-weight: 700;
  color: #6e79ff; background: rgba(110,121,255,0.12);
  border: 1px solid rgba(110,121,255,0.25);
  padding: 2px 8px; border-radius: 999px; flex-shrink: 0; white-space: nowrap;
}
.joined-badge {
  font-size: 10px; font-weight: 700;
  color: #22c55e; background: rgba(34,197,94,0.1);
  border: 1px solid rgba(34,197,94,0.2);
  padding: 2px 8px; border-radius: 999px; flex-shrink: 0; white-space: nowrap;
}

.list-state { padding: 26px 12px; text-align: center; color: #7c86ad; font-size: 12px; }
.theme-light .list-state { color: #9098b8; }

/* Footer */
.sidebar-footer {
  padding: 12px 16px 14px; border-top: 1px solid rgba(255,255,255,0.03);
  position: sticky; bottom: 0; background: rgba(7,10,22,0.98);
}
.theme-light .sidebar-footer { border-top-color: #e8eaf0; background: #ffffff; }
.footer-actions { display: flex; gap: 10px; }
.footer-btn {
  width: 34px; height: 34px; border-radius: 11px; display: grid; place-items: center;
  color: #97a2c8; background: rgba(255,255,255,0.02); border: 1px solid rgba(255,255,255,0.04);
  cursor: pointer; transition: all 0.2s;
}
.footer-btn:hover { background: rgba(255,255,255,0.05); }
.theme-light .footer-btn { color: #7880a0; background: #f3f4f8; border-color: #e2e4ee; }

/* Modals */
.modal-overlay {
  position: absolute; inset: 0; background: rgba(5,8,20,0.55);
  backdrop-filter: blur(4px); display: grid; place-items: center;
  z-index: 20; border-radius: inherit;
}
.theme-light .modal-overlay { background: rgba(200,205,230,0.5); }
.confirm-modal {
  background: linear-gradient(180deg, rgba(22,28,52,0.97), rgba(16,20,38,0.99));
  border: 1px solid rgba(132,144,224,0.15);
  border-radius: 16px; padding: 24px; width: 260px;
  text-align: center; box-shadow: 0 20px 40px rgba(0,0,0,0.4);
}
.theme-light .confirm-modal { background: #ffffff; border-color: #dde1f0; box-shadow: 0 12px 40px rgba(90,106,200,0.15); }
.leave-icon {
  width: 44px; height: 44px; border-radius: 13px;
  display: grid; place-items: center; margin: 0 auto 12px;
  color: #ff4d6d; background: rgba(255,77,109,0.1); border: 1px solid rgba(255,77,109,0.2);
}
.confirm-modal h3 { color: #eef2ff; font-size: 15px; margin-bottom: 8px; }
.confirm-modal p  { color: #8d96ba; font-size: 12px; line-height: 1.6; margin-bottom: 20px; }
.theme-light .confirm-modal h3 { color: #1a1d2e; }
.theme-light .confirm-modal p  { color: #7880a0; }
.modal-actions { display: flex; gap: 10px; justify-content: center; }
.btn-cancel { padding: 8px 14px; border-radius: 10px; background: rgba(255,255,255,0.05); border: 1px solid rgba(255,255,255,0.06); color: #a6afd4; font-size: 13px; cursor: pointer; }
.theme-light .btn-cancel { background: #f3f4f8; border-color: #e2e4ee; color: #7880a0; }
.btn-delete { padding: 8px 14px; border-radius: 10px; background: linear-gradient(135deg, #ff4d6d, #d93856); border: none; color: #fff; font-size: 13px; cursor: pointer; }
.btn-leave  { padding: 8px 14px; border-radius: 10px; background: linear-gradient(135deg, #ff4d6d, #d93856); border: none; color: #fff; font-size: 13px; cursor: pointer; }

/* Animations */
.dropdown-enter-active, .dropdown-leave-active { transition: opacity 0.15s, transform 0.15s; }
.dropdown-enter-from, .dropdown-leave-to { opacity: 0; transform: translateY(-6px) scale(0.97); }

@media (max-width: 760px) {
  .chat-sidebar { border-radius: 0; }
  .sidebar-shell { padding-bottom: env(safe-area-inset-bottom); }
  .theme-light .sidebar-shell { background: #ffffff; }
  .sidebar-header { height: 64px; padding: 12px 14px 10px; }
  .search-wrap { padding: 0 12px 10px; }
  .sidebar-tabs { padding: 0 12px 12px; }
  .sidebar-list { padding: 2px 8px 12px; }
  .confirm-modal { width: calc(100vw - 48px); max-width: 280px; }
  .sidebar-footer { padding-bottom: env(safe-area-inset-bottom, 14px); }
  .compose-dropdown { right: -8px; }
  .leave-btn { opacity: 1; }
}
</style>