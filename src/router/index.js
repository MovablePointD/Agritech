import { createRouter, createWebHistory } from 'vue-router'
import Home from '@/views/Home.vue'
import Login from '@/views/Login.vue'
import Register from '@/views/Register.vue'
import ResetPassword from '@/views/ResetPassword.vue'
import ProductList from '@/views/ProductList.vue'
import ProductDetail from '@/views/ProductDetail.vue'
import PostList from '@/views/PostList.vue'
import PostDetail from '@/views/PostDetail.vue'
import ExpertList from '@/views/ExpertList.vue'
import ExpertDetail from '@/views/ExpertDetail.vue'
import UserCenter from '@/views/UserCenter.vue'
import AddressManage from '@/views/AddressManage.vue'
import OrderList from '@/views/OrderList.vue'
import KnowledgeList from '@/views/KnowledgeList.vue'
import KnowledgeDetail from '@/views/KnowledgeDetail.vue'
import Admin from '@/views/Admin.vue'
import Reviewer from '@/views/Reviewer.vue'
import Cart from '@/views/Cart.vue'
import MyKnowledge from '@/views/MyKnowledge.vue'
import RuralAffairList from '@/views/RuralAffairList.vue'
import RuralAffairDetail from '@/views/RuralAffairDetail.vue'
import RuralAffairSubmit from '@/views/RuralAffairSubmit.vue'
import MyAffairs from '@/views/MyAffairs.vue'
import Notifications from '@/views/Notifications.vue'
import AffairProcessor from '@/views/AffairProcessor.vue'
import AIChat from '@/views/AIChat.vue'
import RichEditor from '@/views/RichEditor.vue'
import Messages from '@/views/Messages.vue'
import MessageDetail from '@/views/MessageDetail.vue'
import RuralInfoList from '@/views/RuralInfoList.vue'
import RuralInfoDetail from '@/views/RuralInfoDetail.vue'
import PolicyNoticeList from '@/views/PolicyNoticeList.vue'
import PolicyNoticeDetail from '@/views/PolicyNoticeDetail.vue'
import RuralHub from '@/views/RuralHub.vue'

const routes = [
  { path: '/', name: 'Home', component: Home },
  { path: '/login', name: 'Login', component: Login },
  { path: '/register', name: 'Register', component: Register },
  { path: '/reset-password', name: 'ResetPassword', component: ResetPassword },
  { path: '/products', name: 'ProductList', component: ProductList },
  { path: '/product/:id', name: 'ProductDetail', component: ProductDetail },
  { path: '/posts', name: 'PostList', component: PostList },
  { path: '/post/:id', name: 'PostDetail', component: PostDetail },
  { path: '/experts', name: 'ExpertList', component: ExpertList },
  { path: '/expert/:id', name: 'ExpertDetail', component: ExpertDetail },
  { path: '/knowledge', name: 'KnowledgeList', component: KnowledgeList },
  { path: '/knowledge/:id', name: 'KnowledgeDetail', component: KnowledgeDetail },
  { path: '/user', name: 'UserCenter', component: UserCenter },
  { path: '/address', name: 'AddressManage', component: AddressManage },
  { path: '/orders', name: 'OrderList', component: OrderList },
  { path: '/cart', name: 'Cart', component: Cart },
  { path: '/my-knowledge', name: 'MyKnowledge', component: MyKnowledge },
  { path: '/admin', name: 'Admin', component: Admin },
  { path: '/reviewer', name: 'Reviewer', component: Reviewer },
  { path: '/affairs', name: 'RuralAffairList', component: RuralAffairList },
  { path: '/affair/:id', name: 'RuralAffairDetail', component: RuralAffairDetail },
  { path: '/affair/submit', name: 'RuralAffairSubmit', component: RuralAffairSubmit },
  { path: '/my-affairs', name: 'MyAffairs', component: MyAffairs },
  { path: '/notifications', name: 'Notifications', component: Notifications },
  { path: '/processor', name: 'AffairProcessor', component: AffairProcessor },
  { path: '/ai-chat', name: 'AIChat', component: AIChat },
  { path: '/rich-editor/:contentType', name: 'RichEditor', component: RichEditor },
  { path: '/messages', name: 'Messages', component: Messages },
  { path: '/messages/:id', name: 'MessageDetail', component: MessageDetail },
  { path: '/rural-info', name: 'RuralHub', component: RuralHub },                          // 统一信息中心
  { path: '/rural-info/:id', name: 'RuralInfoDetail', component: RuralInfoDetail },          // 信息详情
  { path: '/policies', redirect: '/rural-info' },                                           // 重定向到信息中心
  { path: '/policy/:id', name: 'PolicyNoticeDetail', component: PolicyNoticeDetail },        // 公告详情
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
