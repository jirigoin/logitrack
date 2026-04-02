import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { vehicleApi, type Vehicle, type VehicleStatus, type VehicleStatusResponse, type VehicleType } from "../api/vehicles";
import { useAuth } from "../context/AuthContext";

const vehicleTypeLabels: Record<VehicleType, string> = {
  motocicleta: "Motocicleta",
  furgoneta: "Furgoneta",
  camion: "Camión",
  camion_grande: "Camión Grande",
};

const vehicleStatusLabels: Record<VehicleStatus, string> = {
  disponible: "Disponible",
  mantenimiento: "Mantenimiento",
  en_transito: "En Tránsito",
  inactivo: "Inactivo",
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

export function VehicleList() {
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [loading, setLoading] = useState(false);
  const [showForm, setShowForm] = useState(false);
  const [error, setError] = useState<string>("");
  const [success, setSuccess] = useState<string>("");
  const [selectedVehicle, setSelectedVehicle] = useState<VehicleStatusResponse | null>(null);
  const { hasRole } = useAuth();

  const isAdmin = hasRole("admin");

  const [formData, setFormData] = useState({
    license_plate: "",
    type: "furgoneta" as VehicleType,
    capacity_kg: 0,
  });

  const loadVehicles = async () => {
    setLoading(true);
    try {
      const data = await vehicleApi.list();
      setVehicles(data ?? []);
    } catch (err) {
      console.error("Failed to load vehicles:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { loadVehicles(); }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setSuccess("");

    if (!formData.license_plate.trim()) {
      setError("La patente es obligatoria");
      return;
    }
    if (formData.capacity_kg <= 0) {
      setError("La capacidad debe ser mayor a 0");
      return;
    }

    try {
      await vehicleApi.create(formData);
      setSuccess("Vehículo registrado exitosamente");
      setShowForm(false);
      setFormData({ license_plate: "", type: "furgoneta", capacity_kg: 0 });
      loadVehicles();
    } catch (err: any) {
      if (err.response?.status === 409) {
        setError("Ya existe un vehículo con la misma patente");
      } else if (err.response?.data?.error) {
        setError(err.response.data.error);
      } else {
        setError("Error al registrar el vehículo");
      }
    }
  };

  const handleCancel = () => {
    setShowForm(false);
    setFormData({ license_plate: "", type: "furgoneta", capacity_kg: 0 });
    setError("");
    setSuccess("");
  };

  const handleViewVehicle = async (plate: string) => {
    try {
      const data = await vehicleApi.getByPlate(plate);
      setSelectedVehicle(data);
    } catch (err) {
      console.error("Failed to load vehicle details:", err);
    }
  };

  const closeVehicleDetail = () => {
    setSelectedVehicle(null);
  };

  return (
    <div style={{ padding: 24 }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 20 }}>
        <h1 style={{ margin: 0 }}>Gestión de Flota</h1>
        {isAdmin && (
          <button
            onClick={() => setShowForm(!showForm)}
            style={{
              background: "#1e3a5f",
              color: "#fff",
              border: "none",
              borderRadius: 6,
              padding: "8px 16px",
              cursor: "pointer",
              fontWeight: 600,
            }}
          >
            {showForm ? "Cancelar" : "+ Nuevo Vehículo"}
          </button>
        )}
      </div>

      {/* Formulario de alta */}
      {showForm && isAdmin && (
        <div style={{
          background: "#f9fafb",
          border: "1px solid #e5e7eb",
          borderRadius: 8,
          padding: 20,
          marginBottom: 20,
          maxWidth: 500,
        }}>
          <h2 style={{ margin: "0 0 16px 0", fontSize: 18 }}>Registrar Nuevo Vehículo</h2>

          {error && (
            <div style={{
              background: "#fef2f2",
              border: "1px solid #fecaca",
              color: "#dc2626",
              padding: "8px 12px",
              borderRadius: 6,
              marginBottom: 12,
              fontSize: 14,
            }}>
              {error}
            </div>
          )}

          {success && (
            <div style={{
              background: "#f0fdf4",
              border: "1px solid #bbf7d0",
              color: "#16a34a",
              padding: "8px 12px",
              borderRadius: 6,
              marginBottom: 12,
              fontSize: 14,
            }}>
              {success}
            </div>
          )}

          <form onSubmit={handleSubmit}>
            <div style={{ marginBottom: 16 }}>
              <label style={{ display: "block", marginBottom: 6, fontWeight: 500, fontSize: 14 }}>
                Patente *
              </label>
              <input
                type="text"
                value={formData.license_plate}
                onChange={(e) => setFormData({ ...formData, license_plate: e.target.value.toUpperCase() })}
                placeholder="Ej: AB123CD"
                style={{
                  width: "100%",
                  padding: "8px 12px",
                  borderRadius: 6,
                  border: "1px solid #d1d5db",
                  fontSize: 14,
                  textTransform: "uppercase",
                }}
              />
            </div>

            <div style={{ marginBottom: 16 }}>
              <label style={{ display: "block", marginBottom: 6, fontWeight: 500, fontSize: 14 }}>
                Tipo *
              </label>
              <select
                value={formData.type}
                onChange={(e) => setFormData({ ...formData, type: e.target.value as VehicleType })}
                style={{
                  width: "100%",
                  padding: "8px 12px",
                  borderRadius: 6,
                  border: "1px solid #d1d5db",
                  fontSize: 14,
                  background: "#fff",
                }}
              >
                {Object.entries(vehicleTypeLabels).map(([value, label]) => (
                  <option key={value} value={value}>{label}</option>
                ))}
              </select>
            </div>

            <div style={{ marginBottom: 16 }}>
              <label style={{ display: "block", marginBottom: 6, fontWeight: 500, fontSize: 14 }}>
                Capacidad (kg) *
              </label>
              <input
                type="number"
                value={formData.capacity_kg || ""}
                onChange={(e) => setFormData({ ...formData, capacity_kg: parseFloat(e.target.value) || 0 })}
                placeholder="Ej: 500"
                min="1"
                step="0.1"
                style={{
                  width: "100%",
                  padding: "8px 12px",
                  borderRadius: 6,
                  border: "1px solid #d1d5db",
                  fontSize: 14,
                }}
              />
            </div>

            <div style={{ display: "flex", gap: 8, justifyContent: "flex-end" }}>
              <button
                type="button"
                onClick={handleCancel}
                style={{
                  background: "#e5e7eb",
                  color: "#374151",
                  border: "none",
                  borderRadius: 6,
                  padding: "8px 16px",
                  cursor: "pointer",
                  fontWeight: 500,
                }}
              >
                Cancelar
              </button>
              <button
                type="submit"
                style={{
                  background: "#1e3a5f",
                  color: "#fff",
                  border: "none",
                  borderRadius: 6,
                  padding: "8px 16px",
                  cursor: "pointer",
                  fontWeight: 500,
                }}
              >
                Registrar
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Lista de vehículos */}
      {loading ? (
        <p>Cargando...</p>
      ) : vehicles.length === 0 ? (
        <p style={{ color: "#6b7280" }}>No hay vehículos registrados en la flota.</p>
      ) : (
        <>
          <p style={{ fontSize: 13, color: "#6b7280", marginBottom: 8 }}>
            {vehicles.length} vehículo{vehicles.length !== 1 ? "s" : ""} en la flota
          </p>
          <div style={{ overflowX: "auto" }}>
            <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 14, minWidth: 500 }}>
              <thead>
                <tr style={{ background: "#f9fafb", textAlign: "left" }}>
                  <th style={thStyle}>Patente</th>
                  <th style={thStyle}>Tipo</th>
                  <th style={thStyle}>Capacidad (kg)</th>
                  <th style={thStyle}>Estado</th>
                </tr>
              </thead>
              <tbody>
                {vehicles.map((v) => (
                  <tr
                    key={v.id}
                    style={{ borderBottom: "1px solid #e5e7eb" }}
                    onMouseEnter={(e) => (e.currentTarget.style.background = "#f0f9ff")}
                    onMouseLeave={(e) => (e.currentTarget.style.background = "")}
                  >
                    <td style={{ ...tdStyle, cursor: "pointer" }} onClick={() => handleViewVehicle(v.license_plate)}>
                      <code style={{ fontWeight: 600, fontSize: 15 }}>{v.license_plate}</code>
                    </td>
                    <td style={{ ...tdStyle, cursor: "pointer" }} onClick={() => handleViewVehicle(v.license_plate)}>{vehicleTypeLabels[v.type]}</td>
                    <td style={{ ...tdStyle, cursor: "pointer" }} onClick={() => handleViewVehicle(v.license_plate)}>{v.capacity_kg} kg</td>
                    <td style={{ ...tdStyle, cursor: "pointer" }} onClick={() => handleViewVehicle(v.license_plate)}>
                      <span
                        style={{
                          display: "inline-flex",
                          alignItems: "center",
                          gap: 6,
                          padding: "4px 10px",
                          borderRadius: 9999,
                          fontSize: 13,
                          fontWeight: 500,
                          background: `${getStatusColor(v.status)}20`,
                          color: getStatusColor(v.status),
                        }}
                      >
                        <span
                          style={{
                            width: 8,
                            height: 8,
                            borderRadius: "50%",
                            background: getStatusColor(v.status),
                          }}
                        />
                        {vehicleStatusLabels[v.status]}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      )}

      {/* Modal de detalle de vehículo */}
      {selectedVehicle && (
        <VehicleDetailModal
          vehicle={selectedVehicle}
          onClose={closeVehicleDetail}
        />
      )}
    </div>
  );
}

// Vehicle Detail Modal Component
function VehicleDetailModal({ vehicle, onClose }: { vehicle: VehicleStatusResponse; onClose: () => void }) {
  return (
    <div
      style={{
        position: "fixed",
        inset: 0,
        background: "rgba(0,0,0,0.45)",
        zIndex: 1000,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        padding: 16,
      }}
      onClick={onClose}
    >
      <div
        style={{
          background: "#fff",
          borderRadius: 12,
          padding: 24,
          maxWidth: 560,
          width: "100%",
          maxHeight: "90vh",
          overflowY: "auto",
          boxShadow: "0 20px 60px rgba(0,0,0,0.3)",
        }}
        onClick={(e) => e.stopPropagation()}
      >
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 20 }}>
          <div>
            <p style={{ fontSize: 13, color: "#6b7280", margin: 0, textTransform: "uppercase" }}>Detalle del Vehículo</p>
            <h2 style={{ fontSize: 24, fontWeight: 700, margin: "4px 0 0", color: "#111827" }}>
              {vehicle.license_plate}
            </h2>
          </div>
          <button
            onClick={onClose}
            style={{
              background: "none",
              border: "none",
              fontSize: 24,
              cursor: "pointer",
              color: "#6b7280",
              padding: "4px 8px",
            }}
          >
            ✕
          </button>
        </div>

        <div style={{ display: "flex", alignItems: "center", gap: 12, marginBottom: 24 }}>
          <div
            style={{
              width: 56,
              height: 56,
              borderRadius: 12,
              background: getStatusColor(vehicle.status) + "20",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              flexShrink: 0,
            }}
          >
            <svg style={{ width: 28, height: 28, color: getStatusColor(vehicle.status) }} fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 17a2 2 0 11-4 0 2 2 0 014 0zM19 17a2 2 0 11-4 0 2 2 0 014 0z" />
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M13 16V6a1 1 0 00-1-1H4a1 1 0 00-1 1v10a1 1 0 001 1h1m8-1a1 1 0 01-1 1H9m4-1V8a1 1 0 011-1h2.586a1 1 0 01.707.293l3.414 3.414a1 1 0 01.293.707V16a1 1 0 01-1 1h-1m-6-1a1 1 0 001 1h1M5 17a1 1 0 100-2 1 1 0 000 2z" />
            </svg>
          </div>
          <div style={{ flex: 1 }}>
            <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
              <span
                style={{
                  padding: "4px 12px",
                  borderRadius: 9999,
                  background: getStatusColor(vehicle.status) + "20",
                  fontSize: 13,
                  fontWeight: 600,
                  color: getStatusColor(vehicle.status),
                }}
              >
                {vehicleStatusLabels[vehicle.status]}
              </span>
            </div>
            <p style={{ fontSize: 13, color: "#6b7280", margin: "4px 0 0" }}>
              ID: #{vehicle.id}
            </p>
          </div>
        </div>

        <div
          style={{
            background: "#f9fafb",
            border: "1px solid #e5e7eb",
            borderRadius: 8,
            padding: 16,
            marginBottom: 20,
          }}
        >
          <h3 style={{ fontSize: 14, fontWeight: 600, color: "#374151", margin: "0 0 12px" }}>Información del Vehículo</h3>
          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 12 }}>
            <div>
              <p style={{ fontSize: 12, color: "#6b7280", margin: "0 0 2px" }}>Tipo</p>
              <p style={{ fontSize: 15, fontWeight: 600, color: "#111827", margin: 0 }}>{vehicleTypeLabels[vehicle.type]}</p>
            </div>
            <div>
              <p style={{ fontSize: 12, color: "#6b7280", margin: "0 0 2px" }}>Capacidad</p>
              <p style={{ fontSize: 15, fontWeight: 600, color: "#111827", margin: 0 }}>{vehicle.capacity_kg} kg</p>
            </div>
            {vehicle.updated_at && (
              <div>
                <p style={{ fontSize: 12, color: "#6b7280", margin: "0 0 2px" }}>Última actualización</p>
                <p style={{ fontSize: 14, fontWeight: 500, color: "#374151", margin: 0 }}>
                  {new Date(vehicle.updated_at).toLocaleString()}
                </p>
              </div>
            )}
            {vehicle.updated_by && (
              <div>
                <p style={{ fontSize: 12, color: "#6b7280", margin: "0 0 2px" }}>Actualizado por</p>
                <p style={{ fontSize: 14, fontWeight: 500, color: "#374151", margin: 0 }}>{vehicle.updated_by}</p>
              </div>
            )}
          </div>
        </div>

        {/* Envíos asignados */}
        <div
          style={{
            background: "#f9fafb",
            border: "1px solid #e5e7eb",
            borderRadius: 8,
            padding: 16,
          }}
        >
          <h3 style={{ fontSize: 14, fontWeight: 600, color: "#374151", margin: "0 0 12px" }}>Envíos Asignados</h3>
          {vehicle.assigned_shipment ? (
            <div>
              <div
                style={{
                  background: "#fff",
                  border: "1px solid #e5e7eb",
                  borderRadius: 8,
                  padding: 12,
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "space-between",
                }}
              >
                <div>
                  <p style={{ fontSize: 16, fontWeight: 700, color: "#1e3a5f", margin: 0 }}>
                    {vehicle.assigned_shipment}
                  </p>
                  <p style={{ fontSize: 12, color: "#6b7280", margin: "2px 0 0" }}>
                    Tracking ID del envío asignado
                  </p>
                </div>
                <Link
                  to={`/shipments/${vehicle.assigned_shipment}`}
                  style={{
                    background: "#1e3a5f",
                    color: "#fff",
                    textDecoration: "none",
                    borderRadius: 6,
                    padding: "6px 12px",
                    fontSize: 13,
                    fontWeight: 500,
                  }}
                  onClick={onClose}
                >
                  Ver envío
                </Link>
              </div>
            </div>
          ) : (
            <div
              style={{
                textAlign: "center",
                padding: "20px 0",
                color: "#6b7280",
              }}
            >
              <svg style={{ width: 32, height: 32, margin: "0 auto 8px", opacity: 0.5 }} fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
              </svg>
              <p style={{ fontSize: 14, margin: 0 }}>Sin envíos asignados</p>
              <p style={{ fontSize: 12, margin: "4px 0 0" }}>Este vehículo no tiene envíos asignados actualmente</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

const thStyle: React.CSSProperties = { padding: "10px 14px", fontWeight: 600, color: "#374151" };
const tdStyle: React.CSSProperties = { padding: "10px 14px" };