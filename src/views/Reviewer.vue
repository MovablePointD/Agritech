<template>
  <div class="reviewer-container">
    <header class="header">
      <div class="container">
        <router-link to="/" class="logo">农技知识分享平台 - 审核中心</router-link>
        <nav class="nav">
          <router-link to="/">首页</router-link>
          <router-link to="/user">个人中心</router-link>
        </nav>
        <div class="user-area">
          <span class="user-info">
            <el-avatar :src="currentUser.avatar_url" :size="32">{{ currentUser.username?.[0] }}</el-avatar>
            <span>{{ currentUser.nickname || currentUser.username }}</span>
            <el-tag type="warning" size="small" style="margin-left: 4px;">审核员</el-tag>
          </span>
          <el-button size="small" @click="handleLogout">退出</el-button>
        </div>
      </div>
    </header>

    <div class="container main">
      <el-tabs v-model="activeTab" type="border-card">
        <!-- 专家申请审核 -->
        <el-tab-pane label="专家申请" name="experts">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-card>
                <template #header>
                  <span class="card-title">待审核 ({{ pendingExperts.length }})</span>
                  <el-button size="small" @click="loadPendingExperts" style="float: right;">刷新</el-button>
                </template>
                <el-table :data="pendingExperts" v-loading="loadingExperts" height="400" empty-text="暂无待审核申请">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="real_name" label="真实姓名" width="80" />
                  <el-table-column prop="profession" label="专业" width="80">
                    <template #default="{ row }">
                      <el-tag size="small">{{ row.profession }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="company" label="单位" min-width="100" show-overflow-tooltip />
                  <el-table-column label="操作" width="120" fixed="right">
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
                  <el-button size="small" @click="loadRejectedExperts" style="float: right;">刷新</el-button>
                </template>
                <el-table :data="rejectedExperts" v-loading="loadingExperts" height="400" empty-text="暂无已拒绝申请">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="real_name" label="真实姓名" width="80" />
                  <el-table-column prop="profession" label="专业" width="80">
                    <template #default="{ row }">
                      <el-tag size="small">{{ row.profession }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="company" label="单位" min-width="100" show-overflow-tooltip />
                  <el-table-column label="操作" width="80" fixed="right">
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
              <span>已通过专家 ({{ approvedExperts.length }})</span>
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
              <el-table-column label="操作" width="100" fixed="right">
                <template #default="{ row }">
                  <el-button size="small" type="danger" @click="removeExpert(row.id)">移除</el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-tab-pane>

        <!-- 商品审核 -->
        <el-tab-pane label="商品审核" name="products">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-card>
                <template #header>
                  <span class="card-title">待审核 ({{ pendingProducts.length }})</span>
                  <el-button size="small" @click="loadPendingProducts" style="float: right;">刷新</el-button>
                </template>
                <el-table :data="pendingProducts" v-loading="loadingAllProducts" height="400" empty-text="暂无待审核商品">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="title" label="商品名称" min-width="120" show-overflow-tooltip />
                  <el-table-column prop="price" label="价格" width="80">
                    <template #default="{ row }">¥{{ row.price }}</template>
                  </el-table-column>
                  <el-table-column prop="publisher_user?.username" label="发布者" width="80" />
                  <el-table-column label="操作" width="150" fixed="right">
                    <template #default="{ row }">
                      <el-button size="small" @click="showProductDetail(row)">详情</el-button>
                      <el-button size="small" type="success" @click="passProduct(row.id)">通过</el-button>
                      <el-button size="small" type="danger" @click="rejectProduct(row.id)">驳回</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
            <el-col :span="12">
              <el-card>
                <template #header>
                  <span class="card-title">已下架 ({{ offShelfProducts.length }})</span>
                  <el-button size="small" @click="loadOffShelfProducts" style="float: right;">刷新</el-button>
                </template>
                <el-table :data="offShelfProducts" v-loading="loadingAllProducts" height="400" empty-text="暂无已驳回商品">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="title" label="商品名称" min-width="120" show-overflow-tooltip />
                  <el-table-column prop="price" label="价格" width="80">
                    <template #default="{ row }">¥{{ row.price }}</template>
                  </el-table-column>
                  <el-table-column prop="publisher_user?.username" label="发布者" width="80" />
                  <el-table-column label="操作" width="100" fixed="right">
                    <template #default="{ row }">
                      <el-button size="small" @click="showProductDetail(row)">详情</el-button>
                      <el-button size="small" type="success" @click="reOnShelfProduct(row.id)">重审</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
          </el-row>
          <el-card style="margin-top: 20px;">
            <template #header>
              <span>已上架商品 ({{ onShelfProducts.length }})</span>
              <el-button size="small" @click="loadOnShelfProducts" style="float: right;">刷新</el-button>
            </template>
            <el-table :data="onShelfProducts" v-loading="loadingAllProducts" empty-text="暂无上架商品">
              <el-table-column prop="id" label="ID" width="60" />
              <el-table-column prop="title" label="商品名称" min-width="150" show-overflow-tooltip />
              <el-table-column prop="price" label="价格" width="100">
                <template #default="{ row }">¥{{ row.price }}</template>
              </el-table-column>
              <el-table-column prop="type" label="类型" width="100" />
              <el-table-column prop="publisher_user?.username" label="发布者" width="100" />
              <el-table-column prop="created_at" label="发布时间" width="160" :formatter="formatTime" />
              <el-table-column label="操作" width="120" fixed="right">
                <template #default="{ row }">
                  <el-button size="small" @click="showProductDetail(row)">详情</el-button>
                  <el-button size="small" type="danger" @click="offShelfProduct(row.id)">下架</el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-tab-pane>

        <!-- 动态审核 -->
        <el-tab-pane label="动态审核" name="posts">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-card>
                <template #header>
                  <span class="card-title">待审核 ({{ pendingPosts.length }})</span>
                  <el-button size="small" @click="loadPendingPosts" style="float: right;">刷新</el-button>
                </template>
                <el-table :data="pendingPosts" v-loading="loadingAllPosts" height="400" empty-text="暂无待审核动态">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="title" label="标题" min-width="120" show-overflow-tooltip />
                  <el-table-column prop="type" label="类型" width="70" />
                  <el-table-column prop="user?.username" label="发布者" width="80" />
                  <el-table-column label="操作" width="150" fixed="right">
                    <template #default="{ row }">
                      <el-button size="small" @click="showPostDetail(row)">详情</el-button>
                      <el-button size="small" type="success" @click="passPostReview(row.id)">通过</el-button>
                      <el-button size="small" type="danger" @click="rejectPostReview(row.id)">驳回</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
            <el-col :span="12">
              <el-card>
                <template #header>
                  <span class="card-title">已退回 ({{ rejectedPosts.length }})</span>
                  <el-button size="small" @click="loadRejectedPosts" style="float: right;">刷新</el-button>
                </template>
                <el-table :data="rejectedPosts" v-loading="loadingAllPosts" height="400" empty-text="暂无已退回动态">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="title" label="标题" min-width="120" show-overflow-tooltip />
                  <el-table-column prop="type" label="类型" width="70" />
                  <el-table-column prop="user?.username" label="发布者" width="80" />
                  <el-table-column label="操作" width="100" fixed="right">
                    <template #default="{ row }">
                      <el-button size="small" @click="showPostDetail(row)">详情</el-button>
                      <el-button size="small" type="success" @click="restorePost(row.id)">重审</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
          </el-row>
          <el-card style="margin-top: 20px;">
            <template #header>
              <span>已发布动态 ({{ publishedPosts.length }})</span>
              <el-button size="small" @click="loadPublishedPosts" style="float: right;">刷新</el-button>
            </template>
            <el-table :data="publishedPosts" v-loading="loadingAllPosts" empty-text="暂无已发布动态">
              <el-table-column prop="id" label="ID" width="60" />
              <el-table-column prop="title" label="标题" min-width="150" show-overflow-tooltip />
              <el-table-column prop="type" label="类型" width="100" />
              <el-table-column prop="user?.username" label="发布者" width="100" />
              <el-table-column prop="likes" label="点赞" width="80" />
              <el-table-column prop="created_at" label="发布时间" width="160" :formatter="formatTime" />
              <el-table-column label="操作" width="120" fixed="right">
                <template #default="{ row }">
                  <el-button size="small" @click="showPostDetail(row)">详情</el-button>
                  <el-button size="small" type="danger" @click="deletePost(row.id)">退回</el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-tab-pane>

        <!-- 知识审核 -->
        <el-tab-pane label="知识审核" name="knowledge">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-card>
                <template #header>
                  <span class="card-title">待审核 ({{ pendingKnowledge.length }})</span>
                  <el-button size="small" @click="loadPendingKnowledge" style="float: right;">刷新</el-button>
                </template>
                <el-table :data="pendingKnowledge" v-loading="loadingAllKnowledge" height="400" empty-text="暂无待审核知识">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="title" label="标题" min-width="150" show-overflow-tooltip />
                  <el-table-column prop="user?.username" label="发布者" width="100" />
                  <el-table-column label="操作" width="150" fixed="right">
                    <template #default="{ row }">
                      <el-button size="small" @click="showKnowledgeDetail(row)">详情</el-button>
                      <el-button size="small" type="success" @click="passKnowledgeReview(row.id)">通过</el-button>
                      <el-button size="small" type="danger" @click="rejectKnowledgeReview(row.id)">驳回</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
            <el-col :span="12">
              <el-card>
                <template #header>
                  <span class="card-title">已退回 ({{ rejectedKnowledges.length }})</span>
                  <el-button size="small" @click="loadRejectedKnowledge" style="float: right;">刷新</el-button>
                </template>
                <el-table :data="rejectedKnowledges" v-loading="loadingAllKnowledge" height="400" empty-text="暂无已退回知识">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="title" label="标题" min-width="150" show-overflow-tooltip />
                  <el-table-column prop="user?.username" label="发布者" width="100" />
                  <el-table-column label="操作" width="100" fixed="right">
                    <template #default="{ row }">
                      <el-button size="small" @click="showKnowledgeDetail(row)">详情</el-button>
                      <el-button size="small" type="success" @click="restoreKnowledge(row.id)">恢复</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
          </el-row>
          <el-card style="margin-top: 20px;">
            <template #header>
              <span>已发布知识 ({{ publishedKnowledge.length }})</span>
              <el-button size="small" @click="loadPublishedKnowledge" style="float: right;">刷新</el-button>
            </template>
            <el-table :data="publishedKnowledge" v-loading="loadingAllKnowledge" empty-text="暂无已发布知识">
              <el-table-column prop="id" label="ID" width="60" />
              <el-table-column prop="title" label="标题" min-width="150" show-overflow-tooltip />
              <el-table-column prop="user?.username" label="发布者" width="100" />
              <el-table-column prop="created_at" label="发布时间" width="160" :formatter="formatTime" />
              <el-table-column label="操作" width="150" fixed="right">
                <template #default="{ row }">
                  <el-button size="small" @click="showKnowledgeDetail(row)">详情</el-button>
                  <el-button size="small" type="danger" @click="deleteKnowledge(row.id)">退回</el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-tab-pane>

        <!-- 处理人员管理 -->
        <el-tab-pane label="处理人员" name="processors">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-card>
                <template #header>
                  <span class="card-title">待审核 ({{ pendingProcessors.length }})</span>
                  <el-button size="small" @click="loadPendingProcessors" style="float: right;">刷新</el-button>
                </template>
                <el-table :data="pendingProcessors" v-loading="loadingProcessors" height="400" empty-text="暂无待审核申请">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="real_name" label="真实姓名" width="80" />
                  <el-table-column prop="phone" label="联系电话" width="110" />
                  <el-table-column prop="address" label="负责区域" min-width="100" show-overflow-tooltip />
                  <el-table-column label="操作" width="120" fixed="right">
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
                  <el-button size="small" @click="loadRejectedProcessors" style="float: right;">刷新</el-button>
                </template>
                <el-table :data="rejectedProcessors" v-loading="loadingProcessors" height="400" empty-text="暂无已拒绝申请">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="real_name" label="真实姓名" width="80" />
                  <el-table-column prop="phone" label="联系电话" width="110" />
                  <el-table-column prop="address" label="负责区域" min-width="100" show-overflow-tooltip />
                  <el-table-column label="操作" width="80" fixed="right">
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
              <span>已通过处理人员 ({{ approvedProcessors.length }})</span>
              <el-button size="small" @click="loadApprovedProcessors" style="float: right;">刷新</el-button>
            </template>
            <el-table :data="approvedProcessors" v-loading="loadingProcessors" empty-text="暂无已通过处理人员">
              <el-table-column prop="id" label="ID" width="60" />
              <el-table-column prop="real_name" label="真实姓名" width="100" />
              <el-table-column prop="phone" label="联系电话" width="120" />
              <el-table-column prop="address" label="负责区域" min-width="150" show-overflow-tooltip />
              <el-table-column prop="user?.username" label="申请人" width="100" />
              <el-table-column prop="audit_time" label="审核时间" width="160" :formatter="formatTime" />
              <el-table-column label="操作" width="120" fixed="right">
                <template #default="{ row }">
                  <el-button size="small" type="danger" @click="disableProcessor(row.id)">禁用</el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-tab-pane>

        <!-- 事务审核 -->
        <el-tab-pane label="事务审核" name="affairs">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-card>
                <template #header>
                  <span class="card-title">待审核 ({{ pendingAffairs.length }})</span>
                  <el-button size="small" @click="loadPendingAffairs" style="float: right;">刷新</el-button>
                </template>
                <el-table :data="pendingAffairs" v-loading="loadingAffairs" height="350" empty-text="暂无待审核事务">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="title" label="事务标题" min-width="120" show-overflow-tooltip />
                  <el-table-column prop="type" label="类型" width="80">
                    <template #default="{ row }">
                      <el-tag size="small">{{ getAffairTypeText(row.type) }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="user?.username" label="提交人" width="80" />
                  <el-table-column label="操作" width="150" fixed="right">
                    <template #default="{ row }">
                      <el-button size="small" @click="showAffairDetail(row)">详情</el-button>
                      <el-button size="small" type="success" @click="auditAffair(row, true)">通过</el-button>
                      <el-button size="small" type="danger" @click="auditAffair(row, false)">驳回</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
            <el-col :span="12">
              <el-card>
                <template #header>
                  <span class="card-title">已驳回 ({{ rejectedAffairs.length }})</span>
                  <el-button size="small" @click="loadRejectedAffairs" style="float: right;">刷新</el-button>
                </template>
                <el-table :data="rejectedAffairs" v-loading="loadingAllAffairs" height="350" empty-text="暂无已驳回事务">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="title" label="事务标题" min-width="120" show-overflow-tooltip />
                  <el-table-column prop="type" label="类型" width="80">
                    <template #default="{ row }">
                      <el-tag size="small">{{ getAffairTypeText(row.type) }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="user?.username" label="提交人" width="80" />
                  <el-table-column label="操作" width="100" fixed="right">
                    <template #default="{ row }">
                      <el-button size="small" @click="showAffairDetail(row)">详情</el-button>
                      <el-button size="small" type="success" @click="reApproveAffair(row.id)">重审</el-button>
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
                  <span class="card-title">待处理 ({{ processingAffairs.length }})</span>
                  <el-button size="small" @click="loadProcessingAffairs" style="float: right;">刷新</el-button>
                </template>
                <el-table :data="processingAffairs" v-loading="loadingAllAffairs" height="350" empty-text="暂无待处理事务">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="title" label="事务标题" min-width="120" show-overflow-tooltip />
                  <el-table-column prop="type" label="类型" width="80">
                    <template #default="{ row }">
                      <el-tag size="small">{{ getAffairTypeText(row.type) }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="user?.username" label="提交人" width="80" />
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
                  <span class="card-title">待确认 ({{ pendingConfirmAffairs.length }})</span>
                  <el-button size="small" @click="loadPendingConfirmAffairs" style="float: right;">刷新</el-button>
                </template>
                <el-table :data="pendingConfirmAffairs" v-loading="loadingAllAffairs" height="350" empty-text="暂无待确认事务">
                  <el-table-column prop="id" label="ID" width="60" />
                  <el-table-column prop="title" label="事务标题" min-width="120" show-overflow-tooltip />
                  <el-table-column prop="type" label="类型" width="80">
                    <template #default="{ row }">
                      <el-tag size="small">{{ getAffairTypeText(row.type) }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="user?.username" label="提交人" width="80" />
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
              <span>已完成事务 ({{ completedAffairs.length }})</span>
              <el-button size="small" @click="loadCompletedAffairs" style="float: right;">刷新</el-button>
            </template>
            <el-table :data="completedAffairs" v-loading="loadingAllAffairs" empty-text="暂无已完成事务">
              <el-table-column prop="id" label="ID" width="60" />
              <el-table-column prop="title" label="事务标题" min-width="150" show-overflow-tooltip />
              <el-table-column prop="type" label="类型" width="100">
                <template #default="{ row }">
                  <el-tag size="small">{{ getAffairTypeText(row.type) }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="address" label="事发地点" width="120" show-overflow-tooltip />
              <el-table-column prop="user?.username" label="提交人" width="100" />
              <el-table-column prop="handler_name" label="处理人" width="100" />
              <el-table-column prop="created_at" label="提交时间" width="160" :formatter="formatTime" />
              <el-table-column label="操作" width="80" fixed="right">
                <template #default="{ row }">
                  <el-button size="small" @click="showAffairDetail(row)">详情</el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-tab-pane>

        <!-- 敏感词与审核设置 -->
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
      </el-tabs>
    </div>

    <!-- 事务详情对话框 -->
    <el-dialog v-model="showAffairDialog" title="事务详情" width="700px">
      <div v-if="currentAffair" class="affair-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="事务标题" :span="2">{{ currentAffair.title }}</el-descriptions-item>
          <el-descriptions-item label="事务类型">
            <el-tag size="small">{{ getAffairTypeText(currentAffair.type) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="提交人">{{ currentAffair.user?.username }}</el-descriptions-item>
          <el-descriptions-item label="事务状态">
            <el-tag :type="getAffairStatusType(currentAffair.status)" size="small">{{ getAffairStatusText(currentAffair.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="事发地点" :span="2">{{ currentAffair.address }}</el-descriptions-item>
          <el-descriptions-item label="提交时间" :span="2">{{ formatTimeStr(currentAffair.created_at) }}</el-descriptions-item>
        </el-descriptions>
        <el-divider>事务描述</el-divider>
        <div class="affair-content">{{ currentAffair.content }}</div>
        <div v-if="currentAffair.images" class="affair-images">
          <el-image
            v-for="(img, idx) in parseImages(currentAffair.images)"
            :key="idx"
            :src="img"
            fit="cover"
            class="affair-image"
            :preview-src-list="parseImages(currentAffair.images)"
          />
        </div>
      </div>
      <template #footer>
        <el-button @click="showAffairDialog = false">关闭</el-button>
        <el-button v-if="currentAffair?.status === 0" type="success" @click="auditAffair(currentAffair, true)">通过</el-button>
        <el-button v-if="currentAffair?.status === 0" type="danger" @click="openRejectReasonDialog('affair')">驳回</el-button>
      </template>
    </el-dialog>

    <!-- 商品详情对话框 -->
    <el-dialog v-model="showProductDialog" title="商品详情" width="700px">
      <div v-if="currentProduct" class="product-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="商品名称">{{ currentProduct.title }}</el-descriptions-item>
          <el-descriptions-item label="价格">¥{{ currentProduct.price }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ currentProduct.type }}</el-descriptions-item>
          <el-descriptions-item label="库存">{{ currentProduct.stock }}</el-descriptions-item>
          <el-descriptions-item label="发布者">{{ currentProduct.user?.username }}</el-descriptions-item>
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
      </div>
      <template #footer>
        <el-button @click="showProductDialog = false">关闭</el-button>
        <el-button type="success" @click="passProduct(currentProduct?.id)">通过</el-button>
        <el-button type="danger" @click="openRejectReasonDialog('product')">驳回</el-button>
      </template>
    </el-dialog>

    <!-- 动态详情对话框 -->
    <el-dialog v-model="showPostDialog" title="动态详情" width="700px">
      <div v-if="currentPost" class="post-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="标题">{{ currentPost.title }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ currentPost.type }}</el-descriptions-item>
          <el-descriptions-item label="发布者">{{ currentPost.user?.username }}</el-descriptions-item>
          <el-descriptions-item label="点赞">{{ currentPost.likes }}</el-descriptions-item>
          <el-descriptions-item label="内容" :span="2"><div v-html="currentPost.content" style="max-height: 400px; overflow-y: auto; line-height: 1.8;" /></el-descriptions-item>
        </el-descriptions>
        <div v-if="currentPost.images" style="margin-top: 20px;">
          <p>图片：</p>
          <div class="post-images">
            <img v-for="(img, idx) in parseImages(currentPost.images)" :key="idx" :src="img" style="max-width: 200px; margin-right: 10px; border-radius: 8px;" />
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="showPostDialog = false">关闭</el-button>
        <el-button type="success" @click="passPost(currentPost?.id)">通过</el-button>
        <el-button type="danger" @click="openRejectReasonDialog('post')">退回</el-button>
      </template>
    </el-dialog>

    <!-- 知识详情对话框 -->
    <el-dialog v-model="showKnowledgeDialog" title="知识详情" width="700px">
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
      </div>
      <template #footer>
        <el-button @click="showKnowledgeDialog = false">关闭</el-button>
        <el-button type="success" @click="passKnowledge(currentKnowledge?.id)">通过</el-button>
        <el-button type="danger" @click="openRejectReasonDialog('knowledge')">退回</el-button>
      </template>
    </el-dialog>

    <!-- 驳回原因对话框 -->
    <el-dialog v-model="showRejectDialog" title="填写驳回原因" width="500px">
      <el-form :model="rejectForm" label-width="100px">
        <el-form-item label="驳回原因" required>
          <el-input v-model="rejectForm.reason" type="textarea" :rows="4" placeholder="请输入驳回原因" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRejectDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmReject">确认驳回</el-button>
      </template>
    </el-dialog>

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
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/utils/api'
import { getUserInfo, removeToken, removeUserInfo } from '@/utils/auth'

const router = useRouter()
const user = computed(() => getUserInfo())

const currentUser = ref({
  username: '',
  nickname: '',
  avatar_url: ''
})

const activeTab = ref('experts')
const reviewSubTab = ref('products')

// 专家申请
const pendingExperts = ref([])
const approvedExperts = ref([])
const rejectedExperts = ref([])
const loadingExperts = ref(false)

// 事务审核
const pendingAffairs = ref([])
const approvedAffairs = ref([])
const rejectedAffairs = ref([])
const processingAffairs = ref([])
const pendingConfirmAffairs = ref([])
const completedAffairs = ref([])
const loadingAffairs = ref(false)
const loadingAllAffairs = ref(false)

// 内容复核 - 商品
const onShelfProducts = ref([])
const offShelfProducts = ref([])
const pendingProducts = ref([])
const loadingAllProducts = ref(false)

// 内容复核 - 动态
const publishedPosts = ref([])
const deletedPosts = ref([])
const rejectedPosts = ref([])
const pendingPosts = ref([])
const loadingAllPosts = ref(false)

// 内容复核 - 知识
const publishedKnowledge = ref([])
const deletedKnowledge = ref([])
const rejectedKnowledges = ref([])
const pendingKnowledge = ref([])
const loadingAllKnowledge = ref(false)

// 处理人员管理
const pendingProcessors = ref([])
const approvedProcessors = ref([])
const rejectedProcessors = ref([])
const loadingProcessors = ref(false)

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

// 对话框状态
const showAffairDialog = ref(false)
const showProductDialog = ref(false)
const showPostDialog = ref(false)
const showKnowledgeDialog = ref(false)
const showRejectDialog = ref(false)

const currentAffair = ref(null)
const currentProduct = ref(null)
const currentPost = ref(null)
const currentKnowledge = ref(null)

const rejectForm = reactive({
  reason: '',
  type: '',
  id: null
})

// 工具函数
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

const getAffairTypeText = (type) => {
  const map = {
    'infrastructure': '基础设施',
    'hardware': '硬件问题',
    'env': '环境问题',
    'safety': '安全隐患',
    'other': '其他'
  }
  return map[type] || type
}

const getAffairStatusText = (status) => {
  const map = {
    1: '待审核',
    2: '待处理',
    3: '已驳回',
    4: '待确认',
    5: '已完成'
  }
  return map[status] || status
}

const getAffairStatusType = (status) => {
  const map = {
    1: 'warning',  // 待审核
    2: 'primary',  // 待处理
    3: 'danger',   // 已驳回
    4: 'warning',  // 待确认
    5: 'success'   // 已完成
  }
  return map[status] || ''
}

// 处理人员状态
const getProcessorStatusText = (status) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[status] || '')
const getProcessorStatusType = (status) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[status] || '')

// 专家状态
const getExpertStatusText = (status) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[status] || '')

// 敏感词类别
const getCategoryText = (cat) => {
  const map = { 'politics': '政治敏感', 'porn': '色情低俗', 'advertising': '广告推广', 'gamble': '赌博博彩', 'fraud': '欺诈诈骗', 'other': '其他违规' }
  return map[cat] || cat
}

// 敏感词级别
const getLevelText = (level) => ({ 1: '轻度', 2: '中度', 3: '重度' }[level] || '')
const getLevelType = (level) => ({ 1: 'info', 2: 'warning', 3: 'danger' }[level] || '')

// ========== 专家申请 ==========
const loadPendingExperts = async () => {
  loadingExperts.value = true
  try {
    const res = await api.getPendingExperts()
    pendingExperts.value = res?.list || []
  } catch (error) {
    ElMessage.error('加载待审核专家失败')
  } finally {
    loadingExperts.value = false
  }
}

const loadRejectedExperts = async () => {
  try {
    const res = await api.getAllExperts()
    rejectedExperts.value = (res?.list || []).filter(e => e.status === 2)
  } catch (error) {
    console.error('加载已拒绝专家失败:', error)
  }
}

const approveExpert = async (id) => {
  try {
    await api.approveExpert(id)
    ElMessage.success('已通过')
    loadPendingExperts()
    loadRejectedExperts()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '操作失败'
    ElMessage.error(msg)
  }
}

const rejectExpert = async (id) => {
  try {
    await ElMessageBox.confirm('确定要拒绝该专家申请吗？', '提示', { type: 'warning' })
    await api.rejectExpert(id)
    ElMessage.success('已拒绝')
    loadPendingExperts()
    loadApprovedExperts()
  } catch (error) {
    if (error !== 'cancel') {
      const msg = error?.response?.data?.error || error?.message || '操作失败'
      ElMessage.error(msg)
    }
  }
}

const loadApprovedExperts = async () => {
  try {
    const res = await api.getAllExperts()
    approvedExperts.value = (res?.list || []).filter(e => e.status === 1)
  } catch (error) {
    console.error('加载已通过专家失败:', error)
  }
}

const reApproveExpert = async (id) => {
  try {
    await api.approveExpert(id)
    ElMessage.success('已重新通过')
    loadApprovedExperts()
    loadRejectedExperts()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '操作失败'
    ElMessage.error(msg)
  }
}

const removeExpert = async (id) => {
  try {
    await ElMessageBox.confirm('确定要移除该专家身份吗？', '提示', { type: 'warning' })
    await api.deleteExpert(id)
    ElMessage.success('已移除')
    loadApprovedExperts()
  } catch (error) {
    if (error !== 'cancel') {
      const msg = error?.response?.data?.error || error?.message || '操作失败'
      ElMessage.error(msg)
    }
  }
}

// ========== 处理人员 ==========
const loadPendingProcessors = async () => {
  loadingProcessors.value = true
  try {
    const pendingRes = await api.getPendingProcessors()
    pendingProcessors.value = pendingRes?.list || []
  } catch (error) {
    console.error('加载待审核处理人员失败:', error)
  } finally {
    loadingProcessors.value = false
  }
}

const loadApprovedProcessors = async () => {
  try {
    const res = await api.getAllProcessors()
    approvedProcessors.value = (res?.list || []).filter(p => p.status === 1)
  } catch (error) {
    console.error('加载已通过处理人员失败:', error)
  }
}

const loadRejectedProcessors = async () => {
  try {
    const res = await api.getAllProcessors()
    rejectedProcessors.value = (res?.list || []).filter(p => p.status === 2)
  } catch (error) {
    console.error('加载已拒绝处理人员失败:', error)
  }
}

const approveProcessor = async (id) => {
  try {
    await api.approveProcessor(id)
    ElMessage.success('已通过')
    loadPendingProcessors()
    loadApprovedProcessors()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '操作失败'
    ElMessage.error(msg)
  }
}

const rejectProcessor = async (id) => {
  try {
    await ElMessageBox.confirm('确定要拒绝该申请吗？', '提示', { type: 'warning' })
    await api.rejectProcessor(id, { reason: '管理员拒绝' })
    ElMessage.success('已拒绝')
    loadPendingProcessors()
    loadRejectedProcessors()
  } catch (error) {
    if (error !== 'cancel') {
      const msg = error?.response?.data?.error || error?.message || '操作失败'
      ElMessage.error(msg)
    }
  }
}

const disableProcessor = async (id) => {
  try {
    await ElMessageBox.confirm('确定要禁用该处理人员吗？', '提示', { type: 'warning' })
    await api.disableProcessor(id)
    ElMessage.success('已禁用')
    loadApprovedProcessors()
  } catch (error) {
    if (error !== 'cancel') {
      const msg = error?.response?.data?.error || error?.message || '操作失败'
      ElMessage.error(msg)
    }
  }
}

const reApproveProcessor = async (id) => {
  try {
    await api.approveProcessor(id)
    ElMessage.success('已重新通过')
    loadApprovedProcessors()
    loadRejectedProcessors()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '操作失败'
    ElMessage.error(msg)
  }
}

// ========== 事务审核 ==========
const loadPendingAffairs = async () => {
  loadingAffairs.value = true
  try {
    const res = await api.getPendingAuditAffairs()
    pendingAffairs.value = res?.list || res || []
  } catch (error) {
    ElMessage.error('加载待审核事务失败')
  } finally {
    loadingAffairs.value = false
  }
}

const loadApprovedAffairs = async () => {
  loadingAllAffairs.value = true
  try {
    const res = await api.getAllRuralAffairs({ page: 1, page_size: 100 })
    // status 2 = 审核通过
    approvedAffairs.value = (res?.list || []).filter(a => a.status === 2)
  } catch (error) {
    console.error('加载已通过事务失败:', error)
  } finally {
    loadingAllAffairs.value = false
  }
}

const loadRejectedAffairs = async () => {
  loadingAllAffairs.value = true
  try {
    const res = await api.getAllRuralAffairs({ page: 1, page_size: 100 })
    // status 3 = 审核不通过（驳回）
    rejectedAffairs.value = (res?.list || []).filter(a => a.status === 3)
  } catch (error) {
    console.error('加载已驳回事务失败:', error)
  } finally {
    loadingAllAffairs.value = false
  }
}

// 加载待处理事务 (status=2)
const loadProcessingAffairs = async () => {
  loadingAllAffairs.value = true
  try {
    const res = await api.getAllRuralAffairs({ page: 1, page_size: 100 })
    // status 2 = 审核通过，待处理
    processingAffairs.value = (res?.list || []).filter(a => a.status === 2)
  } catch (error) {
    console.error('加载待处理事务失败:', error)
  } finally {
    loadingAllAffairs.value = false
  }
}

// 加载待确认事务 (status=4)
const loadPendingConfirmAffairs = async () => {
  loadingAllAffairs.value = true
  try {
    const res = await api.getAllRuralAffairs({ page: 1, page_size: 100 })
    // status 4 = 处理中，待提交者确认
    pendingConfirmAffairs.value = (res?.list || []).filter(a => a.status === 4)
  } catch (error) {
    console.error('加载待确认事务失败:', error)
  } finally {
    loadingAllAffairs.value = false
  }
}

// 加载已完成事务 (status=5)
const loadCompletedAffairs = async () => {
  loadingAllAffairs.value = true
  try {
    const res = await api.getAllRuralAffairs({ page: 1, page_size: 100 })
    // status 5 = 已完成
    completedAffairs.value = (res?.list || []).filter(a => a.status === 5)
  } catch (error) {
    console.error('加载已完成事务失败:', error)
  } finally {
    loadingAllAffairs.value = false
  }
}

const showAffairDetail = (affair) => {
  currentAffair.value = affair
  showAffairDialog.value = true
}

const auditAffair = async (affair, approved) => {
  if (!approved) {
    currentAffair.value = affair
    openRejectReasonDialog('affair')
    return
  }
  try {
    await api.auditRuralAffair(affair.id, { approved: true })
    ElMessage.success('已通过')
    showAffairDialog.value = false
    loadPendingAffairs()
    loadApprovedAffairs()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '操作失败'
    ElMessage.error(msg)
  }
}

const reApproveAffair = async (id) => {
  try {
    await api.auditRuralAffair(id, { approved: true })
    ElMessage.success('已重新通过')
    loadApprovedAffairs()
    loadRejectedAffairs()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '操作失败'
    ElMessage.error(msg)
  }
}

// ========== 内容复核 - 商品 ==========
const loadOnShelfProducts = async () => {
  loadingAllProducts.value = true
  try {
    const res = await api.getAllProducts({ page: 1, page_size: 100 })
    onShelfProducts.value = (res?.list || []).filter(p => p.status === 1)
  } catch (error) {
    console.error('加载上架商品失败:', error)
  } finally {
    loadingAllProducts.value = false
  }
}

const loadOffShelfProducts = async () => {
  loadingAllProducts.value = true
  try {
    const res = await api.getAllProducts({ page: 1, page_size: 100 })
    offShelfProducts.value = (res?.list || []).filter(p => p.status === 0)
  } catch (error) {
    console.error('加载下架商品失败:', error)
  } finally {
    loadingAllProducts.value = false
  }
}

const loadPendingProducts = async () => {
  loadingAllProducts.value = true
  try {
    const res = await api.getPendingProducts()
    pendingProducts.value = res?.list || []
  } catch (error) {
    console.error('加载待审核商品失败:', error)
  } finally {
    loadingAllProducts.value = false
  }
}

const passProduct = async (id) => {
  try {
    await api.auditProduct(id, { approved: true })
    ElMessage.success('已通过')
    loadPendingProducts()
    loadOnShelfProducts()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '操作失败'
    ElMessage.error(msg)
  }
}

const rejectProduct = async (id) => {
  try {
    await api.auditProduct(id, { approved: false })
    ElMessage.success('已驳回')
    loadPendingProducts()
    loadOffShelfProducts()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '操作失败'
    ElMessage.error(msg)
  }
}

const loadReviewProducts = () => {
  loadOnShelfProducts()
  loadOffShelfProducts()
  loadPendingProducts()
}

const showProductDetail = (product) => {
  currentProduct.value = product
  showProductDialog.value = true
}

const offShelfProduct = async (id) => {
  try {
    await ElMessageBox.confirm('确定要下架该商品吗？', '提示', { type: 'warning' })
    await api.offShelfProduct(id)
    ElMessage.success('已下架')
    loadOnShelfProducts()
    loadOffShelfProducts()
  } catch (error) {
    if (error !== 'cancel') {
      const msg = error?.response?.data?.error || error?.message || '操作失败'
      ElMessage.error(msg)
    }
  }
}

const reOnShelfProduct = async (id) => {
  try {
    await api.restoreProduct(id)
    ElMessage.success('已重新上架')
    loadOnShelfProducts()
    loadOffShelfProducts()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '操作失败'
    ElMessage.error(msg)
  }
}

// ========== 内容复核 - 动态 ==========
const loadPublishedPosts = async () => {
  loadingAllPosts.value = true
  try {
    const res = await api.getAllPosts({ page: 1, page_size: 100 })
    publishedPosts.value = (res?.list || []).filter(p => p.status === 1)
  } catch (error) {
    console.error('加载已发布动态失败:', error)
  } finally {
    loadingAllPosts.value = false
  }
}

const loadDeletedPosts = async () => {
  loadingAllPosts.value = true
  try {
    const res = await api.getDeletedPosts()
    deletedPosts.value = res?.list || []
  } catch (error) {
    console.error('加载已删除动态失败:', error)
  } finally {
    loadingAllPosts.value = false
  }
}

const loadPendingPosts = async () => {
  loadingAllPosts.value = true
  try {
    const res = await api.getPendingPosts()
    pendingPosts.value = res?.list || []
  } catch (error) {
    console.error('加载待审核动态失败:', error)
  } finally {
    loadingAllPosts.value = false
  }
}

const passPostReview = async (id) => {
  try {
    await api.auditPost(id, { approved: true })
    ElMessage.success('已通过')
    loadPendingPosts()
    loadPublishedPosts()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '操作失败'
    ElMessage.error(msg)
  }
}

const rejectPostReview = async (id) => {
  try {
    await api.auditPost(id, { approved: false })
    ElMessage.success('已驳回')
    loadPendingPosts()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '操作失败'
    ElMessage.error(msg)
  }
}

const loadReviewPosts = () => {
  loadPublishedPosts()
  loadRejectedPosts()
  loadPendingPosts()
}

const passPost = async (id) => {
  try {
    await api.auditPost(id, { approved: true })
    ElMessage.success('已通过')
    showPostDialog.value = false
    loadPendingPosts()
    loadPublishedPosts()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '操作失败'
    ElMessage.error(msg)
  }
}

const showPostDetail = (post) => {
  currentPost.value = post
  showPostDialog.value = true
}

const deletePost = async (id) => {
  try {
    await ElMessageBox.confirm('确定要退回该动态吗？', '提示', { type: 'warning' })
    // 退回 = 设置 status = 0（审核不通过），不是 status = 2（删除）
    await api.auditPost(id, { approved: false })
    ElMessage.success('已退回')
    loadPublishedPosts()
    loadDeletedPosts()
  } catch (error) {
    if (error !== 'cancel') {
      const msg = error?.response?.data?.error || error?.message || '操作失败'
      ElMessage.error(msg)
    }
  }
}

const restorePost = async (id) => {
  try {
    // 恢复 = 重新提交审核，设置 status = -1（待审核）
    await api.updatePost({ id, status: -1 })
    ElMessage.success('已恢复')
    loadRejectedPosts()
    loadPendingPosts()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '操作失败'
    ElMessage.error(msg)
  }
}

const loadRejectedPosts = async () => {
  loadingAllPosts.value = true
  try {
    const res = await api.getAllPosts({ page: 1, page_size: 100 })
    // 已退回 = status 0
    rejectedPosts.value = (res?.list || []).filter(p => p.status === 0)
  } catch (error) {
    console.error('加载已退回动态失败:', error)
  } finally {
    loadingAllPosts.value = false
  }
}

// ========== 内容复核 - 知识 ==========
const loadPublishedKnowledge = async () => {
  loadingAllKnowledge.value = true
  try {
    const res = await api.getAllKnowledge({ page: 1, pageSize: 100 })
    publishedKnowledge.value = (res?.data || res?.list || []).filter(k => k.status === 1)
  } catch (error) {
    console.error('加载已发布知识失败:', error)
  } finally {
    loadingAllKnowledge.value = false
  }
}

const loadRejectedKnowledge = async () => {
  loadingAllKnowledge.value = true
  try {
    const res = await api.getAllKnowledge({ page: 1, pageSize: 100 })
    // 已退回 = status 0
    rejectedKnowledges.value = (res?.data || res?.list || []).filter(k => k.status === 0)
  } catch (error) {
    console.error('加载已退回知识失败:', error)
  } finally {
    loadingAllKnowledge.value = false
  }
}

const loadPendingKnowledge = async () => {
  loadingAllKnowledge.value = true
  try {
    const res = await api.getPendingKnowledge()
    pendingKnowledge.value = res?.list || []
  } catch (error) {
    console.error('加载待审核知识失败:', error)
  } finally {
    loadingAllKnowledge.value = false
  }
}

const passKnowledgeReview = async (id) => {
  try {
    await api.auditKnowledge(id, { approved: true })
    ElMessage.success('已通过')
    loadPendingKnowledge()
    loadPublishedKnowledge()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '操作失败'
    ElMessage.error(msg)
  }
}

const passKnowledge = async (id) => {
  try {
    await api.auditKnowledge(id, { approved: true })
    ElMessage.success('已通过')
    showKnowledgeDialog.value = false
    loadPendingKnowledge()
    loadPublishedKnowledge()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '操作失败'
    ElMessage.error(msg)
  }
}

const rejectKnowledgeReview = async (id) => {
  try {
    await api.auditKnowledge(id, { approved: false })
    ElMessage.success('已驳回')
    loadPendingKnowledge()
    loadRejectedKnowledge()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '操作失败'
    ElMessage.error(msg)
  }
}

const loadReviewKnowledge = () => {
  loadPublishedKnowledge()
  loadRejectedKnowledge()
  loadPendingKnowledge()
}

const showKnowledgeDetail = (knowledge) => {
  currentKnowledge.value = knowledge
  showKnowledgeDialog.value = true
}

const deleteKnowledge = async (id) => {
  try {
    await ElMessageBox.confirm('确定要退回该知识吗？', '提示', { type: 'warning' })
    // 退回 = 设置 status = 0（审核不通过），不是 status = 2（删除）
    await api.auditKnowledge(id, { approved: false })
    ElMessage.success('已退回')
    loadPublishedKnowledge()
    loadRejectedKnowledge()
  } catch (error) {
    if (error !== 'cancel') {
      const msg = error?.response?.data?.error || error?.message || '操作失败'
      ElMessage.error(msg)
    }
  }
}

const restoreKnowledge = async (id) => {
  try {
    // 恢复 = 重新提交审核，设置 status = -1（待审核）
    await api.updateKnowledge(id, { status: -1 })
    ElMessage.success('已恢复')
    loadRejectedKnowledge()
    loadPendingKnowledge()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '操作失败'
    ElMessage.error(msg)
  }
}

const loadDeletedKnowledge = async () => {
  loadingAllKnowledge.value = true
  try {
    const res = await api.getDeletedKnowledge()
    deletedKnowledge.value = res?.list || []
  } catch (error) {
    console.error('加载已删除知识失败:', error)
  } finally {
    loadingAllKnowledge.value = false
  }
}



// ========== 驳回对话框 ==========
const openRejectReasonDialog = (type) => {
  rejectForm.type = type
  rejectForm.reason = ''
  showRejectDialog.value = true
}

const confirmReject = async () => {
  if (!rejectForm.reason) {
    ElMessage.warning('请输入驳回原因')
    return
  }
  try {
    if (rejectForm.type === 'affair' && currentAffair.value) {
      await api.auditRuralAffair(currentAffair.value.id, { approved: false, reject_reason: rejectForm.reason })
      ElMessage.success('已驳回')
      showAffairDialog.value = false
      loadPendingAffairs()
      loadRejectedAffairs()
    } else if (rejectForm.type === 'product' && currentProduct.value) {
      await api.offShelfProduct(currentProduct.value.id)
      ElMessage.success('已下架')
      showProductDialog.value = false
      loadReviewProducts()
    } else if (rejectForm.type === 'post' && currentPost.value) {
      // 退回动态 = 设置 status = 0
      await api.auditPost(currentPost.value.id, { approved: false })
      ElMessage.success('已退回')
      showPostDialog.value = false
      loadReviewPosts()
    } else if (rejectForm.type === 'knowledge' && currentKnowledge.value) {
      // 退回知识 = 设置 status = 0
      await api.auditKnowledge(currentKnowledge.value.id, { approved: false })
      ElMessage.success('已退回')
      showKnowledgeDialog.value = false
      loadReviewKnowledge()
    }
    showRejectDialog.value = false
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '操作失败'
    ElMessage.error(msg)
  }
}

// ========== 敏感词管理 ==========
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

// ========== 自动审核设置 ==========
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

const saveAutoReviewSetting = async () => {
  try {
    await api.setAutoReviewSetting(autoReviewSetting)
    ElMessage.success('设置已保存')
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '保存失败'
    ElMessage.error(msg)
  }
}

// ========== 退出登录 ==========
const handleLogout = () => {
  removeToken()
  removeUserInfo()
  ElMessage.success('已退出登录')
  router.push('/')
}

onMounted(() => {
  if (user.value) {
    const avatar_url = user.value.avatar_url?.startsWith('/') ? user.value.avatar_url : '/' + (user.value.avatar_url || '')
    currentUser.value = { ...user.value, avatar_url }

    if (user.value.role !== 'admin' && user.value.role !== 'sysadmin') {
      ElMessage.error('您没有权限访问此页面')
      router.push('/')
      return
    }
  } else {
    ElMessage.error('请先登录')
    router.push('/login')
    return
  }

  // 加载数据
  loadPendingExperts()
  loadApprovedExperts()
  loadRejectedExperts()
  loadPendingProcessors()
  loadApprovedProcessors()
  loadRejectedProcessors()
  loadPendingAffairs()
  loadRejectedAffairs()
  loadProcessingAffairs()
  loadPendingConfirmAffairs()
  loadCompletedAffairs()
  loadPendingProducts()
  loadOnShelfProducts()
  loadOffShelfProducts()
  loadPendingPosts()
  loadPublishedPosts()
  loadDeletedPosts()
  loadPendingKnowledge()
  loadPublishedKnowledge()
  loadDeletedKnowledge()

  // 敏感词管理与审核设置（admin 和 sysadmin 都可使用）
  if (user.value?.role === 'admin' || user.value?.role === 'sysadmin') {
    loadSensitiveWords()
    loadAutoReviewSetting()
  }
})
</script>

<style scoped>
.reviewer-container { min-height: 100vh; background: #f5f5f5; }
.header { background: #fff; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
.container { max-width: 1400px; margin: 0 auto; padding: 0 20px; }
.header .container { display: flex; align-items: center; height: 60px; }
.logo { font-size: 20px; font-weight: bold; color: #e6a23c; text-decoration: none; }
.nav { flex: 1; margin-left: 40px; }
.nav a { margin: 0 15px; color: #666; text-decoration: none; font-size: 15px; }
.nav a:hover, .nav a.router-link-active { color: #409eff; }

.main { padding: 20px; }

.affair-detail { padding: 10px 0; }
.affair-content { color: #333; line-height: 1.8; white-space: pre-wrap; }
.affair-images { display: flex; flex-wrap: wrap; gap: 10px; margin-top: 15px; }
.affair-image { width: 120px; height: 120px; border-radius: 4px; cursor: pointer; }

.post-images { display: flex; flex-wrap: wrap; gap: 10px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
</style>
