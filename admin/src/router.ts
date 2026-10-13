import { createRouter, createWebHistory } from "vue-router";
import { api } from "./api";
export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/login", component: () => import("./views/Login.vue") },
    {
      path: "/",
      component: () => import("./views/Dashboard.vue"),
      meta: { auth: true },
    },
    {
      path: "/products",
      component: () => import("./views/Products.vue"),
      meta: { auth: true },
    },
    {
      path: "/categories",
      component: () => import("./views/Categories.vue"),
      meta: { auth: true },
    },
    {
      path: "/orders",
      component: () => import("./views/Orders.vue"),
      meta: { auth: true },
    },
    {
      path: "/users",
      component: () => import("./views/Users.vue"),
      meta: { auth: true, admin: true },
    },
    {
      path: "/after-sales",
      component: () => import("./views/AfterSales.vue"),
      meta: { auth: true },
    },
    { path: "/:pathMatch(.*)*", redirect: "/" },
  ],
});
router.beforeEach(async (to) => {
  if (!to.meta.auth) return;
  if (!api.token.value) return "/login";
  try {
    await api.refresh();
  } catch {
    return "/login";
  }
  if (!["admin", "staff"].includes(api.user.value?.role ?? "")) {
    api.clear();
    return "/login";
  }
  if (to.meta.admin && api.user.value?.role !== "admin") return "/";
});
