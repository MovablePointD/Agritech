<template>
  <div class="messages-page">
    <AppHeader />
    <div class="container main">
      <div class="page-header">
        <h2><el-icon><ChatDotRound /></el-icon> 我的私信</h2>
        <el-button type="primary" @click="showNewDialog = true">
          <el-icon><Plus /></el-icon> 新建私信
        </el-button>
      </div>

      <el-card v-loading="loading" class="conversation-card">
        <div v-if="conversations.length === 0 && !loading" class="empty-state">
          <el-empty description="暂无私信" />
        </div>
        <div v-else class="conversation-list">
          <div
            v-for="conv in conversations"
            :key="conv.id"
            class="conversation-item"
            :class="{ unread: conv.unread_count > 0 }"
            @click="openConversation(conv.id)"
          >
            <div class="conv-avatar">
              <el-avatar :size="44" :src="conv.other_user?.avatar_url">
                {{ conv.other_user?.nickname?.[0] || conv.other_user?.username?.[0] || '?' }}
              </el-avatar>
            </div>
            <div class="conv-body">
              <div class="conv-top">
                <span class="conv-name">
                  {{ conv.other_user?.nickname || conv.other_user?.username || '未知用户' }}
                  <el-tag v-if="conv.conversation_type === 'consultation'" type="warning" size="small" effect="plain" style="margin-left: 6px;">
                    咨询
                  </el-tag>
                </span>
                <span class="conv-time">{{ conv.updated_at }}</span>
              </div>
              <div class="conv-bottom">
                <span class="conv-preview">{{ conv.last_message || '暂无消息' }}</span>
                <el-badge v-if="conv.unread_count > 0" :value="conv.unread_count" :max="99" />
              </div>
            </div>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 新建私信对话框 -->
    <el-dialog v-model="showNewDialog" title="新建私信" width="450px">
      <el-form label-width="80px">
        <el-form-item label="搜索用户">
          <el-select
            v-model="selectedUser"
            filterable
            remote
            reserve-keyword
            placeholder="输入用户名搜索"
            :remote-method="searchUsers"
            :loading="searchLoading"
            style="width: 100%;"
          >
            <el-option
              v-for="u in searchResults"
              :key="u.id"
              :label="`${u.nickname || u.username} (${u.username})`"
              :value="u.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showNewDialog = false">取消</el-button>
        <el-button type="primary" :disabled="!selectedUser" @click="startConversation">开始私信</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ChatDotRound, Plus } from '@element-plus/icons-vue'
import api from '@/utils/api'
import AppHeader from '@/components/AppHeader.vue'

const router = useRouter()

const conversations = ref([])
const loading = ref(false)
const showNewDialog = ref(false)
const selectedUser = ref(null)
const searchResults = ref([])
const searchLoading = ref(false)
let pollTimer = null

const loadConversations = async () => {
  loading.value = true
  try {
    const res = await api.getConversations()
    conversations.value = res.conversations || []
  } catch (error) {
    console.error('加载私信列表失败:', error)
  } finally {
    loading.value = false
  }
}

const openConversation = (id) => {
  router.push(`/messages/${id}`)
}

const searchUsers = async (query) => {
  if (!query || query.length < 1) {
    searchResults.value = []
    return
  }
  searchLoading.value = true
  try {
    const res = await api.getUserList({ keyword: query, page: 1, page_size: 10 })
    searchResults.value = res.list || []
  } catch (error) {
    console.error('搜索用户失败:', error)
  } finally {
    searchLoading.value = false
  }
}

const startConversation = async () => {
  if (!selectedUser.value) return
  try {
    const res = await api.createConversation({ target_user_id: selectedUser.value })
    showNewDialog.value = false
    selectedUser.value = null
    router.push(`/messages/${res.conversation_id}`)
  } catch (error) {
    ElMessage.error(error?.response?.data?.error || '创建会话失败')
  }
}

// 定时轮询未读数
const startPolling = () => {
  pollTimer = setInterval(loadConversations, 10000)
}

onMounted(() => {
  loadConversations()
  startPolling()
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<style scoped>
.messages-page {
  min-height: 100vh;
  background: #f5f5f5;
}

.container.main {
  max-width: 700px;
  margin: 0 auto;
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  font-size: 20px;
  color: #333;
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
}

.conversation-card {
  border-radius: 12px;
}

.empty-state {
  padding: 40px 0;
}

.conversation-list {
  display: flex;
  flex-direction: column;
}

.conversation-item {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 16px;
  cursor: pointer;
  border-bottom: 1px solid #f0f0f0;
  transition: background 0.2s;
}

.conversation-item:last-child {
  border-bottom: none;
}

.conversation-item:hover {
  background: #f5f7fa;
}

.conversation-item.unread {
  background: #ecf5ff;
}

.conversation-item.unread:hover {
  background: #d9ecff;
}

.conv-avatar {
  flex-shrink: 0;
}

.conv-body {
  flex: 1;
  min-width: 0;
}

.conv-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.conv-name {
  font-size: 15px;
  font-weight: 600;
  color: #333;
}

.conv-time {
  font-size: 12px;
  color: #999;
  flex-shrink: 0;
  margin-left: 10px;
}

.conv-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.conv-preview {
  font-size: 13px;
  color: #999;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  max-width: 420px;
}

.conversation-item.unread .conv-preview {
  color: #555;
  font-weight: 500;
}
</style>
