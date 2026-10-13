<script setup lang="ts">
import { ref, onMounted } from "vue";
import { api } from "../api";
import type { User, AfterSale, Page } from "../../../shared/types";
import { statusLabel } from "../../../shared/helpers";
const name = ref(""),
  phone = ref(""),
  avatar = ref(""),
  message = ref(""),
  error = ref(""),
  busy = ref(false),
  sales = ref<AfterSale[]>([]);
onMounted(async () => {
  try {
    await api.refresh();
    name.value = api.user.value?.name ?? "";
    phone.value = api.user.value?.phone ?? "";
    avatar.value = api.user.value?.avatar ?? "";
    sales.value = (await api.request<Page<AfterSale>>("/after-sales")).items;
  } catch (e) {
    error.value = (e as Error).message;
  }
});
async function save() {
  busy.value = true;
  error.value = "";
  message.value = "";
  try {
    api.user.value = await api.request<User>("/me", "PATCH", {
      name: name.value,
      phone: phone.value,
      avatar: avatar.value,
    });
    message.value = "资料已保存";
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    busy.value = false;
  }
}
</script>
<template>
  <div class="page-title">
    <p class="eyebrow">MY ACCOUNT</p>
    <h1>个人中心</h1>
    <p>{{ api.user.value?.email }}</p>
  </div>
  <div class="two-columns">
    <form class="panel" @submit.prevent="save">
      <h2>个人资料</h2>
      <label>昵称<input v-model="name" required maxlength="80" /></label
      ><label>联系电话<input v-model="phone" maxlength="30" /></label
      ><label
        >头像地址<input
          v-model="avatar"
          type="url"
          placeholder="https://…"
          maxlength="500"
      /></label>
      <p v-if="message" role="status" class="success">{{ message }}</p>
      <p v-if="error" role="alert" class="error">{{ error }}</p>
      <button class="button" :disabled="busy">保存资料</button>
    </form>
    <section class="panel">
      <h2>我的售后</h2>
      <p v-if="!sales.length" class="muted">暂无售后申请。</p>
      <article v-for="s in sales" :key="s.id" class="sale-row">
        <RouterLink :to="`/orders/${s.order_id}`"
          >订单 {{ s.order?.number }}</RouterLink
        ><span class="badge">{{ statusLabel(s.status) }}</span>
        <p>{{ s.reason }}</p>
        <small>{{ s.note }}</small>
      </article>
    </section>
  </div>
</template>
