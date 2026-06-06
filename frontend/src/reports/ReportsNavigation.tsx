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

import { SideNavigationList } from "../components/SideNavigationList";
import { reportDefinitions, type ReportSlug, type ReportsPageText } from "./reportTypes";

type ReportsNavigationProps = {
  activeReport: ReportSlug;
  t: ReportsPageText;
};

export function ReportsNavigation({ activeReport, t }: ReportsNavigationProps) {
  return (
    <SideNavigationList
      ariaLabel={t.reportList}
      items={reportDefinitions.map((report) => ({ href: `#reports/${report.slug}`, id: report.slug, label: t.titles[report.slug] }))}
      selectedId={activeReport}
      title={t.reportList}
    />
  );
}
