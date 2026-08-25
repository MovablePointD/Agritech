<template>
  <div class="order-list">
    <AppHeader />

    <div class="container main">
      <el-card>
        <template #header>
          <span>我的订单</span>
        </template>

        <el-tabs v-model="activeTab" @tab-change="fetchData">
          <el-tab-pane label="全部" name="" />
          <el-tab-pane label="待付款" name="0" />
          <el-tab-pane label="待发货" name="1" />
          <el-tab-pane label="待收货" name="2" />
          <el-tab-pane label="已完成" name="3" />
        </el-tabs>

        <el-table v-if="!loading && list.length > 0" :data="list" v-loading="loading">
          <el-table-column prop="order_no" label="订单编号" width="180" />
          <el-table-column label="商品" min-width="200">
            <template #default="{ row }">
              <div v-if="row.items && row.items.length > 0">
                <div v-for="item in row.items" :key="item.id" class="order-item-row">
                  <span class="item-name">{{ item.product_name }}</span>
                  <span class="item-qty">× {{ item.quantity }}</span>
                  <span class="item-price">¥{{ item.subtotal?.toFixed(2) }}</span>
                </div>
              </div>
              <div v-else class="price">¥{{ row.total_price }}</div>
            </template>
          </el-table-column>
          <el-table-column label="订单类型" width="100">
            <template #default="{ row }">
              <el-tag size="small">{{ row.order_type === 'buy' ? '购买' : '出售' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="订单状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getStatusType(row.status)" size="small">{{ getStatusText(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="下单时间" width="160" :formatter="formatTime" />
          <el-table-column label="操作" width="200">
            <template #default="{ row }">
              <template v-if="row.status === 0">
                <el-button size="small" type="primary" @click="handlePay(row)">付款</el-button>
                <el-button size="small" type="danger" @click="handleCancel(row)">取消</el-button>
              </template>
              <template v-else-if="row.status === 2">
                <el-button size="small" type="success" @click="handleConfirm(row)">确认收货</el-button>
              </template>
              <el-button size="small" @click="showDetail(row)">详情</el-button>
            </template>
          </el-table-column>
        </el-table>

        <!-- 空状态 -->
        <div v-if="!loading && list.length === 0" class="empty-state">
          <div class="empty-icon">
            <svg class="empty-svg" viewBox="0 0 120 120" fill="none" xmlns="http://www.w3.org/2000/svg">
              <circle cx="60" cy="60" r="50" fill="#f5f5f5" stroke="#e0e0e0" stroke-width="2"/>
              <path d="M40 55C40 52.2386 42.2386 50 45 50H75C77.7614 50 80 52.2386 80 55V70C80 72.7614 77.7614 75 75 75H45C42.2386 75 40 72.7614 40 70V55Z" fill="#fff" stroke="#d0d0d0" stroke-width="2"/>
              <path d="M45 50L50 40H70L75 50" stroke="#d0d0d0" stroke-width="2" stroke-linecap="round"/>
              <line x1="50" y1="60" x2="70" y2="60" stroke="#d0d0d0" stroke-width="2" stroke-linecap="round"/>
              <line x1="50" y1="65" x2="65" y2="65" stroke="#d0d0d0" stroke-width="2" stroke-linecap="round"/>
              <circle cx="60" cy="60" r="25" fill="#fafafa"/>
            </svg>
          </div>
          <p class="empty-text">暂无订单</p>
          <el-button type="primary" @click="$router.push('/products')">去购物</el-button>
        </div>
      </el-card>
    </div>

    <!-- 订单详情对话框 -->
    <el-dialog v-model="showDetailDialog" title="订单详情" width="600px">
      <div v-if="detail" class="order-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="订单编号">{{ detail.order_no }}</el-descriptions-item>
          <el-descriptions-item label="订单状态">
            <el-tag :type="getStatusType(detail.status)" size="small">{{ getStatusText(detail.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="收货人">{{ detail.receiver }}</el-descriptions-item>
          <el-descriptions-item label="联系电话">{{ detail.phone }}</el-descriptions-item>
          <el-descriptions-item label="收货地址" :span="2">{{ detail.address }}</el-descriptions-item>
          <el-descriptions-item label="总金额">¥{{ (detail.total_price || 0).toFixed(2) }}</el-descriptions-item>
          <el-descriptions-item label="下单时间">{{ formatTime(detail) }}</el-descriptions-item>
        </el-descriptions>
        <el-divider>商品清单</el-divider>
        <div v-if="detail.items && detail.items.length > 0">
          <div v-for="item in detail.items" :key="item.id" class="order-item-detail">
            <span>商品：{{ item.product_name }}</span>
            <span>数量：{{ item.quantity }}</span>
            <span>单价：¥{{ item.price?.toFixed(2) }}</span>
            <span>小计：¥{{ (item.subtotal || 0).toFixed(2) }}</span>
          </div>
        </div>
        <div v-else class="empty-items">暂无商品信息</div>
      </div>
      <template #footer>
        <el-button @click="showDetailDialog = false">关闭</el-button>
        <el-button v-if="detail?.status === 0" type="primary" @click="handlePay(detail)">立即付款</el-button>
        <el-button v-if="detail?.status === 2" type="success" @click="handleConfirm(detail)">确认收货</el-button>
      </template>
    </el-dialog>

    <!-- 购买对话框 -->
    <el-dialog v-model="showBuyDialog" title="确认订单" width="500px">
      <el-form :model="buyForm" label-width="80px">
        <el-form-item label="商品">
          <div>{{ currentProduct?.title }} - ¥{{ currentProduct?.price }}</div>
        </el-form-item>
        <el-form-item label="数量">
          <el-input-number v-model="buyForm.quantity" :min="1" :max="99" />
        </el-form-item>
        <el-form-item label="收货地址">
          <el-select v-model="buyForm.address_id" placeholder="请选择地址" style="width: 100%">
            <el-option v-for="addr in addresses" :key="addr.id" :label="`${addr.receiver} ${addr.province}${addr.city}${addr.district}${addr.detail}`" :value="addr.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="buyForm.remark" type="textarea" :rows="2" placeholder="选填" />
        </el-form-item>
        <el-form-item label="总价">
          <span class="total-price">¥{{ (currentProduct?.price || 0) * buyForm.quantity }}</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showBuyDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreateOrder">提交订单</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import api from '@/utils/api'
import { getToken, getUserInfo } from '@/utils/auth'

const route = useRoute()
const router = useRouter()

const token = computed(() => getToken())

const currentUser = ref({
  username: '',
  nickname: '',
  avatar_url: ''
})

const list = ref([])
const loading = ref(false)
const activeTab = ref('')
const showDetailDialog = ref(false)
const showBuyDialog = ref(false)
const detail = ref(null)
const currentProduct = ref(null)
const addresses = ref([])

const buyForm = reactive({
  product_id: null,
  quantity: 1,
  address_id: null,
  remark: ''
})

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

// 用户下拉菜单处理
const handleUserCommand = (command) => {
  const routes = { user: '/user', orders: '/orders', address: '/address', cart: '/cart', knowledge: '/my-knowledge' }
  if (command === 'logout') {
    removeToken()
    removeUserInfo()
    ElMessage.success('已退出登录')
    router.push('/')
  } else if (routes[command]) {
    router.push(routes[command])
  }
}

const getStatusType = (s) => ({ 0: 'warning', 1: 'info', 2: 'primary', 3: 'success', 4: 'danger' }[s] || '')
const getStatusText = (s) => ({ 0: '待付款', 1: '待发货', 2: '待收货', 3: '已完成', 4: '已取消' }[s] || '')
const formatTime = (row) => row.created_at ? new Date(row.created_at).toLocaleString('zh-CN') : ''

const fetchData = async () => {
  loading.value = true
  try {
    const res = activeTab.value ? await api.getOrders({ status: activeTab.value }) : await api.getMyOrders()
    const orders = res.list || res || []
    // 去重：根据订单ID和商品ID去重items
    list.value = orders.map(order => {
      if (order.items) {
        const seen = new Set()
        order.items = order.items.filter(item => {
          const key = `${item.product_id}-${item.id}`
          if (seen.has(key)) return false
          seen.add(key)
          return true
        })
      }
      return order
    })
  } finally {
    loading.value = false
  }
}

const showDetail = (row) => {
  detail.value = row
  showDetailDialog.value = true
}

const handlePay = async (row) => {
  try {
    await api.payOrder(row.id)
    ElMessage.success('付款成功')
    fetchData()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '付款失败'
    ElMessage.error(msg)
  }
}

const handleCancel = async (row) => {
  try {
    await api.cancelOrder(row.id)
    ElMessage.success('订单已取消')
    fetchData()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '取消失败'
    ElMessage.error(msg)
  }
}

const handleConfirm = async (row) => {
  try {
    await api.confirmReceive(row.id)
    ElMessage.success('已确认收货')
    fetchData()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '确认失败'
    ElMessage.error(msg)
  }
}

const initBuyDialog = async (productId) => {
  if (productId) {
    currentProduct.value = await api.getProduct(productId)
    buyForm.product_id = productId
  }
  const addrRes = await api.getAddresses()
  addresses.value = addrRes?.list || addrRes || []
  showBuyDialog.value = true
}

const handleCreateOrder = async () => {
  if (!buyForm.address_id) {
    ElMessage.warning('请选择收货地址')
    return
  }
  try {
    // 使用直接购买API
    await api.directBuy({
      product_id: buyForm.product_id,
      quantity: buyForm.quantity,
      address_id: buyForm.address_id,
      remark: buyForm.remark
    })
    ElMessage.success('订单创建成功')
    showBuyDialog.value = false
    fetchData()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '创建失败'
    ElMessage.error(msg)
  }
}

onMounted(async () => {
  initUserInfo()
  loadUserDetail()
  await fetchData()
  const productId = route.query.product_id
  if (productId) {
    await initBuyDialog(Number(productId))
  }
})
</script>

<style scoped>
.order-list { min-height: 100vh; background: #f5f5f5; }
.container { max-width: 1200px; margin: 0 auto; padding: 0 20px; }
.main { padding: 20px; }
.price { color: #f56c6c; font-size: 14px; }
.total-price { font-size: 24px; color: #f56c6c; font-weight: bold; }
.order-item-row { display: flex; justify-content: space-between; align-items: center; padding: 2px 0; }
.order-item-row .item-name { flex: 1; color: #333; }
.order-item-row .item-qty { color: #999; margin: 0 8px; }
.order-item-row .item-price { color: #f56c6c; font-weight: 500; }

/* 订单详情对话框 */
.order-detail { padding: 10px 0; }
.order-item-detail {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid #eee;
}
.order-item-detail:last-child { border-bottom: none; }
.order-item-detail span { color: #333; }
.empty-items { color: #999; text-align: center; padding: 20px; }

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
