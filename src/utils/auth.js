export function getToken() {
  return localStorage.getItem('token')
}

export function setToken(token) {
  localStorage.setItem('token', token)
}

export function removeToken() {
  localStorage.removeItem('token')
}

export function getUserInfo() {
  const userStr = localStorage.getItem('user')
  return userStr ? JSON.parse(userStr) : null
}

export function setUserInfo(user) {
  localStorage.setItem('user', JSON.stringify(user))
}

export function removeUserInfo() {
  localStorage.removeItem('user')
}
