<script setup lang="ts">
import { ref, onMounted, watch } from "vue";
import { api } from "../api";
import type { Product, Category, Page } from "../../../shared/types";
import ProductCard from "../components/ProductCard.vue";
import StateMessage from "../components/StateMessage.vue";
const products = ref<Product[]>([]),
  categories = ref<Category[]>([]),
  query = ref(""),
  category = ref(0),
  page = ref(1),
  total = ref(0),
  loading = ref(false),
  error = ref("");
let sequence = 0;
async function load() {
  const seq = ++sequence;
  loading.value = true;
  error.value = "";
  try {
    const r = await api.request<Page<Product>>(
      `/products?q=${encodeURIComponent(query.value)}&category_id=${category.value}&page=${page.value}`,
    );
    if (seq === sequence) {
      products.value = r.items;
      total.value = r.total;
    }
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    if (seq === sequence) loading.value = false;
  }
}
function search() {
  page.value = 1;
  load();
}
watch(category, search);
onMounted(async () => {
  await load();
  try {
    categories.value = await api.request<Category[]>("/categories");
  } catch (e) {
    error.value = (e as Error).message;
  }
});
</script>
<template>
  <section class="hero">
    <div>
      <p class="eyebrow">CURATED FOR EVERYDAY</p>
      <h1>好好生活，<br />从一件好物开始。</h1>
      <p>舒适、实用，也有一点不一样。<br />发现那些值得陪伴你很久的日常。</p>
      <a class="button" href="#collection">探索精选好物 ↗</a>
      <div class="hero-note">
        <span>精选品质</span><span>真实库存</span><span>安心售后</span>
      </div>
    </div>
    <div class="hero-visual">
      <div class="orbit"></div>
      <div class="vase"></div>
      <div class="leaf leaf-one"></div>
      <div class="leaf leaf-two"></div>
      <div class="hero-stamp">LESS, BUT BETTER<br /><b>日常的美好</b></div>
    </div>
  </section>
  <section id="collection">
    <div class="section-title">
      <div>
        <p class="eyebrow">THE COLLECTION</p>
        <h2>发现你的下一件好物</h2>
      </div>
      <span>{{ total }} 件精选商品</span>
    </div>
    <form class="filters" @submit.prevent="search">
      <div class="category-tabs">
        <button
          type="button"
          :class="{ selected: category === 0 }"
          @click="category = 0"
        >
          全部好物</button
        ><button
          v-for="c in categories"
          :key="c.id"
          type="button"
          :class="{ selected: category === c.id }"
          @click="category = c.id"
        >
          {{ c.name }}
        </button>
      </div>
      <div class="search">
        <input
          v-model="query"
          aria-label="搜索商品"
          placeholder="搜索你喜欢的好物"
        /><button aria-label="搜索" type="submit">⌕</button>
      </div>
    </form>
    <StateMessage
      :loading="loading"
      :error="error"
      :empty="!loading && !error && !products.length"
    />
    <div v-if="!loading" class="product-grid">
      <ProductCard v-for="p in products" :key="p.id" :product="p" />
    </div>
    <div v-if="total > 12" class="pagination">
      <button
        :disabled="page === 1"
        @click="
          page--;
          load();
        "
      >
        上一页</button
      ><span>{{ page }} / {{ Math.ceil(total / 12) }}</span
      ><button
        :disabled="page * 12 >= total"
        @click="
          page++;
          load();
        "
      >
        下一页
      </button>
    </div>
  </section>
  <section class="promise-grid">
    <article>
      <span>01</span>
      <h3>好物有标准</h3>
      <p>从材质到细节，认真对待每一次选择。</p>
    </article>
    <article>
      <span>02</span>
      <h3>购物更轻松</h3>
      <p>清晰的规格与库存，简单的购物体验。</p>
    </article>
    <article>
      <span>03</span>
      <h3>售后有回应</h3>
      <p>订单进度随时查看，售后申请在线处理。</p>
    </article>
  </section>
</template>
