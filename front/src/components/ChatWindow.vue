<template>
  <section class="chat-window" :class="{ 'theme-light': isLight }">
    <header class="chat-header">
      <div class="chat-user">
        <div class="chat-avatar">
          <img v-if="avatarUrl" :src="avatarUrl" alt="" class="chat-avatar-image" />
          <span v-else>{{ avatarLetter }}</span>
        </div>
        <div class="chat-user-meta">
          <div class="chat-username">{{ chatTitle }}</div>
          <div class="chat-status" :class="{ online: isOnline, offline: !isOnline }">
            <span class="status-dot"></span>
            <span v-if="isOnline">Online</span>
            <span v-else-if="lastSeen">{{ formatLastSeen(lastSeen) }}</span>
            <span v-else>Offline</span>
          </div>
        </div>
      </div>

      <div class="chat-actions">
        <button class="chat-icon-btn" title="Search" type="button">
          <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="1.8">
            <circle cx="11" cy="11" r="7"></circle>
            <path d="M20 20l-3.5-3.5"></path>
          </svg>
        </button>
        <button class="chat-icon-btn" title="More" type="button">
          <svg viewBox="0 0 24 24" width="17" height="17" fill="currentColor">
            <circle cx="12" cy="5" r="1.8"></circle>
            <circle cx="12" cy="12" r="1.8"></circle>
            <circle cx="12" cy="19" r="1.8"></circle>
          </svg>
        </button>
      </div>
    </header>

    <div class="messages-area-wrapper">
      <div class="messages-area" ref="messagesArea">
        <div v-if="messages.length === 0" class="day-separator">Today</div>

        <MessageBubble
          v-for="message in messages"
          :key="message.id"
          :message="message"
          :isMine="isMine(message)"
          :isLight="isLight"
          @delete="confirmDelete"
        />
      </div>

      <!-- Модалка подтверждения удаления сообщения -->
      <div v-if="deleteModalOpen" class="delete-modal-overlay">
        <div class="delete-modal">
          <h3>Delete message?</h3>
          <p>This will permanently delete the message for both users.</p>
          <div class="modal-actions">
            <button type="button" class="btn-cancel" @click="closeDeleteModal">Cancel</button>
            <button type="button" class="btn-delete" @click="executeDelete">Delete</button>
          </div>
        </div>
      </div>
    </div>

    <div class="composer-wrap">
      <form class="composer" @submit.prevent="sendMessage">
        <button type="button" class="composer-side-btn" title="Attach">
          <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="1.8">
            <path d="M21.44 11.05l-8.49 8.49a5.5 5.5 0 0 1-7.78-7.78l9.2-9.19a3.5 3.5 0 0 1 4.95 4.95l-9.19 9.2a1.5 1.5 0 0 1-2.12-2.13l8.49-8.48"></path>
          </svg>
        </button>

        <input
          v-model="newMessage"
          type="text"
          class="message-input"
          placeholder="Type a message..."
        />

        <button type="button" class="composer-side-btn" title="Emoji">
          <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="1.8">
            <circle cx="12" cy="12" r="9"></circle>
            <path d="M8 14s1.5 2 4 2 4-2 4-2"></path>
            <path d="M9 9h.01"></path>
            <path d="M15 9h.01"></path>
          </svg>
        </button>

        <button type="submit" class="send-btn" title="Send">
          <svg viewBox="0 0 24 24" width="17" height="17" fill="currentColor">
            <path d="M21.8 2.2a1 1 0 0 0-1.04-.23L2.76 8.97a1 1 0 0 0 .08 1.89l7.14 2.38 2.38 7.14a1 1 0 0 0 .91.68h.05a1 1 0 0 0 .9-.59l7-18a1 1 0 0 0-.22-1.03z"></path>
          </svg>
        </button>
      </form>
    </div>
  </section>
</template>

<script>
import MessageBubble from './MessageBubble.vue'

const BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'
const STATUS_POLL_INTERVAL = 5000 // 5 сек вместо 20, статус живее

export default {
  name: 'ChatWindow',
  components: { MessageBubble },

  props: {
    chat: { type: Object, default: null },
    chatId: { type: [String, Number], default: null },
    currentUserId: { type: [String, Number], default: null },
    recipientId: { type: [String, Number], default: null },
    selectedCompanion: { type: Object, default: null },
    isLight: { type: Boolean, default: false }
  },

  emits: ['message-sent'],

  data() {
    return {
      messages: [],
      newMessage: '',
      deleteModalOpen: false,
      messageToDelete: null,
      loading: false,
      isOnline: false,
      lastSeen: null,
      statusPollTimer: null
    }
  },

  computed: {
    chatTitle() {
      return (
        this.selectedCompanion?.nickname ||
        this.selectedCompanion?.username ||
        this.chat?.companion_name ||
        this.chat?.companion_nickname ||
        this.chat?.nickname ||
        this.chat?.username ||
        (this.recipientId ? `User ${String(this.recipientId).slice(0, 6)}` : 'Unknown')
      )
    },

    avatarUrl() {
      return (
        this.selectedCompanion?.photo_url ||
        this.chat?.companion_photo_url ||
        this.chat?.photo_url ||
        null
      )
    },

    avatarLetter() {
      return this.chatTitle?.[0]?.toUpperCase() || ''
    }
  },

  watch: {
    chatId: {
      immediate: true,
      async handler(value) {
        // сбрасываем старый статус при смене чата
        this.isOnline = false
        this.lastSeen = null

        if (value) {
          await this.loadMessages()
          this.startStatusPolling()
        } else {
          this.messages = []
          this.stopStatusPolling()
        }
      }
    },

    recipientId: {
      immediate: true,
      handler(value) {
        // сбрасываем статус при смене собеседника
        this.isOnline = false
        this.lastSeen = null

        if (value) {
          this.fetchStatus()
        }
      }
    }
  },

  beforeUnmount() {
    this.stopStatusPolling()
  },

  methods: {
    startStatusPolling() {
      this.stopStatusPolling()
      this.fetchStatus()
      this.statusPollTimer = setInterval(() => {
        this.fetchStatus()
      }, STATUS_POLL_INTERVAL)
    },

    stopStatusPolling() {
      if (this.statusPollTimer) {
        clearInterval(this.statusPollTimer)
        this.statusPollTimer = null
      }
    },

    async fetchStatus() {
      const targetId =
        this.recipientId ||
        this.selectedCompanion?.id ||
        this.chat?.companion_id ||
        this.chat?.companionid

      if (!targetId) return

      try {
        const res = await fetch(`${BASE}/users/${targetId}/status`, {
          headers: { Authorization: `Bearer ${sessionStorage.getItem('access_token') || ''}` }
        })
        if (!res.ok) return
        const data = await res.json()
        this.isOnline = data.online === true
        this.lastSeen = data.lastseen || data.last_seen || null
      } catch {
        // не ломаем UI
      }
    },

    formatLastSeen(raw) {
      if (!raw) return 'Offline'

      const normalized = raw.endsWith('Z') || raw.includes('+') ? raw : raw + 'Z'
      const date = new Date(normalized)
      if (isNaN(date.getTime())) return 'Offline'

      const diff = Math.floor((Date.now() - date.getTime()) / 1000)

      if (diff < 60) return `last seen ${diff}s ago`
      if (diff < 3600) return `last seen ${Math.floor(diff / 60)}m ago`
      if (diff < 86400)
        return `last seen today at ${date.toLocaleTimeString('en', {
          hour: '2-digit',
          minute: '2-digit'
        })}`
      if (diff < 172800)
        return `last seen yesterday at ${date.toLocaleTimeString('en', {
          hour: '2-digit',
          minute: '2-digit'
        })}`
      return `last seen ${date.toLocaleDateString('en', { day: 'numeric', month: 'short' })}`
    },

    normalizeMessage(message) {
      return {
        ...message,
        sender_id: message.sender_id ?? message.senderid,
        created_at: message.created_at ?? message.createdat,
        status: message.status || 'delivered'
      }
    },

    async loadMessages(skipReadPatch = false) {
      if (!this.chatId) return
      this.loading = true
      try {
        const url = new URL(`${BASE}/chat/messages`)
        url.searchParams.set('chat_id', this.chatId)

        const res = await fetch(url.toString(), {
          headers: { Authorization: `Bearer ${sessionStorage.getItem('access_token') || ''}` }
        })

        if (!res.ok) {
          console.error('Failed to fetch messages, status', res.status)
          return
        }

        const data = await res.json()
        const apiMessages = Array.isArray(data.messages) ? data.messages : []

        this.messages = apiMessages
          .map(this.normalizeMessage)
          .sort((a, b) => new Date(a.created_at) - new Date(b.created_at))

        if (!skipReadPatch) await this.markIncomingAsRead()
        this.scrollToBottom()
      } catch (e) {
        console.error('Failed to load messages', e)
      } finally {
        this.loading = false
      }
    },

    async markIncomingAsRead() {
      if (!this.chatId || !this.currentUserId) return
      const hasUnread = this.messages.some(
        m => String(m.sender_id) !== String(this.currentUserId) && m.status !== 'read'
      )
      if (!hasUnread) return
      try {
        const url = new URL(`${BASE}/chat/messages/read`)
        url.searchParams.set('chatid', this.chatId)
        url.searchParams.set('userid', this.currentUserId)

        const res = await fetch(url.toString(), {
          method: 'PATCH',
          headers: { Authorization: `Bearer ${sessionStorage.getItem('access_token') || ''}` }
        })
        if (!res.ok) return
        await this.loadMessages(true)
      } catch (e) {
        console.error('markIncomingAsRead error', e)
      }
    },

    scrollToBottom() {
      this.$nextTick(() => {
        const el = this.$refs.messagesArea
        if (el) el.scrollTop = el.scrollHeight
      })
    },

    isMine(message) {
      return String(message.sender_id ?? message.senderid) === String(this.currentUserId)
    },

    async sendMessage() {
      const text = this.newMessage.trim()
      if (!text || !this.chatId) return

      const optimistic = {
        id: `local-${Date.now()}`,
        sender_id: this.currentUserId,
        content: text,
        created_at: new Date().toISOString(),
        status: 'delivered'
      }

      this.messages.push(optimistic)
      this.newMessage = ''
      this.scrollToBottom()

      this.$emit('message-sent', {
        chatId: this.chatId,
        content: text,
        date: optimistic.created_at
      })

      try {
        const res = await fetch(`${BASE}/chat/messages`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${sessionStorage.getItem('access_token') || ''}`
          },
          body: JSON.stringify({
            chat_id: this.chatId,
            sender_id: this.currentUserId,
            content: text
          })
        })

        if (!res.ok) {
          console.error('Failed to send message, status', res.status)
          return
        }

        const data = await res.json()
        const savedRaw = data.message || null

        if (savedRaw) {
          const saved = this.normalizeMessage(savedRaw)
          this.messages = this.messages.map(m => (m.id === optimistic.id ? saved : m))
        } else {
          this.messages = this.messages.map(m =>
            m.id === optimistic.id ? { ...m, status: 'delivered' } : m
          )
        }

        await this.loadMessages()
      } catch (e) {
        console.error('Failed to send message', e)
      }
    },

    confirmDelete(messageId) {
      this.messageToDelete = messageId
      this.deleteModalOpen = true
    },

    closeDeleteModal() {
      this.deleteModalOpen = false
      this.messageToDelete = null
    },

    async executeDelete() {
      if (!this.messageToDelete) return
      const messageId = this.messageToDelete
      this.closeDeleteModal()

      try {
        const res = await fetch(`${BASE}/chat/messages/${messageId}`, {
          method: 'DELETE',
          headers: { Authorization: `Bearer ${sessionStorage.getItem('access_token') || ''}` }
        })

        if (!res.ok) {
          const text = await res.text()
          console.error('Delete message failed', res.status, text)
          return
        }

        this.messages = this.messages.filter(m => m.id !== messageId)
      } catch (e) {
        console.error('Delete message error', e)
      }
    }
  }
}
</script>

<style scoped>
.chat-window {
  height: 100%; max-height: 100%; min-height: 0;
  display: flex; flex-direction: column; min-width: 0; overflow: hidden;
  background: linear-gradient(180deg, rgba(10, 13, 28, 0.72), rgba(7, 10, 22, 0.82));
  transition: background 0.3s;
}
.chat-window.theme-light {
  background: #f5f6fc;
}

/* HEADER */
.chat-header {
  height: 78px; min-height: 78px; flex-shrink: 0;
  padding: 14px 20px;
  display: flex; align-items: center; justify-content: space-between;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  background: rgba(255, 255, 255, 0.015);
  transition: all 0.3s;
}
.theme-light .chat-header {
  background: #ffffff;
  border-bottom-color: #e4e6f0;
  box-shadow: 0 1px 0 #e4e6f0;
}

.chat-user { display: flex; align-items: center; gap: 12px; min-width: 0; }
.chat-avatar {
  width: 38px; height: 38px; border-radius: 50%;
  display: grid; place-items: center; overflow: hidden; flex-shrink: 0;
  background: linear-gradient(135deg, #6d78ff, #8866ff);
  color: white; font-weight: 700; font-size: 13px;
}
.chat-avatar-image { width: 100%; height: 100%; object-fit: cover; }
.chat-user-meta { min-width: 0; }
.chat-username { color: #f2f4ff; font-size: 14px; font-weight: 700; line-height: 1.2; }
.theme-light .chat-username { color: #1a1d2e; }

/* Статус */
.chat-status {
  margin-top: 3px;
  font-size: 11px; font-weight: 600;
  display: flex; align-items: center; gap: 5px;
  color: #8691b7;
  transition: color 0.3s;
}
.theme-light .chat-status { color: #9098b8; }
.status-dot {
  width: 7px; height: 7px; border-radius: 50%;
  background: #4d5270;
  flex-shrink: 0;
  transition: background 0.3s;
}
.chat-status.online .status-dot { background: #22c55e; }
.chat-status.online { color: #22c55e; }
.theme-light .chat-status.online { color: #16a34a; }
.theme-light .chat-status.online .status-dot { background: #16a34a; }

.chat-actions { display: flex; align-items: center; gap: 8px; }
.chat-icon-btn {
  width: 32px; height: 32px; border-radius: 10px;
  display: grid; place-items: center;
  color: #95a0c8; background: transparent; border: 1px solid transparent; cursor: pointer;
}
.theme-light .chat-icon-btn { color: #7880a0; }
.chat-icon-btn:hover { background: rgba(255, 255, 255, 0.035); border-color: rgba(255, 255, 255, 0.04); }
.theme-light .chat-icon-btn:hover { background: #f0f1f8; border-color: #e4e6f0; }

/* MESSAGES */
.messages-area-wrapper {
  flex: 1; min-height: 0; overflow: hidden;
  position: relative; display: flex; flex-direction: column;
}
.messages-area {
  flex: 1; min-height: 0; overflow-y: auto; overflow-x: hidden;
  padding: 22px 28px 18px;
  display: flex; flex-direction: column; gap: 6px;
  -webkit-overflow-scrolling: touch; overscroll-behavior: contain;
}
.messages-area::-webkit-scrollbar { width: 6px; }
.messages-area::-webkit-scrollbar-thumb { background: rgba(148, 159, 212, 0.16); border-radius: 999px; }

.day-separator {
  align-self: center; margin: 4px 0 16px;
  padding: 5px 12px; border-radius: 999px;
  color: #8a94bc; font-size: 11px; font-weight: 700;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.04);
}
.theme-light .day-separator { background: rgba(91, 106, 255, 0.07); border-color: rgba(91, 106, 255, 0.1); color: #7880a0; }

/* COMPOSER */
.composer-wrap {
  flex-shrink: 0; padding: 14px 28px 18px;
  background: linear-gradient(180deg, rgba(8, 12, 24, 0.18), rgba(8, 12, 24, 0.32));
}
.theme-light .composer-wrap { background: #f5f6fc; }

.composer {
  height: 56px; display: flex; align-items: center; gap: 12px;
  padding: 0 12px 0 14px; border-radius: 18px;
  border: 1px solid rgba(110, 123, 255, 0.18);
  background: linear-gradient(180deg, rgba(25, 30, 58, 0.68), rgba(18, 23, 46, 0.78));
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.18), inset 0 1px 0 rgba(255, 255, 255, 0.04);
}
.theme-light .composer {
  background: #ffffff;
  border-color: rgba(91, 106, 255, 0.2);
  box-shadow: 0 4px 16px rgba(91, 106, 200, 0.08);
}

.composer-side-btn, .send-btn {
  width: 34px; height: 34px; border-radius: 11px;
  display: grid; place-items: center; flex-shrink: 0; cursor: pointer; border: none;
}
.composer-side-btn {
  color: #a6afd4;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.04);
}
.theme-light .composer-side-btn { color: #9098b8; background: #f3f4f8; border-color: #e4e6f0; }

.message-input {
  flex: 1; min-width: 0; background: transparent; border: none; outline: none;
  color: #eef2ff; font-size: 14px; font-weight: 500;
}
.theme-light .message-input { color: #1a1d2e; }
.message-input::placeholder { color: #747ea2; }
.theme-light .message-input::placeholder { color: #aab0cc; }

.send-btn {
  color: white;
  background: linear-gradient(135deg, #6e79ff, #8669ff);
  box-shadow: 0 8px 18px rgba(94, 102, 255, 0.28);
}

/* DELETE MODAL */
.delete-modal-overlay {
  position: absolute; inset: 0;
  background: rgba(5, 8, 18, 0.6); backdrop-filter: blur(4px);
  display: grid; place-items: center; z-index: 10;
}
.theme-light .delete-modal-overlay { background: rgba(200, 205, 230, 0.5); }

.delete-modal {
  background: linear-gradient(180deg, rgba(22, 28, 52, 0.95), rgba(16, 20, 38, 0.98));
  border: 1px solid rgba(132, 144, 224, 0.15);
  border-radius: 16px; padding: 24px; width: 300px;
  text-align: center; box-shadow: 0 20px 40px rgba(0, 0, 0, 0.4);
}
.theme-light .delete-modal {
  background: #ffffff;
  border-color: #dde1f0;
  box-shadow: 0 12px 40px rgba(90, 106, 200, 0.15);
}
.delete-modal h3 { color: #eef2ff; font-size: 16px; margin-bottom: 8px; }
.delete-modal p { color: #8d96ba; font-size: 13px; margin-bottom: 20px; }
.theme-light .delete-modal h3 { color: #1a1d2e; }
.theme-light .delete-modal p { color: #7880a0; }
.modal-actions { display: flex; gap: 10px; justify-content: center; }
.btn-cancel {
  padding: 8px 16px; border-radius: 10px;
  background: rgba(255,255,255,0.05); border: 1px solid rgba(255,255,255,0.05);
  color: #a6afd4; font-size: 13px; cursor: pointer;
}
.theme-light .btn-cancel { background: #f3f4f8; border-color: #e2e4ee; color: #7880a0; }
.btn-delete {
  padding: 8px 16px; border-radius: 10px;
  background: linear-gradient(135deg, #ff4d6d, #d93856);
  border: none; color: white; font-size: 13px; cursor: pointer;
}

@media (max-width: 760px) {
  .chat-header { padding: 12px 16px; }
  .messages-area { padding: 18px 16px 12px; }
  .composer-wrap { padding: 10px 16px 16px; }
}
</style>