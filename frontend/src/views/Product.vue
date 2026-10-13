<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { api } from "../api";
import type { Product, Review, Page } from "../../../shared/types";
import { images, money } from "../../../shared/helpers";
import StateMessage from "../components/StateMessage.vue";
const route = useRoute(),
  router = useRouter();
const product = ref<Product | null>(null),
  reviews = ref<Review[]>([]),
  reviewPage = ref(1),
  reviewTotal = ref(0),
  reviewLoading = ref(false),
  variantID = ref(0),
  quantity = ref(1),
  selectedImage = ref(0),
  busy = ref(false),
  loading = ref(false),
  error = ref(""),
  message = ref("");
const variant = computed(() =>
  product.value?.variants.find((v) => v.id === variantID.value),
);
const gallery = computed(() => images(product.value?.images ?? "[]"));
watch(
  () => route.params.id,
  async () => {
    loading.value = true;
    error.value = "";
    message.value = "";
    try {
      product.value = await api.request<Product>(
        `/products/${route.params.id}`,
      );
      variantID.value = product.value.variants[0]?.id ?? 0;
      selectedImage.value = 0;
      reviewPage.value = 1;
      await loadReviews();
    } catch (e) {
      error.value = (e as Error).message;
    } finally {
      loading.value = false;
    }
  },
  { immediate: true },
);
async function loadReviews() {
  reviewLoading.value = true;
  try {
    const data = await api.request<Page<Review>>(
      `/products/${route.params.id}/reviews?page=${reviewPage.value}`,
    );
    reviews.value = data.items;
    reviewTotal.value = data.total;
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    reviewLoading.value = false;
  }
}
async function add() {
  if (!api.token.value) {
    router.push({ path: "/login", query: { redirect: route.fullPath } });
    return;
  }
  busy.value = true;
  error.value = "";
  try {
    await api.request("/cart", "POST", {
      variant_id: variantID.value,
      quantity: quantity.value,
    });
    message.value = "已加入购物袋";
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    busy.value = false;
  }
}
</script>
<template>
  <StateMessage :loading="loading" :error="error" /><template
    v-if="product && !loading"
    ><p class="breadcrumbs">
      <RouterLink to="/">探索好物</RouterLink> / {{ product.category.name }} /
      {{ product.name }}
    </p>
    <div class="product-detail">
      <div>
        <div class="detail-image">
          <img
            v-if="gallery[selectedImage]"
            :src="gallery[selectedImage]"
            :alt="product.name"
          /><span v-else class="image-placeholder">MORI</span>
        </div>
        <div class="thumbnails">
          <button
            v-for="(img, i) in gallery"
            :key="img"
            :class="{ selected: selectedImage === i }"
            @click="selectedImage = i"
          >
            <img :src="img" :alt="`商品图片 ${i + 1}`" />
          </button>
        </div>
      </div>
      <section class="detail-copy">
        <p class="eyebrow">{{ product.category.name }} / MORI COLLECTION</p>
        <h1>{{ product.name }}</h1>
        <p class="description">{{ product.description }}</p>
        <strong class="detail-price">{{ money(variant?.price ?? 0) }}</strong>
        <p class="muted">规格选择</p>
        <div class="variant-options">
          <button
            v-for="v in product.variants"
            :key="v.id"
            :class="{ selected: variantID === v.id }"
            @click="
              variantID = v.id;
              quantity = 1;
            "
          >
            {{ v.specification }}
          </button>
        </div>
        <p>库存：{{ variant?.stock ?? 0 }} 件</p>
        <label class="quantity-label"
          >数量<input
            v-model.number="quantity"
            type="number"
            min="1"
            :max="Math.min(99, variant?.stock ?? 0)" /></label
        ><button
          class="button full"
          :disabled="
            busy ||
            !variant?.stock ||
            quantity < 1 ||
            quantity > Math.min(99, variant?.stock ?? 0)
          "
          @click="add"
        >
          {{ variant?.stock ? "加入购物袋" : "暂时售罄" }} ↗
        </button>
        <p v-if="message" class="success" role="status">
          {{ message }} · <RouterLink to="/cart">去结算</RouterLink>
        </p>
        <div class="detail-guarantees">
          <span>规格清晰</span><span>订单可追踪</span><span>在线售后</span>
        </div>
      </section>
    </div>
    <section class="reviews">
      <div class="section-title">
        <h2>真实评价</h2>
        <span>{{ reviewTotal }} 条评价</span>
      </div>
      <p v-if="!reviews.length" class="muted">还没有评价，期待你的体验分享。</p>
      <article v-for="r in reviews" :key="r.id" class="review">
        <div>
          <b>{{ r.user.name }}</b
          ><span class="stars"
            >{{ "★".repeat(r.rating) }}{{ "☆".repeat(5 - r.rating) }}</span
          >
        </div>
        <p>{{ r.content }}</p>
        <small>{{ new Date(r.created_at).toLocaleDateString() }}</small>
      </article>
      <div v-if="reviewTotal > 12" class="pagination">
        <button
          :disabled="reviewPage === 1 || reviewLoading"
          @click="
            reviewPage--;
            loadReviews();
          "
        >
          上一页评价
        </button>
        <span>{{ reviewPage }} / {{ Math.ceil(reviewTotal / 12) }}</span>
        <button
          :disabled="reviewPage * 12 >= reviewTotal || reviewLoading"
          @click="
            reviewPage++;
            loadReviews();
          "
        >
          下一页评价
        </button>
      </div>
    </section></template
  >
</template>
