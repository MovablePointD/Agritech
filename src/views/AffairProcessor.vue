<template>
  <div class="processor-page">
    <AppHeader />

    <div class="container main">
      <div class="page-header">
        <h1>事务处理中心</h1>
        <p class="subtitle">优先展示与您负责区域相近的事务</p>
      </div>

      <!-- 处理人员申请状态 / 申请入口 -->
      <div v-if="!myProcessor" class="apply-section">
        <el-button type="primary" size="large" @click="showProcessorApply = true" :icon="Plus">申请成为事务处理人员</el-button>
      </div>
      <el-card v-else-if="myProcessor.status !== 1" class="processor-status-card">
        <template #header><span>我的申请状态</span></template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="真实姓名">{{ myProcessor.real_name }}</el-descriptions-item>
          <el-descriptions-item label="联系电话">{{ myProcessor.phone }}</el-descriptions-item>
          <el-descriptions-item label="负责区域">{{ myProcessor.address }}</el-descriptions-item>
          <el-descriptions-item label="申请状态">
            <el-tag :type="myProcessor.status === 0 ? 'warning' : 'danger'">{{ myProcessor.status === 0 ? '待审核' : '已拒绝' }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item v-if="myProcessor.reject_reason" label="拒绝原因" :span="2">
            <span style="color: #f56c6c;">{{ myProcessor.reject_reason }}</span>
          </el-descriptions-item>
        </el-descriptions>
        <div style="margin-top: 12px;">
          <el-button v-if="myProcessor.status === 2" type="warning" @click="showProcessorApply = true">重新申请</el-button>
          <el-button v-if="myProcessor.status === 0" type="primary" disabled>审核中</el-button>
        </div>
      </el-card>

      <!-- 统计卡片 -->
      <div class="stats-cards">
        <div class="stat-card pending">
          <div class="stat-icon"><el-icon><Clock /></el-icon></div>
          <div class="stat-info">
            <span class="stat-num">{{ stats.pending }}</span>
            <span class="stat-label">待处理</span>
          </div>
        </div>
        <div class="stat-card processing">
          <div class="stat-icon"><el-icon><Setting /></el-icon></div>
          <div class="stat-info">
            <span class="stat-num">{{ stats.processing }}</span>
            <span class="stat-label">处理中</span>
          </div>
        </div>
        <div class="stat-card completed">
          <div class="stat-icon"><el-icon><CircleCheck /></el-icon></div>
          <div class="stat-info">
            <span class="stat-num">{{ stats.completed }}</span>
            <span class="stat-label">已完成</span>
          </div>
        </div>
        <div class="stat-card region">
          <div class="stat-icon"><el-icon><Location /></el-icon></div>
          <div class="stat-info">
            <span class="stat-num region-name">{{ myProcessor?.address || '未设置' }}</span>
            <span class="stat-label">负责区域</span>
          </div>
        </div>
      </div>

      <!-- 筛选 -->
      <div class="filters">
        <el-button :type="showMyAffairs ? 'success' : 'default'" @click="toggleMyAffairs">
          {{ showMyAffairs ? '查看我的事务' : '查看我的事务' }}
        </el-button>
        <el-button v-if="showMyAffairs" @click="loadRecommended">推荐列表</el-button>
        <el-select v-if="!showMyAffairs" v-model="filterStatus" placeholder="处理状态" clearable @change="handleFilter">
          <el-option label="全部状态" value="" />
          <el-option label="待处理" value="2" />
          <el-option label="处理中" value="4" />
          <el-option label="已完成" value="5" />
        </el-select>
        <el-select v-if="!showMyAffairs" v-model="filterType" placeholder="事务类型" clearable @change="handleFilter">
          <el-option label="全部类型" value="" />
          <el-option label="基础设施" value="infrastructure" />
          <el-option label="硬件问题" value="hardware" />
          <el-option label="环境问题" value="env" />
          <el-option label="安全隐患" value="safety" />
          <el-option label="其他" value="other" />
        </el-select>
        <el-input v-if="!showMyAffairs" v-model="keyword" placeholder="搜索关键词" style="width: 200px;" clearable @keyup.enter="handleFilter">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-button v-if="!showMyAffairs" type="primary" @click="handleFilter">搜索</el-button>
      </div>

      <!-- 我的事务模块（放大展示） -->
      <div v-if="showMyAffairs" class="my-affairs-section">
        <div class="section-header">
          <h2>我处理的事务</h2>
          <div class="filter-tabs">
            <el-radio-group v-model="myAffairsFilter" size="default" @change="filterMyAffairs">
              <el-radio-button label="pending">待处理 ({{ stats.pending }})</el-radio-button>
              <el-radio-button label="processing">处理中 ({{ stats.processing }})</el-radio-button>
              <el-radio-button label="followup">追问中 ({{ myAffairsAll.filter(a => a.status === 6).length }})</el-radio-button>
              <el-radio-button label="completed">已完成 ({{ stats.completed }})</el-radio-button>
              <el-radio-button label="appealed">申诉中 ({{ myAffairsAll.filter(a => a.status === 7).length }})</el-radio-button>
              <el-radio-button label="">全部 ({{ myAffairsAll.length }})</el-radio-button>
            </el-radio-group>
          </div>
        </div>

        <div class="my-affairs-list" v-loading="loadingMyAffairs">
          <el-card v-for="item in filteredMyAffairs" :key="item.id" class="my-affair-card" shadow="hover">
            <div class="my-affair-content">
              <div class="affair-left">
                <div class="affair-header">
                  <h3 class="affair-title">{{ item.title }}</h3>
                  <el-tag :type="getStatusType(item.status)" size="default">{{ getStatusText(item.status) }}</el-tag>
                </div>
                <div class="affair-meta">
                  <span class="type-tag">{{ getTypeText(item.type) }}</span>
                  <span class="location"><el-icon><Location /></el-icon> {{ item.address }}</span>
                  <span class="submitter">
                    <el-avatar :src="item.user?.avatar_url" :size="20">{{ item.user?.username?.[0] }}</el-avatar>
                    {{ item.user?.nickname || item.user?.username }}
                  </span>
                  <span class="time">{{ formatTime(item.created_at) }}</span>
                </div>
                <p class="affair-content">{{ item.content }}</p>
                <div v-if="item.process_content" class="process-result">
                  <el-divider content-position="left">处理结果</el-divider>
                  <p>{{ item.process_content }}</p>
                  <div v-if="item.process_images" class="process-images">
                    <el-image
                      v-for="(img, idx) in JSON.parse(item.process_images)"
                      :key="idx"
                      :src="img"
                      fit="cover"
                      class="process-img"
                      :preview-src-list="JSON.parse(item.process_images)"
                    />
                  </div>
                </div>
              </div>
              <div class="affair-right">
                <el-button type="primary" @click="openDetail(item)">详情</el-button>
                <el-button v-if="item.status === 2" type="success" @click="startProcess(item)">开始处理</el-button>
                <el-button v-if="item.status === 4 && item.handler_id === currentUserId" type="warning" @click="openProcessDialog(item)">填写处理记录</el-button>
                <el-button v-if="item.status === 6 && item.handler_id === currentUserId" type="primary" @click="openAnswerDialog(item)">追答</el-button>
                <el-button v-if="(item.status === 4 || item.status === 6) && item.handler_id === currentUserId && item.process_content" type="danger" plain size="small" @click="openHandlerAppeal(item)">申诉</el-button>
                <el-button v-if="item.status === 7" type="info" disabled>申诉处理中</el-button>
              </div>
            </div>
          </el-card>
          <el-empty v-if="filteredMyAffairs.length === 0" description="暂无相关事务" />
        </div>
      </div>

      <!-- 推荐说明 -->
      <div v-if="showRecommendation && !showMyAffairs" class="recommendation-tip">
        <el-icon><InfoFilled /></el-icon>
        <span>当前显示的是根据您负责区域推荐的优先处理事务（地址匹配优先，然后按时间由老到新排序）</span>
      </div>

      <!-- 事务列表（推荐列表） -->
      <div v-if="!showMyAffairs" class="affairs-list" v-loading="loading">
        <el-card v-for="item in affairs" :key="item.id" class="affair-card" :class="{ 'matched': item.matchLevel === 'high' }" shadow="hover">
          <div class="affair-item">
            <div class="affair-images" v-if="item.images">
              <el-image :src="JSON.parse(item.images)[0]" fit="cover" class="affair-img" />
            </div>
            <div class="affair-info">
              <div class="affair-header">
                <h3 class="affair-title">{{ item.title }}</h3>
                <el-tag :type="getStatusType(item.status)" size="small">{{ getStatusText(item.status) }}</el-tag>
                <el-tag v-if="item.matchLevel === 'high'" type="success" size="small" effect="plain">区域匹配</el-tag>
              </div>
              <div class="affair-meta">
                <span class="type-tag">{{ getTypeText(item.type) }}</span>
                <span class="location"><el-icon><Location /></el-icon> {{ item.address }}</span>
                <span class="time">{{ formatTime(item.created_at) }}</span>
                <span class="submitter">
                  <el-avatar :src="item.user?.avatar_url" :size="16">{{ item.user?.username?.[0] }}</el-avatar>
                  {{ item.user?.nickname || item.user?.username }}
                </span>
              </div>
              <p class="affair-content">{{ item.content?.substring(0, 120) }}{{ item.content?.length > 120 ? '...' : '' }}</p>
              <div class="affair-actions">
                <el-button type="primary" size="small" @click="openDetail(item)">查看详情</el-button>
                <el-button v-if="item.status === 2" type="success" size="small" @click="startProcess(item)">开始处理</el-button>
                <el-button v-if="item.status === 4 && item.handler_id === currentUserId" type="warning" size="small" @click="openProcessDialog(item)">填写处理记录</el-button>
                <el-button v-if="item.status === 6 && item.handler_id === currentUserId" type="primary" size="small" @click="openAnswerDialog(item)">追答</el-button>
                <el-button v-if="item.status === 7" type="info" size="small" disabled>申诉处理中</el-button>
              </div>
            </div>
          </div>
        </el-card>
        <el-empty v-if="!loading && affairs.length === 0" description="暂无待处理事务" />
      </div>

      <div class="pagination-wrapper" v-if="total > 0 && !showMyAffairs">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="prev, pager, next"
          @current-change="fetchData"
        />
      </div>
    </div>

    <!-- 处理结果对话框 -->
    <el-dialog v-model="showProcessDialog" title="填写处理结果" width="600px">
      <el-form :model="processForm" label-width="100px">
        <el-form-item label="处理详情" required>
          <el-input v-model="processForm.content" type="textarea" :rows="6" placeholder="请详细描述处理过程和结果" />
        </el-form-item>
        <el-form-item label="处理图片">
          <el-upload
            action="/api/upload"
            :headers="{ Authorization: tokenValue }"
            list-type="picture-card"
            :on-success="handleProcessImageSuccess"
            :before-upload="beforeProcessUpload"
            :on-remove="handleProcessImageRemove"
            :file-list="processImageList"
          >
            <el-icon><Plus /></el-icon>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showProcessDialog = false">取消</el-button>
        <el-button type="primary" @click="submitProcess" :loading="processing">提交处理结果</el-button>
      </template>
    </el-dialog>

    <!-- 申请成为处理人员对话框 -->
    <el-dialog v-model="showProcessorApply" title="申请成为事务处理人员" width="600px">
      <el-form :model="processorApplyForm" label-width="100px">
        <el-form-item label="真实姓名" required>
          <el-input v-model="processorApplyForm.real_name" placeholder="请输入真实姓名" />
        </el-form-item>
        <el-form-item label="联系电话" required>
          <el-input v-model="processorApplyForm.phone" placeholder="请输入联系电话" />
        </el-form-item>
        <el-form-item label="负责区域" required>
          <el-input v-model="processorApplyForm.address" placeholder="请输入您负责的区域地址" />
        </el-form-item>
        <el-form-item label="申请理由" required>
          <el-input v-model="processorApplyForm.intro" type="textarea" :rows="4" placeholder="请说明您的能力和经验" />
        </el-form-item>
        <el-form-item label="证明材料">
          <el-upload
            action="/api/upload"
            :headers="{ Authorization: tokenValue }"
            list-type="picture-card"
            :on-success="handleProcessorImageSuccess"
            :before-upload="beforeProcessorUpload"
            :on-remove="handleProcessorImageRemove"
            :file-list="processorImageList"
            :limit="6"
          >
            <el-icon><Plus /></el-icon>
          </el-upload>
          <span class="upload-tip">上传相关资质证明或工作材料，最多6张</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showProcessorApply = false">取消</el-button>
        <el-button type="primary" :loading="processorApplying" @click="submitProcessorApply">提交申请</el-button>
      </template>
    </el-dialog>

    <AppFooter />
  </div>

  <!-- 追答对话框 -->
  <el-dialog v-model="showAnswerDialog" title="回复追问" width="500px">
    <el-input v-model="answerContent" type="textarea" :rows="4" placeholder="请输入追答内容" />
    <p style="color: #999; font-size: 12px; margin-top: 5px;">正在处理事务「{{ currentAffair?.title }}」的追问</p>
    <template #footer>
      <el-button @click="showAnswerDialog = false">取消</el-button>
      <el-button type="primary" @click="submitAnswer" :loading="processing">提交追答</el-button>
    </template>
  </el-dialog>

  <!-- 申诉弹窗 -->
  <el-dialog v-model="showAppealDialog" title="发起申诉" width="500px">
    <el-input v-model="appealReason" type="textarea" :rows="4" placeholder="请说明申诉理由" />
    <p style="color: #999; font-size: 12px; margin-top: 5px;">申诉事务「{{ currentAffair?.title }}」，将由管理员审核</p>
    <template #footer>
      <el-button @click="showAppealDialog = false">取消</el-button>
      <el-button type="danger" @click="submitHandlerAppeal" :loading="processing">提交申诉</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Clock, CircleCheck, Setting, Location, Search, Plus, InfoFilled } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import api from '@/utils/api'
import { getToken, getUserInfo } from '@/utils/auth'

const router = useRouter()

const tokenValue = computed(() => getToken() || '')
const currentUserId = computed(() => getUserInfo()?.id)

const affairs = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const keyword = ref('')
const filterStatus = ref('')
const filterType = ref('')
const showRecommendation = ref(false)
const showMyAffairs = ref(false)

// 我的事务相关
const myAffairsFilter = ref('pending')
const myAffairsAll = ref([])
const loadingMyAffairs = ref(false)
const filteredMyAffairs = computed(() => {
  if (myAffairsFilter.value === '') {
    return myAffairsAll.value
  }
  const statusMap = { pending: 2, processing: 4, followup: 6, completed: 5, appealed: 7 }
  return myAffairsAll.value.filter(a => a.status === statusMap[myAffairsFilter.value])
})

const stats = ref({ pending: 0, processing: 0, completed: 0 })
const myProcessor = ref(null)

// 处理人员申请
const showProcessorApply = ref(false)
const processorApplying = ref(false)
const processorApplyForm = ref({
  real_name: '',
  phone: '',
  address: '',
  intro: '',
  cert_images: ''
})
const processorImageList = ref([])
const processorCertImagesArr = ref([])

const beforeProcessorUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) ElMessage.error('只能上传图片')
  if (!isLt5M) ElMessage.error('图片大小不能超过5MB')
  return isImage && isLt5M
}

const handleProcessorImageSuccess = (res) => {
  if (res.url) {
    processorCertImagesArr.value.push(res.url)
    processorApplyForm.value.cert_images = JSON.stringify(processorCertImagesArr.value)
  }
}

const handleProcessorImageRemove = (file) => {
  const url = file.url || file.response?.url
  processorCertImagesArr.value = processorCertImagesArr.value.filter(u => u !== url)
  processorApplyForm.value.cert_images = processorCertImagesArr.value.length > 0 ? JSON.stringify(processorCertImagesArr.value) : ''
}

const submitProcessorApply = async () => {
  const f = processorApplyForm.value
  if (!f.real_name || !f.phone || !f.address || !f.intro) {
    ElMessage.warning('请填写必填项')
    return
  }
  processorApplying.value = true
  try {
    await api.applyProcessor(processorApplyForm.value)
    ElMessage.success('申请已提交')
    showProcessorApply.value = false
    // 重置
    processorApplyForm.value = { real_name: '', phone: '', address: '', intro: '', cert_images: '' }
    processorImageList.value = []
    processorCertImagesArr.value = []
    loadMyProcessor()
  } catch (error) {
    ElMessage.error(error?.response?.data?.error || '申请失败')
  } finally {
    processorApplying.value = false
  }
}

// 处理对话框
const showProcessDialog = ref(false)
const currentAffair = ref(null)
const processing = ref(false)
const processForm = reactive({
  content: '',
  images: ''
})
const processImageList = ref([])

// 追答相关
const showAnswerDialog = ref(false)
const answerContent = ref('')

// 申诉相关
const showAppealDialog = ref(false)
const appealReason = ref('')

const getStatusText = (status) => ({ 1: '待审核', 2: '待处理', 3: '审核不通过', 4: '处理中', 5: '已完成', 6: '追问中', 7: '申诉中' }[status] || '未知')
const getStatusType = (status) => ({ 1: 'warning', 2: 'primary', 3: 'danger', 4: 'info', 5: 'success', 6: 'warning', 7: 'danger' }[status] || 'info')
const getTypeText = (type) => ({ infrastructure: '基础设施', hardware: '硬件问题', env: '环境问题', safety: '安全隐患', other: '其他' }[type] || '其他')

const formatTime = (time) => {
  if (!time) return ''
  return new Date(time).toLocaleString('zh-CN')
}

const handleFilter = () => {
  page.value = 1
  showRecommendation.value = false
  showMyAffairs.value = false
  fetchData()
}

const loadRecommended = () => {
  page.value = 1
  showRecommendation.value = true
  showMyAffairs.value = false
  fetchRecommendedData()
}

const toggleMyAffairs = () => {
  page.value = 1
  showMyAffairs.value = !showMyAffairs.value
  if (showMyAffairs.value) {
    showRecommendation.value = false
    myAffairsFilter.value = 'pending'
    loadMyHandledAffairs()
  } else {
    loadRecommended()
  }
}

const filterMyAffairs = () => {
  // 筛选变化时不需要额外操作，computed会自动计算
}

const loadMyHandledAffairs = async () => {
  loadingMyAffairs.value = true
  try {
    const res = await api.getMyHandledAffairs({ page: 1, page_size: 100 })
    myAffairsAll.value = res?.list || []
    // 更新统计数据
    updateMyStats()
  } catch (error) {
    console.error('获取我的事务失败:', error)
    myAffairsAll.value = []
  } finally {
    loadingMyAffairs.value = false
  }
}

const updateMyStats = () => {
  stats.value.pending = myAffairsAll.value.filter(a => a.status === 2).length
  stats.value.processing = myAffairsAll.value.filter(a => a.status === 4).length
  stats.value.completed = myAffairsAll.value.filter(a => a.status === 5).length
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
    affairs.value.forEach(a => a.matchLevel = 'none')
  } catch (error) {
    console.error('获取数据失败:', error)
  } finally {
    loading.value = false
  }
}

const fetchRecommendedData = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value
    }
    const res = await api.getRecommendedAffairs(params)
    // 根据请求结果设置匹配级别
    affairs.value = (res?.list || []).map(a => ({
      ...a,
      matchLevel: a.matchLevel || 'none'
    }))
    total.value = res?.total || 0
  } catch (error) {
    console.error('获取推荐数据失败:', error)
    // 如果推荐接口失败，尝试获取普通列表
    fetchData()
  } finally {
    loading.value = false
  }
}

const fetchStats = async () => {
  try {
    const [pending, processing, completed] = await Promise.allSettled([
      api.getRuralAffairs({ status: '2', page_size: 1 }),
      api.getRuralAffairs({ status: '4', page_size: 1 }),
      api.getRuralAffairs({ status: '5', page_size: 1 })
    ])
    stats.value.pending = pending.status === 'fulfilled' ? pending.value?.total || 0 : 0
    stats.value.processing = processing.status === 'fulfilled' ? processing.value?.total || 0 : 0
    stats.value.completed = completed.status === 'fulfilled' ? completed.value?.total || 0 : 0
  } catch (error) {
    console.error('获取统计数据失败:', error)
  }
}

const loadMyProcessor = async () => {
  try {
    const res = await api.getMyProcessor()
    myProcessor.value = res || null
  } catch (error) {
    myProcessor.value = null
  }
}

const openDetail = (affair) => {
  router.push(`/affair/${affair.id}`)
}

const startProcess = async (affair) => {
  try {
    await api.startProcessRuralAffair(affair.id)
    ElMessage.success('已开始处理该事务')
    fetchData()
    fetchStats()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

const openProcessDialog = (affair) => {
  currentAffair.value = affair
  processForm.content = ''
  processForm.images = ''
  processImageList.value = []
  showProcessDialog.value = true
}

const beforeProcessUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) ElMessage.error('只能上传图片')
  if (!isLt5M) ElMessage.error('图片大小不能超过5MB')
  return isImage && isLt5M
}

const handleProcessImageSuccess = (res) => {
  const newUrl = res.url
  if (processForm.images) {
    try {
      const currentImages = JSON.parse(processForm.images)
      processForm.images = JSON.stringify([...currentImages, newUrl])
    } catch {
      processForm.images = JSON.stringify([newUrl])
    }
  } else {
    processForm.images = JSON.stringify([newUrl])
  }
}

const handleProcessImageRemove = () => {
  processForm.images = ''
}

const submitProcess = async () => {
  if (!processForm.content) {
    ElMessage.warning('请填写处理详情')
    return
  }
  processing.value = true
  try {
    await api.processRuralAffair(currentAffair.value.id, {
      process_content: processForm.content,
      process_images: processForm.images
    })
    ElMessage.success('处理记录已提交，等待用户确认完成')
    showProcessDialog.value = false
    if (showMyAffairs.value) {
      loadMyHandledAffairs()
    } else {
      fetchData()
    }
    fetchStats()
  } catch (error) {
    ElMessage.error('提交失败')
  } finally {
    processing.value = false
  }
}

// 追答
const openAnswerDialog = (affair) => {
  currentAffair.value = affair
  answerContent.value = ''
  showAnswerDialog.value = true
}

const submitAnswer = async () => {
  if (!answerContent.value.trim()) {
    ElMessage.warning('追答内容不能为空')
    return
  }
  processing.value = true
  try {
    await api.addFollowUpAnswer(currentAffair.value.id, {
      content: answerContent.value,
      images: ''
    })
    ElMessage.success('追答成功')
    showAnswerDialog.value = false
    if (showMyAffairs.value) {
      loadMyHandledAffairs()
    } else {
      fetchData()
    }
    fetchStats()
  } catch (e) {
    ElMessage.error('追答失败')
  } finally {
    processing.value = false
  }
}

// 处理人员申诉
const openHandlerAppeal = (affair) => {
  currentAffair.value = affair
  appealReason.value = ''
  showAppealDialog.value = true
}

const submitHandlerAppeal = async () => {
  if (!appealReason.value.trim()) {
    ElMessage.warning('申诉理由不能为空')
    return
  }
  processing.value = true
  try {
    await api.createAppeal(currentAffair.value.id, {
      applicant_type: 'handler',
      reason: appealReason.value,
      images: ''
    })
    ElMessage.success('申诉已提交，等待管理员审核')
    showAppealDialog.value = false
    if (showMyAffairs.value) {
      loadMyHandledAffairs()
    } else {
      fetchData()
    }
    fetchStats()
  } catch (e) {
    ElMessage.error('申诉提交失败')
  } finally {
    processing.value = false
  }
}

onMounted(() => {
  loadMyProcessor().then(() => {
    loadRecommended()
  })
  fetchStats()
})
</script>

<style scoped>
.processor-page { min-height: 100vh; background: #f5f5f5; }

.container { max-width: 1200px; margin: 0 auto; padding: 0 20px; }
.main { padding: 30px 20px; }
.page-header { text-align: center; margin-bottom: 30px; }
.page-header h1 { font-size: 28px; color: #333; margin-bottom: 10px; }
.subtitle { color: #666; font-size: 14px; }

.stats-cards { display: grid; grid-template-columns: repeat(4, 1fr); gap: 15px; margin-bottom: 30px; }
.stat-card { background: #fff; border-radius: 12px; padding: 20px; display: flex; align-items: center; gap: 15px; }
.stat-icon { width: 50px; height: 50px; border-radius: 12px; display: flex; align-items: center; justify-content: center; font-size: 24px; }
.stat-card.pending .stat-icon { background: #fdf6ec; color: #e6a23c; }
.stat-card.processing .stat-icon { background: #ecf5ff; color: #409eff; }
.stat-card.completed .stat-icon { background: #f0f9eb; color: #67c23a; }
.stat-card.region .stat-icon { background: #f4f4f5; color: #909399; }
.stat-info { display: flex; flex-direction: column; }
.stat-num { font-size: 28px; font-weight: bold; color: #333; }
.stat-num.region-name { font-size: 14px; font-weight: normal; color: #666; }
.stat-label { font-size: 13px; color: #999; }

.filters { display: flex; gap: 15px; margin-bottom: 20px; flex-wrap: wrap; align-items: center; }

.recommendation-tip {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  background: #f0f9eb;
  border-radius: 8px;
  margin-bottom: 20px;
  color: #67c23a;
  font-size: 14px;
}

.affairs-list { display: flex; flex-direction: column; gap: 15px; }
.affair-card { transition: transform 0.2s; }
.affair-card.matched { border-left: 4px solid #67c23a; }
.affair-item { display: flex; gap: 20px; }
.affair-images { width: 200px; height: 140px; flex-shrink: 0; border-radius: 8px; overflow: hidden; }
.affair-img { width: 100%; height: 100%; }
.affair-info { flex: 1; min-width: 0; }
.affair-header { display: flex; justify-content: flex-start; align-items: center; gap: 10px; margin-bottom: 10px; flex-wrap: wrap; }
.affair-title { font-size: 18px; color: #333; margin: 0; }
.affair-meta { display: flex; gap: 15px; align-items: center; margin-bottom: 10px; font-size: 13px; color: #999; flex-wrap: wrap; }
.type-tag { background: #ecf5ff; color: #409eff; padding: 2px 8px; border-radius: 4px; }
.location { display: flex; align-items: center; gap: 4px; }
.submitter { display: flex; align-items: center; gap: 4px; }
.affair-content { color: #666; font-size: 14px; line-height: 1.6; margin-bottom: 10px; }
.affair-actions { display: flex; gap: 10px; }

/* 我的事务模块样式 */
.my-affairs-section { margin-bottom: 30px; }
.my-affairs-section .section-header { margin-bottom: 20px; }
.my-affairs-section h2 { font-size: 24px; color: #333; margin: 0 0 15px 0; }
.my-affairs-section .filter-tabs { margin-top: 10px; }
.my-affairs-list { display: flex; flex-direction: column; gap: 20px; }
.my-affair-card { padding: 5px; }
.my-affair-content { display: flex; gap: 30px; }
.my-affair-content .affair-left { flex: 1; min-width: 0; }
.my-affair-content .affair-right { display: flex; flex-direction: column; gap: 10px; min-width: 120px; justify-content: center; }
.my-affair-content .affair-header { margin-bottom: 12px; }
.my-affair-content .affair-title { font-size: 20px; font-weight: 600; }
.my-affair-content .affair-meta { margin-bottom: 15px; }
.my-affair-content .affair-content { font-size: 15px; line-height: 1.8; margin-bottom: 15px; white-space: pre-wrap; }
.my-affair-content .process-result { background: #f5f7fa; padding: 15px; border-radius: 8px; margin-top: 15px; }
.my-affair-content .process-result p { color: #666; line-height: 1.6; margin: 0; }
.my-affair-content .process-images { display: flex; gap: 10px; flex-wrap: wrap; margin-top: 10px; }
.my-affair-content .process-img { width: 80px; height: 80px; border-radius: 6px; }

.detail-images { margin-top: 20px; }
.images-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 10px; }
.detail-img { width: 100%; height: 100px; border-radius: 8px; }

.detail-video { margin-top: 20px; }

.detail-process { margin-top: 20px; }
.process-images { display: flex; gap: 10px; flex-wrap: wrap; margin-top: 10px; }
.process-img { width: 100px; height: 100px; border-radius: 8px; }
.process-time { color: #999; font-size: 13px; margin-top: 10px; }

.pagination-wrapper { display: flex; justify-content: center; margin-top: 30px; }
.upload-tip { font-size: 12px; color: #999; display: block; margin-top: 8px; }
.apply-section { text-align: center; padding: 20px 0 30px; }
.processor-status-card { margin-bottom: 20px; }

@media (max-width: 768px) {
  .stats-cards { grid-template-columns: repeat(2, 1fr); }
  .affair-item { flex-direction: column; }
  .affair-images { width: 100%; height: 180px; }
  .images-grid { grid-template-columns: repeat(2, 1fr); }
}
</style>
