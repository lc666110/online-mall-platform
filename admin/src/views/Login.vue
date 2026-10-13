<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import { NCard, NInput, NButton, NAlert } from "naive-ui";
import { api } from "../api";
const email = ref(""),
  password = ref(""),
  busy = ref(false),
  error = ref("");
const router = useRouter();
async function login() {
  busy.value = true;
  error.value = "";
  try {
    const u = await api.login(email.value, password.value);
    if (!["admin", "staff"].includes(u.role)) {
      await api.logout();
      throw new Error("此账号没有后台权限");
    }
    router.push("/");
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    busy.value = false;
  }
}
</script>
<template>
  <div class="admin-login">
    <section>
      <p class="admin-brand">MORI<span>COMMERCE ADMIN</span></p>
      <h1>每一笔好生意，<br />都从这里开始。</h1>
      <p>商品、订单、用户与数据，一站式管理。</p>
    </section>
    <NCard class="login-card" title="登录商城工作台"
      ><form @submit.prevent="login">
        <label
          >管理员邮箱<NInput
            v-model:value="email"
            placeholder="管理员邮箱" /></label
        ><label
          >密码<NInput
            v-model:value="password"
            type="password"
            show-password-on="click"
            placeholder="密码" /></label
        ><NAlert v-if="error" type="error" style="margin: 16px 0">{{
          error
        }}</NAlert
        ><NButton
          attr-type="submit"
          type="primary"
          block
          :loading="busy"
          style="margin-top: 24px"
          >登录工作台</NButton
        >
        <p class="muted" style="font-size: 12px; margin-top: 20px">
          使用部署时配置的管理员账号登录。
        </p>
      </form></NCard
    >
  </div>
</template>
