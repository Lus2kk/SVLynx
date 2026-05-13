<template>
  <section class="chat-window" :class="{ 'theme-light': isLight }">
    <header class="chat-header">
      <div class="chat-user">
        <button v-if="showBackButton" class="back-btn" type="button" title="Back" @click="$emit('back')">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.2">
            <path d="M15 18l-6-6 6-6"></path>
          </svg>
        </button>
        <div class="chat-avatar" :style="!avatarUrl ? { background: avatarColor } : {}">
          <img v-if="avatarUrl" :src="avatarUrl" alt="" class="chat-avatar-image" />
          <span v-else>{{ avatarLetter }}</span>
        </div>
        <div class="chat-user-meta">
          <div class="chat-username">{{ chatTitle }}</div>
          <div class="chat-status" :class="{ online: !isTyping && presence.online, offline: !isTyping && !presence.online, typing: isTyping }">
  <span class="status-dot"></span>
  <span>{{ isTyping ? 'печатает...' : presenceText }}</span>
</div>
        </div>
      </div>
      <div class="chat-actions">
         <button class="chat-icon-btn" :class="{ active: searchOpen }" title="Search" type="button" @click="openSearch">
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
      <transition name="search-slide">
        <div v-if="searchOpen" class="search-bar">
  <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="1.8" class="search-icon">
    <circle cx="11" cy="11" r="7"></circle>
    <path d="M20 20l-3.5-3.5"></path>
  </svg>
  <input
    ref="searchInput"
    v-model="searchQuery"
    type="text"
    class="search-input"
    placeholder="Поиск сообщений..."
    @input="onSearchInput"
    @keydown.escape="closeSearch"
  />
          <span v-if="searchResults.length" class="search-count">
            {{ searchIndex + 1 }} / {{ searchResults.length }}
          </span>
          <button v-if="searchResults.length" class="search-nav-btn" @click="searchPrev" type="button">
            <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M15 18l-6-6 6-6"/>
            </svg>
          </button>
          <button v-if="searchResults.length" class="search-nav-btn" @click="searchNext" type="button">
            <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M9 18l6-6-6-6"/>
            </svg>
          </button>
          <button class="search-close-btn" @click="closeSearch" type="button">
            <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 6L6 18M6 6l12 12"/>
            </svg>
          </button>
        </div>
      </transition>

    </header>

    <div class="messages-area-wrapper">
      <div class="messages-area" ref="messagesArea" @scroll="onScroll">
  <template v-for="item in messagesWithSeparators" :key="item.id">
    <div v-if="item.type === 'separator'" class="day-separator">{{ item.label }}</div>
    <MessageBubble
      v-else
      :data-message-id="item.id"
      :message="item"
      :isMine="isMine(item)"
      :isLight="isLight"
      :highlight="searchResults.includes(item.id)"
      :highlightActive="searchResults[searchIndex] === item.id"
      @delete="confirmDelete"
      @reply="onReply"
      @select="onSelectMessage"
      :isSelecting="isSelecting"
      :isSelected="selectedMessages.some(m => String(m.id) === String(item.id))"
    />
  </template>
</div>
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
     <div v-if="isSelecting" class="selection-bar">
  <button class="sel-cancel" @click="cancelSelect">Отмена</button>
  <span class="sel-count">{{ selectedMessages.length }} выбрано</span>
  <button class="sel-delete" @click="deleteSelected">
    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
      <path d="M3 6h18"/>
      <path d="M8 6V4h8v2"/>
      <path d="M18 6l-1 14H7L6 6"/>
    </svg>
    Удалить
  </button>
</div>
    <div class="composer-wrap">
      <div v-if="replyTo" class="reply-preview">
  <div class="reply-preview-content">
    <span class="reply-preview-name">{{ String(replyTo.sender_id) === String(currentUserId) ? 'Вы' : 'Собеседник' }}</span>
    <span class="reply-preview-text">{{ replyTo.type === 'voice' ? '🎤 Голосовое' : replyTo.type === 'image' ? '📷 Фото' : replyTo.content }}</span>
  </div>
  <button class="reply-preview-close" @click="replyTo = null">✕</button>
</div>
      <form class="composer" @submit.prevent="sendMessage">
        <MediaUploader
  v-if="chatId && currentUserId && recipientId"
  :chatId="String(chatId)"
  :senderId="String(currentUserId)"
  :recipientId="String(recipientId)"
  :senderName="chatTitle"
  @media-sent="onMediaSent"
/>

        <textarea
  v-model="newMessage"
  class="message-input"
  placeholder="Сообщение..."
  ref="messageInput"
  rows="1"
  @input="onTypingAndResize"
  @keydown.enter.exact.prevent="sendMessage"
  @keydown.enter.shift.exact="newLine"
        />

        <button type="button" class="composer-side-btn" title="Emoji">
          <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="1.8">
            <circle cx="12" cy="12" r="9"></circle>
            <path d="M8 14s1.5 2 4 2 4-2 4-2"></path>
            <path d="M9 9h.01"></path>
            <path d="M15 9h.01"></path>
          </svg>
        </button>

        <!-- Кнопка отправки текста / переключения в voiceMode -->
        <button
  v-show="!voiceMode"
  type="button"
  class="send-btn"
  title="Send"
  aria-label="Отправить"
  @click="onSendClick"
>
  <!-- Diagonal arrow-up. Original geometry, no paper-plane. -->
  <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round">
    <path d="M7 17 L17 7"></path>
    <path d="M9 7 H17 V15"></path>
  </svg>
</button>

        <!-- Кнопка записи: короткий клик = выйти, долгое нажатие = запись -->
        <!-- Обработчики вешаются через ref в watch, а не через Vue директивы -->
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
          <span v-if="isRecordingVoice" class="rec-timer-inline">{{ voiceTimerText }}</span>
        </button>
      </form>
    </div>
  </section>
</template>

<script>
import MessageBubble from './MessageBubble.vue'
import { apiFetch, getCookie } from '../api.js'
import MediaUploader from './MediaUploader.vue'

const BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'
const VOICE_BASE = import.meta.env.VITE_VOICE_API_URL || 'http://localhost:9090'

export default {
  name: 'ChatWindow',
  components: { MessageBubble, MediaUploader },

  props: {
    chat: { type: Object, default: null },
    chatId: { type: [String, Number], default: null },
    isVisible: { type: Boolean, default: true },
    currentUserId: { type: [String, Number], default: null },
    recipientId: { type: [String, Number], default: null },
    selectedCompanion: { type: Object, default: null },
    isLight: { type: Boolean, default: false },
    showBackButton: { type: Boolean, default: false },
    presence: { type: Object, default: () => ({ online: false, lastSeen: null }) },
    isTyping: { type: Boolean, default: false }
  },

  emits: ['message-sent', 'message-deleted', 'mark-as-read', 'back', 'typing'],

  data() {
    return {
      searchOpen: false,
      searchQuery: '',
      searchResults: [],
      searchIndex: 0,
      searchDebounce: null,
      waveformData: [],
      audioContext: null,
      analyser: null,
      voiceMode: false,
      hasMore: true,
      loadingMore: false,
      messages: [],
      newMessage: '',
      deleteModalOpen: false,
      messageToDelete: null,
      loading: false,
      nowTick: Date.now(),
      presenceTimer: null,
      isRecordingVoice: false,
      voiceMediaRecorder: null,
      voiceChunks: [],
      voiceTimerSeconds: 0,
      voiceTimerInterval: null,
      replyTo: null,
      isSelecting: false,
      selectedMessages: [],
      typingTimer: null,
    }
  },

  async mounted() {
    this.presenceTimer = setInterval(() => {
      this.nowTick = Date.now()
    }, 6000)

    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
      stream.getTracks().forEach(t => t.stop())
    } catch (e) {
      console.warn('Microphone permission denied', e)
    }
  },

  beforeUnmount() {
    clearInterval(this.presenceTimer)
    this._removeBtnListeners()
  },

  computed: {
    messagesWithSeparators() {
  if (!this.messages.length) return [{ type: 'separator', id: 'sep-today', label: 'Сегодня' }]

  const result = []
  let lastDateKey = null

  for (const msg of this.messages) {
    const date = new Date(msg.created_at)
    const dateKey = date.toDateString()

    if (dateKey !== lastDateKey) {
      lastDateKey = dateKey
      result.push({ type: 'separator', id: `sep-${dateKey}`, label: this.formatDateLabel(date) })
    }
    result.push(msg)
  }
  return result
},
    avatarColor() {
      return this.chat?.companion_avatar_color || 'linear-gradient(135deg, #6d78ff, #8866ff)'
    },
    voiceTimerText() {
      const m = Math.floor(this.voiceTimerSeconds / 60).toString().padStart(2, '0')
      const s = (this.voiceTimerSeconds % 60).toString().padStart(2, '0')
      return `${m}:${s}`
    },
    chatTitle() {
      const name =
        this.selectedCompanion?.nickname ||
        this.selectedCompanion?.name ||
        this.selectedCompanion?.username ||
        this.chat?.companion_name ||
        this.chat?.companion_nickname ||
        this.chat?.nickname ||
        this.chat?.username
      return name || (this.recipientId ? `User ${String(this.recipientId).slice(0, 6)}` : 'Unknown')
    },
    avatarUrl() {
      return this.selectedCompanion?.photo_url || this.chat?.companion_photo_url || this.chat?.photo_url || null
    },
    avatarLetter() {
      return this.chatTitle?.[0]?.toUpperCase() || ''
    },
    presenceText() {
      void this.nowTick
      if (this.presence?.online) return 'В сети'
      if (this.presence?.lastSeen) return this.formatLastSeen(this.presence.lastSeen)
      return 'Не в сети'
    }
  },

  watch: {
    isTyping(val) {
    console.log('ChatWindow isTyping changed:', val)  
  },
    chatId: {
      immediate: true,
      async handler(value) {
        if (value) {
          await this.loadMessages()
        } else {
          this.messages = []
        }
      }
    },

    voiceMode(val) {
      this._removeBtnListeners()
      if (val) {
        this.$nextTick(() => this._addBtnListeners())
      }
    }
  },

  methods: {
    onTyping() {
  clearTimeout(this.typingTimer)
  this.$emit('typing', {
    chat_id: this.chatId,
    sender_id: this.currentUserId,
    recipient_id: this.recipientId
  })
  this.typingTimer = setTimeout(() => {}, 2000)
},

onTypingAndResize() {
  this.onTyping()
  const el = this.$refs.messageInput
  if (!el) return
  el.style.height = 'auto'
  el.style.height = Math.min(el.scrollHeight, 120) + 'px'
},

newLine() {
  this.newMessage += '\n'
  this.$nextTick(() => this.onTypingAndResize())
},
    onSelectMessage(message) {
  if (!this.isSelecting) this.isSelecting = true
  const idx = this.selectedMessages.findIndex(m => String(m.id) === String(message.id))
  if (idx === -1) {
    this.selectedMessages.push(message)
  } else {
    this.selectedMessages.splice(idx, 1)
    if (this.selectedMessages.length === 0) this.isSelecting = false
  }
},
formatDateLabel(date) {
  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const yesterday = new Date(today - 86400000)
  const msgDay = new Date(date.getFullYear(), date.getMonth(), date.getDate())

  if (msgDay.getTime() === today.getTime()) return 'Сегодня'
  if (msgDay.getTime() === yesterday.getTime()) return 'Вчера'
  return date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' })
},
cancelSelect() {
  this.isSelecting = false
  this.selectedMessages = []
},
async deleteSelected() {
  const mine = this.selectedMessages.filter(m => String(m.sender_id) === String(this.currentUserId))
  for (const msg of mine) {
    try {
      await apiFetch(`${BASE}/chat/messages/${msg.id}`, { method: 'DELETE' })
      this.messages = this.messages.filter(m => String(m.id) !== String(msg.id))
    } catch (e) { console.error(e) }
  }
  this.cancelSelect()
},
    onMediaSent(message) {
  if (!message) return
  const msg = this.normalizeMessage(message)
  if (this.messages.find(m => String(m.id) === String(msg.id))) return
  this.messages.push(msg)
  this.scrollToBottom()
  this.$emit('message-sent', {
    chatId: this.chatId,
    content: '📎 Медиафайл',
    date: msg.created_at
  })
},
  onReply(message) {
  this.replyTo = message
  if (!('ontouchstart' in window)) {
    this.$refs.messageInput?.focus()
    if (this.$refs.messageInput) {
  this.$refs.messageInput.style.height = 'auto'
}
  }
},

  openSearch() {
    this.searchOpen = true
    this.$nextTick(() => this.$refs.searchInput?.focus())
  },

  closeSearch() {
    this.searchOpen = false
    this.searchQuery = ''
    this.searchResults = []
    this.searchIndex = 0
  },

  onSearchInput() {
    clearTimeout(this.searchDebounce)
    if (!this.searchQuery.trim()) {
      this.searchResults = []
      this.searchIndex = 0
      return
    }
    this.searchDebounce = setTimeout(() => this.doSearch(), 400)
  },

    async doSearch() {
      if (!this.chatId || !this.searchQuery.trim()) return
      try {
        const url = new URL(`${BASE}/chat/messages/search`)
        url.searchParams.set('chat_id', this.chatId)
        url.searchParams.set('content', this.searchQuery.trim())
        const res = await apiFetch(url.toString())
        if (!res.ok) return
        const data = await res.json()
        const found = Array.isArray(data.messages) ? data.messages : []
        this.searchResults = found.map(m => m.id)
        this.searchIndex = 0
        if (this.searchResults.length) this.scrollToMessage(this.searchResults[0])
      } catch (e) {
        console.error('Search error', e)
      }
    },

    searchNext() {
      if (!this.searchResults.length) return
      this.searchIndex = (this.searchIndex + 1) % this.searchResults.length
      this.scrollToMessage(this.searchResults[this.searchIndex])
    },

    searchPrev() {
      if (!this.searchResults.length) return
      this.searchIndex = (this.searchIndex - 1 + this.searchResults.length) % this.searchResults.length
      this.scrollToMessage(this.searchResults[this.searchIndex])
    },

    scrollToMessage(id) {
      this.$nextTick(() => {
        const el = document.querySelector(`[data-message-id="${id}"]`)
        if (el) el.scrollIntoView({ behavior: 'smooth', block: 'center' })
      })
    },

    _addBtnListeners() {
      const btn = this.$refs.voiceBtn
      if (!btn) return

      this._onPressDown = () => {
        this._isLongPress = false
        clearTimeout(this._pressTimer)
        this._pressTimer = setTimeout(() => {
          this._isLongPress = true
          this.startVoice()
        }, 300)
      }

      this._onPressUp = () => {
      clearTimeout(this._pressTimer)
      if (this.isRecordingVoice) {
        // запись идёт — останавливаем в любом случае
        this.stopVoice()
      } else if (!this._isLongPress) {
        // короткий клик без записи — выйти из voiceMode
        this.voiceMode = false
      }
      this._isLongPress = false
    }

      this._onTouchStartNative = (e) => {
  e.preventDefault()
  e.stopPropagation()
  this._touchMoved = false
  this._isLongPress = false
  clearTimeout(this._pressTimer)
  this._pressTimer = setTimeout(() => {
    this._isLongPress = true
    this.startVoice()
  }, 300)
}

this._onTouchMoveNative = (e) => {
  this._touchMoved = true
  clearTimeout(this._pressTimer)
}

this._onTouchEndNative = (e) => {
  e.preventDefault()
  e.stopPropagation()
  clearTimeout(this._pressTimer)
  if (this.isRecordingVoice) {
    this.stopVoice()
  } else if (!this._isLongPress && !this._touchMoved) {
    this.voiceMode = false
  }
  this._isLongPress = false
}

      this._onMouseLeaveNative = () => {
        clearTimeout(this._pressTimer)
        if (this.isRecordingVoice) this.stopVoice()
        this._isLongPress = false
      }

      btn.addEventListener('mousedown', this._onPressDown)
      btn.addEventListener('mouseup', this._onPressUp)
      btn.addEventListener('mouseleave', this._onMouseLeaveNative)
      btn.addEventListener('touchstart', this._onTouchStartNative, { passive: false })
      btn.addEventListener('touchmove', this._onTouchMoveNative, { passive: false })
      btn.addEventListener('touchend', this._onTouchEndNative, { passive: false })
      btn.addEventListener('touchcancel', this._onTouchEndNative, { passive: false })
    },

    _removeBtnListeners() {
  const btn = this.$refs.voiceBtn
  if (!btn) return
  if (this._onPressDown) btn.removeEventListener('mousedown', this._onPressDown)
  if (this._onPressUp) btn.removeEventListener('mouseup', this._onPressUp)
  if (this._onMouseLeaveNative) btn.removeEventListener('mouseleave', this._onMouseLeaveNative)
  if (this._onTouchStartNative) btn.removeEventListener('touchstart', this._onTouchStartNative)
  if (this._onTouchMoveNative) btn.removeEventListener('touchmove', this._onTouchMoveNative)
  if (this._onTouchEndNative) {
    btn.removeEventListener('touchend', this._onTouchEndNative)
    btn.removeEventListener('touchcancel', this._onTouchEndNative)
  }
},

    onSendClick() {
      if (this.newMessage.trim()) {
        this.sendMessage()
      } else {
        this.voiceMode = !this.voiceMode
      }
    },

    async startVoice() {
      if (this.isRecordingVoice) return
      try {
        const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
        this.voiceChunks = []
        this.waveformData = []

        this.audioContext = new AudioContext()
        this.analyser = this.audioContext.createAnalyser()
        this.analyser.fftSize = 256
        const source = this.audioContext.createMediaStreamSource(stream)
        source.connect(this.analyser)

        const mimeType = MediaRecorder.isTypeSupported('audio/webm;codecs=opus')
          ? 'audio/webm;codecs=opus'
          : MediaRecorder.isTypeSupported('audio/mp4')
            ? 'audio/mp4'
            : 'audio/webm'

        this.voiceMediaRecorder = new MediaRecorder(stream, { mimeType })
        this.voiceMediaRecorder.ondataavailable = e => {
          if (e.data.size > 0) this.voiceChunks.push(e.data)
        }
        this.voiceMediaRecorder.onstop = this.handleVoiceStop
        this.voiceMediaRecorder.start()
        this.isRecordingVoice = true
        this.voiceTimerSeconds = 0
        this.voiceTimerInterval = setInterval(() => this.voiceTimerSeconds++, 1000)

        this.waveformInterval = setInterval(() => {
          const dataArray = new Uint8Array(this.analyser.frequencyBinCount)
          this.analyser.getByteFrequencyData(dataArray)
          const avg = dataArray.reduce((a, b) => a + b, 0) / dataArray.length
          this.waveformData.push(avg / 255)
        }, 100)

      } catch (e) {
        console.error('Microphone error', e)
      }
    },

    stopVoice() {
      if (!this.voiceMediaRecorder || !this.isRecordingVoice) return
      this.isRecordingVoice = false
      clearInterval(this.voiceTimerInterval)
      clearInterval(this.waveformInterval)
      if (this.audioContext) {
        this.audioContext.close()
        this.audioContext = null
      }
      this.voiceMediaRecorder.requestData()
      this.voiceMediaRecorder.stream.getTracks().forEach(t => t.stop())
      this.voiceMediaRecorder.stop()
    },

    async handleVoiceStop() {
      if (this.voiceChunks.length === 0) return
      const mimeType = MediaRecorder.isTypeSupported('audio/webm;codecs=opus')
        ? 'audio/webm;codecs=opus'
        : MediaRecorder.isTypeSupported('audio/mp4')
          ? 'audio/mp4'
          : 'audio/webm'

      const blob = new Blob(this.voiceChunks, { type: mimeType })
      const ext = mimeType.includes('mp4') ? '.mp4' : '.webm'

      const form = new FormData()
      form.append('file', blob, `voice_${Date.now()}${ext}`)
      form.append('chat_id', String(this.chatId))
      form.append('sender_id', String(this.currentUserId))
      form.append('recipient_id', String(this.recipientId))
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
        this.onVoiceSent(data.message)
      } catch (e) {
        console.error('Voice upload error', e)
      } finally {
        this.voiceMode = false
      }
    },

    onVoiceSent(message) {
      if (!message) return
      const msg = this.normalizeMessage({
        ...message,
        waveform: this.waveformData,
        duration: this.voiceTimerSeconds
      })
      if (this.messages.find(m => String(m.id) === String(msg.id))) return
      this.messages.push(msg)
      this.scrollToBottom()
      this.$emit('message-sent', {
        chatId: this.chatId,
        content: 'Голосовое сообщение',
        date: msg.created_at
      })
    },

    onScroll() {
      const area = this.$refs.messagesArea
      if (!area) return
      if (area.scrollTop < 100 && this.hasMore && !this.loadingMore) {
        this.loadMoreMessages()
      }
    },

    async loadMoreMessages() {
      if (!this.chatId || this.loadingMore || !this.hasMore) return
      this.loadingMore = true
      try {
        const oldest = this.messages[0]?.created_at
        if (!oldest) return
        const url = new URL(`${BASE}/chat/messages`)
        url.searchParams.set('chat_id', this.chatId)
        url.searchParams.set('before', oldest)
        url.searchParams.set('limit', '50')
        const res = await apiFetch(url.toString())
        if (!res.ok) return
        const data = await res.json()
        const older = Array.isArray(data.messages) ? data.messages : []
        if (older.length === 0) {
          this.hasMore = false
          return
        }
        const area = this.$refs.messagesArea
        const prevScrollHeight = area.scrollHeight
        this.messages = [
          ...older.map(this.normalizeMessage).sort((a, b) => new Date(a.created_at) - new Date(b.created_at)),
          ...this.messages
        ]
        this.$nextTick(() => {
          area.scrollTop = area.scrollHeight - prevScrollHeight
        })
      } catch (e) {
        console.error('loadMoreMessages error', e)
      } finally {
        this.loadingMore = false
      }
    },

    formatLastSeen(raw) {
      if (!raw) return 'не в сети'
      const normalized = raw.endsWith('Z') || raw.includes('+') ? raw : raw + 'Z'
      const date = new Date(normalized)
      if (isNaN(date.getTime())) return 'в сети'
      const diff = Math.floor((Date.now() - date.getTime()) / 1000)
      if (diff < 60) return 'был(a) только что'
if (diff < 3600) return `был(a) ${Math.floor(diff / 60)} мин назад`
if (diff < 86400) return `был(a) сегодня в ${date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}`
if (diff < 172800) return `был(a) вчера в ${date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}`
return `был ${date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })}`
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
      this.hasMore = true
      if (!this.chatId) return
      this.loading = true
      try {
        const url = new URL(`${BASE}/chat/messages`)
        url.searchParams.set('chat_id', this.chatId)
        const res = await apiFetch(url.toString())
        if (!res.ok) return
        const data = await res.json()
        const apiMessages = Array.isArray(data.messages) ? data.messages : []
        this.messages = apiMessages
          .map(this.normalizeMessage)
          .sort((a, b) => new Date(a.created_at) - new Date(b.created_at))
        if (!skipReadPatch && !document.hidden) await this.markIncomingAsRead()
        this.scrollToBottom()
      } catch (e) {
        console.error('Failed to load messages', e)
      } finally {
        this.loading = false
      }
    },

    handleIncomingMessage(rawPayload) {
  const msg = this.normalizeMessage(rawPayload)
  if (this.messages.find(m => String(m.id) === String(msg.id))) return
  this.messages.push(msg)
  this.scrollToBottom()
  if (!document.hidden) this.markIncomingAsRead()
},

    handleDeleteMessage(payload) {
      const id = payload?.id ?? payload?.message_id
      if (!id) return
      this.messages = this.messages.filter(m => String(m.id) !== String(id))
    },

    handleMessagesRead() {
      this.messages = this.messages.map(m =>
        String(m.sender_id) === String(this.currentUserId) ? { ...m, status: 'read' } : m
      )
    },

    async markIncomingAsRead() {
      if (!this.chatId || !this.currentUserId) return
      if (!this.isVisible || document.hidden) return 
      const hasUnread = this.messages.some(
        m => String(m.sender_id) !== String(this.currentUserId) && m.status !== 'read'
      )
      if (!hasUnread) return
      this.$emit('mark-as-read', {
        chat_id: this.chatId,
        user_id: this.currentUserId,
        recipient_id: this.recipientId
      })
      this.messages = this.messages.map(m =>
        String(m.sender_id) !== String(this.currentUserId) ? { ...m, status: 'read' } : m
      )
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

      this.$refs.messageInput?.focus()

      const optimistic = {
        id: `local-${Date.now()}`,
        sender_id: this.currentUserId,
        content: text,
        created_at: new Date().toISOString(),
        status: 'sent'
      }

      const pendingReplyTo = this.replyTo ? {
    id: String(this.replyTo.id),
    content: this.replyTo.content || '',
    type: this.replyTo.type || 'text',
    file_name: this.replyTo.file_name || '',
    is_mine: String(this.replyTo.sender_id) === String(this.currentUserId)
} : null
this.messages.push(optimistic)
this.newMessage = ''
this.replyTo = null
this.scrollToBottom()

      this.$emit('message-sent', {
        chatId: this.chatId,
        content: text,
        date: optimistic.created_at
      })

      try {
        const res = await apiFetch(`${BASE}/chat/messages`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
    chat_id: this.chatId,
    sender_id: this.currentUserId,
    recipient_id: this.recipientId,
    content: text,
    reply_to: pendingReplyTo,
    sender_name: this.chatTitle
})
        })

        if (!res.ok) throw new Error('Network response was not ok')

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
      } catch (e) {
        console.error('Failed to send message', e)
        this.messages = this.messages.map(m =>
          m.id === optimistic.id ? { ...m, status: 'failed' } : m
        )
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
        const res = await apiFetch(`${BASE}/chat/messages/${messageId}`, {
          method: 'DELETE'
        })
        if (!res.ok) return
        this.messages = this.messages.filter(m => String(m.id) !== String(messageId))
        this.$emit('message-deleted', { id: messageId, chat_id: this.chatId, recipient_id: this.recipientId })
      } catch (e) {
        console.error('Delete message error', e)
      }
    }
  }
}
</script>

<style scoped>
.chat-window {
  height: 100%;
  display: flex; flex-direction: column; min-width: 0; overflow: hidden;
  background: linear-gradient(180deg, rgba(10, 13, 28, 0.72), rgba(7, 10, 22, 0.82));
  transition: background 0.3s;
}
.chat-window.theme-light { background: #f5f6fc; }

.chat-header {
  height: 78px; min-height: 78px; flex-shrink: 0;
  position: sticky;
  padding: 14px 20px;
  display: flex; align-items: center; justify-content: space-between;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  background: rgba(255, 255, 255, 0.015);
  transition: all 0.3s;
}
.theme-light .chat-header { background: #ffffff; border-bottom-color: #e4e6f0; box-shadow: 0 1px 0 #e4e6f0; }

.chat-user { display: flex; align-items: center; gap: 12px; min-width: 0; }
.chat-avatar {
  width: 38px; height: 38px; border-radius: 50%;
  display: grid; place-items: center; overflow: hidden; flex-shrink: 0;
  color: white; font-weight: 700; font-size: 13px;
}
.chat-avatar-image { width: 100%; height: 100%; object-fit: cover; }
.chat-user-meta { min-width: 0; }
.chat-username { color: #f2f4ff; font-size: 14px; font-weight: 700; line-height: 1.2; }
.theme-light .chat-username { color: #1a1d2e; }

.chat-status {
  margin-top: 3px; font-size: 11px; font-weight: 600;
  display: flex; align-items: center; gap: 5px;
  color: #8691b7; transition: color 0.3s;
}
.theme-light .chat-status { color: #9098b8; }
.status-dot { width: 7px; height: 7px; border-radius: 50%; background: #4d5270; flex-shrink: 0; transition: background 0.3s; }
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

.messages-area-wrapper {
  flex: 1; min-height: 0; overflow: hidden;
  position: relative; display: flex; flex-direction: column;
  background-color: transparent;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.04) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.02) 1px, transparent 1px);
  background-size: 48px 48px;
}
.theme-light .messages-area-wrapper {
  background-image:
    linear-gradient(rgba(91, 106, 255, 0.06) 1px, transparent 1px),
    linear-gradient(90deg, rgba(91, 106, 255, 0.04) 1px, transparent 1px);
}

.messages-area {
  flex: 1; min-height: 0; overflow-y: auto; overflow-x: hidden;
  padding: 22px 28px 18px;
  display: flex; flex-direction: column; gap: 2px;
  -webkit-overflow-scrolling: touch; overscroll-behavior: contain;
  background: transparent;
}

.messages-area > *:first-child {
  margin-top: auto;
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

.composer-wrap {
  flex-shrink: 0;
  position: sticky;
  bottom: 0;
  z-index: 10;
  padding: 14px 28px 18px;
  background: linear-gradient(180deg, rgba(8, 12, 24, 0.18), rgba(8, 12, 24, 0.32));
}
.theme-light .composer-wrap { background: #f5f6fc; }

.composer {
  min-height: 56px; height: auto; display: flex; align-items: flex-end; gap: 12px;
  padding: 10px 12px 10px 14px; border-radius: 18px;
  border: 1px solid rgba(110, 123, 255, 0.18);
  background: linear-gradient(180deg, rgba(25, 30, 58, 0.68), rgba(18, 23, 46, 0.78));
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.18), inset 0 1px 0 rgba(255, 255, 255, 0.04);
}
.theme-light .composer { background: #ffffff; border-color: rgba(91, 106, 255, 0.2); box-shadow: 0 4px 16px rgba(91, 106, 200, 0.08); }

.composer-side-btn, .send-btn {
  width: 34px; height: 34px; border-radius: 11px;
  display: grid; place-items: center; flex-shrink: 0; cursor: pointer; border: none;
}
.composer-side-btn { color: #a6afd4; background: rgba(255, 255, 255, 0.03); border: 1px solid rgba(255, 255, 255, 0.04); }
.theme-light .composer-side-btn { color: #9098b8; background: #f3f4f8; border-color: #e4e6f0; }

.message-input {
  flex: 1; min-width: 0; background: transparent; border: none; outline: none;
  color: #eef2ff; font-size: 16px; font-weight: 500;
  resize: none; overflow-y: hidden; line-height: 1.4;
  max-height: 120px; overflow-y: auto;
  padding: 4px 0; align-self: center;
  font-family: inherit;
  -webkit-appearance: none;
  appearance: none;
}
.theme-light .message-input { color: #1a1d2e; }
.message-input::placeholder { color: #747ea2; }
.theme-light .message-input::placeholder { color: #aab0cc; }

.send-btn {
  color: white;
  background: linear-gradient(135deg, #6e79ff, #8669ff);
  box-shadow: 0 8px 18px rgba(94, 102, 255, 0.28);
  position: relative;
  user-select: none;
  -webkit-user-select: none;
  touch-action: none;
}

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
.theme-light .delete-modal { background: #ffffff; border-color: #dde1f0; box-shadow: 0 12px 40px rgba(90, 106, 200, 0.15); }
.delete-modal h3 { color: #eef2ff; font-size: 16px; margin-bottom: 8px; }
.delete-modal p { color: #8d96ba; font-size: 13px; margin-bottom: 20px; }
.theme-light .delete-modal h3 { color: #1a1d2e; }
.theme-light .delete-modal p { color: #7880a0; }
.modal-actions { display: flex; gap: 10px; justify-content: center; }
.btn-cancel { padding: 8px 16px; border-radius: 10px; background: rgba(255,255,255,0.05); border: 1px solid rgba(255,255,255,0.05); color: #a6afd4; font-size: 13px; cursor: pointer; }
.theme-light .btn-cancel { background: #f3f4f8; border-color: #e2e4ee; color: #7880a0; }
.btn-delete { padding: 8px 16px; border-radius: 10px; background: linear-gradient(135deg, #ff4d6d, #d93856); border: none; color: white; font-size: 13px; cursor: pointer; }

.back-btn {
  width: 34px; height: 34px; border-radius: 11px;
  display: grid; place-items: center; flex-shrink: 0;
  color: #a6afd4; background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.06); cursor: pointer;
  margin-right: 2px; transition: all 0.2s;
}
.back-btn:hover { background: rgba(255, 255, 255, 0.08); }
.theme-light .back-btn { color: #7880a0; background: #f3f4f8; border-color: #e4e6f0; }
.theme-light .back-btn:hover { background: #e8eaf5; }

@media (max-width: 760px) {
  .chat-header { padding: 10px 14px; height: 64px; min-height: 64px; flex-shrink: 0; }
  .messages-area { padding: 14px 12px 10px; }
  .composer-wrap {
    padding: 8px 12px calc(8px + env(safe-area-inset-bottom));
    flex-shrink: 0;
  }
  .composer { min-height: 50px; height: auto; border-radius: 16px; padding: 8px 10px 8px 12px; gap: 8px; align-items: flex-end; }
  .delete-modal { width: calc(100vw - 48px); max-width: 300px; }
  .messages-area { padding: 14px 12px 10px; }
}

.send-btn.recording {
  background: linear-gradient(135deg, #ff4d6d, #d93856);
  box-shadow: 0 0 0 4px rgba(255, 77, 109, 0.2);
  animation: pulse 1s infinite;
}
.rec-timer-inline {
  position: absolute; top: -22px; left: 50%;
  transform: translateX(-50%);
  font-size: 10px; font-weight: 700;
  color: #ff4d6d; white-space: nowrap;
  background: rgba(0,0,0,0.5);
  padding: 2px 6px; border-radius: 6px;
}
@keyframes pulse {
  0%, 100% { box-shadow: 0 0 0 4px rgba(255, 77, 109, 0.2); }
  50% { box-shadow: 0 0 0 8px rgba(255, 77, 109, 0.1); }
}
.search-bar {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  display: flex; align-items: center; gap: 8px;
  padding: 0 16px;
  background: rgba(13, 17, 32, 0.98);
  backdrop-filter: blur(12px);
  z-index: 20;
}
.theme-light .search-bar { background: #ffffff; }
.search-icon { color: #6e79ff; flex-shrink: 0; }
.search-input {
  flex: 1; background: transparent; border: none; outline: none;
  color: #eef2ff; font-size: 14px; font-weight: 500;
}
.theme-light .search-input { color: #1a1d2e; }
.search-input::placeholder { color: #4a5270; }
.search-count { font-size: 11px; color: #6e79ff; font-weight: 600; white-space: nowrap; flex-shrink: 0; }
.search-nav-btn {
  width: 26px; height: 26px; border-radius: 8px;
  display: grid; place-items: center; flex-shrink: 0;
  color: #a6afd4; background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.06); cursor: pointer;
  transition: all 0.15s;
}
.search-nav-btn:hover { color: #fff; background: rgba(110,121,255,0.2); }
.search-close-btn {
  width: 26px; height: 26px; border-radius: 8px;
  display: grid; place-items: center; flex-shrink: 0;
  color: #a6afd4; background: transparent; border: none; cursor: pointer;
}
.search-close-btn:hover { color: #ff4d6d; }
.chat-icon-btn.active { background: rgba(110,121,255,0.15); border-color: rgba(110,121,255,0.3); color: #6e79ff; }
.search-slide-enter-active, .search-slide-leave-active { transition: opacity 0.2s, transform 0.2s; }
.search-slide-enter-from, .search-slide-leave-to { opacity: 0; transform: translateY(-6px); }
.reply-preview {
  display: flex; align-items: center; justify-content: space-between;
  padding: 8px 16px; margin: 0 0 4px;
  background: rgba(110,121,255,0.1);
  border-left: 3px solid #6e79ff;
  border-radius: 8px;
}
.reply-preview-content { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.reply-preview-name { font-size: 12px; font-weight: 700; color: #6e79ff; }
.reply-preview-text { font-size: 13px; color: #a6afd4; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.reply-preview-close { background: none; border: none; color: #a6afd4; cursor: pointer; font-size: 16px; flex-shrink: 0; }
.selection-bar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 12px 20px;
  background: rgba(16,20,40,0.95);
  backdrop-filter: blur(12px);
  border-top: 1px solid rgba(110,121,255,0.15);
  flex-shrink: 0;
  animation: slideUp 0.2s ease;
}
.sel-cancel {
  background: none; border: none; color: #7880a0;
  cursor: pointer; font-size: 14px; font-family: inherit;
  padding: 6px 10px; border-radius: 8px; transition: color 0.15s;
}
.sel-cancel:hover { color: #eef1fb; }
.sel-count {
  font-size: 14px; color: #eef1fb; font-weight: 700;
  background: rgba(110,121,255,0.12);
  padding: 4px 12px; border-radius: 20px;
  border: 1px solid rgba(110,121,255,0.2);
}
.sel-delete {
  display: flex; align-items: center; gap: 6px;
  background: rgba(255,77,109,0.12);
  border: 1px solid rgba(255,77,109,0.25);
  color: #ff4d6d; border-radius: 10px;
  padding: 7px 16px; cursor: pointer;
  font-size: 14px; font-family: inherit;
  font-weight: 600; transition: all 0.15s;
}
.sel-delete:hover { background: rgba(255,77,109,0.22); }
@keyframes slideUp { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: translateY(0); } }
.chat-status.typing { color: #6e79ff !important; }
.chat-status.typing .status-dot { background: #6e79ff !important; animation: typingPulse 1s infinite; }
@keyframes typingPulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}
</style>