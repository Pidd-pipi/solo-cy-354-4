import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import Products from '../pages/Products.vue'
import Publish from '../pages/Publish.vue'
import Messages from '../pages/Messages.vue'
import Orders from '../pages/Orders.vue'
import BookExchange from '../pages/BookExchange.vue'
import Graduation from '../pages/Graduation.vue'
import Profile from '../pages/Profile.vue'
import Login from '../pages/Login.vue'
import Register from '../pages/Register.vue'

const routes: RouteRecordRaw[] = [
  { path: '/', redirect: '/products' },
  { path: '/products', component: Products },
  { path: '/publish', component: Publish, meta: { requiresAuth: true } },
  { path: '/messages', component: Messages, meta: { requiresAuth: true } },
  { path: '/orders', component: Orders, meta: { requiresAuth: true } },
  { path: '/book-exchange', component: BookExchange },
  { path: '/graduation', component: Graduation },
  { path: '/profile', component: Profile, meta: { requiresAuth: true } },
  { path: '/login', component: Login },
  { path: '/register', component: Register },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
