<template>
  <div class="rich-editor-page">
    <!-- 顶部工具栏 -->
    <header class="editor-header">
      <div class="header-left">
        <el-button @click="goBack" :icon="ArrowLeft" text>
          返回
        </el-button>
        <el-divider direction="vertical" />
        <span class="header-title">
          <el-icon><Document /></el-icon>
          <template v-if="isPost">发布动态</template>
          <template v-else-if="isRuralInfo">发布农村信息</template>
          <template v-else-if="isPolicyNotice">发布政策公告</template>
          <template v-else>发布知识</template>
          <span v-if="editingId" class="editing-badge">· 编辑中</span>
        </span>
      </div>
      <div class="header-right">
        <el-button-group>
          <!-- <el-button 
            @click="handleSaveDraft" 
            :loading="saving" 
            :disabled="!canPublish"
            plain
          >
            <el-icon><Folder /></el-icon>
            保存草稿
          </el-button> -->
          <el-button 
            type="primary" 
            @click="handlePublish" 
            :loading="publishing" 
            :disabled="!canPublish"
          >
            <el-icon><Upload /></el-icon>
            发布内容
          </el-button>
        </el-button-group>
      </div>
    </header>

    <!-- 主编辑区域 -->
    <div class="editor-body">
      <!-- 左侧元数据面板 -->
      <aside class="editor-sidebar">
        <div class="sidebar-card">
          <h3 class="sidebar-title">
            <el-icon><EditPen /></el-icon>
            内容信息
          </h3>

          <!-- 标题 -->
          <div class="field-group">
            <label class="field-label">标题 <span class="required">*</span></label>
            <el-input 
              v-model="form.title" 
              placeholder="请输入标题" 
              maxlength="100" 
              show-word-limit
              size="large"
            />
          </div>

          <!-- 类型选择 -->
          <div class="field-group">
            <label class="field-label">分类</label>
            <el-select v-model="form.type" style="width: 100%">
              <template v-if="isPost">
                <el-option label="普通" value="normal" />
                <el-option label="提问" value="question" />
                <el-option label="分享" value="share" />
              </template>
              <template v-else-if="isRuralInfo">
                <el-option label="综合概况" value="overview" />
                <el-option label="村情介绍" value="village_intro" />
                <el-option label="农业资源" value="resource" />
                <el-option label="乡村文化" value="culture" />
                <el-option label="交通设施" value="transportation" />
                <el-option label="教育医疗" value="education" />
                <el-option label="其他信息" value="other" />
              </template>
              <template v-else-if="isPolicyNotice">
                <el-option label="补贴政策" value="subsidy" />
                <el-option label="土地政策" value="land" />
                <el-option label="环保政策" value="environmental" />
                <el-option label="科技政策" value="technology" />
                <el-option label="综合政策" value="comprehensive" />
                <el-option label="医疗养老" value="healthcare" />
                <el-option label="教育政策" value="education" />
                <el-option label="其他政策" value="other" />
              </template>
              <template v-else>
                <el-option label="种植技术" value="planting" />
                <el-option label="养殖技术" value="breeding" />
                <el-option label="农机设备" value="machinery" />
                <el-option label="病虫害防治" value="pest" />
                <el-option label="加工储存" value="processing" />
              </template>
            </el-select>
          </div>

          <!-- 封面图片（仅知识需要） -->
          <div class="field-group" v-if="!isPost && !isRuralInfo && !isPolicyNotice">
            <label class="field-label">封面图片</label>
            <el-upload
              :action="uploadUrl"
              :headers="uploadHeaders"
              list-type="picture-card"
              :on-success="handleImageSuccess"
              :before-upload="beforeImageUpload"
              :on-remove="handleImageRemove"
              :file-list="imageFileList"
              :limit="1"
            >
              <el-icon><Plus /></el-icon>
            </el-upload>
            <span class="field-hint">支持jpg、png、webp，单张不超过5MB</span>
          </div>

          <!-- 农村地址（农村信息/政策公告必填） -->
          <div class="field-group" v-if="isRuralInfo || isPolicyNotice">
            <label class="field-label">农村地址 <span class="required">*</span></label>
            <el-input 
              v-model="form.address" 
              placeholder="请输入关联的农村地址，如：XX省XX市XX区XX村（将用于自动关联同地址的公告和事务）" 
              maxlength="200" 
              show-word-limit
            />
            <span class="field-hint">输入地址后，系统将自动匹配并展示同地址的政策公告与事务</span>
          </div>

          <!-- 政策公告专属字段 -->
          <template v-if="isPolicyNotice">
            <div class="field-group">
              <label class="field-label">发布部门</label>
              <el-input 
                v-model="form.publish_dept" 
                placeholder="如：XX省农业农村厅" 
                maxlength="100"
              />
            </div>
            <div class="field-group">
              <label class="field-label">生效日期</label>
              <el-date-picker
                v-model="form.publish_date"
                type="date"
                placeholder="选择政策生效日期"
                style="width: 100%"
                value-format="YYYY-MM-DD"
              />
            </div>
            <div class="field-group">
              <label class="field-label">置顶</label>
              <el-switch v-model="form.is_top" active-text="置顶" />
            </div>
          </template>

          <!-- 视频上传 -->
          <div class="field-group">
            <label class="field-label">视频</label>
            <el-upload
              ref="videoUploadRef"
              :auto-upload="false"
              accept="video/*"
              :limit="1"
              :on-change="handleVideoChange"
              :on-remove="handleVideoRemove"
            >
              <el-button plain>
                <el-icon><VideoCamera /></el-icon>
                选择视频
              </el-button>
              <template #tip>
                <span class="field-hint">支持mp4、avi、mov，最大100MB</span>
              </template>
            </el-upload>
            <div v-if="form.video_url" class="video-preview-card">
              <video :src="form.video_url" controls class="mini-video" />
              <el-button 
                type="danger" 
                size="small" 
                :icon="Delete"
                circle
                @click="removeVideo"
              />
            </div>
          </div>

          <!-- 标签/关键词（可选） -->
          <!-- <div class="field-group">
            <label class="field-label">标签</label>
            <el-input 
              v-model="form.tags" 
              placeholder="多个标签用逗号分隔" 
              size="default"
            />
          </div> -->
        </div>

        <!-- 字数统计卡片 -->
        <div class="sidebar-card stats-card">
          <h3 class="sidebar-title">
            <el-icon><DataAnalysis /></el-icon>
            统计信息
          </h3>
          <div class="stats-grid">
            <div class="stat-item">
              <span class="stat-value">{{ wordCount }}</span>
              <span class="stat-label">字数</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ charCount }}</span>
              <span class="stat-label">字符数</span>
            </div>
            <div class="stat-item" v-if="isKnowledge">
              <span class="stat-value">{{ imageCount }}</span>
              <span class="stat-label">封面</span>
            </div>
          </div>
        </div>

        <!-- 操作提示卡片 -->
        <div class="sidebar-card tips-card">
          <h3 class="sidebar-title">
            <el-icon><InfoFilled /></el-icon>
            编辑提示
          </h3>
          <ul class="tips-list">
            <li>使用工具栏进行文本格式化</li>
            <li>支持插入图片、表格和链接</li>
            <!-- <li>可以随时保存草稿稍后编辑</li> -->
            <li>发布后将进入审核流程</li>
            <li>专业内容建议详细配图说明</li>
          </ul>
        </div>
      </aside>

      <!-- 中央编辑区 -->
      <main class="editor-main">
        <!-- 标题展示区 -->
        <div class="title-display" v-if="form.title">
          <h1>{{ form.title }}</h1>
          <div class="title-meta">
            <el-tag size="small" :type="contentTagType">
              {{ contentTypeLabel }}
            </el-tag>
            <el-tag size="small" type="info" v-if="form.type">
              {{ typeLabel }}
            </el-tag>
            <el-tag size="small" v-if="form.address" type="success">
              <el-icon><Location /></el-icon>
              {{ form.address }}
            </el-tag>
          </div>
        </div>

        <!-- 富文本编辑器 -->
        <div class="editor-container" :class="{ 'has-title': !!form.title }">
          <QuillEditor
            v-model:content="form.content"
            contentType="html"
            theme="snow"
            :toolbar="fullToolbar"
            :placeholder="editorPlaceholder"
            class="editor-quill"
          />
        </div>

        <!-- 内容预览区 -->
        <div v-if="showPreview" class="preview-panel">
          <el-divider>
            <el-icon><View /></el-icon>
            内容预览
          </el-divider>
          <div class="preview-content" v-html="form.content" />
        </div>
      </main>
    </div>

    <!-- 底部操作栏 -->
    <footer class="editor-footer" v-if="form.title || form.content">
      <div class="footer-inner">
        <div class="footer-left">
          <el-button text @click="showPreview = !showPreview">
            <el-icon><View /></el-icon>
            {{ showPreview ? '隐藏预览' : '预览效果' }}
          </el-button>
        </div>
        <div class="footer-right">
          <span class="auto-save-hint" v-if="lastSaved">
            <el-icon><Clock /></el-icon>
            上次保存：{{ lastSaved }}
          </span>
          <el-button @click="goBack">取消</el-button>
          <!-- <el-button @click="handleSaveDraft" :loading="saving" plain>
            <el-icon><Folder /></el-icon>
            保存草稿
          </el-button> -->
          <el-button type="primary" @click="handlePublish" :loading="publishing">
            <el-icon><Upload /></el-icon>
            发布内容
          </el-button>
        </div>
      </div>
    </footer>

    <!-- 发布确认对话框 -->
    <el-dialog v-model="showConfirmDialog" title="确认发布" width="550px" :close-on-click-modal="false">
      <div class="confirm-body">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="类型">
            <el-tag size="small" :type="contentTagType">
              {{ contentTypeLabel }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="标题">{{ form.title }}</el-descriptions-item>
          <el-descriptions-item label="分类">{{ typeLabel }}</el-descriptions-item>
          <el-descriptions-item v-if="form.address" label="农村地址">{{ form.address }}</el-descriptions-item>
          <el-descriptions-item v-if="isPolicyNotice && form.publish_dept" label="发布部门">{{ form.publish_dept }}</el-descriptions-item>
          <el-descriptions-item label="字数">{{ wordCount }} 字</el-descriptions-item>
          <el-descriptions-item label="图片">{{ isKnowledge ? (imageCount + ' 张') : '通过编辑器插入' }}</el-descriptions-item>
          <el-descriptions-item label="视频">{{ form.video_url ? '已添加' : '无' }}</el-descriptions-item>
        </el-descriptions>
        <el-alert 
          type="info" 
          :closable="false" 
          show-icon 
          style="margin-top: 16px;"
          title="提示"
          description="发布后内容将进入审核流程，审核通过后才会在平台公开展示。您可以在个人中心查看内容状态。"
        />
      </div>
      <template #footer>
        <el-button @click="showConfirmDialog = false">再检查一下</el-button>
        <el-button type="primary" @click="confirmPublish" :loading="publishing">
          确认发布
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { 
  ArrowLeft, Document, Upload, Folder, EditPen, Plus, 
  VideoCamera, Delete, View, DataAnalysis, InfoFilled, Clock, Location
} from '@element-plus/icons-vue'
import { QuillEditor } from '@vueup/vue-quill'
import api from '@/utils/api'
import { getToken } from '@/utils/auth'

const router = useRouter()
const route = useRoute()

// 路由参数
const contentType = computed(() => route.params.contentType || 'post')
const editingId = computed(() => route.query.id || null)
const isPost = computed(() => contentType.value === 'post')
const isRuralInfo = computed(() => contentType.value === 'rural-info')
const isPolicyNotice = computed(() => contentType.value === 'policy-notice')
const isKnowledge = computed(() => !isPost.value && !isRuralInfo.value && !isPolicyNotice.value)

// 内容类型标签
const contentTypeLabel = computed(() => {
  if (isPost.value) return '动态'
  if (isRuralInfo.value) return '农村信息'
  if (isPolicyNotice.value) return '政策公告'
  return '知识'
})

// 内容类型标签颜色
const contentTagType = computed(() => {
  if (isPost.value) return 'success'
  if (isRuralInfo.value) return ''
  if (isPolicyNotice.value) return 'danger'
  return 'warning'
})

// 上传配置
const uploadUrl = '/api/upload'
const uploadHeaders = computed(() => ({ Authorization: getToken() || '' }))

// 富文本工具栏配置 - 完整Word级工具栏
const fullToolbar = [
  [{ header: [1, 2, 3, 4, 5, 6, false] }],
  ['bold', 'italic', 'underline', 'strike'],
  [{ color: [] }, { background: [] }],
  [{ list: 'ordered' }, { list: 'bullet' }, { list: 'check' }],
  [{ indent: '-1' }, { indent: '+1' }, { align: [] }],
  ['blockquote', 'code-block'],
  [{ script: 'sub' }, { script: 'super' }],
  ['link', 'image', 'video'],
  ['table'],
  ['clean']
]

// 编辑器占位符
const editorPlaceholder = computed(() => {
  if (isPost.value) return '开始写作...分享您的农技心得、经验或提出问题'
  if (isRuralInfo.value) return '开始写作...介绍农村基本信息、资源、文化等内容'
  if (isPolicyNotice.value) return '开始写作...发布惠农政策公告，助力乡村振兴'
  return '开始写作...分享您的专业知识、技术指导和实践经验'
})

// 类型标签
const typeLabel = computed(() => {
  if (isPost.value) {
    return { normal: '普通', question: '提问', share: '分享' }[form.type] || ''
  }
  if (isRuralInfo.value) {
    return { overview:'综合概况', village_intro:'村情介绍', resource:'农业资源', culture:'乡村文化', transportation:'交通设施', education:'教育医疗', other:'其他信息' }[form.type] || ''
  }
  if (isPolicyNotice.value) {
    return { subsidy:'补贴政策', land:'土地政策', environmental:'环保政策', technology:'科技政策', comprehensive:'综合政策', healthcare:'医疗养老', education:'教育政策', other:'其他政策' }[form.type] || ''
  }
  return { 
    planting: '种植技术', breeding: '养殖技术', machinery: '农机设备', 
    pest: '病虫害防治', processing: '加工储存' 
  }[form.type] || ''
})

// 表单数据
const form = reactive({
  id: null,
  title: '',
  type: isPost.value ? 'normal' : isRuralInfo.value ? 'overview' : isPolicyNotice.value ? 'subsidy' : 'planting',
  content: '',
  image_url: '',
  video_url: '',
  tags: '',
  // 农村信息/政策公告 地址字段
  address: '',
  // 政策公告专属
  publish_dept: '',
  publish_date: '',
  is_top: false
})

// 图片列表
const imagesArray = ref([])
const imageFileList = computed(() => 
  imagesArray.value.map(url => ({ url }))
)

// 视频文件
const videoFile = ref(null)
const videoUploadRef = ref(null)

// 状态
const saving = ref(false)
const publishing = ref(false)
const showPreview = ref(false)
const showConfirmDialog = ref(false)
const lastSaved = ref('')

// 自动保存定时器
let autoSaveTimer = null

// 字数统计
const wordCount = computed(() => {
  if (!form.content) return 0
  const text = form.content.replace(/<[^>]*>/g, '')
  return text.replace(/\s/g, '').length
})

const charCount = computed(() => {
  if (!form.content) return 0
  return form.content.replace(/<[^>]*>/g, '').length
})

const imageCount = computed(() => imagesArray.value.length)

// 是否可以发布
const canPublish = computed(() => {
  const hasTitle = form.title.trim()
  const hasContent = form.content.replace(/<[^>]*>/g, '').trim()
  if (!hasTitle || !hasContent) return false
  // 农村信息和政策公告要求必填地址
  if ((isRuralInfo.value || isPolicyNotice.value) && !form.address.trim()) return false
  return true
})

// 加载现有内容（编辑模式）
const loadContent = async () => {
  if (!editingId.value) return
  try {
    if (isPost.value) {
      const res = await api.getPost(editingId.value)
      if (res) {
        form.id = res.id
        form.title = res.title || ''
        form.type = res.type || 'normal'
        form.content = res.content || ''
        form.video_url = res.video_url || ''
        form.tags = res.tags || ''
      }
    } else if (isRuralInfo.value) {
      const res = await api.getRuralInfo(editingId.value)
      if (res) {
        form.id = res.id
        form.title = res.title || ''
        form.type = res.type || 'overview'
        form.content = res.content || ''
        form.address = res.address || ''
        form.video_url = res.video_url || ''
      }
    } else if (isPolicyNotice.value) {
      const res = await api.getPolicyNotice(editingId.value)
      if (res) {
        form.id = res.id
        form.title = res.title || ''
        form.type = res.category || 'subsidy'
        form.content = res.content || ''
        form.address = res.address || ''
        form.publish_dept = res.publish_dept || ''
        form.publish_date = res.publish_date || ''
        form.is_top = res.is_top || false
      }
    } else {
      const res = await api.getKnowledge(editingId.value)
      if (res) {
        form.id = res.id
        form.title = res.title || ''
        form.type = res.type || 'planting'
        form.content = res.content || ''
        form.image_url = res.image_url || ''
        form.video_url = res.video_url || ''
        form.tags = res.tags || ''
        if (res.image_url) {
          imagesArray.value = [res.image_url]
        }
      }
    }
  } catch (error) {
    ElMessage.error('加载内容失败')
    router.back()
  }
}

// 返回
const goBack = () => {
  router.back()
}

// 图片上传前校验
const beforeImageUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) ElMessage.error('只能上传图片文件')
  if (!isLt5M) ElMessage.error('图片大小不能超过5MB')
  return isImage && isLt5M
}

// 图片上传成功（仅知识封面使用）
const handleImageSuccess = (res) => {
  const url = res.url
  imagesArray.value = [url]
  form.image_url = url
}

// 图片移除（仅知识封面使用）
const handleImageRemove = () => {
  imagesArray.value = []
  form.image_url = ''
}

// 视频文件选择
const handleVideoChange = (file) => {
  videoFile.value = file.raw
}

// 视频移除
const handleVideoRemove = () => {
  videoFile.value = null
}

// 删除已添加的视频
const removeVideo = () => {
  form.video_url = ''
  videoFile.value = null
}

// 上传视频
const uploadVideo = async () => {
  if (!videoFile.value) return form.video_url
  const formData = new FormData()
  formData.append('file', videoFile.value)
  formData.append('dir', 'video')
  const res = await api.uploadVideo(formData)
  return res.url
}

// 构建提交数据（按内容类型精确构建，避免发送无关字段导致后端绑定失败）
const buildSubmitData = () => {
  if (isPost.value) {
    return {
      title: form.title,
      type: form.type,
      content: form.content,
      video_url: form.video_url || '',
      tags: form.tags || ''
    }
  }
  if (isRuralInfo.value) {
    return {
      title: form.title,
      type: form.type,
      content: form.content,
      address: form.address || '',
      images: ''
    }
  }
  if (isPolicyNotice.value) {
    return {
      title: form.title,
      category: form.type,       // 政策公告用 category
      content: form.content,
      address: form.address || '',
      publish_dept: form.publish_dept || '',
      publish_date: form.publish_date || '',
      images: '',
      attachment: '',
      is_top: form.is_top || false
    }
  }
  // knowledge
  return {
    title: form.title,
    type: form.type,
    content: form.content,
    image_url: form.image_url || '',
    video_url: form.video_url || '',
    tags: form.tags || ''
  }
}

// 保存草稿
const handleSaveDraft = async () => {
  if (!canPublish.value) {
    if ((isRuralInfo.value || isPolicyNotice.value) && !form.address.trim()) {
      ElMessage.warning('请填写农村地址')
    } else {
      ElMessage.warning('请至少填写标题和内容')
    }
    return
  }
  saving.value = true
  try {
    if (videoFile.value) {
      form.video_url = await uploadVideo()
      videoFile.value = null
    }
    // rural-info 和 policy-notice 不支持草稿状态，直接走创建/更新
    const isRuralType = isRuralInfo.value || isPolicyNotice.value
    const data = isRuralType 
      ? buildSubmitData() 
      : { ...buildSubmitData(), status: 'draft' }
    
    if (form.id) {
      if (isPost.value) {
        await api.updatePost({ id: form.id, ...data })
      } else if (isRuralInfo.value) {
        await api.updateRuralInfo({ id: form.id, ...data })
      } else if (isPolicyNotice.value) {
        await api.updatePolicyNotice({ id: form.id, ...data })
      } else {
        await api.updateKnowledge(form.id, data)
      }
    } else {
      if (isPost.value) {
        const res = await api.createPost(data)
        if (res) form.id = res.id
      } else if (isRuralInfo.value) {
        const res = await api.createRuralInfo(data)
        if (res?.data) form.id = res.data.id
      } else if (isPolicyNotice.value) {
        const res = await api.createPolicyNotice(data)
        if (res?.data) form.id = res.data.id
      } else {
        const res = await api.createKnowledge(data)
        if (res) form.id = res.id
      }
    }
    lastSaved.value = new Date().toLocaleTimeString('zh-CN')
    ElMessage.success('草稿已保存')
    autoSave()
  } catch (error) {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

// 发布 - 先弹确认框
const handlePublish = () => {
  if (!canPublish.value) {
    if ((isRuralInfo.value || isPolicyNotice.value) && !form.address.trim()) {
      ElMessage.warning('请填写农村地址')
      return
    }
    ElMessage.warning('请至少填写标题和内容')
    return
  }
  showConfirmDialog.value = true
}

// 确认发布
const confirmPublish = async () => {
  publishing.value = true
  try {
    if (videoFile.value) {
      form.video_url = await uploadVideo()
      videoFile.value = null
    }
    const data = buildSubmitData()
    
    if (form.id) {
      if (isPost.value) {
        await api.updatePost({ id: form.id, ...data })
      } else if (isRuralInfo.value) {
        await api.updateRuralInfo({ id: form.id, ...data })
      } else if (isPolicyNotice.value) {
        await api.updatePolicyNotice({ id: form.id, ...data })
      } else {
        await api.updateKnowledge(form.id, data)
      }
    } else {
      if (isPost.value) {
        await api.createPost(data)
      } else if (isRuralInfo.value) {
        await api.createRuralInfo(data)
      } else if (isPolicyNotice.value) {
        await api.createPolicyNotice(data)
      } else {
        await api.createKnowledge(data)
      }
    }
    showConfirmDialog.value = false
    ElMessage.success('发布成功，请等待审核')
    
    // 清除自动保存
    if (autoSaveTimer) clearInterval(autoSaveTimer)
    localStorage.removeItem(`draft_${contentType.value}_${editingId.value || 'new'}`)
    
    // 跳转回农村信息中心或个人中心
    if (isRuralInfo.value || isPolicyNotice.value) {
      router.push('/rural-info')
    } else {
      router.push('/user')
    }
  } catch (error) {
    ElMessage.error('发布失败')
  } finally {
    publishing.value = false
  }
}

// 自动保存（本地草稿）
const autoSave = () => {
  try {
    const draftKey = `draft_${contentType.value}_${editingId.value || 'new'}`
    const draft = {
      ...form,
      images: imagesArray.value,
      savedAt: new Date().toISOString()
    }
    localStorage.setItem(draftKey, JSON.stringify(draft))
  } catch { /* localStorage 可能满 */ }
}

// 恢复本地草稿
const restoreDraft = () => {
  if (editingId.value) return // 编辑已有内容时不恢复草稿
  try {
    const draftKey = `draft_${contentType.value}_new`
    const saved = localStorage.getItem(draftKey)
    if (saved) {
      const draft = JSON.parse(saved)
      const defaultType = isPost.value ? 'normal' : isRuralInfo.value ? 'overview' : isPolicyNotice.value ? 'subsidy' : 'planting'
      Object.assign(form, {
        title: draft.title || '',
        type: draft.type || defaultType,
        content: draft.content || '',
        video_url: draft.video_url || '',
        tags: draft.tags || '',
        address: draft.address || '',
        publish_dept: draft.publish_dept || '',
        publish_date: draft.publish_date || '',
        is_top: draft.is_top || false
      })
      if (isKnowledge.value) {
        form.image_url = draft.image_url || ''
        imagesArray.value = draft.image_url ? [draft.image_url] : []
      }
      ElMessage.info('已恢复上次未发布的草稿')
    }
  } catch { /* 忽略 */ }
}

// 键盘快捷键
const handleKeydown = (e) => {
  if ((e.ctrlKey || e.metaKey) && e.key === 's') {
    e.preventDefault()
    handleSaveDraft()
  }
}

onMounted(() => {
  loadContent()
  restoreDraft()
  document.addEventListener('keydown', handleKeydown)
  
  // 每60秒自动保存
  autoSaveTimer = setInterval(autoSave, 60000)
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleKeydown)
  if (autoSaveTimer) clearInterval(autoSaveTimer)
  // 离开时自动保存草稿
  autoSave()
})
</script>

<style scoped>
/* ======== 页面容器 ======== */
.rich-editor-page {
  min-height: 100vh;
  background: #f0f2f5;
  display: flex;
  flex-direction: column;
}

/* ======== 顶部工具栏 ======== */
.editor-header {
  position: sticky;
  top: 0;
  z-index: 100;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  padding: 12px 24px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  display: flex;
  align-items: center;
  gap: 6px;
}

.header-title .el-icon {
  color: #43a047;
}

.editing-badge {
  color: #e6a23c;
  font-weight: normal;
  font-size: 13px;
}

.header-right {
  display: flex;
  gap: 12px;
}

/* ======== 主体布局 ======== */
.editor-body {
  flex: 1;
  display: flex;
  gap: 24px;
  padding: 24px;
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
  box-sizing: border-box;
  align-items: flex-start;
}

/* ======== 左侧面板 ======== */
.editor-sidebar {
  width: 260px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 16px;
  position: sticky;
  top: 80px;
}

.sidebar-card {
  background: #fff;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
}

.sidebar-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 16px 0;
  display: flex;
  align-items: center;
  gap: 6px;
}

.sidebar-title .el-icon {
  color: #43a047;
}

.field-group {
  margin-bottom: 20px;
}

.field-group:last-child {
  margin-bottom: 0;
}

.field-label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: #606266;
  margin-bottom: 6px;
}

.field-label .required {
  color: #f56c6c;
}

.field-hint {
  font-size: 11px;
  color: #b0b3bb;
  margin-top: 4px;
  display: block;
}

/* 统计卡片 */
.stats-card {
  background: linear-gradient(135deg, #e8f5e9, #f1f8e9);
}

.stats-grid {
  display: flex;
  justify-content: space-around;
}

.stat-item {
  text-align: center;
}

.stat-value {
  display: block;
  font-size: 24px;
  font-weight: 700;
  color: #2e7d32;
}

.stat-label {
  font-size: 12px;
  color: #6d8c6f;
}

/* 提示卡片 */
.tips-card {
  background: #fafbfc;
}

.tips-list {
  margin: 0;
  padding: 0 0 0 16px;
  list-style: disc;
}

.tips-list li {
  font-size: 12px;
  color: #909399;
  line-height: 1.8;
}

/* ======== 中央编辑区 ======== */
.editor-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* 标题展示 */
.title-display {
  background: #fff;
  border-radius: 8px;
  padding: 24px 28px 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
}

.title-display h1 {
  margin: 0 0 8px 0;
  font-size: 24px;
  color: #1a1a1a;
  line-height: 1.4;
}

.title-meta {
  display: flex;
  gap: 6px;
}

/* 编辑器容器 */
.editor-container {
  flex: 1;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-height: 500px;
}

.editor-container.has-title {
  border-radius: 8px;
}

.editor-quill {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.editor-quill :deep(.ql-toolbar) {
  border: none;
  border-bottom: 1px solid #e4e7ed;
  padding: 10px 16px;
  background: #fafbfc;
}

.editor-quill :deep(.ql-toolbar button:hover),
.editor-quill :deep(.ql-toolbar button.ql-active) {
  color: #43a047;
}

.editor-quill :deep(.ql-toolbar button:hover .ql-stroke),
.editor-quill :deep(.ql-toolbar button.ql-active .ql-stroke) {
  stroke: #43a047;
}

.editor-quill :deep(.ql-toolbar button:hover .ql-fill),
.editor-quill :deep(.ql-toolbar button.ql-active .ql-fill) {
  fill: #43a047;
}

.editor-quill :deep(.ql-container) {
  border: none;
  font-size: 16px;
  flex: 1;
}

.editor-quill :deep(.ql-editor) {
  padding: 24px 28px;
  min-height: 400px;
  line-height: 1.8;
  font-size: 16px;
}

.editor-quill :deep(.ql-editor.ql-blank::before) {
  font-style: normal;
  color: #c0c4cc;
  font-size: 16px;
  left: 28px;
  right: 28px;
}

.editor-quill :deep(.ql-editor h1) { font-size: 2em; }
.editor-quill :deep(.ql-editor h2) { font-size: 1.5em; }
.editor-quill :deep(.ql-editor h3) { font-size: 1.17em; }
.editor-quill :deep(.ql-editor blockquote) {
  border-left: 4px solid #43a047;
  background: #f1f8e9;
  padding: 8px 16px;
  margin: 12px 0;
}

/* 预览面板 */
.preview-panel {
  background: #fff;
  border-radius: 8px;
  padding: 8px 0 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
}

.preview-content {
  padding: 0 28px;
  line-height: 1.8;
  font-size: 16px;
  color: #303133;
  word-break: break-word;
}

.preview-content :deep(img) {
  max-width: 100%;
  border-radius: 6px;
}

.preview-content :deep(table) {
  border-collapse: collapse;
  width: 100%;
}

.preview-content :deep(td),
.preview-content :deep(th) {
  border: 1px solid #e4e7ed;
  padding: 8px 12px;
}

/* 视频预览 */
.video-preview-card {
  margin-top: 10px;
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.mini-video {
  width: 200px;
  border-radius: 6px;
  border: 1px solid #e4e7ed;
}

/* ======== 底部栏 ======== */
.editor-footer {
  position: sticky;
  bottom: 0;
  z-index: 100;
  background: #fff;
  border-top: 1px solid #e4e7ed;
  padding: 12px 24px;
  box-shadow: 0 -1px 4px rgba(0, 0, 0, 0.04);
}

.footer-inner {
  max-width: 1400px;
  margin: 0 auto;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.footer-left {
  display: flex;
  align-items: center;
}

.footer-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.auto-save-hint {
  font-size: 12px;
  color: #b0b3bb;
  display: flex;
  align-items: center;
  gap: 4px;
}

/* ======== 确认对话框 ======== */
.confirm-body {
  padding: 8px 0;
}

/* ======== 响应式 ======== */
@media (max-width: 1100px) {
  .editor-sidebar {
    width: 200px;
  }
  .editor-quill :deep(.ql-editor) {
    padding: 16px 20px;
  }
}

@media (max-width: 860px) {
  .editor-body {
    flex-direction: column-reverse;
    padding: 12px;
  }
  .editor-sidebar {
    width: 100%;
    position: static;
    flex-direction: row;
    flex-wrap: wrap;
    gap: 8px;
  }
  .sidebar-card {
    flex: 1;
    min-width: 200px;
  }
}
</style>
