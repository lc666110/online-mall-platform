<script setup lang="ts">
import { onMounted } from "vue";
import { api } from "./api";
import { useRouter } from "vue-router";
const router = useRouter();
onMounted(() => api.refresh().catch(() => api.clear()));
async function logout() {
  try {
    await api.logout();
    router.push("/");
  } catch (e) {
    alert(e instanceof Error ? e.message : "退出失败");
  }
}
</script>
<template>
  <div class="announcement">认真挑选每一件好物 · 让日常，多一点喜欢</div>
  <header class="header">
    <RouterLink to="/" class="brand">MORI<span>生活好物</span></RouterLink>
    <nav>
      <RouterLink to="/">探索好物</RouterLink
      ><RouterLink to="/orders">我的订单</RouterLink
      ><RouterLink to="/cart">购物袋</RouterLink
      ><RouterLink v-if="!api.token.value" to="/login">登录 / 注册</RouterLink
      ><template v-else
        ><RouterLink to="/profile">{{
          api.user.value?.name ?? "个人中心"
        }}</RouterLink
        ><button class="text-button" @click="logout">退出</button></template
      >
    </nav>
  </header>
  <main class="container"><RouterView /></main>
  <footer>
    <div class="brand">MORI<span>把喜欢带回家</span></div>
    <p>精选生活好物，让购物简单而安心。</p>
    <small>© {{ new Date().getFullYear() }} MORI · online-mall-platform</small>
  </footer>
</template>
