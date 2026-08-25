<template>
  <div class="post-detail">
    <AppHeader />

    <div class="container main" v-loading="loading">
      <el-card v-if="post">
        <div class="breadcrumb">
          <el-button text @click="$router.push('/posts')">
            <el-icon><ArrowLeft /></el-icon> 返回动态列表
          </el-button>
        </div>
        <div class="post-header">
          <el-tag :type="getTypeTag(post.type)">{{ getTypeName(post.type) }}</el-tag>
          <h1>{{ post.title }}</h1>
          <div class="author">
            <el-avatar :src="post.user?.avatar_url" :size="40">{{ post.user?.username?.[0] }}</el-avatar>
            <div class="author-info">
              <span class="name">{{ post.user?.username }}</span>
              <span class="time">{{ formatTime(post.created_at) }}</span>
            </div>
          </div>
        </div>

        <div class="post-content" v-html="formatContent(post.content)"></div>

        <!-- 图片展示区域 -->
        <div class="images-section" v-if="postImages.length > 0">
          <!-- 单图展示 -->
          <div v-if="postImages.length === 1" class="images-single">
            <el-image
              :src="postImages[0]"
              :preview-src-list="postImages"
              :initial-index="0"
              fit="cover"
              class="single-image"
              preview-teleported
            />
          </div>
          <!-- 多图网格展示 -->
          <div v-else class="images-grid" :class="`images-${Math.min(postImages.length, 3)}`">
            <div
              v-for="(img, i) in postImages"
              :key="i"
              class="image-item"
              @click="previewImage(i)"
            >
              <el-image
                :src="img"
                fit="cover"
                class="grid-image"
                loading="lazy"
              />
              <!-- 图片索引标签 -->
              <div class="image-index" v-if="postImages.length > 3 && i === 2">
                +{{ postImages.length - 3 }}
              </div>
            </div>
          </div>
        </div>

        <!-- 视频展示区域 -->
        <div class="video-section" v-if="post.video_url">
          <video :src="post.video_url" controls class="post-video"></video>
        </div>

        <div class="actions">
          <el-button @click="handleLike" :type="liked ? 'primary' : ''">
            <el-icon><Star /></el-icon> {{ liked ? '已点赞' : '点赞' }} {{ post.likes }}
          </el-button>
          <el-button @click="$router.push('/posts')">返回列表</el-button>
        </div>
      </el-card>

      <!-- 评论区域 -->
      <el-card class="comment-card">
        <template #header>
          <span>评论 ({{ totalComments }})</span>
        </template>

        <div class="comment-list">
          <!-- 遍历评论树 -->
          <div v-for="item in comments" :key="item.id" class="comment-item level-1">
            <div class="comment-main">
              <el-avatar :src="item.user?.avatar_url" :size="36">{{ item.user?.username?.[0] }}</el-avatar>
              <div class="comment-body">
                <div class="comment-header">
                  <span class="name">{{ item.user?.username || item.user?.nickname || '未知用户' }}</span>
                  <span class="time">{{ formatTime(item.created_at) }}</span>
                </div>
                <div class="comment-text">{{ item.content }}</div>
                <div class="comment-actions">
                  <span @click="likeComment(item.id)"><el-icon><Star /></el-icon> {{ item.likes }}</span>
                  <span @click="showReplyForm(item.id, item.user?.username, 1)" class="reply-btn">回复</span>
                </div>

                <!-- 回复表单 -->
                <div v-if="replyingTo === item.id" class="reply-form">
                  <el-input v-model="replyContent" type="textarea" :placeholder="`回复 @${replyToUser}...`" :rows="2" />
                  <div class="reply-form-actions">
                    <el-button size="small" @click="cancelReply">取消</el-button>
                    <el-button type="primary" size="small" @click="submitReply">发送</el-button>
                  </div>
                </div>

                <!-- 二级评论 -->
                <div v-if="item.children?.length" class="children level-2">
                  <div v-for="child in item.children" :key="child.id" class="comment-item child">
                    <el-avatar :src="child.user?.avatar_url" :size="28">{{ child.user?.username?.[0] }}</el-avatar>
                    <div class="comment-body">
                      <div class="comment-header">
                        <span class="name">{{ child.user?.username || child.user?.nickname || '未知用户' }}</span>
                        <span class="time">{{ formatTime(child.created_at) }}</span>
                      </div>
                      <div class="comment-text">
                        <span v-if="child.reply_to_user" class="reply-at">回复 @{{ child.reply_to_user }}：</span>{{ child.content }}
                      </div>
                      <div class="comment-actions">
                        <span @click="likeComment(child.id)"><el-icon><Star /></el-icon> {{ child.likes }}</span>
                        <span @click="showReplyForm(child.id, child.user?.username, 2)" class="reply-btn">回复</span>
                      </div>

                      <!-- 二级回复表单 -->
                      <div v-if="replyingTo === child.id" class="reply-form">
                        <el-input v-model="replyContent" type="textarea" :placeholder="`回复 @${replyToUser}...`" :rows="2" />
                        <div class="reply-form-actions">
                          <el-button size="small" @click="cancelReply">取消</el-button>
                          <el-button type="primary" size="small" @click="submitReply">发送</el-button>
                        </div>
                      </div>

                      <!-- 三级评论（最后一层，不可回复） -->
                      <div v-if="child.children?.length" class="children level-3">
                        <div v-for="grandchild in child.children" :key="grandchild.id" class="comment-item grandchild">
                          <el-avatar :src="grandchild.user?.avatar_url" :size="24">{{ grandchild.user?.username?.[0] }}</el-avatar>
                          <div class="comment-body">
                            <div class="comment-header">
                              <span class="name">{{ grandchild.user?.username || grandchild.user?.nickname || '未知用户' }}</span>
                              <span class="time">{{ formatTime(grandchild.created_at) }}</span>
                            </div>
                            <div class="comment-text">
                              <span v-if="grandchild.reply_to_user" class="reply-at">回复 @{{ grandchild.reply_to_user }}：</span>{{ grandchild.content }}
                            </div>
                            <div class="comment-actions">
                              <span @click="likeComment(grandchild.id)"><el-icon><Star /></el-icon> {{ grandchild.likes }}</span>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <el-empty v-if="comments.length === 0" description="暂无评论，快来抢沙发吧！" />
        </div>

        <div class="comment-form-wrapper">
          <div v-if="token" class="comment-form">
            <el-input v-model="commentContent" type="textarea" placeholder="写下你的评论..." :rows="3" />
            <el-button type="primary" @click="submitComment" style="margin-top: 10px;">发表评论</el-button>
          </div>
          <div v-else class="login-tip">
            <router-link to="/login"><el-button type="primary">登录</el-button></router-link>
            <span>后参与评论</span>
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Star, ArrowLeft } from '@element-plus/icons-vue'
import api from '@/utils/api'
import { getToken } from '@/utils/auth'
import { marked } from 'marked'
import hljs from 'highlight.js'
import 'highlight.js/styles/github.css'

// 配置 marked：支持 GFM、换行转 <br>
marked.setOptions({
  breaks: true,
  gfm: true
})

const route = useRoute()
const router = useRouter()
const token = computed(() => getToken())

const post = ref(null)
const loading = ref(false)
const comments = ref([])
const commentContent = ref('')
const liked = ref(false)

// 回复相关
const replyingTo = ref(null) // 当前回复的评论ID
const replyToUser = ref('') // 当前回复的用户名
const replyContent = ref('')

const getTypeName = (t) => ({ normal: '普通', question: '提问', share: '分享' }[t] || '普通')
const getTypeTag = (t) => ({ normal: '', question: 'warning', share: 'success' }[t] || '')
const formatTime = (t) => t ? new Date(t).toLocaleString('zh-CN') : ''

// Markdown 渲染：将 Markdown 文本转为 HTML
const formatContent = (content) => {
  if (!content) return ''
  let html = marked.parse(content)
  // 对代码块追加 highlight.js 语法高亮
  html = html.replace(/<pre><code class="language-(\w+)">([\s\S]*?)<\/code><\/pre>/g, (match, lang, code) => {
    const decoded = code
      .replace(/&lt;/g, '<')
      .replace(/&gt;/g, '>')
      .replace(/&amp;/g, '&')
      .replace(/&quot;/g, '"')
      .replace(/&#39;/g, "'")
    let highlighted
    try {
      if (lang && hljs.getLanguage(lang)) {
        highlighted = hljs.highlight(decoded, { language: lang, ignoreIllegals: true }).value
      } else {
        highlighted = hljs.highlightAuto(decoded).value
      }
    } catch {
      highlighted = decoded.replace(/</g, '&lt;').replace(/>/g, '&gt;')
    }
    return `<pre class="code-block"><code class="language-${lang} hljs">${highlighted}</code></pre>`
  })
  return html
}

// 计算总评论数（递归计算所有层级）
const countComments = (list) => {
  let count = 0
  list.forEach(c => {
    count++
    if (c.children?.length) {
      count += countComments(c.children)
    }
  })
  return count
}

const totalComments = computed(() => countComments(comments.value))

const parseImages = (imagesStr) => {
  if (!imagesStr) return []
  try {
    const images = JSON.parse(imagesStr)
    return images.map(img => img.startsWith('/') ? img : '/' + img)
  } catch (e) {
    // 如果不是 JSON 格式，可能是逗号分隔或单图
    if (typeof imagesStr === 'string') {
      const imgs = imagesStr.split(',').map(s => s.trim()).filter(Boolean)
      return imgs.map(img => img.startsWith('/') ? img : '/' + img)
    }
    return []
  }
}

// 计算属性：获取处理后的图片列表
const postImages = computed(() => {
  if (!post.value?.images) return []
  return parseImages(post.value.images)
})

// 图片预览
const imageViewerRef = ref(null)
const previewImage = (index) => {
  if (imageViewerRef.value) {
    imageViewerRef.value.setActiveItem(index)
    imageViewerRef.value.show()
  }
}

// 处理评论数据，修正头像路径
const processComments = (list) => {
  return list.map(c => {
    const item = {
      ...c,
      user: c.user ? {
        ...c.user,
        avatar_url: c.user.avatar_url?.startsWith('/') ? c.user.avatar_url : '/' + (c.user.avatar_url || '')
      } : null
    }
    if (c.children?.length) {
      item.children = processComments(c.children)
    }
    return item
  })
}

const fetchData = async () => {
  loading.value = true
  try {
    const id = route.params.id
    const [postRes, commentRes] = await Promise.all([
      api.getPost(id),
      api.getPostComments(id)
    ])
    if (postRes) {
      liked.value = postRes.is_liked || false
      const p = postRes.post || postRes
      // 处理用户头像路径
      if (p.user) {
        p.user.avatar_url = p.user.avatar_url?.startsWith('/') ? p.user.avatar_url : '/' + p.user.avatar_url
      }
      post.value = p
    }
    comments.value = processComments(Array.isArray(commentRes) ? commentRes : [])
  } finally {
    loading.value = false
  }
}

const handleLike = async () => {
  if (!token.value) {
    ElMessage.warning('请先登录')
    return
  }
  try {
    const res = await api.likePost(post.value.id)
    liked.value = res.liked
    post.value.likes = res.likes
    ElMessage.success(res.liked ? '点赞成功' : '已取消点赞')
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

const submitComment = async () => {
  if (!commentContent.value.trim()) return
  try {
    await api.createComment({ post_id: post.value.id, content: commentContent.value, parent_id: 0 })
    ElMessage.success('评论成功')
    commentContent.value = ''
    fetchData()
  } catch (error) {
    ElMessage.error('评论失败')
  }
}

// 显示回复表单
const showReplyForm = (commentId, username, currentLevel) => {
  if (!token.value) {
    ElMessage.warning('请先登录')
    return
  }
  // currentLevel 是被回复评论的层级
  // 回复后新评论的层级 = currentLevel + 1
  if (currentLevel >= 3) {
    ElMessage.info('已达最大回复层级（3级）')
    return
  }
  replyingTo.value = commentId
  replyToUser.value = username || '该用户'
  replyContent.value = ''
}

// 取消回复
const cancelReply = () => {
  replyingTo.value = null
  replyToUser.value = ''
  replyContent.value = ''
}

// 提交回复
const submitReply = async () => {
  if (!replyContent.value.trim()) return
  try {
    await api.createComment({
      post_id: post.value.id,
      content: replyContent.value,
      parent_id: replyingTo.value
    })
    ElMessage.success('回复成功')
    replyContent.value = ''
    replyingTo.value = null
    fetchData()
  } catch (error) {
    ElMessage.error('回复失败')
  }
}

const likeComment = async (id) => {
  if (!token.value) {
    ElMessage.warning('请先登录')
    return
  }
  try {
    const res = await api.likeComment(id)
    // 递归更新评论的likes
    const updateCommentLikes = (list) => {
      for (const item of list) {
        if (item.id === id) {
          item.likes = res.likes
          return true
        }
        if (item.children?.length && updateCommentLikes(item.children)) {
          return true
        }
      }
      return false
    }
    updateCommentLikes(comments.value)
    ElMessage.success(res.liked ? '点赞成功' : '已取消点赞')
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

onMounted(fetchData)
</script>

<style scoped>
.post-detail { min-height: 100vh; background: #f5f5f5; }
.breadcrumb { margin-bottom: 15px; }
.container { max-width: 800px; margin: 0 auto; padding: 0 20px; }
.main { padding: 20px; }
.main > .el-card { margin-bottom: 20px; }

.post-header h1 { font-size: 24px; margin: 15px 0; }
.author { display: flex; align-items: center; gap: 10px; margin-bottom: 20px; }
.author-info { display: flex; flex-direction: column; }
.author-info .name { font-weight: 500; }
.author-info .time { font-size: 13px; color: #999; }

.post-content { line-height: 1.8; color: #333; font-size: 15px; margin-bottom: 20px; word-wrap: break-word; }

.post-content p { margin-bottom: 16px; line-height: 1.8; }
.post-content p:last-child { margin-bottom: 0; }
.post-content img { max-width: 100%; border-radius: 8px; margin: 10px 0; }
.post-content ul, .post-content ol { padding-left: 24px; margin-bottom: 16px; }
.post-content li { margin-bottom: 8px; line-height: 1.6; }
.post-content blockquote { border-left: 4px solid #409eff; padding-left: 16px; margin: 16px 0; color: #666; background: #f9f9f9; padding: 12px 16px; border-radius: 0 4px 4px 0; }
.post-content code { background: #f5f5f5; padding: 2px 6px; border-radius: 4px; font-family: monospace; }
.post-content pre { background: #f5f5f5; padding: 16px; border-radius: 8px; overflow-x: auto; margin: 16px 0; }
.post-content h1, .post-content h2, .post-content h3, .post-content h4 { margin-top: 24px; margin-bottom: 16px; font-weight: 600; }
.post-content h1 { font-size: 24px; }
.post-content h2 { font-size: 20px; }
.post-content h3 { font-size: 18px; }
.post-content a { color: #409eff; text-decoration: none; }
.post-content a:hover { text-decoration: underline; }

/* Markdown 代码块高亮样式（marked + highlight.js 生成的 .code-block） */
.post-content :deep(.code-block) {
  background: #282c34;
  color: #abb2bf;
  padding: 15px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 10px 0;
  border: 1px solid #3e4451;
}

/* 图片展示区域 */
.images-section { margin: 20px 0; }
.images-section .el-image { border-radius: 8px; }

/* 单图展示 */
.images-single { width: 100%; max-width: 500px; }
.single-image { width: 100%; height: auto; max-height: 400px; border-radius: 12px; }

/* 多图网格展示 */
.images-grid { display: grid; gap: 8px; }
.images-1 { grid-template-columns: 1fr; }
.images-2 { grid-template-columns: repeat(2, 1fr); }
.images-3 { grid-template-columns: repeat(3, 1fr); }

.image-item {
  position: relative;
  aspect-ratio: 1;
  overflow: hidden;
  border-radius: 8px;
  cursor: pointer;
}
.grid-image { width: 100%; height: 100%; transition: transform 0.3s; }
.image-item:hover .grid-image { transform: scale(1.05); }

/* 图片索引标签（超过3张时） */
.image-index {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: #fff;
  font-weight: bold;
}

/* 视频展示区域 */
.video-section { margin: 20px 0; }
.post-video { width: 100%; max-width: 600px; border-radius: 12px; background: #000; }

.actions { margin-top: 20px; padding-top: 20px; border-top: 1px solid #eee; }

.comment-list { margin-top: 20px; }
.comment-form-wrapper { margin-top: 20px; padding-top: 20px; border-top: 1px solid #eee; }
.comment-item { padding: 15px 0; border-bottom: 1px solid #eee; }
.comment-main { display: flex; gap: 12px; }
.comment-body { flex: 1; }
.comment-header { display: flex; gap: 10px; margin-bottom: 8px; }
.comment-header .name { font-weight: 500; color: #333; }
.comment-header .time { color: #999; font-size: 13px; }
.comment-text { color: #333; line-height: 1.6; margin-bottom: 8px; word-break: break-all; }
.reply-at { color: #409eff; font-weight: 500; }
.comment-actions { display: flex; gap: 15px; color: #999; font-size: 13px; }
.comment-actions span { cursor: pointer; display: flex; align-items: center; gap: 4px; }
.comment-actions span:hover { color: #409eff; }
.reply-btn { color: #409eff; }

.reply-form { margin-top: 10px; padding: 10px; background: #f9f9f9; border-radius: 4px; }
.reply-form-actions { margin-top: 8px; display: flex; justify-content: flex-end; gap: 8px; }

/* 二级评论 */
.children { margin-top: 12px; }
.level-2 { padding-left: 15px; border-left: 2px solid #e8e8e8; }
.level-2 .comment-item { padding: 10px 0; border-bottom: 1px dashed #f0f0f0; }
.level-2 .child { display: flex; gap: 10px; }

/* 三级评论（最后一层） */
.level-3 { margin-top: 10px; padding-left: 12px; border-left: 1px dashed #ddd; }
.level-3 .comment-item { padding: 8px 0; border-bottom: none; }
.level-3 .grandchild { display: flex; gap: 8px; }

.login-tip { display: flex; align-items: center; gap: 10px; color: #666; }

.comment-card { margin-top: 0; }
</style>
