<template>
  <div class="home">
    <AppHeader />

    <!-- 轮播图 -->
    <div class="banner">
      <el-carousel height="400px" :interval="5000">
        <el-carousel-item>
          <div class="banner-item banner-1">
            <h2>农技综合服务平台</h2>
            <p>汇聚农业专家，提供种植养殖技术指导与咨询服务</p>
          </div>
        </el-carousel-item>
        <el-carousel-item>
          <div class="banner-item banner-2">
            <h2>优质农产品交易</h2>
            <p>安全可靠的农产品买卖平台，打通供需两端</p>
          </div>
        </el-carousel-item>
        <el-carousel-item>
          <div class="banner-item banner-3">
            <h2>农业知识共享</h2>
            <p>涵盖种植、养殖、农机等多领域专业知识</p>
          </div>
        </el-carousel-item>
        <el-carousel-item>
          <div class="banner-item banner-4">
            <h2>农村事务处理</h2>
            <p>基础设施报修、环境问题上报，及时响应解决</p>
          </div>
        </el-carousel-item>
      </el-carousel>
    </div>

    <!-- 内容区域 -->
    <div class="container main-content">
      <!-- 全局加载提示 -->
      <div v-if="loading" class="loading-bar">
        <el-icon class="is-loading"><Loading /></el-icon>
        <span>正在加载推荐内容...</span>
      </div>
      
      <!-- 最新商品 -->
      <section class="section">
        <div class="section-header">
          <h3>最新商品</h3>
          <router-link to="/products" class="more">查看更多</router-link>
        </div>
        <el-empty v-if="!loading && products.length === 0" description="暂无商品推荐" :image-size="80" />
        <el-row :gutter="20" v-else>
          <el-col :span="6" v-for="item in products" :key="item.id">
            <router-link :to="`/product/${item.id}`" class="product-card">
              <el-image :src="item.image_url" fit="cover">
                <template #error><div class="image-placeholder">暂无图片</div></template>
              </el-image>
              <div class="product-info">
                <h4>{{ item.title }}</h4>
                <p class="price">¥{{ item.price }}</p>
              </div>
            </router-link>
          </el-col>
        </el-row>
      </section>

      <!-- 热门动态 -->
      <section class="section">
        <div class="section-header">
          <h3>热门动态</h3>
          <router-link to="/posts" class="more">查看更多</router-link>
        </div>
        <el-empty v-if="!loading && posts.length === 0" description="暂无最新动态" :image-size="80" />
        <div class="post-list" v-else>
          <div v-for="item in posts" :key="item.id" class="post-item">
            <router-link :to="`/post/${item.id}`">
              <h4>{{ item.title }}</h4>
              <p>{{ item.content }}</p>
              <div class="post-meta">
                <span><el-icon><View /></el-icon> {{ item.views }}</span>
                <span><el-icon><Star /></el-icon> {{ item.likes }}</span>
                <span class="type-tag">{{ getTypeName(item.type) }}</span>
              </div>
            </router-link>
          </div>
        </div>
      </section>

      <!-- 农业知识 -->
      <section class="section">
        <div class="section-header">
          <h3>农业知识</h3>
          <router-link to="/knowledge" class="more">查看更多</router-link>
        </div>
        <el-empty v-if="!loading && knowledges.length === 0" description="暂无农业知识" :image-size="80" />
        <div class="knowledge-list" v-else>
          <div v-for="item in knowledges" :key="item.id" class="knowledge-item">
            <router-link :to="`/knowledge/${item.id}`">
              <div class="knowledge-image" v-if="item.image_url">
                <el-image :src="item.image_url" fit="cover" />
              </div>
              <div class="knowledge-content">
                <h4>{{ item.title }}</h4>
                <p class="knowledge-desc">{{ item.content?.slice(0, 80) }}{{ item.content?.length > 80 ? '...' : '' }}</p>
                <div class="knowledge-meta">
                  <span class="author">{{ item.user?.nickname || item.user?.username || '匿名' }}</span>
                  <span>{{ formatDate(item.created_at) }}</span>
                </div>
              </div>
            </router-link>
          </div>
        </div>
      </section>

      <!-- 农村信息专区 -->
      <section class="section">
        <div class="section-header">
          <h3>🌾 农村信息</h3>
          <router-link to="/rural-info" class="more">进入信息中心</router-link>
        </div>
        <el-row :gutter="16">
          <el-col :span="8">
            <router-link to="/rural-info" class="rural-card">
              <div class="rural-card-icon">📋</div>
              <h4>农村信息介绍</h4>
              <p>了解乡村风貌、农业资源与发展动态</p>
            </router-link>
          </el-col>
          <el-col :span="8">
            <router-link to="/rural-info" class="rural-card">
              <div class="rural-card-icon">📜</div>
              <h4>政策公告</h4>
              <p>最新惠农政策、法规解读与补贴指南</p>
            </router-link>
          </el-col>
          <el-col :span="8">
            <router-link to="/rural-info" class="rural-card">
              <div class="rural-card-icon">🔧</div>
              <h4>农村事务</h4>
              <p>基础设施报修、环境问题上报与处理跟踪</p>
            </router-link>
          </el-col>
        </el-row>
      </section>

      <!-- 推荐专家 -->
      <section class="section">
        <div class="section-header">
          <h3>推荐专家</h3>
          <router-link to="/experts" class="more">查看更多</router-link>
        </div>
        <el-row :gutter="20">
          <el-col :span="8" v-for="item in experts" :key="item.id">
            <router-link :to="`/expert/${item.id}`" class="expert-card">
              <el-avatar :src="item.user?.avatar_url" :size="64">{{ item.real_name?.[0] }}</el-avatar>
              <div class="expert-info">
                <h4>{{ item.real_name }}</h4>
                <p>{{ item.profession }}</p>
                <p class="company">{{ item.company }}</p>
              </div>
            </router-link>
          </el-col>
        </el-row>
      </section>
    </div>

    <AppFooter />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { View, Star, Loading } from '@element-plus/icons-vue'
import api from '@/utils/api'

const products = ref([])
const posts = ref([])
const experts = ref([])
const knowledges = ref([])

const getTypeName = (type) => {
  const map = { normal: '普通', question: '提问', share: '分享' }
  return map[type] || '普通'
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

const loading = ref(true)
const loadErrors = ref([])

onMounted(async () => {
  // 使用 Promise.allSettled 替代 Promise.all：
  // 即使某个 API 失败，其他仍正常显示，不会因为一个挂导致全部空白
  // 热门推荐算法综合浏览量、评论数与发布时间
  const results = await Promise.allSettled([
    api.getOnSaleProducts(),
    api.getHotPosts({ page: 1, page_size: 5 }),
    api.getExperts({ page: 1, page_size: 3 }),
    api.getHotKnowledge({ page: 1, page_size: 4 })
  ])

  // 处理热门商品（索引 0）
  if (results[0].status === 'fulfilled') {
    const res = results[0].value
    products.value = (res?.list || []).map(item => ({
      ...item,
      image_url: item.image_url?.startsWith('/') ? item.image_url : '/' + item.image_url
    }))
  } else {
    console.warn('[首页] 热门商品加载失败:', results[0].reason?.message || results[0].reason)
  }

  // 处理最新动态（索引 1）
  if (results[1].status === 'fulfilled') {
    posts.value = results[1].value?.list || []
  } else {
    console.warn('[首页] 最新动态加载失败:', results[1].reason?.message || results[1].reason)
  }

  // 处理推荐专家（索引 2）
  if (results[2].status === 'fulfilled') {
    const res = results[2].value
    experts.value = (res?.list || []).map(item => ({
      ...item,
      user: item.user ? {
        ...item.user,
        avatar_url: item.user.avatar_url?.startsWith('/') ? item.user.avatar_url : '/' + item.user.avatar_url
      } : null
    }))
  } else {
    console.warn('[首页] 推荐专家加载失败:', results[2].reason?.message || results[2].reason)
  }

  // 处理热门知识（索引 3）
  if (results[3].status === 'fulfilled') {
    knowledges.value = results[3].value?.list || []
  } else {
    console.warn('[首页] 热门知识加载失败:', results[3].reason?.message || results[3].reason)
  }

  loading.value = false
})
</script>

<style scoped>
.home { min-height: 100vh; }

.container { max-width: 1200px; margin: 0 auto; padding: 0 20px; }

.header .container {
  display: flex;
  align-items: center;
  height: 60px;
}

.logo { font-size: 20px; font-weight: bold; color: #409eff; }

.nav { flex: 1; margin-left: 40px; }
.nav a {
  margin: 0 15px;
  color: #666;
  text-decoration: none;
  font-size: 15px;
}
.nav a:hover, .nav a.router-link-active { color: #409eff; }

.user-area { display: flex; gap: 10px; align-items: center; }
.user-info { display: flex; align-items: center; gap: 8px; cursor: pointer; }

.banner { background: #f0f2f5; }
.banner-item {
  height: 400px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  color: #fff;
  text-align: center;
}
.banner-1 { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
.banner-2 { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); }
.banner-3 { background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%); }
.banner-4 { background: linear-gradient(135deg, #fa709a 0%, #fee140 100%); }
.banner-item h2 { font-size: 36px; margin-bottom: 15px; }
.banner-item p { font-size: 18px; }

.main-content { padding: 40px 20px; }

.loading-bar {
  display: flex; align-items: center; justify-content: center; gap: 10px;
  padding: 20px; color: #909399; font-size: 15px;
}

.section { margin-bottom: 50px; }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.section-header h3 { font-size: 20px; color: #333; }
.more { color: #409eff; text-decoration: none; }

.product-card {
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  text-decoration: none;
  display: block;
  margin-bottom: 20px;
}
.product-card .el-image { width: 100%; height: 180px; }
.image-placeholder {
  width: 100%; height: 180px;
  display: flex; align-items: center; justify-content: center;
  background: #f5f7fa; color: #c0c4cc; font-size: 14px;
}
.product-info { padding: 15px; }
.product-info h4 { color: #333; font-size: 16px; margin-bottom: 8px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.price { color: #f56c6c; font-size: 18px; font-weight: bold; }

.post-list { background: #fff; border-radius: 8px; overflow: hidden; }
.post-item { padding: 20px; border-bottom: 1px solid #eee; }
.post-item:last-child { border-bottom: none; }
.post-item a { text-decoration: none; color: inherit; }
.post-item h4 { color: #333; margin-bottom: 8px; }
.post-item p { color: #666; font-size: 14px; overflow: hidden; text-overflow: ellipsis; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; margin-bottom: 10px; }
.post-meta { display: flex; gap: 20px; color: #999; font-size: 13px; }
.post-meta span { display: flex; align-items: center; gap: 4px; }
.type-tag { background: #ecf5ff; color: #409eff; padding: 2px 8px; border-radius: 4px; }

.rural-card {
  background: #fff; border-radius: 10px; padding: 24px 20px; text-align: center;
  text-decoration: none; display: block; box-shadow: 0 2px 10px rgba(0,0,0,0.05);
  transition: transform 0.2s, box-shadow 0.2s;
}
.rural-card:hover { transform: translateY(-4px); box-shadow: 0 4px 16px rgba(46,125,50,0.15); }
.rural-card-icon { font-size: 40px; margin-bottom: 10px; }
.rural-card h4 { font-size: 17px; color: #333; margin-bottom: 6px; }
.rural-card p { font-size: 13px; color: #888; line-height: 1.5; }

.expert-card {
  background: #fff;
  border-radius: 8px;
  padding: 30px;
  display: flex;
  align-items: center;
  gap: 20px;
  text-decoration: none;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  margin-bottom: 20px;
}
.expert-info h4 { color: #333; margin-bottom: 5px; }
.expert-info p { color: #666; font-size: 14px; }
.expert-info .company { color: #999; font-size: 13px; }

.knowledge-list { display: grid; grid-template-columns: repeat(2, 1fr); gap: 20px; }
.knowledge-item {
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  transition: transform 0.2s, box-shadow 0.2s;
}
.knowledge-item:hover { transform: translateY(-4px); box-shadow: 0 4px 16px rgba(0,0,0,0.15); }
.knowledge-item a { text-decoration: none; display: flex; flex-direction: column; height: 100%; }
.knowledge-image { width: 100%; height: 160px; overflow: hidden; }
.knowledge-image .el-image { width: 100%; height: 100%; }
.knowledge-content { padding: 15px; flex: 1; display: flex; flex-direction: column; }
.knowledge-content h4 { color: #333; margin-bottom: 8px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.knowledge-desc { color: #666; font-size: 13px; line-height: 1.6; flex: 1; }
.knowledge-meta { display: flex; justify-content: space-between; margin-top: 10px; color: #999; font-size: 12px; }
.knowledge-meta .author { color: #409eff; }

.footer {
  background: #333;
  color: #fff;
  text-align: center;
  padding: 30px;
}
</style>
