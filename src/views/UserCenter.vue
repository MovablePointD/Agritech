<template>
  <div class="user-center">
    <AppHeader />

    <div class="container main">
      <!-- 用户信息卡片 -->
      <el-card class="user-card">
        <div class="user-info">
          <div class="avatar-section">
            <el-avatar :src="currentUser.avatar_url" :size="120">{{ currentUser.username?.[0] }}</el-avatar>
            <el-button size="small" @click="showAvatarDialog = true">更换头像</el-button>
          </div>
          <div class="info-section">
            <h2>{{ currentUser.nickname || currentUser.username }}</h2>
            <p class="username-hint" v-if="currentUser.nickname !== currentUser.username">用户名：{{ currentUser.username }}</p>
            <p class="signature" v-if="currentUser.signature">{{ currentUser.signature }}</p>
            <p class="signature empty" v-else>暂无个人签名</p>
            <div class="meta-info">
              <span v-if="currentUser.phone"><el-icon><Phone /></el-icon> {{ currentUser.phone }}</span>
              <span v-if="currentUser.email"><el-icon><Message /></el-icon> {{ currentUser.email }}</span>
            </div>
          </div>
          <div class="action-section">
            <el-button type="primary" @click="showEditDialog = true">编辑资料</el-button>
            <el-button v-if="isSysAdmin" type="danger" @click="$router.push('/admin')">系统管理</el-button>
            <el-button v-if="isAdmin" type="warning" @click="$router.push('/reviewer')">审核中心</el-button>
            <el-button v-if="isProcessor" type="success" @click="$router.push('/processor')">事务处理中心</el-button>
            <el-button type="danger" plain @click="handleDeleteAccount">注销账号</el-button>
            <span style="margin-left: 10px; font-size: 12px; color: #999;">注销后账号资料将被保留，但无法登录</span>
          </div>
        </div>
      </el-card>

      <!-- 我处理的事务（仅处理人员可见）- 优先展示 -->
      <el-card v-if="isProcessor" style="margin-top: 20px;">
        <template #header>
          <div class="card-header">
            <span>我处理的事务</span>
            <div class="header-actions">
              <el-button type="success" size="small" @click="$router.push('/processor')">进入处理大厅</el-button>
              <el-button size="small" :icon="Refresh" @click="loadMyHandledAffairs">刷新</el-button>
            </div>
          </div>
        </template>
        <el-table v-if="!loadingHandledAffairs && myHandledAffairs.length > 0" :data="myHandledAffairs">
          <el-table-column prop="title" label="事务标题" min-width="150" show-overflow-tooltip />
          <el-table-column prop="type" label="类型" width="100">
            <template #default="{ row }">
              <el-tag size="small">{{ getAffairTypeText(row.type) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="address" label="事发地点" width="120" show-overflow-tooltip />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getAffairStatusType(row.status)" size="small">{{ getAffairStatusText(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="提交人" width="100">
            <template #default="{ row }">
              <span>{{ row.user?.nickname || '未知' }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="提交时间" width="160" :formatter="formatTime" />
          <el-table-column label="操作" width="280">
            <template #default="{ row }">
              <el-button size="small" type="primary" @click="goAffairDetail(row.id)">查看详情</el-button>
              <el-button v-if="row.status === 2" size="small" type="success" @click="startProcessAffair(row.id)">开始处理</el-button>
              <el-button v-if="row.status === 4 && row.handler_id === currentUserId" size="small" type="warning" @click="openProcessDialog(row)">填写处理结果</el-button>
              <el-button v-if="row.status === 6 && row.handler_id === currentUserId" size="small" type="primary" @click="openAnswerDialog(row)">追答</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div v-if="!loadingHandledAffairs && myHandledAffairs.length === 0" class="empty-state">
          <div class="empty-icon">
            <el-icon :size="60"><Setting /></el-icon>
          </div>
          <p class="empty-text">暂无需要处理的事务</p>
          <el-button type="primary" @click="$router.push('/processor')">去处理大厅看看</el-button>
        </div>
      </el-card>

      <!-- 我的事务（我提交的事务） -->
      <el-card style="margin-top: 20px;">
        <template #header>
          <div class="card-header">
            <span>我提交的事务</span>
            <div class="header-actions">
              <el-select v-model="affairStatus" placeholder="处理状态" size="small" style="width: 120px; margin-right: 10px;" clearable @change="loadMyAffairs">
                <el-option label="全部" :value="null" />
                <el-option label="待审核" :value="1" />
                <el-option label="审核通过" :value="2" />
                <el-option label="审核不通过" :value="3" />
                <el-option label="处理中" :value="4" />
                <el-option label="已完成" :value="5" />
              </el-select>
              <el-button size="small" :icon="Refresh" @click="loadMyAffairs">刷新</el-button>
            </div>
          </div>
        </template>
        <el-table v-if="!loadingAffairs && myAffairs.length > 0" :data="myAffairs">
          <el-table-column prop="title" label="事务标题" min-width="150" show-overflow-tooltip />
          <el-table-column prop="type" label="类型" width="100">
            <template #default="{ row }">
              <el-tag size="small">{{ getAffairTypeText(row.type) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="address" label="事发地点" width="120" show-overflow-tooltip />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getAffairStatusType(row.status)" size="small">{{ getAffairStatusText(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="handler_name" label="处理人" width="100">
            <template #default="{ row }">
              <span v-if="row.handler_name">{{ row.handler_name }}</span>
              <span v-else style="color: #999;">暂无</span>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="提交时间" width="160" :formatter="formatTime" />
          <el-table-column label="操作" width="240">
            <template #default="{ row }">
              <el-button size="small" type="primary" @click="goAffairDetail(row.id)">查看详情</el-button>
              <el-button v-if="canModifyAffair(row.status)" size="small" type="warning" @click="goModifyAffair(row.id)">修改事务</el-button>
              <el-button v-if="row.status === 4 || row.status === 6" size="small" type="danger" plain @click="openFollowUpDialog(row)">追问</el-button>
              <el-button v-if="row.status === 4 || row.status === 6" size="small" type="success" @click="openConfirmDialog(row)">确认完成</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div v-if="!loadingAffairs && myAffairs.length === 0" class="empty-state">
          <div class="empty-icon">
            <el-icon :size="60"><Bell /></el-icon>
          </div>
          <p class="empty-text">暂无提交的事务</p>
          <el-button type="primary" @click="$router.push('/affair/submit')">提交事务</el-button>
        </div>
      </el-card>

      <!-- 我的商品 -->
      <el-card style="margin-top: 20px;">
        <template #header>
          <div class="card-header">
            <span>我的商品</span>
            <div class="header-actions">
              <el-input v-model="productKeyword" placeholder="搜索商品" size="small" style="width: 200px; margin-right: 10px;" clearable @keyup.enter="loadMyProducts">
                <template #append><el-button :icon="Search" @click="loadMyProducts" /></template>
              </el-input>
              <el-button type="primary" size="small" @click="openProductDialog">发布商品</el-button>
            </div>
          </div>
        </template>
        <el-table v-if="!loadingProducts && myProducts.length > 0" :data="myProducts">
          <el-table-column prop="title" label="商品名称" />
          <el-table-column prop="price" label="价格" width="100">
            <template #default="{ row }">¥{{ row.price }}</template>
          </el-table-column>
          <el-table-column prop="type" label="类型" width="100" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getProductStatusType(row.status)" size="small">{{ getProductStatusText(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="发布时间" width="180" :formatter="formatTime" />
          <el-table-column label="操作" width="220">
            <template #default="{ row }">
              <el-button size="small" @click="showProductDetail(row)">查看</el-button>
              <el-button size="small" type="primary" @click="openEditProductDialog(row)">编辑</el-button>
              <el-button size="small" type="danger" @click="handleDeleteProduct(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div v-if="!loadingProducts && myProducts.length === 0" class="empty-state">
          <div class="empty-icon">
            <el-icon :size="60"><Goods /></el-icon>
          </div>
          <p class="empty-text">暂无发布的商品</p>
          <el-button type="primary" @click="openProductDialog">发布第一个商品</el-button>
        </div>
      </el-card>

      <!-- 我的订单 -->
      <el-card style="margin-top: 20px;">
        <template #header>
          <div class="card-header">
            <span>我的订单</span>
            <div class="header-actions">
              <el-select v-model="orderStatus" placeholder="订单状态" size="small" style="width: 120px; margin-right: 10px;" clearable @change="loadMyOrders">
                <el-option label="全部" :value="null" />
                <el-option label="待付款" :value="0" />
                <el-option label="待发货" :value="1" />
                <el-option label="待收货" :value="2" />
                <el-option label="已完成" :value="3" />
                <el-option label="已取消" :value="4" />
              </el-select>
              <el-input v-model="orderKeyword" placeholder="搜索订单" size="small" style="width: 200px;" clearable @keyup.enter="loadMyOrders">
                <template #append><el-button :icon="Search" @click="loadMyOrders" /></template>
              </el-input>
            </div>
          </div>
        </template>
        <el-table v-if="!loadingOrders && myOrders.length > 0" :data="myOrders">
          <el-table-column prop="id" label="订单号" width="100" />
          <el-table-column label="商品" min-width="200">
            <template #default="{ row }">
              <div v-if="row.items && row.items.length > 0" class="order-product">
                <div v-for="item in row.items" :key="item.id" class="order-item">
                  {{ item.product_name }} × {{ item.quantity }}
                </div>
              </div>
              <div v-else class="order-product">{{ row.product?.title || '商品' }}</div>
            </template>
          </el-table-column>
          <el-table-column prop="total_price" label="金额" width="100">
            <template #default="{ row }">¥{{ (row.total_price || 0).toFixed(2) }}</template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getOrderStatusType(row.status)" size="small">
                {{ getOrderStatusText(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="下单时间" width="180" :formatter="formatTime" />
          <el-table-column label="操作" width="220">
            <template #default="{ row }">
              <el-button size="small" @click="showOrderDetail(row)">查看</el-button>
              <el-button v-if="row.status === 0" size="small" type="primary" @click="handlePay(row)">付款</el-button>
              <el-button v-if="row.status === 0" size="small" type="danger" @click="handleCancelOrder(row.id)">取消</el-button>
              <el-button v-if="row.status === 2" size="small" type="success" @click="handleConfirmReceive(row)">确认收货</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div v-if="!loadingOrders && myOrders.length === 0" class="empty-state">
          <div class="empty-icon">
            <el-icon :size="60"><ShoppingCart /></el-icon>
          </div>
          <p class="empty-text">暂无订单</p>
          <el-button type="primary" @click="$router.push('/products')">去购物</el-button>
        </div>
      </el-card>

      <!-- 我的销售 -->
      <el-card style="margin-top: 20px;">
        <template #header>
          <div class="card-header">
            <span>我的销售</span>
            <div class="header-actions">
              <el-select v-model="sellerStatus" placeholder="订单状态" size="small" style="width: 120px; margin-right: 10px;" clearable @change="loadMySellerOrders">
                <el-option label="全部" :value="null" />
                <el-option label="待发货" :value="1" />
                <el-option label="待收货" :value="2" />
                <el-option label="已完成" :value="3" />
              </el-select>
              <el-button size="small" :icon="Refresh" @click="loadMySellerOrders">刷新</el-button>
            </div>
          </div>
        </template>
        <el-table v-if="!loadingSellerOrders && mySellerOrders.length > 0" :data="mySellerOrders">
          <el-table-column prop="order_no" label="订单编号" width="180" />
          <el-table-column label="商品" min-width="200">
            <template #default="{ row }">
              <div v-if="row.items && row.items.length > 0" class="order-product">
                <div v-for="item in row.items" :key="item.id" class="order-item">
                  {{ item.product_name }} × {{ item.quantity }}
                </div>
              </div>
              <div v-else class="order-product">-</div>
            </template>
          </el-table-column>
          <el-table-column label="买家信息" width="180">
            <template #default="{ row }">
              <div>收货人：{{ row.receiver }}</div>
              <div class="buyer-phone">{{ row.phone }}</div>
            </template>
          </el-table-column>
          <el-table-column prop="address" label="收货地址" min-width="200" show-overflow-tooltip />
          <el-table-column prop="total_price" label="金额" width="100">
            <template #default="{ row }">¥{{ (row.total_price || 0).toFixed(2) }}</template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getOrderStatusType(row.status)" size="small">
                {{ getOrderStatusText(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="{ row }">
              <el-button size="small" @click="showOrderDetail(row)">详情</el-button>
              <el-button v-if="row.status === 1" size="small" type="primary" @click="openShipDialog(row)">发货</el-button>
              <el-button v-if="row.express_no && row.status === 2" size="small" type="info" @click="showExpressInfo(row)">查看物流</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div v-if="!loadingSellerOrders && mySellerOrders.length === 0" class="empty-state">
          <div class="empty-icon">
            <el-icon :size="60"><Box /></el-icon>
          </div>
          <p class="empty-text">暂无销售订单</p>
          <el-button type="primary" @click="$router.push('/products')">去发布商品</el-button>
        </div>
      </el-card>

      <!-- 我的动态（移至最后） -->
      <el-card style="margin-top: 20px;">
        <template #header>
          <div class="card-header">
            <span>我的动态</span>
            <div class="header-actions">
              <el-button size="small" @click="showPostDialog = true">简洁发布</el-button>
              <el-button type="primary" size="small" @click="$router.push('/rich-editor/post')">
                <el-icon><EditPen /></el-icon>
                专业富文本编辑
              </el-button>
            </div>
          </div>
        </template>
        <el-table v-if="!loadingPosts && myPosts.length > 0" :data="myPosts">
          <el-table-column prop="title" label="标题" />
          <el-table-column prop="likes" label="点赞" width="80" />
          <el-table-column prop="views" label="浏览" width="80" />
          <el-table-column prop="created_at" label="发布时间" width="180" :formatter="formatTime" />
          <el-table-column label="操作" width="220">
            <template #default="{ row }">
              <el-button size="small" @click="$router.push(`/post/${row.id}`)">查看</el-button>
              <el-button size="small" type="primary" @click="openEditPostDialog(row)">编辑</el-button>
              <el-button size="small" type="danger" @click="handleDeletePost(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div v-if="!loadingPosts && myPosts.length === 0" class="empty-state">
          <div class="empty-icon">
            <el-icon :size="60"><Document /></el-icon>
          </div>
          <p class="empty-text">暂无发布的动态</p>
          <el-button type="primary" @click="showPostDialog = true">发布第一条动态</el-button>
        </div>
      </el-card>

      <!-- 我的知识（仅专家与系统管理员可见） -->
      <el-card v-if="canPublishKnowledge" style="margin-top: 20px;" ref="knowledgeCardRef">
        <template #header>
          <div class="card-header">
            <span>我的知识</span>
            <div class="header-actions">
              <el-button size="small" @click="openKnowledgeDialog()">简洁发布</el-button>
              <el-button type="primary" size="small" @click="$router.push('/rich-editor/knowledge')">
                <el-icon><EditPen /></el-icon>
                专业富文本编辑
              </el-button>
            </div>
          </div>
        </template>
        <el-table v-if="!loadingKnowledge && myKnowledges.length > 0" :data="myKnowledges">
          <el-table-column prop="title" label="标题" />
          <el-table-column prop="created_at" label="发布时间" width="180" :formatter="formatTime" />
          <el-table-column label="操作" width="220">
            <template #default="{ row }">
              <el-button size="small" @click="$router.push(`/knowledge/${row.id}`)">查看</el-button>
              <el-button size="small" type="primary" @click="editKnowledge(row)">编辑</el-button>
              <el-button size="small" type="danger" @click="handleDeleteKnowledge(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div v-if="!loadingKnowledge && myKnowledges.length === 0" class="empty-state">
          <div class="empty-icon">
            <el-icon :size="60"><Reading /></el-icon>
          </div>
          <p class="empty-text">暂无发布的知识</p>
          <el-button type="primary" @click="openKnowledgeDialog">发布第一条知识</el-button>
        </div>
      </el-card>
    </div>

    <!-- 编辑资料对话框 -->
    <el-dialog v-model="showEditDialog" title="编辑个人资料" width="500px">
      <el-form :model="editForm" label-width="80px">
        <el-form-item label="昵称">
          <el-input v-model="editForm.nickname" placeholder="请输入昵称" />
        </el-form-item>
        <el-form-item label="用户名">
          <el-input :value="currentUser.username" disabled placeholder="用户名不可修改" />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="editForm.phone" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="editForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="个人签名">
          <el-input v-model="editForm.signature" type="textarea" :rows="3" placeholder="请输入个人签名" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" @click="handleUpdateProfile" :loading="saving">保存</el-button>
      </template>
    </el-dialog>

    <!-- 处理结果对话框 -->
    <el-dialog v-model="showProcessDialog" title="填写处理结果" width="600px">
      <el-form :model="processForm" label-width="100px">
        <el-form-item label="事务标题">
          <span>{{ currentAffair?.title }}</span>
        </el-form-item>
        <el-form-item label="处理详情" required>
          <el-input v-model="processForm.content" type="textarea" :rows="6" placeholder="请详细描述处理过程和结果" />
        </el-form-item>
        <el-form-item label="处理图片">
          <el-upload
            action="/api/upload"
            :headers="{ Authorization: tokenValue }"
            list-type="picture-card"
            :on-success="handleProcessImageSuccess"
            :before-upload="beforeProcessUpload"
            :on-remove="handleProcessImageRemove"
            :file-list="processImageList"
          >
            <el-icon><Plus /></el-icon>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showProcessDialog = false">取消</el-button>
        <el-button type="primary" @click="submitProcess" :loading="processing">提交处理结果</el-button>
      </template>
    </el-dialog>

    <!-- 头像上传对话框 -->
    <el-dialog v-model="showAvatarDialog" title="更换头像" width="400px">
      <div class="upload-area">
        <el-upload
          class="avatar-uploader"
          action="/api/upload"
          :headers="{ Authorization: tokenValue }"
          :show-file-list="false"
          :on-success="handleAvatarSuccess"
          :before-upload="beforeAvatarUpload"
        >
          <img v-if="avatarPreview" :src="avatarPreview" class="avatar-preview" />
          <el-icon v-else class="avatar-uploader-icon"><Plus /></el-icon>
        </el-upload>
        <p class="upload-tip">点击上传头像图片，建议尺寸 200x200</p>
      </div>
    </el-dialog>

    <!-- 专家申请对话框 -->
    <el-dialog v-model="showExpertDialog" title="申请专家认证" width="600px">
      <el-form :model="expertForm" label-width="100px">
        <el-form-item label="真实姓名">
          <el-input v-model="expertForm.real_name" placeholder="请输入真实姓名" />
        </el-form-item>
        <el-form-item label="联系电话">
          <el-input v-model="expertForm.phone" placeholder="请输入联系电话" />
        </el-form-item>
        <el-form-item label="专业领域">
          <el-select v-model="expertForm.profession" style="width: 100%">
            <el-option label="种植" value="种植" />
            <el-option label="养殖" value="养殖" />
            <el-option label="农机" value="农机" />
            <el-option label="病虫害防治" value="病虫害防治" />
            <el-option label="农产品加工" value="农产品加工" />
          </el-select>
        </el-form-item>
        <el-form-item label="职位">
          <el-input v-model="expertForm.title" placeholder="请输入职位" />
        </el-form-item>
        <el-form-item label="所属单位">
          <el-input v-model="expertForm.company" placeholder="请输入所属单位" />
        </el-form-item>
        <el-form-item label="个人简介">
          <el-input v-model="expertForm.intro" type="textarea" :rows="3" placeholder="请输入个人简介" />
        </el-form-item>
        <el-form-item label="证书编号">
          <el-input v-model="expertForm.cert_no" placeholder="请输入证书编号" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showExpertDialog = false">取消</el-button>
        <el-button type="primary" @click="handleApplyExpert">提交申请</el-button>
      </template>
    </el-dialog>

    <!-- 发布/编辑动态对话框（简洁模式） -->
    <el-dialog v-model="showPostDialog" :title="postForm.id ? '编辑动态' : '发布动态'" width="600px">
      <el-form :model="postForm" label-width="80px">
        <el-form-item label="标题">
          <el-input v-model="postForm.title" placeholder="请输入标题" maxlength="100" show-word-limit />
        </el-form-item>
        <el-form-item label="内容">
          <el-input v-model="postForm.content" type="textarea" :rows="6" placeholder="分享您的想法、经验或问题..." />
        </el-form-item>
        <el-form-item label="图片">
          <el-upload
            action="/api/upload"
            :headers="{ Authorization: tokenValue }"
            list-type="picture-card"
            :on-success="handlePostUploadSuccess"
            :before-upload="beforePostUpload"
            :on-remove="handlePostRemove"
          >
            <el-icon><Plus /></el-icon>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <el-button type="primary" link @click="goToRichEditor('post', postForm.id)">
            <el-icon><EditPen /></el-icon>
            使用专业富文本编辑器
          </el-button>
          <div>
            <el-button @click="showPostDialog = false">取消</el-button>
            <el-button type="primary" @click="handlePublishPost" :loading="publishingPost">{{ postForm.id ? '保存' : '发布' }}</el-button>
          </div>
        </div>
      </template>
    </el-dialog>

    <!-- 发布/编辑商品对话框 -->
    <el-dialog v-model="showProductDialog" :title="productForm.id ? '编辑商品' : '发布商品'" width="600px">
      <el-form :model="productForm" label-width="80px">
        <el-form-item label="商品名称">
          <el-input v-model="productForm.title" placeholder="请输入商品名称" />
        </el-form-item>
        <el-form-item label="商品类型">
          <el-select v-model="productForm.type" style="width: 100%">
            <el-option label="种子" value="seed" />
            <el-option label="肥料" value="fertilizer" />
            <el-option label="农药" value="pesticide" />
            <el-option label="农具" value="tool" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="价格">
          <el-input-number v-model="productForm.price" :min="0" :precision="2" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="商品详情">
          <el-input v-model="productForm.content" type="textarea" :rows="4" placeholder="请输入商品详情" />
        </el-form-item>
        <el-form-item label="商品图片">
          <el-upload
            action="/api/upload"
            :headers="{ Authorization: tokenValue }"
            list-type="picture-card"
            :on-success="handleProductUploadSuccess"
            :before-upload="beforeProductUpload"
            :on-remove="handleProductRemove"
          >
            <el-icon><Plus /></el-icon>
          </el-upload>
        </el-form-item>
        <el-form-item label="库存">
          <el-input-number v-model="productForm.stock" :min="0" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="产地">
          <el-input v-model="productForm.address" placeholder="请输入产地或发货地" />
        </el-form-item>
        <el-form-item label="商品状态">
          <el-radio-group v-model="productForm.status">
            <el-radio :label="1">立即上架</el-radio>
            <el-radio :label="0">暂不上架</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showProductDialog = false">取消</el-button>
        <el-button type="primary" @click="handlePublishProduct" :loading="publishingProduct">发布</el-button>
      </template>
    </el-dialog>

    <!-- 发布知识对话框（简洁模式） -->
    <el-dialog v-model="showKnowledgeDialog" :title="knowledgeDialogTitle" width="600px">
      <el-form :model="knowledgeForm" label-width="80px">
        <el-form-item label="标题">
          <el-input v-model="knowledgeForm.title" placeholder="请输入知识标题" maxlength="100" show-word-limit />
        </el-form-item>
        <el-form-item label="内容">
          <el-input v-model="knowledgeForm.content" type="textarea" :rows="6" placeholder="请输入知识内容..." />
        </el-form-item>
        <el-form-item label="封面图片">
          <el-upload
            action="/api/upload"
            :headers="{ Authorization: tokenValue }"
            list-type="picture-card"
            :on-success="handleKnowledgeUploadSuccess"
            :before-upload="beforeKnowledgeUpload"
            :on-remove="handleKnowledgeRemove"
            :file-list="knowledgeImageList"
            :limit="1"
          >
            <el-icon><Plus /></el-icon>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <el-button type="primary" link @click="goToRichEditor('knowledge', knowledgeForm.id)">
            <el-icon><EditPen /></el-icon>
            使用专业富文本编辑器
          </el-button>
          <div>
            <el-button @click="showKnowledgeDialog = false">取消</el-button>
            <el-button type="primary" @click="handlePublishKnowledge" :loading="publishingKnowledge">{{ knowledgeForm.id ? '保存修改' : '发布' }}</el-button>
          </div>
        </div>
      </template>
    </el-dialog>

    <!-- 事务处理人员申请对话框 -->
    <el-dialog v-model="showProcessorDialog" title="申请成为事务处理人员" width="500px">
      <el-form :model="processorForm" label-width="100px">
        <el-form-item label="真实姓名" required>
          <el-input v-model="processorForm.real_name" placeholder="请输入真实姓名" />
        </el-form-item>
        <el-form-item label="联系电话" required>
          <el-input v-model="processorForm.phone" placeholder="请输入联系电话" />
        </el-form-item>
        <el-form-item label="负责区域" required>
          <el-input v-model="processorForm.address" placeholder="请输入您负责的区域，如：XX省XX市XX区" />
        </el-form-item>
        <el-form-item label="个人简介" required>
          <el-input v-model="processorForm.intro" type="textarea" :rows="4" placeholder="请简要介绍您的资历和处理能力" />
        </el-form-item>
        <el-form-item label="证明材料">
          <el-upload
            action="/api/upload"
            :headers="{ Authorization: tokenValue }"
            list-type="picture-card"
            :on-success="handleProcessorCertSuccess"
            :before-upload="beforeCertUpload"
            :on-remove="handleProcessorCertRemove"
            :file-list="processorCertList"
          >
            <el-icon><Plus /></el-icon>
          </el-upload>
          <div class="upload-tip">上传相关资质证明或工作证明图片</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showProcessorDialog = false">取消</el-button>
        <el-button type="primary" @click="handleApplyProcessor" :loading="applyingProcessor">提交申请</el-button>
      </template>
    </el-dialog>

    <!-- 商品详情对话框 -->
    <el-dialog v-model="showProductDetailDialog" title="商品详情" width="600px">
      <div v-if="currentProductDetail" class="product-detail">
        <el-image :src="currentProductDetail.image_url" fit="cover" style="width: 100%; height: 200px; margin-bottom: 20px;" />
        <h3>{{ currentProductDetail.title }}</h3>
        <p class="price">¥{{ currentProductDetail.price }}</p>
        <p class="info">类型：{{ currentProductDetail.type }} | 地址：{{ currentProductDetail.address }}</p>
        <el-divider />
        <h4>商品详情</h4>
        <p>{{ currentProductDetail.content }}</p>
      </div>
    </el-dialog>

    <!-- 订单详情对话框 -->
    <el-dialog v-model="showOrderDetailDialog" title="订单详情" width="600px">
      <div v-if="currentOrderDetail" class="order-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="订单编号">{{ currentOrderDetail.order_no }}</el-descriptions-item>
          <el-descriptions-item label="订单状态">
            <el-tag :type="getOrderStatusType(currentOrderDetail.status)" size="small">{{ getOrderStatusText(currentOrderDetail.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="收货人">{{ currentOrderDetail.receiver }}</el-descriptions-item>
          <el-descriptions-item label="联系电话">{{ currentOrderDetail.phone }}</el-descriptions-item>
          <el-descriptions-item label="收货地址" :span="2">{{ currentOrderDetail.address }}</el-descriptions-item>
          <el-descriptions-item label="总金额">¥{{ (currentOrderDetail.total_price || 0).toFixed(2) }}</el-descriptions-item>
          <el-descriptions-item label="物流单号" v-if="currentOrderDetail.express_no">{{ currentOrderDetail.express_no }}</el-descriptions-item>
          <el-descriptions-item label="下单时间">{{ formatTime({ created_at: currentOrderDetail.created_at }) }}</el-descriptions-item>
        </el-descriptions>
        <el-divider>商品清单</el-divider>
        <div v-if="currentOrderDetail.items && currentOrderDetail.items.length > 0">
          <div v-for="item in currentOrderDetail.items" :key="item.id" class="order-item-detail">
            <span>商品：{{ item.product_name }}</span>
            <span>数量：{{ item.quantity }}</span>
            <span>单价：¥{{ item.price?.toFixed(2) }}</span>
            <span>小计：¥{{ (item.subtotal || 0).toFixed(2) }}</span>
          </div>
        </div>
        <div v-else class="empty-items">暂无商品信息</div>
      </div>
      <template #footer>
        <el-button @click="showOrderDetailDialog = false">关闭</el-button>
        <el-button v-if="currentOrderDetail?.status === 0" type="primary" @click="handlePay(currentOrderDetail)">立即付款</el-button>
        <el-button v-if="currentOrderDetail?.status === 2" type="success" @click="handleConfirmReceive(currentOrderDetail)">确认收货</el-button>
      </template>
    </el-dialog>

    <!-- 发货对话框 -->
    <el-dialog v-model="showShipDialog" title="发货" width="500px">
      <el-form :model="shipForm" label-width="100px">
        <el-form-item label="订单编号">
          <div>{{ shipForm.orderNo }}</div>
        </el-form-item>
        <el-form-item label="收货信息">
          <div>{{ shipForm.receiver }} {{ shipForm.phone }}</div>
          <div class="ship-address">{{ shipForm.address }}</div>
        </el-form-item>
        <el-form-item label="物流单号" required>
          <el-input v-model="shipForm.expressNo" placeholder="请输入物流单号" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="shipForm.remark" type="textarea" :rows="2" placeholder="选填" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showShipDialog = false">取消</el-button>
        <el-button type="primary" @click="handleShip" :loading="shipping">确认发货</el-button>
      </template>
    </el-dialog>

    <!-- 物流信息对话框 -->
    <el-dialog v-model="showExpressDialog" title="物流信息" width="500px">
      <el-descriptions :column="1" border v-if="currentExpressOrder">
        <el-descriptions-item label="订单编号">{{ currentExpressOrder.order_no }}</el-descriptions-item>
        <el-descriptions-item label="物流单号">{{ currentExpressOrder.express_no }}</el-descriptions-item>
        <el-descriptions-item label="收货人">{{ currentExpressOrder.receiver }}</el-descriptions-item>
        <el-descriptions-item label="联系电话">{{ currentExpressOrder.phone }}</el-descriptions-item>
        <el-descriptions-item label="收货地址">{{ currentExpressOrder.address }}</el-descriptions-item>
        <el-descriptions-item label="订单状态">
          <el-tag :type="getOrderStatusType(currentExpressOrder.status)" size="small">{{ getOrderStatusText(currentExpressOrder.status) }}</el-tag>
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="showExpressDialog = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 事务详情对话框 -->
    <el-dialog v-model="showAffairDialog" title="事务详情" width="600px">
      <div v-if="currentAffair" class="affair-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="事务标题" :span="2">{{ currentAffair.title }}</el-descriptions-item>
          <el-descriptions-item label="事务类型">
            <el-tag size="small">{{ getAffairTypeText(currentAffair.type) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="处理状态">
            <el-tag :type="getAffairStatusType(currentAffair.status)" size="small">{{ getAffairStatusText(currentAffair.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="事发地点" :span="2">{{ currentAffair.address }}</el-descriptions-item>
          <el-descriptions-item label="提交时间" :span="2">{{ formatTime({ created_at: currentAffair.created_at }) }}</el-descriptions-item>
        </el-descriptions>

        <el-divider>事务描述</el-divider>
        <div class="affair-content">{{ currentAffair.content }}</div>

        <!-- 图片展示 -->
        <div v-if="currentAffair.images" class="affair-images">
          <el-image
            v-for="(img, idx) in JSON.parse(currentAffair.images || '[]')"
            :key="idx"
            :src="img"
            fit="cover"
            class="affair-image"
            :preview-src-list="JSON.parse(currentAffair.images || '[]')"
          />
        </div>

        <!-- 视频展示 -->
        <div v-if="currentAffair.video_url" class="affair-video">
          <el-divider>视频</el-divider>
          <video :src="currentAffair.video_url" controls class="affair-video-player" />
        </div>

        <!-- 审核信息（如果被驳回） -->
        <div v-if="currentAffair.status === 3" class="reject-info">
          <el-divider>驳回原因</el-divider>
          <el-alert type="error" :closable="false" show-icon>
            {{ currentAffair.reject_reason || '未提供原因' }}
          </el-alert>
        </div>

        <!-- 处理信息 -->
        <div v-if="currentAffair.status >= 4" class="process-info">
          <el-divider>处理信息</el-divider>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="处理人员">{{ currentAffair.handler_name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="处理时间">{{ currentAffair.process_time ? new Date(currentAffair.process_time).toLocaleString('zh-CN') : '-' }}</el-descriptions-item>
            <el-descriptions-item label="处理详情" :span="2">{{ currentAffair.process_content || '-' }}</el-descriptions-item>
          </el-descriptions>
          <!-- 处理图片 -->
          <div v-if="currentAffair.process_images" class="process-images">
            <el-image
              v-for="(img, idx) in JSON.parse(currentAffair.process_images || '[]')"
              :key="idx"
              :src="img"
              fit="cover"
              class="process-image"
              :preview-src-list="JSON.parse(currentAffair.process_images || '[]')"
            />
          </div>
        </div>

        <!-- 完成信息 -->
        <div v-if="currentAffair.status === 5" class="completed-info">
          <el-divider>完成信息</el-divider>
          <el-alert type="success" :closable="false" show-icon>
            事务已于 {{ currentAffair.completed_time ? new Date(currentAffair.completed_time).toLocaleString('zh-CN') : '' }} 完成处理
          </el-alert>
        </div>
      </div>
      <template #footer>
        <el-button @click="showAffairDialog = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 追答对话框（处理人员回复追问） -->
    <el-dialog v-model="showAnswerDialog" title="回复追问" width="500px">
      <el-form label-width="80px">
        <el-form-item label="事务标题">
          <span>{{ currentAffair?.title }}</span>
        </el-form-item>
        <el-form-item label="追答内容">
          <el-input v-model="answerContent" type="textarea" :rows="4" placeholder="请输入追答内容" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAnswerDialog = false">取消</el-button>
        <el-button type="primary" @click="submitAnswer" :loading="processing">提交追答</el-button>
      </template>
    </el-dialog>

    <!-- 追问对话框（用户对处理结果追问） -->
    <el-dialog v-model="showFollowUpDialog" title="追问" width="500px">
      <el-form label-width="80px">
        <el-form-item label="事务标题">
          <span>{{ currentAffair?.title }}</span>
        </el-form-item>
        <el-form-item label="追问内容">
          <el-input v-model="followUpContent" type="textarea" :rows="4" placeholder="请输入追问内容" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showFollowUpDialog = false">取消</el-button>
        <el-button type="warning" @click="submitFollowUp" :loading="processing">提交追问</el-button>
      </template>
    </el-dialog>

    <!-- 确认完成对话框 -->
    <el-dialog v-model="showConfirmDialog" title="确认事务完成" width="450px">
      <div style="text-align: center; padding: 20px;">
        <el-icon :size="60" color="#67c23a"><CircleCheck /></el-icon>
        <p style="margin-top: 16px; font-size: 15px; color: #555;">
          确认事务「{{ currentAffair?.title }}」已处理完成？
        </p>
        <p style="color: #999; font-size: 13px;">确认后将通知处理人员，事务状态变为已完成</p>
      </div>
      <template #footer>
        <el-button @click="showConfirmDialog = false">取消</el-button>
        <el-button type="success" @click="handleConfirmComplete" :loading="processing">确认完成</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Phone, Message, Plus, Search, Document, Goods, ShoppingCart, Reading, Box, Refresh, Bell, EditPen, CircleCheck } from '@element-plus/icons-vue'
import api from '@/utils/api'
import { getToken, getUserInfo, setUserInfo, removeToken } from '@/utils/auth'

const router = useRouter()
const token = computed(() => getToken())
const tokenValue = computed(() => getToken() || '')

const currentUser = ref({
  username: '',
  nickname: '',
  phone: '',
  email: '',
  signature: '',
  avatar_url: ''
})

const editForm = reactive({
  nickname: '',
  phone: '',
  email: '',
  signature: ''
})

const saving = ref(false)
const loadingPosts = ref(false)
const myPosts = ref([])
const myExpert = ref(null)
const showEditDialog = ref(false)
const showAvatarDialog = ref(false)
const showExpertDialog = ref(false)
const showProductDialog = ref(false)
const avatarPreview = ref('')

// 发布动态相关
const showPostDialog = ref(false)
const postForm = reactive({
  id: null,
  title: '',
  type: 'normal',
  content: '',
  images: ''
})
const publishingPost = ref(false)

// 我的商品
const myProducts = ref([])
const loadingProducts = ref(false)
const showProductDetailDialog = ref(false)
const currentProductDetail = ref(null)

// 订单详情
const showOrderDetailDialog = ref(false)
const currentOrderDetail = ref(null)
const productKeyword = ref('')

// 我的订单
const myOrders = ref([])
const loadingOrders = ref(false)
const orderStatus = ref('')
const orderKeyword = ref('')

// 我的销售
const mySellerOrders = ref([])
const loadingSellerOrders = ref(false)
const sellerStatus = ref('')
const showShipDialog = ref(false)
const showExpressDialog = ref(false)
const currentExpressOrder = ref(null)
const shipping = ref(false)
const shipForm = reactive({
  orderId: null,
  orderNo: '',
  receiver: '',
  phone: '',
  address: '',
  expressNo: '',
  remark: ''
})

// 我的事务
const myAffairs = ref([])
const loadingAffairs = ref(false)
const affairStatus = ref('')
const showAffairDialog = ref(false)
const currentAffair = ref(null)

// 我处理的事务
const myHandledAffairs = ref([])
const loadingHandledAffairs = ref(false)
const handledAffairsCardRef = ref(null)
const showProcessDialog = ref(false)
const processing = ref(false)
const processForm = reactive({
  content: '',
  images: ''
})
const processImageList = ref([])

// 我的知识
const myKnowledges = ref([])
const loadingKnowledge = ref(false)
const knowledgeCardRef = ref(null)
const showKnowledgeDialog = ref(false)
const knowledgeForm = reactive({
  id: null,
  title: '',
  type: 'planting',
  content: '',
  image_url: ''
})
const publishingKnowledge = ref(false)
const knowledgeDialogTitle = computed(() => knowledgeForm.id ? '编辑知识' : '发布知识')
const knowledgeImageList = computed(() => {
  if (knowledgeForm.image_url) {
    return [{ url: knowledgeForm.image_url }]
  }
  return []
})

// 判断是否为系统管理员
const isSysAdmin = computed(() => {
  return currentUser.value?.role === 'sysadmin'
})

// 判断是否为审核人员
const isAdmin = computed(() => {
  return currentUser.value?.role === 'admin'
})

// 判断是否为专家用户（审核通过或系统管理员直接拥有）
// 修复双角色问题：当用户同时拥有 processor 和 expert 角色时，role 字段可能为 "processor"，
// 但 myExpert 表中 status=1 表示已通过专家认证，应同时识别专家身份
const isExpert = computed(() => {
  if (isSysAdmin.value) return true
  if (currentUser.value?.role === 'expert') return true
  // 双角色场景：role 不是 expert，但专家认证已通过
  return myExpert.value?.status === 1
})

// 判断用户角色是否为 expert（不管审核状态）
const isExpertUser = computed(() => {
  if (isSysAdmin.value) return true
  return currentUser.value?.role === 'expert'
})

// 事务处理人员
const myProcessor = ref(null)
const isProcessor = ref(false)
const showProcessorDialog = ref(false)
const applyingProcessor = ref(false)
const processorForm = reactive({
  real_name: '',
  phone: '',
  address: '',
  intro: '',
  cert_images: ''
})
const processorCertList = ref([])

// 判断是否为处理人员（审核通过或系统管理员直接拥有）
const isProcessorUser = computed(() => {
  if (isSysAdmin.value) return true
  return isProcessor.value
})

// 获取处理人员状态文本
const getProcessorStatusText = (status) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[status] || '')
const getProcessorStatusType = (status) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[status] || '')

// 用户身份文本（支持展示多重角色）
const roleText = computed(() => {
  const roles = []
  if (isSysAdmin.value) roles.push('系统管理员')
  if (isAdmin.value) roles.push('审核人员')
  if (isExpert.value) roles.push('专家')
  if (isProcessorUser.value) roles.push('事务处理人员')
  if (roles.length === 0) roles.push('用户')
  return roles.join('、')
})

// 用户身份标签类型
const roleTagType = computed(() => {
  if (isSysAdmin.value || isAdmin.value) return 'danger'
  if (isExpert.value) return 'warning'
  if (isProcessorUser.value) return 'success'
  return 'info'
})

// 商品表单
const productForm = reactive({
  id: null,
  title: '',
  type: 'seed',
  price: 0,
  content: '',
  image_url: '',
  stock: 0,
  address: '',
  status: 1,
  publisher: null
})
const publishingProduct = ref(false)

const expertForm = reactive({
  real_name: '',
  phone: '',
  profession: '',
  title: '',
  company: '',
  intro: '',
  cert_no: ''
})

const formatTime = (row) => row.created_at ? new Date(row.created_at).toLocaleString('zh-CN') : ''

const getStatusType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || '')
const getStatusText = (s) => ({ 0: '待审核', 1: '已通过', 2: '未通过' }[s] || '')

// 商品状态
const getProductStatusText = (status) => {
  const map = { '-1': '待审核', '0': '退回/下架', '1': '上架' }
  return map[String(status)] || '未知'
}
const getProductStatusType = (status) => {
  const map = { '-1': 'warning', '0': 'danger', '1': 'success' }
  return map[String(status)] || ''
}

// 知识状态
const getKnowledgeStatusText = (status) => {
  const map = { '-1': '待审核', '0': '退回/隐藏', '1': '已发布', '2': '删除' }
  return map[String(status)] || '未知'
}
const getKnowledgeStatusType = (status) => {
  const map = { '-1': 'warning', '0': 'danger', '1': 'success', '2': 'info' }
  return map[String(status)] || ''
}

// 判断用户是否可以发布知识（仅系统管理员与已通过认证的专家）
const canPublishKnowledge = computed(() => {
  return isSysAdmin.value || isExpert.value
})

// 初始化用户信息
const initUserInfo = () => {
  const userData = getUserInfo()
  if (userData) {
    const avatar_url = userData.avatar_url?.startsWith('/') ? userData.avatar_url : '/' + (userData.avatar_url || '')
    const nickname = userData.nickname || userData.username || ''
    currentUser.value = { ...userData, avatar_url, nickname }
    Object.assign(editForm, {
      nickname: nickname,
      phone: userData.phone || '',
      email: userData.email || '',
      signature: userData.signature || ''
    })
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
      setUserInfo({ ...res, avatar_url, nickname })
      Object.assign(editForm, {
        nickname: nickname,
        phone: res.phone || '',
        email: res.email || '',
        signature: res.signature || ''
      })
    }
  } catch (error) {
    console.error('加载用户信息失败:', error)
  }
}

// 更新个人资料
const handleUpdateProfile = async () => {
  if (!editForm.nickname) {
    ElMessage.warning('请输入昵称')
    return
  }
  saving.value = true
  try {
    // 只提交允许修改的字段，不提交username
    await api.updateUser({
      nickname: editForm.nickname,
      phone: editForm.phone,
      email: editForm.email,
      signature: editForm.signature
    })
    const avatar_url = currentUser.value.avatar_url
    currentUser.value = { ...currentUser.value, ...editForm, avatar_url }
    setUserInfo({ ...currentUser.value })
    ElMessage.success('资料更新成功')
    showEditDialog.value = false
  } catch (error) {
    ElMessage.error('更新失败')
  } finally {
    saving.value = false
  }
}

// 头像上传前校验
const beforeAvatarUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) {
    ElMessage.error('只能上传图片文件')
    return false
  }
  if (!isLt5M) {
    ElMessage.error('图片大小不能超过5MB')
    return false
  }
  // 预览头像
  avatarPreview.value = URL.createObjectURL(file)
  return true
}

// 头像上传成功
const handleAvatarSuccess = async (res) => {
  try {
    const newAvatarUrl = res.url
    await api.updateUser({ avatar_url: newAvatarUrl })
    const avatar_url = newAvatarUrl.startsWith('/') ? newAvatarUrl : '/' + newAvatarUrl
    currentUser.value.avatar_url = avatar_url
    setUserInfo({ ...currentUser.value })
    ElMessage.success('头像更新成功')
    showAvatarDialog.value = false
    avatarPreview.value = ''
  } catch (error) {
    ElMessage.error('头像更新失败')
  }
}

// 加载我的动态
const loadMyPosts = async () => {
  loadingPosts.value = true
  try {
    const res = await api.getMyPosts()
    // 处理返回数据，可能直接是数组或 {list: [...]}
    myPosts.value = res?.list || res || []
  } catch (error) {
    console.error('加载动态失败:', error)
    myPosts.value = []
  } finally {
    loadingPosts.value = false
  }
}

// 删除动态
const handleDeletePost = async (id) => {
  try {
    await api.deletePost(id)
    ElMessage.success('删除成功')
    loadMyPosts()
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

// 发布动态 - 图片上传前校验
const beforePostUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) ElMessage.error('只能上传图片')
  if (!isLt5M) ElMessage.error('图片大小不能超过5MB')
  return isImage && isLt5M
}

// 发布动态 - 图片上传成功
const handlePostUploadSuccess = (res) => {
  // 处理图片URL
  const newUrl = res.url
  if (postForm.images) {
    try {
      const currentImages = JSON.parse(postForm.images)
      postForm.images = JSON.stringify([...currentImages, newUrl])
    } catch {
      postForm.images = JSON.stringify([newUrl])
    }
  } else {
    postForm.images = JSON.stringify([newUrl])
  }
}

// 发布动态 - 图片移除
const handlePostRemove = () => {
  postForm.images = ''
}

// 跳转到专业富文本编辑器
const goToRichEditor = (contentType, id = null) => {
  const path = `/rich-editor/${contentType}`
  router.push(id ? `${path}?id=${id}` : path)
}

// 打开编辑动态对话框
const openEditPostDialog = (post) => {
  postForm.id = post.id
  postForm.title = post.title
  postForm.content = post.content
  postForm.images = post.images || ''
  showPostDialog.value = true
}

// 发布/编辑动态
const handlePublishPost = async () => {
  if (!postForm.title) {
    ElMessage.warning('请输入标题')
    return
  }
  if (!postForm.content) {
    ElMessage.warning('请输入内容')
    return
  }
  publishingPost.value = true
  try {
    const postData = {
      title: postForm.title,
      type: 'normal',
      content: postForm.content,
      images: postForm.images || ''
    }

    if (postForm.id) {
      await api.updatePost({ id: postForm.id, ...postData })
      ElMessage.success('保存成功')
    } else {
      await api.createPost(postData)
      ElMessage.success('发布成功')
    }
    showPostDialog.value = false
    postForm.id = null
    postForm.title = ''
    postForm.type = 'normal'
    postForm.content = ''
    postForm.images = ''
    loadMyPosts()
  } catch (error) {
    ElMessage.error(postForm.id ? '保存失败' : '发布失败')
  } finally {
    publishingPost.value = false
  }
}

// 加载专家信息
const loadMyExpert = async () => {
  try {
    const res = await api.getMyExpert()
    myExpert.value = res || null
  } catch (error) {
    myExpert.value = null
  }
}

// 申请专家认证
const handleApplyExpert = async () => {
  if (!expertForm.real_name || !expertForm.profession) {
    ElMessage.warning('请填写必填项')
    return
  }
  try {
    await api.applyExpert(expertForm)
    ElMessage.success('申请已提交，请等待审核')
    showExpertDialog.value = false
    loadMyExpert()
  } catch (error) {
    ElMessage.error('申请失败')
  }
}

// 加载我的处理人员申请状态
const loadMyProcessor = async () => {
  try {
    const res = await api.getMyProcessor()
    myProcessor.value = res || null
    // 修复双角色问题：当用户同时拥有 processor 和 expert 角色时，可能 role 不为 "processor"，
    // 但 processor 表中 status=1 表示已通过处理人员认证，应同时识别处理人员身份
    isProcessor.value = currentUser.value?.role === 'processor' || myProcessor.value?.status === 1
  } catch (error) {
    myProcessor.value = null
    isProcessor.value = false
  }
}

// 申请成为事务处理人员
const handleApplyProcessor = async () => {
  if (!processorForm.real_name || !processorForm.phone || !processorForm.address || !processorForm.intro) {
    ElMessage.warning('请填写必填项')
    return
  }
  applyingProcessor.value = true
  try {
    await api.applyProcessor(processorForm)
    ElMessage.success('申请已提交，请等待审核')
    showProcessorDialog.value = false
    loadMyProcessor()
    // 重置表单
    processorForm.real_name = ''
    processorForm.phone = ''
    processorForm.address = ''
    processorForm.intro = ''
    processorForm.cert_images = ''
    processorCertList.value = []
  } catch (error) {
    ElMessage.error(error.message || '申请失败')
  } finally {
    applyingProcessor.value = false
  }
}

// 事务处理人员证明材料上传前校验
const beforeCertUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) ElMessage.error('只能上传图片文件')
  if (!isLt5M) ElMessage.error('图片大小不能超过5MB')
  return isImage && isLt5M
}

// 事务处理人员证明材料上传成功
const handleProcessorCertSuccess = (res) => {
  const newUrl = res.url
  if (processorForm.cert_images) {
    try {
      const currentImages = JSON.parse(processorForm.cert_images)
      processorForm.cert_images = JSON.stringify([...currentImages, newUrl])
    } catch {
      processorForm.cert_images = JSON.stringify([newUrl])
    }
  } else {
    processorForm.cert_images = JSON.stringify([newUrl])
  }
}

// 事务处理人员证明材料移除
const handleProcessorCertRemove = () => {
  processorForm.cert_images = ''
}

// 加载我的商品
const loadMyProducts = async () => {
  loadingProducts.value = true
  try {
    const res = await api.getMyProducts()
    let products = res?.list || res || []
    if (productKeyword.value) {
      products = products.filter(p => p.title?.includes(productKeyword.value))
    }
    myProducts.value = products.map(item => ({
      ...item,
      image_url: item.image_url?.startsWith('/') ? item.image_url : '/' + item.image_url
    }))
  } catch (error) {
    console.error('加载商品失败:', error)
    myProducts.value = []
  } finally {
    loadingProducts.value = false
  }
}

// 删除商品
const handleDeleteProduct = async (id) => {
  try {
    await api.deleteProduct(id)
    ElMessage.success('删除成功')
    loadMyProducts()
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

// 查看商品详情
const showProductDetail = async (product) => {
  currentProductDetail.value = product
  showProductDetailDialog.value = true
}

// 查看订单详情
const showOrderDetail = async (order) => {
  currentOrderDetail.value = order
  showOrderDetailDialog.value = true
}

// 加载我的事务
const loadMyAffairs = async () => {
  loadingAffairs.value = true
  try {
    const res = await api.getMyRuralAffairs()
    let affairs = res?.list || res || []
    if (affairStatus.value !== null && affairStatus.value !== '') {
      affairs = affairs.filter(a => a.status === affairStatus.value)
    }
    myAffairs.value = affairs
  } catch (error) {
    console.error('加载事务失败:', error)
    myAffairs.value = []
  } finally {
    loadingAffairs.value = false
  }
}

// 加载我处理的事务
const loadMyHandledAffairs = async () => {
  loadingHandledAffairs.value = true
  try {
    const res = await api.getMyHandledAffairs({ page: 1, page_size: 100 })
    myHandledAffairs.value = res?.list || []
  } catch (error) {
    console.error('加载处理事务失败:', error)
    myHandledAffairs.value = []
  } finally {
    loadingHandledAffairs.value = false
  }
}

// 当前用户ID（用于判断是否为事务处理人）
const currentUserId = computed(() => currentUser.value?.id)

// 导航到事务详情页
const goAffairDetail = (id) => {
  router.push(`/affair/${id}`)
}

// 导航到事务详情页并打开修改（通过 query 参数提示）
const goModifyAffair = (id) => {
  router.push(`/affair/${id}?action=modify`)
}

// 判断事务是否可以被提交者修改（未进入处理流程或审核未通过）
const canModifyAffair = (status) => {
  return status === 1 || status === 3
}

// 开始处理事务
const startProcessAffair = async (affairId) => {
  try {
    await api.startProcessRuralAffair(affairId)
    ElMessage.success('已开始处理该事务')
    loadMyHandledAffairs()
  } catch (error) {
    ElMessage.error(error?.response?.data?.error || '操作失败')
  }
}

// 查看事务详情（保留旧弹窗兼容）
const showAffairDetail = (affair) => {
  goAffairDetail(affair.id)
}

// 追答对话框相关
const showAnswerDialog = ref(false)
const answerContent = ref('')

const openAnswerDialog = (affair) => {
  currentAffair.value = affair
  answerContent.value = ''
  showAnswerDialog.value = true
}

const submitAnswer = async () => {
  if (!answerContent.value.trim()) {
    ElMessage.warning('追答内容不能为空')
    return
  }
  processing.value = true
  try {
    await api.addFollowUpAnswer(currentAffair.value.id, { content: answerContent.value, images: '' })
    ElMessage.success('追答成功')
    showAnswerDialog.value = false
    loadMyHandledAffairs()
  } catch (e) {
    ElMessage.error(e?.response?.data?.error || '追答失败')
  } finally {
    processing.value = false
  }
}

// 追问对话框
const showFollowUpDialog = ref(false)
const followUpContent = ref('')

const openFollowUpDialog = (affair) => {
  currentAffair.value = affair
  followUpContent.value = ''
  showFollowUpDialog.value = true
}

const submitFollowUp = async () => {
  if (!followUpContent.value.trim()) {
    ElMessage.warning('追问内容不能为空')
    return
  }
  processing.value = true
  try {
    await api.addFollowUpQuestion(currentAffair.value.id, { content: followUpContent.value, images: '' })
    ElMessage.success('追问成功')
    showFollowUpDialog.value = false
    loadMyAffairs()
  } catch (e) {
    ElMessage.error(e?.response?.data?.error || '追问失败')
  } finally {
    processing.value = false
  }
}

// 确认完成对话框
const showConfirmDialog = ref(false)

const openConfirmDialog = (affair) => {
  currentAffair.value = affair
  showConfirmDialog.value = true
}

const handleConfirmComplete = async () => {
  try {
    await ElMessageBox.confirm('确认该事务已处理完成？', '提示', { type: 'info' })
    await api.confirmCompleteAffair(currentAffair.value.id)
    ElMessage.success('已确认完成')
    showConfirmDialog.value = false
    loadMyAffairs()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(e?.response?.data?.error || '操作失败')
  }
}

// 打开处理对话框
const openProcessDialog = (affair) => {
  currentAffair.value = affair
  processForm.content = ''
  processForm.images = ''
  processImageList.value = []
  showProcessDialog.value = true
}

// 处理图片上传前校验
const beforeProcessUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) ElMessage.error('只能上传图片')
  if (!isLt5M) ElMessage.error('图片大小不能超过5MB')
  return isImage && isLt5M
}

// 处理图片上传成功
const handleProcessImageSuccess = (res) => {
  const newUrl = res.url
  if (processForm.images) {
    try {
      const currentImages = JSON.parse(processForm.images)
      processForm.images = JSON.stringify([...currentImages, newUrl])
    } catch {
      processForm.images = JSON.stringify([newUrl])
    }
  } else {
    processForm.images = JSON.stringify([newUrl])
  }
}

// 处理图片移除
const handleProcessImageRemove = () => {
  processForm.images = ''
}

// 提交处理结果
const submitProcess = async () => {
  if (!processForm.content) {
    ElMessage.warning('请填写处理详情')
    return
  }
  processing.value = true
  try {
    await api.processRuralAffair(currentAffair.value.id, {
      process_content: processForm.content,
      process_images: processForm.images
    })
    ElMessage.success('处理完成')
    showProcessDialog.value = false
    loadMyHandledAffairs()
  } catch (error) {
    ElMessage.error('提交失败')
  } finally {
    processing.value = false
  }
}

// 获取事务类型文本
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

// 获取事务状态文本
const getAffairStatusText = (status) => {
  const map = { 1: '待审核', 2: '审核通过', 3: '审核不通过', 4: '处理中', 5: '已完成' }
  return map[status] || '未知'
}

// 获取事务状态类型
const getAffairStatusType = (status) => {
  const map = { 1: 'warning', 2: 'primary', 3: 'danger', 4: 'info', 5: 'success' }
  return map[status] || ''
}

// 加载我的订单
const loadMyOrders = async () => {
  loadingOrders.value = true
  try {
    const res = await api.getMyOrders()
    let orders = res?.list || res || []
    if (orderStatus.value !== null && orderStatus.value !== '') {
      orders = orders.filter(o => o.status === orderStatus.value)
    }
    if (orderKeyword.value) {
      orders = orders.filter(o => o.items?.some(item => item.product_name?.includes(orderKeyword.value)) || o.order_no?.includes(orderKeyword.value))
    }
    myOrders.value = orders
  } catch (error) {
    console.error('加载订单失败:', error)
    myOrders.value = []
  } finally {
    loadingOrders.value = false
  }
}

// 订单付款
const handlePay = async (order) => {
  try {
    await api.payOrder(order.id)
    ElMessage.success('付款成功')
    loadMyOrders()
    showOrderDetailDialog.value = false
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '付款失败'
    ElMessage.error(msg)
  }
}

// 取消订单
const handleCancelOrder = async (id) => {
  try {
    await api.cancelOrder(id)
    ElMessage.success('订单已取消')
    loadMyOrders()
    showOrderDetailDialog.value = false
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '取消失败'
    ElMessage.error(msg)
  }
}

// 确认收货
const handleConfirmReceive = async (order) => {
  try {
    await api.confirmReceive(order.id)
    ElMessage.success('确认收货成功')
    loadMyOrders()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '确认收货失败'
    ElMessage.error(msg)
  }
}

// 加载我的销售订单
const loadMySellerOrders = async () => {
  loadingSellerOrders.value = true
  try {
    const res = await api.getMySellerOrders()
    let orders = res?.list || res || []
    if (sellerStatus.value !== null && sellerStatus.value !== '') {
      orders = orders.filter(o => o.status === sellerStatus.value)
    }
    mySellerOrders.value = orders
  } catch (error) {
    console.error('加载销售订单失败:', error)
    mySellerOrders.value = []
  } finally {
    loadingSellerOrders.value = false
  }
}

// 打开发货对话框
const openShipDialog = (order) => {
  shipForm.orderId = order.id
  shipForm.orderNo = order.order_no
  shipForm.receiver = order.receiver
  shipForm.phone = order.phone
  shipForm.address = order.address
  shipForm.expressNo = ''
  shipForm.remark = ''
  showShipDialog.value = true
}

// 确认发货
const handleShip = async () => {
  if (!shipForm.expressNo) {
    ElMessage.warning('请输入物流单号')
    return
  }
  shipping.value = true
  try {
    await api.shipOrder(shipForm.orderId, { express_no: shipForm.expressNo })
    ElMessage.success('发货成功')
    showShipDialog.value = false
    loadMySellerOrders()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || '发货失败'
    ElMessage.error(msg)
  } finally {
    shipping.value = false
  }
}

// 查看物流信息
const showExpressInfo = (order) => {
  currentExpressOrder.value = order
  showExpressDialog.value = true
}

// 打开商品发布对话框
const openProductDialog = () => {
  // 重置表单
  productForm.id = null
  productForm.title = ''
  productForm.type = 'seed'
  productForm.price = 0
  productForm.content = ''
  productForm.image_url = ''
  productForm.stock = 0
  productForm.address = ''
  productForm.status = 1
  productForm.publisher = null
  showProductDialog.value = true
}

// 打开商品编辑对话框
const openEditProductDialog = (product) => {
  productForm.id = product.id
  productForm.title = product.title
  productForm.type = product.type
  productForm.price = product.price
  productForm.content = product.content
  productForm.image_url = product.image_url
  productForm.stock = product.stock || 0
  productForm.address = product.address
  productForm.status = product.status
  productForm.publisher = product.publisher
  showProductDialog.value = true
}

// 商品图片上传前校验
const beforeProductUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) ElMessage.error('只能上传图片')
  if (!isLt5M) ElMessage.error('图片大小不能超过5MB')
  return isImage && isLt5M
}

// 商品图片上传成功
const handleProductUploadSuccess = (res) => {
  productForm.image_url = res.url
}

// 商品图片移除
const handleProductRemove = () => {
  productForm.image_url = ''
}

// 发布/编辑商品
const handlePublishProduct = async () => {
  if (!productForm.title) {
    ElMessage.warning('请输入商品名称')
    return
  }
  if (!productForm.content) {
    ElMessage.warning('请输入商品详情')
    return
  }
  publishingProduct.value = true
  try {
    const data = {
      title: productForm.title,
      type: productForm.type,
      price: productForm.price,
      content: productForm.content,
      image_url: productForm.image_url,
      stock: productForm.stock,
      address: productForm.address,
      status: productForm.status,
      publisher: productForm.publisher || currentUser.value?.id
    }
    if (productForm.id) {
      // 编辑
      await api.updateProduct({ id: productForm.id, ...data })
      ElMessage.success('商品更新成功')
    } else {
      // 新增
      await api.createProduct(data)
      ElMessage.success('商品发布成功')
    }
    showProductDialog.value = false
    loadMyProducts()
  } catch (error) {
    const msg = error?.response?.data?.error || error?.message || (productForm.id ? '更新失败' : '发布失败')
    ElMessage.error(msg)
  } finally {
    publishingProduct.value = false
  }
}

// 加载我的知识列表
const loadMyKnowledges = async () => {
  loadingKnowledge.value = true
  try {
    const res = await api.getMyKnowledge()
    myKnowledges.value = res || []
  } catch (error) {
    myKnowledges.value = []
  } finally {
    loadingKnowledge.value = false
  }
}

// 打开知识发布/编辑对话框（简洁模式）
const openKnowledgeDialog = (knowledge = null) => {
  if (knowledge) {
    knowledgeForm.id = knowledge.id
    knowledgeForm.title = knowledge.title
    knowledgeForm.content = knowledge.content
    knowledgeForm.image_url = knowledge.image_url || ''
  } else {
    knowledgeForm.id = null
    knowledgeForm.title = ''
    knowledgeForm.content = ''
    knowledgeForm.image_url = ''
  }
  showKnowledgeDialog.value = true
}

// 编辑知识 - 跳转到专业富文本编辑器
const editKnowledge = (knowledge) => {
  goToRichEditor('knowledge', knowledge.id)
}

// 知识封面上传前校验
const beforeKnowledgeUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) ElMessage.error('只能上传图片')
  if (!isLt5M) ElMessage.error('图片大小不能超过5MB')
  return isImage && isLt5M
}

// 知识封面上传成功
const handleKnowledgeUploadSuccess = (res) => {
  knowledgeForm.image_url = res.url
}

// 知识封面移除
const handleKnowledgeRemove = () => {
  knowledgeForm.image_url = ''
}

// 发布/保存知识
const handlePublishKnowledge = async () => {
  if (!knowledgeForm.title) {
    ElMessage.warning('请输入知识标题')
    return
  }
  if (!knowledgeForm.content) {
    ElMessage.warning('请输入知识内容')
    return
  }
  publishingKnowledge.value = true
  try {
    const knowledgeData = {
      title: knowledgeForm.title,
      type: 'planting',
      content: knowledgeForm.content,
      image_url: knowledgeForm.image_url || ''
    }

    if (knowledgeForm.id) {
      await api.updateKnowledge(knowledgeForm.id, knowledgeData)
      ElMessage.success('修改成功')
    } else {
      await api.createKnowledge(knowledgeData)
      ElMessage.success('发布成功')
    }
    showKnowledgeDialog.value = false
    loadMyKnowledges()
  } catch (error) {
    ElMessage.error('操作失败')
  } finally {
    publishingKnowledge.value = false
  }
}

// 删除知识
const handleDeleteKnowledge = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除这条知识吗？', '提示', { type: 'warning' })
    await api.deleteKnowledge(id)
    ElMessage.success('删除成功')
    loadMyKnowledges()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 获取订单状态类型
const getOrderStatusType = (status) => {
  const map = { 0: 'warning', 1: 'info', 2: 'primary', 3: 'success', 4: 'info' }
  return map[status] || ''
}

// 获取订单状态文本
const getOrderStatusText = (status) => {
  const map = { 0: '待付款', 1: '待发货', 2: '待收货', 3: '已完成', 4: '已取消' }
  return map[status] || '未知状态'
}

// 注销账号
const handleDeleteAccount = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要注销账号吗？注销后您将无法登录，但您的个人资料和数据将被保留。此操作不可撤销。',
      '注销确认',
      {
        confirmButtonText: '确认注销',
        cancelButtonText: '取消',
        type: 'warning',
        confirmButtonClass: 'el-button--danger'
      }
    )
    await api.deleteAccount()
    ElMessage.success('账号已注销')
    removeToken()
    router.push('/')
  } catch (error) {
    if (error !== 'cancel') {
      const msg = error?.response?.data?.error || error?.message || '注销失败'
      ElMessage.error(msg)
    }
  }
}

onMounted(() => {
  initUserInfo()
  loadUserDetail()
  loadMyAffairs()
  loadMyPosts()
  loadMyExpert()
  loadMyProducts()
  loadMyOrders()
  loadMySellerOrders()
  loadMyKnowledges()
  loadMyProcessor().then(() => {
    if (isProcessor.value) {
      loadMyHandledAffairs()
    }
  })
})
</script>

<style scoped>
.user-center { min-height: 100vh; background: #f5f5f5; }
.container { max-width: 1200px; margin: 0 auto; padding: 0 20px; }
.main { padding: 20px; }

.user-card .user-info {
  display: flex;
  align-items: flex-start;
  gap: 30px;
}

.avatar-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 15px;
}

.info-section {
  flex: 1;
}

.info-section h2 {
  margin: 0 0 10px 0;
  font-size: 24px;
  color: #333;
}

.username-hint {
  color: #999;
  font-size: 13px;
  margin: 0 0 10px 0;
}

.signature {
  color: #666;
  margin: 0 0 15px 0;
  font-size: 14px;
}

.signature.empty {
  color: #999;
  font-style: italic;
}

.meta-info {
  display: flex;
  gap: 20px;
  color: #666;
  font-size: 14px;
}

.meta-info span {
  display: flex;
  align-items: center;
  gap: 5px;
}

.action-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.card-header { display: flex; justify-content: space-between; align-items: center; }

.expert-empty {
  text-align: center;
  padding: 20px;
  color: #666;
}

.expert-empty p {
  margin-bottom: 15px;
}

.upload-area {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px;
}

.avatar-uploader {
  border: 2px dashed #d9d9d9;
  border-radius: 50%;
  width: 150px;
  height: 150px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: border-color 0.3s;
}

.avatar-uploader:hover {
  border-color: #409eff;
}

.avatar-preview {
  width: 146px;
  height: 146px;
  border-radius: 50%;
  object-fit: cover;
}

.avatar-uploader-icon {
  font-size: 28px;
  color: #8c939d;
}

.upload-tip {
  margin-top: 15px;
  color: #999;
  font-size: 12px;
}

.card-header { display: flex; justify-content: space-between; align-items: center; }
.header-actions { display: flex; align-items: center; }

.order-product { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

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
  color: #dcdfe6;
  margin-bottom: 20px;
}

.empty-text {
  color: #909399;
  font-size: 14px;
  margin-bottom: 20px;
}

/* 商品详情对话框 */
.product-detail h3 { margin-bottom: 10px; }
.product-detail .price { color: #f56c6c; font-size: 20px; font-weight: bold; margin: 10px 0; }
.product-detail .info { color: #666; margin-bottom: 15px; }

/* 订单详情对话框 */
.order-item-detail {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid #eee;
}
.order-item-detail:last-child { border-bottom: none; }
.order-item-detail span { color: #333; }
.empty-items { color: #999; text-align: center; padding: 20px; }

/* 事务处理人员 */
.processor-tip { color: #999; font-size: 13px; margin-bottom: 10px; }
.upload-tip { font-size: 12px; color: #999; margin-top: 5px; }

/* 卖家订单 */
.buyer-phone { color: #999; font-size: 12px; }
.ship-address { color: #666; font-size: 13px; margin-top: 5px; }

/* 事务详情 */
.affair-detail { padding: 10px 0; }
.affair-content { color: #333; line-height: 1.8; white-space: pre-wrap; }
.affair-images { display: flex; flex-wrap: wrap; gap: 10px; margin-top: 15px; }
.affair-image { width: 120px; height: 120px; border-radius: 4px; }
.affair-video { margin-top: 15px; }
.affair-video-player { width: 100%; max-height: 300px; border-radius: 4px; }
.reject-info { margin-top: 15px; }
.process-info { margin-top: 15px; }
.process-images { display: flex; flex-wrap: wrap; gap: 10px; margin-top: 15px; }
.process-image { width: 100px; height: 100px; border-radius: 4px; }
.completed-info { margin-top: 15px; }

/* 富文本编辑器样式 */
.editor-form-item :deep(.el-form-item__content) {
  line-height: normal;
}

.editor-wrapper {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: hidden;
}

.editor-wrapper :deep(.ql-container) {
  min-height: 250px;
  font-size: 15px;
}

.editor-wrapper :deep(.ql-editor) {
  min-height: 250px;
  line-height: 1.8;
}

.editor-wrapper :deep(.ql-editor.ql-blank::before) {
  color: #c0c4cc;
  font-style: normal;
  left: 16px;
  right: 16px;
}

.mode-tip {
  color: #909399;
  font-size: 12px;
  margin-left: 10px;
}

.video-preview {
  margin-top: 10px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  align-items: flex-start;
}
</style>
