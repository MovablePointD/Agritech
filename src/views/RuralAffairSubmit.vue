<template>
  <div class="affair-submit-page">
    <AppHeader />

    <div class="container main">
      <el-card class="submit-card">
        <template #header>
          <div class="card-header">
            <el-button text @click="$router.back()"><el-icon><ArrowLeft /></el-icon> 返回</el-button>
            <span class="title">上报农村事务</span>
          </div>
        </template>

        <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
          <el-form-item label="事务标题" prop="title">
            <el-input v-model="form.title" placeholder="请简要描述您要上报的问题" maxlength="100" show-word-limit />
          </el-form-item>

          <el-form-item label="事务类型" prop="type">
            <el-select v-model="form.type" placeholder="请选择事务类型" style="width: 100%;">
              <el-option label="基础设施" value="infrastructure" />
              <el-option label="硬件问题" value="hardware" />
              <el-option label="环境问题" value="env" />
              <el-option label="安全隐患" value="safety" />
              <el-option label="其他" value="other" />
            </el-select>
          </el-form-item>

          <el-form-item label="事发地点" prop="address">
            <el-input v-model="form.address" placeholder="请输入问题发生的具体地点" maxlength="200" />
          </el-form-item>

          <el-form-item label="问题描述" prop="content">
            <el-input v-model="form.content" type="textarea" :rows="8" placeholder="请详细描述您发现的问题，包括具体情况、影响范围等" maxlength="2000" show-word-limit />
          </el-form-item>

          <el-form-item label="图片上传">
            <div class="upload-area">
              <el-upload
                action="#"
                :auto-upload="false"
                :limit="9"
                list-type="picture-card"
                :on-change="handleImageChange"
                :on-remove="handleImageRemove"
                :file-list="imageList"
              >
                <el-icon><Plus /></el-icon>
              </el-upload>
              <div class="upload-tip">最多上传9张图片，支持jpg、png格式</div>
            </div>
          </el-form-item>

          <el-form-item label="视频上传">
            <div class="video-upload">
              <el-upload
                ref="videoUploadRef"
                action="#"
                :auto-upload="false"
                :limit="1"
                accept="video/*"
                :on-change="handleVideoChange"
                :on-remove="handleVideoRemove"
                :file-list="videoList"
              >
                <el-button type="primary" plain><el-icon><Upload /></el-icon> 选择视频</el-button>
              </el-upload>
              <div class="upload-tip">支持mp4、avi、mov格式，最大100MB</div>
            </div>
          </el-form-item>

          <el-form-item>
            <el-button type="primary" :loading="submitting" @click="handleSubmit" size="large">
              {{ submitting ? '提交中...' : '提交事务' }}
            </el-button>
            <el-button @click="$router.back()" size="large">取消</el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <el-card class="tips-card">
        <template #header>温馨提示</template>
        <div class="tips">
          <div class="tip-item"><el-icon color="#409eff"><InfoFilled /></el-icon> 请确保提供的信息真实有效</div>
          <div class="tip-item"><el-icon color="#67c23a"><Clock /></el-icon> 我们会在1-3个工作日内完成审核</div>
          <div class="tip-item"><el-icon color="#e6a23c"><Bell /></el-icon> 审核结果会通过系统消息通知您</div>
          <div class="tip-item"><el-icon color="#f56c6c"><Warning /></el-icon> 对于紧急安全隐患，请同时联系当地相关部门</div>
        </div>
      </el-card>
    </div>

    <AppFooter />
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Plus, Upload, InfoFilled, Clock, Bell, Warning } from '@element-plus/icons-vue'
import api from '@/utils/api'
import { getToken } from '@/utils/auth'

const router = useRouter()
const token = getToken()

const formRef = ref(null)
const submitting = ref(false)
const imageList = ref([])
const videoList = ref([])
const videoUploadRef = ref(null)

const form = reactive({
  title: '',
  type: '',
  address: '',
  content: '',
  images: '',
  video_url: ''
})

const rules = {
  title: [{ required: true, message: '请输入事务标题', trigger: 'blur' }],
  type: [{ required: true, message: '请选择事务类型', trigger: 'change' }],
  address: [{ required: true, message: '请输入事发地点', trigger: 'blur' }],
  content: [{ required: true, message: '请输入问题描述', trigger: 'blur' }]
}

const handleImageChange = (file, files) => {
  imageList.value = files
}

const handleImageRemove = (file, files) => {
  imageList.value = files
}

const handleVideoChange = (file, files) => {
  videoList.value = files
}

const handleVideoRemove = (file, files) => {
  videoList.value = files
}

const uploadImages = async () => {
  const uploadedUrls = []
  for (const file of imageList.value) {
    if (file.raw) {
      const formData = new FormData()
      formData.append('file', file.raw)
      try {
        const res = await api.upload(formData)
        uploadedUrls.push(res.url)
      } catch (error) {
        console.error('图片上传失败:', error)
      }
    } else if (file.url) {
      uploadedUrls.push(file.url)
    }
  }
  return uploadedUrls
}

const uploadVideo = async () => {
  if (videoList.value.length > 0 && videoList.value[0].raw) {
    const formData = new FormData()
    formData.append('file', videoList.value[0].raw)
    try {
      const res = await api.uploadVideo(formData)
      return res.url
    } catch (error) {
      console.error('视频上传失败:', error)
      ElMessage.warning('视频上传失败，将继续提交其他内容')
    }
  }
  return ''
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    
    if (!token) {
      ElMessage.warning('请先登录')
      router.push('/login')
      return
    }

    submitting.value = true
    try {
      // 上传图片
      const imageUrls = await uploadImages()
      form.images = JSON.stringify(imageUrls)
      
      // 上传视频
      form.video_url = await uploadVideo()

      await api.createRuralAffair(form)
      ElMessage.success('提交成功！我们会尽快审核您的事务')
      router.push('/affairs')
    } catch (error) {
      ElMessage.error('提交失败，请重试')
    } finally {
      submitting.value = false
    }
  })
}
</script>

<style scoped>
.affair-submit-page { min-height: 100vh; background: #f5f5f5; }

.container { max-width: 900px; margin: 0 auto; padding: 0 20px; }
.main { padding: 30px 20px; }
.submit-card { margin-bottom: 20px; }
.card-header { display: flex; align-items: center; gap: 15px; }
.title { font-size: 18px; font-weight: 500; }

.upload-area { display: flex; flex-direction: column; gap: 10px; }
.upload-tip { font-size: 12px; color: #999; }
.video-upload { display: flex; flex-direction: column; gap: 10px; }

.tips-card .tips { display: flex; flex-direction: column; gap: 12px; }
.tip-item { display: flex; align-items: center; gap: 10px; font-size: 14px; color: #666; }
</style>
