import axios from "axios";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? "http://localhost:8080/api/v1",
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem("token");
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

export interface MLFactors {
  shipment_type: number;
  distance_km: number;
  restrictions: number;
  time_window: number;
  volume_score: number;
  route_saturation: number;
}

export interface MLConfig {
  id: number;
  factors: MLFactors;
  alta_threshold: number;
  media_threshold: number;
  is_active: boolean;
  created_by: string;
  created_at: string;
  notes: string;
}

export interface RegenerateRequest {
  factors: MLFactors;
  alta_threshold: number;
  media_threshold: number;
  notes: string;
}

export interface RegenerateResponse {
  config: MLConfig;
  recalculated_count: number;
}

export interface ActivateResponse {
  config: MLConfig;
  recalculated_count: number;
}

export const mlConfigApi = {
  getActive: () => api.get<MLConfig>("/ml/config").then((r) => r.data),
  getHistory: () => api.get<MLConfig[]>("/ml/config/history").then((r) => r.data),
  regenerate: (data: RegenerateRequest) =>
    api.post<RegenerateResponse>("/ml/config/regenerate", data).then((r) => r.data),
  activate: (id: number) =>
    api.post<ActivateResponse>(`/ml/config/${id}/activate`).then((r) => r.data),
};
