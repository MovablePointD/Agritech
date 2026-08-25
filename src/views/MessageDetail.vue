<template>
  <div class="chat-detail-page">
    <AppHeader />
    <div class="chat-container">
      <!-- 顶部信息栏 -->
      <div class="chat-topbar">
        <el-button class="back-btn" @click="$router.push('/messages')" text>
          <el-icon><ArrowLeft /></el-icon> 返回
        </el-button>
        <div class="chat-user-info" v-if="otherUser">
          <el-avatar :size="36" :src="otherUser.avatar_url">
            {{ otherUser.nickname?.[0] || otherUser.username?.[0] || '?' }}
          </el-avatar>
          <span class="chat-username">
            {{ otherUser.nickname || otherUser.username }}
            <el-tag v-if="conversationType === 'consultation'" type="warning" size="small" effect="plain" style="margin-left: 6px;">
              咨询 · {{ expertName || '专家' }}
            </el-tag>
          </span>
        </div>
        <el-button type="danger" text size="small" @click="deleteConversation">
          <el-icon><Delete /></el-icon>
        </el-button>
      </div>

      <!-- 消息列表 -->
      <div class="message-list" ref="msgListRef" @scroll="onScroll">
        <div v-if="loading" class="loading-more" v-loading="loading" />
        <div v-if="!hasMore && messages.length > 0" class="no-more">— 没有更多消息了 —</div>
        
        <div
          v-for="msg in messages"
          :key="msg.id"
          class="msg-item"
          :class="{ mine: msg.sender_id === currentUserId }"
        >
          <template v-if="msg.sender_id !== currentUserId">
            <el-avatar :size="32" :src="otherUser?.avatar_url" class="msg-avatar">
              {{ otherUser?.nickname?.[0] || otherUser?.username?.[0] || '?' }}
            </el-avatar>
            <div class="msg-bubble other">
              {{ msg.content }}
            </div>
          </template>
          <template v-else>
            <div class="msg-bubble mine">
              {{ msg.content }}
            </div>
            <el-avatar :size="32" :src="currentUser?.avatar_url" class="msg-avatar">
              {{ currentUser?.nickname?.[0] || currentUser?.username?.[0] || '我' }}
            </el-avatar>
          </template>
        </div>
        
        <div v-if="messages.length === 0 && !loading" class="empty-chat">
          <el-empty description="暂无消息，发送第一条私信吧" />
        </div>
      </div>

      <!-- 输入区域 -->
      <div class="chat-input-area">
        <el-input
          v-model="inputContent"
          type="textarea"
          :rows="2"
          placeholder="输入消息... (Enter 发送)"
          resize="none"
          @keydown.enter.exact.prevent="send"
        />
        <el-button type="primary" :disabled="!inputContent.trim()" @click="send">
          发送
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Delete } from '@element-plus/icons-vue'
import api from '@/utils/api'
import { getUserInfo } from '@/utils/auth'
import AppHeader from '@/components/AppHeader.vue'

const route = useRoute()
const router = useRouter()

const conversationId = computed(() => Number(route.params.id))
const currentUser = ref(getUserInfo() || {})
const currentUserId = computed(() => currentUser.value?.id || 0)

const otherUser = ref(null)
const conversationType = ref('')
const expertName = ref('')
const messages = ref([])
const inputContent = ref('')
const loading = ref(false)
const sending = ref(false)
const page = ref(1)
const hasMore = ref(true)
const msgListRef = ref(null)
let pollTimer = null

const loadConversation = async () => {
  try {
    const res = await api.getConversationDetail(conversationId.value)
    const conv = res.conversation || {}
    otherUser.value = conv.other_user || null
    conversationType.value = conv.conversation_type || ''
    expertName.value = conv.expert_name || ''
  } catch (error) {
    console.error('加载会话失败:', error)
  }
}

const loadMessages = async (reset = false) => {
  if (reset) {
    page.value = 1
    hasMore.value = true
    messages.value = []
  }
  if (!hasMore.value) return

  loading.value = true
  try {
    const res = await api.getMessages(conversationId.value, {
      page: page.value,
      page_size: 30
    })
    const newMsgs = res.messages || []
    if (newMsgs.length < 30) {
      hasMore.value = false
    }
    if (reset) {
      messages.value = newMsgs
    } else {
      messages.value = [...newMsgs, ...messages.value]
    }
    if (reset) {
      scrollToBottom()
    }
  } catch (error) {
    console.error('加载消息失败:', error)
  } finally {
    loading.value = false
  }
}

const send = async () => {
  const content = inputContent.value.trim()
  if (!content || sending.value) return
  inputContent.value = ''
  sending.value = true

  try {
    const res = await api.sendMessage({
      conversation_id: conversationId.value,
      content
    })
    messages.value.push({
      id: res.message?.id || Date.now(),
      sender_id: currentUserId.value,
      content,
      is_read: 0,
      created_at: new Date().toISOString()
    })
    scrollToBottom()
  } catch (error) {
    ElMessage.error(error?.response?.data?.error || '发送失败')
  } finally {
    sending.value = false
  }
}

const markAsRead = async () => {
  try {
    await api.markMessageAsRead(conversationId.value)
  } catch (error) {
    // 静默失败
  }
}

const deleteConversation = async () => {
  try {
    await ElMessageBox.confirm('确定要删除该私信会话吗？所有消息将被删除。', '提示', { type: 'warning' })
    await api.deleteConversation(conversationId.value)
    ElMessage.success('已删除')
    router.push('/messages')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const scrollToBottom = () => {
  nextTick(() => {
    if (msgListRef.value) {
      msgListRef.value.scrollTop = msgListRef.value.scrollHeight
    }
  })
}

const onScroll = () => {
  if (!msgListRef.value || loading.value || !hasMore.value) return
  if (msgListRef.value.scrollTop <= 30) {
    const prevHeight = msgListRef.value.scrollHeight
    page.value++
    loadMessages(false).then(() => {
      nextTick(() => {
        if (msgListRef.value) {
          msgListRef.value.scrollTop = msgListRef.value.scrollHeight - prevHeight
        }
      })
    })
  }
}

// 定时刷新
const startPolling = () => {
  pollTimer = setInterval(() => {
    loadConversation()
    loadMessages(true)
  }, 8000)
}

watch(conversationId, (newId) => {
  if (newId) {
    page.value = 1
    hasMore.value = true
    messages.value = []
    loadConversation()
    loadMessages(true)
    markAsRead()
  }
})

onMounted(() => {
  loadConversation()
  loadMessages(true)
  markAsRead()
  startPolling()
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<style scoped>
.chat-detail-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f5f5f5;
}

.chat-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  max-width: 700px;
  width: 100%;
  margin: 0 auto;
  background: #fff;
  box-shadow: 0 0 20px rgba(0, 0, 0, 0.06);
}

.chat-topbar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  border-bottom: 1px solid #eee;
  background: #fff;
  flex-shrink: 0;
}

.back-btn {
  color: #666 !important;
}

.chat-user-info {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
}

.chat-username {
  font-size: 15px;
  font-weight: 600;
  color: #333;
}

.message-list {
  flex: 1;
  overflow-y: auto;
  padding: 20px 16px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.loading-more {
  text-align: center;
  padding: 10px;
  height: 40px;
}

.no-more {
  text-align: center;
  color: #ccc;
  font-size: 12px;
  padding: 10px 0;
}

.empty-chat {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.msg-item {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  max-width: 75%;
}

.msg-item.mine {
  align-self: flex-end;
  flex-direction: row-reverse;
}

.msg-avatar {
  flex-shrink: 0;
}

.msg-bubble {
  padding: 10px 14px;
  border-radius: 16px;
  font-size: 14px;
  line-height: 1.5;
  word-break: break-word;
}

.msg-bubble.other {
  background: #f0f0f0;
  color: #333;
  border-bottom-left-radius: 4px;
}

.msg-bubble.mine {
  background: #409eff;
  color: #fff;
  border-bottom-right-radius: 4px;
}

.chat-input-area {
  display: flex;
  gap: 10px;
  align-items: flex-end;
  padding: 12px 16px;
  border-top: 1px solid #eee;
  background: #fff;
  flex-shrink: 0;
}

.chat-input-area .el-textarea {
  flex: 1;
}
</style>
