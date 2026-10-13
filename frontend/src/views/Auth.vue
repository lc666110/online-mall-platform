<script setup lang="ts">
import { ref } from "vue";
import { useRouter, useRoute } from "vue-router";
import { api } from "../api";
const registering = ref(false),
  email = ref(""),
  password = ref(""),
  name = ref(""),
  busy = ref(false),
  error = ref("");
const router = useRouter(),
  route = useRoute();
async function submit() {
  busy.value = true;
  error.value = "";
  try {
    if (registering.value)
      await api.request("/auth/register", "POST", {
        email: email.value,
        password: password.value,
        name: name.value,
      });
    await api.login(email.value, password.value);
    const target =
      typeof route.query.redirect === "string" &&
      route.query.redirect.startsWith("/") &&
      !route.query.redirect.startsWith("//")
        ? route.query.redirect
        : "/";
    router.push(target);
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    busy.value = false;
  }
}
</script>
<template>
  <section class="auth-layout">
    <div class="auth-story">
      <p class="eyebrow">WELCOME TO MORI</p>
      <h1>把喜欢，<br />留在日常里。</h1>
      <p>加入我们，发现属于你的生活好物。</p>
      <span class="auth-art">mori.</span>
    </div>
    <form class="panel auth-form" @submit.prevent="submit">
      <p class="eyebrow">YOUR MORI ACCOUNT</p>
      <h2>{{ registering ? "创建你的账号" : "欢迎回来" }}</h2>
      <div class="tabs">
        <button
          type="button"
          :class="{ selected: !registering }"
          @click="registering = false"
        >
          登录</button
        ><button
          type="button"
          :class="{ selected: registering }"
          @click="registering = true"
        >
          创建账号
        </button>
      </div>
      <label v-if="registering"
        >昵称<input
          v-model="name"
          required
          maxlength="80"
          autocomplete="nickname" /></label
      ><label
        >邮箱<input
          v-model="email"
          type="email"
          required
          maxlength="191"
          autocomplete="email" /></label
      ><label
        >密码<input
          v-model="password"
          type="password"
          required
          :minlength="registering ? 10 : 1"
          maxlength="72"
          :autocomplete="
            registering ? 'new-password' : 'current-password'
          " /></label
      ><small v-if="registering"
        >密码至少 10 位，建议同时使用字母、数字和符号。</small
      >
      <p v-if="error" class="error" role="alert">{{ error }}</p>
      <button class="button full" :disabled="busy">
        {{ busy ? "请稍候…" : registering ? "注册并登录" : "登录" }}
      </button>
    </form>
  </section>
</template>
