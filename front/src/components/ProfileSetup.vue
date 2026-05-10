<template>
  <div class="profile-wrapper">
    <div class="card profile-card">
      <div class="card-content">
        <Logo />

        <div class="headline">Настройки профиля</div>
        <div class="subline">Заполните данные чтобы начать общение</div>

        <StatusMsg :type="status.type" :message="status.message" />

        <div class="form">
          <div class="field">
            <label class="label">Никнейм <span class="required">*</span></label>
            <input
              v-model="nickname"
              type="text"
              placeholder="Введите никнейм"
              class="input"
              maxlength="25"
              spellcheck="false"
            />
            <div class="field-footer">
              <span class="hint">От 3 до 25 символов</span>
              <span class="counter">{{ nickname.length }}/25</span>
            </div>
          </div>

          <div class="field">
            <label class="label">Имя</label>
            <input
              v-model="name"
              type="text"
              placeholder="Ваше имя"
              class="input"
              maxlength="50"
            />
            <div class="field-footer"></div>
          </div>

          <div class="field">
            <label class="label">Статус</label>
            <input
              v-model="statusText"
              type="text"
              placeholder="Расскажите немного о себе"
              class="input"
              maxlength="100"
            />
            <div class="field-footer"></div>
          </div>
        </div>

        <button class="submit-btn" @click="submit" :disabled="loading || nickname.length < 3">
          <span v-if="!loading">Начать общение →</span>
          <span v-else class="spinner"></span>
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import Logo from './Logo.vue'
import StatusMsg from './StatusMsg.vue'
import { apiFetch } from '../api.js'

const BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

function getCookie(name) {
  const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'))
  return match ? match[2] : null
}

export default {
  components: { Logo, StatusMsg },
  emits: ['done'],

  data() {
    return {
      nickname: '',
      name: '',
      statusText: '',
      loading: false,
      status: { type: '', message: '' }
    }
  },

  methods: {
    async submit() {
      if (this.nickname.length < 3) return
      this.loading = true
      this.status = { type: '', message: '' }

      try {
        const res = await apiFetch(`${BASE}/auth/email/complete`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            nickname: this.nickname,
            name: this.name,
            status: this.statusText
          })
        })
        const data = await res.json()
        if (!res.ok) throw new Error(data.error || 'Ошибка')

        this.status = { type: 'success', message: 'Профиль создан! Добро пожаловать в SVLynx' }
        setTimeout(() => this.$emit('done'), 1000)
      } catch (e) {
        this.status = { type: 'error', message: e.message }
      } finally {
        this.loading = false
      }
    }
  }
}
</script>

<style scoped>
.profile-wrapper {
  width: 100%;
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 0;
  box-sizing: border-box;
}

.profile-card {
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

.profile-card::before { display: none; }

.headline {
  font-family: 'Syne', sans-serif;
  font-size: 26px;
  font-weight: 700;
  color: #f0f2f8;
  margin-bottom: 8px;
  animation: slideUp 0.4s ease both;
}

.subline {
  font-size: 14px;
  color: #5a6480;
  margin-bottom: 24px;
  line-height: 1.4;
  animation: slideUp 0.4s 0.05s ease both;
}

.form { margin-bottom: 24px; }
.field { margin-bottom: 16px; animation: slideUp 0.6s ease both; }
.field:nth-child(1) { animation-delay: 0.1s; }
.field:nth-child(2) { animation-delay: 0.25s; }
.field:nth-child(3) { animation-delay: 0.4s; }

.label {
  display: block;
  font-size: 11px;
  color: #839dea;
  margin-bottom: 8px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 1px;
}
.required { color: #4f8ef7; }

.input {
  width: 100%;
  box-sizing: border-box;
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.09);
  border-radius: 12px;
  padding: 14px 16px;
  color: #f0f2f8;
  font-size: 16px;
  font-family: 'DM Sans', sans-serif;
  outline: none;
  transition: border-color 0.2s, box-shadow 0.2s;
  min-height: 48px;
}
.input::placeholder { color: #3a4060; }
.input:focus {
  border-color: rgba(79,142,247,0.5);
  box-shadow: 0 0 0 3px rgba(79,142,247,0.1);
}

.field-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 6px;
}
.hint, .counter { font-size: 11px; color: #5a6480; }
.counter.warn { color: #f87171; }

.submit-btn {
  width: 100%;
  background: linear-gradient(135deg, #4f8ef7, #7c5ef7);
  border: none;
  border-radius: 12px;
  padding: 14px;
  color: #fff;
  font-size: 15px;
  font-family: 'DM Sans', sans-serif;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 48px;
  animation: slideUp 0.4s 0.4s ease both;
}
.submit-btn:disabled { opacity: 0.4; cursor: not-allowed; }

.spinner {
  display: inline-block;
  width: 16px; height: 16px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
@keyframes fadeUp { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: translateY(0); } }
@keyframes slideUp { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: translateY(0); } }

@media (min-width: 480px) {
  .profile-wrapper { padding: 16px; }
  .profile-card {
    max-width: 420px;
    min-height: auto;
    border: 1px solid rgba(255,255,255,0.07);
    border-radius: 24px;
    padding: 48px 44px 40px;
    overflow-y: visible;
    box-shadow: 0 0 0 1px rgba(79,142,247,0.08), 0 32px 80px rgba(0,0,0,0.6), inset 0 1px 0 rgba(255,255,255,0.06);
  }
  .card-content { margin: 0 auto; }
  .profile-card::before {
    display: block;
    content: '';
    position: absolute;
    top: 0; left: 10%; right: 10%;
    height: 1px;
    background: linear-gradient(90deg, transparent, rgba(79,142,247,0.5), transparent);
  }
  .headline { font-size: 28px; margin-bottom: 10px; }
  .subline { margin-bottom: 24px; }
  .input { font-size: 14px; padding: 12px 16px; }
}
</style>