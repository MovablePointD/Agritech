<template>
  <div class="knowledge-detail-page">
    <AppHeader />

    <div class="container main" v-loading="loading">
      <div v-if="knowledge" class="content">
        <el-card>
          <div class="breadcrumb">
            <el-button text @click="$router.push('/knowledge')">
              <el-icon><ArrowLeft /></el-icon> 返回知识列表
            </el-button>
          </div>
          <div class="article-header">
            <h1 class="title">{{ knowledge.title }}</h1>
            <div class="meta">
              <div class="author-info">
                <el-avatar :src="knowledge.user?.avatar_url" :size="40">{{ knowledge.user?.username?.[0] }}</el-avatar>
                <div class="author-detail">
                  <span class="author-name">{{ knowledge.user?.nickname || knowledge.user?.username || '匿名' }}</span>
                  <span class="date">{{ formatTime(knowledge.created_at) }}</span>
                </div>
              </div>
            </div>
          </div>

          <div class="article-cover" v-if="coverImage">
            <el-image :src="coverImage" fit="cover" :preview-src-list="[coverImage]" />
          </div>

          <div class="article-video" v-if="knowledge.video_url">
            <video :src="knowledge.video_url" controls class="knowledge-video"></video>
          </div>

          <div class="article-content" v-html="knowledge.content"></div>

          <div class="article-footer">
            <div class="footer-actions">
              <el-button @click="handleLike" :type="liked ? 'primary' : ''">
                <el-icon><Star /></el-icon> {{ liked ? '已点赞' : '点赞' }} {{ knowledge.likes || 0 }}
              </el-button>
              <el-button @click="$router.push('/knowledge')">返回列表</el-button>
            </div>
          </div>
        </el-card>

        <!-- 知识评论 -->
        <el-card class="comment-section">
          <template #header>
            <span>知识评论 ({{ totalComments }})</span>
          </template>

          <div class="comment-list">
            <!-- 一级评论 -->
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
                    <span @click="likeComment(item.id)"><el-icon><Star /></el-icon> {{ item.likes || 0 }}</span>
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
                          <span @click="likeComment(child.id)"><el-icon><Star /></el-icon> {{ child.likes || 0 }}</span>
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
                                <span @click="likeComment(grandchild.id)"><el-icon><Star /></el-icon> {{ grandchild.likes || 0 }}</span>
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

      <div v-else class="empty-state">
        <el-icon :size="80" color="#dcdfe6"><Document /></el-icon>
        <p>知识不存在或已被删除</p>
        <el-button type="primary" @click="$router.push('/knowledge')">返回列表</el-button>
      </div>
    </div>

    <AppFooter />
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Document, Star, ArrowLeft } from '@element-plus/icons-vue'
import api from '@/utils/api'
import { getToken } from '@/utils/auth'

const route = useRoute()
const router = useRouter()
const token = getToken()

const knowledge = ref(null)
const loading = ref(false)
const liked = ref(false)

// 智能封面图：优先image_url，否则从HTML内容提取第一张图片
const coverImage = computed(() => {
  if (!knowledge.value) return null
  if (knowledge.value.image_url) return knowledge.value.image_url
  if (!knowledge.value.content) return null
  const match = knowledge.value.content.match(/<img[^>]+src="([^">]+)"/)
  return match ? match[1] : null
})
const comments = ref([])
const commentContent = ref('')

// 回复相关
const replyingTo = ref(null)
const replyToUser = ref('')
const replyContent = ref('')

const formatTime = (time) => {
  if (!time) return ''
  return new Date(time).toLocaleString('zh-CN')
}

// 计算总评论数
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

// 处理评论数据
const processComments = (list) => {
  if (!Array.isArray(list)) {
    return []
  }
  return list.map(c => {
    if (!c) return null
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
  }).filter(c => c !== null)
}

const fetchData = async () => {
  loading.value = true
  try {
    const id = route.params.id
    // 分别获取知识和评论
    const [knowledgeRes, commentRes] = await Promise.allSettled([
      api.getKnowledge(id),
      api.getKnowledgeComments(id)
    ])
    
    // 处理知识数据
    if (knowledgeRes.status === 'fulfilled' && knowledgeRes.value) {
      const res = knowledgeRes.value
      liked.value = res.is_liked || false
      const k = res.knowledge || res
      k.image_url = k.image_url?.startsWith('/') ? k.image_url : '/' + (k.image_url || '')
      knowledge.value = k
    } else {
      knowledge.value = null
    }
    
    // 处理评论数据
    if (commentRes.status === 'fulfilled') {
      comments.value = processComments(Array.isArray(commentRes.value) ? commentRes.value : [])
    } else {
      comments.value = []
    }
  } catch (error) {
    console.error('获取数据失败:', error)
  } finally {
    loading.value = false
  }
}

const submitComment = async () => {
  if (!commentContent.value.trim()) return
  try {
    await api.createKnowledgeComment({
      knowledge_id: knowledge.value.id,
      content: commentContent.value,
      parent_id: 0
    })
    ElMessage.success('评论成功')
    commentContent.value = ''
    fetchData()
  } catch (error) {
    ElMessage.error('评论失败')
  }
}

// 显示回复表单
const showReplyForm = (commentId, username, currentLevel) => {
  if (!token) {
    ElMessage.warning('请先登录')
    return
  }
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
    await api.createKnowledgeComment({
      knowledge_id: knowledge.value.id,
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

const handleLike = async () => {
  if (!token) {
    ElMessage.warning('请先登录')
    return
  }
  try {
    const res = await api.likeKnowledge(knowledge.value.id)
    liked.value = res.liked
    knowledge.value.likes = res.likes
    ElMessage.success(res.liked ? '点赞成功' : '已取消点赞')
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

const likeComment = async (id) => {
  if (!token) {
    ElMessage.warning('请先登录')
    return
  }
  try {
    const res = await api.likeKnowledgeComment(id)
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
.knowledge-detail-page { min-height: 100vh; background: #f5f5f5; }
.breadcrumb { margin-bottom: 15px; }
.container { max-width: 900px; margin: 0 auto; padding: 0 20px; }
.nav a:hover, .nav a.active { color: #409eff; }

.user-area { display: flex; gap: 10px; align-items: center; }
.user-info { display: flex; align-items: center; gap: 8px; cursor: pointer; }

.main { padding: 30px 20px; }

.content { display: flex; flex-direction: column; gap: 20px; }

.article-header { margin-bottom: 30px; }
.title { font-size: 28px; color: #333; margin-bottom: 20px; line-height: 1.4; }

.meta { display: flex; align-items: center; gap: 20px; }
.author-info { display: flex; align-items: center; gap: 12px; }
.author-detail { display: flex; flex-direction: column; }
.author-name { font-size: 14px; color: #333; font-weight: 500; }
.date { font-size: 12px; color: #999; }

.article-cover { margin-bottom: 30px; border-radius: 8px; overflow: hidden; }
.article-cover .el-image { width: 100%; max-height: 400px; }

.article-video { margin-bottom: 30px; }
.knowledge-video { width: 100%; max-width: 800px; border-radius: 12px; background: #000; }

.article-content {
  font-size: 16px;
  line-height: 1.8;
  color: #333;
  white-space: pre-wrap;
}
.article-content p { margin-bottom: 16px; line-height: 1.8; }
.article-content img { max-width: 100%; border-radius: 8px; margin: 16px 0; }
.article-content blockquote {
  margin: 16px 0;
  padding: 12px 20px;
  background: #f8f9fa;
  border-left: 4px solid #409eff;
  border-radius: 0 8px 8px 0;
  color: #666;
}
.article-content blockquote p { margin-bottom: 0; }
.article-content ul, .article-content ol {
  margin: 16px 0;
  padding-left: 24px;
}
.article-content li { margin: 8px 0; line-height: 1.8; }
.article-content h1, .article-content h2, .article-content h3,
.article-content h4, .article-content h5, .article-content h6 {
  margin-top: 24px;
  margin-bottom: 16px;
  line-height: 1.4;
}
.article-content pre {
  background: #f6f8fa;
  padding: 16px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 16px 0;
}
.article-content code {
  background: #f6f8fa;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 14px;
}
.article-content pre code {
  background: none;
  padding: 0;
}

.article-footer {
  margin-top: 40px;
  padding-top: 30px;
  border-top: 1px solid #eee;
}
.footer-actions {
  display: flex;
  justify-content: center;
  gap: 12px;
}

/* 评论样式 */
.comment-section { margin-top: 0; }
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

/* 三级评论 */
.level-3 { margin-top: 10px; padding-left: 12px; border-left: 1px dashed #ddd; }
.level-3 .comment-item { padding: 8px 0; border-bottom: none; }
.level-3 .grandchild { display: flex; gap: 8px; }

.login-tip { display: flex; align-items: center; gap: 10px; color: #666; }

.empty-state {
  background: #fff;
  border-radius: 12px;
  padding: 80px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
  color: #909399;
}
</style>
