
<template>
  <div class="login-wrapper">
    <div class="card">
      <div class="card-content">
        <Logo />
        
        <div class="headline">Добро пожаловать</div>
        <div class="subline">Войдите, чтобы начать общение. Быстро и безопасно.</div>

        <div class="divider">
          <div class="divider-line"></div>
          <div class="divider-text">войти через</div>
          <div class="divider-line"></div>
        </div>

        <StatusMsg :type="status.type" :message="status.message" />

        <EmailAuth @status="onStatus" @success="onEmailSuccess" />

        <div class="divider" style="margin-bottom: 24px;">
          <div class="divider-line"></div>
          <div class="divider-text">или</div>
          <div class="divider-line"></div>
        </div>

        <TgButton @auth="onTelegramAuth" />

        <div class="footer">
          Входя, вы соглашаетесь с <a href="#">условиями использования</a><br class="desktop-break">
          и <a href="#">политикой конфиденциальности</a>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import Logo from './Logo.vue'
import StatusMsg from './StatusMsg.vue'
import TgButton from './TgButton.vue'
import EmailAuth from './EmailAuth.vue'
import { usePush } from '../composables/usePush.js'
import { translateError } from '../composables/ErrorMessages.js'

function setCookie(name, value, maxAgeSeconds) {
  document.cookie = `${name}=${value}; path=/; max-age=${maxAgeSeconds}; SameSite=Strict`
}

function getCookie(name) {
  const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'))
  return match ? match[2] : null
}

export default {
  components: { Logo, StatusMsg, TgButton, EmailAuth },
  emits: ['show-profile', 'show-chat'],

  data() {
    return {
      status: { type: '', message: '' }
    }
  },

  methods: {
    async onEmailSuccess({ needsProfile }) {
      const { subscribe } = usePush()
      try { await subscribe() } catch (e) { console.warn('Push subscribe error:', e) }
      
      if (needsProfile) {
        setTimeout(() => this.$emit('show-profile'), 100)
      } else {
        setTimeout(() => this.$emit('show-chat'), 100)
      }
    },

    onStatus(event) {
      if (event.type === 'error') {
        this.status = { type: 'error', message: translateError(event.message) }
      } else {
        this.status = event
      }
    },

    async onTelegramAuth(data) {
      if (data.access_token) {
        setCookie('access_token', data.access_token, 900)
        setCookie('refresh_token', data.refresh_token, 2592000)
        sessionStorage.setItem('current_user_name', data.sender_name || '')

        this.status = { type: 'success', message: 'Вы вошли! Переход в SVLynx...' }
        
        const { subscribe } = usePush()
        try { await subscribe() } catch (e) { console.warn('Push subscribe error:', e) }
        
        setTimeout(() => this.$emit('show-chat'), 500)
      } else {
        this.status = { type: 'error', message: translateError(data.error) }
      }
    }
  }
}
</script>

<style scoped>
.login-wrapper {
  width: 100%;
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 0;
  box-sizing: border-box;
}

.card {
  position: relative;
  z-index: 10;
  width: 100%;
  min-height: 100vh;
  background: hwb(224 5% 87%);
  border: none;
  border-radius: 0;
  padding: 24px 24px;
  box-shadow: none;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  animation: fadeUp 0.7s cubic-bezier(0.16,1,0.3,1) both;
}

.card-content {
  width: 100%;
  max-width: 420px;
  margin: auto auto;
  display: flex;
  flex-direction: column;
}

.card::before {
  display: none;
}

.headline {
  font-family: 'Syne', sans-serif;
  font-size: 26px;
  font-weight: 700;
  color: #f0f2f8;
  margin-bottom: 8px;
  animation: fadeUp 0.5s 0.1s ease both;
}

.subline {
  font-size: 14px;
  color: #5a6480;
  margin-bottom: 32px;
  line-height: 1.4;
  animation: fadeUp 0.5s 0.2s ease both;
}

.divider {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 24px;
  animation: fadeUp 0.5s 0.3s ease both;
}

.divider-line {
  flex: 1;
  height: 1px;
  background: rgba(255,255,255,0.07);
}

.divider-text {
  font-size: 11px;
  color: #5a6480;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.footer {
  text-align: center;
  font-size: 12px;
  color: #5a6480;
  line-height: 1.5;
  margin-top: 16px;
}

.footer a {
  color: #4f8ef7;
  text-decoration: none;
}

.desktop-break {
  display: none;
}

@keyframes fadeUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

@media (min-width: 480px) {
  .login-wrapper {
    padding: 16px;
  }

  .card {
    max-width: 420px;
    min-height: auto;
    border: 1px solid rgba(255,255,255,0.07);
    border-radius: 24px;
    padding: 48px 44px 40px;
    overflow-y: visible;
    box-shadow:
      0 0 0 1px rgba(79,142,247,0.08),
      0 32px 80px rgba(0,0,0,0.6),
      inset 0 1px 0 rgba(255,255,255,0.06);
  }

  .card-content {
    margin: 0 auto;
  }

  .card::before {
    display: block;
    content: '';
    position: absolute;
    top: 0; left: 10%; right: 10%;
    height: 1px;
    background: linear-gradient(90deg, transparent, rgba(79,142,247,0.5), transparent);
  }

  .headline {
    font-size: 28px;
    margin-bottom: 10px;
  }

  .subline {
    margin-bottom: 36px;
  }

  .divider {
    margin-bottom: 28px;
  }

  .divider-text {
    letter-spacing: 1.5px;
  }

  .footer {
    line-height: 1.6;
  }

  .desktop-break {
    display: inline;
  }
}

@media (max-width: 760px) {
  .login-wrapper {
    min-height: 100dvh;
    background: hwb(224 5% 87%);
    display: flex;
    flex-direction: column;
    padding-bottom: 0;
  }

  .card {
    flex: 1;
    min-height: 100dvh;
    padding-top: calc(24px + env(safe-area-inset-top));
    padding-bottom: calc(24px + env(safe-area-inset-bottom));
    background: hwb(224 5% 87%);
    position: sticky;
    bottom: 0;
  }
}
</style>