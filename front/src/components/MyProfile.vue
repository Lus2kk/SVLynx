<template>
  <div class="mp-root">
    <div class="mp-bg">
      <div class="mp-bg-orb orb-1"></div>
      <div class="mp-bg-orb orb-2"></div>
    </div>

    <div class="mp-panel">

      <div class="mp-header">
        <button class="mp-back" @click="$emit('close')">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.2">
            <path d="M19 12H5M12 5l-7 7 7 7"/>
          </svg>
          <span>Назад</span>
        </button>
        <div class="mp-header-label">Профиль</div>
      </div>

      <div class="mp-avatar-block">
        <div class="mp-avatar" :style="avatarStyle">
          <img v-if="user.photo_url" :src="user.photo_url" alt="" class="mp-avatar-img" />
          <span v-else class="mp-avatar-letter">{{ avatarLetter }}</span>
        </div>
        <div class="mp-avatar-info">
          <div class="mp-display-name">{{ user.name || user.first_name || user.nickname || '—' }}</div>
          <div class="mp-nickname" v-if="user.nickname">@{{ user.nickname }}</div>
          <div class="mp-nickname" v-else-if="user.username">@{{ user.username }}</div>
          <div class="mp-status-text" v-if="user.status">{{ user.status }}</div>
        </div>
      </div>

      <div class="mp-chips" v-if="!loadingProfile">
        <div class="mp-chip" v-if="user.email">
          <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="2" y="4" width="20" height="16" rx="2"/><path d="M2 7l10 7 10-7"/>
          </svg>
          {{ user.email }}
        </div>
        <div class="mp-chip" v-if="user.telegram_id">
          <svg viewBox="0 0 24 24" width="12" height="12" fill="currentColor">
            <path d="M12 0C5.373 0 0 5.373 0 12s5.373 12 12 12 12-5.373 12-12S18.627 0 12 0zm5.562 8.248l-2.04 9.613c-.15.666-.54.832-1.094.517l-3-2.21-1.447 1.393c-.16.16-.295.295-.605.295l.215-3.053 5.56-5.023c.242-.215-.053-.334-.373-.12L7.19 14.6l-2.95-.92c-.641-.2-.654-.641.134-.949l11.532-4.448c.534-.194 1.001.13.656.965z"/>
          </svg>
          Telegram
        </div>
        <div class="mp-chip mp-chip-online" v-if="user.is_online">
          <span class="mp-online-dot"></span>
          Онлайн
        </div>
      </div>

      <div v-if="loadingProfile" class="mp-loading">
        <div class="mp-spinner"></div>
        <span>Загрузка...</span>
      </div>

      <div class="mp-divider" v-if="!loadingProfile">
        <span>{{ isTelegram ? 'Информация' : 'Редактировать' }}</span>
      </div>

      <!-- Telegram: readonly -->
      <div class="mp-tg-note" v-if="!loadingProfile && isTelegram">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
          <path d="M12 0C5.373 0 0 5.373 0 12s5.373 12 12 12 12-5.373 12-12S18.627 0 12 0zm5.562 8.248l-2.04 9.613c-.15.666-.54.832-1.094.517l-3-2.21-1.447 1.393c-.16.16-.295.295-.605.295l.215-3.053 5.56-5.023c.242-.215-.053-.334-.373-.12L7.19 14.6l-2.95-.92c-.641-.2-.654-.641.134-.949l11.532-4.448c.534-.194 1.001.13.656.965z"/>
        </svg>
        Профиль управляется через Telegram
      </div>

      <!-- Email: editable form -->
      <div class="mp-form" v-if="!loadingProfile && !isTelegram">
        <StatusMsg :type="status.type" :message="status.message" />

        <div class="mp-field">
          <label class="mp-label">Никнейм <span class="mp-req">*</span></label>
          <div class="mp-input-wrap">
            <input
              v-model="nickname"
              type="text"
              class="mp-input"
              placeholder="your_nickname"
              maxlength="25"
              spellcheck="false"
            />
            <span class="mp-counter" :class="{ warn: nickname.length >= 23 }">{{ nickname.length }}/25</span>
          </div>
          <div class="mp-hint">От 3 до 25 символов</div>
        </div>

        <div class="mp-field">
          <label class="mp-label">Имя</label>
          <input v-model="name" type="text" class="mp-input" placeholder="Ваше имя" maxlength="50" />
        </div>

        <div class="mp-field">
          <label class="mp-label">Статус</label>
          <input v-model="statusText" type="text" class="mp-input" placeholder="Расскажите немного о себе" maxlength="100" />
        </div>

        <button class="mp-save-btn" @click="save" :disabled="loading || nickname.length < 3">
          <span v-if="!loading">Сохранить изменения</span>
          <span v-else class="mp-btn-spinner"></span>
        </button>
      </div>

    </div>
  </div>
</template>

<script>
import StatusMsg from './StatusMsg.vue'
import { apiFetch } from '../api.js'
import { translateError } from '../composables/ErrorMessages.js'

const BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

const AVATAR_COLORS = [
  'linear-gradient(135deg, #6572ff, #8a67ff)',
  'linear-gradient(135deg, #f7604f, #ff9a5c)',
  'linear-gradient(135deg, #34d399, #059669)',
  'linear-gradient(135deg, #f472b6, #a855f7)',
  'linear-gradient(135deg, #38bdf8, #6366f1)',
]

export default {
  components: { StatusMsg },
  emits: ['close'],

  data() {
    return {
      user: {},
      nickname: '',
      name: '',
      statusText: '',
      isTelegram: false,
      loading: false,
      loadingProfile: true,
      status: { type: '', message: '' }
    }
  },

  computed: {
    avatarLetter() {
      const src = this.user.name || this.user.first_name || this.user.nickname || '?'
      return src[0].toUpperCase()
    },
    avatarStyle() {
      if (this.user.photo_url) return {}
      const color = this.user.avatar_color || AVATAR_COLORS[0]
      return { background: color }
    }
  },

  async mounted() {
    try {
      const res = await apiFetch(`${BASE}/auth/me`)
      if (!res.ok) throw new Error()
      const data = await res.json()
      this.user = data
      this.isTelegram = !!data.telegram_id
      this.nickname = data.nickname || ''
      this.name = data.name || ''
      this.statusText = data.status || ''
    } catch {
      this.status = { type: 'error', message: 'Не удалось загрузить профиль' }
    } finally {
      this.loadingProfile = false
    }
  },

  methods: {
    async save() {
      if (this.nickname.length < 3) return
      this.loading = true
      this.status = { type: '', message: '' }
      try {
        const res = await apiFetch(`${BASE}/auth/email/complete`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ nickname: this.nickname, name: this.name, status: this.statusText })
        })
        const data = await res.json()
        if (!res.ok) throw new Error(data.error || 'Ошибка')
        this.user.nickname = this.nickname
        this.user.name = this.name
        this.user.status = this.statusText
        this.status = { type: 'success', message: 'Профиль обновлён' }
      } catch (e) {
        this.status = { type: 'error', message: translateError(e.message) }
      } finally {
        this.loading = false
      }
    }
  }
}
</script>

<style scoped>
.mp-root {
  position: relative;
  width: 100%;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px 16px;
  box-sizing: border-box;
  overflow: hidden;
}

.mp-bg {
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 0;
}

.mp-bg-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.18;
}

.orb-1 {
  width: 500px; height: 500px;
  background: #6572ff;
  top: -120px; left: -100px;
  animation: orbFloat 8s ease-in-out infinite alternate;
}

.orb-2 {
  width: 400px; height: 400px;
  background: #a855f7;
  bottom: -100px; right: -80px;
  animation: orbFloat 10s ease-in-out infinite alternate-reverse;
}

@keyframes orbFloat {
  from { transform: translate(0, 0) scale(1); }
  to { transform: translate(30px, 20px) scale(1.05); }
}

.mp-panel {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 480px;
  background: rgba(9, 13, 30, 0.85);
  backdrop-filter: blur(24px);
  border: 1px solid rgba(255, 255, 255, 0.07);
  border-radius: 28px;
  padding: 28px;
  box-shadow: 0 32px 80px rgba(0, 0, 0, 0.5), inset 0 1px 0 rgba(255,255,255,0.06);
  animation: panelIn 0.5s cubic-bezier(0.16,1,0.3,1) both;
}

@keyframes panelIn {
  from { opacity: 0; transform: translateY(24px) scale(0.97); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}

.mp-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 28px;
}

.mp-back {
  display: flex;
  align-items: center;
  gap: 6px;
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.07);
  border-radius: 10px;
  padding: 7px 12px;
  color: #8090b8;
  font-size: 12px;
  font-family: 'DM Sans', sans-serif;
  cursor: pointer;
  transition: all 0.18s;
}
.mp-back:hover { background: rgba(255,255,255,0.08); color: #e0e6ff; }

.mp-header-label {
  font-family: 'Syne', sans-serif;
  font-size: 13px;
  font-weight: 700;
  color: #4a5580;
  text-transform: uppercase;
  letter-spacing: 2px;
}

.mp-avatar-block {
  display: flex;
  align-items: center;
  gap: 18px;
  margin-bottom: 20px;
  padding: 20px;
  background: rgba(255,255,255,0.025);
  border: 1px solid rgba(255,255,255,0.05);
  border-radius: 20px;
}

.mp-avatar {
  width: 72px; height: 72px;
  border-radius: 50%;
  flex-shrink: 0;
  display: grid;
  place-items: center;
  overflow: hidden;
  box-shadow: 0 8px 24px rgba(0,0,0,0.35), 0 0 0 3px rgba(255,255,255,0.07);
}

.mp-avatar-img { width: 100%; height: 100%; object-fit: cover; }

.mp-avatar-letter {
  font-family: 'Syne', sans-serif;
  font-size: 28px;
  font-weight: 800;
  color: #fff;
}

.mp-avatar-info { flex: 1; min-width: 0; }

.mp-display-name {
  font-family: 'Syne', sans-serif;
  font-size: 20px;
  font-weight: 700;
  color: #eef2ff;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-bottom: 3px;
}

.mp-nickname {
  font-size: 13px;
  color: #6572ff;
  font-weight: 500;
  margin-bottom: 4px;
}

.mp-status-text {
  font-size: 12px;
  color: #5a6480;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.mp-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 22px;
}

.mp-chip {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 5px 10px;
  border-radius: 8px;
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.07);
  color: #6d7ba8;
  font-size: 11.5px;
  font-weight: 500;
}

.mp-chip-online {
  border-color: rgba(52, 211, 153, 0.25);
  background: rgba(52, 211, 153, 0.08);
  color: #34d399;
}

.mp-online-dot {
  width: 6px; height: 6px;
  border-radius: 50%;
  background: #34d399;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.5; transform: scale(0.8); }
}

.mp-loading {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #5a6480;
  font-size: 13px;
  margin-bottom: 24px;
  padding: 16px;
}

.mp-spinner {
  width: 18px; height: 18px;
  border: 2px solid rgba(101, 114, 255, 0.2);
  border-top-color: #6572ff;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

.mp-divider {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
}

.mp-divider::before,
.mp-divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: rgba(255,255,255,0.05);
}

.mp-divider span {
  font-size: 10px;
  font-weight: 700;
  color: #3a4260;
  text-transform: uppercase;
  letter-spacing: 1.5px;
  white-space: nowrap;
}

.mp-tg-note {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 16px;
  border-radius: 12px;
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.06);
  color: #4a5580;
  font-size: 13px;
}

.mp-form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.mp-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.mp-label {
  font-size: 10.5px;
  font-weight: 700;
  color: #4a5a88;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.mp-req { color: #6572ff; }

.mp-input-wrap { position: relative; }

.mp-input {
  width: 100%;
  box-sizing: border-box;
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.07);
  border-radius: 12px;
  padding: 12px 16px;
  color: #e8eeff;
  font-size: 14px;
  font-family: 'DM Sans', sans-serif;
  outline: none;
  transition: border-color 0.2s, box-shadow 0.2s, background 0.2s;
  min-height: 46px;
}

.mp-input-wrap .mp-input { padding-right: 52px; }
.mp-input::placeholder { color: #2e3555; }
.mp-input:focus {
  border-color: rgba(101, 114, 255, 0.45);
  background: rgba(101, 114, 255, 0.05);
  box-shadow: 0 0 0 3px rgba(101, 114, 255, 0.1);
}

.mp-counter {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 10px;
  color: #3a4260;
  pointer-events: none;
}
.mp-counter.warn { color: #f87171; }

.mp-hint { font-size: 11px; color: #3a4260; }

.mp-save-btn {
  width: 100%;
  margin-top: 6px;
  background: linear-gradient(135deg, #6572ff 0%, #a855f7 100%);
  border: none;
  border-radius: 14px;
  padding: 14px;
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  font-family: 'DM Sans', sans-serif;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 48px;
  transition: opacity 0.2s, transform 0.15s, box-shadow 0.2s;
  box-shadow: 0 8px 24px rgba(101, 114, 255, 0.3);
  letter-spacing: 0.3px;
}
.mp-save-btn:hover:not(:disabled) {
  opacity: 0.92;
  transform: translateY(-1px);
  box-shadow: 0 12px 32px rgba(101, 114, 255, 0.4);
}
.mp-save-btn:active:not(:disabled) { transform: scale(0.98); }
.mp-save-btn:disabled { opacity: 0.35; cursor: not-allowed; box-shadow: none; }

.mp-btn-spinner {
  display: inline-block;
  width: 16px; height: 16px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }

@media (max-width: 600px) {
  .mp-root {
    align-items: flex-start;
    padding: 0;
  }

  .mp-panel {
    max-width: 100%;
    min-height: 100svh;
    border-radius: 0;
    border: none;
    box-shadow: none;
    padding: 20px;
    padding-top: calc(20px + env(safe-area-inset-top));
    padding-bottom: calc(20px + env(safe-area-inset-bottom));
    backdrop-filter: none;
    background: rgba(8, 12, 26, 0.99);
  }

  .mp-avatar { width: 60px; height: 60px; }
  .mp-avatar-letter { font-size: 22px; }
  .mp-display-name { font-size: 17px; }
}
</style>