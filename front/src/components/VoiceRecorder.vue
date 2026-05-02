<template>
  <div class="voice-recorder">
    <button
      type="button"
      class="record-btn"
      :class="{ recording: isRecording }"
      @mousedown="startRecording"
      @mouseup="stopRecording"
      @touchstart.prevent="startRecording"
      @touchend.prevent="stopRecording"
    >
      <svg viewBox="0 0 24 24" width="17" height="17" fill="currentColor">
        <path d="M12 1a4 4 0 0 1 4 4v6a4 4 0 0 1-8 0V5a4 4 0 0 1 4-4z"/>
        <path d="M19 10a7 7 0 0 1-14 0H3a9 9 0 0 0 18 0h-2z"/>
        <line x1="12" y1="19" x2="12" y2="23" stroke="currentColor" stroke-width="2"/>
        <line x1="8" y1="23" x2="16" y2="23" stroke="currentColor" stroke-width="2"/>
      </svg>
      <span v-if="isRecording" class="rec-timer">{{ timerText }}</span>
    </button>
  </div>
</template>

<script>
const VOICE_BASE = import.meta.env.VITE_VOICE_API_URL || 'http://localhost:9090'

export default {
  name: 'VoiceRecorder',

  props: {
    chatId: { type: String, required: true },
    senderId: { type: String, required: true },
    recipientId: { type: String, required: true },
    isLight: { type: Boolean, default: false }
  },

  emits: ['voice-sent'],

  data() {
    return {
      isRecording: false,
      mediaRecorder: null,
      chunks: [],
      timerSeconds: 0,
      timerInterval: null
    }
  },

  computed: {
    timerText() {
      const m = Math.floor(this.timerSeconds / 60).toString().padStart(2, '0')
      const s = (this.timerSeconds % 60).toString().padStart(2, '0')
      return `${m}:${s}`
    }
  },

  methods: {
    async startRecording() {
      try {
        const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
        this.chunks = []
        this.mediaRecorder = new MediaRecorder(stream)
        this.mediaRecorder.ondataavailable = e => {
          if (e.data.size > 0) this.chunks.push(e.data)
        }
        this.mediaRecorder.onstop = this.handleStop
        this.mediaRecorder.start()
        this.isRecording = true
        this.timerSeconds = 0
        this.timerInterval = setInterval(() => this.timerSeconds++, 1000)
      } catch (e) {
        console.error('Microphone error', e)
      }
    },

    stopRecording() {
      if (!this.mediaRecorder || !this.isRecording) return
      this.mediaRecorder.stop()
      this.mediaRecorder.stream.getTracks().forEach(t => t.stop())
      this.isRecording = false
      clearInterval(this.timerInterval)
    },

    async handleStop() {
      if (this.chunks.length === 0) return
      const blob = new Blob(this.chunks, { type: 'audio/webm' })

      const form = new FormData()
      form.append('file', blob, `voice_${Date.now()}.webm`)
      form.append('chat_id', this.chatId)
      form.append('sender_id', this.senderId)
      form.append('recipient_id', this.recipientId)

      try {
        const res = await fetch(`${VOICE_BASE}/voice/upload`, {
          method: 'POST',
          headers: {
            Authorization: `Bearer ${sessionStorage.getItem('access_token') || ''}`
          },
          body: form
        })
        if (!res.ok) return
        const data = await res.json()
        this.$emit('voice-sent', data.message)
      } catch (e) {
        console.error('Voice upload error', e)
      }
    }
  }
}
</script>

<style scoped>
.voice-recorder { display: flex; align-items: center; }

.record-btn {
  width: 34px; height: 34px; border-radius: 11px;
  display: grid; place-items: center; flex-shrink: 0;
  cursor: pointer; position: relative;
  color: #a6afd4;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.04);
  transition: all 0.2s;
}
.record-btn.recording {
  color: white;
  background: linear-gradient(135deg, #ff4d6d, #d93856);
  border-color: transparent;
  box-shadow: 0 0 0 4px rgba(255, 77, 109, 0.2);
}
.rec-timer {
  position: absolute; top: -24px; left: 50%;
  transform: translateX(-50%);
  font-size: 10px; font-weight: 700;
  color: #ff4d6d; white-space: nowrap;
  background: rgba(0,0,0,0.4);
  padding: 2px 6px; border-radius: 6px;
}
</style>