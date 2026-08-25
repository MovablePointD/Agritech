<template>
  <header class="header">
    <div class="container">
      <router-link to="/" class="logo">农技综合服务平台</router-link>
      <nav class="nav">
        <router-link to="/" :class="{ active: $route.path === '/' }">首页</router-link>
        <router-link to="/affairs" :class="{ active: $route.path.startsWith('/affair') || $route.path.startsWith('/my-affairs') }">事务</router-link>
        <router-link to="/products" :class="{ active: $route.path.startsWith('/product') }">商品</router-link>
        <router-link to="/posts" :class="{ active: $route.path.startsWith('/post') }">动态</router-link>
        <router-link to="/knowledge" :class="{ active: $route.path.startsWith('/knowledge') || $route.path.startsWith('/my-knowledge') }">知识</router-link>
        <router-link to="/experts" :class="{ active: $route.path.startsWith('/expert') }">专家</router-link>
        <router-link to="/rural-info" :class="{ active: $route.path.startsWith('/rural-info') || $route.path.startsWith('/polic') }">农村信息</router-link>
        <router-link to="/ai-chat" :class="{ active: $route.path.startsWith('/ai-chat') }">
          <!-- <el-icon style="vertical-align: middle; margin-right: 2px;"><ChatDotRound /></el-icon> -->AI助手
        </router-link>
      </nav>
      <div class="user-area">
        <template v-if="token">
          <el-badge :value="msgUnreadCount" :hidden="msgUnreadCount === 0" :max="99">
            <el-button :icon="Message" circle @click="$router.push('/messages')" />
          </el-badge>
          <el-badge :value="unreadCount" :hidden="unreadCount === 0" :max="99">
            <el-button :icon="Bell" circle @click="$router.push('/notifications')" />
          </el-badge>
          <el-dropdown @command="handleUserCommand">
            <span class="user-info">
              <el-avatar :src="user?.avatar_url" :size="32">{{ user?.username?.[0] }}</el-avatar>
              <span>{{ user?.username }}</span>
              <el-tag :type="roleTagType" size="small" style="margin-left: 4px;">{{ roleText }}</el-tag>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="user">个人中心</el-dropdown-item> 
                <el-dropdown-item command="ai-chat">AI 助手</el-dropdown-item>
                
                <el-dropdown-item command="orders">我的订单</el-dropdown-item>
                <el-dropdown-item command="address">地址管理</el-dropdown-item>
                <el-dropdown-item command="cart">购物车</el-dropdown-item>
                <el-dropdown-item v-if="isExpert || isSysAdmin" command="knowledge">我的知识</el-dropdown-item>
                <el-dropdown-item v-if="isProcessor || isSysAdmin" command="processor">事务处理中心</el-dropdown-item>
                <el-dropdown-item command="messages">
                  我的私信<span v-if="msgUnreadCount > 0" style="color: #f56c6c; margin-left: 5px;">({{ msgUnreadCount }}未读)</span>
                </el-dropdown-item>
                <el-dropdown-item command="notifications">
                  我的通知<span v-if="unreadCount > 0" style="color: #f56c6c; margin-left: 5px;">({{ unreadCount }}未读)</span>
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
        <template v-else>
          <router-link to="/login"><el-button type="primary" size="small">登录</el-button></router-link>
          <router-link to="/register"><el-button size="small">注册</el-button></router-link>
        </template>
      </div>
    </div>
  </header>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { Bell, ChatDotRound, Message } from '@element-plus/icons-vue'
import { removeToken, removeUserInfo } from '@/utils/auth'
import api from '@/utils/api'

const router = useRouter()
const token = ref(localStorage.getItem('token'))
const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))

const unreadCount = ref(0)
const msgUnreadCount = ref(0)

const isExpert = computed(() => user.value?.role === 'expert')
const isProcessor = computed(() => user.value?.role === 'processor')
const isAdmin = computed(() => user.value?.role === 'admin')
const isSysAdmin = computed(() => user.value?.role === 'sysadmin')

const roleText = computed(() => {
  if (isSysAdmin.value) return '系统管理员'
  if (isAdmin.value) return '审核人员'
  if (isExpert.value) return '专家'
  if (isProcessor.value) return '处理人员'
  return '用户'
})

const roleTagType = computed(() => {
  if (isSysAdmin.value) return 'danger'
  if (isAdmin.value) return 'danger'
  if (isExpert.value) return 'warning'
  if (isProcessor.value) return 'success'
  return 'info'
})

const fetchUnreadCount = async () => {
  if (!token.value) return
  try {
    const res = await api.getUnreadCount()
    unreadCount.value = res?.count || 0
  } catch (error) {
    // 静默失败
  }
}

const fetchMsgUnreadCount = async () => {
  if (!token.value) return
  try {
    const res = await api.getMessageUnreadCount()
    msgUnreadCount.value = res?.count || 0
  } catch (error) {
    // 静默失败
  }
}

const handleUserCommand = (command) => {
  const routes = {
    'ai-chat': '/ai-chat',
    user: '/user',
    orders: '/orders',
    address: '/address',
    cart: '/cart',
    knowledge: '/my-knowledge',
    processor: '/processor',
    messages: '/messages',
    notifications: '/notifications'
  }
  if (command === 'logout') {
    removeToken()
    removeUserInfo()
    token.value = null
    user.value = null
    unreadCount.value = 0
    msgUnreadCount.value = 0
    router.push('/')
  } else if (routes[command]) {
    router.push(routes[command])
  }
}

// 监听通知更新事件
const handleNotificationUpdate = () => {
  fetchUnreadCount()
}

onMounted(() => {
  fetchUnreadCount()
  fetchMsgUnreadCount()
  // 定时轮询私信未读数
  const msgTimer = setInterval(fetchMsgUnreadCount, 15000)
  window.addEventListener('notification-update', handleNotificationUpdate)
  onUnmounted(() => {
    window.removeEventListener('notification-update', handleNotificationUpdate)
    clearInterval(msgTimer)
  })
})
</script>

<style scoped>
.header {
  background: #fff;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  position: sticky;
  top: 0;
  z-index: 100;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
  display: flex;
  align-items: center;
  height: 60px;
}

.logo {
  font-size: 20px;
  font-weight: bold;
  color: #409eff;
  text-decoration: none;
}

.nav {
  flex: 1;
  margin-left: 40px;
}

.nav a {
  margin: 0 15px;
  color: #666;
  text-decoration: none;
  font-size: 15px;
  transition: color 0.2s;
}

.nav a:hover,
.nav a.active {
  color: #409eff;
}

.user-area {
  display: flex;
  gap: 10px;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}
</style>
