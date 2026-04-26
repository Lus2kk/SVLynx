<template>
  <div class="input-area">
    <textarea
      v-model="text"
      class="msg-input"
      placeholder="Написать сообщение..."
      rows="1"
      :disabled="disabled"
      @keydown.enter.exact.prevent="submit"
      @input="autoResize"
      ref="textarea"
    ></textarea>
    <button
      class="send-btn"
      :disabled="!text.trim() || disabled"
      @click="submit"
    >
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
        <line x1="22" y1="2" x2="11" y2="13"></line>
        <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
      </svg>
    </button>
  </div>
</template>

<script>
export default {
  props: {
    disabled: { type: Boolean, default: false }
  },
  emits: ['send'],

  data() {
    return { text: '' }
  },

  methods: {
    submit() {
      if (!this.text.trim() || this.disabled) return
      this.$emit('send', this.text.trim())
      this.text = ''
      this.$nextTick(() => {
        if (this.$refs.textarea) {
          this.$refs.textarea.style.height = 'auto'
        }
      })
    },

    autoResize(e) {
      const el = e.target
      el.style.height = 'auto'
      el.style.height = Math.min(el.scrollHeight, 120) + 'px'
    }
  }
}
</script>

<style scoped>
.input-area {
  display: flex;
  align-items: flex-end;
  gap: 10px;
  padding: 16px 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(13, 19, 32, 0.8);
  backdrop-filter: blur(12px);
  flex-shrink: 0;
}

.msg-input {
  flex: 1;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(79, 142, 247, 0.15);
  border-radius: 14px;
  padding: 12px 16px;
  color: #f0f2f8;
  font-size: 14px;
  font-family: 'DM Sans', sans-serif;
  outline: none;
  resize: none;
  line-height: 1.5;
  max-height: 120px;
  overflow-y: auto;
  transition: border-color 0.2s, box-shadow 0.2s;
}
.msg-input::placeholder { color: #3a4060; }
.msg-input:focus:not(:disabled) {
  border-color: rgba(79, 142, 247, 0.45);
  box-shadow: 0 0 0 3px rgba(79, 142, 247, 0.08);
}
.msg-input:disabled { opacity: 0.4; cursor: not-allowed; }

.send-btn {
  width: 44px; height: 44px;
  border-radius: 12px;
  background: linear-gradient(135deg, #4f8ef7, #7c5ef7);
  border: none;
  color: #fff;
  cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
  transition: opacity 0.2s, transform 0.1s, box-shadow 0.2s;
}
.send-btn:disabled { opacity: 0.3; cursor: not-allowed; }
.send-btn:hover:not(:disabled) {
  opacity: 0.9;
  box-shadow: 0 6px 20px rgba(79, 142, 247, 0.4);
  transform: translateY(-1px);
}
.send-btn:active:not(:disabled) { transform: scale(0.95); }
</style>