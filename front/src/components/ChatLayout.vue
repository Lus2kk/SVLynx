<template>
  <div class="direct-page" :class="{ 'theme-light': isLight }">
    <div class="direct-shell" :class="{ 'theme-light': isLight }">
      <ChatSidebar
        :directs="directs"
        :activeId="activeChatId"
        :currentUserId="currentUserId"
        :isLight="isLight"
        :class="{ 'mobile-hidden': mobileView === 'chat' }"
        :userStatuses="userStatuses"
        @select="selectChat"
        @start-chat="startChat"
        @toggle-theme="toggleTheme"
        @chat-deleted="onChatDeleted"
        @open-profile="$emit('open-profile')"
      />


      <div class="content-area" :class="{ 'mobile-hidden': mobileView === 'sidebar' }">
       <ChatWindow
  v-if="activeChat"
  ref="chatWindow"
  :key="`${activeChatId}-${activeRecipientId}-${chatWindowKey}`"
  :chat="activeChat"
  :chatId="activeChatId"
  :currentUserId="currentUserId"
  :currentUserName="currentUserName"
  :recipientId="activeRecipientId"
  :presence="activePresence"
  :isLight="isLight"
  :showBackButton="isMobile"
  :isVisible="!isMobile || mobileView === 'chat'"
  :isTyping="isTyping"
  @message-sent="updateChatPreview"
  @message-deleted="onMessageDeleted"
  @mark-as-read="onMarkAsRead"
  @typing="onTyping"
  @back="goBackToSidebar"
/>


        <div v-else class="empty-chat">
          <div class="empty-card">
            <div class="empty-logo">
              <svg viewBox="0 0 24 24" width="28" height="28" fill="none" stroke="currentColor" stroke-width="1.9">
                <path d="M4 6.5A2.5 2.5 0 0 1 6.5 4h11A2.5 2.5 0 0 1 20 6.5v7A2.5 2.5 0 0 1 17.5 16H11l-4.5 4V16A2.5 2.5 0 0 1 4 13.5v-7z" />
              </svg>
            </div>
            <h2 class="empty-title">Ваши сообщения</h2>
            <p class="empty-text">Выберите чат или найдите собеседника через поиск.</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>


<script>
import ChatSidebar from './ChatSidebar.vue'
import ChatWindow from './ChatWindow.vue'
import { apiFetch, getCookie } from '../api.js'


const BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'
const WS_BASE = BASE.replace(/^http/, 'ws')


export default {
  name: 'ChatLayout',
  components: { ChatSidebar, ChatWindow },

  emits: ['theme-changed', 'open-profile'],

  data() {
    return {
      directs: [],
      activeChatId: null,
      activeRecipientId: null,
      currentUserId: null,
      loading: false,
      isLight: localStorage.getItem('svlynx-theme') === 'light',
      ws: null,
      chatWindowKey: 0,
      reconnectTimer: null,
      userStatuses: {},
      mobileView: 'sidebar',
      isMobile: false,
      currentUserName: null,
      isTyping: false,
      typingTimer: null,

    }
  },

  computed: {
    activeChat() {
      return this.directs.find(d => String(d.id) === String(this.activeChatId)) || null
    },
    activePresence() {
      const key = String(this.activeRecipientId || '')
      return this.userStatuses[key] || { online: false, lastSeen: null }
    }
  },
  watch: {
  isLight(val) {
    this.$nextTick(() => {
      this.updateThemeColor()
    })
  }
},
  async mounted() {
  this.currentUserId = this.parseUserIdFromToken()
  
  const token = sessionStorage.getItem('access_token')
  if (token) {
    const payload = this.parseJwt(token)
    this.currentUserName = sessionStorage.getItem('current_user_name') || payload?.name || payload?.nickname || payload?.username || ''
  }
  
  if ('Notification' in window && Notification.permission === 'default') {
    const { subscribe } = await import('../composables/usePush.js')
    try { await subscribe() } catch (e) { console.warn('Push subscribe error:', e) }
  }
  
  this.updateThemeColor()
  await this.loadDirects()
  this.connectWebSocket()
  this.checkMobile()
  window.addEventListener('resize', this.checkMobile)
},

  beforeUnmount() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    window.removeEventListener('resize', this.checkMobile)
  },

  methods: {
    
    connectWebSocket() {
      if (!this.currentUserId) return
      if (this.reconnectTimer) { clearTimeout(this.reconnectTimer); this.reconnectTimer = null }
      if (this.ws) { this.ws.close(); this.ws = null }

      const wsUrl = `${WS_BASE}/ws?user_id=${this.currentUserId}`
      this.ws = new WebSocket(wsUrl)

      
      
      this.ws.onmessage = (event) => {
        if (import.meta.env.DEV) console.debug('WS RAW:', event.data)
      try {
        const data = JSON.parse(event.data)
        let payload = data.Payload ?? data.payload ?? null
        let type = data.Type ?? data.type ?? null
        if (import.meta.env.DEV) console.debug('WS TYPE:', type, 'PAYLOAD:', payload)

          if (!payload && (data.chat_id || data.chatId || data.content)) {
            payload = data
            type = 'SendMessage'
          }

          if (typeof payload === 'string') {
            try { payload = JSON.parse(payload) } catch { payload = null }
          }

                if (type === 'SendMessage' || type === 'send_message' || type === 'new_message') {
        const chatId =
          payload?.chat_id ??
          payload?.chatId ??
          payload?.chatid ??
          payload?.ChatID

        const content =
          payload?.content ??
          payload?.Content ??
          ''

        const createdAt =
          payload?.created_at ??
          payload?.createdAt ??
          payload?.createdat ??
          payload?.CreatedAT ??
          new Date().toISOString()
          const senderId = payload?.sender_id ?? payload?.senderId ?? payload?.senderid

        if (chatId) {
          this.updateChatPreview({
  chatId,
  content,
  date: createdAt,
  senderId,
  isIncoming: String(senderId) !== String(this.currentUserId)
})

          if (String(this.activeChatId) === String(chatId)) {
  const isViewingChat = !this.isMobile || this.mobileView === 'chat'
  if (isViewingChat) {
    this.$refs.chatWindow?.handleIncomingMessage?.(payload)
  } else {
    this.$refs.chatWindow?.handleIncomingMessage?.(payload)
  }
}
        }

        return
      }

          if (type === 'user_online') {
            const userId = payload?.user_id ?? payload?.userId ?? payload?.id
            if (userId) this.setUserPresence(userId, { online: true, lastSeen: null })
            return
          }

          if (type === 'user_offline') {
            const userId = payload?.user_id ?? payload?.userId ?? payload?.id
            const lastSeen = payload?.last_seen ?? payload?.lastSeen ?? new Date().toISOString()
            if (userId) this.setUserPresence(userId, { online: false, lastSeen })
            return
          }

          if (type === 'new_chat') {
            
            this.loadDirects()
            return
          }

          if (type === 'delete_message') {
            const msgId = payload?.id ?? payload?.message_id
            if (msgId) {
              this.$refs.chatWindow?.handleDeleteMessage?.(payload)
            }
            return
          }

          if (type === 'delete_chat') {
            const chatId = payload?.chat_id
            if (chatId) this.onChatDeleted(chatId)
            return
          }

          if (type === 'mark_as_read') {
  const chatId = payload?.chat_id
  if (chatId) {
    const chat = this.directs.find(c => String(c.id) === String(chatId))
    if (chat) chat.last_message_status = 'read'  
    if (String(this.activeChatId) === String(chatId)) {
      this.$refs.chatWindow?.handleMessagesRead?.()
    }
  }
  return
}

if (type === 'typing') {  
   console.log('TYPING received', payload)
   console.log('activeChatId:', this.activeChatId, 'payload chatId:', payload?.chat_id)
  const chatId = payload?.chat_id
  if (chatId && String(chatId) === String(this.activeChatId)) {
    this.isTyping = true
    clearTimeout(this.typingTimer)
    this.typingTimer = setTimeout(() => {
      this.isTyping = false
    }, 3000)
  }
  return
}

          } catch (e) { 
            console.error('WS Parse Error:', e)
          }
      }

      this.ws.onclose = () => {
        this.ws = null
        this.reconnectTimer = setTimeout(() => this.connectWebSocket(), 5000)
      }
    },

    toggleTheme() {
    this.isLight = !this.isLight
    localStorage.setItem('svlynx-theme', this.isLight ? 'light' : 'dark')
    this.updateThemeColor()
    this.$emit('theme-changed')
  },

    updateThemeColor() {
      const color = this.isLight ? '#ffffff' : 'rgb(8, 12, 26)'
      document.body.style.background = color
      document.body.style.backgroundColor = color
      document.documentElement.style.background = color
      document.documentElement.style.backgroundColor = color
      const meta = document.querySelector('meta[name="theme-color"]')
      if (meta) meta.setAttribute('content', color)
    },

    parseJwt(token) {
      try {
        if (!token) return null
        const parts = token.split('.')
        if (parts.length !== 3) return null
        const base64 = parts[1].replace(/-/g, '+').replace(/_/g, '/')
        const padded = base64.padEnd(base64.length + (4 - (base64.length % 4)) % 4, '=')
        return JSON.parse(atob(padded))
      } catch { return null }
    },

    parseUserIdFromToken() {
      try {
        const token = getCookie('access_token')
        if (!token) return null
        const payload = this.parseJwt(token)
        if (!payload) return null
        return String(payload?.user_id ?? payload?.sub ?? payload?.id ?? '').trim()
      } catch { return null }
    },

    async loadDirects() {
      this.loading = true
      try {
        const userId = String(this.currentUserId || '').trim()
        if (!userId) return

        const url = new URL(`${BASE}/chat/direct/list`)
        url.searchParams.set('user_id', userId)

        const res = await apiFetch(url.toString())

        if (!res.ok) return
        const data = await res.json()
        const apiChats = data.directs || data.chats || []

        const chatMap = new Map()
        this.directs.forEach(c => chatMap.set(c.id, c))
        apiChats.forEach(c => chatMap.set(c.id, c))

        this.directs = Array.from(chatMap.values()).sort((a, b) => {
          const dateA = new Date(a.last_message_at || a.updated_at || a.created_at || 0)
          const dateB = new Date(b.last_message_at || b.updated_at || b.created_at || 0)
          return dateB - dateA
        })


        this.saveChatsToLocal()
        this.fetchAllStatuses()

      } catch (e) {
        console.error('loadDirects crash:', e)
      } finally {
        this.loading = false
      }
    },

    saveChatsToLocal() {
      try { localStorage.setItem('svlynx-saved-chats', JSON.stringify(this.directs)) } catch {}
    },

    selectChat({ chatId, recipientId }) {
  this.activeChatId = chatId
  this.activeRecipientId = recipientId
  this.chatWindowKey++

  const idx = this.directs.findIndex(c => String(c.id) === String(chatId))
  if (idx !== -1) {
    this.directs[idx] = { ...this.directs[idx], unread_count: 0, unreadcount: 0 }
    this.directs = [...this.directs]
  }

  if (recipientId) this.fetchUserStatus(recipientId)
  if (this.isMobile) this.mobileView = 'chat'
  this.saveChatsToLocal()
},

    onMessageDeleted({ id, chat_id, recipient_id }) {
      console.log('onMessageDeleted called', { id, chat_id, recipient_id }) // ← добавь
      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
        console.log('WS not open:', this.ws?.readyState) // ← и это
        return
      }
      const msg = JSON.stringify({
        type: 'delete_message',
        payload: { id, chat_id, recipient_id }
      })
      this.ws.send(msg)
    },

    goBackToSidebar() {
  this.mobileView = 'sidebar'
  this.activeChatId = null
  this.activeRecipientId = null
},

    checkMobile() {
      this.isMobile = window.innerWidth <= 760
      if (!this.isMobile) this.mobileView = 'sidebar'
    },

    setUserPresence(userId, patch = {}) {
      const key = String(userId)
      const prev = this.userStatuses[key] || { online: false, lastSeen: null }
      this.userStatuses = { ...this.userStatuses, [key]: { ...prev, ...patch } }
    },

    async fetchUserStatus(userId) {
      try {
        const res = await apiFetch(`${BASE}/users/${userId}/status`)
        if (!res.ok) return
        const data = await res.json()
        this.setUserPresence(userId, {
          online: data?.online === true,
          lastSeen: data?.last_seen ?? data?.lastseen ?? null
        })
      } catch (e) {
        console.error('fetchUserStatus error', e)
      }
    },

    async fetchAllStatuses() {
  const ids = this.directs.map(d => {
    const first = d.first_user_id ?? d.firstuserid
    const second = d.second_user_id ?? d.seconduserid
    return String(first) === String(this.currentUserId) ? second : first
  }).filter(Boolean)

  await Promise.all(ids.map(id => this.fetchUserStatus(id)))
},
    onChatDeleted(chatId) {
      this.directs = this.directs.filter(d => String(d.id) !== String(chatId))
      if (String(this.activeChatId) === String(chatId)) {
        this.activeChatId = null
        this.activeRecipientId = null
      }
      this.saveChatsToLocal()
    },

    onMarkAsRead({ chat_id, user_id, recipient_id }) {
      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return
      this.ws.send(JSON.stringify({
        type: 'mark_as_read',
        payload: { chat_id, user_id, recipient_id }
      }))
    },

    onTyping({ chat_id, sender_id, recipient_id }) {  // ← сюда
  if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return
  this.ws.send(JSON.stringify({
    type: 'typing',
    payload: { chat_id, sender_id, recipient_id }
  }))
},

    updateChatPreview({ chatId, content, date, isIncoming, senderId }) {
  const idx = this.directs.findIndex(c => String(c.id) === String(chatId))
  if (idx === -1) return

  const chat = { ...this.directs[idx] }
  
  chat.last_message_content = content
  chat.last_message_at = date
  chat.last_message_sender_id = isIncoming ? senderId : this.currentUserId
  chat.last_message_status = isIncoming ? 'delivered' : 'sent'

  if (isIncoming && String(this.activeChatId) !== String(chatId)) {
    chat.unread_count = (Number(chat.unread_count || chat.unreadcount) || 0) + 1
  }

  this.directs[idx] = chat
  this.directs = [...this.directs].sort((a, b) => {
    const dateA = new Date(a.last_message_at || a.updated_at || a.created_at || 0)
    const dateB = new Date(b.last_message_at || b.updated_at || b.created_at || 0)
    return dateB - dateA
  })
  this.saveChatsToLocal()
},

    async startChat(userId, nickname) {
      try {
        const res = await apiFetch(`${BASE}/chat/direct`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ first_user_id: this.currentUserId, second_user_id: userId })
        })

        const text = await res.text()
        let data
        try { data = JSON.parse(text) } catch { return }
        if (!res.ok) return

        const direct = data.direct || data.chat || data

        await this.loadDirects()

        this.activeChatId = direct.id
        this.activeRecipientId = userId
        this.chatWindowKey++
        this.fetchUserStatus(userId)

        if (this.isMobile) this.mobileView = 'chat'
      } catch (e) {
        console.error('startChat crash:', e)
      }
    }
  }
}
</script>


<style scoped>
@import url('https://api.fontshare.com/v2/css?f[]=satoshi@400,500,700&display=swap');


.direct-page {
  height: 100svh;
  width: 100vw;
  min-height: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  box-sizing: border-box;
  position: relative;
  overflow: hidden;
  background: rgba(8, 12, 26, 0.98);
  font-family: 'Satoshi', sans-serif;
  transition: background 0.3s;
} 


/* Light page background */
.direct-page.theme-light {
  background:
    radial-gradient(circle at 12% 20%, rgba(91, 106, 255, 0.06), transparent 22%),
    radial-gradient(circle at 85% 88%, rgba(108, 86, 255, 0.06), transparent 26%),
    linear-gradient(180deg, #eef0fb 0%, #e8ebf8 100%);
}


.direct-page::before {
  content: '';
  position: absolute; inset: 0;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.025) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.025) 1px, transparent 1px);
  background-size: 46px 46px;
  opacity: 0.45;
  pointer-events: none;
  transition: opacity 0.3s;
}
.direct-page.theme-light::before {
  background-image:
    linear-gradient(rgba(91, 106, 255, 0.06) 1px, transparent 1px),
    linear-gradient(90deg, rgba(91, 106, 255, 0.06) 1px, transparent 1px);
  opacity: 1;
}


.direct-shell {
  position: relative; z-index: 1;
  width: 100%; max-width: 1600px; height: 100%; min-height: 0;
  display: grid; grid-template-columns: 340px 1fr;
  border-radius: 26px; overflow: hidden;
  border: 1px solid rgba(132, 144, 224, 0.14);
  background: linear-gradient(180deg, rgba(9, 13, 28, 0.94), rgba(7, 10, 22, 0.96));
  backdrop-filter: blur(18px);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.34), inset 0 1px 0 rgba(255, 255, 255, 0.04);
  transition: background 0.3s, border-color 0.3s;
}


.direct-shell.theme-light {
  background: #ffffff;
  border-color: rgba(91, 106, 255, 0.15);
  box-shadow: 0 20px 60px rgba(91, 106, 200, 0.12), inset 0 1px 0 rgba(255, 255, 255, 0.8);
}


.content-area {
  min-width: 0; height: 100%; min-height: 0;
  display: flex; flex-direction: column; overflow: hidden;
  border-left: 1px solid rgba(255, 255, 255, 0.06);
  background: linear-gradient(180deg, rgba(10, 14, 30, 0.78), rgba(7, 10, 22, 0.84));
  transition: background 0.3s, border-color 0.3s;
}
.direct-shell.theme-light .content-area {
  background: #f5f6fc;
  border-left-color: #e4e6f0;
}


.empty-chat {
  flex: 1; min-height: 0;
  display: flex; align-items: center; justify-content: center; padding: 24px;
}


.empty-card {
  width: 100%; max-width: 480px; padding: 34px 26px;
  border-radius: 22px; text-align: center;
  background: rgba(255, 255, 255, 0.025);
  border: 1px solid rgba(255, 255, 255, 0.05);
}
.direct-shell.theme-light .empty-card {
  background: #ffffff;
  border-color: #e4e6f0;
  box-shadow: 0 8px 30px rgba(91, 106, 200, 0.08);
}


.empty-logo {
  width: 68px; height: 68px; margin: 0 auto 18px;
  border-radius: 18px; display: grid; place-items: center; color: #e9edff;
  background: linear-gradient(135deg, rgba(96, 111, 255, 0.28), rgba(117, 93, 255, 0.18));
  border: 1px solid rgba(130, 141, 255, 0.22);
}
.empty-title { color: #eef2ff; font-size: 26px; font-weight: 700; margin-bottom: 10px; }
.empty-text { color: #8d96ba; font-size: 14px; line-height: 1.7; max-width: 34ch; margin: 0 auto; }
.direct-shell.theme-light .empty-title { color: #1a1d2e; }
.direct-shell.theme-light .empty-text { color: #7880a0; }

.chat-avatar-wrap {
  position: relative;
  flex-shrink: 0;
  width: 54px;
  height: 54px;
  overflow: visible;
}

.online-dot {
  position: absolute;
  right: -2px;
  bottom: -2px;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: #22c55e;
  border: 2px solid rgba(8, 12, 26, 0.98);
  z-index: 2;
}
@media (max-width: 980px) {
  .direct-shell { grid-template-columns: 300px 1fr; }
}
@media (max-width: 760px) {
  .direct-page {
    padding: 0;
    align-items: stretch;
    position: fixed;
    inset: 0;
    background: rgb(8, 12, 26);
  }
  .direct-page.theme-light {
    background: #ffffff;
  }

  .direct-shell {
    flex: 1;
    height: 100%;
    min-height: 0;
    border-radius: 0;
    border: none;
    padding-top: env(safe-area-inset-top);
    grid-template-columns: 1fr;
    grid-template-rows: 1fr;
    align-items: stretch;
  }

  .direct-shell.theme-light {
    background: #ffffff;
  }

  .content-area {
    grid-column: 1;
    grid-row: 1;
    height: 100%;
    overflow: hidden;
  }

  .direct-shell.theme-light .content-area {
    background: #ffffff;
  }

  .mobile-hidden {
    display: none !important;
  }

}
</style>