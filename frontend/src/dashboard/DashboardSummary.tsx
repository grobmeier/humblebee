/*
 * Copyright 2026 Grobmeier Solutions GmbH. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
