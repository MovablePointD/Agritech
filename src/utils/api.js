import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000
})

api.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = token
  }
  return config
})

api.interceptors.response.use(
  response => response.data,
  error => {
    const status = error.response?.status
    // 提取服务端返回的错误信息
    const serverMsg = error.response?.data?.error || error.response?.data?.message || ''

    if (status === 401) {
      const currentPath = router.currentRoute?.value?.path
      // 登录页面的401是密码错误/用户不存在，不要跳转和清除token
      if (currentPath !== '/login' && currentPath !== '/register') {
        localStorage.removeItem('token')
        localStorage.removeItem('user')
        router.push('/login')
        ElMessage.error('登录已过期，请重新登录')
      }
      return Promise.reject(error)
    }

    // 全局错误提示：根据状态码和服务端消息给出友好提示
    if (status === 400) {
      ElMessage.error(serverMsg || '请求参数有误')
    } else if (status === 403) {
      ElMessage.error(serverMsg || '没有权限执行此操作')
    } else if (status === 404) {
      ElMessage.error(serverMsg || '请求的资源不存在')
    } else if (status === 500) {
      ElMessage.error(serverMsg || '服务器内部错误，请稍后重试')
    } else if (status) {
      // 其他HTTP状态码
      ElMessage.error(serverMsg || `请求异常 (${status})`)
    } else {
      // 无response（网络断开、超时等）
      ElMessage.error('网络连接异常，请检查网络后重试')
    }

    return Promise.reject(error)
  }
)

export default {
  // 用户
  login: (data) => api.post('/user/login', data),
  register: (data) => api.post('/user/register', data),
  resetPassword: (data) => api.post('/user/reset-password', data),
  getUserInfo: () => api.get('/user/info'),
  updateUser: (data) => api.put('/user/', data),

  // 商品
  getProducts: (params) => api.get('/products', { params }),
  getProduct: (id) => api.get(`/product/${id}`),
  getOnSaleProducts: () => api.get('/products/on_sale'),
  getAllProducts: (params) => api.get('/admin/product/all', { params }),
  getPendingProducts: () => api.get('/admin/product/pending'),
  getRejectedProducts: () => api.get('/admin/product/rejected'),
  getHiddenProducts: () => api.get('/admin/product/hidden'),
  auditProduct: (id, data) => api.post(`/admin/product/audit/${id}`, data),
  restoreProduct: (id) => api.post(`/admin/product/restore/${id}`),
  offShelfProduct: (id) => api.post(`/admin/product/offshelf/${id}`),
  createProduct: (data) => api.post('/product/', data),
  updateProduct: (data) => api.put('/product/', data),
  deleteProduct: (id) => api.delete(`/product/${id}`),
  getMyProducts: () => api.get('/product/my'),

  // 动态
  getPosts: (params) => api.get('/posts', { params }),
  getHotPosts: (params) => api.get('/posts/hot', { params }),
  getPost: (id) => api.get(`/post/${id}`),
  getAllPosts: (params) => api.get('/admin/post/all', { params }),
  getPendingPosts: () => api.get('/admin/post/pending'),
  getRejectedPosts: () => api.get('/admin/post/rejected'),
  getDeletedPosts: () => api.get('/admin/post/deleted'),
  auditPost: (id, data) => api.post(`/admin/post/audit/${id}`, data),
  restorePost: (id) => api.post(`/admin/post/restore/${id}`),
  createPost: (data) => api.post('/post/', data),
  updatePost: (data) => api.put('/post/', data),
  deletePost: (id) => api.delete(`/post/${id}`),
  likePost: (id) => api.post(`/post/like/${id}`),
  getMyPosts: () => api.get('/post/my'),
  // 动态草稿
  createDraftPost: (data) => api.post('/post/draft', data),
  updateDraftPost: (data) => api.put('/post/draft', data),
  getDraftPosts: () => api.get('/post/drafts'),
  deleteDraftPost: (id) => api.delete(`/post/draft/${id}`),
  publishDraftPost: (id) => api.post(`/post/draft/publish/${id}`),

  // 动态评论
  getPostComments: (postId) => api.get(`/post/${postId}/comments`),
  createComment: (data) => api.post('/post/comment', data),
  updateComment: (data) => api.put('/post/comment', data),
  deleteComment: (id) => api.delete(`/post/comment/${id}`),
  likeComment: (id) => api.post(`/post/comment/like/${id}`),

  // 专家
  getExperts: (params) => api.get('/experts', { params }),
  getExpert: (id) => api.get(`/expert/${id}`),
  getExpertsByProfession: (profession) => api.get(`/experts/profession/${profession}`),
  searchExperts: (keyword) => api.get('/experts/search', { params: { keyword } }),
  applyExpert: (data) => api.post('/expert/apply', data),
  getMyExpert: () => api.get('/expert/my'),

  // 管理员专家管理
  getPendingExperts: () => api.get('/admin/expert/pending'),
  getAllExperts: () => api.get('/admin/expert/all'),
  approveExpert: (id) => api.post(`/admin/expert/approve/${id}`),
  rejectExpert: (id) => api.post(`/admin/expert/reject/${id}`),

  // 地址
  getAddresses: () => api.get('/address/'),
  getDefaultAddress: () => api.get('/address/default'),
  getAddress: (id) => api.get(`/address/${id}`),
  createAddress: (data) => api.post('/address/', data),
  updateAddress: (data) => api.put('/address/', data),
  deleteAddress: (id) => api.delete(`/address/${id}`),
  setDefaultAddress: (id) => api.post(`/address/default/${id}`),

  // 订单
  getOrders: (params) => api.get('/orders', { params }),
  getOrder: (id) => api.get(`/order/${id}`),
  createOrder: (data) => api.post('/order/', data),
  updateOrder: (data) => api.put('/order/', data),
  deleteOrder: (id) => api.delete(`/order/${id}`),
  payOrder: (id) => api.post(`/order/pay/${id}`),
  cancelOrder: (id) => api.post(`/order/cancel/${id}`),
  confirmReceive: (id) => api.post(`/order/confirm/${id}`),
  getMyOrders: () => api.get('/order/my'),
  getMySellerOrders: () => api.get('/order/seller'),
  shipOrder: (id, data) => api.post(`/order/ship/${id}`, data),

  // 知识
  getKnowledgeList: (params) => api.get('/knowledge', { params }),
  getHotKnowledge: (params) => api.get('/knowledge/hot', { params }),
  getKnowledge: (id) => api.get(`/knowledge/${id}`),
  createKnowledge: (data) => api.post('/knowledge/', data),
  updateKnowledge: (id, data) => api.put(`/knowledge/${id}`, data),
  deleteKnowledge: (id) => api.delete(`/knowledge/${id}`),
  likeKnowledge: (id) => api.post(`/knowledge/like/${id}`),
  getMyKnowledge: () => api.get('/knowledge/my'),
  // 知识草稿
  createDraftKnowledge: (data) => api.post('/knowledge/draft', data),
  updateDraftKnowledge: (data) => api.put('/knowledge/draft', data),
  getDraftKnowledge: () => api.get('/knowledge/drafts'),
  deleteDraftKnowledge: (id) => api.delete(`/knowledge/draft/${id}`),
  publishDraftKnowledge: (id) => api.post(`/knowledge/draft/publish/${id}`),
  getAllKnowledge: (params) => api.get('/admin/knowledge/all', { params }),
  getPendingKnowledge: () => api.get('/admin/knowledge/pending'),
  getRejectedKnowledge: () => api.get('/admin/knowledge/rejected'),
  getDeletedKnowledge: () => api.get('/admin/knowledge/deleted'),
  auditKnowledge: (id, data) => api.post(`/admin/knowledge/audit/${id}`, data),
  restoreKnowledge: (id) => api.post(`/admin/knowledge/restore/${id}`),

  // 管理员事务管理
  getRejectedAffairs: () => api.get('/admin/affair/rejected'),
  getProcessingAffairs: () => api.get('/admin/affair/processing'),
  getCompletedAffairs: () => api.get('/admin/affair/completed'),

  // 商品评论
  getProductComments: (productId) => api.get(`/comment_product/${productId}`),
  createProductComment: (data) => api.post('/comment_product/', data),
  likeProductComment: (id) => api.post(`/comment_product/like/${id}`),

  // 知识评论
  getKnowledgeComments: (knowledgeId) => api.get(`/comment_knowledge/${knowledgeId}`),
  createKnowledgeComment: (data) => api.post('/comment_knowledge/', data),
  likeKnowledgeComment: (id) => api.post(`/comment_knowledge/like/${id}`),

  // 文件上传
  upload: (formData) => api.post('/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  }),
  uploadVideo: (formData) => api.post('/upload/video', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 120000 // 视频上传超时2分钟
  }),

  // 购物车
  getCart: () => api.get('/cart/'),
  getCartItem: (id) => api.get(`/cart/${id}`),
  addToCart: (data) => api.post('/cart/', data),
  batchAddToCart: (data) => api.post('/cart/batch', data),
  updateCartItem: (id, data) => api.put(`/cart/${id}`, data),
  removeFromCart: (id) => api.delete(`/cart/${id}`),
  clearCart: () => api.delete('/cart/'),
  getCartCount: () => api.get('/cart/count'),
  checkoutCart: (data) => api.post('/cart/checkout', data),
  directBuy: (data) => api.post('/cart/direct_buy', data),

  // 农村事务
  getRuralAffairs: (params) => api.get('/affairs', { params }),
  getRuralAffair: (id) => api.get(`/affair/${id}`),
  getAffairFullDetail: (id) => api.get(`/affair/${id}/full`),
  createRuralAffair: (data) => api.post('/affair/', data),
  updateRuralAffair: (data) => api.put('/affair/', data),
  deleteRuralAffair: (id) => api.delete(`/affair/${id}`),
  getMyRuralAffairs: (params) => api.get('/affair/my', { params }),
  auditRuralAffair: (id, data) => api.post(`/affair/audit/${id}`, data),
  startProcessRuralAffair: (id) => api.post(`/affair/start-process/${id}`),
  processRuralAffair: (id, data) => api.post(`/affair/process/${id}`, data),
  // 修改
  modifyRuralAffair: (id, data) => api.put(`/affair/${id}/modify`, data),
  getAffairModifications: (id) => api.get(`/affair/${id}/modifications`),
  // 追问追答
  addFollowUpQuestion: (id, data) => api.post(`/affair/${id}/follow-up`, data),
  addFollowUpAnswer: (id, data) => api.post(`/affair/${id}/follow-up/answer`, data),
  getAffairFollowUps: (id) => api.get(`/affair/${id}/follow-ups`),
  // 确认完成
  confirmCompleteAffair: (id) => api.post(`/affair/${id}/confirm`),
  // 申诉
  createAppeal: (id, data) => api.post(`/affair/${id}/appeal`, data),
  addAppealMaterial: (id, data) => api.post(`/affair/appeal/${id}/material`, data),
  // 管理员申诉处理
  adminProcessAppeal: (id, data) => api.post(`/admin/affair/appeal/${id}/process`, data),
  adminGetAppeals: (params) => api.get('/admin/affair/appeals', { params }),

  // 管理员事务管理
  getPendingAuditAffairs: (params) => api.get('/admin/affair/pending-audit', { params }),
  getPendingProcessAffairs: (params) => api.get('/admin/affair/pending-process', { params }),
  getAllRuralAffairs: (params) => api.get('/admin/affair/all', { params }),

  // 通知
  getNotifications: (params) => api.get('/notification/', { params }),
  getUnreadCount: () => api.get('/notification/unread-count'),
  markAsRead: (id) => api.post(`/notification/read/${id}`),
  markAllAsRead: () => api.post('/notification/read-all'),

  // 事务处理人员
  applyProcessor: (data) => api.post('/affair-processor/apply', data),
  getMyProcessor: () => api.get('/affair-processor/my'),
  checkIsProcessor: () => api.get('/affair-processor/check'),
  getRecommendedAffairs: (params) => api.get('/affair-processor/recommended', { params }),
  getMyHandledAffairs: (params) => api.get('/affair-processor/handled', { params }),

  // 管理员处理人员管理
  getPendingProcessors: (params) => api.get('/admin/processor/pending', { params }),
  getAllProcessors: (params) => api.get('/admin/processor/all', { params }),
  approveProcessor: (id) => api.post(`/admin/processor/approve/${id}`),
  rejectProcessor: (id, data) => api.post(`/admin/processor/reject/${id}`, data),
  disableProcessor: (id) => api.post(`/admin/processor/disable/${id}`),

  // 敏感词管理
  getSensitiveWords: () => api.get('/admin/sensitive-word/'),
  addSensitiveWord: (data) => api.post('/admin/sensitive-word/', data),
  updateSensitiveWord: (data) => api.put('/admin/sensitive-word/', data),
  deleteSensitiveWord: (data) => api.delete('/admin/sensitive-word/', data),

  // 自动审核设置
  getAutoReviewSetting: () => api.get('/admin/auto-review/setting'),
  setAutoReviewSetting: (data) => api.post('/admin/auto-review/setting', data),

  // 内容检测
  checkContent: (text) => api.get('/check-content', { params: { text } }),

  // 私信
  getConversations: () => api.get('/message/conversations'),
  createConversation: (data) => api.post('/message/conversations', data),
  getConversationDetail: (id) => api.get(`/message/conversation/${id}`),
  deleteConversation: (id) => api.delete(`/message/conversation/${id}`),
  getMessages: (id, params) => api.get(`/message/messages/${id}`, { params }),
  sendMessage: (data) => api.post('/message/send', data),
  markMessageAsRead: (id) => api.post(`/message/read/${id}`),
  getMessageUnreadCount: () => api.get('/message/unread-count'),
  // 专家咨询
  createConsultation: (data) => api.post('/message/consultation', data),

  // 用户列表（用于私信搜索等）
  getUserList: (params) => api.get('/user/list', { params }),

  // 用户管理（系统管理员）
  getAllUsersForAdmin: (params) => api.get('/admin/user/all', { params }),
  banUser: (id, data) => api.post(`/admin/user/${id}/ban`, data),
  unbanUser: (id) => api.post(`/admin/user/${id}/unban`),
  deleteAccount: () => api.post('/user/delete-account'),
  getDeletedUsers: (params) => api.get('/admin/user/deleted', { params }),
  updateDeletedUser: (id, data) => api.put(`/admin/user/deleted/${id}`, data),

  // 农村信息介绍
  getRuralInfos: (params) => api.get('/info', { params }),
  getRuralInfo: (id) => api.get(`/info/${id}`),
  createRuralInfo: (data) => api.post('/info/', data),
  updateRuralInfo: (data) => api.put('/info/', data),
  deleteRuralInfo: (id) => api.delete(`/info/${id}`),
  adminAllRuralInfos: (params) => api.get('/admin/info/all', { params }),
  auditRuralInfo: (id, data) => api.post(`/admin/info/audit/${id}`, data),

  // 农村政策公告
  getPolicyNotices: (params) => api.get('/policy', { params }),
  getPolicyNotice: (id) => api.get(`/policy/${id}`),
  createPolicyNotice: (data) => api.post('/policy/', data),
  updatePolicyNotice: (data) => api.put('/policy/', data),
  deletePolicyNotice: (id) => api.delete(`/policy/${id}`),
  adminAllPolicyNotices: (params) => api.get('/admin/policy/all', { params }),
  auditPolicyNotice: (id, data) => api.post(`/admin/policy/audit/${id}`, data),

  // 农村信息评论
  getCommentRuralInfos: (params) => api.get('/comment_rural', { params }),
  createCommentRuralInfo: (data) => api.post('/comment_rural/', data),
  updateCommentRuralInfo: (data) => api.put('/comment_rural/', data),
  deleteCommentRuralInfo: (id) => api.delete(`/comment_rural/${id}`),
  likeCommentRuralInfo: (id) => api.post(`/comment_rural/like/${id}`),

  // 地址关联查询：根据农村地址获取关联的政策公告和事务
  getAssociatedByAddress: (address) => api.get('/info/associated', { params: { address } }),
}
