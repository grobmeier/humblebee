type DashboardSummaryProps = {
  monthWorkTime: string;
  weekWorkTime: string;
};

export function DashboardSummary({ monthWorkTime, weekWorkTime }: DashboardSummaryProps) {
  return (
    <section className="summary-metrics" aria-label="Zeituebersicht">
      <Metric label="Arbeitszeit (Woche)" value={weekWorkTime} />
      <Metric label="Arbeitszeit (Monat)" value={monthWorkTime} />
    </section>
  );
}

type MetricProps = {
  label: string;
  value: string;
};

function Metric({ label, value }: MetricProps) {
  return (
    <div className="summary-metric">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}
