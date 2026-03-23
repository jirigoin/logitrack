import axios from "axios";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? "http://localhost:8080/api/v1",
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem("token");
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

export interface Branch {
  id: string;
  name: string;
  city: string;
  province: string;
}

export const branchApi = {
  list: () => api.get<Branch[]>("/branches").then((r) => r.data),
};

// branchLabel looks up a branch by city string (used for event locations).
export const branchLabel = (city: string, branches: Branch[]): string => {
  const branch = branches.find((b) => b.city === city);
  return branch ? branch.name : city;
};

// branchLabelById looks up a branch by its ID (used for shipment.current_location).
export const branchLabelById = (id: string, branches: Branch[]): string => {
  const branch = branches.find((b) => b.id === id);
  return branch ? branch.name : id;
};
