<template>
  <div>
    <div class="bg-scene">
      <div class="orb orb-1"></div>
      <div class="orb orb-2"></div>
      <div class="orb orb-3"></div>
    </div>
    <div class="grid-overlay"></div>
  </div>
</template>

<style scoped>
.bg-scene {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
}

.orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.18;
  animation: drift 12s ease-in-out infinite alternate;
}
.orb-1 {
  width: 500px; height: 500px;
  background: radial-gradient(circle, #4f8ef7, transparent);
  top: -100px; left: -100px;
}
.orb-2 {
  width: 400px; height: 400px;
  background: radial-gradient(circle, #7c5ef7, transparent);
  bottom: -80px; right: -80px;
  animation-delay: -4s;
}
.orb-3 {
  width: 300px; height: 300px;
  background: radial-gradient(circle, #2dd4bf, transparent);
  top: 50%; left: 50%;
  transform: translate(-50%, -50%);
  animation-delay: -8s;
  opacity: 0.08;
}

@keyframes drift {
  from { transform: translate(0, 0) scale(1); }
  to   { transform: translate(30px, 20px) scale(1.1); }
}

.grid-overlay {
  position: fixed;
  inset: 0;
  background-image:
    linear-gradient(rgba(255,255,255,0.025) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255,255,255,0.025) 1px, transparent 1px);
  background-size: 48px 48px;
  z-index: 0;
}
</style>