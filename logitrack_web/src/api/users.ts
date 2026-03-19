import axios from "axios";
import type { User } from "./auth";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? "http://localhost:8080/api/v1",
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem("token");
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

export const usersApi = {
  listDrivers: () =>
    api.get<{ drivers: User[] }>("/users/drivers").then((r) => r.data.drivers ?? []),
};
