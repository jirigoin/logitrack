import { useState } from "react";
import { vehicleApi, type VehicleStatusResponse, type VehicleStatus, type VehicleType } from "../api/vehicles";
import { useAuth } from "../context/AuthContext";
import { Navigate } from "react-router-dom";

const vehicleTypeLabels: Record<VehicleType, string> = {
  motocicleta: "Motocicleta",
  furgoneta: "Furgoneta",
  camion: "Camión",
  camion_grande: "Camión Grande",
};

const getStatusColor = (status: VehicleStatus): string => {
  switch (status) {
    case "disponible":
      return "#10b981";
    case "mantenimiento":
      return "#f59e0b";
    case "en_transito":
      return "#3b82f6";
    case "inactivo":
      return "#6b7280";
    default:
      return "#9ca3af";
  }
};

const formatDate = (dateString: string): string => {
  const date = new Date(dateString);
  return date.toLocaleString("es-AR", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
};

export function VehicleStatus() {
  const { hasRole } = useAuth();
  const [plate, setPlate] = useState("");
  const [vehicle, setVehicle] = useState<VehicleStatusResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>("");
  const [notFound, setNotFound] = useState(false);

  // Solo supervisor, manager y admin pueden consultar
  if (!hasRole("supervisor") && !hasRole("manager") && !hasRole("admin")) {
    return <Navigate to="/dashboard" replace />;
  }

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!plate.trim()) {
      setError("La patente es obligatoria");
      return;
    }

    setLoading(true);
    setError("");
    setVehicle(null);
    setNotFound(false);

    try {
      const data = await vehicleApi.getByPlate(plate.toUpperCase().trim());
      setVehicle(data);
    } catch (err: any) {
      if (err.response?.status === 404) {
        setNotFound(true);
      } else if (err.response?.status === 400) {
        setError(err.response?.data?.error || "Error en la búsqueda");
      } else {
        setError("Error al consultar el vehículo");
      }
    } finally {
      setLoading(false);
    }
  };

  const handleClear = () => {
    setPlate("");
    setVehicle(null);
    setError("");
    setNotFound(false);
  };

  return (
    <div style={{ padding: 24, maxWidth: 800, margin: "0 auto" }}>
      <h1 style={{ marginBottom: 24, fontSize: 24 }}>Consulta de Estado de Vehículo</h1>

      {/* Formulario de búsqueda */}
      <form onSubmit={handleSearch} style={{ marginBottom: 24 }}>
        <div style={{ display: "flex", gap: 12, alignItems: "flex-end" }}>
          <div style={{ flex: 1 }}>
            <label style={{ display: "block", marginBottom: 6, fontWeight: 500, fontSize: 14 }}>
              Patente del vehículo *
            </label>
            <input
              type="text"
              value={plate}
              onChange={(e) => setPlate(e.target.value.toUpperCase())}
              placeholder="Ej: AB123CD"
              style={{
                width: "100%",
                padding: "10px 14px",
                borderRadius: 6,
                border: "1px solid #d1d5db",
                fontSize: 16,
                textTransform: "uppercase",
                fontWeight: 500,
              }}
            />
          </div>
          <button
            type="submit"
            disabled={loading}
            style={{
              background: "#1e3a5f",
              color: "#fff",
              border: "none",
              borderRadius: 6,
              padding: "10px 20px",
              cursor: loading ? "not-allowed" : "pointer",
              fontWeight: 600,
              fontSize: 14,
              opacity: loading ? 0.7 : 1,
            }}
          >
            {loading ? "Consultando..." : "Consultar"}
          </button>
          <button
            type="button"
            onClick={handleClear}
            style={{
              background: "#e5e7eb",
              color: "#374151",
              border: "none",
              borderRadius: 6,
              padding: "10px 20px",
              cursor: "pointer",
              fontWeight: 500,
              fontSize: 14,
            }}
          >
            Limpiar
          </button>
        </div>
      </form>

      {/* Mensaje de error */}
      {error && (
        <div
          style={{
            background: "#fef2f2",
            border: "1px solid #fecaca",
            color: "#dc2626",
            padding: "12px 16px",
            borderRadius: 6,
            marginBottom: 20,
            fontSize: 14,
          }}
        >
          {error}
        </div>
      )}

      {/* Vehículo no encontrado */}
      {notFound && (
        <div
          style={{
            background: "#fffbeb",
            border: "1px solid #fde68a",
            color: "#92400e",
            padding: "16px 20px",
            borderRadius: 8,
            marginBottom: 20,
            textAlign: "center",
          }}
        >
          <svg
            style={{ width: 48, height: 48, margin: "0 auto 12px", display: "block" }}
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1.5}
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
            />
          </svg>
          <p style={{ fontSize: 16, fontWeight: 600, margin: 0 }}>Vehículo no registrado</p>
          <p style={{ fontSize: 14, margin: "4px 0 0", opacity: 0.8 }}>
            No existe un vehículo con la patente <strong>{plate.toUpperCase()}</strong> en el sistema.
          </p>
        </div>
      )}

      {/* Resultado de la consulta */}
      {vehicle && (
        <div
          style={{
            background: "#fff",
            border: "1px solid #e5e7eb",
            borderRadius: 12,
            overflow: "hidden",
            boxShadow: "0 1px 3px rgba(0,0,0,0.1)",
          }}
        >
          {/* Encabezado con estado */}
          <div
            style={{
              background: `${getStatusColor(vehicle.status)}15`,
              padding: 24,
              borderBottom: `2px solid ${getStatusColor(vehicle.status)}30`,
            }}
          >
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", flexWrap: "wrap", gap: 16 }}>
              <div>
                <p style={{ fontSize: 13, color: "#6b7280", margin: 0, textTransform: "uppercase", letterSpacing: "0.5px" }}>
                  Patente
                </p>
                <h2 style={{ fontSize: 28, fontWeight: 700, margin: "4px 0 0", color: "#111827" }}>
                  {vehicle.license_plate}
                </h2>
              </div>
              <div
                style={{
                  display: "inline-flex",
                  alignItems: "center",
                  gap: 8,
                  padding: "8px 16px",
                  borderRadius: 9999,
                  background: `${getStatusColor(vehicle.status)}20`,
                }}
              >
                <span
                  style={{
                    width: 12,
                    height: 12,
                    borderRadius: "50%",
                    background: getStatusColor(vehicle.status),
                    animation: "pulse 2s infinite",
                  }}
                />
                <span
                  style={{
                    fontSize: 16,
                    fontWeight: 600,
                    color: getStatusColor(vehicle.status),
                  }}
                >
                  {vehicle.status_label}
                </span>
              </div>
            </div>
          </div>

          {/* Información del vehículo */}
          <div style={{ padding: 24 }}>
            <h3 style={{ fontSize: 14, fontWeight: 600, color: "#6b7280", margin: "0 0 16px", textTransform: "uppercase", letterSpacing: "0.5px" }}>
              Información del Vehículo
            </h3>
            <div
              style={{
                display: "grid",
                gridTemplateColumns: "repeat(auto-fit, minmax(180px, 1fr))",
                gap: 16,
              }}
            >
              <InfoCard
                label="Tipo"
                value={vehicleTypeLabels[vehicle.type] || vehicle.type}
                icon={
                  <svg style={{ width: 20, height: 20 }} fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 17a2 2 0 11-4 0 2 2 0 014 0zM19 17a2 2 0 11-4 0 2 2 0 014 0z" />
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M13 16V6a1 1 0 00-1-1H4a1 1 0 00-1 1v10a1 1 0 001 1h1m8-1a1 1 0 01-1 1H9m4-1V8a1 1 0 011-1h2.586a1 1 0 01.707.293l3.414 3.414a1 1 0 01.293.707V16a1 1 0 01-1 1h-1m-6-1a1 1 0 001 1h1M5 17a1 1 0 100-2 1 1 0 000 2z" />
                  </svg>
                }
              />
              <InfoCard
                label="Capacidad"
                value={`${vehicle.capacity_kg} kg`}
                icon={
                  <svg style={{ width: 20, height: 20 }} fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M3 6l3 1m0 0l-3 9a5.002 5.002 0 006.001 0M6 7l3 9M6 7l6-2m6 2l3-1m-3 1l-3 9a5.002 5.002 0 006.001 0M18 7l3 9m-3-9l-6-2m0-2v2m0 16V5m0 16H9m3 0h3" />
                  </svg>
                }
              />
              <InfoCard
                label="Última Actualización"
                value={formatDate(vehicle.updated_at)}
                icon={
                  <svg style={{ width: 20, height: 20 }} fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                }
              />
              <InfoCard
                label="ID Vehículo"
                value={`#${vehicle.id}`}
                icon={
                  <svg style={{ width: 20, height: 20 }} fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" />
                  </svg>
                }
              />
            </div>

            {/* Envío asignado */}
            {vehicle.assigned_shipment && (
              <div
                style={{
                  marginTop: 24,
                  padding: 16,
                  background: "#eff6ff",
                  border: "1px solid #bfdbfe",
                  borderRadius: 8,
                }}
              >
                <h3 style={{ fontSize: 14, fontWeight: 600, color: "#1e40af", margin: "0 0 8px", display: "flex", alignItems: "center", gap: 8 }}>
                  <svg style={{ width: 18, height: 18 }} fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  Envío Asignado
                </h3>
                <p style={{ fontSize: 16, fontWeight: 600, color: "#1e3a5f", margin: 0 }}>
                  Tracking: {vehicle.assigned_shipment}
                </p>
                <p style={{ fontSize: 13, color: "#6b7280", margin: "4px 0 0" }}>
                  Este vehículo está actualmente asignado a un envío en curso.
                </p>
              </div>
            )}

            {/* Sin envío asignado */}
            {!vehicle.assigned_shipment && vehicle.status === "disponible" && (
              <div
                style={{
                  marginTop: 24,
                  padding: 16,
                  background: "#f0fdf4",
                  border: "1px solid #bbf7d0",
                  borderRadius: 8,
                }}
              >
                <p style={{ fontSize: 14, color: "#16a34a", margin: 0, display: "flex", alignItems: "center", gap: 8 }}>
                  <svg style={{ width: 18, height: 18 }} fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  Vehículo disponible para asignación
                </p>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Instrucciones iniciales */}
      {!vehicle && !error && !notFound && (
        <div
          style={{
            textAlign: "center",
            padding: "60px 20px",
            color: "#6b7280",
          }}
        >
          <svg
            style={{ width: 64, height: 64, margin: "0 auto 16px", opacity: 0.5 }}
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1}
              d="M9 17a2 2 0 11-4 0 2 2 0 014 0zM19 17a2 2 0 11-4 0 2 2 0 014 0z"
            />
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1}
              d="M13 16V6a1 1 0 00-1-1H4a1 1 0 00-1 1v10a1 1 0 001 1h1m8-1a1 1 0 01-1 1H9m4-1V8a1 1 0 011-1h2.586a1 1 0 01.707.293l3.414 3.414a1 1 0 01.293.707V16a1 1 0 01-1 1h-1m-6-1a1 1 0 001 1h1M5 17a1 1 0 100-2 1 1 0 000 2z"
            />
          </svg>
          <p style={{ fontSize: 16, margin: 0 }}>
            Ingrese la patente del vehículo para consultar su estado actual
          </p>
        </div>
      )}
    </div>
  );
}

function InfoCard({ label, value, icon }: { label: string; value: string; icon: React.ReactNode }) {
  return (
    <div
      style={{
        padding: 16,
        background: "#f9fafb",
        border: "1px solid #e5e7eb",
        borderRadius: 8,
        display: "flex",
        alignItems: "flex-start",
        gap: 12,
      }}
    >
      <div style={{ color: "#6b7280", flexShrink: 0 }}>{icon}</div>
      <div>
        <p style={{ fontSize: 12, color: "#6b7280", margin: "0 0 4px", textTransform: "uppercase", letterSpacing: "0.5px" }}>
          {label}
        </p>
        <p style={{ fontSize: 16, fontWeight: 600, color: "#111827", margin: 0 }}>{value}</p>
      </div>
    </div>
  );
}