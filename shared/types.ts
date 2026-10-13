export interface Base {
  id: number;
  created_at: string;
  updated_at: string;
}
export interface User extends Base {
  email: string;
  name: string;
  phone: string;
  avatar: string;
  role: "user" | "staff" | "admin";
  enabled: boolean;
}
export interface Category extends Base {
  name: string;
}
export interface Variant extends Base {
  product_id: number;
  sku: string;
  specification: string;
  price: number;
  stock: number;
}
export interface Product extends Base {
  name: string;
  description: string;
  images: string;
  category_id: number;
  category: Category;
  active: boolean;
  variants: Variant[];
}
export interface CartItem extends Base {
  user_id: number;
  variant_id: number;
  variant: Variant;
  quantity: number;
}
export type OrderStatus =
  "pending" | "confirmed" | "shipped" | "completed" | "cancelled" | "refunded";
export interface OrderItem extends Base {
  order_id: number;
  product_id: number;
  variant_id: number;
  name: string;
  specification: string;
  price: number;
  quantity: number;
}
export interface Order extends Base {
  number: string;
  user_id: number;
  status: OrderStatus;
  total: number;
  recipient: string;
  phone: string;
  address: string;
  carrier: string;
  tracking: string;
  items: OrderItem[];
}
export interface Review extends Base {
  user_id: number;
  user: Pick<User, "name" | "avatar">;
  product_id: number;
  order_item_id: number;
  rating: number;
  content: string;
}
export interface AfterSale extends Base {
  order_id: number;
  order: Order;
  reason: string;
  status: "pending" | "approved" | "rejected";
  note: string;
}
export interface Page<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}
export interface LoginResult {
  token: string;
  user: User;
}
export interface Dashboard {
  sales: number;
  orders: number;
  users: number;
  products: number;
  trend: { date: string; sales: number; orders: number }[];
  growth: { date: string; users: number }[];
  ranking: {
    product_id: number;
    name: string;
    quantity: number;
    sales: number;
  }[];
}
