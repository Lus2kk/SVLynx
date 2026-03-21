<template>
  <div class="logo-row">
    <div class="logo-img-wrap">
    <img src="../assets/logo_clean.png" alt="SVLynx" class="logo-img" />
    </div>
    <div class="logo-text">SV<span>Lynx</span></div>
  </div>
</template>

<style scoped>
.logo-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 36px;
}
.logo-img-wrap {
  width: 48px; height: 48px;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(79,142,247,0.35);
  animation: fadeUp 0.7s cubic-bezier(0.16,1,0.3,1) both;
}
.logo-img {
  width: 100%; height: 100%;
  object-fit: cover;
  display: block;
}
.logo-text {
  font-family: 'Syne', sans-serif;
  font-size: 22px;
  font-weight: 800;
  color: #fff;
  letter-spacing: -0.5px;
  animation: fadeUp 0.7s 0.1s cubic-bezier(0.16,1,0.3,1) both;
}
.logo-text span { color: #4f8ef7; }

@keyframes fadeUp {
  from { opacity: 0; transform: translateY(16px); }
  to   { opacity: 1; transform: translateY(0); }
}
</style>