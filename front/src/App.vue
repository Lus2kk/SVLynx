<template>
  <BgScene />
  <transition name="fade" mode="out-in">
    <ProfileSetup v-if="showProfile" key="profile" @done="onProfileDone" />
    <LoginCard v-else key="login" @show-profile="showProfile = true" />
  </transition>
</template>

<script>
import BgScene from './components/BgScene.vue'
import LoginCard from './components/LoginCard.vue'
import ProfileSetup from './components/ProfileSetup.vue'

export default {
  components: { BgScene, LoginCard, ProfileSetup },

  data() {
    return {
      showProfile: false
    }
  },

  methods: {
    onProfileDone() {
      this.showProfile = false
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
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: 'DM Sans', sans-serif;
  overflow: hidden;
}

.fade-enter-active, .fade-leave-active {
  transition: all 0.35s ease;
}
.fade-enter-from { opacity: 0; transform: translateY(16px) scale(0.98); }
.fade-leave-to { opacity: 0; transform: translateY(-16px) scale(0.98); }
</style>