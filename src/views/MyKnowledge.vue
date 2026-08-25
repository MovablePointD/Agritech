<template>
  <div class="my-knowledge">
    <AppHeader />

    <div class="container main">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>我的知识</span>
            <el-button type="primary" size="small" @click="showPublishDialog = true">发布知识</el-button>
          </div>
        </template>

        <el-table v-if="!loading && list.length > 0" :data="list" v-loading="loading">
          <el-table-column prop="title" label="标题" />
          <el-table-column prop="created_at" label="发布时间" width="180" :formatter="formatTime" />
          <el-table-column label="操作" width="220">
            <template #default="{ row }">
              <el-button size="small" @click="$router.push(`/knowledge/${row.id}`)">查看</el-button>
              <el-button size="small" type="primary" @click="editKnowledge(row)">编辑</el-button>
              <el-button size="small" type="danger" @click="handleDelete(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>

        <div v-if="!loading && list.length === 0" class="empty-state">
          <div class="empty-icon">
            <el-icon :size="60"><Reading /></el-icon>
          </div>
          <p class="empty-text">暂无发布的知识</p>
          <el-button type="primary" @click="showPublishDialog = true">发布第一条知识</el-button>
        </div>
      </el-card>
    </div>

    <!-- 发布/编辑知识对话框（简洁模式） -->
    <el-dialog v-model="showPublishDialog" :title="editForm.id ? '编辑知识' : '发布知识'" width="600px">
      <el-form :model="editForm" label-width="80px">
        <el-form-item label="标题">
          <el-input v-model="editForm.title" placeholder="请输入标题" maxlength="100" show-word-limit />
        </el-form-item>
        <el-form-item label="内容">
          <el-input v-model="editForm.content" type="textarea" :rows="6" placeholder="详细描述您的农业知识..." />
        </el-form-item>
        <el-form-item label="封面图">
          <el-upload action="/api/upload" :headers="{ Authorization: tokenValue }" list-type="picture-card" :on-success="handleUploadSuccess" :before-upload="beforeUpload" :on-remove="handleRemove" :file-list="fileList" :limit="1">
            <el-icon><Plus /></el-icon>
          </el-upload>
          <div class="upload-tip">上传知识封面图片（可选）</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <el-button type="primary" link @click="$router.push(`/rich-editor/knowledge${editForm.id ? '?id=' + editForm.id : ''}`)">
            <el-icon><EditPen /></el-icon>
            使用专业富文本编辑器
          </el-button>
          <div>
            <el-button @click="showPublishDialog = false">取消</el-button>
            <el-button type="primary" @click="handleSave" :loading="saving">保存</el-button>
          </div>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Reading, Plus, EditPen } from '@element-plus/icons-vue'
import api from '@/utils/api'
import { getToken, getUserInfo } from '@/utils/auth'

const router = useRouter()

const token = computed(() => getToken())
const tokenValue = computed(() => token.value ? `Bearer ${token.value}` : '')

const list = ref([])
const loading = ref(false)
const showPublishDialog = ref(false)
const editForm = ref({
  id: null,
  title: '',
  type: 'planting',
  content: '',
  image_url: ''
})
const fileList = ref([])
const saving = ref(false)

const initUserInfo = () => {
  const userData = getUserInfo()
  if (userData) {
    const avatar_url = userData.avatar_url?.startsWith('/') ? userData.avatar_url : '/' + (userData.avatar_url || '')
    const nickname = userData.nickname || userData.username || ''
    currentUser.value = { ...userData, avatar_url, nickname }
  }
}

const formatTime = (row) => row.created_at ? new Date(row.created_at).toLocaleString('zh-CN') : ''

const fetchData = async () => {
  loading.value = true
  try {
    const res = await api.getMyKnowledge()
    list.value = res?.list || res || []
  } finally {
    loading.value = false
  }
}

const editKnowledge = (row) => {
  editForm.value = {
    id: row.id,
    title: row.title,
    type: row.type,
    content: row.content,
    image_url: row.image_url || ''
  }
  if (row.image_url) {
    fileList.value = [{ url: row.image_url }]
  } else {
    fileList.value = []
  }
  showPublishDialog.value = true
}

const handleUploadSuccess = (res) => {
  editForm.value.image_url = res.url || res.data || ''
}

const beforeUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) ElMessage.error('只能上传图片文件')
  if (!isLt5M) ElMessage.error('图片大小不能超过5MB')
  return isImage && isLt5M
}

const handleRemove = () => {
  editForm.value.image_url = ''
}

const handleSave = async () => {
  if (!editForm.value.title) {
    ElMessage.warning('请填写标题')
    return
  }
  if (!editForm.value.content?.trim()) {
    ElMessage.warning('请填写内容')
    return
  }
  saving.value = true
  try {
    if (editForm.value.id) {
      await api.updateKnowledge(editForm.value.id, editForm.value)
    } else {
      await api.createKnowledge(editForm.value)
    }
    ElMessage.success('保存成功')
    showPublishDialog.value = false
    editForm.value = { id: null, title: '', type: 'planting', content: '', image_url: '' }
    fileList.value = []
    fetchData()
  } catch (error) {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

const handleDelete = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除该知识吗？', '提示', { type: 'warning' })
    await api.deleteKnowledge(id)
    ElMessage.success('删除成功')
    fetchData()
  } catch (error) {
    if (error !== 'cancel') ElMessage.error('删除失败')
  }
}

onMounted(() => {
  initUserInfo()
  fetchData()
})
</script>

<style scoped>
.my-knowledge { min-height: 100vh; background: #f5f5f5; }
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
.empty-icon { margin-bottom: 20px; }
.empty-text { color: #909399; font-size: 14px; margin-bottom: 20px; }

.upload-tip {
  color: #909399;
  font-size: 12px;
  margin-top: 5px;
}
</style>
