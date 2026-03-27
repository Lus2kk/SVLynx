<template>
  <div class="email-auth">
    <div v-if="step === 'email'" class="input-group">
      <input
        v-model="email"
        type="email"
        placeholder="your@gmail.com"
        class="email-input"
        @keyup.enter="sendCode"
      />
      <button class="email-btn" @click="sendCode" :disabled="loading">
        <span v-if="!loading">Получить код</span>
        <span v-else class="spinner"></span>
      </button>
    </div>
    <div v-if="step === 'code'" class="input-group">
      <input
        v-model="code"
        type="text"
        placeholder="000000"
        maxlength="6"
        class="email-input code-input"
        @keyup.enter="verifyCode"
      />
      <button class="email-btn" @click="verifyCode" :disabled="loading">
        <span v-if="!loading">Войти</span>
        <span v-else class="spinner"></span>
      </button>
    </div>
    <div v-if="step === 'code'" class="resend-hint">
      Код отправлен на {{ email }}
      <span class="resend-link" @click="reset">Изменить</span>
    </div>
  </div>
</template>

<script>
const BASE = 'https://course-runtime-albuquerque-links.trycloudflare.com'
export default {
  emits: ['status'],
  data() {
    return { step: 'email', email: '', code: '', sessionId: null, loading: false }
  },
  methods: {
    async sendCode() {
      if (!this.email) return
      this.loading = true
      this.$emit('status', { type: '', message: '' })
      try {
        const initRes = await fetch(`${BASE}/auth/email/init`, { method: 'POST' })
        const initData = await initRes.json()
        if (!initRes.ok) throw new Error(initData.error || 'Ошибка инициализации')
        this.sessionId = initData.sessionid
        const sendRes = await fetch(`${BASE}/auth/email/send-code`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ sessionid: this.sessionId, email: this.email })
        })
        const sendData = await sendRes.json()
        if (!sendRes.ok) throw new Error(sendData.error || 'Ошибка отправки кода')
        this.step = 'code'
        this.$emit('status', { type: 'success', message: `Код отправлен на ${this.email}` })
      } catch (e) {
        this.$emit('status', { type: 'error', message: e.message })
      } finally {
        this.loading = false
      }
    },
    async verifyCode() {
      if (!this.code || this.code.length !== 6) return
      this.loading = true
      this.$emit('status', { type: '', message: '' })
      try {
        const res = await fetch(`${BASE}/auth/email/verify-code`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ sessionid: this.sessionId, code: this.code })
        })
        const data = await res.json()
        if (!res.ok) throw new Error(data.error || 'Неверный код')
        localStorage.setItem('access_token', data.accesstoken)
        localStorage.setItem('refresh_token', data.refreshtoken)
        if (data.isnew) {
          this.$emit('status', { type: 'success', message: 'Добро пожаловать! Заполните профиль...' })
        } else {
          this.$emit('status', { type: 'success', message: 'Вы вошли! Переход в SVLynx...' })
        }
      } catch (e) {
        this.$emit('status', { type: 'error', message: e.message })
      } finally {
        this.loading = false
      }
    },
    reset() {
      this.step = 'email'
      this.code = ''
      this.sessionId = null
      this.$emit('status', { type: '', message: '' })
    }
  }
}
</script>
<style scoped>
.email-auth { margin-bottom: 32px; }
.input-group { display: flex; gap: 8px; }
.email-input {
  flex: 1;
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.09);
  border-radius: 12px;
  padding: 12px 16px;
  color: #f0f2f8;
  font-size: 14px;
  font-family: 'DM Sans', sans-serif;
  outline: none;
  transition: border-color 0.2s;
}
.email-input::placeholder { color: #3a4060; }
.email-input:focus { border-color: rgba(79,142,247,0.5); }
.code-input { letter-spacing: 6px; font-size: 18px; text-align: center; }
.email-btn {
  background: linear-gradient(135deg, #4f8ef7, #7c5ef7);
  border: none;
  border-radius: 12px;
  padding: 12px 18px;
  color: #fff;
  font-size: 13px;
  font-family: 'DM Sans', sans-serif;
  cursor: pointer;
  white-space: nowrap;
  transition: opacity 0.2s;
  min-width: 110px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.email-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.email-btn:hover:not(:disabled) { opacity: 0.85; }
.resend-hint { margin-top: 10px; font-size: 12px; color: #5a6480; text-align: center; }
.resend-link { color: #4f8ef7; cursor: pointer; margin-left: 6px; }
.resend-link:hover { text-decoration: underline; }
.spinner {
  display: inline-block;
  width: 14px; height: 14px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
</style>
