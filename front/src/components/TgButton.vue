<template>
  <div class="tg-wrapper" ref="wrapper"></div>
</template>
<script>
export default {
  emits: ['auth', 'error'],
  mounted() {
    window.onTelegramAuth = async (user) => {
      console.log('Telegram user data:', JSON.stringify(user))
      console.log('sender_name будет:', user.first_name || user.username || '')

      try {
        const response = await fetch(`${import.meta.env.VITE_API_URL}/auth/telegram/callback`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(user)
        })
        const data = await response.json()
        if (!response.ok) {
          throw new Error(data.error || 'Telegram auth failed')
        }
        sessionStorage.setItem('access_token', data.access_token)
        sessionStorage.setItem('refresh_token', data.refresh_token)
        const name = user.first_name || user.username || ''
        sessionStorage.setItem('current_user_name', name)
        this.$emit('auth', { ...data, sender_name: user.first_name || user.username || '' })
      } catch (err) {
        console.error('Telegram auth failed', err)
        this.$emit('error', err)
      }
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
  },
  beforeUnmount() {
    delete window.onTelegramAuth
  }
}
</script>
<style scoped>
.tg-wrapper {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  margin-bottom: 32px;
}
.tg-wrapper :deep(iframe),
.tg-wrapper :deep(span),
.tg-wrapper :deep(a) {
  display: block;
  margin: 0 auto;
  max-width: 100%;
}
</style>