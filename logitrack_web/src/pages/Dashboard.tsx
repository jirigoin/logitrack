import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { shipmentApi, type Stats, type Shipment, type ShipmentStatus } from "../api/shipments";
import { fmtDateTime } from "../utils/date";
import { StatusBadge } from "../components/StatusBadge";

const statusConfig: Record<ShipmentStatus, { label: string; color: string; bg: string }> = {
  pending:     { label: "Draft",       color: "#374151", bg: "#f3f4f6" },
  in_progress: { label: "In Progress", color: "#92400e", bg: "#fef3c7" },
  in_transit:  { label: "In Transit", color: "#1e40af", bg: "#dbeafe" },
  at_branch:   { label: "At Branch",  color: "#5b21b6", bg: "#ede9fe" },
  delivering:       { label: "Delivering",     color: "#9a3412", bg: "#ffedd5" },
  delivery_failed:  { label: "Delivery Failed",   color: "#991b1b", bg: "#fee2e2" },
  delivered:        { label: "Delivered",          color: "#065f46", bg: "#d1fae5" },
  ready_for_pickup: { label: "Ready for pickup",      color: "#0e7490", bg: "#cffafe" },
  ready_for_return: { label: "Ready for return",      color: "#5b21b6", bg: "#ede9fe" },
  returned:         { label: "Returned",               color: "#374151", bg: "#f3f4f6" },
  cancelled:        { label: "Cancelled",               color: "#b91c1c", bg: "#fee2e2" },
};

function toDateInput(d: Date): string {
  return d.toISOString().slice(0, 10);
}

function defaultRange(): { from: string; to: string } {
  const to = new Date();
  const from = new Date();
  from.setDate(from.getDate() - 29);
  return { from: toDateInput(from), to: toDateInput(to) };
}

export function Dashboard() {
  const [stats, setStats] = useState<Stats | null>(null);
  const [recent, setRecent] = useState<Shipment[]>([]);
  const range = defaultRange();
  const [dateFrom, setDateFrom] = useState(range.from);
  const [dateTo, setDateTo] = useState(range.to);
  const navigate = useNavigate();

  useEffect(() => {
    shipmentApi.stats({ date_from: dateFrom, date_to: dateTo }).then(setStats);
  }, [dateFrom, dateTo]);

  useEffect(() => {
    shipmentApi.list().then((all) => {
      const sorted = [...all].sort(
        (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
      );
      setRecent(sorted.slice(0, 5));
    });
  }, []);

  return (
    <div style={{ padding: 24 }}>
      <h1 style={{ marginBottom: 24 }}>Dashboard</h1>

      {/* Stats cards */}
      <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(160px, 1fr))", gap: 16, marginBottom: 32 }}>
        <StatCard label="Total Shipments" value={stats?.total ?? 0} color="#1e3a5f" bg="#e0eaff" />
        {(Object.keys(statusConfig) as ShipmentStatus[]).map((s) => (
          <StatCard
            key={s}
            label={statusConfig[s].label}
            value={stats?.by_status?.[s] ?? 0}
            color={statusConfig[s].color}
            bg={statusConfig[s].bg}
          />
        ))}
      </div>

      {/* Shipments per day chart */}
      <div style={{ marginBottom: 32 }}>
        <div style={{ display: "flex", alignItems: "center", gap: 16, marginBottom: 16, flexWrap: "wrap" }}>
          <h2 style={{ margin: 0, fontSize: "1rem" }}>Shipments Created per Day</h2>
          <div style={{ display: "flex", alignItems: "center", gap: 8, fontSize: 13 }}>
            <label htmlFor="date-from" style={{ color: "#6b7280" }}>From</label>
            <input
              id="date-from"
              type="date"
              value={dateFrom}
              max={dateTo}
              onChange={(e) => setDateFrom(e.target.value)}
              style={{ border: "1px solid #d1d5db", borderRadius: 6, padding: "4px 8px", fontSize: 13 }}
            />
            <label htmlFor="date-to" style={{ color: "#6b7280" }}>To</label>
            <input
              id="date-to"
              type="date"
              value={dateTo}
              min={dateFrom}
              onChange={(e) => setDateTo(e.target.value)}
              style={{ border: "1px solid #d1d5db", borderRadius: 6, padding: "4px 8px", fontSize: 13 }}
            />
          </div>
        </div>
        <DayChart byDay={stats?.by_day ?? {}} dateFrom={dateFrom} dateTo={dateTo} />
      </div>

      {/* Recent shipments */}
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 12 }}>
        <h2 style={{ margin: 0, fontSize: "1rem" }}>Recent Shipments</h2>
        <button
          onClick={() => navigate("/")}
          style={{ background: "none", border: "none", color: "#2563eb", cursor: "pointer", fontSize: 14 }}
        >
          View all →
        </button>
      </div>

      {recent.length === 0 ? (
        <p style={{ color: "#6b7280" }}>No shipments yet.</p>
      ) : (
        <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 14 }}>
          <thead>
            <tr style={{ background: "#f9fafb", textAlign: "left" }}>
              <th style={th}>Tracking ID</th>
              <th style={th}>Recipient</th>
              <th style={th}>Destination</th>
              <th style={th}>Status</th>
              <th style={th}>Created At</th>
            </tr>
          </thead>
          <tbody>
            {recent.map((s) => (
              <tr
                key={s.tracking_id}
                onClick={() => navigate(`/shipments/${s.tracking_id}`)}
                style={{ borderBottom: "1px solid #e5e7eb", cursor: "pointer" }}
                onMouseEnter={(e) => (e.currentTarget.style.background = "#f0f9ff")}
                onMouseLeave={(e) => (e.currentTarget.style.background = "")}
              >
                <td style={td}><code>{s.tracking_id}</code></td>
                <td style={td}>{s.recipient.name}</td>
                <td style={td}>{s.recipient.address.city}</td>
                <td style={td}><StatusBadge status={s.status} /></td>
                <td style={td}>{fmtDateTime(s.created_at)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}

// --- DayChart ---

interface DayChartProps {
  byDay: Record<string, number>;
  dateFrom: string;
  dateTo: string;
}

function DayChart({ byDay, dateFrom, dateTo }: DayChartProps) {
  // Build ordered list of days in the range.
  const days: { date: string; count: number }[] = [];
  if (dateFrom && dateTo) {
    const cur = new Date(dateFrom + "T00:00:00");
    const end = new Date(dateTo + "T00:00:00");
    while (cur <= end) {
      const key = cur.toISOString().slice(0, 10);
      days.push({ date: key, count: byDay[key] ?? 0 });
      cur.setDate(cur.getDate() + 1);
    }
  }

  if (days.length === 0) {
    return <p style={{ color: "#6b7280", fontSize: 14 }}>Select a date range to see the chart.</p>;
  }

  const maxCount = Math.max(...days.map((d) => d.count), 1);
  const chartH = 160;
  const barW = Math.max(4, Math.min(32, Math.floor(700 / days.length) - 2));
  const gap = Math.max(1, Math.floor(700 / days.length) - barW);
  const svgW = days.length * (barW + gap) + 40; // 40px left margin for y-axis labels

  // Y-axis ticks (0, max/2, max)
  const yTicks = [0, Math.round(maxCount / 2), maxCount].filter((v, i, a) => a.indexOf(v) === i);

  return (
    <div style={{ overflowX: "auto", background: "#f9fafb", borderRadius: 10, padding: "16px 8px 8px 8px", border: "1px solid #e5e7eb" }}>
      <svg width={svgW} height={chartH + 48} style={{ display: "block" }}>
        {/* Y-axis ticks and gridlines */}
        {yTicks.map((tick) => {
          const y = chartH - Math.round((tick / maxCount) * chartH) + 8;
          return (
            <g key={tick}>
              <line x1={34} x2={svgW} y1={y} y2={y} stroke="#e5e7eb" strokeWidth={1} />
              <text x={30} y={y + 4} textAnchor="end" fontSize={10} fill="#9ca3af">{tick}</text>
            </g>
          );
        })}

        {/* Bars */}
        {days.map((d, i) => {
          const barH = Math.max(1, Math.round((d.count / maxCount) * chartH));
          const x = 40 + i * (barW + gap);
          const y = chartH - barH + 8;
          const showLabel = days.length <= 31 || i % Math.ceil(days.length / 15) === 0;
          const isWeekend = [0, 6].includes(new Date(d.date + "T00:00:00").getDay());
          return (
            <g key={d.date}>
              <rect
                x={x}
                y={y}
                width={barW}
                height={barH}
                rx={2}
                fill={isWeekend ? "#93c5fd" : "#3b82f6"}
                opacity={0.85}
              >
                <title>{d.date}: {d.count} shipment{d.count !== 1 ? "s" : ""}</title>
              </rect>
              {d.count > 0 && barH > 14 && (
                <text x={x + barW / 2} y={y + barH - 4} textAnchor="middle" fontSize={9} fill="white" fontWeight={600}>
                  {d.count}
                </text>
              )}
              {d.count > 0 && barH <= 14 && (
                <text x={x + barW / 2} y={y - 3} textAnchor="middle" fontSize={9} fill="#3b82f6" fontWeight={600}>
                  {d.count}
                </text>
              )}
              {showLabel && (
                <text
                  x={x + barW / 2}
                  y={chartH + 22}
                  textAnchor="middle"
                  fontSize={9}
                  fill="#6b7280"
                  transform={`rotate(-40, ${x + barW / 2}, ${chartH + 22})`}
                >
                  {d.date.slice(5)} {/* MM-DD */}
                </text>
              )}
            </g>
          );
        })}
      </svg>
      <div style={{ fontSize: 11, color: "#9ca3af", marginTop: 4, paddingLeft: 40 }}>
        Total in period: <strong style={{ color: "#374151" }}>{days.reduce((s, d) => s + d.count, 0)}</strong> shipments
      </div>
    </div>
  );
}

function StatCard({ label, value, color, bg }: { label: string; value: number; color: string; bg: string }) {
  return (
    <div style={{ background: bg, borderRadius: 10, padding: "16px 20px" }}>
      <div style={{ fontSize: 28, fontWeight: 700, color }}>{value}</div>
      <div style={{ fontSize: 13, color, opacity: 0.8, marginTop: 4 }}>{label}</div>
    </div>
  );
}

const th: React.CSSProperties = { padding: "10px 14px", fontWeight: 600, color: "#374151" };
const td: React.CSSProperties = { padding: "10px 14px" };
