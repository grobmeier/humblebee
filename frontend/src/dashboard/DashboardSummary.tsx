type DashboardSummaryProps = {
  projectTime: string;
  workTime: string;
};

export function DashboardSummary({ projectTime, workTime }: DashboardSummaryProps) {
  return (
    <section className="summary-metrics" aria-label="Zeituebersicht">
      <Metric label="Projektzeit" value={projectTime} weekLabel="Woche" weekValue={projectTime} />
      <Metric label="Abwesenheiten" value="00:00" weekLabel="Woche" weekValue="00:00" />
      <Metric label="Arbeitszeit" value={workTime} weekLabel="Woche" weekValue={workTime} />
      <Metric label="Ist" value="-40:00" valueClassName="negative" weekLabel="Soll" weekValue="40:00" />
      <Metric className="pause-metric" label="Pause" value="00:00" />
    </section>
  );
}

type MetricProps = {
  className?: string;
  label: string;
  value: string;
  valueClassName?: string;
  weekLabel?: string;
  weekValue?: string;
};

function Metric({ className = "", label, value, valueClassName, weekLabel, weekValue }: MetricProps) {
  return (
    <div className={`summary-metric ${className}`}>
      <span>{label}</span>
      <strong className={valueClassName}>{value}</strong>
      {weekLabel ? <small>{weekLabel}</small> : null}
      {weekValue ? <em>{weekValue}</em> : null}
    </div>
  );
}
