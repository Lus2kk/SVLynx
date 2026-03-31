<template>
  <div class="card">
    <Logo />
    <div class="headline">Добро пожаловать</div>
    <div class="subline">Войдите, чтобы начать общение. Быстро и безопасно.</div>

    <div class="divider">
      <div class="divider-line"></div>
      <div class="divider-text">войти через</div>
      <div class="divider-line"></div>
    </div>

    <StatusMsg :type="status.type" :message="status.message" />

    <EmailAuth @status="status = $event" @success="onEmailSuccess" />

    <div class="divider" style="margin-bottom: 24px;">
      <div class="divider-line"></div>
      <div class="divider-text">или</div>
      <div class="divider-line"></div>
    </div>

    <TgButton @auth="onTelegramAuth" />

    <div class="footer">
      Входя, вы соглашаетесь с <a href="#">условиями использования</a><br>
      и <a href="#">политикой конфиденциальности</a>
    </div>
  </div>
</template>

<script>
import Logo from './Logo.vue'
import StatusMsg from './StatusMsg.vue'
import TgButton from './TgButton.vue'
import EmailAuth from './EmailAuth.vue'

export default {
  components: { Logo, StatusMsg, TgButton, EmailAuth },
  emits: ['show-profile'],

  data() {
    return {
      status: { type: '', message: '' }
    }
  },

  methods: {
    onEmailSuccess({ isNew }) {
      if (isNew) {
        setTimeout(() => this.$emit('show-profile'), 100)
      }
    },

    onTelegramAuth(user) {
      this.status = { type: 'success', message: 'Подключение к серверу...' }
     fetch('https://svlynx.site/auth/telegram/callback', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(user)
      })
      .then(res => res.json())
      .then(data => {
        if (data.status === 'approved') {
          this.status = { type: 'success', message: 'Вы вошли! Переход в SVLynx...' }
        } else {
          this.status = { type: 'error', message: 'Ошибка: ' + data.error }
        }
      })
      .catch(() => {
        this.status = { type: 'error', message: 'Не удалось подключиться к серверу' }
      })
    }
  }
}
</script>

<style scoped>
.card {
  position: relative;
  z-index: 10;
  width: 420px;
  background: hwb(224 5% 87%);
  border: 1px solid rgba(255,255,255,0.07);
  border-radius: 24px;
  padding: 48px 44px 40px;
  box-shadow:
    0 0 0 1px rgba(79,142,247,0.08),
    0 32px 80px rgba(0,0,0,0.6),
    inset 0 1px 0 rgba(255,255,255,0.06);
  animation: fadeUp 0.7s cubic-bezier(0.16,1,0.3,1) both;
}
.card::before {
  content: '';
  position: absolute;
  top: 0; left: 10%; right: 10%;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(79,142,247,0.5), transparent);
}
.headline {
  font-family: 'Syne', sans-serif;
  font-size: 28px;
  font-weight: 700;
  color: #f0f2f8;
  margin-bottom: 10px;
  animation: fadeUp 0.5s 0.1s ease both;
}
.subline {
  font-size: 14px;
  color: #5a6480;
  margin-bottom: 36px;
  animation: fadeUp 0.5s 0.2s ease both;
}
.divider {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 28px;
  animation: fadeUp 0.5s 0.3s ease both;
}
.divider-line { flex: 1; height: 1px; background: rgba(255,255,255,0.07); }
.divider-text { font-size: 11px; color: #5a6480; text-transform: uppercase; letter-spacing: 1.5px; }
.footer { text-align: center; font-size: 12px; color: #5a6480; line-height: 1.6; }
.footer a { color: #4f8ef7; text-decoration: none; }

@keyframes fadeUp {
  from { opacity: 0; transform: translateY(20px); }
  to   { opacity: 1; transform: translateY(0); }
}
</style>