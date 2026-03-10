import ApprovalsPageClient from "./_client";

export function generateStaticParams() {
  return [{ id: "__placeholder__" }];
}

export default function ApprovalsPage() {
  return <ApprovalsPageClient />;
}
