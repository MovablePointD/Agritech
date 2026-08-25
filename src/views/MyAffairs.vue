<template>
  <div class="my-affairs-page">
    <AppHeader />

    <div class="container main">
      <div class="page-header">
        <h1>我的事务</h1>
        <el-button type="primary" @click="$router.push('/affair/submit')">
          <el-icon><Plus /></el-icon> 上报新事务
        </el-button>
      </div>

      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="全部" name="all" />
        <el-tab-pane label="待审核" name="1" />
        <el-tab-pane label="审核通过" name="2" />
        <el-tab-pane label="处理中" name="4" />
        <el-tab-pane label="已完成" name="5" />
      </el-tabs>

      <div class="affairs-list" v-loading="loading">
        <el-card v-for="item in affairs" :key="item.id" class="affair-card" shadow="hover">
          <div class="affair-item" @click="$router.push(`/affair/${item.id}`)">
            <div class="affair-images" v-if="item.images">
              <el-image :src="JSON.parse(item.images)[0]" fit="cover" class="affair-img" />
            </div>
            <div class="affair-info">
              <div class="affair-header">
                <h3 class="affair-title">{{ item.title }}</h3>
                <el-tag :type="getStatusType(item.status)" size="small">{{ getStatusText(item.status) }}</el-tag>
              </div>
              <div class="affair-meta">
                <span class="type-tag">{{ getTypeText(item.type) }}</span>
                <span class="location"><el-icon><Location /></el-icon> {{ item.address }}</span>
                <span class="time">{{ formatTime(item.created_at) }}</span>
              </div>
              <p class="affair-content">{{ item.content.substring(0, 80) }}{{ item.content.length > 80 ? '...' : '' }}</p>
              <div class="affair-footer">
                <span class="status-text">处理人员：{{ item.handler_name || '待分配' }}</span>
                <el-button text size="small" @click.stop="$router.push(`/affair/${item.id}`)">查看详情</el-button>
              </div>
            </div>
          </div>
        </el-card>
        <el-empty v-if="!loading && affairs.length === 0" description="暂无相关事务" />
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
    </div>
  </div>
</template>

<script setup>
      
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Plus, Location } from '@element-plus/icons-vue'
import api from '@/utils/api'
import { getToken } from '@/utils/auth'

const router = useRouter()
const token = getToken()

const affairs = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const activeTab = ref('all')

const getStatusText = (status) => {
  const map = { 1: '待审核', 2: '审核通过', 3: '审核不通过', 4: '处理中', 5: '已完成' }
  return map[status] || '未知'
}

const getStatusType = (status) => {
  const map = { 1: 'warning', 2: 'primary', 3: 'danger', 4: 'info', 5: 'success' }
  return map[status] || 'info'
}

const getTypeText = (type) => {
  const map = { infrastructure: '基础设施', hardware: '硬件问题', env: '环境问题', safety: '安全隐患', other: '其他' }
  return map[type] || '其他'
}

const formatTime = (time) => {
  if (!time) return ''
  return new Date(time).toLocaleString('zh-CN')
}

const handleTabChange = () => {
  page.value = 1
  fetchData()
}

const fetchData = async () => {
  if (!token) {
    ElMessage.warning('请先登录')
    router.push('/login')
    return
  }

  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value }
    if (activeTab.value !== 'all') {
      params.status = activeTab.value
    }
    const res = await api.getMyRuralAffairs(params)
    affairs.value = res?.list || []
    total.value = res?.total || 0
  } catch (error) {
    console.error('获取数据失败:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.my-affairs-page { min-height: 100vh; background: #f5f5f5; }

.container { max-width: 1000px; margin: 0 auto; padding: 0 20px; }
.main { padding: 30px 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-header h1 { font-size: 24px; color: #333; margin: 0; }

.affairs-list { display: flex; flex-direction: column; gap: 15px; margin-top: 20px; }
.affair-card { transition: transform 0.2s; cursor: pointer; }
.affair-card:hover { transform: translateY(-2px); }
.affair-item { display: flex; gap: 20px; }
.affair-images { width: 160px; height: 120px; flex-shrink: 0; border-radius: 8px; overflow: hidden; }
.affair-img { width: 100%; height: 100%; }
.affair-info { flex: 1; min-width: 0; }
.affair-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 10px; }
.affair-title { font-size: 16px; color: #333; margin: 0; flex: 1; }
.affair-meta { display: flex; gap: 15px; align-items: center; margin-bottom: 10px; font-size: 13px; color: #999; flex-wrap: wrap; }
.type-tag { background: #ecf5ff; color: #409eff; padding: 2px 8px; border-radius: 4px; }
.location { display: flex; align-items: center; gap: 4px; }
.affair-content { color: #666; font-size: 14px; line-height: 1.5; margin-bottom: 10px; }
.affair-footer { display: flex; justify-content: space-between; align-items: center; }
.status-text { font-size: 13px; color: #999; }

.pagination-wrapper { display: flex; justify-content: center; margin-top: 30px; }

@media (max-width: 768px) {
  .affair-item { flex-direction: column; }
  .affair-images { width: 100%; height: 160px; }
}
</style>
