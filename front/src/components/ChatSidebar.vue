<template>
  <aside class="chat-sidebar" :class="{ 'theme-light': isLight }">
    <div class="sidebar-shell">
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

        <!-- Карандаш = режим удаления чатов -->
        <button
          class="header-btn"
          :class="{ 'delete-mode-active': deleteMode }"
          title="Manage chats"
          type="button"
          @click="deleteMode = !deleteMode"
        >
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.8">
            <path d="M5 20h14"></path>
            <path d="M15.5 4.5l4 4L10 18l-4 1 1-4 9.5-10.5z"></path>
          </svg>
        </button>
      </header>

      <div class="search-wrap">
        <div class="search-box">
          <svg class="search-icon" viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="1.8">
            <circle cx="11" cy="11" r="7"></circle>
            <path d="M20 20l-3.5-3.5"></path>
          </svg>
          <input
            v-model="search"
            type="text"
            class="search-input"
            placeholder="Search..."
          />
        </div>
      </div>

      <div class="sidebar-tabs">
        <button class="tab-btn" :class="{ active: activeTab === 'all' }" type="button" @click="activeTab = 'all'">All</button>
        <button class="tab-btn" :class="{ active: activeTab === 'chats' }" type="button" @click="activeTab = 'chats'">Chats</button>
        <button class="tab-btn" :class="{ active: activeTab === 'groups' }" type="button" @click="activeTab = 'groups'">Groups</button>
        <button class="tab-btn" :class="{ active: activeTab === 'channels' }" type="button" @click="activeTab = 'channels'">Channels</button>
      </div>

      <div class="sidebar-list">
        <!-- Поиск пользователей -->
        <template v-if="search.trim().length > 0">
          <div v-if="isSearching" class="list-state">Searching...</div>

          <button
            v-else
            v-for="user in searchResults"
            :key="user.id"
            class="chat-item"
            type="button"
            @click="$emit('start-chat', user.id, user.nickname)"
          >
            <div class="chat-avatar">
              <img v-if="user.photo_url" :src="user.photo_url" alt="" class="avatar-image" />
              <span v-else>{{ (user.name || user.nickname)?.[0]?.toUpperCase() || '?' }}</span>
            </div>
            <div class="chat-body">
              <div class="chat-topline">
                <!-- Верхняя строка: имя пользователя (name или nickname) -->
                <span class="chat-name">{{ user.name || user.nickname || 'Unknown' }}</span>
              </div>
              <div class="chat-bottomline">
                <!-- Нижняя строка: @nickname -->
                <span class="chat-preview">@{{ user.nickname || user.username || '' }}</span>
              </div>
            </div>
          </button>

          <div v-if="!isSearching && searchResults.length === 0" class="list-state">No users found</div>
        </template>

        <!-- Список диалогов -->
        <template v-else>
          <div
            v-for="direct in filteredDirects"
            :key="direct.id"
            class="chat-item-wrap"
          >
            <button
              class="chat-item"
              :class="{ active: String(activeId) === String(direct.id) }"
              type="button"
              @click="!deleteMode && $emit('select', { chatId: direct.id, recipientId: getRecipientId(direct) })"
            >
              <div class="chat-avatar">
                <img v-if="getAvatarUrl(direct)" :src="getAvatarUrl(direct)" alt="" class="avatar-image" />
                <span v-else>{{ getAvatarLetter(direct) }}</span>
              </div>

              <div class="chat-body">
                <div class="chat-topline">
                  <span class="chat-name">{{ getRecipientName(direct) }}</span>
                  <span class="chat-time">{{ getChatTime(direct) }}</span>
                </div>
                <div class="chat-bottomline">
                  <span class="chat-preview">{{ getChatPreview(direct) }}</span>
                  <span v-if="getUnreadCount(direct) > 0" class="unread-badge">{{ getUnreadCount(direct) }}</span>
                </div>
              </div>
            </button>

            <!-- Кнопка удаления чата (только в режиме deleteMode) -->
            <button
              v-if="deleteMode"
              class="delete-chat-btn"
              type="button"
              title="Delete chat"
              @click="askDeleteChat(direct)"
            >
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.8">
                <path d="M3 6h18"></path>
                <path d="M8 6V4.8c0-.99.81-1.8 1.8-1.8h4.4c.99 0 1.8.81 1.8 1.8V6"></path>
                <path d="M18.2 6l-.72 11.02A2 2 0 0 1 15.48 19H8.52a2 2 0 0 1-1.99-1.98L5.8 6"></path>
              </svg>
            </button>
          </div>

          <div v-if="filteredDirects.length === 0" class="list-state">
            <span v-if="activeTab === 'chats' || activeTab === 'all'">No chats yet. Search to start one.</span>
            <span v-else>No {{ activeTab }} yet.</span>
          </div>
        </template>
      </div>

      <footer class="sidebar-footer">
        <div class="footer-actions">
          <button class="footer-btn" title="Settings" type="button">
            <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="1.8">
              <circle cx="12" cy="12" r="3"></circle>
              <path d="M19.4 15a1.7 1.7 0 0 0 .33 1.82l.05.05a2 2 0 1 1-2.83 2.83l-.05-.05a1.7 1.7 0 0 0-1.82-.33 1.7 1.7 0 0 0-1.03 1.56V21a2 2 0 1 1-4 0v-.08a1.7 1.7 0 0 0-1.03-1.56 1.7 1.7 0 0 0-1.82.33l-.05.05a2 2 0 1 1-2.83-2.83l.05-.05A1.7 1.7 0 0 0 4.6 15a1.7 1.7 0 0 0-1.56-1.03H3a2 2 0 1 1 0-4h.08A1.7 1.7 0 0 0 4.64 8.94a1.7 1.7 0 0 0-.33-1.82l-.05-.05a2 2 0 1 1 2.83-2.83l.05.05A1.7 1.7 0 0 0 8.96 4.6a1.7 1.7 0 0 0 1.03-1.56V3a2 2 0 1 1 4 0v.08a1.7 1.7 0 0 0 1.03 1.56 1.7 1.7 0 0 0 1.82-.33l.05-.05a2 2 0 1 1 2.83 2.83l-.05.05A1.7 1.7 0 0 0 19.4 8.94c.12.39.6 1.03 1.56 1.03H21a2 2 0 1 1 0 4h-.08A1.7 1.7 0 0 0 19.4 15z"></path>
            </svg>
          </button>

          <!-- Кнопка переключения темы -->
          <button class="footer-btn" title="Toggle theme" type="button" @click="toggleTheme">
            <svg v-if="!isLight" viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="1.8">
              <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path>
            </svg>
            <svg v-else viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="1.8">
              <circle cx="12" cy="12" r="5"></circle>
              <line x1="12" y1="1" x2="12" y2="3"></line>
              <line x1="12" y1="21" x2="12" y2="23"></line>
              <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line>
              <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line>
              <line x1="1" y1="12" x2="3" y2="12"></line>
              <line x1="21" y1="12" x2="23" y2="12"></line>
              <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line>
              <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line>
            </svg>
          </button>
        </div>
      </footer>
    </div>

    <!-- Модалка подтверждения удаления чата -->
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
  </aside>
</template>

<script>
const BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

export default {
  name: 'ChatSidebar',

  props: {
    directs: { type: Array, default: () => [] },
    activeId: { type: [String, Number], default: null },
    currentUserId: { type: String, default: null },
    isLight: { type: Boolean, default: false }
  },

  emits: ['select', 'start-chat', 'toggle-theme', 'chat-deleted'],

  data() {
    return {
      search: '',
      searchResults: [],
      isSearching: false,
      searchTimeout: null,
      activeTab: 'chats',
      deleteMode: false,
      chatToDelete: null
    }
  },

  computed: {
    filteredDirects() {
      if (this.activeTab === 'all' || this.activeTab === 'chats') {
        return this.directs
      }
      return []
    }
  },

  watch: {
    search(value) {
      if (!value.trim()) {
        this.searchResults = []
        this.isSearching = false
        return
      }
      clearTimeout(this.searchTimeout)
      this.searchTimeout = setTimeout(() => {
        this.fetchUsers(value.trim())
      }, 350)
    }
  },

  methods: {
    toggleTheme() {
      this.$emit('toggle-theme')
    },

    async fetchUsers(query) {
      this.isSearching = true
      try {
        const url = new URL(`${BASE}/users/search`)
        url.searchParams.set('q', query)
        url.searchParams.set('user_id', this.currentUserId || '')

        const res = await fetch(url.toString(), {
          headers: { Authorization: `Bearer ${sessionStorage.getItem('access_token') || ''}` }
        })

        if (!res.ok) throw new Error('Search error')

        const data = await res.json()
        this.searchResults = data.users || []
      } catch (e) {
        console.error('Search error', e)
        this.searchResults = []
      } finally {
        this.isSearching = false
      }
    },

    getRecipientId(direct) {
      const first = direct.first_user_id ?? direct.firstuserid
      const second = direct.second_user_id ?? direct.seconduserid
      return String(first) === String(this.currentUserId) ? second : first
    },

    getRecipientName(direct) {
      return direct.companion_name || direct.companion_nickname || direct.nickname || 'Unknown'
    },

    getAvatarLetter(direct) {
      return this.getRecipientName(direct)?.[0]?.toUpperCase() || '?'
    },

    getAvatarUrl(direct) {
      return direct.companion_photo_url || direct.photo_url || null
    },

    getChatPreview(direct) {
      return direct.last_message_content || ''
    },

    getUnreadCount(direct) {
      return Number(direct.unread_count || direct.unreadcount || 0)
    },

    getChatTime(direct) {
      const raw = direct.last_message_at || direct.updated_at || direct.created_at || direct.creation_time
      if (!raw) return ''
      const date = new Date(raw)
      if (Number.isNaN(date.getTime())) return ''
      return date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
    },

    askDeleteChat(direct) {
      this.chatToDelete = direct
    },

    async confirmDeleteChat() {
      if (!this.chatToDelete) return
      const chatId = this.chatToDelete.id
      this.chatToDelete = null

      try {
        const res = await fetch(`${BASE}/chat/direct/${chatId}`, {
          method: 'DELETE',
          headers: { Authorization: `Bearer ${sessionStorage.getItem('access_token') || ''}` }
        })
        if (!res.ok) {
          const text = await res.text()
          console.error('Delete chat failed', res.status, text)
          return
        }
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
  width: 100%;
  height: 100%;
  background: transparent;
  font-family: 'Satoshi', sans-serif;
  position: relative;
}

/* ===== DARK (default) ===== */
.sidebar-shell {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: linear-gradient(180deg, rgba(8, 12, 26, 0.98), rgba(7, 10, 22, 0.98));
}

/* ===== LIGHT THEME ===== */
.theme-light .sidebar-shell {
  background: #ffffff;
}
.theme-light .sidebar-header {
  border-bottom: none;
}
.theme-light .brand-main { color: #1a1d2e; }
.theme-light .brand-accent { color: #5b6aff; }
.theme-light .brand-mark {
  background: linear-gradient(135deg, #5b6aff, #7b68ff);
}
.theme-light .header-btn {
  color: #5b6aff;
  background: rgba(91, 106, 255, 0.08);
  border-color: rgba(91, 106, 255, 0.15);
}
.theme-light .search-box {
  background: #f3f4f8;
  border-color: #e2e4ee;
}
.theme-light .search-icon { color: #9098b8; }
.theme-light .search-input { color: #1a1d2e; }
.theme-light .search-input::placeholder { color: #aab0cc; }
.theme-light .tab-btn { color: #8890b4; }
.theme-light .tab-btn.active {
  color: #1a1d2e;
  background: rgba(91, 106, 255, 0.1);
  border-color: rgba(91, 106, 255, 0.18);
}
.theme-light .chat-item:hover { background: #f3f4f8; }
.theme-light .chat-item.active {
  background: linear-gradient(180deg, rgba(91, 106, 255, 0.1), rgba(91, 106, 255, 0.07));
  border-color: rgba(91, 106, 255, 0.22);
}
.theme-light .chat-name { color: #1a1d2e; }
.theme-light .chat-time { color: #9098b8; }
.theme-light .chat-preview { color: #7880a0; }
.theme-light .list-state { color: #9098b8; }
.theme-light .sidebar-footer { border-top-color: #e8eaf0; }
.theme-light .footer-btn {
  color: #7880a0;
  background: #f3f4f8;
  border-color: #e2e4ee;
}
.theme-light .delete-chat-btn {
  background: rgba(255, 60, 80, 0.06);
  border-color: rgba(255, 60, 80, 0.15);
  color: #ff3c50;
}

/* ===== HEADER ===== */
.sidebar-header {
  height: 78px;
  padding: 18px 16px 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.brand { display: flex; align-items: center; gap: 10px; }
.brand-mark {
  width: 31px; height: 31px;
  border-radius: 10px;
  display: grid; place-items: center;
  color: #ffffff;
  background: linear-gradient(135deg, #6675ff, #7b68ff);
  box-shadow: 0 8px 18px rgba(90, 98, 255, 0.22);
}
.brand-text { font-size: 18px; font-weight: 800; letter-spacing: -0.02em; }
.brand-main { color: #f3f5ff; }
.brand-accent { color: #92a0ff; }

.header-btn {
  width: 32px; height: 32px;
  border-radius: 10px;
  display: grid; place-items: center;
  color: #98a2ca;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.04);
  cursor: pointer;
  transition: all 0.2s;
}
.header-btn.delete-mode-active {
  color: #ff4d6d;
  background: rgba(255, 77, 109, 0.12);
  border-color: rgba(255, 77, 109, 0.25);
}

/* ===== SEARCH ===== */
.search-wrap { padding: 0 16px 12px; }
.search-box {
  height: 40px; display: flex; align-items: center; gap: 9px; padding: 0 12px;
  border-radius: 13px;
  background: rgba(255, 255, 255, 0.028);
  border: 1px solid rgba(255, 255, 255, 0.035);
}
.search-icon { color: #7480a8; }
.search-input {
  flex: 1; min-width: 0; background: transparent; border: none; outline: none;
  color: #eef1ff; font-size: 13px; font-weight: 500;
}
.search-input::placeholder { color: #6d7798; }

/* ===== TABS ===== */
.sidebar-tabs { display: flex; gap: 8px; padding: 0 16px 14px; overflow-x: auto; }
.sidebar-tabs::-webkit-scrollbar { display: none; }
.tab-btn {
  height: 28px; padding: 0 11px; border-radius: 10px; white-space: nowrap;
  color: #7d87ab; font-size: 11px; font-weight: 700;
  background: transparent; border: 1px solid transparent; cursor: pointer; transition: all 0.2s;
}
.tab-btn.active {
  color: #f1f4ff;
  background: rgba(96, 108, 255, 0.14);
  border-color: rgba(114, 126, 255, 0.16);
}

/* ===== LIST ===== */
.sidebar-list { flex: 1; overflow-y: auto; min-height: 0; padding: 2px 10px 12px; }
.sidebar-list::-webkit-scrollbar { width: 6px; }
.sidebar-list::-webkit-scrollbar-thumb { background: rgba(147, 158, 211, 0.16); border-radius: 999px; }

.chat-item-wrap {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
}

.chat-item {
  flex: 1;
  display: flex; align-items: center; gap: 11px;
  padding: 11px 10px;
  border-radius: 16px;
  text-align: left;
  background: transparent; border: 1px solid transparent;
  transition: all 0.16s ease; cursor: pointer;
}
.chat-item:hover { background: rgba(255, 255, 255, 0.025); }
.chat-item.active {
  background: linear-gradient(180deg, rgba(75, 88, 228, 0.17), rgba(64, 78, 210, 0.12));
  border-color: rgba(110, 122, 255, 0.26);
  box-shadow: 0 0 0 1px rgba(98, 112, 255, 0.08) inset, 0 10px 22px rgba(49, 61, 180, 0.16);
}

.delete-chat-btn {
  flex-shrink: 0;
  width: 30px; height: 30px;
  border-radius: 10px;
  display: grid; place-items: center;
  color: #ff4d6d;
  background: rgba(255, 77, 109, 0.1);
  border: 1px solid rgba(255, 77, 109, 0.2);
  cursor: pointer;
  transition: all 0.15s;
}
.delete-chat-btn:hover {
  background: rgba(255, 77, 109, 0.2);
}

.chat-avatar {
  width: 42px; height: 42px; border-radius: 14px; flex-shrink: 0;
  display: grid; place-items: center; overflow: hidden;
  color: #fff; font-size: 13px; font-weight: 700;
  background: linear-gradient(135deg, #6572ff, #8a67ff);
}
.avatar-image { width: 100%; height: 100%; object-fit: cover; }
.chat-body { flex: 1; min-width: 0; }
.chat-topline, .chat-bottomline {
  display: flex; align-items: center; justify-content: space-between; gap: 8px;
}
.chat-topline { margin-bottom: 3px; }
.chat-name { color: #eef2ff; font-size: 13px; font-weight: 700; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.chat-time { flex-shrink: 0; color: #7580a6; font-size: 10.5px; font-weight: 600; }
.chat-preview { color: #8590b4; font-size: 11.5px; font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.unread-badge {
  min-width: 18px; height: 18px; padding: 0 5px; border-radius: 999px;
  display: inline-grid; place-items: center; flex-shrink: 0;
  color: #fff; font-size: 10px; font-weight: 700;
  background: linear-gradient(135deg, #6a76ff, #8866ff);
}
.list-state { padding: 26px 12px; text-align: center; color: #7c86ad; font-size: 12px; }

/* ===== FOOTER ===== */
.sidebar-footer {
  padding: 12px 16px 14px;
  border-top: 1px solid rgba(255, 255, 255, 0.03);
}
.footer-actions { display: flex; gap: 10px; }
.footer-btn {
  width: 34px; height: 34px; border-radius: 11px;
  display: grid; place-items: center;
  color: #97a2c8;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.04);
  cursor: pointer; transition: all 0.2s;
}
.footer-btn:hover { background: rgba(255, 255, 255, 0.05); }

/* ===== MODAL ===== */
.modal-overlay {
  position: absolute; inset: 0;
  background: rgba(5, 8, 20, 0.55);
  backdrop-filter: blur(4px);
  display: grid; place-items: center;
  z-index: 20;
  border-radius: inherit;
}
.theme-light .modal-overlay { background: rgba(200, 205, 230, 0.5); }

.confirm-modal {
  background: linear-gradient(180deg, rgba(22, 28, 52, 0.97), rgba(16, 20, 38, 0.99));
  border: 1px solid rgba(132, 144, 224, 0.15);
  border-radius: 16px; padding: 24px; width: 260px;
  text-align: center;
  box-shadow: 0 20px 40px rgba(0,0,0,0.4);
}
.theme-light .confirm-modal {
  background: #ffffff;
  border-color: #dde1f0;
  box-shadow: 0 12px 40px rgba(90, 106, 200, 0.15);
}
.confirm-modal h3 { color: #eef2ff; font-size: 15px; margin-bottom: 8px; }
.confirm-modal p { color: #8d96ba; font-size: 12px; line-height: 1.6; margin-bottom: 20px; }
.theme-light .confirm-modal h3 { color: #1a1d2e; }
.theme-light .confirm-modal p { color: #7880a0; }
.modal-actions { display: flex; gap: 10px; justify-content: center; }
.btn-cancel {
  padding: 8px 14px; border-radius: 10px;
  background: rgba(255,255,255,0.05); border: 1px solid rgba(255,255,255,0.06);
  color: #a6afd4; font-size: 13px; cursor: pointer;
}
.theme-light .btn-cancel { background: #f3f4f8; border-color: #e2e4ee; color: #7880a0; }
.btn-delete {
  padding: 8px 14px; border-radius: 10px;
  background: linear-gradient(135deg, #ff4d6d, #d93856);
  border: none; color: white; font-size: 13px; cursor: pointer;
}
</style>