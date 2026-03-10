import StoryDetailPageClient from "./_client";

export function generateStaticParams() {
  return [{ id: "__placeholder__", storyId: "__placeholder__" }];
}

export default function StoryDetailPage() {
  return <StoryDetailPageClient />;
}
