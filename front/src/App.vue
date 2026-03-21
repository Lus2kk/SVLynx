<template>
  <BgScene />
  <LoginCard />
</template>

<script>
import BgScene from './components/BgScene.vue'
import LoginCard from './components/LoginCard.vue'

export default {
  components: { BgScene, LoginCard }
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
</style>