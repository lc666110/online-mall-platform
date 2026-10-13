<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRoute } from "vue-router";
import { api } from "../api";
import type { Order } from "../../../shared/types";
import { money, statusLabel } from "../../../shared/helpers";
const route = useRoute();
const order = ref<Order | null>(null),
  error = ref(""),
  message = ref(""),
  busy = ref(false),
  loading = ref(false),
  reviewItem = ref(0),
  rating = ref(5),
  content = ref(""),
  reason = ref(""),
  showSale = ref(false);
async function load() {
  loading.value = true;
  error.value = "";
  try {
    order.value = await api.request<Order>(`/orders/${route.params.id}`);
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    loading.value = false;
  }
}
async function transition(status: string) {
  busy.value = true;
  error.value = "";
  try {
    order.value = await api.request<Order>(
      `/orders/${order.value?.id}/status`,
      "PATCH",
      { status },
    );
    message.value = "订单状态已更新";
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    busy.value = false;
  }
}
async function review() {
  busy.value = true;
  error.value = "";
  try {
    await api.request("/reviews", "POST", {
      order_item_id: reviewItem.value,
      rating: rating.value,
      content: content.value,
    });
    message.value = "感谢你的评价";
    reviewItem.value = 0;
    content.value = "";
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    busy.value = false;
  }
}
async function sale() {
  busy.value = true;
  error.value = "";
  try {
    await api.request(`/orders/${order.value?.id}/after-sales`, "POST", {
      reason: reason.value,
    });
    message.value = "售后申请已提交，可在个人中心查看进度";
    showSale.value = false;
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    busy.value = false;
  }
}
onMounted(load);
</script>
<template>
  <div class="page-title">
    <p class="eyebrow">ORDER DETAILS</p>
    <h1>订单详情</h1>
  </div>
  <p v-if="error" class="error" role="alert">{{ error }}</p>
  <p v-if="loading" class="muted" role="status">正在加载订单…</p>
  <p v-if="message" class="success" role="status">{{ message }}</p>
  <template v-if="order"
    ><section class="panel">
      <div class="order-heading">
        <h2>{{ order.number }}</h2>
        <span class="badge">{{ statusLabel(order.status) }}</span>
      </div>
      <div class="order-progress">
        <span :class="{ done: true }">提交订单</span
        ><span
          :class="{
            done: ['confirmed', 'shipped', 'completed'].includes(order.status),
          }"
          >订单确认</span
        ><span
          :class="{ done: ['shipped', 'completed'].includes(order.status) }"
          >物流发货</span
        ><span :class="{ done: order.status === 'completed' }">确认收货</span>
      </div>
      <article v-for="i in order.items" :key="i.id" class="cart-row">
        <div>
          <RouterLink :to="`/products/${i.product_id}`"
            ><h3>{{ i.name }}</h3></RouterLink
          >
          <p>{{ i.specification }} × {{ i.quantity }}</p>
        </div>
        <b>{{ money(i.price * i.quantity) }}</b
        ><button
          v-if="order.status === 'completed'"
          class="button secondary"
          @click="reviewItem = i.id"
        >
          评价商品
        </button>
      </article>
      <div class="total-line">
        <span>订单合计</span><strong>{{ money(order.total) }}</strong>
      </div>
      <div class="actions">
        <button
          v-if="order.status === 'pending'"
          class="button secondary"
          :disabled="busy"
          @click="transition('cancelled')"
        >
          取消订单</button
        ><button
          v-if="order.status === 'shipped'"
          class="button"
          :disabled="busy"
          @click="transition('completed')"
        >
          确认收货</button
        ><button
          v-if="['shipped', 'completed'].includes(order.status)"
          class="button secondary"
          @click="showSale = true"
        >
          申请售后
        </button>
      </div>
    </section>
    <div class="two-columns">
      <section class="panel">
        <h2>收货信息</h2>
        <p>{{ order.recipient }} · {{ order.phone }}</p>
        <p>{{ order.address }}</p>
      </section>
      <section class="panel">
        <h2>物流信息</h2>
        <p>{{ order.carrier || "等待发货" }}</p>
        <p>{{ order.tracking || "发货后将显示运单号" }}</p>
      </section>
    </div>
    <form v-if="reviewItem" class="panel" @submit.prevent="review">
      <h2>分享你的使用体验</h2>
      <label
        >评分<select v-model.number="rating">
          <option v-for="n in 5" :key="n" :value="n">{{ n }} 星</option>
        </select></label
      ><label
        >评价内容<textarea
          v-model="content"
          required
          maxlength="2000"
        ></textarea>
      </label>
      <div class="actions">
        <button class="button" :disabled="busy">提交评价</button
        ><button type="button" class="button secondary" @click="reviewItem = 0">
          取消
        </button>
      </div>
    </form>
    <form v-if="showSale" class="panel" @submit.prevent="sale">
      <h2>申请售后</h2>
      <label
        >申请原因<textarea
          v-model="reason"
          required
          maxlength="1000"
        ></textarea>
      </label>
      <div class="actions">
        <button class="button" :disabled="busy">提交售后申请</button
        ><button
          type="button"
          class="button secondary"
          @click="showSale = false"
        >
          取消
        </button>
      </div>
    </form></template
  >
</template>
