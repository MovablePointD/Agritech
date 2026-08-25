<template>
  <div class="cart-page">
    <AppHeader />

    <div class="container main">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>我的购物车</span>
            <div class="header-actions">
              <span class="cart-summary">共 {{ list.length }} 件商品，已选 {{ selectedItems.length }} 件</span>
              <el-button v-if="list.length > 0" type="danger" size="small" @click="handleClearCart">清空购物车</el-button>
            </div>
          </div>
        </template>

        <!-- 购物车列表 -->
        <el-table v-if="!loading && list.length > 0" :data="list" v-loading="loading" @selection-change="handleSelectionChange">
          <el-table-column type="selection" width="55" />
          <el-table-column label="商品" min-width="300">
            <template #default="{ row }">
              <div class="product-info">
                <router-link :to="`/product/${row.product_id}`" class="product-link">
                  <el-image :src="row.product?.image_url" fit="cover" class="product-image" />
                  <div class="product-detail">
                    <div class="product-title">{{ row.product?.title || '商品' }}</div>
                    <div class="product-desc">{{ row.product?.content?.substring(0, 50) || '' }}...</div>
                  </div>
                </router-link>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="price" label="单价" width="120">
            <template #default="{ row }">¥{{ row.price?.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column label="数量" width="150">
            <template #default="{ row }">
              <el-input-number v-model="row.quantity" :min="1" :max="99" size="small" @change="handleQuantityChange(row)" />
            </template>
          </el-table-column>
          <el-table-column prop="total_price" label="小计" width="120">
            <template #default="{ row }">
              <span class="subtotal">¥{{ row.total_price?.toFixed(2) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="160">
            <template #default="{ row }">
              <el-button size="small" type="primary" @click="handleBuyNow(row)">购买</el-button>
              <el-button size="small" type="danger" @click="handleRemove(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>

        <!-- 空状态 -->
        <div v-if="!loading && list.length === 0" class="empty-state">
          <div class="empty-icon">
            <svg class="empty-svg" viewBox="0 0 120 120" fill="none" xmlns="http://www.w3.org/2000/svg">
              <circle cx="60" cy="60" r="50" fill="#f5f5f5" stroke="#e0e0e0" stroke-width="2"/>
              <path d="M35 40H85V85H35V40Z" fill="#fff" stroke="#d0d0d0" stroke-width="2"/>
              <path d="M35 40L50 25H70L85 40" stroke="#d0d0d0" stroke-width="2" stroke-linecap="round"/>
              <circle cx="45" cy="55" r="5" fill="#d0d0d0"/>
              <circle cx="75" cy="55" r="5" fill="#d0d0d0"/>
              <path d="M50 70H70" stroke="#d0d0d0" stroke-width="2" stroke-linecap="round"/>
            </svg>
          </div>
          <p class="empty-text">购物车空空如也</p>
          <el-button type="primary" @click="$router.push('/products')">去购物</el-button>
        </div>

        <!-- 底部结算栏 -->
        <div v-if="!loading && list.length > 0" class="settlement-bar">
          <div class="settlement-left">
            <el-checkbox v-model="selectAll" @change="handleSelectAll">全选</el-checkbox>
            <el-button type="text" @click="handleBatchDelete" :disabled="selectedItems.length === 0">批量删除</el-button>
          </div>
          <div class="settlement-right">
            <div class="total-info">
              <span class="total-label">合计：</span>
              <span class="total-price">¥{{ totalPrice.toFixed(2) }}</span>
            </div>
            <el-button type="primary" size="large" @click="handleSettlement" :disabled="selectedItems.length === 0">结算</el-button>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 购买对话框 -->
    <el-dialog v-model="showBuyDialog" title="确认订单" width="600px">
      <el-form :model="buyForm" label-width="80px">
        <el-form-item label="商品">
          <div v-for="item in selectedBuyItems" :key="item.id" class="buy-item">
            <span>{{ item.product?.title }}</span>
            <span>× {{ item.quantity }}</span>
            <span class="price">¥{{ item.total_price?.toFixed(2) }}</span>
          </div>
        </el-form-item>
        <el-form-item label="收货地址">
          <el-select v-model="buyForm.address_id" placeholder="请选择地址" style="width: 100%">
            <el-option v-for="addr in addresses" :key="addr.id" :label="`${addr.receiver} ${addr.province}${addr.city}${addr.district}${addr.detail}`" :value="addr.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="buyForm.remark" type="textarea" :rows="2" placeholder="选填" />
        </el-form-item>
        <el-form-item label="订单总价">
          <span class="total-price">¥{{ settlementTotal.toFixed(2) }}</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showBuyDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitOrder" :loading="submitting">提交订单</el-button>
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
const selectedItems = ref([])
const selectAll = ref(false)
const cartCount = ref(0)
const showBuyDialog = ref(false)
const submitting = ref(false)
const selectedBuyItems = ref([])
const addresses = ref([])

const buyForm = reactive({
  address_id: null,
  remark: ''
})

// 计算总价
const totalPrice = computed(() => {
  return selectedItems.value.reduce((sum, item) => sum + (item.total_price || 0), 0)
})

const settlementTotal = computed(() => {
  return selectedBuyItems.value.reduce((sum, item) => sum + (item.total_price || 0), 0)
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

// 加载购物车数据
const fetchData = async () => {
  loading.value = true
  try {
    const res = await api.getCart()
    list.value = (res?.list || res || []).map(item => ({
      ...item,
      product: item.product ? {
        ...item.product,
        image_url: item.product.image_url?.startsWith('/') ? item.product.image_url : '/' + item.product.image_url
      } : null
    }))
  } catch (error) {
    console.error('加载购物车失败:', error)
  } finally {
    loading.value = false
  }
}

// 加载购物车数量
const loadCartCount = async () => {
  try {
    const res = await api.getCartCount()
    cartCount.value = res?.count || 0
  } catch (error) {
    cartCount.value = 0
  }
}

// 加载收货地址
const loadAddresses = async () => {
  try {
    const res = await api.getAddresses()
    addresses.value = res?.list || res || []
  } catch (error) {
    addresses.value = []
  }
}

// 数量变化
const handleQuantityChange = async (row) => {
  try {
    const updated = await api.updateCartItem(row.id, { quantity: row.quantity })
    if (updated && updated.total_price !== undefined) {
      row.total_price = updated.total_price
    }
    await loadCartCount()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '更新数量失败'
    ElMessage.error(msg)
    fetchData()
  }
}

// 删除单个商品
const handleRemove = async (id) => {
  try {
    await ElMessageBox.confirm('确定要从购物车移除该商品吗？', '提示', { type: 'warning' })
    await api.removeFromCart(id)
    ElMessage.success('已移除')
    fetchData()
    loadCartCount()
  } catch (error) {
    if (error !== 'cancel') {
      const msg = error?.response?.data?.error || error?.message || '移除失败'
      ElMessage.error(msg)
    }
  }
}

// 清空购物车
const handleClearCart = async () => {
  try {
    await ElMessageBox.confirm('确定要清空购物车吗？', '提示', { type: 'warning' })
    await api.clearCart()
    ElMessage.success('已清空')
    fetchData()
    loadCartCount()
  } catch (error) {
    if (error !== 'cancel') {
      const msg = error?.response?.data?.error || error?.message || '清空失败'
      ElMessage.error(msg)
    }
  }
}

// 选择变化
const handleSelectionChange = (selection) => {
  selectedItems.value = selection
  selectAll.value = list.value.length > 0 && selection.length === list.value.length
}

// 全选/取消全选
const handleSelectAll = (val) => {
  if (val) {
    selectedItems.value = [...list.value]
  } else {
    selectedItems.value = []
  }
}

// 批量删除
const handleBatchDelete = async () => {
  if (selectedItems.value.length === 0) return
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedItems.value.length} 件商品吗？`, '提示', { type: 'warning' })
    for (const item of selectedItems.value) {
      await api.removeFromCart(item.id)
    }
    ElMessage.success('已删除')
    fetchData()
    loadCartCount()
    selectedItems.value = []
  } catch (error) {
    if (error !== 'cancel') {
      const msg = error?.response?.data?.error || error?.message || '删除失败'
      ElMessage.error(msg)
    }
  }
}

// 立即购买单个商品
const handleBuyNow = async (item) => {
  selectedBuyItems.value = [item]
  await loadAddresses()
  showBuyDialog.value = true
}

// 结算
const handleSettlement = async () => {
  if (selectedItems.value.length === 0) {
    ElMessage.warning('请选择要结算的商品')
    return
  }
  selectedBuyItems.value = [...selectedItems.value]
  await loadAddresses()
  showBuyDialog.value = true
}

// 提交订单
const handleSubmitOrder = async () => {
  if (!buyForm.address_id) {
    ElMessage.warning('请选择收货地址')
    return
  }

  submitting.value = true
  try {
    const itemIds = selectedBuyItems.value.map(item => item.id)

    // 使用购物车结算API（每个商品单独创建订单）
    const res = await api.checkoutCart({
      item_ids: itemIds,
      address_id: buyForm.address_id,
      remark: buyForm.remark
    })

    const orderCount = res?.orders?.length || 1
    ElMessage.success(`成功创建 ${orderCount} 笔订单`)
    showBuyDialog.value = false
    fetchData()
    loadCartCount()
    router.push('/orders')
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '创建订单失败'
    ElMessage.error(msg)
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  initUserInfo()
  loadUserDetail()
  fetchData()
  loadCartCount()
})
</script>

<style scoped>
.cart-page { min-height: 100vh; background: #f5f5f5; }
.container { max-width: 1200px; margin: 0 auto; padding: 0 20px; }
.main { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.header-actions { display: flex; align-items: center; gap: 15px; }
.cart-summary { color: #666; font-size: 14px; }

/* 商品信息 */
.product-info { display: flex; align-items: center; }
.product-link { display: flex; align-items: center; text-decoration: none; color: inherit; }
.product-image { width: 80px; height: 80px; border-radius: 4px; margin-right: 15px; }
.product-detail { flex: 1; }
.product-title { color: #333; font-weight: 500; margin-bottom: 5px; }
.product-desc { color: #999; font-size: 12px; }

.subtotal { color: #f56c6c; font-weight: bold; font-size: 16px; }

/* 结算栏 */
.settlement-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 0;
  margin-top: 20px;
  border-top: 1px solid #eee;
}
.settlement-left { display: flex; align-items: center; gap: 20px; }
.settlement-right { display: flex; align-items: center; gap: 20px; }
.total-info { display: flex; align-items: baseline; }
.total-label { color: #666; font-size: 14px; }
.total-price { color: #f56c6c; font-size: 24px; font-weight: bold; }

/* 购买对话框 */
.buy-item { display: flex; justify-content: space-between; padding: 5px 0; }
.buy-item .price { color: #f56c6c; font-weight: bold; }

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
.empty-svg { width: 120px; height: 120px; }
.empty-text { color: #909399; font-size: 14px; margin-bottom: 20px; }
</style>
