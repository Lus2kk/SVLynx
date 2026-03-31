<template>
  <div class="email-auth">
    <transition name="slide" mode="out-in">
      <!-- Шаг 1: email -->
      <div v-if="step === 'email'" key="email" class="email-step">
        <input
          ref="emailInput"
          v-model.trim="email"
          type="email"
          placeholder="your@gmail.com"
          class="email-input"
          @keyup.enter="sendCode"
          autocomplete="email"
        />
        <button class="email-btn" @click="sendCode" :disabled="loading">
          <span v-if="!loading">Получить код</span>
          <span v-else class="spinner"></span>
        </button>
      </div>

      <!-- Шаг 2: код -->
      <div v-else key="code">
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
            @input="onPinInput(index, $event)"
            @keydown="onPinKeydown(index, $event)"
            @paste="onPinPaste($event)"
          />
        </div>

        <button
          class="email-btn pin-btn"
          @click="verifyCode"
          :disabled="loading || code.length !== 6"
        >
          <span v-if="!loading">Войти →</span>
          <span v-else class="spinner"></span>
        </button>

        <div class="resend-row">
          <span v-if="cooldown > 0" class="cooldown-text">
            Отправить снова через {{ cooldown }} сек
          </span>
          <span v-else class="resend-link" @click="resendCode">
            Отправить код снова
          </span>
        </div>
      </div>
    </transition>
  </div>
</template>

<script>
const BASE = '  http://localhost:8080'                                 

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
      cooldownTimer: null
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
    // ------- PIN -------

    onPinInput(index, e) {
      const val = e.target.value.replace(/\D/g, '')
      this.codeDigits[index] = val ? val[val.length - 1] : ''
      this.code = this.codeDigits.join('')
      if (val && index < 5) {
        this.pinInputs[index + 1]?.focus()
      }
    },

    onPinKeydown(index, e) {
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
      e.preventDefault()
      const paste = e.clipboardData.getData('text').replace(/\D/g, '').slice(0, 6)
      paste.split('').forEach((char, i) => {
        if (i < 6) this.codeDigits[i] = char
      })
      this.code = this.codeDigits.join('')
      const nextIndex = Math.min(paste.length, 5)
      this.pinInputs[nextIndex]?.focus()
    },

    // ------- COOLDOWN -------

    startCooldown(seconds = 60) {
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

    // ------- EMAIL CHECK -------

    isEmailValid() {
      const input = this.$refs.emailInput
      if (!input) return false
      // Явно запускаем HTML5-валидацию
      if (!input.checkValidity()) {
        input.reportValidity()
        return false
      }
      return true
    },

    // ------- API CALLS -------

    async sendCode() {
      this.email = this.email.trim()
      if (!this.email) return

      if (!this.isEmailValid()) {
        return
      }

      this.loading = true
      this.$emit('status', { type: '', message: '' })

      try {
        // Шаг 1: init
        const initRes = await fetch(`${BASE}/auth/email/init`, { method: 'POST' })
        const initData = await initRes.json()
        if (!initRes.ok) throw new Error(initData.error || 'Ошибка инициализации')
        this.sessionId = initData.session_id

        // Шаг 2: отправка кода
        const sendRes = await fetch(`${BASE}/auth/email/send-code`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ session_id: this.sessionId, email: this.email })
        })
        const sendData = await sendRes.json()
        if (!sendRes.ok) throw new Error(sendData.error || 'Ошибка отправки кода')

        this.step = 'code'
        this.startCooldown(60)
        this.$emit('status', { type: 'success', message: `Код отправлен на ${this.maskedEmail}` })
        this.$nextTick(() => this.pinInputs[0]?.focus())
      } catch (e) {
        this.$emit('status', { type: 'error', message: e.message || 'Ошибка отправки кода' })
      } finally {
        this.loading = false
      }
    },

    async resendCode() {
      if (this.cooldown > 0 || !this.sessionId) return
      await this.sendCode()
    },

    async verifyCode() {
      if (!this.code || this.code.length !== 6 || !this.sessionId) return

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

        if (data.is_new) {
          this.$emit('status', {
            type: 'success',
            message: 'Добро пожаловать! Заполните профиль...'
          })
          this.$emit('success', { isNew: true })
        } else {
          this.$emit('status', {
            type: 'success',
            message: 'Вы вошли! Переход в SVLynx...'
          })
          this.$emit('success', { isNew: false })
        }
      } catch (e) {
        this.$emit('status', { type: 'error', message: e.message || 'Ошибка проверки кода' })
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
      if (this.cooldownTimer) {
        clearInterval(this.cooldownTimer)
        this.cooldownTimer = null
      }
      this.cooldown = 0
      this.$emit('status', { type: '', message: '' })
      this.$nextTick(() => this.$refs.emailInput?.focus())
    }
  },

  beforeUnmount() {
    if (this.cooldownTimer) {
      clearInterval(this.cooldownTimer)
      this.cooldownTimer = null
    }
  }
}
</script>

<style scoped>
.email-auth {
  margin-bottom: 32px;
}

.email-step {
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: 100%;
}

.email-step .email-btn {
  width: 100%;
}

.email-input {
  flex: 1;
  min-width: 0;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(79, 142, 247, 0.15);
  border-radius: 12px;
  padding: 12px 16px;
  color: #f0f2f8;
  font-size: 14px;
  font-family: 'DM Sans', sans-serif;
  outline: none;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.email-input:-webkit-autofill,
.email-input:-webkit-autofill:hover,
.email-input:-webkit-autofill:focus {
  -webkit-box-shadow: 0 0 0px 1000px #0d1320 inset !important;
  -webkit-text-fill-color: #f0f2f8 !important;
  caret-color: #f0f2f8;
  transition: background-color 5000s ease-in-out 0s;
}

.email-input::placeholder {
  color: #3a4060;
}

.email-input:focus {
  border-color: rgba(79, 142, 247, 0.5);
  box-shadow: 0 0 0 3px rgba(79, 142, 247, 0.1);
}

/* PIN */

.pin-container {
  display: flex;
  gap: 10px;
  justify-content: center;
  margin-bottom: 20px;
}

.pin-input {
  width: 48px;
  height: 56px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(79, 142, 247, 0.2);
  border-radius: 12px;
  color: #f0f2f8;
  font-size: 22px;
  font-weight: 700;
  text-align: center;
  outline: none;
  transition: border-color 0.2s, box-shadow 0.2s;
  font-family: 'DM Sans', sans-serif;
  caret-color: #4f8ef7;
}

.pin-input:focus {
  border-color: #4f8ef7;
  box-shadow: 0 0 0 3px rgba(79, 142, 247, 0.15);
}

/* BUTTONS */

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
  transition: opacity 0.2s, transform 0.1s, box-shadow 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.pin-btn {
  width: 100%;
  padding: 14px;
  font-size: 15px;
  font-weight: 500;
}

.email-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.email-btn:hover:not(:disabled) {
  opacity: 0.9;
  box-shadow: 0 8px 24px rgba(79, 142, 247, 0.4);
  transform: translateY(-1px);
}

.email-btn:active:not(:disabled) {
  transform: scale(0.97);
}

/* TEXT */

.code-hint {
  font-size: 12px;
  color: #5a6480;
  margin-bottom: 16px;
  text-align: center;
}

.code-hint strong {
  color: #a0aec0;
}

.change-link {
  color: #4f8ef7;
  cursor: pointer;
  margin-left: 8px;
  font-size: 12px;
}

.change-link:hover {
  text-decoration: underline;
}

/* RESEND */

.resend-row {
  margin-top: 14px;
  text-align: center;
  font-size: 12px;
}

.cooldown-text {
  color: #3a4060;
}

.resend-link {
  color: #4f8ef7;
  cursor: pointer;
}

.resend-link:hover {
  text-decoration: underline;
}

/* LOADER */

.spinner {
  display: inline-block;
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* ANIMATION */

.slide-enter-active,
.slide-leave-active {
  transition: all 0.25s ease;
}

.slide-enter-from {
  opacity: 0;
  transform: translateX(20px);
}

.slide-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}
</style>