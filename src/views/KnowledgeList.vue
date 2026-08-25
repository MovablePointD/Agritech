<template>
  <div class="knowledge-list-page">
    <AppHeader />

    <div class="container main">
      <div class="page-header">
        <h1>农业知识</h1>
        <p class="subtitle">分享专业种植养殖经验，传播农业技术</p>
      </div>

      <div class="filter-bar">
        <el-input v-model="keyword" placeholder="搜索知识标题" clearable @keyup.enter="fetchData" style="width: 300px;">
          <template #append><el-button :icon="Search" @click="fetchData" /></template>
        </el-input>
      </div>

      <div v-loading="loading" class="knowledge-grid">
        <div v-for="item in list" :key="item.id" class="knowledge-card" :class="{ 'pending-card': item.status !== 1 }" @click="$router.push(`/knowledge/${item.id}`)">
          <div class="card-image" v-if="getCoverImage(item)">
            <el-image :src="getCoverImage(item)" fit="cover" />
          </div>
          <div class="card-image card-image-placeholder" v-else>
            <el-icon :size="40"><Reading /></el-icon>
          </div>
          <div class="card-body">
            <h3 class="card-title">{{ item.title }}</h3>
            <p class="card-desc">{{ truncateText(stripHtml(item.content), 80) }}</p>
            <div class="card-footer">
              <div class="author">
                <el-avatar :src="item.user?.avatar_url" :size="24">{{ item.user?.username?.[0] }}</el-avatar>
                <span>{{ item.user?.nickname || item.user?.username || '匿名' }}</span>
                <el-tag v-if="item.status === -1" size="small" type="warning">审核中</el-tag>
                <el-tag v-else-if="item.status === 0" size="small" type="danger">已退回</el-tag>
              </div>
              <span class="date">{{ formatDate(item.created_at) }}</span>
            </div>
          </div>
        </div>

        <div v-if="!loading && list.length === 0" class="empty-state">
          <el-icon :size="80" color="#dcdfe6"><Document /></el-icon>
          <p>暂无相关知识</p>
        </div>
      </div>

      <div class="pagination" v-if="total > 0">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="prev, pager, next"
          @current-change="fetchData"
        />
      </div>
    </div>
    
    <AppFooter />
  </div>
</template>

<script setup>
      
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Reading, Document } from '@element-plus/icons-vue'
import api from '@/utils/api'

const list = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(12)
const total = ref(0)
const keyword = ref('')

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

// 去除HTML标签获取纯文本
const stripHtml = (html) => {
  if (!html) return ''
  const doc = new DOMParser().parseFromString(html, 'text/html')
  return doc.body.textContent || ''
}

// 截断文本
const truncateText = (text, maxLen) => {
  if (!text) return ''
  return text.length > maxLen ? text.slice(0, maxLen) + '...' : text
}

// 获取封面图片：优先使用image_url，否则从HTML内容中提取第一张图片
const getCoverImage = (item) => {
  if (item.image_url) return item.image_url
  if (!item.content) return null
  const match = item.content.match(/<img[^>]+src="([^">]+)"/)
  return match ? match[1] : null
}

const fetchData = async () => {
  loading.value = true
  try {
    const res = await api.getKnowledgeList({
      page: page.value,
      pageSize: pageSize.value,
      keyword: keyword.value
    })
    list.value = res?.data || []
    total.value = res?.total || 0
  } catch (error) {
    list.value = []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.knowledge-list-page { min-height: 100vh; background: #f5f7fa; }
.container { max-width: 1200px; margin: 0 auto; padding: 0 20px; }

.main { padding: 30px 20px; }

.page-header { text-align: center; margin-bottom: 30px; }
.page-header h1 { font-size: 28px; color: #333; margin-bottom: 10px; }
.subtitle { color: #999; font-size: 14px; }

.filter-bar { display: flex; justify-content: center; margin-bottom: 30px; }

.knowledge-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  min-height: 400px;
}

.knowledge-card {
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
  box-shadow: 0 2px 8px rgba(0,0,0,0.06);
  display: flex;
  flex-direction: column;
  min-height: 380px;
}
.knowledge-card.pending-card { opacity: 0.82; border: 2px solid #e6a23c; }
.knowledge-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0,0,0,0.12);
}

.card-image { width: 100%; height: 180px; overflow: hidden; flex-shrink: 0; }
.card-image .el-image { width: 100%; height: 100%; }
.card-image-placeholder {
  background: linear-gradient(135deg, #66bb6a 0%, #43a047 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}

.card-body { padding: 16px; display: flex; flex-direction: column; flex: 1; }
.card-title {
  font-size: 16px;
  color: #333;
  margin-bottom: 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.card-desc {
  color: #666; font-size: 13px; line-height: 1.6;
  margin-bottom: 12px; min-height: 42px;
  overflow: hidden; text-overflow: ellipsis;
  display: -webkit-box; -webkit-line-clamp: 3; -webkit-box-orient: vertical;
}

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
  margin-top: auto;
}
.author { display: flex; align-items: center; gap: 6px; font-size: 12px; color: #666; flex-wrap: wrap; }
.author span { max-width: 80px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.date { font-size: 12px; color: #999; white-space: nowrap; }

.empty-state {
  grid-column: 1 / -1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px;
  color: #909399;
}
.empty-state p { margin-top: 20px; font-size: 14px; }

.pagination { display: flex; justify-content: center; margin-top: 30px; }
</style>
