<template>
  <div class="expert-list">
    <AppHeader />

    <div class="container main">
      <div class="filter-bar">
        <el-input v-model="keyword" placeholder="搜索专家姓名或专业" clearable @keyup.enter="fetchData" style="width: 300px;">
          <template #append><el-button :icon="Search" @click="fetchData" /></template>
        </el-input>
        <el-select v-model="profession" placeholder="专业领域" clearable @change="fetchData" style="width: 200px;">
          <el-option label="种植" value="种植" />
          <el-option label="养殖" value="养殖" />
          <el-option label="农机" value="农机" />
          <el-option label="病虫害防治" value="病虫害防治" />
          <el-option label="农产品加工" value="农产品加工" />
        </el-select>
        <el-button v-if="token && !myExpert" type="primary" @click="showApplyDialog = true">申请专家认证</el-button>
        <el-button v-if="myExpert && myExpert.status === 2" type="warning" @click="showApplyDialog = true">重新申请</el-button>
      </div>

      <!-- 我的专家申请状态 -->
      <el-card v-if="myExpert" style="margin-bottom: 20px;">
        <template #header>
          <span>我的专家申请状态</span>
        </template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="真实姓名">{{ myExpert.real_name }}</el-descriptions-item>
          <el-descriptions-item label="专业领域">{{ myExpert.profession }}</el-descriptions-item>
          <el-descriptions-item label="职位">{{ myExpert.title }}</el-descriptions-item>
          <el-descriptions-item label="所属单位">{{ myExpert.company }}</el-descriptions-item>
          <el-descriptions-item label="状态" :span="2">
            <el-tag :type="getExpertStatusType(myExpert.status)">{{ getExpertStatusText(myExpert.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item v-if="myExpert.reject_reason" label="拒绝原因" :span="2">
            <span style="color: #f56c6c;">{{ myExpert.reject_reason }}</span>
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <el-row :gutter="20" v-loading="loading">
        <el-col :span="8" v-for="item in list" :key="item.id">
          <el-card class="expert-card" shadow="hover" @click="$router.push('/expert/' + item.id)">
            <div class="expert-header">
              <el-avatar :src="item.user?.avatar_url" :size="80">{{ item.real_name?.[0] }}</el-avatar>
              <el-tag v-if="item.status === 1" type="success" size="small">已认证</el-tag>
              <el-tag v-else type="warning" size="small">待认证</el-tag>
            </div>
            <div class="expert-body">
              <h3>{{ item.real_name }}</h3>
              <p class="title">{{ item.title }} · {{ item.profession }}</p>
              <p class="company">{{ item.company }}</p>
              <p class="intro">{{ item.intro }}</p>
              <div class="contact">
                <span><el-icon><Phone /></el-icon> {{ item.phone }}</span>
              </div>
            </div>
            <div class="expert-footer">
              <el-button type="primary" size="small" @click="handleConsult(item)">咨询</el-button>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <el-empty v-if="!loading && list.length === 0" description="暂无专家" />

      <div class="pagination" v-if="total > 0">
        <el-pagination v-model:current-page="page" :page-size="pageSize" :total="total" layout="prev, pager, next" @current-change="fetchData" />
      </div>
    </div>

    <!-- 咨询对话框 -->
    <el-dialog v-model="showConsultDialog" title="咨询专家" width="500px">
      <p>专家：{{ currentExpert?.real_name }}</p>
      <p>专业：{{ currentExpert?.profession }}</p>
      <el-input v-model="consultContent" type="textarea" :rows="4" placeholder="请输入咨询内容" style="margin-top: 15px;" />
      <template #footer>
        <el-button @click="showConsultDialog = false">取消</el-button>
        <el-button type="primary" @click="submitConsult">提交</el-button>
      </template>
    </el-dialog>

    <!-- 专家申请对话框 -->
    <el-dialog v-model="showApplyDialog" title="申请专家认证" width="600px">
      <el-form :model="applyForm" label-width="100px">
        <el-form-item label="真实姓名" required>
          <el-input v-model="applyForm.real_name" placeholder="请输入真实姓名" />
        </el-form-item>
        <el-form-item label="专业领域" required>
          <el-select v-model="applyForm.profession" placeholder="请选择专业领域" style="width: 100%;">
            <el-option label="种植" value="种植" />
            <el-option label="养殖" value="养殖" />
            <el-option label="农机" value="农机" />
            <el-option label="病虫害防治" value="病虫害防治" />
            <el-option label="农产品加工" value="农产品加工" />
          </el-select>
        </el-form-item>
        <el-form-item label="职位">
          <el-input v-model="applyForm.title" placeholder="请输入职位" />
        </el-form-item>
        <el-form-item label="所属单位">
          <el-input v-model="applyForm.company" placeholder="请输入所属单位" />
        </el-form-item>
        <el-form-item label="联系电话" required>
          <el-input v-model="applyForm.phone" placeholder="请输入联系电话" />
        </el-form-item>
        <el-form-item label="证书编号">
          <el-input v-model="applyForm.cert_no" placeholder="请输入证书编号（选填）" />
        </el-form-item>
        <el-form-item label="个人简介" required>
          <el-input v-model="applyForm.intro" type="textarea" :rows="4" placeholder="请输入个人简介" />
        </el-form-item>
        <el-form-item label="证书/资质图片">
          <el-upload
            action="/api/upload"
            :headers="{ Authorization: token }"
            list-type="picture-card"
            :on-success="handleExpertImageSuccess"
            :before-upload="beforeUpload"
            :on-remove="handleExpertImageRemove"
            :file-list="expertImageList"
            :limit="3"
          >
            <el-icon><Plus /></el-icon>
          </el-upload>
          <span class="upload-tip">上传资质证书或证明材料，最多3张</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showApplyDialog = false">取消</el-button>
        <el-button type="primary" :loading="applying" @click="submitApply">提交申请</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { Search, Phone, Plus } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import api from '@/utils/api'
import { getToken, getUserInfo } from '@/utils/auth'

const router = useRouter()
const token = getToken()
const currentUser = computed(() => getUserInfo())

const list = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(9)
const total = ref(0)
const keyword = ref('')
const profession = ref('')
const showConsultDialog = ref(false)
const currentExpert = ref(null)
const consultContent = ref('')

// 专家申请
const myExpert = ref(null)
const showApplyDialog = ref(false)
const applying = ref(false)
const applyForm = ref({
  real_name: '',
  profession: '',
  title: '',
  company: '',
  phone: '',
  intro: '',
  cert_no: '',
  cert_image: ''
})
const expertImageList = ref([])
const certImagesArr = ref([])

const beforeUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) ElMessage.error('只能上传图片')
  if (!isLt5M) ElMessage.error('图片大小不能超过5MB')
  return isImage && isLt5M
}

const handleExpertImageSuccess = (res) => {
  if (res.url) {
    certImagesArr.value.push(res.url)
    applyForm.value.cert_image = JSON.stringify(certImagesArr.value)
  }
}

const handleExpertImageRemove = (file) => {
  const url = file.url || file.response?.url
  certImagesArr.value = certImagesArr.value.filter(u => u !== url)
  applyForm.value.cert_image = certImagesArr.value.length > 0 ? JSON.stringify(certImagesArr.value) : ''
}

const getExpertStatusText = (status) => {
  const map = { 0: '待审核', 1: '已通过', 2: '已拒绝' }
  return map[status] || '未知'
}
const getExpertStatusType = (status) => {
  const map = { 0: 'warning', 1: 'success', 2: 'danger' }
  return map[status] || ''
}

const fetchData = async () => {
  loading.value = true
  try {
    let res
    if (profession.value) {
      res = await api.getExpertsByProfession(profession.value)
    } else if (keyword.value) {
      res = await api.searchExperts(keyword.value)
    } else {
      res = await api.getExperts({ page: page.value, page_size: pageSize.value })
    }
    list.value = (res.list || res || []).map(item => ({
      ...item,
      user: item.user ? {
        ...item.user,
        avatar_url: item.user.avatar_url?.startsWith('/') ? item.user.avatar_url : '/' + item.user.avatar_url
      } : null
    }))
    total.value = res.total || list.value.length
  } finally {
    loading.value = false
  }
}

const loadMyExpert = async () => {
  try {
    const res = await api.getMyExpert()
    if (res) {
      myExpert.value = res
    }
  } catch (error) {
    // 未申请专家
    myExpert.value = null
  }
}

const handleConsult = (expert) => {
  if (!token) {
    ElMessage.warning('请先登录')
    return
  }
  currentExpert.value = expert
  showConsultDialog.value = true
}

const submitConsult = async () => {
  if (!consultContent.value.trim()) {
    ElMessage.warning('请输入咨询内容')
    return
  }
  try {
    const res = await api.createConsultation({
      expert_id: currentExpert.value.id,
      content: consultContent.value.trim()
    })
    ElMessage.success('咨询已提交')
    showConsultDialog.value = false
    consultContent.value = ''
    // 跳转到咨询会话
    router.push(`/messages/${res.conversation_id}`)
  } catch (error) {
    ElMessage.error(error?.response?.data?.error || '咨询提交失败')
  }
}

const submitApply = async () => {
  if (!applyForm.value.real_name || !applyForm.value.profession || !applyForm.value.phone || !applyForm.value.intro) {
    ElMessage.warning('请填写必填项')
    return
  }
  applying.value = true
  try {
    await api.applyExpert(applyForm.value)
    ElMessage.success('申请已提交')
    showApplyDialog.value = false
    loadMyExpert()
  } catch (error) {
    ElMessage.error('申请失败')
  } finally {
    applying.value = false
  }
}

onMounted(() => {
  fetchData()
  if (token) {
    loadMyExpert()
  }
})
</script>

<style scoped>
.expert-list { min-height: 100vh; background: #f5f5f5; }
.container { max-width: 1200px; margin: 0 auto; padding: 0 20px; }
.main { padding: 20px; }
.filter-bar { display: flex; gap: 15px; margin-bottom: 20px; align-items: center; }

.expert-card { text-align: center; cursor: pointer; }
.expert-header { display: flex; flex-direction: column; align-items: center; gap: 10px; margin-bottom: 15px; }
.expert-body h3 { margin-bottom: 8px; }
.expert-body .title { color: #409eff; margin-bottom: 5px; }
.expert-body .company { color: #999; font-size: 13px; margin-bottom: 10px; }
.expert-body .intro { color: #666; font-size: 14px; overflow: hidden; text-overflow: ellipsis; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; height: 42px; margin-bottom: 10px; }
.expert-body .contact { color: #666; font-size: 13px; display: flex; justify-content: center; }
.expert-body .contact span { display: flex; align-items: center; gap: 5px; }
.expert-footer { margin-top: 15px; padding-top: 15px; border-top: 1px solid #eee; }

.pagination { display: flex; justify-content: center; margin-top: 20px; }
.upload-tip { font-size: 12px; color: #999; display: block; margin-top: 8px; }
</style>
