import { createRouter, createWebHistory } from "vue-router";
import { api } from "./api";
export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", component: () => import("./views/Home.vue") },
    { path: "/products/:id", component: () => import("./views/Product.vue") },
    { path: "/login", component: () => import("./views/Auth.vue") },
    {
      path: "/cart",
      component: () => import("./views/Cart.vue"),
      meta: { auth: true },
    },
    {
      path: "/orders",
      component: () => import("./views/Orders.vue"),
      meta: { auth: true },
    },
    {
      path: "/orders/:id",
      component: () => import("./views/OrderDetail.vue"),
      meta: { auth: true },
    },
    {
      path: "/profile",
      component: () => import("./views/Profile.vue"),
      meta: { auth: true },
    },
    { path: "/:pathMatch(.*)*", redirect: "/" },
  ],
  scrollBehavior: () => ({ top: 0 }),
});
router.beforeEach((to) => {
  if (to.meta.auth && !api.token.value)
    return { path: "/login", query: { redirect: to.fullPath } };
});
