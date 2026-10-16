import { Clock } from "@homie/ui/components/clock";
import { HomeEvents, HomeLayout } from "@homie/ui/components/layout";
import { PhotoPanel, type Photo } from "@homie/ui/components/photo-panel";
import { formatShortDate } from "@homie/ui/lib/time";
import { useNow } from "@homie/ui/hooks/use-now";

// Placeholder photos to preview the slideshow's crossfade and per-shape
// resizing until real ones come from the API. Remove once that lands.
const PLACEHOLDER_PHOTOS: Photo[] = [
  { src: "https://picsum.photos/seed/homie-1/1600/1200", alt: "", place: "Lake Tahoe, CA", takenAt: new Date("2026-08-14T20:12:00") },
  { src: "https://picsum.photos/seed/homie-2/1200/1600", alt: "", place: "Portland, OR" },
  { src: "https://picsum.photos/seed/homie-3/1400/1400", alt: "" },
  { src: "https://picsum.photos/seed/homie-4/1920/1080", alt: "", place: "Banff National Park", takenAt: new Date("2025-07-02T09:30:00") },
];

export function App() {
  const now = useNow();

  return (
    <HomeLayout>
      <PhotoPanel photos={PLACEHOLDER_PHOTOS} intervalMs={4_000} />
      <HomeEvents>
        <div className="flex flex-col gap-hm-8">
          <Clock date={now} />
          <div className="text-hm-lead text-hm-ink-2">{formatShortDate(now)}</div>
        </div>
      </HomeEvents>
    </HomeLayout>
  );
}
