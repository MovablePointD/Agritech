<template>
  <div class="ai-chat-page">
    <AppHeader />

    <div class="chat-container">
      <!-- 顶部工具栏 -->
      <div class="chat-toolbar">
        <div class="toolbar-left">
          <el-button class="back-btn" @click="$router.push('/')" text>
            <el-icon><ArrowLeft /></el-icon> 返回首页
          </el-button>
          <div class="toolbar-divider"></div>
          <span class="toolbar-icon">🌾</span>
          <h2>农技 AI 助手</h2>
        </div>
        <div class="toolbar-center">
          <el-radio-group v-model="currentModel" size="default" @change="handleModelChange" class="assistant-tabs">
            <el-radio-button value="chat-assistant">
              <el-icon><ChatDotRound /></el-icon> 聊天助手
            </el-radio-button>
            <el-radio-button value="system-assistant">
              <el-icon><Setting /></el-icon> 系统助手
            </el-radio-button>
          </el-radio-group>
        </div>
        <div class="toolbar-right">
          <el-button size="small" @click="showNewChatDialog = true">新建对话</el-button>
          <!-- <el-button size="small" type="primary" plain @click="showContentManager = true">
            <el-icon><Document /></el-icon> 我的内容
          </el-button> -->
          <!-- <el-button
            v-if="currentModel === 'system-assistant'"
            size="small"
            type="success"
            plain
            @click="showWorkspacePanel = !showWorkspacePanel"
          >
            <el-icon><Folder /></el-icon> 工作区
          </el-button> -->
        </div>
      </div>

      <div class="chat-body">
        <!-- 会话列表 -->
        <div class="session-list">
          <div class="session-header">
            <span>会话列表</span>
            <el-button size="small" text @click="showNewChatDialog = true">
              <el-icon><Plus /></el-icon>
            </el-button>
          </div>
          <div class="session-items">
            <div
              v-for="session in sessions"
              :key="session.id"
              class="session-item"
              :class="{ active: currentSession?.id === session.id }"
              @click="selectSession(session)"
            >
              <div class="session-info">
                <span class="session-title">{{ session.title }}</span>
                <span class="session-time">{{ formatTime(session.updated_at) }}</span>
              </div>
              <div class="session-actions">
                <el-button size="small" text @click.stop="deleteSession(session.id)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
            </div>
            <div v-if="sessions.length === 0" class="no-sessions">暂无会话</div>
          </div>
        </div>

        <!-- 聊天区域 -->
        <div class="chat-main">
          <div class="message-list" ref="messageListRef">
            <div v-if="messages.length === 0 && !isStreaming" class="welcome">
              <div class="welcome-icon">
                <span class="agri-bot-icon">🤖</span>
              </div>
              <h3>您好，我是{{ currentModel === 'system-assistant' ? '系统' : '农技' }} AI 助手</h3>
              <p v-if="currentModel === 'system-assistant'">
                我可以帮您读取项目文件、生成代码、自动发布动态内容
              </p>
              <p v-else>我可以帮您解答农业知识、撰写文章、分析问题，为乡村振兴贡献力量</p>
              <div v-if="!isLoggedIn" class="login-callout">
                <p class="login-cta-text">登录后即可与 AI 助手畅聊</p>
                <el-button type="primary" size="large" round @click="$router.push('/login')">前往登录</el-button>
              </div>
              <div v-else-if="currentModel === 'chat-assistant'" class="quick-prompts">
                <el-tag v-for="prompt in chatQuickPrompts" :key="prompt" @click="sendQuickMessage(prompt)" class="prompt-tag">
                  {{ prompt }}
                </el-tag>
              </div>
              <div v-else-if="currentModel === 'system-assistant'" class="quick-prompts">
                <el-tag v-for="prompt in systemQuickPrompts" :key="prompt" @click="sendQuickMessage(prompt)" class="prompt-tag sys-tag">
                  {{ prompt }}
                </el-tag>
              </div>
            </div>

            <div v-else class="messages">
              <div v-for="(msg, index) in messages" :key="index" class="message-item" :class="msg.role">
                <div class="message-avatar">
                  <el-avatar v-if="msg.role === 'user'" :size="36" style="background: #409eff;">
                    {{ user?.nickname?.[0] || user?.username?.[0] || 'U' }}
                  </el-avatar>
                  <el-avatar v-else :size="36" style="background: #67c23a;">
                    <el-icon><Service /></el-icon>
                  </el-avatar>
                </div>
                <div class="message-content">
                  <div class="message-header">
                    <span class="message-sender">{{ msg.role === 'user' ? '您' : assistantName }}</span>
                    <div v-if="msg.role === 'assistant' && msg.content" class="message-actions">
                      <el-button size="small" text @click="saveToDraft(msg.content, 'post')">
                        <el-icon><Document /></el-icon> 存为动态
                      </el-button>
                      <el-button
                        v-if="currentModel === 'system-assistant' && hasCodeBlock(msg.content)"
                        size="small"
                        text
                        type="success"
                        @click="applyCodeBlock(msg.content, index)"
                      >
                        <el-icon><Edit /></el-icon> 应用代码
                      </el-button>
                    </div>
                  </div>
                  <div class="message-text" v-html="formatMessage(msg.content)"></div>
                  <div class="message-time">{{ formatTimeStr(msg.created_at) }}</div>
                </div>
              </div>

              <div v-if="isStreaming" class="message-item assistant">
                <div class="message-avatar">
                  <el-avatar :size="36" style="background: #67c23a;">
                    <el-icon><Service /></el-icon>
                  </el-avatar>
                </div>
                <div class="message-content">
                  <div class="message-header">
                    <span class="message-sender">{{ assistantName }}</span>
                    <span class="streaming-indicator">▊</span>
                  </div>
                  <div class="message-text streaming" v-html="formatMessage(streamingContent)"></div>
                </div>
              </div>
            </div>
          </div>

          <!-- 输入区域 -->
          <div class="chat-input">
            <div class="input-actions">
              <el-button v-if="currentSession && messages.length > 0" size="small" @click="clearHistory">
                <el-icon><Delete /></el-icon> 清空历史
              </el-button>
            </div>
            <div class="input-row">
              <el-input
                v-model="inputMessage"
                type="textarea"
                :rows="2"
                :placeholder="inputPlaceholder"
                resize="none"
                @keydown.enter.exact.prevent="handleEnterKey"
                @keydown.shift.enter="handleShiftEnter"
              />
              <el-button type="primary" :loading="loading" @click="sendMessage" :disabled="!inputMessage.trim() || !isLoggedIn">
                发送
              </el-button>
            </div>
            <div v-if="!isLoggedIn" class="login-tip">
              <el-link type="primary" @click="$router.push('/login')">登录</el-link> 后即可使用 AI 助手
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 系统助手工作区面板 -->
    <el-drawer v-if="currentModel === 'system-assistant'" v-model="showWorkspacePanel" title="工作区" size="400px" direction="rtl">
      <div class="workspace-panel">
        <div class="workspace-path">
          <span class="ws-label">工作区根目录：</span>
          <div class="ws-path">{{ workspaceRoot }}</div>
          <el-button size="small" text @click="refreshWorkspace">
            <el-icon><Refresh /></el-icon>
          </el-button>
        </div>

        <el-input v-model="workspaceSearch" placeholder="搜索文件..." size="small" style="margin-bottom: 10px;" clearable>
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>

        <div class="workspace-tree" v-if="workspaceItems.length > 0">
          <div v-for="item in filteredWorkspaceItems" :key="item.path" class="ws-item" @click="handleWorkspaceItemClick(item)">
            <el-icon v-if="item.type === 'dir'" color="#409eff"><Folder /></el-icon>
            <el-icon v-else color="#67c23a"><Document /></el-icon>
            <span class="ws-name">{{ item.name }}</span>
            <span v-if="item.type === 'file'" class="ws-size">{{ formatFileSize(item.size) }}</span>
          </div>
        </div>
        <div v-else class="empty-state">
          <p>{{ loadingWorkspace ? '加载中...' : '点击刷新加载工作区' }}</p>
        </div>
      </div>
    </el-drawer>

    <!-- 代码应用对话框 -->
    <el-dialog v-model="showApplyCodeDialog" title="应用代码到文件" width="700px" :close-on-click-modal="false">
      <div class="apply-code-form">
        <el-form-item label="目标文件路径">
          <el-input v-model="applyCodePath" placeholder="请输入目标文件路径（相对于工作区）">
            <template #prepend>{{ workspaceRoot }}/</template>
          </el-input>
        </el-form-item>
        <el-form-item label="代码内容">
          <el-input v-model="applyCodeContent" type="textarea" :rows="15" placeholder="代码内容" />
        </el-form-item>
      </div>
      <template #footer>
        <el-button @click="showApplyCodeDialog = false">取消</el-button>
        <el-button type="primary" @click="doApplyCode">写入文件</el-button>
      </template>
    </el-dialog>

    <!-- 新建对话对话框 -->
    <el-dialog v-model="showNewChatDialog" title="新建对话" width="400px">
      <p>是否继续当前对话或开启新的对话？</p>
      <template #footer>
        <el-button @click="showNewChatDialog = false">取消</el-button>
        <el-button @click="showNewChatDialog = false">继续当前</el-button>
        <el-button type="primary" @click="createNewSession">新建对话</el-button>
      </template>
    </el-dialog>

    <!-- 预览对话框 -->
    <el-dialog v-model="showPreviewDialog" title="存为动态" width="700px" :close-on-click-modal="false">
      <div class="preview-container">
        <div class="preview-header">
          <el-input v-model="previewContent.title" placeholder="请输入动态标题" style="margin-bottom: 12px;">
            <template #prepend>标题</template>
          </el-input>
          <el-select v-if="previewContent.id" v-model="previewContent.type" placeholder="选择内容类型" style="margin-bottom: 12px; width: 100%;">
            <el-option v-if="permissions.can_create_content && permissions.content_types?.includes('knowledge')" label="知识" value="knowledge" />
            <el-option v-if="permissions.can_create_content && permissions.content_types?.includes('post')" label="动态" value="post" />
          </el-select>
          <div v-if="previewContent.type === 'post'" class="post-body-section">
            <div class="section-label">动态正文（必填）</div>
            <el-input v-model="previewContent.postBody" type="textarea" :rows="5" placeholder="请输入您的动态正文内容..." resize="vertical" />
          </div>
          <div v-else class="post-body-section">
            <div class="section-label">知识内容</div>
            <el-input v-model="previewContent.content" type="textarea" :rows="8" placeholder="请输入知识内容..." resize="vertical" />
          </div>
        </div>
        <div v-if="previewContent.type === 'post' && previewContent.content" class="preview-body" style="margin-top: 15px;">
          <div class="preview-label">AI 回答引用</div>
          <div class="preview-content" v-html="formatMessage(previewContent.content)"></div>
        </div>
        <div v-if="previewContent.status === 'pending'" class="preview-notice">
          <el-alert type="info" :closable="false">当前内容将提交审核，审核通过后方可发布</el-alert>
        </div>
        <div v-if="previewContent.status === 'published'" class="preview-notice">
          <el-alert type="success" :closable="false">您拥有免审核发布权限，内容将直接发布</el-alert>
        </div>
      </div>
      <template #footer>
        <el-button @click="showPreviewDialog = false">取消</el-button>
        <!-- <el-button @click="confirmSaveDraft">保存草稿</el-button> -->
        <el-button v-if="permissions.can_publish_directly" type="success" @click="confirmPublish">直接发布</el-button>
        <el-button v-else type="primary" @click="confirmSubmitReview">提交审核</el-button>
      </template>
    </el-dialog>

    <!-- 内容管理面板 -->
    <el-drawer v-model="showContentManager" title="我的内容" size="500px" direction="rtl">
      <div class="content-manager">
        <el-tabs v-model="contentTab">
          <el-tab-pane label="草稿" name="draft">
            <div v-if="draftContents.length === 0" class="empty-state">暂无草稿</div>
            <div v-else class="content-list">
              <div v-for="item in draftContents" :key="item.id" class="content-item">
                <div class="content-info">
                  <el-tag size="small">{{ item.type === 'knowledge' ? '知识' : '动态' }}</el-tag>
                  <div class="content-title">{{ item.title || '无标题' }}</div>
                  <div class="content-time">{{ formatTimeStr(item.updated_at) }}</div>
                </div>
                <div class="content-actions">
                  <el-button size="small" text @click="editContent(item)">编辑</el-button>
                  <el-button size="small" text type="primary" @click="submitForReview(item)">提交审核</el-button>
                  <el-button size="small" text type="danger" @click="deleteContent(item)">删除</el-button>
                </div>
              </div>
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
    </el-drawer>

    <AppFooter />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, nextTick, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, ChatDotRound, Service, Document, ArrowLeft, Setting, Folder, Search, Refresh, Edit } from '@element-plus/icons-vue'
import AppHeader from '@/components/AppHeader.vue'
import AppFooter from '@/components/AppFooter.vue'
import { getToken, getUserInfo } from '@/utils/auth'
import mainApi from '@/utils/api'
import axios from 'axios'
import { marked } from 'marked'
import hljs from 'highlight.js'
import 'highlight.js/styles/github.css'

marked.setOptions({ breaks: true, gfm: true })

const router = useRouter()

const token = computed(() => getToken())
const user = computed(() => getUserInfo())
const isLoggedIn = computed(() => !!token.value)
const userRole = computed(() => user.value?.role || 'user')
const userId = computed(() => user.value?.id || user.value?.username || 'anonymous')

const sessions = ref([])
const currentSession = ref(null)
const messages = ref([])
const inputMessage = ref('')
const loading = ref(false)
const isStreaming = ref(false)
const streamingContent = ref('')
const showNewChatDialog = ref(false)
const currentModel = ref('chat-assistant')
const assistantName = computed(() => currentModel.value === 'system-assistant' ? '系统助手' : '聊天助手')
const messageListRef = ref(null)

const showContentManager = ref(false)
const contentTab = ref('draft')
const permissions = ref({ can_create_content: true, can_publish_directly: false, can_review: false, content_types: ['knowledge', 'post'] })
const draftContents = ref([])
const pendingContents = ref([])
const publishedContents = ref([])
const allPendingContents = ref([])

const showPreviewDialog = ref(false)
const previewContent = reactive({ id: null, title: '', content: '', postBody: '', type: 'post', status: 'draft' })
const currentEditContent = ref(null)

// 系统助手 - 工作区
const showWorkspacePanel = ref(false)
const workspaceRoot = ref('')
const workspaceItems = ref([])
const loadingWorkspace = ref(false)
const workspaceSearch = ref('')
const showApplyCodeDialog = ref(false)
const applyCodePath = ref('')
const applyCodeContent = ref('')

const models = ref([
  { id: 'chat-assistant', name: '聊天助手' },
  { id: 'system-assistant', name: '系统助手' }
])

const inputPlaceholder = computed(() => {
  return currentModel.value === 'system-assistant'
    ? '输入消息或操作指令... (如：读取文件 / 生成代码 / 发布动态)'
    : '输入消息... (Shift+Enter换行, Enter发送)'
})

const chatQuickPrompts = [
  '如何科学种植水稻？',
  '常见农作物病虫害防治方法',
  '现代智慧农业技术有哪些？',
  '如何提高土壤肥力？'
]

const systemQuickPrompts = [
  '查看项目结构',
  '生成一个API接口代码',
  '帮我写一篇关于智慧农业的动态并发布',
  '分析这段代码的性能问题'
]

const filteredWorkspaceItems = computed(() => {
  if (!workspaceSearch.value) return workspaceItems.value
  const kw = workspaceSearch.value.toLowerCase()
  return workspaceItems.value.filter(item => item.name.toLowerCase().includes(kw))
})

const API_BASE = 'http://localhost:5000/api'

const api = {
  getSessions: () => axios.get(`${API_BASE}/sessions`, { params: { user_id: userId.value } }),
  createSession: (data) => axios.post(`${API_BASE}/sessions`, { ...data, user_id: userId.value }),
  getSession: (id) => axios.get(`${API_BASE}/sessions/${id}`, { params: { user_id: userId.value } }),
  deleteSession: (id) => axios.delete(`${API_BASE}/sessions/${id}`, { params: { user_id: userId.value } }),
  clearSession: (id) => axios.post(`${API_BASE}/sessions/${id}/clear`, { user_id: userId.value }),
  chat: (data) => axios.post(`${API_BASE}/chat`, { ...data, token: token.value }),
  getModels: () => axios.get(`${API_BASE}/models`),
  health: () => axios.get(`${API_BASE}/health`),
  getPrompts: (type) => axios.get(`${API_BASE}/prompts`, { params: { type } }),
  chatStream: async (data, onChunk, onDone, onError, onAction) => {
    try {
      const response = await fetch(`${API_BASE}/chat/stream`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ...data, user_id: userId.value, token: token.value })
      })
      if (!response.ok) {
        const errText = await response.text()
        let errMsg = `请求失败 (${response.status})`
        try { const errJson = JSON.parse(errText); errMsg = errJson.error || errMsg } catch {}
        onError(errMsg)
        return
      }
      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''
      while (true) {
        const { done, value } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() || ''
        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const dataStr = line.slice(6)
            if (dataStr === '[DONE]') continue
            try {
              const parsed = JSON.parse(dataStr)
              if (parsed.done) {
                onDone(parsed.session_id, { actions_executed: parsed.actions_executed || 0 })
              } else if (parsed.action) {
                if (onAction) onAction(parsed.action)
              } else if (parsed.error) {
                onError(parsed.error)
              } else if (parsed.content) {
                onChunk(parsed.content)
              }
            } catch {}
          }
        }
      }
    } catch (error) {
      onError(error.message || '网络错误')
    }
  },
  getPermissions: (role) => axios.get(`${API_BASE}/permissions`, { params: { role } }),
  getContents: (params) => axios.get(`${API_BASE}/contents`, { params }),
  getContent: (id) => axios.get(`${API_BASE}/contents/${id}`),
  createContent: (data) => axios.post(`${API_BASE}/contents`, data),
  updateContent: (id, data) => axios.put(`${API_BASE}/contents/${id}`, data),
  deleteContent: (id, data) => axios.delete(`${API_BASE}/contents/${id}`, { data }),
  submitContent: (id, data) => axios.post(`${API_BASE}/contents/${id}/submit`, data),
  publishContent: (id, data) => axios.post(`${API_BASE}/contents/${id}/publish`, data),
  rejectContent: (id, data) => axios.post(`${API_BASE}/contents/${id}/reject`, data),
  // 工作区接口
  workspaceList: () => axios.get(`${API_BASE}/workspace/list`),
  workspaceRead: (path) => axios.post(`${API_BASE}/workspace/read`, { path }),
  workspaceWrite: (path, content) => axios.post(`${API_BASE}/workspace/write`, { path, content })
}

// ==================== 会话管理 ====================
const loadSessions = async () => {
  if (!isLoggedIn.value) return
  try {
    const res = await api.getSessions()
    const list = res.data?.sessions || res.data || []
    sessions.value = Array.isArray(list) ? list : []
  } catch (error) {
    console.error('加载会话列表失败:', error)
  }
}

let sessionRefreshTimer = null
const startSessionRefresh = () => { stopSessionRefresh(); sessionRefreshTimer = setInterval(loadSessions, 10000) }
const stopSessionRefresh = () => { if (sessionRefreshTimer) { clearInterval(sessionRefreshTimer); sessionRefreshTimer = null } }

const selectSession = async (session) => {
  try {
    const res = await api.getSession(session.id)
    currentSession.value = res.data
    messages.value = res.data.messages || []
    currentModel.value = res.data.model || 'chat-assistant'
    scrollToBottom()
  } catch (error) {
    ElMessage.error('加载会话失败')
    await loadSessions()
  }
}

const createNewSession = async () => {
  showNewChatDialog.value = false
  try {
    const res = await api.createSession({ model: currentModel.value })
    currentSession.value = res.data
    messages.value = []
    await loadSessions()
  } catch (error) {
    ElMessage.error('创建会话失败')
  }
}

const deleteSession = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除该会话吗？', '提示', { type: 'warning' })
    await api.deleteSession(id)
    ElMessage.success('已删除')
    if (currentSession.value?.id === id) { currentSession.value = null; messages.value = [] }
    await loadSessions()
  } catch (error) {
    if (error !== 'cancel') ElMessage.error('删除失败')
  }
}

const clearHistory = async () => {
  if (!currentSession.value) return
  try {
    await ElMessageBox.confirm('确定要清空聊天历史吗？', '提示', { type: 'warning' })
    await api.clearSession(currentSession.value.id)
    messages.value = []
    ElMessage.success('历史已清空')
  } catch (error) {
    if (error !== 'cancel') ElMessage.error('操作失败')
  }
}

// ==================== 消息发送 ====================
const sendMessage = async () => {
  if (!isLoggedIn.value) { ElMessage.warning('请先登录'); return }
  const content = inputMessage.value.trim()
  if (!content) return
  inputMessage.value = ''

  const userMsg = { role: 'user', content, created_at: new Date().toISOString() }
  messages.value.push(userMsg)

  isStreaming.value = true
  streamingContent.value = ''
  scrollToBottom()

  let fullContent = ''
  let finalSessionId = currentSession.value?.id

  await api.chatStream(
    { session_id: currentSession.value?.id, message: content, model: currentModel.value },
    (chunk) => { fullContent += chunk; streamingContent.value = fullContent; scrollToBottom() },
    async (sessionId, doneData) => {
      finalSessionId = sessionId
      if (finalSessionId && (!currentSession.value || currentSession.value.id !== finalSessionId)) {
        currentSession.value = { id: finalSessionId, model: currentModel.value }
      }
      // Reload session to get cleaned messages (without raw ACTION tags)
      // Keep streaming content visible until reload finishes for seamless transition
      await reloadCurrentSession()
      isStreaming.value = false
      streamingContent.value = ''
      await loadSessions()
      scrollToBottom()
      // Notify about action results
      const count = doneData?.actions_executed || 0
      if (count > 0) {
        ElMessage.success(`System assistant completed ${count} action(s)`)
      }
    },
    (errorMsg) => {
      isStreaming.value = false
      streamingContent.value = ''
      if (messages.value.length > 0 && messages.value[messages.value.length - 1].role === 'user') {
        messages.value.pop()
      }
      ElMessage.error(errorMsg)
      scrollToBottom()
    },
    (action) => {
      // Handle individual action events from server
      if (action.action === 'publish_post') {
        ElMessage.success(`Post published: "${action.title}"`)
      } else if (action.action === 'publish_knowledge') {
        ElMessage.success(`Knowledge published: "${action.title}"`)
      }
    }
  )
}

// Reload messages from the current session (to get cleaned content after action parsing)
const reloadCurrentSession = async () => {
  if (!currentSession.value?.id) return
  try {
    const res = await api.getSession(currentSession.value.id)
    if (res.data?.messages) {
      messages.value = res.data.messages
    }
  } catch (error) {
    console.error('Reload session failed:', error)
  }
}

const sendQuickMessage = async (content) => {
  if (!isLoggedIn.value) { ElMessage.warning('请先登录后使用 AI 助手'); return }
  inputMessage.value = content
  await sendMessage()
}

const handleEnterKey = (e) => { e.preventDefault(); sendMessage() }
const handleShiftEnter = () => {}

const handleModelChange = async (model) => {
  if (currentSession.value) {
    try {
      await axios.put(`${API_BASE}/sessions/${currentSession.value.id}`, { model, user_id: userId.value })
      currentSession.value.model = model
    } catch (error) {
      console.error('切换模型失败:', error)
    }
  }
  // 切换系统助手时自动加载工作区
  if (model === 'system-assistant') {
    await loadWorkspace()
  }
}

// ==================== 工作区操作 ====================
const loadWorkspace = async () => {
  loadingWorkspace.value = true
  try {
    const res = await api.workspaceList()
    const workspaces = res.data?.workspaces || []
    if (workspaces.length > 0) {
      workspaceRoot.value = workspaces[0].path
      await refreshWorkspace()
    }
  } catch (error) {
    console.error('加载工作区失败:', error)
  } finally {
    loadingWorkspace.value = false
  }
}

const refreshWorkspace = async () => {
  if (!workspaceRoot.value) return
  loadingWorkspace.value = true
  try {
    const res = await api.workspaceRead(workspaceRoot.value)
    if (res.data?.type === 'directory') {
      workspaceItems.value = res.data.items || []
    }
  } catch (error) {
    console.error('刷新工作区失败:', error)
  } finally {
    loadingWorkspace.value = false
  }
}

const handleWorkspaceItemClick = async (item) => {
  if (item.type === 'file') {
    const path = workspaceRoot.value + '/' + item.name
    try {
      const res = await api.workspaceRead(path)
      if (res.data?.type === 'file') {
        ElMessage.success(`已读取文件: ${item.name}`)
        // 将文件内容发送到聊天
        inputMessage.value = `请分析以下文件的内容：\n\n文件: ${path}\n\`\`\`\n${res.data.content?.substring(0, 3000) || ''}\n\`\`\``
      }
    } catch (error) {
      ElMessage.error('读取文件失败')
    }
  } else {
    // 进入子目录
    try {
      const path = workspaceRoot.value + '/' + item.name
      const res = await api.workspaceRead(path)
      if (res.data?.type === 'directory') {
        workspaceItems.value = res.data.items || []
        workspaceRoot.value = path
      }
    } catch (error) {
      console.error('进入目录失败:', error)
    }
  }
}

// 检测消息是否包含代码块
const hasCodeBlock = (content) => {
  return /```[\s\S]*?```/.test(content)
}

const applyCodeBlock = (content) => {
  const codeMatch = content.match(/```(?:(\w+)\n)?([\s\S]*?)```/)
  if (codeMatch) {
    applyCodeContent.value = codeMatch[2].trim()
    applyCodePath.value = ''
    showApplyCodeDialog.value = true
  }
}

const doApplyCode = async () => {
  if (!applyCodePath.value.trim()) { ElMessage.warning('请输入目标文件路径'); return }
  if (!applyCodeContent.value.trim()) { ElMessage.warning('代码内容为空'); return }
  try {
    await api.workspaceWrite(applyCodePath.value, applyCodeContent.value)
    ElMessage.success(`已写入文件: ${applyCodePath.value}`)
    showApplyCodeDialog.value = false
  } catch (error) {
    ElMessage.error(error?.response?.data?.error || '写入文件失败')
  }
}

// ==================== 格式化 ====================
const formatMessage = (content) => {
  if (!content) return ''
  let html = marked.parse(content)
  html = html.replace(/<pre><code class="language-(\w+)">([\s\S]*?)<\/code><\/pre>/g, (match, lang, code) => {
    const decoded = code.replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&amp;/g, '&').replace(/&quot;/g, '"').replace(/&#39;/g, "'")
    let highlighted
    try {
      highlighted = lang && hljs.getLanguage(lang) ? hljs.highlight(decoded, { language: lang, ignoreIllegals: true }).value : hljs.highlightAuto(decoded).value
    } catch {
      highlighted = decoded.replace(/</g, '&lt;').replace(/>/g, '&gt;')
    }
    return `<pre class="code-block"><code class="language-${lang} hljs">${highlighted}</code></pre>`
  })
  return html
}

const formatTime = (timestamp) => {
  if (!timestamp) return ''
  const date = new Date(timestamp * 1000)
  const now = new Date()
  const diff = now - date
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`
  return date.toLocaleDateString('zh-CN')
}

const formatTimeStr = (time) => { if (!time) return ''; return new Date(time).toLocaleString('zh-CN') }

const formatFileSize = (bytes) => {
  if (!bytes) return ''
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

const scrollToBottom = () => { nextTick(() => { if (messageListRef.value) messageListRef.value.scrollTop = messageListRef.value.scrollHeight }) }

// ==================== 服务 & 权限 ====================
const checkHealth = async () => {
  try { await api.health() } catch (error) { ElMessage.warning('无法连接到 AI 服务，请确保后端服务已启动') }
}

const loadPermissions = async () => {
  try {
    const res = await api.getPermissions(userRole.value)
    permissions.value = res.data.permissions || { can_create_content: true, can_publish_directly: false, can_review: false, content_types: ['knowledge', 'post'] }
  } catch (error) { console.error('加载权限失败:', error) }
}

const loadContents = async () => {
  try {
    const uid = userId.value
    draftContents.value = (await api.getContents({ user_id: uid, status: 'draft' })).data.contents || []
    pendingContents.value = (await api.getContents({ user_id: uid, status: 'pending' })).data.contents || []
    publishedContents.value = (await api.getContents({ user_id: uid, status: 'published' })).data.contents || []
    if (permissions.value.can_review) {
      allPendingContents.value = (await api.getContents({ status: 'pending' })).data.contents || []
    }
  } catch (error) { console.error('加载内容列表失败:', error) }
}

// ==================== 内容管理 ====================
const saveToDraft = (content) => {
  if (!isLoggedIn.value) { ElMessage.warning('请先登录'); return }
  currentEditContent.value = null
  previewContent.id = null
  previewContent.title = ''
  previewContent.postBody = ''
  previewContent.content = content
  previewContent.type = 'post'
  previewContent.status = permissions.value.can_publish_directly ? 'published' : 'draft'
  showPreviewDialog.value = true
}

const confirmSaveDraft = async () => {
  if (!previewContent.title.trim()) { ElMessage.warning('请输入标题'); return }
  if (previewContent.type === 'post' && !previewContent.postBody.trim()) { ElMessage.warning('请输入动态正文内容'); return }
  if (previewContent.type === 'knowledge' && !previewContent.content.trim()) { ElMessage.warning('请输入知识内容'); return }

  try {
    const postContent = previewContent.type === 'post' ? buildPostContent() : previewContent.content
    if (previewContent.type === 'post' || previewContent.type === 'share') {
      if (previewContent.id) {
        await mainApi.updateDraftPost({ id: previewContent.id, title: previewContent.title, content: postContent, images: '' })
        ElMessage.success('草稿已更新')
      } else {
        const res = await mainApi.createDraftPost({ title: previewContent.title, content: postContent, type: 'share', images: '' })
        if (res.data?.id) previewContent.id = res.data.id
        ElMessage.success('已保存为草稿')
      }
    } else {
      if (previewContent.id) {
        await mainApi.updateDraftKnowledge({ id: previewContent.id, title: previewContent.title, content: postContent })
        ElMessage.success('草稿已更新')
      } else {
        const res = await mainApi.createDraftKnowledge({ title: previewContent.title, content: postContent })
        if (res.data?.id) previewContent.id = res.data.id
        ElMessage.success('已保存为草稿')
      }
    }
    try {
      const aiData = { user_id: userId.value, role: userRole.value, title: previewContent.title, content: postContent, type: previewContent.type }
      if (previewContent.id) await api.updateContent(previewContent.id, aiData)
      else await api.createContent(aiData)
    } catch {}
    showPreviewDialog.value = false
    await loadContents()
  } catch (error) { ElMessage.error(error?.response?.data?.error || error?.message || '保存失败') }
}

const buildPostContent = () => {
  const parts = []
  if (previewContent.postBody.trim()) parts.push(previewContent.postBody.trim())
  if (previewContent.content.trim()) {
    const { body, refs } = extractReferences(previewContent.content.trim())
    parts.push('\n\n---\n\n**🤖 AI 助手回答：**\n\n' + body)
    if (refs) parts.push(formatRefBlock(refs))
  }
  return parts.join('')
}

const extractReferences = (content) => {
  if (!content) return { body: '', refs: '' }
  const refPatterns = [
    /(\n\n#{1,3}\s*(?:参考|引用)(?:文献|资料|来源|信息)[\s\S]*?)$/m,
    /(\n\n(?:\[[\d,]+\][^\n]*\n?)+)$/,
    /(\n\n(?:数据来源|信息来源|参考来源|文章来源|source|reference)[^\n]*)$/im,
    /(\n\n(?:&gt;|>)\s*\*\*引用[\s\S]*?)$/m,
    /(\n\n---\n[\s\S]*?参考[\s\S]*?)$/m
  ]
  let refs = '', body = content
  for (const pattern of refPatterns) {
    const match = content.match(pattern)
    if (match) { refs = match[1].trim(); body = content.substring(0, match.index).trim(); break }
  }
  return { body, refs }
}

const formatRefBlock = (refs) => {
  if (!refs) return ''
  const lines = refs.split('\n')
  const formatted = lines.map(line => `> ${line.trim().replace(/^&gt;\s*/, '').replace(/^>\s*/, '')}`).join('\n')
  return `\n\n> **📚 引用来源：**\n${formatted}`
}

const confirmSubmitReview = async () => {
  if (!previewContent.title.trim()) { ElMessage.warning('请输入标题'); return }
  if (previewContent.type === 'post' && !previewContent.postBody.trim()) { ElMessage.warning('请输入动态正文内容'); return }
  try {
    const postContent = previewContent.type === 'post' ? buildPostContent() : previewContent.content
    await mainApi.createPost({ title: previewContent.title, content: postContent, type: 'share', images: '', video_url: '' })
    ElMessage.success('提交成功')
    showPreviewDialog.value = false
    await loadContents()
  } catch (error) { ElMessage.error(error?.message || error?.response?.data?.error || '提交失败') }
}

const confirmPublish = async () => {
  if (!previewContent.title.trim()) { ElMessage.warning('请输入标题'); return }
  if (previewContent.type === 'post' && !previewContent.postBody.trim()) { ElMessage.warning('请输入动态正文内容'); return }
  try {
    const postContent = previewContent.type === 'post' ? buildPostContent() : previewContent.content
    const res = await mainApi.createPost({ title: previewContent.title, content: postContent, type: 'share', images: '', video_url: '' })
    ElMessage.success(res.msg || '发布成功')
    showPreviewDialog.value = false
    await loadContents()
  } catch (error) { ElMessage.error(error?.message || error?.response?.data?.error || '发布失败') }
}

const editContent = (item) => {/* ... */ }
const previewItem = (item) => {/* ... */ }
const submitForReview = async (item) => {
  try { await api.submitContent(item.id, { user_id: userId.value }); ElMessage.success('已提交审核'); await loadContents() }
  catch (error) { ElMessage.error(error?.response?.data?.error || '提交失败') }
}
const deleteContent = async (item) => {
  try {
    await ElMessageBox.confirm('确定要删除该内容吗？', '提示', { type: 'warning' })
    await api.deleteContent(item.id, { user_id: userId.value, role: userRole.value })
    ElMessage.success('已删除'); await loadContents()
  } catch (error) { if (error !== 'cancel') ElMessage.error(error?.response?.data?.error || '删除失败') }
}

onMounted(async () => {
  await loadModels()
  if (isLoggedIn.value) {
    await loadPermissions(); await loadContents(); await loadSessions()
    if (sessions.value.length > 0) await selectSession(sessions.value[0])
    startSessionRefresh()
  }
  checkHealth()
})

onUnmounted(() => { stopSessionRefresh() })

const loadModels = async () => {
  try {
    const res = await api.getModels()
    if (res.data?.models?.length > 0) models.value = res.data.models
  } catch (error) { console.error('加载模型列表失败:', error) }
}
</script>

<style scoped>
.ai-chat-page {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: #f5f5f5;
}

.chat-container {
  display: flex;
  flex-direction: column;
  flex: 1;
  height: calc(100vh - 120px);
  max-height: calc(100vh - 120px);
  overflow: hidden;
}

/* ========== 工具栏 ========== */
.chat-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 20px;
  background: linear-gradient(135deg, #2e7d32 0%, #43a047 100%);
  box-shadow: 0 2px 8px rgba(46, 125, 50, 0.2);
  flex-shrink: 0;
}

.toolbar-left { display: flex; align-items: center; gap: 0; }
.back-btn { color: rgba(255, 255, 255, 0.9) !important; font-size: 13px; padding: 6px 12px; border-radius: 6px; }
.back-btn:hover { color: #fff !important; background: rgba(255, 255, 255, 0.15) !important; }
.toolbar-divider { width: 1px; height: 24px; background: rgba(255, 255, 255, 0.3); margin: 0 14px; }
.toolbar-icon { font-size: 18px; margin-right: 6px; }
.toolbar-left h2 { margin: 0; font-size: 16px; color: #fff; font-weight: 600; white-space: nowrap; }

.toolbar-center { flex: 1; display: flex; justify-content: center; }
.assistant-tabs { --el-radio-button-checked-bg-color: rgba(255,255,255,0.2); --el-radio-button-checked-border-color: rgba(255,255,255,0.5); }
.assistant-tabs .el-radio-button__inner { background: rgba(255,255,255,0.1); border-color: rgba(255,255,255,0.2); color: #fff; }

.toolbar-right { display: flex; align-items: center; gap: 6px; }

/* ========== 主体 ========== */
.chat-body { display: flex; flex: 1; overflow: hidden; min-height: 0; }

.session-list {
  width: 240px; min-width: 240px; background: #fff; border-right: 1px solid #e0e0e0;
  display: flex; flex-direction: column; box-shadow: 2px 0 6px rgba(0, 0, 0, 0.04); overflow: hidden;
}
.session-header { display: flex; justify-content: space-between; align-items: center; padding: 12px 14px; border-bottom: 2px solid #e8f5e9; font-weight: 600; color: #2e7d32; font-size: 13px; flex-shrink: 0; }
.session-items { flex: 1; overflow-y: auto; padding: 8px; }
.session-item { display: flex; justify-content: space-between; align-items: center; padding: 10px 12px; border-radius: 8px; cursor: pointer; transition: all 0.2s; margin-bottom: 2px; border: 1px solid transparent; }
.session-item:hover { background: #e8f5e9; border-color: #c8e6c9; }
.session-item.active { background: #e8f5e9; border-color: #66bb6a; }
.session-info { flex: 1; min-width: 0; overflow: hidden; }
.session-title { display: block; font-size: 12px; color: #333; font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; margin-bottom: 2px; }
.session-time { font-size: 10px; color: #999; }
.session-actions { flex-shrink: 0; margin-left: 4px; opacity: 0; transition: opacity 0.2s; }
.session-item:hover .session-actions { opacity: 1; }
.no-sessions { text-align: center; padding: 30px 15px; color: #bbb; font-size: 12px; }

/* ========== 聊天区域 ========== */
.chat-main { flex: 1; display: flex; flex-direction: column; overflow: hidden; background: #fafdf8; }
.message-list { flex: 1; overflow-y: auto; padding: 16px 20px; }

.welcome { display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100%; text-align: center; padding: 30px; }
.agri-bot-icon { font-size: 56px; display: block; margin-bottom: 8px; animation: float 3s ease-in-out infinite; }
@keyframes float { 0%, 100% { transform: translateY(0); } 50% { transform: translateY(-8px); } }
.welcome h3 { margin-bottom: 6px; color: #2e7d32; font-size: 20px; font-weight: 600; }
.welcome p { color: #666; margin-bottom: 20px; font-size: 13px; max-width: 400px; line-height: 1.5; }
.quick-prompts { display: flex; flex-wrap: wrap; justify-content: center; max-width: 480px; gap: 8px; }
.prompt-tag { cursor: pointer; padding: 6px 14px !important; font-size: 12px; border-radius: 18px !important; background: #e8f5e9 !important; color: #2e7d32 !important; border: 1px solid #a5d6a7 !important; transition: all 0.2s; }
.prompt-tag:hover { background: #c8e6c9 !important; border-color: #66bb6a !important; transform: translateY(-1px); }
.sys-tag { background: #e3f2fd !important; color: #1565c0 !important; border-color: #90caf9 !important; }
.sys-tag:hover { background: #bbdefb !important; }
.login-callout { margin-top: 10px; }
.login-cta-text { color: #999; margin-bottom: 10px; }

.messages { max-width: 800px; margin: 0 auto; }
.message-item { display: flex; margin-bottom: 16px; }
.message-item.user { flex-direction: row-reverse; }
.message-avatar { flex-shrink: 0; margin: 0 8px; }
.message-content { max-width: 70%; }
.message-header { display: flex; align-items: center; gap: 8px; margin-bottom: 2px; padding: 0 4px; }
.message-sender { font-size: 12px; color: #888; font-weight: 500; }
.message-actions { margin-left: auto; display: flex; gap: 4px; }
.message-text { padding: 10px 14px; border-radius: 12px; line-height: 1.6; word-break: break-word; font-size: 13px; }
.user .message-text { background: linear-gradient(135deg, #43a047 0%, #66bb6a 100%); color: #fff; border-bottom-right-radius: 4px; box-shadow: 0 1px 3px rgba(46, 125, 50, 0.2); }
.assistant .message-text { background: #fff; border: 1px solid #e8e8e8; border-bottom-left-radius: 4px; box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04); }
.message-text.streaming { border: 1px dashed #a5d6a7; }
.message-time { font-size: 10px; color: #bbb; margin-top: 2px; padding: 0 4px; }
.streaming-indicator { color: #43a047; font-size: 14px; animation: blink 1s infinite; }
@keyframes blink { 0%, 50% { opacity: 1; } 51%, 100% { opacity: 0; } }

/* ========== 输入区域 ========== */
.chat-input { border-top: 1px solid #e0e0e0; padding: 12px 20px; background: #fff; }
.input-actions { display: flex; margin-bottom: 6px; }
.input-row { display: flex; gap: 10px; align-items: flex-end; }
.input-row .el-textarea { flex: 1; }
.login-tip { text-align: center; margin-top: 8px; font-size: 12px; color: #999; }

/* ========== 工作区面板 ========== */
.workspace-panel { padding: 0; }
.workspace-path { display: flex; align-items: center; gap: 6px; margin-bottom: 10px; font-size: 12px; }
.ws-label { color: #666; white-space: nowrap; }
.ws-path { flex: 1; color: #409eff; font-family: monospace; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.workspace-tree { max-height: calc(100vh - 200px); overflow-y: auto; }
.ws-item { display: flex; align-items: center; gap: 6px; padding: 6px 8px; cursor: pointer; border-radius: 4px; font-size: 13px; transition: background 0.15s; }
.ws-item:hover { background: #f0f7ff; }
.ws-name { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.ws-size { font-size: 11px; color: #999; }
.empty-state { text-align: center; padding: 30px; color: #999; font-size: 13px; }

/* ========== 代码应用 ========== */
.apply-code-form { display: flex; flex-direction: column; gap: 12px; }

/* ========== 内容管理 ========== */
.content-manager { padding: 0; }
.content-list { max-height: calc(100vh - 160px); overflow-y: auto; }
.content-item { display: flex; justify-content: space-between; padding: 10px 0; border-bottom: 1px solid #f0f0f0; }
.content-info { flex: 1; min-width: 0; }
.content-title { font-size: 13px; color: #333; margin-top: 4px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.content-time { font-size: 11px; color: #999; margin-top: 2px; }
.content-actions { display: flex; gap: 4px; align-items: center; }

/* ========== 预览 ========== */
.preview-container { max-height: 60vh; overflow-y: auto; }
.section-label { font-size: 12px; color: #666; margin-bottom: 4px; }
.preview-content { padding: 10px; background: #f9f9f9; border-radius: 6px; max-height: 300px; overflow-y: auto; }
.preview-notice { margin-top: 10px; }

/* ========== 消息文本内样式 ========== */
:deep(.message-text pre) { background: #282c34; color: #abb2bf; border-radius: 8px; padding: 12px; overflow-x: auto; margin: 8px 0; }
:deep(.message-text code) { font-family: 'Fira Code', Consolas, monospace; font-size: 12px; }
:deep(.message-text table) { border-collapse: collapse; width: 100%; margin: 8px 0; }
:deep(.message-text th), :deep(.message-text td) { border: 1px solid #ddd; padding: 6px 10px; text-align: left; font-size: 12px; }
:deep(.message-text th) { background: #f5f5f5; }
:deep(.user .message-text pre) { background: rgba(0,0,0,0.2); }
:deep(.user .message-text code) { color: #fff; }
</style>
