<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import { api } from "../api";
import type { CartItem, Order, Product } from "../../../shared/types";
import { money } from "../../../shared/helpers";
import StateMessage from "../components/StateMessage.vue";
const items = ref<CartItem[]>([]),
  names = ref<Record<number, string>>({}),
  loading = ref(false),
  busy = ref(false),
  error = ref(""),
  recipient = ref(""),
  phone = ref(""),
  address = ref("");
const router = useRouter();
let checkoutKey = Array.from(crypto.getRandomValues(new Uint8Array(24)), (x) =>
  x.toString(16).padStart(2, "0"),
).join("");
const total = computed(() =>
  items.value.reduce((sum, i) => sum + i.variant.price * i.quantity, 0),
);
async function load() {
  loading.value = true;
  try {
    items.value = await api.request<CartItem[]>("/cart");
    await Promise.all(
      items.value.map(async (i) => {
        try {
          const p = await api.request<Product>(
            `/products/${i.variant.product_id}`,
          );
          names.value[p.id] = p.name;
        } catch {
          names.value[i.variant.product_id] = "商品暂不可售";
        }
      }),
    );
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    loading.value = false;
  }
}
async function update(i: CartItem) {
  busy.value = true;
  error.value = "";
  try {
    await api.request(`/cart/${i.id}`, "PATCH", { quantity: i.quantity });
    checkoutKey = Array.from(crypto.getRandomValues(new Uint8Array(24)), (x) =>
      x.toString(16).padStart(2, "0"),
    ).join("");
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    await load();
    busy.value = false;
  }
}
async function remove(i: CartItem) {
  busy.value = true;
  try {
    await api.request(`/cart/${i.id}`, "DELETE");
    checkoutKey = Array.from(crypto.getRandomValues(new Uint8Array(24)), (x) =>
      x.toString(16).padStart(2, "0"),
    ).join("");
    await load();
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    busy.value = false;
  }
}
async function checkout() {
  busy.value = true;
  error.value = "";
  try {
    const o = await api.request<Order>(
      "/orders",
      "POST",
      {
        recipient: recipient.value,
        phone: phone.value,
        address: address.value,
      },
      checkoutKey,
    );
    router.push(`/orders/${o.id}`);
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
    <p class="eyebrow">YOUR SHOPPING BAG</p>
    <h1>喜欢的，都在这里。</h1>
  </div>
  <StateMessage
    :loading="loading"
    :error="error"
    :empty="!loading && !items.length"
  />
  <div v-if="items.length" class="cart-layout">
    <section class="panel">
      <article v-for="i in items" :key="i.id" class="cart-row">
        <div>
          <RouterLink :to="`/products/${i.variant.product_id}`"
            ><h3>
              {{ names[i.variant.product_id] ?? i.variant.sku }}
            </h3></RouterLink
          >
          <p class="muted">{{ i.variant.specification }}</p>
          <b>{{ money(i.variant.price) }}</b>
        </div>
        <div>
          <label
            >数量<input
              v-model.number="i.quantity"
              type="number"
              min="1"
              :max="Math.min(99, i.variant.stock)"
              :disabled="busy"
              @change="update(i)" /></label
          ><button
            class="text-button danger"
            :disabled="busy"
            @click="remove(i)"
          >
            移除
          </button>
        </div>
        <strong>{{ money(i.variant.price * i.quantity) }}</strong>
      </article>
    </section>
    <form class="panel checkout-panel" @submit.prevent="checkout">
      <h2>订单结算</h2>
      <label
        >收货人<input
          v-model="recipient"
          required
          maxlength="80"
          autocomplete="name" /></label
      ><label
        >联系电话<input
          v-model="phone"
          required
          maxlength="30"
          autocomplete="tel" /></label
      ><label
        >收货地址<textarea
          v-model="address"
          required
          maxlength="500"
          autocomplete="street-address"
        ></textarea>
      </label>
      <div class="total-line">
        <span>合计</span><strong>{{ money(total) }}</strong>
      </div>
      <small>提交后由商城确认订单；最终价格以服务端库存与价格校验为准。</small
      ><button class="button full" :disabled="busy">
        {{ busy ? "正在提交…" : "提交订单" }}
      </button>
    </form>
  </div>
</template>
