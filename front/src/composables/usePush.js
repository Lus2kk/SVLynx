const VAPID_PUBLIC_KEY = 'BDxJcOA4xjBkROFCF1-DO9YpUzWVaYR2Ex9i8yPZm8gwo5YVUhsSHp6iuS70EiMRCL4r3WDFrMbAJh_GMw5tA-4'
const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

function urlBase64ToUint8Array(base64String) {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4)
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/')
  const rawData = atob(base64)
  return Uint8Array.from([...rawData].map((c) => c.charCodeAt(0)))
}

export function usePush() {
  async function subscribe() {
    if (!('serviceWorker' in navigator) || !('PushManager' in window)) {
      console.warn('Push не поддерживается')
      return
    }

    const permission = await Notification.requestPermission()
    if (permission !== 'granted') {
      console.warn('Разрешение на уведомления отклонено')
      return
    }

    const reg = await navigator.serviceWorker.register('/sw.js')
    await navigator.serviceWorker.ready

    const sub = await reg.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: urlBase64ToUint8Array(VAPID_PUBLIC_KEY),
    })

    const key = sub.getKey('p256dh')
    const auth = sub.getKey('auth')

    const token = sessionStorage.getItem('access_token')

    await fetch(`${API_URL}/push/subscribe`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({
        endpoint: sub.endpoint,
        p256dh: btoa(String.fromCharCode(...new Uint8Array(key))),
        auth: btoa(String.fromCharCode(...new Uint8Array(auth))),
      }),
    })

    console.log('Push подписка оформлена')
  }

  async function unsubscribe() {
    const reg = await navigator.serviceWorker.getRegistration('/sw.js')
    if (!reg) return

    const sub = await reg.pushManager.getSubscription()
    if (!sub) return

    const token = sessionStorage.getItem('access_token')

    await fetch(`${API_URL}/push/unsubscribe`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ endpoint: sub.endpoint }),
    })

    await sub.unsubscribe()
    console.log('Push подписка отменена')
  }

  return { subscribe, unsubscribe }
}