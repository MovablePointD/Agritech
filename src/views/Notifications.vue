<template>
  <div class="notifications-page">
    <AppHeader />

    <div class="container main">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>我的通知</span>
            <el-button v-if="notifications.length > 0" type="primary" link @click="handleMarkAllAsRead">
              全部标记为已读
            </el-button>
          </div>
        </template>

        <div class="notification-list" v-loading="loading">
          <div
            v-for="item in notifications"
            :key="item.id"
            class="notification-item"
            :class="{ unread: item.is_read === 0 }"
            @click="handleItemClick(item)"
          >
            <div class="notification-icon" :class="getIconClass(item.type)">
              <el-icon v-if="item.type === 'stock_low'"><Warning /></el-icon>
              <el-icon v-else-if="item.type === 'order'"><ShoppingCart /></el-icon>
              <el-icon v-else-if="item.type === 'comment_post' || item.type === 'comment_knowledge'"><ChatDotRound /></el-icon>
              <el-icon v-else-if="item.type === 'reply_comment'"><ChatLineRound /></el-icon>
              <el-icon v-else-if="item.type === 'at_comment' || item.type === 'new_message'"><Message /></el-icon>
              <el-icon v-else><Bell /></el-icon>
            </div>
            <div class="notification-content">
              <div class="notification-header">
                <span class="notification-title">{{ item.title }}</span>
                <el-tag :type="getTagType(item.type)" size="small">{{ getTypeText(item.type) }}</el-tag>
              </div>
              <div class="notification-text">{{ item.content }}</div>
              <div class="notification-time">{{ formatTime(item.created_at) }}</div>
            </div>
          </div>
          <el-empty v-if="!loading && notifications.length === 0" description="暂无通知" />
        </div>

        <div class="pagination-wrapper" v-if="total > 0">
          <el-pagination
            v-model:current-page="page"
            :page-size="pageSize"
            :total="total"
            layout="prev, pager, next"
            @current-change="fetchData"
          />
        </div>
      </el-card>
    </div>

    <AppFooter />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Warning, ShoppingCart, Bell, ChatDotRound, ChatLineRound, Message } from '@element-plus/icons-vue'
import api from '@/utils/api'
import { getToken } from '@/utils/auth'

const router = useRouter()
const token = getToken()

const notifications = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

const formatTime = (time) => {
  if (!time) return ''
  return new Date(time).toLocaleString('zh-CN')
}

const getTypeText = (type) => {
  const map = {
    stock_low: '库存提醒',
    order: '订单通知',
    comment_post: '动态评论',
    comment_knowledge: '知识评论',
    reply_comment: '回复通知',
    at_comment: '@通知',
    new_message: '新私信',
    system: '系统通知'
  }
  return map[type] || '通知'
}

const getTagType = (type) => {
  const map = {
    stock_low: 'warning',
    order: 'primary',
    comment_post: 'success',
    comment_knowledge: 'success',
    reply_comment: 'info',
    at_comment: 'danger',
    new_message: 'primary',
    system: ''
  }
  return map[type] || 'info'
}

const getIconClass = (type) => {
  const map = {
    stock_low: 'stock_low',
    order: 'order',
    comment_post: 'comment',
    comment_knowledge: 'comment',
    reply_comment: 'reply',
    at_comment: 'at',
    new_message: 'message',
    system: 'system'
  }
  return map[type] || 'system'
}

const fetchData = async () => {
  if (!token) {
    ElMessage.warning('请先登录')
    router.push('/login')
    return
  }

  loading.value = true
  try {
    const res = await api.getNotifications({ page: page.value, page_size: pageSize.value })
    notifications.value = res?.list || []
    total.value = res?.total || 0
  } catch (error) {
    console.error('获取通知失败:', error)
  } finally {
    loading.value = false
  }
}

const handleItemClick = async (item) => {
  if (item.is_read === 0) {
    try {
      await api.markAsRead(item.id)
      item.is_read = 1
      // 通知 Header 更新未读数
      window.dispatchEvent(new Event('notification-update'))
    } catch (error) {
      console.error('标记已读失败:', error)
    }
  }

  // 根据关联类型跳转到对应页面
  if (item.related_type === 'product' && item.related_id) {
    router.push(`/product/${item.related_id}`)
  } else if (item.related_type === 'post' && item.related_id) {
    router.push(`/post/${item.related_id}`)
  } else if (item.related_type === 'knowledge' && item.related_id) {
    router.push(`/knowledge/${item.related_id}`)
  } else if (item.related_type === 'conversation' && item.related_id) {
    router.push(`/messages/${item.related_id}`)
  }
}

const handleMarkAllAsRead = async () => {
  try {
    await api.markAllAsRead()
    notifications.value.forEach(item => item.is_read = 1)
    window.dispatchEvent(new Event('notification-update'))
    ElMessage.success('已全部标记为已读')
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

onMounted(fetchData)
</script>

<style scoped>
.notifications-page { min-height: 100vh; background: #f5f5f5; }

.container { max-width: 900px; margin: 0 auto; padding: 0 20px; }
.main { padding: 30px 20px; }

.card-header { display: flex; justify-content: space-between; align-items: center; }

.notification-list { display: flex; flex-direction: column; gap: 10px; }
.notification-item {
  display: flex;
  gap: 15px;
  padding: 15px;
  background: #fff;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid #eee;
}
.notification-item:hover { background: #f9f9f9; }
.notification-item.unread { background: #ecf5ff; border-color: #409eff; }

.notification-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  flex-shrink: 0;
}
.notification-icon.stock_low { background: #fdf6ec; color: #e6a23c; }
.notification-icon.order { background: #ecf5ff; color: #409eff; }
.notification-icon.comment { background: #f0f9eb; color: #67c23a; }
.notification-icon.reply { background: #ecf5ff; color: #409eff; }
.notification-icon.at { background: #fef0f0; color: #f56c6c; }
.notification-icon.system { background: #f4f4f5; color: #909399; }

.notification-content { flex: 1; min-width: 0; }
.notification-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.notification-title { font-weight: 500; color: #333; font-size: 15px; }
.notification-text { color: #666; font-size: 14px; line-height: 1.5; margin-bottom: 8px; }
.notification-time { font-size: 12px; color: #999; }

.pagination-wrapper { display: flex; justify-content: center; margin-top: 20px; }
</style>
