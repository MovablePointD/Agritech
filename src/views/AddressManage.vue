<template>
  <div class="address-manage">
    <AppHeader />

    <div class="container main">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>收货地址</span>
            <el-button type="primary" size="small" @click="openAddDialog">新增地址</el-button>
          </div>
        </template>

        <el-table v-if="!loading && list.length > 0" :data="list" v-loading="loading">
          <el-table-column prop="receiver" label="收货人" width="100" />
          <el-table-column prop="phone" label="联系电话" width="130" />
          <el-table-column label="地址">
            <template #default="{ row }">
              {{ row.province }} {{ row.city }} {{ row.district }} {{ row.detail }}
            </template>
          </el-table-column>
          <el-table-column label="标签" width="100">
            <template #default="{ row }">
              <el-tag v-if="row.label" size="small">{{ row.label }}</el-tag>
              <span v-else style="color: #999;">无</span>
            </template>
          </el-table-column>
          <el-table-column label="默认" width="80">
            <template #default="{ row }">
              <el-tag v-if="row.is_default === 1" type="success" size="small">默认</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="220">
            <template #default="{ row }">
              <el-button size="small" @click="handleEdit(row)">编辑</el-button>
              <el-button size="small" type="warning" @click="setDefault(row.id)" v-if="row.is_default !== 1">设为默认</el-button>
              <el-button size="small" type="danger" @click="handleDelete(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>

        <!-- 空状态 -->
        <div v-if="!loading && list.length === 0" class="empty-state">
          <div class="empty-icon">
            <svg class="empty-svg" viewBox="0 0 120 120" fill="none" xmlns="http://www.w3.org/2000/svg">
              <circle cx="60" cy="60" r="50" fill="#f5f5f5" stroke="#e0e0e0" stroke-width="2"/>
              <path d="M35 45H85V85H35V45Z" fill="#fff" stroke="#d0d0d0" stroke-width="2"/>
              <path d="M35 45L50 30H70L85 45" stroke="#d0d0d0" stroke-width="2" stroke-linecap="round"/>
              <circle cx="60" cy="65" r="12" fill="#fff" stroke="#d0d0d0" stroke-width="2"/>
              <path d="M56 65L59 68L64 62" stroke="#d0d0d0" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </div>
          <p class="empty-text">暂无收货地址</p>
          <el-button type="primary" @click="openAddDialog">添加地址</el-button>
        </div>
      </el-card>
    </div>

    <!-- 地址编辑对话框 -->
    <el-dialog v-model="showDialog" :title="form.id ? '编辑地址' : '新增地址'" width="600px">
      <el-form :model="form" label-width="80px" :rules="formRules" ref="formRef">
        <el-form-item label="收货人" prop="receiver">
          <el-input v-model="form.receiver" placeholder="请输入收货人姓名" />
        </el-form-item>
        <el-form-item label="联系电话" prop="phone">
          <el-input v-model="form.phone" placeholder="请输入联系电话" />
        </el-form-item>
        <el-form-item label="省份" prop="province">
          <el-input v-model="form.province" placeholder="请输入省份" />
        </el-form-item>
        <el-form-item label="城市" prop="city">
          <el-input v-model="form.city" placeholder="请输入城市" />
        </el-form-item>
        <el-form-item label="区县" prop="district">
          <el-input v-model="form.district" placeholder="请输入区县" />
        </el-form-item>
        <el-form-item label="详细地址" prop="detail">
          <el-input v-model="form.detail" placeholder="请输入详细地址" />
        </el-form-item>
        <el-form-item label="标签">
          <el-input v-model="form.label" placeholder="如：家、公司" />
        </el-form-item>
        <el-form-item label="设为默认">
          <el-switch v-model="isDefaultSwitch" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/utils/api'
import { getToken, getUserInfo } from '@/utils/auth'

const router = useRouter()

const token = computed(() => getToken())

const currentUser = ref({
  username: '',
  nickname: '',
  avatar_url: ''
})

const list = ref([])
const loading = ref(false)
const showDialog = ref(false)
const formRef = ref(null)
const isDefaultSwitch = ref(false)

const form = ref({
  id: null,
  receiver: '',
  phone: '',
  province: '',
  city: '',
  district: '',
  detail: '',
  label: '',
  is_default: 0
})

// 表单验证规则
const formRules = {
  receiver: [{ required: true, message: '请输入收货人姓名', trigger: 'blur' }],
  phone: [{ required: true, message: '请输入联系电话', trigger: 'blur' }],
  province: [{ required: true, message: '请输入省份', trigger: 'blur' }],
  city: [{ required: true, message: '请输入城市', trigger: 'blur' }],
  district: [{ required: true, message: '请输入区县', trigger: 'blur' }],
  detail: [{ required: true, message: '请输入详细地址', trigger: 'blur' }]
}

// 初始化用户信息
const initUserInfo = () => {
  const userData = getUserInfo()
  if (userData) {
    const avatar_url = userData.avatar_url?.startsWith('/') ? userData.avatar_url : '/' + (userData.avatar_url || '')
    const nickname = userData.nickname || userData.username || ''
    currentUser.value = { ...userData, avatar_url, nickname }
  }
}

// 加载用户详细信息
const loadUserDetail = async () => {
  try {
    const res = await api.getUserInfo()
    if (res && !res.error) {
      const avatar_url = res.avatar_url?.startsWith('/') ? res.avatar_url : '/' + (res.avatar_url || '')
      const nickname = res.nickname || res.username || ''
      currentUser.value = { ...res, avatar_url, nickname }
    }
  } catch (error) {
    console.error('加载用户信息失败:', error)
  }
}

const fetchData = async () => {
  loading.value = true
  try {
    const res = await api.getAddresses()
    list.value = res?.list || res || []
  } finally {
    loading.value = false
  }
}

// 打开新增地址对话框
const openAddDialog = () => {
  form.value = {
    id: null,
    receiver: '',
    phone: '',
    province: '',
    city: '',
    district: '',
    detail: '',
    label: '',
    is_default: 0
  }
  isDefaultSwitch.value = false
  showDialog.value = true
}

const handleEdit = (row) => {
  form.value = {
    id: row.id,
    receiver: row.receiver,
    phone: row.phone,
    province: row.province,
    city: row.city,
    district: row.district,
    detail: row.detail,
    label: row.label || '',
    is_default: row.is_default
  }
  isDefaultSwitch.value = row.is_default === 1
  showDialog.value = true
}

const handleSave = async () => {
  // 验证表单
  if (!form.value.receiver || !form.value.phone || !form.value.province || !form.value.city || !form.value.district || !form.value.detail) {
    ElMessage.warning('请填写完整信息')
    return
  }

  // 设置默认地址标识
  form.value.is_default = isDefaultSwitch.value ? 1 : 0

  try {
    if (form.value.id) {
      // 更新地址时，如果设置为默认，先取消其他默认
      if (isDefaultSwitch.value) {
        await api.setDefaultAddress(form.value.id)
      }
      await api.updateAddress(form.value)
    } else {
      await api.createAddress(form.value)
    }
    ElMessage.success('保存成功')
    showDialog.value = false
    fetchData()
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

const setDefault = async (id) => {
  try {
    await api.setDefaultAddress(id)
    ElMessage.success('设置成功')
    fetchData()
  } catch (error) {
    ElMessage.error('设置失败')
  }
}

const handleDelete = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除该地址吗？', '提示', { type: 'warning' })
    await api.deleteAddress(id)
    ElMessage.success('删除成功')
    fetchData()
  } catch (error) {
    if (error !== 'cancel') ElMessage.error('删除失败')
  }
}

onMounted(() => {
  initUserInfo()
  loadUserDetail()
  fetchData()
})
</script>

<style scoped>
.address-manage { min-height: 100vh; background: #f5f5f5; }
.container { max-width: 1200px; margin: 0 auto; padding: 0 20px; }
.main { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }

/* 空状态样式 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
}

.empty-icon {
  margin-bottom: 20px;
}

.empty-svg {
  width: 120px;
  height: 120px;
}

.empty-text {
  color: #909399;
  font-size: 14px;
  margin-bottom: 20px;
}
</style>
