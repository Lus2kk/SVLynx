function getCookie(name) {
  const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'))
  return match ? match[2] : null
}

function setCookie(name, value, maxAgeSeconds) {
  document.cookie = `${name}=${value}; path=/; max-age=${maxAgeSeconds}; SameSite=Strict`
}

function deleteCookie(name) {
  document.cookie = `${name}=; path=/; max-age=0`
}

const BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

let isRefreshing = false
let refreshQueue = []

async function refreshAccessToken() {
  const refreshToken = getCookie('refresh_token')
  if (!refreshToken) throw new Error('No refresh token')

  const res = await fetch(`${BASE}/auth/refresh`, {
    method: 'POST',
    headers: {
      'X-Refresh-Token': refreshToken
    }
  })

  if (!res.ok) {
    deleteCookie('access_token')
    deleteCookie('refresh_token')
    window.location.reload()
    throw new Error('Refresh failed')
  }

  const data = await res.json()
  setCookie('access_token', data.access_token, 900)
  if (data.refresh_token) {
    setCookie('refresh_token', data.refresh_token, 2592000)
  }

  return data.access_token
}

export async function apiFetch(url, options = {}) {
  let token = getCookie('access_token')

  if (!token) {
    try {
      token = await refreshAccessToken()
    } catch (e) {
      return new Response(null, { status: 401 })
    }
  }

  const res = await fetch(url, {
    ...options,
    headers: {
      ...options.headers,
      Authorization: `Bearer ${token}`
    }
  })


  if (res.status === 401) {
    if (isRefreshing) {
      return new Promise((resolve, reject) => {
        refreshQueue.push({ resolve, reject, url, options })
      })
    }

    isRefreshing = true

    try {
      const newToken = await refreshAccessToken()
      isRefreshing = false

      refreshQueue.forEach(({ resolve, reject, url, options }) => {
        apiFetch(url, options).then(resolve).catch(reject)
      })
      refreshQueue = []

      return fetch(url, {
        ...options,
        headers: {
          ...options.headers,
          Authorization: `Bearer ${newToken}`
        }
      })
    } catch (e) {
      isRefreshing = false
      refreshQueue.forEach(({ reject }) => reject(e))
      refreshQueue = []
      throw e
    }
  }

  return res
}

export { getCookie, setCookie, deleteCookie }