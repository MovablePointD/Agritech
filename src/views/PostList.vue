<template>
  <div class="post-list">
    <AppHeader />

    <div class="container main">
      <div class="toolbar">
        <el-radio-group v-model="type" @change="fetchData">
          <el-radio-button label="">全部</el-radio-button>
          <el-radio-button label="normal">普通</el-radio-button>
          <el-radio-button label="question">提问</el-radio-button>
          <el-radio-button label="share">分享</el-radio-button>
        </el-radio-group>
        <div class="toolbar-right">
          <el-input v-model="keyword" placeholder="搜索动态" clearable @keyup.enter="fetchData" style="width: 220px;" class="search-input">
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <el-button type="primary" @click="fetchData" :icon="Search">搜索</el-button>
          <el-button type="primary" @click="showPublishDialog = true" v-if="token">发布动态</el-button>
        </div>
      </div>

      <div class="post-grid" v-loading="loading">
        <div v-for="item in list" :key="item.id" class="post-card" :class="{ 'pending-card': item.status !== 1 }">
          <router-link :to="`/post/${item.id}`">
            <div class="post-header">
              <el-avatar :src="item.user?.avatar_url" :size="40">{{ item.user?.username?.[0] }}</el-avatar>
              <span class="username">{{ item.user?.username }}</span>
              <el-tag size="small" :type="getTypeTag(item.type)">{{ getTypeName(item.type) }}</el-tag>
              <el-tag v-if="item.status === -1" size="small" type="warning" class="status-tag">审核中</el-tag>
              <el-tag v-else-if="item.status === 0" size="small" type="danger" class="status-tag">已退回</el-tag>
            </div>
            <h3>{{ item.title }}</h3>
            <p class="content">{{ stripHtml(item.content) }}</p>
            <div class="images" v-if="item.images">
              <el-image v-for="(img, i) in parseImages(item.images)" :key="i" :src="img" :preview-src-list="parseImages(item.images)" fit="cover" />
            </div>
            <div class="post-footer">
              <span><el-icon><View /></el-icon> {{ item.views }}</span>
              <span><el-icon><Star /></el-icon> {{ item.likes }}</span>
            </div>
          </router-link>
        </div>
      </div>

      <el-empty v-if="!loading && list.length === 0" description="暂无动态" />

      <div class="pagination" v-if="total > 0">
        <el-pagination v-model:current-page="page" :page-size="pageSize" :total="total" layout="prev, pager, next" @current-change="fetchData" />
      </div>
    </div>

    <AppFooter />

    <!-- 发布动态对话框（简洁模式） -->
    <el-dialog v-model="showPublishDialog" title="发布动态" width="600px">
      <el-form :model="postForm" label-width="80px">
        <el-form-item label="标题">
          <el-input v-model="postForm.title" placeholder="请输入标题" maxlength="100" show-word-limit />
        </el-form-item>
        <el-form-item label="内容">
          <el-input v-model="postForm.content" type="textarea" :rows="6" placeholder="分享您的想法、经验或问题..." />
        </el-form-item>
        <el-form-item label="图片">
          <el-upload action="/api/upload" :headers="{ Authorization: tokenValue }" list-type="picture-card" :on-success="handleUploadSuccess" :before-upload="beforeUpload">
            <el-icon><Plus /></el-icon>
          </el-upload>
          <div class="upload-tip">支持拖拽或点击上传多张图片（可选）</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <el-button type="primary" link @click="$router.push('/rich-editor/post')">
            <el-icon><EditPen /></el-icon>
            使用专业富文本编辑器
          </el-button>
          <div>
            <el-button @click="showPublishDialog = false">取消</el-button>
            <el-button type="primary" @click="handlePublish" :loading="publishing">发布</el-button>
          </div>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { View, Star, Plus, EditPen, Search } from '@element-plus/icons-vue'
import api from '@/utils/api'
import { getToken } from '@/utils/auth'

const router = useRouter()
const token = computed(() => getToken())
const tokenValue = computed(() => getToken() || '')

const list = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(12)
const total = ref(0)
const type = ref('')
const keyword = ref('')
const showPublishDialog = ref(false)
const publishing = ref(false)

const postForm = ref({ title: '', content: '', type: 'normal', images: '' })

const getTypeName = (t) => ({ normal: '普通', question: '提问', share: '分享' }[t] || '普通')
const getTypeTag = (t) => ({ normal: '', question: 'warning', share: 'success' }[t] || '')

const fetchData = async () => {
  loading.value = true
  try {
    const res = await api.getPosts({ page: page.value, page_size: pageSize.value, type: type.value, keyword: keyword.value })
    list.value = (res.list || []).map(item => ({
      ...item,
      user: item.user ? {
        ...item.user,
        avatar_url: item.user.avatar_url?.startsWith('/') ? item.user.avatar_url : '/' + item.user.avatar_url
      } : null
    }))
    total.value = res.total || 0
  } finally {
    loading.value = false
  }
}

const beforeUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) ElMessage.error('只能上传图片')
  if (!isLt5M) ElMessage.error('图片大小不能超过5MB')
  return isImage && isLt5M
}

const handleUploadSuccess = (res) => {
  try {
    const currentImages = JSON.parse(postForm.value.images || '[]')
    postForm.value.images = JSON.stringify([...currentImages, res.url])
  } catch (e) {
    postForm.value.images = JSON.stringify([res.url])
  }
}

const parseImages = (imagesStr) => {
  if (!imagesStr) return []
  try {
    const images = JSON.parse(imagesStr)
    return images.map(img => img.startsWith('/') ? img : '/' + img)
  } catch (e) {
    return []
  }
}

// 去除HTML标签获取纯文本
const stripHtml = (html) => {
  if (!html) return ''
  const doc = new DOMParser().parseFromString(html, 'text/html')
  return doc.body.textContent || ''
}

const handlePublish = async () => {
  if (!postForm.value.title) {
    ElMessage.warning('请填写标题')
    return
  }
  if (!postForm.value.content?.trim()) {
    ElMessage.warning('请填写内容')
    return
  }
  publishing.value = true
  try {
    await api.createPost(postForm.value)
    ElMessage.success('发布成功')
    showPublishDialog.value = false
    postForm.value = { title: '', content: '', type: 'normal', images: '' }
    fetchData()
  } catch (error) {
    ElMessage.error('发布失败')
  } finally {
    publishing.value = false
  }
}

onMounted(fetchData)
</script>

<style scoped>
.post-list { min-height: 100vh; background: #f5f5f5; }
.container { max-width: 1200px; margin: 0 auto; padding: 0 20px; }
.main { padding: 20px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 20px; align-items: center; flex-wrap: wrap; gap: 10px; }
.toolbar-right { display: flex; align-items: center; gap: 10px; }
.search-input { flex-shrink: 0; }

.post-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; }
.post-card {
  background: #fff; border-radius: 8px; padding: 20px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  display: flex; flex-direction: column;
  min-height: 260px;
}
.post-card.pending-card { opacity: 0.82; border-left: 3px solid #e6a23c; }
.post-card a { text-decoration: none; color: inherit; display: flex; flex-direction: column; flex: 1; }
.post-header { display: flex; align-items: center; gap: 8px; margin-bottom: 15px; flex-wrap: wrap; }
.status-tag { margin-left: auto; }
.username { font-weight: 500; color: #333; }
.post-card h3 { margin-bottom: 10px; color: #333; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.post-card .content { color: #666; font-size: 14px; overflow: hidden; text-overflow: ellipsis; display: -webkit-box; -webkit-line-clamp: 3; -webkit-box-orient: vertical; margin-bottom: 10px; flex: 1; }
.images { display: flex; gap: 5px; margin-bottom: 10px; overflow: hidden; }
.images .el-image { width: 80px; height: 80px; border-radius: 4px; flex-shrink: 0; }
.post-footer { display: flex; gap: 20px; color: #999; font-size: 13px; margin-top: auto; }
.post-footer span { display: flex; align-items: center; gap: 4px; }

.pagination { display: flex; justify-content: center; margin-top: 20px; }

.upload-tip {
  color: #909399;
  font-size: 12px;
  margin-top: 5px;
}
</style>
