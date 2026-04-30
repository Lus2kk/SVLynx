<template>
  <div class="email-auth">
    <transition name="slide" mode="out-in">
      <div v-if="step === 'email'" key="email" class="email-step">
        <input
          ref="emailInput"
          v-model.trim="email"
          type="email"
          placeholder="your@gmail.com"
          class="email-input"
          @keyup.enter="sendCode"
          autocomplete="email"
          :disabled="loading || cooldown > 0"
        />
        <button 
          class="email-btn" 
          @click="sendCode" 
          :disabled="loading || cooldown > 0"
        >
          <span v-if="!loading">
            {{ cooldown > 0 ? `Повтор через ${cooldown} сек` : 'Получить код' }}
          </span>
          <span v-else class="spinner"></span>
        </button>
      </div>

      <div v-else key="code" class="code-step">
        <div class="code-hint">
          Код отправлен на <strong>{{ maskedEmail }}</strong>
          <span class="change-link" @click="reset">Изменить</span>
        </div>

        <div class="pin-container">
          <input
            v-for="(digit, index) in 6"
            :key="index"
            :ref="el => { if (el) pinInputs[index] = el }"
            v-model="codeDigits[index]"
            type="text"
            inputmode="numeric"
            maxlength="1"
            class="pin-input"
            :disabled="loading || verifyLock > 0"
            @input="onPinInput(index, $event)"
            @keydown="onPinKeydown(index, $event)"
            @paste="onPinPaste($event)"
          />
        </div>

        <button
          class="email-btn pin-btn"
          @click="verifyCode"
          :disabled="loading || code.length !== 6 || verifyLock > 0"
        >
          <span v-if="!loading">
            {{ verifyLock > 0 ? `Повтор через ${verifyLock} сек` : 'Войти →' }}
          </span>
          <span v-else class="spinner"></span>
        </button>
      </div>
    </transition>
  </div>
</template>

<script>

const BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

export default {
  emits: ['status', 'success'],

  data() {
    return {
      step: 'email',
      email: '',
      code: '',
      codeDigits: ['', '', '', '', '', ''],
      pinInputs: [],
      sessionId: null,
      loading: false,
      cooldown: 0,
      cooldownTimer: null,
      verifyLock: 0,
      verifyLockTimer: null
    }
  },

  computed: {
    maskedEmail() {
      if (!this.email) return ''
      const [user, domain] = this.email.split('@')
      if (!domain || !user) return this.email
      const masked = user[0] + '***'
      return `${masked}@${domain}`
    }
  },

  mounted() {
    this.$nextTick(() => {
      this.$refs.emailInput?.focus()
    })
  },

  methods: {
    onPinInput(index, e) {
      if (this.verifyLock > 0) return
      const val = e.target.value.replace(/\D/g, '')
      this.codeDigits[index] = val ? val[val.length - 1] : ''
      this.code = this.codeDigits.join('')
      if (val && index < 5) this.pinInputs[index + 1]?.focus()
    },

    onPinKeydown(index, e) {
      if (this.verifyLock > 0) return
      if (e.key === 'Backspace' && !this.codeDigits[index] && index > 0) {
        this.pinInputs[index - 1]?.focus()
      }
      if (e.key === 'ArrowLeft' && index > 0) {
        e.preventDefault()
        this.pinInputs[index - 1]?.focus()
      }
      if (e.key === 'ArrowRight' && index < 5) {
        e.preventDefault()
        this.pinInputs[index + 1]?.focus()
      }
    },

    onPinPaste(e) {
      if (this.verifyLock > 0) return
      e.preventDefault()
      const paste = e.clipboardData.getData('text').replace(/\D/g, '').slice(0, 6)
      paste.split('').forEach((char, i) => {
        if (i < 6) this.codeDigits[i] = char
      })
      this.code = this.codeDigits.join('')
      const nextIndex = Math.min(paste.length, 5)
      this.pinInputs[nextIndex]?.focus()
    },

    startCooldown(seconds = 30) {
      this.cooldown = seconds
      if (this.cooldownTimer) clearInterval(this.cooldownTimer)
      this.cooldownTimer = setInterval(() => {
        this.cooldown--
        if (this.cooldown <= 0) {
          clearInterval(this.cooldownTimer)
          this.cooldownTimer = null
          this.cooldown = 0
        }
      }, 1000)
    },

    startVerifyLock(seconds = 5) {
      this.verifyLock = seconds
      if (this.verifyLockTimer) clearInterval(this.verifyLockTimer)
      this.verifyLockTimer = setInterval(() => {
        this.verifyLock--
        if (this.verifyLock <= 0) {
          clearInterval(this.verifyLockTimer)
          this.verifyLockTimer = null
          this.verifyLock = 0
        }
      }, 1000)
    },

    isEmailValid() {
      const input = this.$refs.emailInput
      if (!input) return false
      if (!input.checkValidity()) {
        input.reportValidity()
        return false
      }
      return true
    },

    async sendCode() {
      this.email = this.email.trim()
      if (!this.email || !this.isEmailValid() || this.cooldown > 0) return

      this.loading = true
      this.$emit('status', { type: '', message: '' })

      try {
        const initRes = await fetch(`${BASE}/auth/email/init`, { method: 'POST' })
        const initData = await initRes.json()
        if (!initRes.ok) throw new Error(initData.error || 'Ошибка инициализации')

        this.sessionId = initData.session_id

        const sendRes = await fetch(`${BASE}/auth/email/send-code`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ session_id: this.sessionId, email: this.email })
        })

        const sendData = await sendRes.json()
        if (!sendRes.ok) throw new Error(sendData.error || 'Ошибка отправки кода')

        this.step = 'code'
        this.code = ''
        this.codeDigits = ['', '', '', '', '', '']
        this.startCooldown(30)

        this.$emit('status', { type: '', message: '' })

        this.$nextTick(() => this.pinInputs[0]?.focus())
      } catch (e) {
        this.startCooldown(5)
        this.$emit('status', {
          type: 'error',
          message: `${e.message || 'Ошибка отправки кода'}. Повторите через 5 секунд`
        })
      } finally {
        this.loading = false
      }
    },

    async verifyCode() {
      if (!this.code || this.code.length !== 6 || !this.sessionId || this.verifyLock > 0) return

      this.loading = true
      this.$emit('status', { type: '', message: '' })

      try {
        const res = await fetch(`${BASE}/auth/email/verify-code`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ session_id: this.sessionId, code: this.code })
        })

        const data = await res.json()
        if (!res.ok) throw new Error(data.error || 'Неверный код')

        sessionStorage.setItem('access_token', data.access_token)
        sessionStorage.setItem('refresh_token', data.refresh_token)

        if (data.needs_profile) {
          this.$emit('status', { type: 'success', message: 'Добро пожаловать! Заполните профиль...' })
          this.$emit('success', { needsProfile: true })
        } else {
          this.$emit('status', { type: 'success', message: 'Вы вошли! Переход в SVLynx...' })
          this.$emit('success', { needsProfile: false })
        }
      } catch (e) {
        this.startVerifyLock(5)
        this.code = ''
        this.codeDigits = ['', '', '', '', '', '']
        this.$emit('status', {
          type: 'error',
          message: `${e.message || 'Ошибка проверки кода'}. Повторите через 5 секунд`
        })
        this.$nextTick(() => this.pinInputs[0]?.focus())
      } finally {
        this.loading = false
      }
    },

    reset() {
      this.step = 'email'
      this.code = ''
      this.codeDigits = ['', '', '', '', '', '']
      this.pinInputs = []
      this.sessionId = null

      if (this.verifyLockTimer) {
        clearInterval(this.verifyLockTimer)
        this.verifyLockTimer = null
      }
      this.verifyLock = 0

      this.$emit('status', { type: '', message: '' })
      this.$nextTick(() => this.$refs.emailInput?.focus())
    }
  },

  beforeUnmount() {
    if (this.cooldownTimer) clearInterval(this.cooldownTimer)
    if (this.verifyLockTimer) clearInterval(this.verifyLockTimer)
  }
}
</script>

<style scoped>
.email-auth {
  margin-bottom: 24px;
  width: 100%;
}

.email-step, .code-step {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
}

.email-input {
  width: 100%;
  box-sizing: border-box;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(79, 142, 247, 0.15);
  border-radius: 12px;
  padding: 14px 16px;
  color: #f0f2f8;
  font-size: 16px;
  font-family: 'DM Sans', sans-serif;
  outline: none;
  transition: border-color 0.2s, box-shadow 0.2s;
  min-height: 48px;
}

.email-input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.email-input:-webkit-autofill {
  -webkit-box-shadow: 0 0 0px 1000px #0d1320 inset !important;
  -webkit-text-fill-color: #f0f2f8 !important;
  caret-color: #f0f2f8;
}

.email-input::placeholder { color: #3a4060; }
.email-input:focus:not(:disabled) { border-color: rgba(79, 142, 247, 0.5); box-shadow: 0 0 0 3px rgba(79, 142, 247, 0.1); }

.email-btn {
  width: 100%;
  background: linear-gradient(135deg, #4f8ef7, #7c5ef7);
  border: none;
  border-radius: 12px;
  padding: 14px 18px;
  color: #fff;
  font-size: 15px;
  font-weight: 500;
  font-family: 'DM Sans', sans-serif;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 48px;
  transition: opacity 0.2s, transform 0.1s, box-shadow 0.2s;
}

.email-btn:disabled { opacity: 0.4; cursor: not-allowed; box-shadow: none; transform: none; }
.email-btn:hover:not(:disabled) { opacity: 0.9; box-shadow: 0 8px 24px rgba(79, 142, 247, 0.4); transform: translateY(-1px); }
.email-btn:active:not(:disabled) { transform: scale(0.97); }

.pin-container {
  display: flex;
  gap: 6px;
  justify-content: center;
  margin-bottom: 20px;
  width: 100%;
}

.pin-input {
  flex: 1;
  width: 100%;
  max-width: 44px;
  height: 50px;
  box-sizing: border-box;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(79, 142, 247, 0.2);
  border-radius: 10px;
  color: #f0f2f8;
  font-size: 18px;
  font-weight: 700;
  text-align: center;
  outline: none;
  transition: border-color 0.2s, box-shadow 0.2s;
  font-family: 'DM Sans', sans-serif;
  caret-color: #4f8ef7;
  padding: 0;
}

.pin-input:focus:not(:disabled) { border-color: #4f8ef7; box-shadow: 0 0 0 3px rgba(79, 142, 247, 0.15); }
.pin-input:disabled { opacity: 0.5; cursor: not-allowed; }

.code-hint { font-size: 12px; color: #5a6480; margin-bottom: 12px; text-align: center; }
.code-hint strong { color: #a0aec0; }
.change-link { color: #4f8ef7; cursor: pointer; margin-left: 8px; font-size: 12px; }
.change-link:hover { text-decoration: underline; }

.spinner {
  display: inline-block;
  width: 16px; height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }

.slide-enter-active, .slide-leave-active { transition: all 0.25s ease; }
.slide-enter-from { opacity: 0; transform: translateX(20px); }
.slide-leave-to { opacity: 0; transform: translateX(-20px); }

@media (min-width: 480px) {
  .pin-container { gap: 8px; }
  .pin-input { max-width: 48px; height: 52px; font-size: 20px; border-radius: 12px; }
}
</style>