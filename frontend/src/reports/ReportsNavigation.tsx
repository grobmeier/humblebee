import { SideNavigationList } from "../components/SideNavigationList";
import { reportDefinitions, type ReportSlug } from "./reportTypes";

export function ReportsNavigation({ activeReport }: { activeReport: ReportSlug }) {
  return (
    <SideNavigationList
      ariaLabel="Reports"
      items={reportDefinitions.map((report) => ({ href: `#reports/${report.slug}`, id: report.slug, label: report.title }))}
      selectedId={activeReport}
      title="Reports"
    />
  );
}
