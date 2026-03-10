import ReviewPageClient from "./_client";

export function generateStaticParams() {
  return [{ id: "__placeholder__", approvalId: "__placeholder__" }];
}

export default function ReviewPage() {
  return <ReviewPageClient />;
}
