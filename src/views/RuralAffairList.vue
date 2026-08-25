<template>
  <div class="rural-affairs-page">
    <AppHeader />

    <div class="container main">
      <div class="page-header">
        <h1>农村事务大厅</h1>
        <p class="subtitle">您可以通过本平台上报农村各类问题，我们会及时处理并反馈</p>
      </div>

      <div class="filters">
        <el-select v-model="filterType" placeholder="事务类型" clearable @change="handleFilter">
          <el-option label="全部类型" value="" />
          <el-option label="基础设施" value="infrastructure" />
          <el-option label="硬件问题" value="hardware" />
          <el-option label="环境问题" value="env" />
          <el-option label="安全隐患" value="safety" />
          <el-option label="其他" value="other" />
        </el-select>
        <el-select v-model="filterStatus" placeholder="处理状态" clearable @change="handleFilter">
          <el-option label="全部状态" value="" />
          <el-option label="待审核" value="1" />
          <el-option label="审核通过" value="2" />
          <el-option label="审核不通过" value="3" />
          <el-option label="处理中" value="4" />
          <el-option label="已完成" value="5" />
        </el-select>
        <el-input v-model="keyword" placeholder="搜索关键词" style="width: 200px;" clearable @change="handleFilter">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-button type="primary" @click="handleFilter">搜索</el-button>
        <el-button v-if="token && !myProcessor" type="success" @click="showProcessorDialog = true">申请成为处理人员</el-button>
        <el-button v-if="myProcessor && myProcessor.status === 2" type="warning" @click="showProcessorDialog = true">重新申请</el-button>
        <div class="filters-spacer"></div>
        <el-button v-if="token" type="primary" @click="$router.push('/affair/submit')">
          <el-icon><Plus /></el-icon>
          上报事务
        </el-button>
      </div>
      <div class="filter-hint" v-if="!filterStatus && !filterType && !keyword">
        <el-icon><InfoFilled /></el-icon> 当前显示所有状态的事务，您可以使用上方筛选器按类型或状态筛选
      </div>

      <!-- 我的处理人员申请状态 -->
      <el-card v-if="myProcessor" style="margin-bottom: 20px;">
        <template #header>
          <span>我的处理人员申请状态</span>
          <el-tag v-if="isProcessor" type="success" size="small" style="margin-left: 10px;">已认证</el-tag>
        </template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="真实姓名">{{ myProcessor.real_name }}</el-descriptions-item>
          <el-descriptions-item label="联系电话">{{ myProcessor.phone }}</el-descriptions-item>
          <el-descriptions-item label="负责区域">{{ myProcessor.address }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getProcessorStatusType(myProcessor.status)">{{ getProcessorStatusText(myProcessor.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="个人简介" :span="2">{{ myProcessor.intro }}</el-descriptions-item>
          <el-descriptions-item v-if="myProcessor.reject_reason" label="拒绝原因" :span="2">
            <span style="color: #f56c6c;">{{ myProcessor.reject_reason }}</span>
          </el-descriptions-item>
        </el-descriptions>
        <div v-if="isProcessor" style="margin-top: 15px;">
          <el-button type="primary" @click="$router.push('/processor')">进入处理界面</el-button>
        </div>
      </el-card>

      <div class="stats-cards">
        <div class="stat-card" @click="filterStatus = '1'; handleFilter()">
          <div class="stat-icon pending"><el-icon><Clock /></el-icon></div>
          <div class="stat-info">
            <span class="stat-num">{{ stats.pending }}</span>
            <span class="stat-label">待审核</span>
          </div>
        </div>
        <div class="stat-card" @click="filterStatus = '2'; handleFilter()">
          <div class="stat-icon approved"><el-icon><CircleCheck /></el-icon></div>
          <div class="stat-info">
            <span class="stat-num">{{ stats.approved }}</span>
            <span class="stat-label">审核通过</span>
          </div>
        </div>
        <div class="stat-card" @click="filterStatus = '4'; handleFilter()">
          <div class="stat-icon processing"><el-icon><Setting /></el-icon></div>
          <div class="stat-info">
            <span class="stat-num">{{ stats.processing }}</span>
            <span class="stat-label">处理中</span>
          </div>
        </div>
        <div class="stat-card" @click="filterStatus = '5'; handleFilter()">
          <div class="stat-icon completed"><el-icon><CircleCheckFilled /></el-icon></div>
          <div class="stat-info">
            <span class="stat-num">{{ stats.completed }}</span>
            <span class="stat-label">已完成</span>
          </div>
        </div>
      </div>

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
              <p class="affair-content">{{ item.content.substring(0, 100) }}{{ item.content.length > 100 ? '...' : '' }}</p>
              <div class="affair-footer">
                <span class="author">
                  <el-avatar :src="item.user?.avatar_url" :size="20">{{ item.user?.username?.[0] }}</el-avatar>
                  {{ item.user?.nickname || item.user?.username }}
                </span>
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

    <el-button type="primary" class="fab" @click="$router.push('/affair/submit')" v-if="token">
      <el-icon><Plus /></el-icon>
      <span class="fab-text">上报事务</span>
    </el-button>

    <!-- 处理人员申请对话框 -->
    <el-dialog v-model="showProcessorDialog" title="申请成为事务处理人员" width="600px">
      <el-form :model="processorForm" label-width="100px">
        <el-form-item label="真实姓名" required>
          <el-input v-model="processorForm.real_name" placeholder="请输入真实姓名" />
        </el-form-item>
        <el-form-item label="联系电话" required>
          <el-input v-model="processorForm.phone" placeholder="请输入联系电话" />
        </el-form-item>
        <el-form-item label="负责区域" required>
          <el-input v-model="processorForm.address" placeholder="请输入负责区域" />
        </el-form-item>
        <el-form-item label="个人简介" required>
          <el-input v-model="processorForm.intro" type="textarea" :rows="4" placeholder="请输入个人简介" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showProcessorDialog = false">取消</el-button>
        <el-button type="primary" :loading="applying" @click="submitProcessorApply">提交申请</el-button>
      </template>
    </el-dialog>

    <AppFooter />
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { Clock, CircleCheck, CircleCheckFilled, Setting, Search, Location, Plus, InfoFilled } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import api from '@/utils/api'
import { getToken } from '@/utils/auth'

const token = getToken()

const affairs = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const keyword = ref('')
const filterType = ref('')
const filterStatus = ref('')

const stats = ref({ pending: 0, approved: 0, processing: 0, completed: 0 })

// 处理人员申请
const myProcessor = ref(null)
const isProcessor = ref(false)
const showProcessorDialog = ref(false)
const applying = ref(false)
const processorForm = ref({
  real_name: '',
  phone: '',
  address: '',
  intro: ''
})

const getStatusText = (status) => {
  const map = { 1: '待审核', 2: '审核通过', 3: '审核不通过', 4: '处理中', 5: '已完成' }
  return map[status] || '未知'
}

const getStatusType = (status) => {
  const map = { 1: 'warning', 2: 'primary', 3: 'danger', 4: 'info', 5: 'success' }
  return map[status] || 'info'
}

const getProcessorStatusText = (status) => {
  const map = { 0: '待审核', 1: '已通过', 2: '已拒绝' }
  return map[status] || '未知'
}

const getProcessorStatusType = (status) => {
  const map = { 0: 'warning', 1: 'success', 2: 'danger' }
  return map[status] || ''
}

const getTypeText = (type) => {
  const map = { infrastructure: '基础设施', hardware: '硬件问题', env: '环境问题', safety: '安全隐患', other: '其他' }
  return map[type] || '其他'
}

const formatTime = (time) => {
  if (!time) return ''
  return new Date(time).toLocaleString('zh-CN')
}

const handleFilter = () => {
  page.value = 1
  fetchData()
}

const fetchData = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: keyword.value,
      type: filterType.value,
      status: filterStatus.value
    }
    const res = await api.getRuralAffairs(params)
    affairs.value = res?.list || []
    total.value = res?.total || 0
  } catch (error) {
    console.error('获取数据失败:', error)
  } finally {
    loading.value = false
  }
}

const fetchStats = async () => {
  try {
    const [p1, p2, p4, p5] = await Promise.allSettled([
      api.getRuralAffairs({ status: '1', page_size: 1 }),
      api.getRuralAffairs({ status: '2', page_size: 1 }),
      api.getRuralAffairs({ status: '4', page_size: 1 }),
      api.getRuralAffairs({ status: '5', page_size: 1 })
    ])
    stats.value.pending = p1.status === 'fulfilled' ? p1.value?.total || 0 : 0
    stats.value.approved = p2.status === 'fulfilled' ? p2.value?.total || 0 : 0
    stats.value.processing = p4.status === 'fulfilled' ? p4.value?.total || 0 : 0
    stats.value.completed = p5.status === 'fulfilled' ? p5.value?.total || 0 : 0
  } catch (error) {
    console.error('获取统计数据失败:', error)
  }
}

const loadMyProcessor = async () => {
  try {
    const res = await api.getMyProcessor()
    if (res) {
      myProcessor.value = res
      isProcessor.value = res.status === 1
    }
  } catch (error) {
    myProcessor.value = null
    isProcessor.value = false
  }
}

const submitProcessorApply = async () => {
  if (!processorForm.value.real_name || !processorForm.value.phone || !processorForm.value.address || !processorForm.value.intro) {
    ElMessage.warning('请填写必填项')
    return
  }
  applying.value = true
  try {
    await api.applyProcessor(processorForm.value)
    ElMessage.success('申请已提交')
    showProcessorDialog.value = false
    loadMyProcessor()
  } catch (error) {
    ElMessage.error('申请失败')
  } finally {
    applying.value = false
  }
}

onMounted(() => {
  fetchData()
  fetchStats()
  if (token) {
    loadMyProcessor()
  }
})
</script>

<style scoped>
.rural-affairs-page { min-height: 100vh; background: #f5f5f5; }

.container { max-width: 1200px; margin: 0 auto; padding: 0 20px; }
.main { padding: 30px 20px; }
.page-header { text-align: center; margin-bottom: 30px; }
.page-header h1 { font-size: 28px; color: #333; margin-bottom: 10px; }
.subtitle { color: #666; font-size: 14px; }

.filters { display: flex; gap: 15px; margin-bottom: 25px; flex-wrap: wrap; align-items: center; }
.filters-spacer { flex: 1; }
.filter-hint { 
  display: flex; align-items: center; gap: 6px; 
  font-size: 13px; color: #909399; 
  margin-bottom: 25px; padding: 8px 14px; 
  background: #f4f4f5; border-radius: 6px; 
}

.stats-cards { display: grid; grid-template-columns: repeat(4, 1fr); gap: 15px; margin-bottom: 30px; }
.stat-card { background: #fff; border-radius: 12px; padding: 20px; display: flex; align-items: center; gap: 15px; cursor: pointer; transition: transform 0.2s, box-shadow 0.2s; }
.stat-card:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,0.1); }
.stat-icon { width: 50px; height: 50px; border-radius: 12px; display: flex; align-items: center; justify-content: center; font-size: 24px; }
.stat-icon.pending { background: #fdf6ec; color: #e6a23c; }
.stat-icon.approved { background: #ecf5ff; color: #409eff; }
.stat-icon.processing { background: #f4f4f5; color: #909399; }
.stat-icon.completed { background: #f0f9eb; color: #67c23a; }
.stat-info { display: flex; flex-direction: column; }
.stat-num { font-size: 28px; font-weight: bold; color: #333; }
.stat-label { font-size: 13px; color: #999; }

.affairs-list { display: flex; flex-direction: column; gap: 15px; }
.affair-card { transition: transform 0.2s; cursor: pointer; }
.affair-card:hover { transform: translateY(-2px); }
.affair-item { display: flex; gap: 20px; }
.affair-images { width: 200px; height: 140px; flex-shrink: 0; border-radius: 8px; overflow: hidden; }
.affair-img { width: 100%; height: 100%; }
.affair-info { flex: 1; min-width: 0; }
.affair-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 10px; }
.affair-title { font-size: 18px; color: #333; margin: 0; flex: 1; }
.affair-meta { display: flex; gap: 15px; align-items: center; margin-bottom: 10px; font-size: 13px; color: #999; }
.type-tag { background: #ecf5ff; color: #409eff; padding: 2px 8px; border-radius: 4px; }
.location { display: flex; align-items: center; gap: 4px; }
.affair-content { color: #666; font-size: 14px; line-height: 1.6; margin-bottom: 10px; }
.affair-footer { display: flex; justify-content: space-between; align-items: center; }
.author { display: flex; align-items: center; gap: 8px; font-size: 13px; color: #666; }

.pagination-wrapper { display: flex; justify-content: center; margin-top: 30px; }

.fab { position: fixed; bottom: 80px; right: 30px; width: auto; height: 48px; border-radius: 24px; padding: 0 20px; font-size: 15px; box-shadow: 0 4px 16px rgba(64,158,255,0.45); z-index: 100; }
.fab-text { margin-left: 4px; white-space: nowrap; }

@media (max-width: 768px) {
  .stats-cards { grid-template-columns: repeat(2, 1fr); }
  .affair-item { flex-direction: column; }
  .affair-images { width: 100%; height: 180px; }
}
</style>
