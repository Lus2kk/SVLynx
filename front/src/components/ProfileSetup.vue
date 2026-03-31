<template>
  <div class="card profile-card">
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
</template>

<script>
import Logo from './Logo.vue'
import StatusMsg from './StatusMsg.vue'

const BASE = '  http://localhost:8080'

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
        const accessToken = sessionStorage.getItem('access_token')
        const res = await fetch(`${BASE}/auth/email/complete`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${accessToken}`
          },
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
.field-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 5px;
}
.counter {
  font-size: 11px;
  color: #3a4060;
}
.counter.warn {
  color: #f87171;
}
.profile-card {
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
.profile-card::before {
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
  animation: slideUp 0.4s ease both;
}
.form { margin-bottom: 24px; }
.field {
  margin-bottom: 14px;
  animation: slideUp 0.6s ease both;
}
.field:nth-child(1) { animation-delay: 0.1s; }
.field:nth-child(2) { animation-delay: 0.25s; }
.field:nth-child(3) { animation-delay: 0.4s; }
.subline {
  font-size: 14px;
  color: #5a6480;
  margin-bottom: 24px;
  animation: slideUp 0.4s 0.05s ease both;
}
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
  padding: 12px 16px;
  color: #f0f2f8;
  font-size: 14px;
  font-family: 'DM Sans', sans-serif;
  outline: none;
  transition: border-color 0.2s, box-shadow 0.2s;
}
.input::placeholder { color: #3a4060; }
.input:focus {
  border-color: rgba(79,142,247,0.5);
  box-shadow: 0 0 0 3px rgba(79,142,247,0.1);
}
.hint {
  display: block;
  font-size: 11px;
  color: #5a6480;
  margin-top: 5px;
}
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
  transition: opacity 0.2s, transform 0.1s;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 16px;
  animation: slideUp 0.4s 0.4s ease both;
}
.submit-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.submit-btn:hover:not(:disabled) { opacity: 0.85; }
.submit-btn:active:not(:disabled) { transform: scale(0.98); }
.footer {
  text-align: center;
  font-size: 12px;
  color: #5a6480;
}
.spinner {
  display: inline-block;
  width: 16px; height: 16px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
@keyframes fadeUp {
  from { opacity: 0; transform: translateY(32px) scale(0.97); }
  to   { opacity: 1; transform: translateY(0) scale(1); }
}
@keyframes slideUp {
  from { opacity: 0; transform: translateY(30px); }
  to   { opacity: 1; transform: translateY(0); }
}
</style>