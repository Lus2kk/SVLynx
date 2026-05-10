<template>
  <div class="media-uploader">
    <input
      ref="fileInput"
      type="file"
      accept="image/*,video/*,audio/*,.pdf,.zip,.txt"
      style="display: none"
      @change="handleFile"
    />
    <button
      type="button"
      class="attach-btn"
      :class="{ uploading: isUploading }"
      @click="$refs.fileInput.click()"
      :disabled="isUploading"
    >
      <svg v-if="!isUploading" viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="1.8">
        <path d="M21.44 11.05l-8.49 8.49a5.5 5.5 0 0 1-7.78-7.78l9.2-9.19a3.5 3.5 0 0 1 4.95 4.95l-9.19 9.2a1.5 1.5 0 0 1-2.12-2.13l8.49-8.48"></path>
      </svg>
      <svg v-else viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="1.8">
        <path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/>
      </svg>
    </button>
  </div>
</template>

<script>
const MEDIA_BASE = import.meta.env.VITE_MEDIA_API_URL || 'http://localhost:9091'

function getCookie(name) {
  const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'))
  return match ? match[2] : null
}

export default {
  name: 'MediaUploader',
  props: {
    chatId:      { type: String, required: true },
    senderId:    { type: String, required: true },
    recipientId: { type: String, required: true }
  },
  emits: ['media-sent'],
  data() {
    return { isUploading: false }
  },
  methods: {
    async handleFile(e) {
      const file = e.target.files[0]
      if (!file) return
      this.isUploading = true
      const form = new FormData()
      form.append('file', file, file.name)
      form.append('chat_id', this.chatId)
      form.append('sender_id', this.senderId)
      form.append('recipient_id', this.recipientId)
      form.append('sender_name', '')
      try {
        const res = await fetch(`${MEDIA_BASE}/media/upload`, {
          method: 'POST',
          headers: { Authorization: `Bearer ${getCookie('access_token') || localStorage.getItem('access_token') || ''}` },
          body: form
        })
        if (!res.ok) { console.error('Media upload failed', await res.text()); return }
        const data = await res.json()
        this.$emit('media-sent', data.message)
      } catch (e) {
        console.error('Media upload error', e)
      } finally {
        this.isUploading = false
        this.$refs.fileInput.value = ''
      }
    }
  }
}
</script>

<style scoped>
.media-uploader { display: flex; align-items: center; }
.attach-btn {
  width: 34px; height: 34px; border-radius: 11px;
  display: grid; place-items: center; flex-shrink: 0;
  cursor: pointer; color: #a6afd4;
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.04);
  transition: all 0.2s;
}
.attach-btn:hover { color: #6e79ff; background: rgba(110,121,255,0.1); }
.attach-btn.uploading { color: #6e79ff; animation: spin 1s linear infinite; }
.attach-btn:disabled { cursor: not-allowed; opacity: 0.5; }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
</style>