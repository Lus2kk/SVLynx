<template>
  <transition name="modal-fade">
    <div class="modal-overlay" @click.self="$emit('close')">
      <div class="modal" :class="{ 'theme-light': isLight }">

        <!-- Header -->
        <div class="modal-header">
          <div class="modal-icon">
            <!-- Иконка канала (такая же как в sidebar) -->
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
              <path d="M4 10h3l6-4v12l-6-4H4z" />
              <path d="M7 14.5v3a2 2 0 0 0 2 2h1" />
              <path d="M18 9a3 3 0 0 1 0 6" />
              <path d="M20.5 7.5a5 5 0 0 1 0 9" />
            </svg>
          </div>
          <div>
            <h2 class="modal-title">Create Channel</h2>
            <p class="modal-subtitle">Broadcast posts to your subscribers</p>
          </div>
          <button class="close-btn" @click="$emit('close')">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 6L6 18M6 6l12 12"/>
            </svg>
          </button>
        </div>

        <!-- Avatar picker -->
        <div class="avatar-section">
          <div class="avatar-preview" :style="{ background: selectedColor }">
            <span>{{ form.name?.[0]?.toUpperCase() || '#' }}</span>
          </div>
          <div class="color-picker">
            <button
              v-for="color in colors"
              :key="color"
              class="color-dot"
              :class="{ active: selectedColor === color }"
              :style="{ background: color }"
              @click="selectedColor = color"
            />
          </div>
        </div>

        <!-- Form -->
        <div class="form">
          <div class="field">
            <label>Channel Name <span class="required">*</span></label>
            <input
              v-model="form.name"
              type="text"
              placeholder="e.g. Daily Tech Digest"
              maxlength="100"
              @input="autoHandle"
            />
          </div>

          <div class="field">
            <label>
              Channel Username
              <span class="required">*</span>
              <span class="label-hint">— used to find and mention your channel</span>
            </label>
            <div class="handle-input">
              <span class="handle-prefix">@</span>
              <input
                v-model="form.handle"
                type="text"
                placeholder="daily_tech_digest"
                maxlength="32"
                @input="sanitizeHandle"
              />
            </div>
            <span class="field-hint">3–32 characters: letters, digits, underscores only</span>
          </div>

          <div class="field">
            <label>Description</label>
            <textarea
              v-model="form.description"
              placeholder="What will you post about?"
              rows="3"
              maxlength="500"
            ></textarea>
          </div>

          <div class="field">
            <label>Type</label>
            <div class="type-selector">
              <button
                class="type-btn"
                :class="{ active: form.type === 'public' }"
                @click="form.type = 'public'"
                type="button"
              >
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.8">
                  <circle cx="12" cy="12" r="10"/>
                  <path d="M2 12h20M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/>
                </svg>
                Public
                <span class="type-hint">Anyone can find & join</span>
              </button>
              <button
                class="type-btn"
                :class="{ active: form.type === 'private' }"
                @click="form.type = 'private'"
                type="button"
              >
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.8">
                  <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
                  <path d="M7 11V7a5 5 0 0 1 10 0v4"/>
                </svg>
                Private
                <span class="type-hint">Invite link only</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Error -->
        <div v-if="error" class="error-msg">{{ error }}</div>

        <!-- Actions -->
        <div class="modal-footer">
          <button class="btn-cancel" @click="$emit('close')" type="button">Cancel</button>
          <button
            class="btn-create"
            :disabled="!canSubmit || loading"
            @click="submit"
            type="button"
          >
            <span v-if="loading" class="spinner"></span>
            <span v-else>Create Channel</span>
          </button>
        </div>

      </div>
    </div>
  </transition>
</template>

<script>
import { apiFetch } from '../api.js'

const BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

export default {
  name: 'CreateChannelModal',

  props: {
    currentUserId: { type: String, required: true },
    isLight: { type: Boolean, default: false }
  },

  emits: ['close', 'created'],

  data() {
    return {
      form: {
        name: '',
        handle: '',
        description: '',
        type: 'public'
      },
      selectedColor: 'linear-gradient(135deg, #6572ff, #8a67ff)',
      colors: [
        'linear-gradient(135deg, #6572ff, #8a67ff)',
        'linear-gradient(135deg, #1D9E75, #16c77e)',
        'linear-gradient(135deg, #D85A30, #f0703a)',
        'linear-gradient(135deg, #378ADD, #4a9ef5)',
        'linear-gradient(135deg, #b51ed7, #d42af5)',
        'linear-gradient(135deg, #e0a020, #f5b830)',
        'linear-gradient(135deg, #e0205a, #f53070)',
        'linear-gradient(135deg, #20c0e0, #30d5f5)',
      ],
      loading: false,
      error: ''
    }
  },

  computed: {
    canSubmit() {
      return this.form.name.trim().length >= 1 && this.form.handle.length >= 3
    }
  },

  methods: {
    autoHandle() {
      if (!this.form.handle || this.form.handle === this.prevAutoHandle) {
        const h = this.form.name.toLowerCase()
          .replace(/\s+/g, '_')
          .replace(/[^a-z0-9_]/g, '')
          .slice(0, 32)
        this.form.handle = h
        this.prevAutoHandle = h
      }
    },

    sanitizeHandle() {
      this.form.handle = this.form.handle
        .toLowerCase()
        .replace(/[^a-z0-9_]/g, '')
        .slice(0, 32)
      this.prevAutoHandle = null
    },

    async submit() {
      if (!this.canSubmit || this.loading) return
      this.error = ''
      this.loading = true

      try {
        const res = await apiFetch(`${BASE}/channels`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            name: this.form.name.trim(),
            handle: this.form.handle,
            description: this.form.description.trim(),
            type: this.form.type,
            avatar_color: this.selectedColor,
            owner_id: this.currentUserId
          })
        })

        const data = await res.json()

        if (!res.ok) {
          this.error = data.error || 'Failed to create channel'
          return
        }

        this.$emit('created', { ...data.channel, user_role: 'owner' })
        this.$emit('close')
      } catch (e) {
        this.error = 'Network error'
        console.error('createChannel error', e)
      } finally {
        this.loading = false
      }
    }
  }
}
</script>

<style scoped>
.modal-overlay {
  position: fixed; inset: 0;
  background: rgba(4, 6, 16, 0.7);
  backdrop-filter: blur(8px);
  display: grid; place-items: center;
  z-index: 200; padding: 16px;
}

.modal {
  width: 100%; max-width: 420px;
  background: linear-gradient(180deg, rgba(18, 22, 44, 0.98), rgba(12, 16, 32, 0.99));
  border: 1px solid rgba(132, 144, 224, 0.14);
  border-radius: 22px;
  box-shadow: 0 32px 64px rgba(0, 0, 0, 0.5), 0 0 0 1px rgba(255,255,255,0.03) inset;
  overflow: hidden;
}
.modal.theme-light {
  background: #ffffff;
  border-color: rgba(91, 106, 255, 0.15);
  box-shadow: 0 24px 48px rgba(91, 106, 200, 0.15);
}

.modal-header {
  display: flex; align-items: center; gap: 14px;
  padding: 22px 22px 18px;
  border-bottom: 1px solid rgba(255,255,255,0.05);
}
.theme-light .modal-header { border-bottom-color: #f0f1f8; }

.modal-icon {
  width: 44px; height: 44px; border-radius: 13px; flex-shrink: 0;
  display: grid; place-items: center;
  color: #6e79ff;
  background: rgba(110, 121, 255, 0.12);
  border: 1px solid rgba(110, 121, 255, 0.2);
}

.modal-title { color: #eef2ff; font-size: 16px; font-weight: 700; margin: 0 0 2px; }
.theme-light .modal-title { color: #1a1d2e; }
.modal-subtitle { color: #7d87ab; font-size: 12px; margin: 0; }

.close-btn {
  margin-left: auto; width: 30px; height: 30px; border-radius: 9px;
  display: grid; place-items: center; flex-shrink: 0;
  color: #7d87ab; background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.06); cursor: pointer; transition: all 0.15s;
}
.close-btn:hover { color: #ff4d6d; background: rgba(255,77,109,0.1); border-color: rgba(255,77,109,0.2); }

/* Avatar */
.avatar-section { display: flex; align-items: center; gap: 16px; padding: 18px 22px; }
.avatar-preview {
  width: 56px; height: 56px; border-radius: 16px; flex-shrink: 0;
  display: grid; place-items: center;
  color: #fff; font-size: 22px; font-weight: 800;
  box-shadow: 0 8px 20px rgba(0,0,0,0.25); transition: background 0.2s;
}
.color-picker { display: flex; flex-wrap: wrap; gap: 8px; }
.color-dot {
  width: 22px; height: 22px; border-radius: 50%;
  cursor: pointer; border: 2px solid transparent; transition: transform 0.15s, border-color 0.15s;
}
.color-dot:hover { transform: scale(1.15); }
.color-dot.active { border-color: #fff; transform: scale(1.15); }

/* Form */
.form { padding: 0 22px; display: flex; flex-direction: column; gap: 14px; }
.field { display: flex; flex-direction: column; gap: 6px; }
.field label {
  color: #7d87ab; font-size: 11px; font-weight: 700;
  text-transform: uppercase; letter-spacing: 0.05em;
  display: flex; align-items: center; gap: 4px; flex-wrap: wrap;
}
.theme-light .field label { color: #9098b8; }
.label-hint { color: #5d6888; font-size: 10px; font-weight: 500; text-transform: none; letter-spacing: 0; }
.required { color: #ff4d6d; }

.field input, .field textarea {
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 11px; padding: 10px 13px;
  color: #eef2ff; font-size: 13px; font-weight: 500;
  outline: none; font-family: inherit; transition: border-color 0.15s;
  box-sizing: border-box; width: 100%;
}
.theme-light .field input, .theme-light .field textarea { background: #f5f6fc; border-color: #e4e6f0; color: #1a1d2e; }
.field input:focus, .field textarea:focus { border-color: rgba(110,121,255,0.4); }
.field textarea { resize: none; line-height: 1.5; }

.handle-input {
  display: flex; align-items: center;
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 11px; overflow: hidden; transition: border-color 0.15s;
}
.theme-light .handle-input { background: #f5f6fc; border-color: #e4e6f0; }
.handle-input:focus-within { border-color: rgba(110,121,255,0.4); }
.handle-prefix { padding: 10px 4px 10px 13px; color: #6e79ff; font-size: 13px; font-weight: 700; flex-shrink: 0; }
.handle-input input { background: transparent; border: none; border-radius: 0; padding: 10px 13px 10px 0; flex: 1; }
.handle-input input:focus { border-color: transparent; }
.field-hint { color: #5d6888; font-size: 11px; margin-top: 2px; }

/* Type selector */
.type-selector { display: flex; gap: 8px; }
.type-btn {
  flex: 1; display: flex; flex-direction: column; align-items: flex-start; gap: 2px;
  padding: 10px 12px; border-radius: 11px; cursor: pointer; text-align: left;
  color: #7d87ab; font-size: 12px; font-weight: 600;
  background: rgba(255,255,255,0.03); border: 1px solid rgba(255,255,255,0.07); transition: all 0.15s;
}
.theme-light .type-btn { background: #f5f6fc; border-color: #e4e6f0; color: #9098b8; }
.type-btn svg { margin-bottom: 4px; }
.type-btn.active { color: #6e79ff; background: rgba(110,121,255,0.1); border-color: rgba(110,121,255,0.3); }
.type-hint { font-size: 10px; font-weight: 500; opacity: 0.7; }

/* Error */
.error-msg {
  margin: 8px 22px 0; padding: 10px 13px; border-radius: 10px;
  background: rgba(255,77,109,0.1); border: 1px solid rgba(255,77,109,0.2);
  color: #ff4d6d; font-size: 12px; font-weight: 500;
}

/* Footer */
.modal-footer { display: flex; gap: 10px; justify-content: flex-end; padding: 18px 22px 22px; }
.btn-cancel {
  padding: 10px 18px; border-radius: 11px;
  background: rgba(255,255,255,0.05); border: 1px solid rgba(255,255,255,0.07);
  color: #a6afd4; font-size: 13px; font-weight: 600; cursor: pointer; transition: all 0.15s;
}
.theme-light .btn-cancel { background: #f3f4f8; border-color: #e2e4ee; color: #7880a0; }
.btn-cancel:hover { background: rgba(255,255,255,0.08); }
.btn-create {
  padding: 10px 22px; border-radius: 11px; border: none;
  background: linear-gradient(135deg, #6e79ff, #8669ff);
  color: #fff; font-size: 13px; font-weight: 700; cursor: pointer;
  box-shadow: 0 8px 20px rgba(94,102,255,0.3); transition: all 0.15s;
  display: flex; align-items: center; gap: 8px; min-width: 130px; justify-content: center;
}
.btn-create:disabled { opacity: 0.45; cursor: not-allowed; box-shadow: none; }
.btn-create:not(:disabled):hover { transform: translateY(-1px); box-shadow: 0 12px 24px rgba(94,102,255,0.4); }

.spinner {
  width: 14px; height: 14px; border-radius: 50%;
  border: 2px solid rgba(255,255,255,0.3); border-top-color: #fff;
  animation: spin 0.7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

.modal-fade-enter-active, .modal-fade-leave-active { transition: opacity 0.2s, transform 0.2s; }
.modal-fade-enter-from, .modal-fade-leave-to { opacity: 0; transform: scale(0.96) translateY(8px); }

@media (max-width: 760px) {
  .modal-overlay { align-items: flex-end; padding: 0; }
  .modal { border-radius: 22px 22px 0 0; max-width: 100%; }
  .modal-fade-enter-from, .modal-fade-leave-to { transform: translateY(20px); }
}
</style>