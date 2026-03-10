import EpicOverviewPageClient from "./_client";

// generateStaticParams is required for dynamic routes with output: 'export'.
// We return a placeholder since epics are loaded dynamically at runtime.
export function generateStaticParams() {
  return [{ id: "__placeholder__" }];
}

export default function EpicOverviewPage() {
  return <EpicOverviewPageClient />;
}
