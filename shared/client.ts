import { ref } from "vue";
import type { User, LoginResult } from "./types";
export function createClient(storageKey: string) {
  const token = ref(sessionStorage.getItem(storageKey) ?? "");
  const user = ref<User | null>(null);
  async function request<T>(
    path: string,
    method = "GET",
    body?: unknown,
    key?: string,
  ): Promise<T> {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };
    if (token.value) headers.Authorization = `Bearer ${token.value}`;
    if (key) headers["Idempotency-Key"] = key;
    const response = await fetch(`/api/v1${path}`, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
      signal: AbortSignal.timeout(15000),
    });
    const result = (await response.json()) as { data: T; error?: string };
    if (!response.ok) {
      if (response.status === 401) {
        clear();
      }
      throw new Error(result.error ?? "请求失败");
    }
    return result.data;
  }
  function clear() {
    token.value = "";
    user.value = null;
    sessionStorage.removeItem(storageKey);
  }
  async function login(email: string, password: string) {
    const r = await request<LoginResult>("/auth/login", "POST", {
      email,
      password,
    });
    token.value = r.token;
    user.value = r.user;
    sessionStorage.setItem(storageKey, r.token);
    return r.user;
  }
  async function refresh() {
    if (token.value) user.value = await request<User>("/me");
  }
  async function logout() {
    if (token.value) await request("/auth/logout", "POST");
    clear();
  }
  return { request, login, refresh, logout, clear, token, user };
}
