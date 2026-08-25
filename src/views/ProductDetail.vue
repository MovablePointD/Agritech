<template>
  <div class="product-detail">
    <AppHeader />

    <div class="container main" v-loading="loading">
      <div v-if="product" class="content">
        <el-card>
          <div class="breadcrumb">
            <el-button text @click="$router.push('/products')">
              <el-icon><ArrowLeft /></el-icon> 返回商品列表
            </el-button>
          </div>
          <div class="detail">
            <el-image :src="product.image_url" class="image" />
            <div class="info">
              <h1>{{ product.title }}</h1>
              <div class="price">¥{{ product.price }}</div>
              <div class="meta">
                <span>类型：{{ product.type }}</span>
                <span>地址：{{ product.address }}</span>
              </div>
              <!-- 库存信息（仅对上传者可见） -->
              <div v-if="isOwner" class="stock-info" :class="getStockClass(product.stock)">
                <span class="stock-icon"><el-icon><Box /></el-icon></span>
                <span>库存：{{ product.stock || 0 }}</span>
                <el-tag v-if="product.stock < 20" type="warning" size="small" effect="plain" style="margin-left: 10px;">
                  库存不足
                </el-tag>
              </div>
              <div class="desc">
                <h3>商品详情</h3>
                <p>{{ product.content }}</p>
              </div>
              <div class="actions">
                <el-button type="primary" size="large" @click="handleBuy" :disabled="!token || product.stock <= 0">立即购买</el-button>
                <el-button type="warning" size="large" @click="handleAddToCart" :disabled="!token || product.stock <= 0">加入购物车</el-button>
                <el-button size="large" @click="handleContact" :loading="contacting" :disabled="!token">联系卖家</el-button>
              </div>
            </div>
          </div>
        </el-card>

        <!-- 商品评论 -->
        <el-card class="comment-section">
          <template #header>
            <span>商品评论 ({{ totalComments }})</span>
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
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Star, Box, ArrowLeft } from '@element-plus/icons-vue'
import api from '@/utils/api'
import { getToken, getUserInfo } from '@/utils/auth'

const route = useRoute()
const router = useRouter()
const token = computed(() => getToken())
const currentUser = ref(getUserInfo() || {})
const currentUserId = computed(() => currentUser.value?.id || 0)
const isOwner = computed(() => token.value && product.value?.publisher === currentUserId.value)

const product = ref(null)
const loading = ref(false)
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

// 获取库存样式类
const getStockClass = (stock) => {
  if (!stock || stock <= 0) return 'danger'
  if (stock < 20) return 'warning'
  return ''
}

// 处理评论数据
const processComments = (list) => {
  if (!Array.isArray(list)) {
    console.warn('评论数据不是数组:', list)
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
    // 分别获取商品和评论，避免一个失败影响另一个
    const [productRes, commentRes] = await Promise.allSettled([
      api.getProduct(id),
      api.getProductComments(id)
    ])
    
    // 处理商品数据
    if (productRes.status === 'fulfilled' && productRes.value) {
      const p = productRes.value
      p.image_url = p.image_url?.startsWith('/') ? p.image_url : '/' + (p.image_url || '')
      product.value = p
    } else {
      console.error('获取商品失败:', productRes.reason)
      product.value = null
    }
    
    // 处理评论数据
    if (commentRes.status === 'fulfilled') {
      comments.value = processComments(Array.isArray(commentRes.value) ? commentRes.value : [])
    } else {
      console.error('获取评论失败:', commentRes.reason)
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
    await api.createProductComment({
      product_id: product.value.id,
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
    await api.createProductComment({
      product_id: product.value.id,
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
    const res = await api.likeProductComment(id)
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

const handleBuy = () => {
  if (product.value.stock <= 0) {
    ElMessage.warning('该商品库存不足，暂时无法购买')
    return
  }
  if (product.value.stock < 20) {
    ElMessage.warning('该商品库存不足（当前库存：' + product.value.stock + '）')
  }
  router.push({ path: '/orders', query: { product_id: product.value.id } })
}

const contacting = ref(false)
const handleContact = async () => {
  if (!token.value) {
    ElMessage.warning('请先登录')
    return
  }
  if (!product.value?.publisher) {
    ElMessage.warning('无法获取卖家信息')
    return
  }
  if (product.value.publisher === currentUserId.value) {
    ElMessage.info('这是您自己的商品')
    return
  }
  contacting.value = true
  try {
    const res = await api.createConversation({ target_user_id: product.value.publisher })
    const convId = res.conversation_id
    if (convId) {
      ElMessage.success('已创建咨询会话')
      router.push('/messages/' + convId)
    } else {
      ElMessage.error('创建会话失败，请稍后重试')
    }
  } catch (error) {
    ElMessage.error('联系卖家失败，请稍后重试')
  } finally {
    contacting.value = false
  }
}

// 加入购物车
const handleAddToCart = async () => {
  if (!token.value) {
    ElMessage.warning('请先登录')
    return
  }
  if (product.value.stock <= 0) {
    ElMessage.warning('该商品库存不足，暂时无法购买')
    return
  }
  if (product.value.stock < 20) {
    ElMessage.warning('该商品库存不足（当前库存：' + product.value.stock + '），建议尽快下单')
  }
  try {
    await api.addToCart({
      product_id: product.value.id,
      quantity: 1
    })
    ElMessage.success('已加入购物车')
  } catch (error) {
    ElMessage.error('加入购物车失败')
  }
}

onMounted(fetchData)
</script>

<style scoped>
.product-detail { min-height: 100vh; background: #f5f5f5; }
.breadcrumb { margin-bottom: 15px; }
.container { max-width: 1200px; margin: 0 auto; padding: 0 20px; }
.main { padding: 20px; }
.content { display: flex; flex-direction: column; gap: 20px; }

.detail { display: flex; gap: 30px; }
.detail .image { width: 400px; height: 400px; border-radius: 8px; flex-shrink: 0; }
.detail .info { flex: 1; }
.detail h1 { font-size: 24px; margin-bottom: 15px; }
.detail .price { font-size: 28px; color: #f56c6c; font-weight: bold; margin-bottom: 15px; }
.detail .meta { display: flex; gap: 20px; color: #666; margin-bottom: 20px; }
.stock-info {
  display: flex;
  align-items: center;
  padding: 12px 15px;
  background: #f0f9eb;
  border-radius: 8px;
  margin-bottom: 20px;
  color: #67c23a;
  font-weight: 500;
}
.stock-info .stock-icon { margin-right: 8px; font-size: 18px; }
.stock-info.warning {
  background: #fdf6ec;
  color: #e6a23c;
}
.stock-info.danger {
  background: #fef0f0;
  color: #f56c6c;
}
.detail .desc h3 { font-size: 16px; margin-bottom: 10px; }
.detail .desc p { color: #666; line-height: 1.8; }
.detail .actions { margin-top: 30px; }

.comment-section { margin-top: 20px; }
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
</style>
