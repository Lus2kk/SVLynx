<template>
  <div class="tg-wrapper" ref="wrapper"></div>
</template>

<script>
export default {
  emits: ['auth'],

  mounted() {
    window.onTelegramAuth = (user) => {
      this.$emit('auth', user)
    }

    const script = document.createElement('script')
    script.src = 'https://telegram.org/js/telegram-widget.js?22'
    script.setAttribute('data-telegram-login', 'svlynx_auth_bot')
    script.setAttribute('data-size', 'large')
    script.setAttribute('data-onauth', 'onTelegramAuth(user)')
    script.setAttribute('data-request-access', 'write')
    script.async = true
    script.onerror = () => {
      console.warn('Telegram widget failed to load')
    }
    this.$refs.wrapper.appendChild(script)
  }
}
</script>

<style scoped>
.tg-wrapper {
  display: flex;
  justify-content: center;
  margin-bottom: 32px;
  width: 100%;
}

.tg-wrapper :deep(iframe),
.tg-wrapper :deep(span) {
  width: 100% !important;
}

.tg-wrapper :deep(a) {
  width: 100% !important;
  display: block !important;
}
</style>
