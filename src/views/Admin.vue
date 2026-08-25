<template>
  <div class="admin-container">
    <header class="header">
      <div class="container">
        <router-link to="/" class="logo">农技知识分享平台 - 系统管理端</router-link>
        <nav class="nav">
          <router-link to="/">首页</router-link>
          <router-link to="/user">个人中心</router-link>
          <router-link v-if="currentUser.role === 'admin'" to="/reviewer">审核中心</router-link>
        </nav>
        <div class="user-area">
          <span class="user-info">
            <el-avatar :src="currentUser.avatar_url" :size="32">{{ currentUser.username?.[0] }}</el-avatar>
            <span>{{ currentUser.nickname || currentUser.username }}</span>
            <el-tag type="danger" size="small" style="margin-left: 4px;">系统管理员</el-tag>
          </span>
          <el-button size="small" @click="handleLogout">退出</el-button>
        </div>
      </div>
    </header>

    <div class="container main">
      <el-tabs v-model="activeTab" type="border-card">
        <!-- 商品管理 -->
        <el-tab-pane label="商品管理" name="products">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-card>
                <template #header>
                  <span class="card-title">待审核 ({{ pendingProducts.length }})</span>
                  <el-button size="small" @click="loadPendingProducts">刷新</el-button>
                </template>
                <el-table :data="pendingProducts" v-loading="loadingProducts" height="350" empty-text="暂无待审核商品">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="title" label="商品名称" min-width="120" show-overflow-tooltip />
                  <el-table-column prop="price" label="价格" width="80">
                    <template #default="{ row }">¥{{ row.price }}</template>
                  </el-table-column>
                  <el-table-column label="发布者" width="100">
                    <template #default="{ row }">
                      <span>{{ row.publisher_user?.nickname || row.publisher_user?.username || '未知' }}</span>
                    </template>
                  </el-table-column>
                  <el-table-column label="操作" width="150" fixed="right">
                    <template #default="{ row }">
                      <el-button size="small" @click="showProductDetail(row)">详情</el-button>
                      <el-button size="small" type="success" @click="passProduct(row.id)">通过</el-button>
                      <el-button size="small" type="danger" @click="rejectProduct(row.id)">退回</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
            <el-col :span="12">
              <el-card>
                <template #header>
                  <span class="card-title">已退回 ({{ rejectedProducts.length }})</span>
                  <el-button size="small" @click="loadRejectedProducts">刷新</el-button>
                </template>
                <el-table :data="rejectedProducts" v-loading="loadingProducts" height="350" empty-text="暂无已退回商品">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="title" label="商品名称" min-width="120" show-overflow-tooltip />
                  <el-table-column prop="price" label="价格" width="80">
                    <template #default="{ row }">¥{{ row.price }}</template>
                  </el-table-column>
                  <el-table-column label="发布者" width="100">
                    <template #default="{ row }">
                      <span>{{ row.publisher_user?.nickname || row.publisher_user?.username || '未知' }}</span>
                    </template>
                  </el-table-column>
                  <el-table-column label="操作" width="180" fixed="right">
                    <template #default="{ row }">
                      <el-button size="small" @click="showProductDetail(row)">详情</el-button>
                      <el-button size="small" type="primary" @click="editProduct(row)">编辑</el-button>
                      <el-button size="small" type="success" @click="reOnShelfProduct(row.id)">重新提交</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
          </el-row>
          <el-card style="margin-top: 20px;">
            <template #header>
              <span class="card-title">已上架商品 ({{ onShelfProducts.length }})</span>
              <el-button size="small" @click="loadOnShelfProducts">刷新</el-button>
            </template>
            <el-table :data="onShelfProducts" v-loading="loadingProducts" empty-text="暂无上架商品">
              <el-table-column prop="id" label="ID" width="60" />
              <el-table-column prop="title" label="商品名称" min-width="150" show-overflow-tooltip />
              <el-table-column prop="price" label="价格" width="100">
                <template #default="{ row }">¥{{ row.price }}</template>
              </el-table-column>
              <el-table-column prop="type" label="类型" width="100">
                <template #default="{ row }">
                  <el-tag size="small">{{ getProductTypeText(row.type) }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="发布者" width="100">
                <template #default="{ row }">
                  <span>{{ row.publisher_user?.nickname || row.publisher_user?.username || '未知' }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="created_at" label="创建时间" width="160" :formatter="formatTime" />
              <el-table-column label="操作" width="180" fixed="right">
                <template #default="{ row }">
                  <el-button size="small" @click="showProductDetail(row)">详情</el-button>
                  <el-button size="small" type="primary" @click="editProduct(row)">编辑</el-button>
                  <el-button size="small" type="danger" @click="offShelfProduct(row.id)">退回</el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-tab-pane>

        <!-- 动态管理 -->
        <el-tab-pane label="动态管理" name="posts">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-card>
                <template #header>
                  <span class="card-title">待审核 ({{ pendingPosts.length }})</span>
                  <el-button size="small" @click="loadPendingPosts">刷新</el-button>
                </template>
                <el-table :data="pendingPosts" v-loading="loadingPosts" height="350" empty-text="暂无待审核动态">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="title" label="标题" min-width="120" show-overflow-tooltip />
                  <el-table-column label="发布者" width="100">
                    <template #default="{ row }">
                      <span>{{ row.user?.nickname || row.user?.username || '未知' }}</span>
                    </template>
                  </el-table-column>
                  <el-table-column label="操作" width="150" fixed="right">
                    <template #default="{ row }">
                      <el-button size="small" @click="showPostDetail(row)">详情</el-button>
                      <el-button size="small" type="success" @click="passPost(row.id)">通过</el-button>
                      <el-button size="small" type="danger" @click="rejectPost(row.id)">退回</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
            <el-col :span="12">
              <el-card>
                <template #header>
                  <span class="card-title">已退回 ({{ rejectedPosts.length }})</span>
                  <el-button size="small" @click="loadRejectedPosts">刷新</el-button>
                </template>
                <el-table :data="rejectedPosts" v-loading="loadingPosts" height="350" empty-text="暂无已退回动态">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="title" label="标题" min-width="120" show-overflow-tooltip />
                  <el-table-column label="发布者" width="100">
                    <template #default="{ row }">
                      <span>{{ row.user?.nickname || row.user?.username || '未知' }}</span>
                    </template>
                  </el-table-column>
                  <el-table-column label="操作" width="180" fixed="right">
                    <template #default="{ row }">
                      <el-button size="small" @click="showPostDetail(row)">详情</el-button>
                      <el-button size="small" type="primary" @click="editPost(row)">编辑</el-button>
                      <el-button size="small" type="success" @click="restorePost(row.id)">恢复</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
          </el-row>
          <el-card style="margin-top: 20px;">
            <template #header>
              <span class="card-title">已发布动态 ({{ publishedPosts.length }})</span>
              <el-button size="small" @click="loadPublishedPosts">刷新</el-button>
            </template>
            <el-table :data="publishedPosts" v-loading="loadingPosts" empty-text="暂无已发布动态">
              <el-table-column prop="id" label="ID" width="60" />
              <el-table-column prop="title" label="标题" min-width="150" show-overflow-tooltip />
              <el-table-column prop="type" label="类型" width="100">
                <template #default="{ row }">
                  <el-tag size="small">{{ getPostTypeText(row.type) }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="发布者" width="100">
                <template #default="{ row }">
                  <span>{{ row.user?.nickname || row.user?.username || '未知' }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="likes" label="点赞" width="80" />
              <el-table-column prop="views" label="浏览" width="80" />
              <el-table-column prop="created_at" label="创建时间" width="160" :formatter="formatTime" />
              <el-table-column label="操作" width="150" fixed="right">
                <template #default="{ row }">
                  <el-button size="small" @click="showPostDetail(row)">详情</el-button>
                  <el-button size="small" type="primary" @click="editPost(row)">编辑</el-button>
                  <el-button size="small" type="danger" @click="rejectPublishedPost(row.id)">退回</el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
          <el-card style="margin-top: 20px;">
            <template #header>
              <span class="card-title">已删除动态 ({{ deletedPosts.length }})</span>
              <el-button size="small" @click="loadDeletedPosts">刷新</el-button>
            </template>
            <el-table :data="deletedPosts" v-loading="loadingPosts" empty-text="暂无已删除动态">
              <el-table-column prop="id" label="ID" width="60" />
              <el-table-column prop="title" label="标题" min-width="150" show-overflow-tooltip />
              <el-table-column prop="type" label="类型" width="100">
                <template #default="{ row }">
                  <el-tag size="small">{{ getPostTypeText(row.type) }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="发布者" width="100">
                <template #default="{ row }">
                  <span>{{ row.user?.nickname || row.user?.username || '未知' }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="created_at" label="创建时间" width="160" :formatter="formatTime" />
              <el-table-column label="操作" width="150" fixed="right">
                <template #default="{ row }">
                  <el-button size="small" @click="showPostDetail(row)">详情</el-button>
                  <el-button size="small" type="primary" @click="editPost(row)">编辑</el-button>
                  <el-button size="small" type="success" @click="restorePost(row.id)">恢复</el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-tab-pane>

        <!-- 知识管理 -->
        <el-tab-pane label="知识管理" name="knowledge">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-card>
                <template #header>
                  <span class="card-title">待审核 ({{ pendingKnowledges.length }})</span>
                  <el-button size="small" @click="loadPendingKnowledges">刷新</el-button>
                </template>
                <el-table :data="pendingKnowledges" v-loading="loadingKnowledge" height="350" empty-text="暂无待审核知识">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="title" label="标题" min-width="150" show-overflow-tooltip />
                  <el-table-column label="发布者" width="100">
                    <template #default="{ row }">
                      <span>{{ row.user?.nickname || row.user?.username || '未知' }}</span>
                    </template>
                  </el-table-column>
                  <el-table-column label="操作" width="150" fixed="right">
                    <template #default="{ row }">
                      <el-button size="small" @click="showKnowledgeDetail(row)">详情</el-button>
                      <el-button size="small" type="success" @click="passKnowledge(row.id)">通过</el-button>
                      <el-button size="small" type="danger" @click="rejectKnowledge(row.id)">退回</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
            <el-col :span="12">
              <el-card>
                <template #header>
                  <span class="card-title">已退回 ({{ rejectedKnowledges.length }})</span>
                  <el-button size="small" @click="loadRejectedKnowledges">刷新</el-button>
                </template>
                <el-table :data="rejectedKnowledges" v-loading="loadingKnowledge" height="350" empty-text="暂无已退回知识">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="title" label="标题" min-width="150" show-overflow-tooltip />
                  <el-table-column label="发布者" width="100">
                    <template #default="{ row }">
                      <span>{{ row.user?.nickname || row.user?.username || '未知' }}</span>
                    </template>
                  </el-table-column>
                  <el-table-column label="操作" width="180" fixed="right">
                    <template #default="{ row }">
                      <el-button size="small" @click="showKnowledgeDetail(row)">详情</el-button>
                      <el-button size="small" type="primary" @click="editKnowledge(row)">编辑</el-button>
                      <el-button size="small" type="success" @click="restoreKnowledge(row.id)">恢复</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
          </el-row>
          <el-card style="margin-top: 20px;">
            <template #header>
              <span class="card-title">已发布知识 ({{ publishedKnowledges.length }})</span>
              <el-button size="small" @click="loadPublishedKnowledges">刷新</el-button>
            </template>
            <el-table :data="publishedKnowledges" v-loading="loadingKnowledge" empty-text="暂无已发布知识">
              <el-table-column prop="id" label="ID" width="60" />
              <el-table-column prop="title" label="标题" min-width="150" show-overflow-tooltip />
              <el-table-column label="发布者" width="100">
                <template #default="{ row }">
                  <span>{{ row.user?.nickname || row.user?.username || '未知' }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="created_at" label="创建时间" width="160" :formatter="formatTime" />
              <el-table-column label="操作" width="180" fixed="right">
                <template #default="{ row }">
                  <el-button size="small" @click="showKnowledgeDetail(row)">详情</el-button>
                  <el-button size="small" type="primary" @click="editKnowledge(row)">编辑</el-button>
                  <el-button size="small" type="danger" @click="rejectPublishedKnowledge(row.id)">退回</el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
          <el-card style="margin-top: 20px;">
            <template #header>
              <span class="card-title">已删除知识 ({{ deletedKnowledges.length }})</span>
              <el-button size="small" @click="loadDeletedKnowledges">刷新</el-button>
            </template>
            <el-table :data="deletedKnowledges" v-loading="loadingKnowledge" empty-text="暂无已删除知识">
              <el-table-column prop="id" label="ID" width="60" />
              <el-table-column prop="title" label="标题" min-width="150" show-overflow-tooltip />
              <el-table-column label="发布者" width="100">
                <template #default="{ row }">
                  <span>{{ row.user?.nickname || row.user?.username || '未知' }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="created_at" label="创建时间" width="160" :formatter="formatTime" />
              <el-table-column label="操作" width="150" fixed="right">
                <template #default="{ row }">
                  <el-button size="small" @click="showKnowledgeDetail(row)">详情</el-button>
                  <el-button size="small" type="primary" @click="editKnowledge(row)">编辑</el-button>
                  <el-button size="small" type="success" @click="restoreKnowledge(row.id)">恢复</el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-tab-pane>

        <!-- 事务管理 -->
        <el-tab-pane label="事务管理" name="affairs">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-card>
                <template #header>
                  <span class="card-title">待审核 ({{ pendingAuditAffairs.length }})</span>
                  <el-button size="small" @click="loadPendingAuditAffairs">刷新</el-button>
                </template>
                <el-table :data="pendingAuditAffairs" v-loading="loadingAffairs" height="350" empty-text="暂无待审核事务">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="title" label="标题" min-width="120" show-overflow-tooltip />
                  <el-table-column prop="type" label="类型" width="80">
                    <template #default="{ row }">
                      <el-tag size="small">{{ getAffairTypeText(row.type) }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="address" label="地点" width="100" show-overflow-tooltip />
                  <el-table-column label="操作" width="120" fixed="right">
                    <template #default="{ row }">
                      <el-button size="small" type="success" @click="auditAffair(row.id, true)">通过</el-button>
                      <el-button size="small" type="danger" @click="auditAffair(row.id, false)">退回</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
            <el-col :span="12">
              <el-card>
                <template #header>
                  <span class="card-title">待处理 ({{ pendingProcessAffairs.length }})</span>
                  <el-button size="small" @click="loadPendingProcessAffairs">刷新</el-button>
                </template>
                <el-table :data="pendingProcessAffairs" v-loading="loadingAffairs" height="350" empty-text="暂无待处理事务">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="title" label="标题" min-width="120" show-overflow-tooltip />
                  <el-table-column prop="type" label="类型" width="80">
                    <template #default="{ row }">
                      <el-tag size="small">{{ getAffairTypeText(row.type) }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="address" label="地点" width="100" show-overflow-tooltip />
                  <el-table-column label="操作" width="80" fixed="right">
                    <template #default="{ row }">
                      <el-button size="small" @click="showAffairDetail(row)">详情</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
          </el-row>
          <el-row :gutter="20" style="margin-top: 20px;">
            <el-col :span="12">
              <el-card>
                <template #header>
                  <span class="card-title">审核不通过 ({{ rejectedAffairs.length }})</span>
                  <el-button size="small" @click="loadRejectedAffairs">刷新</el-button>
                </template>
                <el-table :data="rejectedAffairs" v-loading="loadingAffairs" height="300" empty-text="暂无审核不通过事务">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="title" label="标题" min-width="120" show-overflow-tooltip />
                  <el-table-column prop="type" label="类型" width="80">
                    <template #default="{ row }">
                      <el-tag size="small">{{ getAffairTypeText(row.type) }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column label="操作" width="80" fixed="right">
                    <template #default="{ row }">
                      <el-button size="small" @click="showAffairDetail(row)">详情</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
            <el-col :span="12">
              <el-card>
                <template #header>
                  <span class="card-title">处理中 ({{ processingAffairs.length }})</span>
                  <el-button size="small" @click="loadProcessingAffairs">刷新</el-button>
                </template>
                <el-table :data="processingAffairs" v-loading="loadingAffairs" height="300" empty-text="暂无处理中事务">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="title" label="标题" min-width="120" show-overflow-tooltip />
                  <el-table-column prop="type" label="类型" width="80">
                    <template #default="{ row }">
                      <el-tag size="small">{{ getAffairTypeText(row.type) }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="handler_name" label="处理人" width="80" />
                  <el-table-column label="操作" width="80" fixed="right">
                    <template #default="{ row }">
                      <el-button size="small" @click="showAffairDetail(row)">详情</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
          </el-row>
          <el-card style="margin-top: 20px;">
            <template #header>
              <span class="card-title">已完成 ({{ completedAffairs.length }})</span>
              <el-button size="small" @click="loadCompletedAffairs">刷新</el-button>
            </template>
            <el-table :data="completedAffairs" v-loading="loadingAffairs" empty-text="暂无已完成事务">
              <el-table-column prop="id" label="ID" width="60" />
              <el-table-column prop="title" label="标题" min-width="150" show-overflow-tooltip />
              <el-table-column prop="type" label="类型" width="80">
                <template #default="{ row }">
                  <el-tag size="small">{{ getAffairTypeText(row.type) }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="handler_name" label="处理人" width="80" />
              <el-table-column label="提交人" width="100">
                <template #default="{ row }">
                  <span>{{ row.user?.nickname || row.user?.username || '未知' }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="completed_time" label="完成时间" width="160" :formatter="formatTime" />
              <el-table-column label="操作" width="80" fixed="right">
                <template #default="{ row }">
                  <el-button size="small" @click="showAffairDetail(row)">详情</el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-tab-pane>

        <!-- 事务处理人员管理 -->
        <el-tab-pane label="处理人员管理" name="processors">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-card>
                <template #header>
                  <span class="card-title">待审核 ({{ pendingProcessors.length }})</span>
                  <el-button size="small" @click="loadPendingProcessors">刷新</el-button>
                </template>
                <el-table :data="pendingProcessors" v-loading="loadingProcessors" height="350" empty-text="暂无待审核申请">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="real_name" label="真实姓名" width="80" />
                  <el-table-column prop="phone" label="电话" width="120" />
                  <el-table-column prop="address" label="负责区域" width="150" show-overflow-tooltip />
                  <el-table-column label="操作" width="120">
                    <template #default="{ row }">
                      <el-button size="small" type="success" @click="approveProcessor(row.id)">通过</el-button>
                      <el-button size="small" type="danger" @click="rejectProcessor(row.id)">拒绝</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
            <el-col :span="12">
              <el-card>
                <template #header>
                  <span class="card-title">已拒绝 ({{ rejectedProcessors.length }})</span>
                  <el-button size="small" @click="loadRejectedProcessors">刷新</el-button>
                </template>
                <el-table :data="rejectedProcessors" v-loading="loadingProcessors" height="350" empty-text="暂无已拒绝申请">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="real_name" label="真实姓名" width="80" />
                  <el-table-column prop="phone" label="电话" width="120" />
                  <el-table-column prop="address" label="负责区域" width="150" show-overflow-tooltip />
                  <el-table-column label="操作" width="80">
                    <template #default="{ row }">
                      <el-button size="small" type="success" @click="reApproveProcessor(row.id)">重审</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
          </el-row>
          <el-card style="margin-top: 20px;">
            <template #header>
              <span class="card-title">已通过处理人员 ({{ approvedProcessors.length }})</span>
              <el-button size="small" @click="loadApprovedProcessors">刷新</el-button>
            </template>
            <el-table :data="approvedProcessors" v-loading="loadingProcessors" empty-text="暂无已通过处理人员">
              <el-table-column prop="id" label="ID" width="60" />
              <el-table-column prop="real_name" label="真实姓名" width="100" />
              <el-table-column prop="phone" label="电话" width="120" />
              <el-table-column prop="address" label="负责区域" width="150" show-overflow-tooltip />
              <el-table-column label="申请人" width="100">
                <template #default="{ row }">
                  <span>{{ row.user?.nickname || row.user?.username || '未知' }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="created_at" label="申请时间" width="160" :formatter="formatTime" />
              <el-table-column label="操作" width="100">
                <template #default="{ row }">
                  <el-button size="small" type="danger" @click="disableProcessor(row.id)">禁用</el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-tab-pane>

        <!-- 专家管理 -->
        <el-tab-pane label="专家管理" name="experts">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-card>
                <template #header>
                  <span class="card-title">待审核 ({{ pendingExperts.length }})</span>
                  <el-button size="small" @click="loadPendingExperts">刷新</el-button>
                </template>
                <el-table :data="pendingExperts" v-loading="loadingExperts" height="350" empty-text="暂无待审核专家">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="real_name" label="真实姓名" width="80" />
                  <el-table-column prop="profession" label="专业" width="80">
                    <template #default="{ row }">
                      <el-tag size="small">{{ row.profession }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="company" label="单位" min-width="100" show-overflow-tooltip />
                  <el-table-column label="操作" width="120">
                    <template #default="{ row }">
                      <el-button size="small" type="success" @click="approveExpert(row.id)">通过</el-button>
                      <el-button size="small" type="danger" @click="rejectExpert(row.id)">拒绝</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
            <el-col :span="12">
              <el-card>
                <template #header>
                  <span class="card-title">已拒绝 ({{ rejectedExperts.length }})</span>
                  <el-button size="small" @click="loadRejectedExperts">刷新</el-button>
                </template>
                <el-table :data="rejectedExperts" v-loading="loadingExperts" height="350" empty-text="暂无已拒绝专家">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="real_name" label="真实姓名" width="80" />
                  <el-table-column prop="profession" label="专业" width="80">
                    <template #default="{ row }">
                      <el-tag size="small">{{ row.profession }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="company" label="单位" min-width="100" show-overflow-tooltip />
                  <el-table-column label="操作" width="80">
                    <template #default="{ row }">
                      <el-button size="small" type="success" @click="reApproveExpert(row.id)">重审</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
          </el-row>
          <el-card style="margin-top: 20px;">
            <template #header>
              <span class="card-title">已通过专家 ({{ approvedExperts.length }})</span>
            </template>
            <el-table :data="approvedExperts" v-loading="loadingExperts" empty-text="暂无已通过专家">
              <el-table-column prop="id" label="ID" width="60" />
              <el-table-column prop="real_name" label="真实姓名" width="100" />
              <el-table-column prop="profession" label="专业领域" width="100">
                <template #default="{ row }">
                  <el-tag size="small">{{ row.profession }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="title" label="职位" width="120" />
              <el-table-column prop="company" label="所属单位" min-width="150" />
              <el-table-column prop="user?.username" label="申请人" width="100" />
              <el-table-column prop="created_at" label="申请时间" width="160" :formatter="formatTime" />
              <el-table-column label="操作" width="100">
                <template #default="{ row }">
                  <el-button size="small" type="danger" @click="disableExpert(row.id)">禁用</el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-tab-pane>

        <!-- 敏感词管理 -->
        <el-tab-pane label="敏感词管理" name="sensitive">
          <el-card>
            <template #header>
              <div class="card-header">
                <span>敏感词列表 ({{ sensitiveWords.length }})</span>
                <el-button type="primary" size="small" @click="showAddWordDialog = true">添加敏感词</el-button>
              </div>
            </template>
            <el-table :data="sensitiveWords" v-loading="loadingSensitive" empty-text="暂无敏感词">
              <el-table-column prop="id" label="ID" width="60" />
              <el-table-column prop="word" label="敏感词" min-width="150" />
              <el-table-column prop="category" label="类别" width="100">
                <template #default="{ row }">
                  <el-tag size="small">{{ getCategoryText(row.category) }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="level" label="级别" width="100">
                <template #default="{ row }">
                  <el-tag :type="getLevelType(row.level)" size="small">{{ getLevelText(row.level) }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="created_at" label="添加时间" width="160" :formatter="formatTime" />
              <el-table-column label="操作" width="100" fixed="right">
                <template #default="{ row }">
                  <el-button size="small" type="danger" @click="deleteSensitiveWord(row.id)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>

          <el-card style="margin-top: 20px;">
            <template #header>
              <span>自动审核设置</span>
            </template>
            <el-form :model="autoReviewSetting" label-width="150px">
              <el-form-item label="审核模式">
                <el-radio-group v-model="autoReviewSetting.mode">
                  <el-radio label="all_pass">全部自动通过</el-radio>
                  <el-radio label="all_reject">全部自动驳回</el-radio>
                  <el-radio label="no_sensitive_pass">不含敏感词自动通过</el-radio>
                  <el-radio label="need_review_images">带图内容需人工复核</el-radio>
                </el-radio-group>
              </el-form-item>
              <el-form-item label="启用敏感词检测">
                <el-switch v-model="autoReviewSetting.enabled_sensitive" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="saveAutoReviewSetting">保存设置</el-button>
              </el-form-item>
            </el-form>
          </el-card>
        </el-tab-pane>

        <!-- 申诉管理 -->
        <el-tab-pane label="申诉管理" name="appeals">
          <el-card style="margin-bottom: 20px;">
            <template #header>
              <div class="card-header">
                <span>申诉列表</span>
                <div style="display: flex; gap: 10px;">
                  <el-select v-model="appealStatusFilter" placeholder="申诉状态" size="small" style="width: 130px;" clearable @change="loadAppeals">
                    <el-option label="待处理" value="pending" />
                    <el-option label="处理中" value="processing" />
                    <el-option label="已处理" value="resolved" />
                  </el-select>
                  <el-button size="small" @click="loadAppeals">刷新</el-button>
                </div>
              </div>
            </template>
            <el-table :data="appeals" v-loading="loadingAppeals" empty-text="暂无申诉记录">
              <el-table-column prop="id" label="ID" width="60" />
              <el-table-column label="事务" min-width="120">
                <template #default="{ row }">
                  <el-link type="primary" @click="$router.push(`/affair/${row.affair_id}`)">
                    {{ row.affair_title || '查看事务' }}
                  </el-link>
                </template>
              </el-table-column>
              <el-table-column label="申诉人" width="100">
                <template #default="{ row }">
                  <span>{{ row.applicant_name }}</span>
                  <el-tag size="small" :type="row.applicant_type === 'user' ? '' : 'warning'" style="margin-left: 4px;">
                    {{ row.applicant_type === 'user' ? '用户' : '处理人员' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="reason" label="申诉理由" min-width="180" show-overflow-tooltip />
              <el-table-column label="状态" width="100">
                <template #default="{ row }">
                  <el-tag :type="getAppealStatusTag(row.status)" size="small">{{ getAppealStatusText(row.status) }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="提交时间" width="160" :formatter="(row) => formatTimeStr(row.created_at)" />
              <el-table-column label="操作" width="220" fixed="right">
                <template #default="{ row }">
                  <el-button size="small" type="primary" @click="openAppealDetail(row)">查看详情</el-button>
                  <el-button v-if="row.status !== 'resolved'" size="small" type="success" @click="openProcessAppealDialog(row)">处理申诉</el-button>
                </template>
              </el-table-column>
            </el-table>
            <div style="margin-top: 15px; display: flex; justify-content: center;" v-if="appealTotal > appealPageSize">
              <el-pagination
                v-model:current-page="appealPage"
                :page-size="appealPageSize"
                :total="appealTotal"
                layout="prev, pager, next"
                @current-change="loadAppeals"
              />
            </div>
          </el-card>
        </el-tab-pane>

        <!-- 用户管理 -->
        <el-tab-pane label="用户管理" name="users">
          <el-tabs v-model="userSubTab" type="card">
            <el-tab-pane label="全部用户" name="allUsers">
              <el-card>
                <template #header>
                  <div class="card-header">
                    <span>注册用户 ({{ userTotal }})</span>
                    <div style="display: flex; gap: 10px;">
                      <el-input v-model="userKeyword" placeholder="搜索用户名/昵称/手机/邮箱" size="small" style="width: 220px;" clearable @keyup.enter="loadAllUsers" />
                      <el-select v-model="userRoleFilter" placeholder="角色筛选" size="small" style="width: 120px;" clearable @change="loadAllUsers">
                        <el-option label="普通用户" value="normal" />
                        <el-option label="农户" value="farmer" />
                        <el-option label="专家" value="expert" />
                        <el-option label="处理人员" value="processor" />
                        <el-option label="审核员" value="admin" />
                        <el-option label="系统管理员" value="sysadmin" />
                      </el-select>
                      <el-select v-model="userBannedFilter" placeholder="封禁状态" size="small" style="width: 120px;" clearable @change="loadAllUsers">
                        <el-option label="已封禁" value="true" />
                        <el-option label="未封禁" value="false" />
                      </el-select>
                      <el-button size="small" @click="loadAllUsers">刷新</el-button>
                    </div>
                  </div>
                </template>
                <el-table :data="allUsers" v-loading="loadingUsers" empty-text="暂无用户">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="username" label="用户名" width="120" />
                  <el-table-column prop="nickname" label="昵称" width="120">
                    <template #default="{ row }">{{ row.nickname || row.username }}</template>
                  </el-table-column>
                  <el-table-column prop="role" label="角色" width="110">
                    <template #default="{ row }">
                      <el-tag :type="getRoleType(row.role)" size="small">{{ getRoleText(row.role) }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="phone" label="手机号" width="130">
                    <template #default="{ row }">{{ row.phone || '-' }}</template>
                  </el-table-column>
                  <el-table-column prop="email" label="邮箱" min-width="160" show-overflow-tooltip>
                    <template #default="{ row }">{{ row.email || '-' }}</template>
                  </el-table-column>
                  <el-table-column label="封禁状态" width="180">
                    <template #default="{ row }">
                      <template v-if="row.banned">
                        <el-tag type="danger" size="small">已封禁</el-tag>
                        <span v-if="row.banned_until" style="margin-left: 4px; font-size: 12px; color: #999;">
                          至 {{ formatShortTime(row.banned_until) }}
                        </span>
                        <span v-else style="margin-left: 4px; font-size: 12px; color: #999;">永久</span>
                      </template>
                      <el-tag v-else type="success" size="small">正常</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column label="注册时间" width="160" :formatter="formatTime" />
                  <el-table-column label="操作" width="180" fixed="right">
                    <template #default="{ row }">
                      <template v-if="row.banned">
                        <el-button size="small" type="success" @click="handleUnbanUser(row.id)">解封</el-button>
                      </template>
                      <template v-else>
                        <el-button size="small" type="danger" @click="openBanDialog(row)">封禁</el-button>
                      </template>
                      <el-button size="small" @click="showUserInfoDialog(row)">详情</el-button>
                    </template>
                  </el-table-column>
                </el-table>
                <div style="margin-top: 15px; display: flex; justify-content: center;">
                  <el-pagination
                    v-model:current-page="userPage"
                    v-model:page-size="userPageSize"
                    :total="userTotal"
                    :page-sizes="[10, 20, 50]"
                    layout="total, sizes, prev, pager, next"
                    @change="loadAllUsers"
                  />
                </div>
              </el-card>
            </el-tab-pane>

            <el-tab-pane label="已注销用户" name="deletedUsers">
              <el-card>
                <template #header>
                  <div class="card-header">
                    <span>已注销用户 ({{ deletedUserTotal }})</span>
                    <div style="display: flex; gap: 10px;">
                      <el-input v-model="deletedUserKeyword" placeholder="搜索用户名/昵称" size="small" style="width: 220px;" clearable @keyup.enter="loadDeletedUsers" />
                      <el-button size="small" @click="loadDeletedUsers">刷新</el-button>
                    </div>
                  </div>
                </template>
                <el-table :data="deletedUsers" v-loading="loadingDeletedUsers" empty-text="暂无已注销用户">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="username" label="用户名" width="120" />
                  <el-table-column prop="nickname" label="昵称" width="120" />
                  <el-table-column prop="role" label="角色" width="110">
                    <template #default="{ row }">
                      <el-tag :type="getRoleType(row.role)" size="small">{{ getRoleText(row.role) }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="phone" label="手机号" width="130">
                    <template #default="{ row }">{{ row.phone || '-' }}</template>
                  </el-table-column>
                  <el-table-column prop="email" label="邮箱" min-width="160" show-overflow-tooltip>
                    <template #default="{ row }">{{ row.email || '-' }}</template>
                  </el-table-column>
                  <el-table-column label="注销时间" width="160" :formatter="(row) => formatTimeStr(row.deleted_at)" />
                  <el-table-column label="操作" width="120" fixed="right">
                    <template #default="{ row }">
                      <el-button size="small" type="primary" @click="openEditDeletedUserDialog(row)">编辑</el-button>
                    </template>
                  </el-table-column>
                </el-table>
                <div style="margin-top: 15px; display: flex; justify-content: center;">
                  <el-pagination
                    v-model:current-page="deletedUserPage"
                    v-model:page-size="deletedUserPageSize"
                    :total="deletedUserTotal"
                    :page-sizes="[10, 20, 50]"
                    layout="total, sizes, prev, pager, next"
                    @change="loadDeletedUsers"
                  />
                </div>
              </el-card>
            </el-tab-pane>
          </el-tabs>
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- 添加敏感词对话框 -->
    <el-dialog v-model="showAddWordDialog" title="添加敏感词" width="500px">
      <el-form :model="newWordForm" label-width="100px">
        <el-form-item label="敏感词" required>
          <el-input v-model="newWordForm.word" placeholder="请输入敏感词" />
        </el-form-item>
        <el-form-item label="类别">
          <el-select v-model="newWordForm.category" style="width: 100%">
            <el-option label="政治敏感" value="politics" />
            <el-option label="色情低俗" value="porn" />
            <el-option label="广告推广" value="advertising" />
            <el-option label="赌博博彩" value="gamble" />
            <el-option label="欺诈诈骗" value="fraud" />
            <el-option label="其他违规" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="级别">
          <el-radio-group v-model="newWordForm.level">
            <el-radio :label="1">轻度</el-radio>
            <el-radio :label="2">中度</el-radio>
            <el-radio :label="3">重度</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddWordDialog = false">取消</el-button>
        <el-button type="primary" @click="addSensitiveWord">添加</el-button>
      </template>
    </el-dialog>

    <!-- 商品详情对话框 -->
    <el-dialog v-model="showProductDialog" title="商品详情" width="800px">
      <div v-if="currentProduct" class="product-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="商品名称">{{ currentProduct.title }}</el-descriptions-item>
          <el-descriptions-item label="价格">¥{{ currentProduct.price }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ currentProduct.type }}</el-descriptions-item>
          <el-descriptions-item label="库存">{{ currentProduct.stock }}</el-descriptions-item>
          <el-descriptions-item label="产地">{{ currentProduct.address || '未设置' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="currentProduct.status === 1 ? 'success' : 'info'">
              {{ currentProduct.status === 1 ? '上架' : '下架' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="详情" :span="2">{{ currentProduct.content }}</el-descriptions-item>
        </el-descriptions>
        <div v-if="currentProduct.image_url" style="margin-top: 20px;">
          <p>商品图片：</p>
          <img :src="currentProduct.image_url" style="max-width: 300px; border-radius: 8px;" />
        </div>
        <el-divider>评论区</el-divider>
        <div class="comments-section">
          <div v-if="loadingComments" v-loading="loadingComments" style="padding: 20px; text-align: center;">加载中...</div>
          <div v-else-if="productComments.length === 0" style="text-align: center; color: #999; padding: 20px;">暂无评论</div>
          <div v-else class="comment-list">
            <div v-for="comment in productComments" :key="comment.id" class="comment-item">
              <div class="comment-header">
                <el-avatar :src="comment.user?.avatar_url" :size="32">{{ comment.user?.username?.[0] }}</el-avatar>
                <span class="comment-username">{{ comment.user?.username }}</span>
                <span class="comment-time">{{ formatTimeStr(comment.created_at) }}</span>
              </div>
              <div class="comment-content">{{ comment.content }}</div>
              <div class="comment-actions">
                <span class="like-btn"><el-icon><Star /></el-icon> {{ comment.likes || 0 }}</span>
              </div>
              <!-- 二级评论 -->
              <div v-if="comment.children && comment.children.length > 0" class="reply-list">
                <div v-for="reply in comment.children" :key="reply.id" class="reply-item">
                  <div class="reply-header">
                    <el-avatar :src="reply.user?.avatar_url" :size="24">{{ reply.user?.username?.[0] }}</el-avatar>
                    <span class="reply-username">{{ reply.user?.username }}</span>
                    <span v-if="reply.reply_to_user" class="reply-to">回复 {{ reply.reply_to_user?.username }}</span>
                    <span class="reply-time">{{ formatTimeStr(reply.created_at) }}</span>
                  </div>
                  <div class="reply-content">{{ reply.content }}</div>
                  <div class="reply-actions">
                    <span class="like-btn"><el-icon><Star /></el-icon> {{ reply.likes || 0 }}</span>
                  </div>
                  <!-- 三级评论 -->
                  <div v-if="reply.children && reply.children.length > 0" class="reply-list">
                    <div v-for="reply2 in reply.children" :key="reply2.id" class="reply-item">
                      <div class="reply-header">
                        <el-avatar :src="reply2.user?.avatar_url" :size="24">{{ reply2.user?.username?.[0] }}</el-avatar>
                        <span class="reply-username">{{ reply2.user?.username }}</span>
                        <span v-if="reply2.reply_to_user" class="reply-to">回复 {{ reply2.reply_to_user?.username }}</span>
                        <span class="reply-time">{{ formatTimeStr(reply2.created_at) }}</span>
                      </div>
                      <div class="reply-content">{{ reply2.content }}</div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 动态详情对话框 -->
    <el-dialog v-model="showPostDialog" title="动态详情" width="800px">
      <div v-if="currentPost" class="post-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="标题">{{ currentPost.title }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ currentPost.type }}</el-descriptions-item>
          <el-descriptions-item label="发布者">{{ currentPost.user?.username }}</el-descriptions-item>
          <el-descriptions-item label="点赞">{{ currentPost.likes }}</el-descriptions-item>
          <el-descriptions-item label="浏览">{{ currentPost.views }}</el-descriptions-item>
          <el-descriptions-item label="内容" :span="2"><div v-html="currentPost.content" style="max-height: 400px; overflow-y: auto; line-height: 1.8;" /></el-descriptions-item>
        </el-descriptions>
        <div v-if="currentPost.images" style="margin-top: 20px;">
          <p>图片：</p>
          <div class="post-images">
            <img v-for="(img, idx) in parseImages(currentPost.images)" :key="idx" :src="img" style="max-width: 200px; margin-right: 10px; border-radius: 8px;" />
          </div>
        </div>
        <el-divider>评论区</el-divider>
        <div class="comments-section">
          <div v-if="loadingPostComments" v-loading="loadingPostComments" style="padding: 20px; text-align: center;">加载中...</div>
          <div v-else-if="postComments.length === 0" style="text-align: center; color: #999; padding: 20px;">暂无评论</div>
          <div v-else class="comment-list">
            <div v-for="comment in postComments" :key="comment.id" class="comment-item">
              <div class="comment-header">
                <el-avatar :src="comment.user?.avatar_url" :size="32">{{ comment.user?.username?.[0] }}</el-avatar>
                <span class="comment-username">{{ comment.user?.username }}</span>
                <span class="comment-time">{{ formatTimeStr(comment.created_at) }}</span>
              </div>
              <div class="comment-content">{{ comment.content }}</div>
              <div class="comment-actions">
                <span class="like-btn"><el-icon><Star /></el-icon> {{ comment.likes || 0 }}</span>
              </div>
              <!-- 二级评论 -->
              <div v-if="comment.children && comment.children.length > 0" class="reply-list">
                <div v-for="reply in comment.children" :key="reply.id" class="reply-item">
                  <div class="reply-header">
                    <el-avatar :src="reply.user?.avatar_url" :size="24">{{ reply.user?.username?.[0] }}</el-avatar>
                    <span class="reply-username">{{ reply.user?.username }}</span>
                    <span v-if="reply.reply_to_user" class="reply-to">回复 {{ reply.reply_to_user?.username }}</span>
                    <span class="reply-time">{{ formatTimeStr(reply.created_at) }}</span>
                  </div>
                  <div class="reply-content">{{ reply.content }}</div>
                  <!-- 三级评论 -->
                  <div v-if="reply.children && reply.children.length > 0" class="reply-list">
                    <div v-for="reply2 in reply.children" :key="reply2.id" class="reply-item">
                      <div class="reply-header">
                        <el-avatar :src="reply2.user?.avatar_url" :size="24">{{ reply2.user?.username?.[0] }}</el-avatar>
                        <span class="reply-username">{{ reply2.user?.username }}</span>
                        <span v-if="reply2.reply_to_user" class="reply-to">回复 {{ reply2.reply_to_user?.username }}</span>
                        <span class="reply-time">{{ formatTimeStr(reply2.created_at) }}</span>
                      </div>
                      <div class="reply-content">{{ reply2.content }}</div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 知识详情对话框 -->
    <el-dialog v-model="showKnowledgeDetailDialog" title="知识详情" width="800px">
      <div v-if="currentKnowledge" class="knowledge-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="标题">{{ currentKnowledge.title }}</el-descriptions-item>
          <el-descriptions-item label="发布者">{{ currentKnowledge.user?.username }}</el-descriptions-item>
          <el-descriptions-item label="内容" :span="2"><div v-html="currentKnowledge.content" style="max-height: 400px; overflow-y: auto; line-height: 1.8;" /></el-descriptions-item>
        </el-descriptions>
        <div v-if="currentKnowledge.image_url" style="margin-top: 20px;">
          <p>封面图片：</p>
          <img :src="currentKnowledge.image_url" style="max-width: 300px; border-radius: 8px;" />
        </div>
        <el-divider>评论区</el-divider>
        <div class="comments-section">
          <div v-if="loadingKnowledgeComments" v-loading="loadingKnowledgeComments" style="padding: 20px; text-align: center;">加载中...</div>
          <div v-else-if="knowledgeComments.length === 0" style="text-align: center; color: #999; padding: 20px;">暂无评论</div>
          <div v-else class="comment-list">
            <div v-for="comment in knowledgeComments" :key="comment.id" class="comment-item">
              <div class="comment-header">
                <el-avatar :src="comment.user?.avatar_url" :size="32">{{ comment.user?.username?.[0] }}</el-avatar>
                <span class="comment-username">{{ comment.user?.username }}</span>
                <span class="comment-time">{{ formatTimeStr(comment.created_at) }}</span>
              </div>
              <div class="comment-content">{{ comment.content }}</div>
              <!-- 二级评论 -->
              <div v-if="comment.children && comment.children.length > 0" class="reply-list">
                <div v-for="reply in comment.children" :key="reply.id" class="reply-item">
                  <div class="reply-header">
                    <el-avatar :src="reply.user?.avatar_url" :size="24">{{ reply.user?.username?.[0] }}</el-avatar>
                    <span class="reply-username">{{ reply.user?.username }}</span>
                    <span class="reply-time">{{ formatTimeStr(reply.created_at) }}</span>
                  </div>
                  <div class="reply-content">{{ reply.content }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 编辑商品对话框 -->
    <el-dialog v-model="showEditProductDialog" title="编辑商品" width="600px">
      <el-form :model="editProductForm" label-width="80px">
        <el-form-item label="商品名称">
          <el-input v-model="editProductForm.title" />
        </el-form-item>
        <el-form-item label="商品类型">
          <el-select v-model="editProductForm.type" style="width: 100%">
            <el-option label="种子" value="seed" />
            <el-option label="肥料" value="fertilizer" />
            <el-option label="农药" value="pesticide" />
            <el-option label="农具" value="tool" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="价格">
          <el-input-number v-model="editProductForm.price" :min="0" :precision="2" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="库存">
          <el-input-number v-model="editProductForm.stock" :min="0" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="商品状态">
          <el-radio-group v-model="editProductForm.status">
            <el-radio :label="1">上架</el-radio>
            <el-radio :label="0">下架</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="商品详情">
          <el-input v-model="editProductForm.content" type="textarea" :rows="4" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditProductDialog = false">取消</el-button>
        <el-button type="primary" @click="handleUpdateProduct" :loading="savingProduct">保存</el-button>
      </template>
    </el-dialog>

    <!-- 编辑知识对话框 -->
    <el-dialog v-model="showEditKnowledgeDialog" title="编辑知识" width="600px">
      <el-form :model="editKnowledgeForm" label-width="80px">
        <el-form-item label="标题">
          <el-input v-model="editKnowledgeForm.title" />
        </el-form-item>
        <el-form-item label="内容">
          <el-input v-model="editKnowledgeForm.content" type="textarea" :rows="6" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditKnowledgeDialog = false">取消</el-button>
        <el-button type="primary" @click="handleUpdateKnowledge" :loading="savingKnowledge">保存</el-button>
      </template>
    </el-dialog>

    <!-- 编辑动态对话框 -->
    <el-dialog v-model="showEditPostDialog" title="编辑动态" width="600px">
      <el-form :model="editPostForm" label-width="80px">
        <el-form-item label="标题">
          <el-input v-model="editPostForm.title" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="editPostForm.type" style="width: 100%">
            <el-option label="种植" value="planting" />
            <el-option label="养殖" value="breeding" />
            <el-option label="技术" value="technology" />
            <el-option label="政策" value="policy" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="内容">
          <el-input v-model="editPostForm.content" type="textarea" :rows="6" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditPostDialog = false">取消</el-button>
        <el-button type="primary" @click="handleUpdatePost" :loading="savingPost">保存</el-button>
      </template>
    </el-dialog>

    <!-- 事务详情对话框 -->
    <el-dialog v-model="showAffairDialog" title="事务详情" width="700px">
      <div v-if="currentAffair" class="affair-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="事务标题" :span="2">{{ currentAffair.title }}</el-descriptions-item>
          <el-descriptions-item label="事务类型">{{ getAffairTypeText(currentAffair.type) }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getAffairStatusType(currentAffair.status)" size="small">{{ getAffairStatusText(currentAffair.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="事发地点" :span="2">{{ currentAffair.address }}</el-descriptions-item>
          <el-descriptions-item label="提交人" :span="2">
            <span>{{ currentAffair.user?.nickname || currentAffair.user?.username || '未知' }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="提交时间" :span="2">{{ formatTimeStr(currentAffair.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="问题描述" :span="2">
            <div style="white-space: pre-wrap;">{{ currentAffair.content }}</div>
          </el-descriptions-item>
        </el-descriptions>
        <div v-if="currentAffair.images" style="margin-top: 20px;">
          <p>现场图片：</p>
          <div class="images-grid">
            <el-image
              v-for="(img, idx) in parseImages(currentAffair.images)"
              :key="idx"
              :src="img"
              fit="cover"
              class="detail-img"
              :preview-src-list="parseImages(currentAffair.images)"
            />
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 申诉详情对话框 -->
    <el-dialog v-model="showAppealDetailDialog" title="申诉详情" width="700px">
      <div v-if="currentAppeal" class="appeal-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="事务标题" :span="2">
            <el-link type="primary" @click="$router.push(`/affair/${currentAppeal.affair_id}`)">
              {{ currentAppeal.affair_title || '查看事务详情' }}
            </el-link>
          </el-descriptions-item>
          <el-descriptions-item label="申诉人">{{ currentAppeal.applicant_name }}</el-descriptions-item>
          <el-descriptions-item label="身份">
            <el-tag :type="currentAppeal.applicant_type === 'user' ? '' : 'warning'" size="small">
              {{ currentAppeal.applicant_type === 'user' ? '用户' : '处理人员' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getAppealStatusTag(currentAppeal.status)" size="small">{{ getAppealStatusText(currentAppeal.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="提交时间" :span="2">{{ formatTimeStr(currentAppeal.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="申诉理由" :span="2">
            <div style="white-space: pre-wrap;">{{ currentAppeal.reason }}</div>
          </el-descriptions-item>
        </el-descriptions>
        <div v-if="currentAppeal.status === 'resolved'" class="appeal-result">
          <el-divider>处理结果</el-divider>
          <el-alert type="success" :closable="false" show-icon>
            <template #title>
              <div><strong>处理人：</strong>{{ currentAppeal.handler_name }}</div>
              <div><strong>处理时间：</strong>{{ formatTimeStr(currentAppeal.processed_at) }}</div>
              <div><strong>处理结果：</strong>{{ currentAppeal.result }}</div>
            </template>
          </el-alert>
        </div>
      </div>
    </el-dialog>

    <!-- 处理申诉对话框 -->
    <el-dialog v-model="showProcessAppealDialog" title="处理申诉" width="600px">
      <el-form :model="processAppealForm" label-width="100px">
        <el-form-item label="申诉人">
          <span>{{ currentAppeal?.applicant_name }}
            <el-tag size="small" :type="currentAppeal?.applicant_type === 'user' ? '' : 'warning'" style="margin-left: 4px;">
              {{ currentAppeal?.applicant_type === 'user' ? '用户' : '处理人员' }}
            </el-tag>
          </span>
        </el-form-item>
        <el-form-item label="申诉理由">
          <div style="white-space: pre-wrap; color: #666; max-height: 120px; overflow-y: auto;">
            {{ currentAppeal?.reason }}
          </div>
        </el-form-item>
        <el-form-item label="处理结果" required>
          <el-input v-model="processAppealForm.result" type="textarea" :rows="4" placeholder="请输入处理结果说明" />
        </el-form-item>
        <el-form-item label="是否批准">
          <el-switch v-model="processAppealForm.approve" active-text="批准" inactive-text="驳回" />
        </el-form-item>
        <el-alert title="提示：管理员处理后，仍需用户确认事务方可最终完成" type="info" :closable="false" show-icon style="margin-top: 10px;" />
      </el-form>
      <template #footer>
        <el-button @click="showProcessAppealDialog = false">取消</el-button>
        <el-button type="primary" @click="handleProcessAppeal" :loading="processingAppeal">确认处理</el-button>
      </template>
    </el-dialog>

    <!-- 管理员直接处理事务对话框 -->
    <el-dialog v-model="showAdminProcessAffairDialog" title="管理员处理事务" width="600px">
      <el-form :model="adminProcessAffairForm" label-width="100px">
        <el-form-item label="事务标题">
          <span>{{ currentAffair?.title }}</span>
        </el-form-item>
        <el-form-item label="当前状态">
          <el-tag :type="getAffairStatusType(currentAffair?.status)" size="small">{{ getAffairStatusText(currentAffair?.status) }}</el-tag>
        </el-form-item>
        <el-form-item label="处理备注" required>
          <el-input v-model="adminProcessAffairForm.note" type="textarea" :rows="4" placeholder="请输入管理员处理备注" />
        </el-form-item>
        <el-alert title="管理员处理后，事务状态不变，仍需用户确认完成" type="warning" :closable="false" show-icon style="margin-top: 10px;" />
      </el-form>
      <template #footer>
        <el-button @click="showAdminProcessAffairDialog = false">取消</el-button>
        <el-button type="primary" @click="handleAdminProcessAffair" :loading="processingAppeal">确认处理</el-button>
      </template>
    </el-dialog>

    <!-- 拒绝处理人员对话框 -->
    <el-dialog v-model="showRejectProcessorDialog" title="拒绝申请" width="500px">
      <el-form :model="rejectProcessorForm" label-width="100px">
        <el-form-item label="拒绝原因" required>
          <el-input v-model="rejectProcessorForm.reason" type="textarea" :rows="4" placeholder="请输入拒绝原因" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRejectProcessorDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmRejectProcessor">确认拒绝</el-button>
      </template>
    </el-dialog>

    <!-- 封禁用户对话框 -->
    <el-dialog v-model="showBanDialog" title="封禁用户" width="500px">
      <el-form :model="banForm" label-width="100px">
        <el-form-item label="用户名">
          <span>{{ banForm.username }}</span>
        </el-form-item>
        <el-form-item label="封禁期限">
          <el-radio-group v-model="banForm.banType" @change="onBanTypeChange">
            <el-radio label="forever">永久封禁</el-radio>
            <el-radio label="duration">设定时长</el-radio>
            <el-radio label="until">指定日期</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="banForm.banType === 'duration'" label="封禁时长">
          <el-input-number v-model="banForm.days" :min="1" :max="3650" style="width: 120px;" /> 天
          <span style="margin-left: 10px; color: #999; font-size: 12px;">预计截止：{{ computedBanEndDate }}</span>
        </el-form-item>
        <el-form-item v-if="banForm.banType === 'until'" label="截止日期">
          <el-date-picker
            v-model="banForm.untilDate"
            type="datetime"
            placeholder="选择封禁截止时间"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DD HH:mm:ss"
            :disabled-date="disabledDate"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showBanDialog = false">取消</el-button>
        <el-button type="danger" @click="confirmBanUser" :loading="banning">确认封禁</el-button>
      </template>
    </el-dialog>

    <!-- 用户信息详情对话框 -->
    <el-dialog v-model="showUserInfoDialogVisible" title="用户信息" width="550px">
      <div v-if="currentUserInfo" class="user-info-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="ID">{{ currentUserInfo.id }}</el-descriptions-item>
          <el-descriptions-item label="用户名">{{ currentUserInfo.username }}</el-descriptions-item>
          <el-descriptions-item label="昵称">{{ currentUserInfo.nickname || currentUserInfo.username }}</el-descriptions-item>
          <el-descriptions-item label="角色">
            <el-tag :type="getRoleType(currentUserInfo.role)" size="small">{{ getRoleText(currentUserInfo.role) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="手机号">{{ currentUserInfo.phone || '-' }}</el-descriptions-item>
          <el-descriptions-item label="邮箱">{{ currentUserInfo.email || '-' }}</el-descriptions-item>
          <el-descriptions-item label="签名" :span="2">{{ currentUserInfo.signature || '-' }}</el-descriptions-item>
          <el-descriptions-item label="封禁状态">
            <el-tag v-if="currentUserInfo.banned" type="danger" size="small">已封禁</el-tag>
            <el-tag v-else type="success" size="small">正常</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="封禁截止">
            <span v-if="currentUserInfo.banned_until">{{ formatTimeStr(currentUserInfo.banned_until) }}</span>
            <span v-else-if="currentUserInfo.banned">永久</span>
            <span v-else>-</span>
          </el-descriptions-item>
          <el-descriptions-item label="注册时间" :span="2">{{ formatTimeStr(currentUserInfo.created_at) }}</el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>

    <!-- 编辑已注销用户对话框 -->
    <el-dialog v-model="showEditDeletedUserDialogVisible" title="编辑已注销用户" width="500px">
      <el-form :model="editDeletedUserForm" label-width="100px">
        <el-form-item label="用户名">
          <span>{{ editDeletedUserForm.username }}</span>
        </el-form-item>
        <el-form-item label="昵称">
          <el-input v-model="editDeletedUserForm.nickname" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="editDeletedUserForm.role" style="width: 100%">
            <el-option label="普通用户" value="normal" />
            <el-option label="农户" value="farmer" />
            <el-option label="专家" value="expert" />
            <el-option label="处理人员" value="processor" />
            <el-option label="审核员" value="admin" />
            <el-option label="系统管理员" value="sysadmin" />
          </el-select>
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="editDeletedUserForm.phone" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="editDeletedUserForm.email" />
        </el-form-item>
        <el-form-item label="签名">
          <el-input v-model="editDeletedUserForm.signature" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditDeletedUserDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmEditDeletedUser" :loading="savingDeletedUser">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Star } from '@element-plus/icons-vue'
import api from '@/utils/api'
import { getUserInfo, removeToken, removeUserInfo } from '@/utils/auth'

const router = useRouter()
const user = computed(() => getUserInfo())

const currentUser = ref({
  username: '',
  nickname: '',
  avatar_url: ''
})

const activeTab = ref('products')

// 商品管理
const pendingProducts = ref([])
const rejectedProducts = ref([])
const onShelfProducts = ref([])
const loadingProducts = ref(false)
const showProductDialog = ref(false)
const showEditProductDialog = ref(false)
const currentProduct = ref(null)
const productComments = ref([])
const loadingComments = ref(false)
const savingProduct = ref(false)
const editProductForm = reactive({
  id: null,
  title: '',
  type: 'seed',
  price: 0,
  content: '',
  stock: 0,
  status: 1
})

// 动态管理
const pendingPosts = ref([])
const rejectedPosts = ref([])
const publishedPosts = ref([])
const deletedPosts = ref([])
const loadingPosts = ref(false)
const showPostDialog = ref(false)
const showEditPostDialog = ref(false)
const currentPost = ref(null)
const postComments = ref([])
const loadingPostComments = ref(false)
const savingPost = ref(false)
const editPostForm = reactive({
  id: null,
  title: '',
  content: '',
  type: 'other',
  images: ''
})

// 知识管理
const pendingKnowledges = ref([])
const rejectedKnowledges = ref([])
const publishedKnowledges = ref([])
const deletedKnowledges = ref([])
const loadingKnowledge = ref(false)
const showKnowledgeDetailDialog = ref(false)
const showEditKnowledgeDialog = ref(false)
const currentKnowledge = ref(null)
const knowledgeComments = ref([])
const loadingKnowledgeComments = ref(false)
const savingKnowledge = ref(false)
const editKnowledgeForm = reactive({
  id: null,
  title: '',
  content: ''
})

// 专家管理
const pendingExperts = ref([])
const rejectedExperts = ref([])
const approvedExperts = ref([])
const loadingExperts = ref(false)

// 事务管理
const pendingAuditAffairs = ref([])
const pendingProcessAffairs = ref([])
const rejectedAffairs = ref([])
const processingAffairs = ref([])
const completedAffairs = ref([])
const loadingAffairs = ref(false)
const showAffairDialog = ref(false)
const currentAffair = ref(null)

// 申诉管理
const appeals = ref([])
const appealTotal = ref(0)
const appealPage = ref(1)
const appealPageSize = ref(10)
const appealStatusFilter = ref('pending')
const loadingAppeals = ref(false)
const currentAppeal = ref(null)
const showAppealDetailDialog = ref(false)
const showProcessAppealDialog = ref(false)
const processingAppeal = ref(false)
const processAppealForm = reactive({
  result: '',
  approve: true
})
const showAdminProcessAffairDialog = ref(false)
const adminProcessAffairForm = reactive({
  note: ''
})

// 处理人员管理
const pendingProcessors = ref([])
const rejectedProcessors = ref([])
const approvedProcessors = ref([])
const loadingProcessors = ref(false)
const showRejectProcessorDialog = ref(false)
const currentRejectProcessorId = ref(null)
const rejectProcessorForm = reactive({
  reason: ''
})

// 敏感词管理
const sensitiveWords = ref([])
const loadingSensitive = ref(false)
const showAddWordDialog = ref(false)
const newWordForm = reactive({
  word: '',
  category: 'other',
  level: 1
})
const autoReviewSetting = reactive({
  mode: 'need_review_images',
  enabled_sensitive: true
})

// 用户管理
const userSubTab = ref('allUsers')
const allUsers = ref([])
const userTotal = ref(0)
const userPage = ref(1)
const userPageSize = ref(10)
const userKeyword = ref('')
const userRoleFilter = ref('')
const userBannedFilter = ref('')
const loadingUsers = ref(false)

const deletedUsers = ref([])
const deletedUserTotal = ref(0)
const deletedUserPage = ref(1)
const deletedUserPageSize = ref(10)
const deletedUserKeyword = ref('')
const loadingDeletedUsers = ref(false)

const showBanDialog = ref(false)
const banning = ref(false)
const banForm = reactive({
  userId: null,
  username: '',
  banType: 'forever',
  days: 7,
  untilDate: ''
})
const showUserInfoDialogVisible = ref(false)
const currentUserInfo = ref(null)

const showEditDeletedUserDialogVisible = ref(false)
const savingDeletedUser = ref(false)
const editDeletedUserForm = reactive({
  id: null,
  username: '',
  nickname: '',
  role: 'normal',
  phone: '',
  email: '',
  signature: ''
})

const getStatusType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || '')
const getStatusText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '')

const getProductTypeText = (type) => {
  const map = { 'seed': '种子', 'fertilizer': '肥料', 'pesticide': '农药', 'tool': '农具', 'other': '其他' }
  return map[type] || type
}

const getPostTypeText = (type) => {
  const map = { 'planting': '种植', 'breeding': '养殖', 'technology': '技术', 'policy': '政策', 'other': '其他' }
  return map[type] || type
}

const getAffairTypeText = (type) => {
  const map = { 'infrastructure': '基础设施', 'hardware': '硬件问题', 'env': '环境问题', 'safety': '安全隐患', 'other': '其他' }
  return map[type] || type
}

const getAffairStatusText = (status) => {
  const map = { 1: '待审核', 2: '审核通过', 3: '审核不通过', 4: '处理中', 5: '已完成' }
  return map[status] || '未知'
}

const getAffairStatusType = (status) => {
  const map = { 1: 'warning', 2: 'primary', 3: 'danger', 4: 'info', 5: 'success' }
  return map[status] || ''
}

const getCategoryText = (cat) => {
  const map = { 'politics': '政治敏感', 'porn': '色情低俗', 'advertising': '广告推广', 'gamble': '赌博博彩', 'fraud': '欺诈诈骗', 'other': '其他违规' }
  return map[cat] || cat
}

const getLevelText = (level) => ({ 1: '轻度', 2: '中度', 3: '重度' }[level] || '')
const getLevelType = (level) => ({ 1: 'info', 2: 'warning', 3: 'danger' }[level] || '')

const formatTime = (row) => row.created_at ? new Date(row.created_at).toLocaleString('zh-CN') : ''
const formatTimeStr = (time) => time ? new Date(time).toLocaleString('zh-CN') : ''

const parseImages = (images) => {
  if (!images) return []
  try {
    return JSON.parse(images)
  } catch {
    return []
  }
}

// 加载待审核商品
const loadPendingProducts = async () => {
  loadingProducts.value = true
  try {
    const res = await api.getPendingProducts()
    pendingProducts.value = res?.list || []
  } catch (error) {
    console.error('加载待审核商品失败:', error)
  } finally {
    loadingProducts.value = false
  }
}

// 加载已退回商品
const loadRejectedProducts = async () => {
  loadingProducts.value = true
  try {
    const res = await api.getRejectedProducts()
    rejectedProducts.value = res?.list || []
  } catch (error) {
    console.error('加载已退回商品失败:', error)
  } finally {
    loadingProducts.value = false
  }
}

// 加载已上架商品
const loadOnShelfProducts = async () => {
  loadingProducts.value = true
  try {
    const res = await api.getProducts({ page: 1, page_size: 100 })
    onShelfProducts.value = (res?.list || []).filter(p => p.status === 1)
  } catch (error) {
    console.error('加载已上架商品失败:', error)
  } finally {
    loadingProducts.value = false
  }
}

// 加载待审核动态
const loadPendingPosts = async () => {
  loadingPosts.value = true
  try {
    const res = await api.getPendingPosts()
    pendingPosts.value = res?.list || []
  } catch (error) {
    console.error('加载待审核动态失败:', error)
  } finally {
    loadingPosts.value = false
  }
}

// 加载已退回动态
const loadRejectedPosts = async () => {
  loadingPosts.value = true
  try {
    const res = await api.getRejectedPosts()
    rejectedPosts.value = res?.list || []
  } catch (error) {
    console.error('加载已退回动态失败:', error)
  } finally {
    loadingPosts.value = false
  }
}

// 加载已发布动态
const loadPublishedPosts = async () => {
  loadingPosts.value = true
  try {
    const res = await api.getPosts({ page: 1, page_size: 100 })
    publishedPosts.value = (res?.list || []).filter(p => p.status === 1)
  } catch (error) {
    console.error('加载已发布动态失败:', error)
  } finally {
    loadingPosts.value = false
  }
}

// 加载已删除动态 (status=2)
const loadDeletedPosts = async () => {
  loadingPosts.value = true
  try {
    const res = await api.getDeletedPosts()
    deletedPosts.value = res?.list || []
  } catch (error) {
    console.error('加载已删除动态失败:', error)
  } finally {
    loadingPosts.value = false
  }
}

// 加载待审核知识
const loadPendingKnowledges = async () => {
  loadingKnowledge.value = true
  try {
    const res = await api.getPendingKnowledge()
    pendingKnowledges.value = res?.list || []
  } catch (error) {
    console.error('加载待审核知识失败:', error)
  } finally {
    loadingKnowledge.value = false
  }
}

// 加载已退回知识
const loadRejectedKnowledges = async () => {
  loadingKnowledge.value = true
  try {
    const res = await api.getRejectedKnowledge()
    rejectedKnowledges.value = res?.list || []
  } catch (error) {
    console.error('加载已退回知识失败:', error)
  } finally {
    loadingKnowledge.value = false
  }
}

// 加载已发布知识
const loadPublishedKnowledges = async () => {
  loadingKnowledge.value = true
  try {
    const res = await api.getKnowledgeList({ page: 1, page_size: 100 })
    publishedKnowledges.value = (res?.data || res?.list || []).filter(k => k.status === 1)
  } catch (error) {
    console.error('加载已发布知识失败:', error)
  } finally {
    loadingKnowledge.value = false
  }
}

// 加载已删除知识 (status=2)
const loadDeletedKnowledges = async () => {
  loadingKnowledge.value = true
  try {
    const res = await api.getDeletedKnowledge()
    deletedKnowledges.value = res?.list || []
  } catch (error) {
    console.error('加载已删除知识失败:', error)
  } finally {
    loadingKnowledge.value = false
  }
}

// 加载专家列表
const loadPendingExperts = async () => {
  loadingExperts.value = true
  try {
    const res = await api.getPendingExperts()
    pendingExperts.value = res?.list || []
  } catch (error) {
    console.error('加载待审核专家失败:', error)
  } finally {
    loadingExperts.value = false
  }
}

// 加载已拒绝专家
const loadRejectedExperts = async () => {
  loadingExperts.value = true
  try {
    const res = await api.getAllExperts()
    rejectedExperts.value = (res?.list || []).filter(e => e.status === 2)
  } catch (error) {
    console.error('加载已拒绝专家失败:', error)
  } finally {
    loadingExperts.value = false
  }
}

// 加载已通过专家
const loadApprovedExperts = async () => {
  loadingExperts.value = true
  try {
    const res = await api.getExperts({ page: 1, page_size: 100 })
    approvedExperts.value = (res?.list || []).filter(e => e.status === 1)
  } catch (error) {
    console.error('加载已通过专家失败:', error)
  } finally {
    loadingExperts.value = false
  }
}

// 加载所有专家列表
const loadAllExperts = async () => {
  loadingExperts.value = true
  try {
    await Promise.all([
      loadPendingExperts(),
      loadRejectedExperts(),
      loadApprovedExperts()
    ])
  } catch (error) {
    console.error('加载专家列表失败:', error)
  } finally {
    loadingExperts.value = false
  }
}

// 事务管理函数
const loadPendingAuditAffairs = async () => {
  loadingAffairs.value = true
  try {
    const res = await api.getPendingAuditAffairs()
    pendingAuditAffairs.value = res?.list || []
  } catch (error) {
    console.error('加载待审核事务失败:', error)
  } finally {
    loadingAffairs.value = false
  }
}

const loadPendingProcessAffairs = async () => {
  loadingAffairs.value = true
  try {
    const res = await api.getPendingProcessAffairs()
    pendingProcessAffairs.value = res?.list || []
  } catch (error) {
    console.error('加载待处理事务失败:', error)
  } finally {
    loadingAffairs.value = false
  }
}

const loadAllAffairs = async () => {
  loadingAffairs.value = true
  try {
    const res = await api.getAllRuralAffairs({ page: 1, page_size: 100 })
    allAffairs.value = res?.list || []
  } catch (error) {
    console.error('加载事务列表失败:', error)
  } finally {
    loadingAffairs.value = false
  }
}

// 加载审核不通过的事务 (status=3)
const loadRejectedAffairs = async () => {
  loadingAffairs.value = true
  try {
    const res = await api.getRejectedAffairs()
    rejectedAffairs.value = res?.list || []
  } catch (error) {
    console.error('加载审核不通过事务失败:', error)
  } finally {
    loadingAffairs.value = false
  }
}

// 加载处理中的事务 (status=4)
const loadProcessingAffairs = async () => {
  loadingAffairs.value = true
  try {
    const res = await api.getProcessingAffairs()
    processingAffairs.value = res?.list || []
  } catch (error) {
    console.error('加载处理中事务失败:', error)
  } finally {
    loadingAffairs.value = false
  }
}

// 加载已完成的事务 (status=5)
const loadCompletedAffairs = async () => {
  loadingAffairs.value = true
  try {
    const res = await api.getCompletedAffairs()
    completedAffairs.value = res?.list || []
  } catch (error) {
    console.error('加载已完成事务失败:', error)
  } finally {
    loadingAffairs.value = false
  }
}

const showAffairDetail = (row) => {
  currentAffair.value = row
  showAffairDialog.value = true
}

const auditAffair = async (id, approved) => {
  try {
    await ElMessageBox.confirm(`确定要${approved ? '通过' : '退回'}该事务吗？`, '提示', { type: 'warning' })
    await api.auditRuralAffair(id, { approved })
    ElMessage.success(approved ? '已通过' : '已退回')
    loadPendingAuditAffairs()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

// ===================== 申诉管理函数 =====================
const getAppealStatusText = (s) => {
  const map = { pending: '待处理', processing: '处理中', resolved: '已处理' }
  return map[s] || s
}
const getAppealStatusTag = (s) => {
  const map = { pending: 'warning', processing: 'info', resolved: 'success' }
  return map[s] || 'info'
}

// 加载申诉列表
const loadAppeals = async () => {
  loadingAppeals.value = true
  try {
    const params = {
      page: appealPage.value,
      page_size: appealPageSize.value
    }
    if (appealStatusFilter.value) {
      params.status = appealStatusFilter.value
    }
    const res = await api.adminGetAppeals(params)
    appeals.value = res?.list || []
    appealTotal.value = res?.total || 0
  } catch (error) {
    console.error('加载申诉列表失败:', error)
  } finally {
    loadingAppeals.value = false
  }
}

// 查看申诉详情
const openAppealDetail = (appeal) => {
  currentAppeal.value = appeal
  showAppealDetailDialog.value = true
}

// 打开处理申诉对话框
const openProcessAppealDialog = (appeal) => {
  currentAppeal.value = appeal
  processAppealForm.result = ''
  processAppealForm.approve = true
  showProcessAppealDialog.value = true
}

// 处理申诉
const handleProcessAppeal = async () => {
  if (!processAppealForm.result.trim()) {
    ElMessage.warning('请填写处理结果')
    return
  }
  processingAppeal.value = true
  try {
    await api.adminProcessAppeal(currentAppeal.value.id, {
      result: processAppealForm.result,
      approve: processAppealForm.approve
    })
    ElMessage.success('申诉已处理')
    showProcessAppealDialog.value = false
    loadAppeals()
  } catch (e) {
    ElMessage.error(e?.response?.data?.error || '处理失败')
  } finally {
    processingAppeal.value = false
  }
}

// 管理员直接处理事务
const openAdminProcessAffairDialog = (affair) => {
  currentAffair.value = affair
  adminProcessAffairForm.note = ''
  showAdminProcessAffairDialog.value = true
}

const handleAdminProcessAffair = async () => {
  if (!adminProcessAffairForm.note.trim()) {
    ElMessage.warning('请填写处理备注')
    return
  }
  processingAppeal.value = true
  try {
    // 管理员标记介入处理，但仍需用户确认完成
    await api.adminProcessAppeal(currentAffair.value.id, {
      result: '管理员介入处理：' + adminProcessAffairForm.note,
      approve: true
    })
    ElMessage.success('已记录管理员处理备注，等待用户确认完成')
    showAdminProcessAffairDialog.value = false
  } catch (e) {
    ElMessage.error(e?.response?.data?.error || '处理失败')
  } finally {
    processingAppeal.value = false
  }
}

// 处理人员管理函数
const loadPendingProcessors = async () => {
  loadingProcessors.value = true
  try {
    const res = await api.getPendingProcessors()
    pendingProcessors.value = res?.list || []
  } catch (error) {
    console.error('加载待审核处理人员失败:', error)
  } finally {
    loadingProcessors.value = false
  }
}

const loadRejectedProcessors = async () => {
  loadingProcessors.value = true
  try {
    const res = await api.getAllProcessors()
    rejectedProcessors.value = (res?.list || []).filter(p => p.status === 2)
  } catch (error) {
    console.error('加载已拒绝处理人员失败:', error)
  } finally {
    loadingProcessors.value = false
  }
}

const loadApprovedProcessors = async () => {
  loadingProcessors.value = true
  try {
    const res = await api.getAllProcessors()
    approvedProcessors.value = (res?.list || []).filter(p => p.status === 1)
  } catch (error) {
    console.error('加载已通过处理人员失败:', error)
  } finally {
    loadingProcessors.value = false
  }
}

const loadAllProcessors = async () => {
  loadingProcessors.value = true
  try {
    await Promise.all([
      loadPendingProcessors(),
      loadRejectedProcessors(),
      loadApprovedProcessors()
    ])
  } catch (error) {
    console.error('加载处理人员列表失败:', error)
  } finally {
    loadingProcessors.value = false
  }
}

const approveProcessor = async (id) => {
  try {
    await api.approveProcessor(id)
    ElMessage.success('已通过审核')
    loadAllProcessors()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

const rejectProcessor = async (id) => {
  currentRejectProcessorId.value = id
  rejectProcessorForm.reason = ''
  showRejectProcessorDialog.value = true
}

const confirmRejectProcessor = async () => {
  if (!rejectProcessorForm.reason) {
    ElMessage.warning('请输入拒绝原因')
    return
  }
  try {
    await api.rejectProcessor(currentRejectProcessorId.value, { reason: rejectProcessorForm.reason })
    ElMessage.success('已拒绝')
    showRejectProcessorDialog.value = false
    loadAllProcessors()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

const reApproveProcessor = async (id) => {
  try {
    await ElMessageBox.confirm('确定要重新审核该申请吗？', '提示', { type: 'warning' })
    await api.approveProcessor(id)
    ElMessage.success('已重新通过')
    loadAllProcessors()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

const disableProcessor = async (id) => {
  try {
    await ElMessageBox.confirm('确定要禁用该处理人员吗？', '提示', { type: 'warning' })
    await api.disableProcessor(id)
    ElMessage.success('已禁用')
    loadAllProcessors()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

// 审核通过专家
const approveExpert = async (id) => {
  try {
    await api.approveExpert(id)
    ElMessage.success('已通过审核')
    loadAllExperts()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

// 审核拒绝专家
const rejectExpert = async (id) => {
  try {
    await ElMessageBox.confirm('确定要拒绝该专家申请吗？', '提示', { type: 'warning' })
    await api.rejectExpert(id)
    ElMessage.success('已拒绝')
    loadAllExperts()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

// 重新审核专家（从拒绝状态恢复）
const reApproveExpert = async (id) => {
  try {
    await ElMessageBox.confirm('确定要重新审核该专家申请吗？', '提示', { type: 'warning' })
    await api.approveExpert(id)
    ElMessage.success('已重新通过审核')
    loadAllExperts()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

// 禁用专家
const disableExpert = async (id) => {
  try {
    await ElMessageBox.confirm('确定要禁用该专家吗？', '提示', { type: 'warning' })
    await api.rejectExpert(id)
    ElMessage.success('已禁用')
    loadAllExperts()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

// 通过商品审核
const passProduct = async (id) => {
  try {
    await api.auditProduct(id, { approved: true })
    ElMessage.success('已通过审核')
    loadPendingProducts()
    loadOnShelfProducts()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

// 退回商品审核
const rejectProduct = async (id) => {
  try {
    await ElMessageBox.confirm('确定要退回该商品吗？', '提示', { type: 'warning' })
    await api.auditProduct(id, { approved: false })
    ElMessage.success('已退回')
    loadPendingProducts()
    loadRejectedProducts()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

// 重新提交商品审核
const reOnShelfProduct = async (id) => {
  try {
    await ElMessageBox.confirm('确定要重新提交该商品审核吗？', '提示', { type: 'warning' })
    await api.restoreProduct(id)
    ElMessage.success('已重新提交审核')
    loadRejectedProducts()
    loadPendingProducts()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

// 下架商品
const offShelfProduct = async (id) => {
  try {
    await ElMessageBox.confirm('确定要退回该商品吗？', '提示', { type: 'warning' })
    await api.offShelfProduct(id)
    ElMessage.success('已退回')
    loadOnShelfProducts()
    loadRejectedProducts()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

// 通过动态审核
const passPost = async (id) => {
  try {
    await api.auditPost(id, { approved: true })
    ElMessage.success('已通过审核')
    loadPendingPosts()
    loadPublishedPosts()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

// 退回动态审核
const rejectPost = async (id) => {
  try {
    await ElMessageBox.confirm('确定要退回该动态吗？', '提示', { type: 'warning' })
    await api.auditPost(id, { approved: false })
    ElMessage.success('已退回')
    loadPendingPosts()
    loadRejectedPosts()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

// 编辑动态
const editPost = (row) => {
  editPostForm.id = row.id
  editPostForm.title = row.title
  editPostForm.content = row.content
  editPostForm.type = row.type
  editPostForm.images = row.images || ''
  showEditPostDialog.value = true
}

// 保存动态
const handleUpdatePost = async () => {
  if (!editPostForm.title || !editPostForm.content) {
    ElMessage.warning('请填写完整信息')
    return
  }
  savingPost.value = true
  try {
    await api.updatePost(editPostForm.id, editPostForm)
    ElMessage.success('保存成功')
    showEditPostDialog.value = false
    loadPublishedPosts()
  } catch (error) {
    ElMessage.error('保存失败')
  } finally {
    savingPost.value = false
  }
}

// 恢复已退回动态
const restorePost = async (id) => {
  try {
    await ElMessageBox.confirm('确定要恢复该动态吗？', '提示', { type: 'warning' })
    await api.restorePost(id)
    ElMessage.success('已恢复')
    loadRejectedPosts()
    loadPendingPosts()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

// 通过知识审核
const passKnowledge = async (id) => {
  try {
    await api.auditKnowledge(id, { approved: true })
    ElMessage.success('已通过审核')
    loadPendingKnowledges()
    loadPublishedKnowledges()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

// 退回知识审核
const rejectKnowledge = async (id) => {
  try {
    await ElMessageBox.confirm('确定要退回该知识吗？', '提示', { type: 'warning' })
    await api.auditKnowledge(id, { approved: false })
    ElMessage.success('已退回')
    loadPendingKnowledges()
    loadRejectedKnowledges()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

// 恢复已退回知识
const restoreKnowledge = async (id) => {
  try {
    await ElMessageBox.confirm('确定要恢复该知识吗？', '提示', { type: 'warning' })
    await api.restoreKnowledge(id)
    ElMessage.success('已恢复')
    loadRejectedKnowledges()
    loadPendingKnowledges()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

// 查看商品详情
const showProductDetail = async (row) => {
  currentProduct.value = row
  showProductDialog.value = true
  loadingComments.value = true
  try {
    const res = await api.getProductComments(row.id)
    // 后端已返回树形结构
    productComments.value = Array.isArray(res) ? res : []
  } catch (error) {
    productComments.value = []
  } finally {
    loadingComments.value = false
  }
}

// 编辑商品
const editProduct = (row) => {
  editProductForm.id = row.id
  editProductForm.title = row.title
  editProductForm.type = row.type
  editProductForm.price = row.price
  editProductForm.content = row.content
  editProductForm.stock = row.stock
  editProductForm.status = row.status
  showEditProductDialog.value = true
}

// 保存商品
const handleUpdateProduct = async () => {
  if (!editProductForm.title || !editProductForm.content) {
    ElMessage.warning('请填写完整信息')
    return
  }
  savingProduct.value = true
  try {
    await api.updateProduct(editProductForm)
    ElMessage.success('保存成功')
    showEditProductDialog.value = false
    loadOnShelfProducts()
  } catch (error) {
    ElMessage.error('保存失败')
  } finally {
    savingProduct.value = false
  }
}

// 查看动态详情
const showPostDetail = async (row) => {
  currentPost.value = row
  showPostDialog.value = true
  loadingPostComments.value = true
  try {
    const res = await api.getPostComments(row.id)
    // 后端已返回树形结构
    postComments.value = Array.isArray(res) ? res : []
  } catch (error) {
    postComments.value = []
  } finally {
    loadingPostComments.value = false
  }
}

// 退回已发布动态 (设置 status=0)
const rejectPublishedPost = async (id) => {
  try {
    await ElMessageBox.confirm('确定要退回该动态吗？', '提示', { type: 'warning' })
    await api.auditPost(id, { approved: false })
    ElMessage.success('已退回')
    loadPublishedPosts()
    loadRejectedPosts()
  } catch (error) {
    if (error !== 'cancel') {
      const msg = error?.response?.data?.error || error?.message || '操作失败'
      ElMessage.error(msg)
    }
  }
}

// 查看知识详情
const showKnowledgeDetail = async (row) => {
  currentKnowledge.value = row
  showKnowledgeDetailDialog.value = true
  loadingKnowledgeComments.value = true
  try {
    const res = await api.getKnowledgeComments(row.id)
    // 后端已返回树形结构
    knowledgeComments.value = Array.isArray(res) ? res : []
  } catch (error) {
    knowledgeComments.value = []
  } finally {
    loadingKnowledgeComments.value = false
  }
}

// 编辑知识
const editKnowledge = (row) => {
  editKnowledgeForm.id = row.id
  editKnowledgeForm.title = row.title
  editKnowledgeForm.content = row.content
  showEditKnowledgeDialog.value = true
}

// 保存知识
const handleUpdateKnowledge = async () => {
  if (!editKnowledgeForm.title || !editKnowledgeForm.content) {
    ElMessage.warning('请填写完整信息')
    return
  }
  savingKnowledge.value = true
  try {
    await api.updateKnowledge(editKnowledgeForm.id, editKnowledgeForm)
    ElMessage.success('保存成功')
    showEditKnowledgeDialog.value = false
    loadPublishedKnowledges()
  } catch (error) {
    ElMessage.error('保存失败')
  } finally {
    savingKnowledge.value = false
  }
}

// 退回已发布知识 (设置 status=0)
const rejectPublishedKnowledge = async (id) => {
  try {
    await ElMessageBox.confirm('确定要退回该知识吗？', '提示', { type: 'warning' })
    await api.auditKnowledge(id, { approved: false })
    ElMessage.success('已退回')
    loadPublishedKnowledges()
    loadRejectedKnowledges()
  } catch (error) {
    if (error !== 'cancel') {
      const msg = error?.response?.data?.error || error?.message || '操作失败'
      ElMessage.error(msg)
    }
  }
}

// 加载敏感词列表
const loadSensitiveWords = async () => {
  loadingSensitive.value = true
  try {
    const res = await api.getSensitiveWords()
    sensitiveWords.value = res?.list || []
  } catch (error) {
    console.error('加载敏感词失败:', error)
  } finally {
    loadingSensitive.value = false
  }
}

// 添加敏感词
const addSensitiveWord = async () => {
  if (!newWordForm.word) {
    ElMessage.warning('请输入敏感词')
    return
  }
  try {
    await api.addSensitiveWord(newWordForm)
    ElMessage.success('添加成功')
    showAddWordDialog.value = false
    newWordForm.word = ''
    newWordForm.category = 'other'
    newWordForm.level = 1
    loadSensitiveWords()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '添加失败'
    ElMessage.error(msg)
  }
}

// 删除敏感词
const deleteSensitiveWord = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除该敏感词吗？', '提示', { type: 'warning' })
    await api.deleteSensitiveWord({ id })
    ElMessage.success('删除成功')
    loadSensitiveWords()
  } catch (error) {
    if (error !== 'cancel') {
      const msg = error?.response?.data?.error || error?.message || '删除失败'
      ElMessage.error(msg)
    }
  }
}

// 加载自动审核设置
const loadAutoReviewSetting = async () => {
  try {
    const res = await api.getAutoReviewSetting()
    if (res) {
      autoReviewSetting.mode = res.mode || 'need_review_images'
      autoReviewSetting.enabled_sensitive = res.enabled_sensitive !== false
    }
  } catch (error) {
    console.error('加载自动审核设置失败:', error)
  }
}

// 保存自动审核设置
const saveAutoReviewSetting = async () => {
  try {
    await api.setAutoReviewSetting(autoReviewSetting)
    ElMessage.success('设置已保存')
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '保存失败'
    ElMessage.error(msg)
  }
}

// 角色相关
const getRoleType = (role) => {
  const map = { normal: 'info', farmer: 'warning', expert: 'success', processor: '', admin: 'danger', sysadmin: 'danger' }
  return map[role] || 'info'
}
const getRoleText = (role) => {
  const map = { normal: '普通用户', farmer: '农户', expert: '专家', processor: '处理人员', admin: '审核员', sysadmin: '系统管理员' }
  return map[role] || role
}

// 格式化短时间
const formatShortTime = (time) => {
  if (!time) return ''
  return new Date(time).toLocaleDateString('zh-CN')
}

// 封禁截止日期计算
const computedBanEndDate = computed(() => {
  if (banForm.banType !== 'duration') return ''
  const end = new Date()
  end.setDate(end.getDate() + banForm.days)
  return end.toLocaleString('zh-CN')
})

// 禁用过去的日期
const disabledDate = (time) => {
  return time.getTime() < Date.now() - 8.64e7 // 不能选择今天之前
}

const onBanTypeChange = () => {
  banForm.untilDate = ''
}

// 加载全部用户
const loadAllUsers = async () => {
  loadingUsers.value = true
  try {
    const params = {
      page: userPage.value,
      page_size: userPageSize.value,
      keyword: userKeyword.value || undefined,
      role: userRoleFilter.value || undefined,
      banned: userBannedFilter.value || undefined
    }
    const res = await api.getAllUsersForAdmin(params)
    allUsers.value = res?.list || []
    userTotal.value = res?.total || 0
  } catch (error) {
    console.error('加载用户列表失败:', error)
  } finally {
    loadingUsers.value = false
  }
}

// 加载已注销用户
const loadDeletedUsers = async () => {
  loadingDeletedUsers.value = true
  try {
    const params = {
      page: deletedUserPage.value,
      page_size: deletedUserPageSize.value,
      keyword: deletedUserKeyword.value || undefined
    }
    const res = await api.getDeletedUsers(params)
    deletedUsers.value = res?.list || []
    deletedUserTotal.value = res?.total || 0
  } catch (error) {
    console.error('加载已注销用户列表失败:', error)
  } finally {
    loadingDeletedUsers.value = false
  }
}

// 打开封禁对话框
const openBanDialog = (user) => {
  banForm.userId = user.id
  banForm.username = user.username
  banForm.banType = 'forever'
  banForm.days = 7
  banForm.untilDate = ''
  showBanDialog.value = true
}

// 确认封禁
const confirmBanUser = async () => {
  banning.value = true
  try {
    let bannedUntil = null
    if (banForm.banType === 'duration') {
      const end = new Date()
      end.setDate(end.getDate() + banForm.days)
      bannedUntil = end.toISOString().slice(0, 19).replace('T', ' ')
    } else if (banForm.banType === 'until') {
      if (!banForm.untilDate) {
        ElMessage.warning('请选择封禁截止时间')
        banning.value = false
        return
      }
      bannedUntil = banForm.untilDate
    }
    await api.banUser(banForm.userId, { banned_until: bannedUntil })
    ElMessage.success('封禁成功')
    showBanDialog.value = false
    loadAllUsers()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '封禁失败'
    ElMessage.error(msg)
  } finally {
    banning.value = false
  }
}

// 解封用户
const handleUnbanUser = async (id) => {
  try {
    await ElMessageBox.confirm('确认解除对该用户的封禁？', '解封确认', { type: 'warning' })
    await api.unbanUser(id)
    ElMessage.success('解封成功')
    loadAllUsers()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('解封失败')
    }
  }
}

// 查看用户信息
const showUserInfoDialog = (user) => {
  currentUserInfo.value = { ...user }
  showUserInfoDialogVisible.value = true
}

// 打开编辑已注销用户对话框
const openEditDeletedUserDialog = (user) => {
  editDeletedUserForm.id = user.id
  editDeletedUserForm.username = user.username
  editDeletedUserForm.nickname = user.nickname || ''
  editDeletedUserForm.role = user.role || 'normal'
  editDeletedUserForm.phone = user.phone || ''
  editDeletedUserForm.email = user.email || ''
  editDeletedUserForm.signature = user.signature || ''
  showEditDeletedUserDialogVisible.value = true
}

// 确认编辑已注销用户
const confirmEditDeletedUser = async () => {
  savingDeletedUser.value = true
  try {
    await api.updateDeletedUser(editDeletedUserForm.id, {
      nickname: editDeletedUserForm.nickname,
      role: editDeletedUserForm.role,
      phone: editDeletedUserForm.phone,
      email: editDeletedUserForm.email,
      signature: editDeletedUserForm.signature
    })
    ElMessage.success('保存成功')
    showEditDeletedUserDialogVisible.value = false
    loadDeletedUsers()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '保存失败'
    ElMessage.error(msg)
  } finally {
    savingDeletedUser.value = false
  }
}

// 退出登录
const handleLogout = () => {
  removeToken()
  removeUserInfo()
  ElMessage.success('已退出登录')
  router.push('/')
}

onMounted(() => {
  // 初始化用户信息
  if (user.value) {
    const avatar_url = user.value.avatar_url?.startsWith('/') ? user.value.avatar_url : '/' + (user.value.avatar_url || '')
    currentUser.value = { ...user.value, avatar_url }

    // 只允许 sysadmin 访问系统管理端
    if (user.value.role !== 'sysadmin') {
      ElMessage.error('您没有权限访问系统管理端')
      router.push('/')
      return
    }
  } else {
    ElMessage.error('请先登录')
    router.push('/login')
    return
  }
  // 加载商品管理数据
  loadPendingProducts()
  loadRejectedProducts()
  loadOnShelfProducts()
  // 加载动态管理数据
  loadPendingPosts()
  loadRejectedPosts()
  loadPublishedPosts()
  loadDeletedPosts()
  // 加载知识管理数据
  loadPendingKnowledges()
  loadRejectedKnowledges()
  loadPublishedKnowledges()
  loadDeletedKnowledges()
  // 加载专家管理数据
  loadAllExperts()
  // 加载事务管理数据
  loadPendingAuditAffairs()
  loadPendingProcessAffairs()
  loadRejectedAffairs()
  loadProcessingAffairs()
  loadCompletedAffairs()
  // 加载处理人员管理数据
  loadAllProcessors()
  // 加载敏感词管理
  loadSensitiveWords()
  loadAutoReviewSetting()
  // 加载用户管理数据
  loadAllUsers()
  // 加载申诉管理数据
  loadAppeals()
})

// 监听用户管理子标签切换，按需加载数据
watch(userSubTab, (tab) => {
  if (tab === 'allUsers' && allUsers.value.length === 0) {
    loadAllUsers()
  } else if (tab === 'deletedUsers' && deletedUsers.value.length === 0) {
    loadDeletedUsers()
  }
})

// 监听主标签切换，按需加载数据
watch(activeTab, (tab) => {
  if (tab === 'appeals' && appeals.value.length === 0) {
    loadAppeals()
  }
  if (tab === 'users' && allUsers.value.length === 0) {
    loadAllUsers()
  }
})
</script>

<style scoped>
.admin-container { min-height: 100vh; background: #f5f5f5; }
.header { background: #fff; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
.container { max-width: 1200px; margin: 0 auto; padding: 0 20px; }
.header .container { display: flex; align-items: center; height: 60px; }
.logo { font-size: 20px; font-weight: bold; color: #f56c6c; text-decoration: none; }
.nav { flex: 1; margin-left: 40px; }
.nav a { margin: 0 15px; color: #666; text-decoration: none; font-size: 15px; }
.nav a:hover, .nav a.router-link-active { color: #409eff; }

.main { padding: 20px; }

.tab-header { display: flex; gap: 10px; margin-bottom: 15px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }

.empty-state { text-align: center; padding: 40px; color: #999; }

/* 评论区样式 */
.comments-section { max-height: 400px; overflow-y: auto; }
.comment-list { display: flex; flex-direction: column; gap: 15px; }
.comment-item { border-bottom: 1px solid #eee; padding-bottom: 15px; }
.comment-header { display: flex; align-items: center; gap: 10px; margin-bottom: 8px; }
.comment-username { font-weight: 500; color: #333; }
.comment-time { color: #999; font-size: 12px; }
.comment-content { color: #666; line-height: 1.6; margin-left: 42px; }
.comment-actions { margin-left: 42px; margin-top: 8px; }

.reply-list { margin-left: 42px; margin-top: 10px; border-left: 2px solid #eee; padding-left: 15px; }
.reply-item { margin-bottom: 10px; }
.reply-header { display: flex; align-items: center; gap: 8px; margin-bottom: 5px; }
.reply-username { font-size: 13px; font-weight: 500; color: #333; }
.reply-to { color: #409eff; font-size: 12px; }
.reply-time { color: #999; font-size: 11px; }
.reply-content { color: #666; font-size: 13px; line-height: 1.5; margin-left: 32px; }

.like-btn { display: inline-flex; align-items: center; gap: 4px; color: #999; font-size: 13px; cursor: pointer; }
.like-btn:hover { color: #409eff; }

.post-images { display: flex; flex-wrap: wrap; gap: 10px; }

.images-grid { display: flex; flex-wrap: wrap; gap: 10px; }
.detail-img { width: 120px; height: 120px; border-radius: 8px; }
.affair-detail { padding: 10px 0; }
</style>
