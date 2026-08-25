<template>
  <div class="expert-detail-page">
    <AppHeader />
    <div class="container main" v-loading="loading">
      <el-card v-if="expert" class="detail-card">
        <div class="breadcrumb">
          <el-button text @click="$router.push('/experts')">
            <el-icon><ArrowLeft /></el-icon> 返回专家列表
          </el-button>
        </div>

        <div class="expert-header">
          <el-avatar :src="expert.user?.avatar_url" :size="80">{{ expert.real_name?.[0] }}</el-avatar>
          <div class="expert-info">
            <h1>{{ expert.real_name }}</h1>
            <p class="expert-title">{{ expert.title }} · {{ expert.profession }}</p>
            <el-tag :type="getStatusType(expert.status)">{{ getStatusText(expert.status) }}</el-tag>
          </div>
        </div>

        <el-divider />

        <el-descriptions :column="2" border>
          <el-descriptions-item label="真实姓名">{{ expert.real_name }}</el-descriptions-item>
          <el-descriptions-item label="专业领域">{{ expert.profession }}</el-descriptions-item>
          <el-descriptions-item label="职位">{{ expert.title || '未填写' }}</el-descriptions-item>
          <el-descriptions-item label="所属单位">{{ expert.company || '未填写' }}</el-descriptions-item>
          <el-descriptions-item label="联系电话">{{ expert.phone || '未填写' }}</el-descriptions-item>
          <el-descriptions-item label="证书编号">{{ expert.cert_no || '未填写' }}</el-descriptions-item>
          <el-descriptions-item v-if="expert.user" label="用户名">{{ expert.user.username }}</el-descriptions-item>
          <el-descriptions-item label="认证状态">
            <el-tag :type="getStatusType(expert.status)">{{ getStatusText(expert.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item v-if="expert.verify_at" label="审核时间">{{ formatTime(expert.verify_at) }}</el-descriptions-item>
          <el-descriptions-item label="申请时间">{{ formatTime(expert.created_at) }}</el-descriptions-item>
        </el-descriptions>

        <!-- 个人简介 -->
        <div class="intro-section" v-if="expert.intro">
          <h3>个人简介</h3>
          <p>{{ expert.intro }}</p>
        </div>

        <!-- 证书/资质图片 -->
        <div class="cert-section" v-if="certImages.length > 0">
          <h3>资质证书</h3>
          <div class="cert-images">
            <el-image
              v-for="(img, idx) in certImages"
              :key="idx"
              :src="img"
              fit="cover"
              :preview-src-list="certImages"
              :initial-index="idx"
              class="cert-image"
            />
          </div>
        </div>

        <!-- 咨询按钮 -->
        <div class="action-bar" v-if="token && expert.status === 1 && expert.user_id !== currentUserId">
          <el-button type="primary" size="large" @click="handleConsult">咨询专家</el-button>
        </div>
      </el-card>
      <div v-else-if="!loading" class="not-found">
        <p>专家不存在</p>
        <el-button type="primary" @click="$router.push('/experts')">返回专家列表</el-button>
      </div>
    </div>
    <AppFooter />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import AppHeader from '@/components/AppHeader.vue'
import AppFooter from '@/components/AppFooter.vue'
import api from '@/utils/api'
import { getToken, getUserInfo } from '@/utils/auth'

const route = useRoute()
const router = useRouter()

const token = computed(() => getToken())
const currentUserId = computed(() => getUserInfo()?.id)

const loading = ref(true)
const expert = ref(null)

const certImages = computed(() => {
  if (!expert.value?.cert_image) return []
  try { return JSON.parse(expert.value.cert_image) } catch { return [] }
})

const getStatusText = (status) => {
  const map = { 0: '待审核', 1: '已通过', 2: '已拒绝' }
  return map[status] || '未知'
}
const getStatusType = (status) => {
  const map = { 0: 'warning', 1: 'success', 2: 'danger' }
  return map[status] || ''
}
const formatTime = (t) => t ? new Date(t).toLocaleString('zh-CN') : ''

const handleConsult = async () => {
  if (!token.value) { ElMessage.warning('请先登录'); return }
  try {
    const res = await api.createConsultation({
      expert_id: expert.value.id,
      content: ''
    })
    ElMessage.success('已创建咨询会话')
    router.push(`/messages/${res.conversation_id}`)
  } catch (error) {
    ElMessage.error('创建咨询失败')
  }
}

onMounted(async () => {
  loading.value = true
  try {
    const res = await api.getExpert(route.params.id)
    expert.value = res
    if (expert.value.user?.avatar_url && !expert.value.user.avatar_url.startsWith('/')) {
      expert.value.user.avatar_url = '/' + expert.value.user.avatar_url
    }
  } catch {
    expert.value = null
  }
  loading.value = false
})
</script>

<style scoped>
.expert-detail-page { min-height: 100vh; background: #f5f7fa; }
.container { max-width: 860px; margin: 0 auto; padding: 0 20px; }
.main { padding: 30px 20px; }
.detail-card { border-radius: 12px; }
.breadcrumb { margin-bottom: 15px; }

.expert-header { display: flex; align-items: center; gap: 20px; margin-bottom: 10px; }
.expert-info h1 { font-size: 24px; color: #333; margin: 0 0 6px; }
.expert-title { color: #409eff; font-size: 14px; margin: 0 0 10px; }

.intro-section { margin-top: 20px; }
.intro-section h3 { font-size: 16px; color: #333; margin-bottom: 10px; }
.intro-section p { color: #666; line-height: 1.8; font-size: 14px; }

.cert-section { margin-top: 20px; }
.cert-section h3 { font-size: 16px; color: #333; margin-bottom: 10px; }
.cert-images { display: flex; flex-wrap: wrap; gap: 10px; }
.cert-image { width: 200px; height: 140px; border-radius: 8px; object-fit: cover; }

.action-bar { margin-top: 24px; text-align: center; }

.not-found { text-align: center; padding: 60px; color: #999; }
</style>
