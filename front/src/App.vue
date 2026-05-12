<template>
  <BgScene />
  <transition name="fade" mode="out-in">
    <div v-if="showChat" key="chat" class="app-view">
     <ChatLayout @theme-changed="onThemeChanged" @open-profile="onOpenProfile" />
    </div>
    <div v-else-if="showMyProfile" key="my-profile" class="app-view">
      <MyProfile @close="showMyProfile = false; showChat = true" />
    </div>
    <div v-else-if="showProfile" key="profile" class="app-view">
      <ProfileSetup @done="onProfileDone" />
    </div>
    <div v-else key="login" class="app-view">
      <LoginCard @show-profile="showProfile = true" @show-chat="showChat = true" />
    </div>
  </transition>
</template>

<script>
import BgScene from './components/BgScene.vue'
import LoginCard from './components/LoginCard.vue'
import ProfileSetup from './components/ProfileSetup.vue'
import ChatLayout from './components/ChatLayout.vue'
import MyProfile from './components/MyProfile.vue'

export default {
  components: { BgScene, LoginCard, ProfileSetup, ChatLayout, MyProfile },

  data() {
    return {
      showProfile: false,
      showChat: false,
      showMyProfile: false,
    }
  },

  async mounted() {
    document.addEventListener('gesturestart', e => e.preventDefault())
    document.addEventListener('gesturechange', e => e.preventDefault())
    document.addEventListener('gestureend', e => e.preventDefault())

    this.applyTheme()

    const accessToken = this.getCookie('access_token')
    const refreshToken = this.getCookie('refresh_token')

    if (accessToken) {
      this.showChat = true
      return
    }

    if (refreshToken) {
      try {
        const BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'
        const res = await fetch(`${BASE}/auth/refresh`, {
          method: 'POST',
          headers: { 'X-Refresh-Token': refreshToken }
        })

        if (res.ok) {
          const data = await res.json()
          this.setCookie('access_token', data.access_token, 60)
          this.setCookie('refresh_token', data.refresh_token, 2592000)
          this.showChat = true
        }
      } catch (e) {
        console.error('refresh error:', e)
      }
    }
  },

  methods: {

    onOpenProfile() {
      this.showChat = false
      this.showMyProfile = true
    },

    getCookie(name) {
      const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'))
      return match ? match[2] : null
    },

    setCookie(name, value, maxAgeSeconds) {
      const secure = location.protocol === 'https:' ? '; Secure' : ''
      document.cookie = `${name}=${value}; path=/; max-age=${maxAgeSeconds}; SameSite=Strict${secure}`
    },

    applyTheme() {
      const isLight = localStorage.getItem('svlynx-theme') === 'light'
      const color = isLight ? '#ffffff' : 'rgb(8, 12, 26)'
      document.documentElement.style.background = color
      document.body.style.background = color
      const meta = document.querySelector('meta[name="theme-color"]')
      if (meta) meta.setAttribute('content', color)
    },

    onThemeChanged() {
      this.applyTheme()
    },

    onProfileDone() {
      this.showProfile = false
      this.showChat = true
    }
  }
}
</script>

<style>
@import url('https://fonts.googleapis.com/css2?family=Syne:wght@400;700;800&family=DM+Sans:wght@300;400;500&display=swap');
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
body {
  min-height: 100vh;
  background: #080c14;
  font-family: 'DM Sans', sans-serif;
  overflow-x: hidden; 
}

.app-view {
  width: 100%;
  min-height: 100dvh;
  display: flex;
  justify-content: center;
  align-items: center;
}
button:focus,
button:focus-visible {
  outline: none;
}

.fade-enter-active, .fade-leave-active { transition: all 0.35s ease; }
.fade-enter-from { opacity: 0; transform: translateY(16px) scale(0.98); }
.fade-leave-to { opacity: 0; transform: translateY(-16px) scale(0.98); }
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
</style>