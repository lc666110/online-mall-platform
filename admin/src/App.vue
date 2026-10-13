<script setup lang="ts">
import {
  NConfigProvider,
  NMessageProvider,
  NDialogProvider,
  NButton,
  NTag,
  zhCN,
  dateZhCN,
} from "naive-ui";
import { useRoute, useRouter } from "vue-router";
import { api } from "./api";
const route = useRoute(),
  router = useRouter();
const links = [
  ["/", "数据概览", "◈"],
  ["/products", "商品管理", "▦"],
  ["/categories", "分类管理", "▤"],
  ["/orders", "订单与发货", "▣"],
  ["/after-sales", "售后管理", "◇"],
  ["/users", "用户与权限", "◎"],
];
async function logout() {
  try {
    await api.logout();
    router.push("/login");
  } catch (e) {
    alert((e as Error).message);
  }
}
</script>
<template>
  <NConfigProvider
    :locale="zhCN"
    :date-locale="dateZhCN"
    :theme-overrides="{
      common: {
        primaryColor: '#28684f',
        primaryColorHover: '#348362',
        primaryColorPressed: '#174b37',
        borderRadius: '7px',
      },
    }"
    ><NMessageProvider
      ><NDialogProvider
        ><RouterView v-if="route.path === '/login'" />
        <div v-else class="admin-layout">
          <aside class="sidebar">
            <RouterLink to="/" class="admin-brand"
              >MORI<span>COMMERCE ADMIN</span></RouterLink
            >
            <div class="workspace-label">商城工作台</div>
            <nav>
              <template v-for="[path, label, icon] in links" :key="path"
                ><RouterLink
                  v-if="path !== '/users' || api.user.value?.role === 'admin'"
                  :to="path"
                  :class="{ active: route.path === path }"
                  ><span>{{ icon }}</span
                  >{{ label }}</RouterLink
                ></template
              >
            </nav>
            <div class="sidebar-bottom">
              ONLINE MALL PLATFORM<br /><small>让经营更有条理</small>
            </div>
          </aside>
          <div class="admin-main">
            <header class="admin-header">
              <span
                >商城运营 /
                {{ links.find((l) => l[0] === route.path)?.[1] }}</span
              >
              <div>
                <NTag size="small" type="success" :bordered="false">{{
                  api.user.value?.role === "admin" ? "管理员" : "运营人员"
                }}</NTag
                ><span>{{ api.user.value?.name }}</span
                ><NButton size="small" quaternary @click="logout"
                  >退出登录</NButton
                >
              </div>
            </header>
            <main class="admin-content"><RouterView /></main>
          </div></div></NDialogProvider></NMessageProvider
  ></NConfigProvider>
</template>
