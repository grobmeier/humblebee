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
