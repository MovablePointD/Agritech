<template>
  <div class="product-list">
    <AppHeader />

    <div class="container main">
      <div class="filter-bar">
        <el-select v-model="type" placeholder="选择类型" clearable @change="fetchData">
          <el-option label="全部" value="" />
          <el-option label="种子" value="seed" />
          <el-option label="肥料" value="fertilizer" />
          <el-option label="农药" value="pesticide" />
          <el-option label="农具" value="tool" />
        </el-select>
        <el-input v-model="keyword" placeholder="搜索商品" clearable @keyup.enter="fetchData" style="width: 200px;">
          <template #append><el-button :icon="Search" @click="fetchData" /></template>
        </el-input>
      </div>

      <el-row :gutter="20" v-loading="loading">
        <el-col :span="6" v-for="item in list" :key="item.id">
          <router-link :to="`/product/${item.id}`" class="product-card" :class="{ 'pending-card': item.status !== 1 }">
            <el-image :src="item.image_url" fit="cover" />
            <div class="info">
              <div class="title-row">
                <h4>{{ item.title }}</h4>
                <el-tag v-if="item.status === -1" size="small" type="warning">审核中</el-tag>
                <el-tag v-else-if="item.status === 0" size="small" type="danger">已下架</el-tag>
              </div>
              <p class="desc">{{ item.content }}</p>
              <div class="bottom">
                <span class="price">¥{{ item.price }}</span>
                <span class="stock" :class="getStockClass(item.stock)">库存: {{ item.stock || 0 }}</span>
              </div>
            </div>
          </router-link>
        </el-col>
      </el-row>

      <el-empty v-if="!loading && list.length === 0" description="暂无商品" />

      <div class="pagination" v-if="total > 0">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="prev, pager, next"
          @current-change="fetchData"
        />
      </div>
    </div>

    <AppFooter />
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import api from '@/utils/api'

const list = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(12)
const total = ref(0)
const type = ref('')
const keyword = ref('')

// 获取库存样式类
const getStockClass = (stock) => {
  if (!stock || stock <= 0) return 'danger'
  if (stock < 20) return 'warning'
  return ''
}

const fetchData = async () => {
  loading.value = true
  try {
    const res = await api.getProducts({ page: page.value, page_size: pageSize.value, type: type.value })
    list.value = (res.list || []).map(item => ({
      ...item,
      image_url: item.image_url?.startsWith('/') ? item.image_url : '/' + item.image_url
    }))
    total.value = res.total || 0
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>

<style scoped>
.product-list { min-height: 100vh; background: #f5f5f5; }
.container { max-width: 1200px; margin: 0 auto; padding: 0 20px; }
.main { padding: 20px; }
.filter-bar { display: flex; gap: 15px; margin-bottom: 20px; }

.product-card {
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  text-decoration: none;
  display: flex;
  flex-direction: column;
  margin-bottom: 20px;
  transition: transform 0.2s;
  min-height: 320px;
}
.product-card.pending-card { opacity: 0.82; border: 2px solid #e6a23c; }
.product-card:hover { transform: translateY(-5px); }
.product-card .el-image { width: 100%; height: 180px; flex-shrink: 0; }
.info { padding: 15px; display: flex; flex-direction: column; flex: 1; }
.title-row { display: flex; align-items: center; gap: 6px; margin-bottom: 8px; }
.info h4 { color: #333; margin: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; flex: 1; }
.info .desc { color: #999; font-size: 13px; overflow: hidden; text-overflow: ellipsis; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; min-height: 36px; margin-bottom: 8px; flex: 1; }
.bottom { display: flex; justify-content: space-between; align-items: center; margin-top: auto; }
.price { color: #f56c6c; font-size: 18px; font-weight: bold; }
.stock { font-size: 12px; color: #67c23a; }
.stock.warning { color: #e6a23c; }
.stock.danger { color: #f56c6c; }

.pagination { display: flex; justify-content: center; margin-top: 20px; }
</style>
