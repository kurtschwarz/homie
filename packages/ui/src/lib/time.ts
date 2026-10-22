/** Formatting that follows Homie's content rules: lowercase am/pm, no leading zero, en dash ranges, relative days. */

const DAYS = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];
const MONTHS = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];

const pad = (n: number) => (n < 10 ? `0${n}` : `${n}`);

export const ampm = (d: Date) => (d.getHours() < 12 ? "am" : "pm");
export const clockParts = (d: Date) => ({
  hours: String(d.getHours() % 12 || 12),
  minutes: pad(d.getMinutes()),
  ampm: ampm(d),
});
const hm = (d: Date) => `${d.getHours() % 12 || 12}:${pad(d.getMinutes())}`;

/** "5:45 pm" */
export const formatTime = (d: Date) => `${hm(d)} ${ampm(d)}`;

/** "3:30–4:15 pm", or "11:30 am–1:00 pm" across noon */
export const formatRange = (start: Date, end: Date) =>
  ampm(start) === ampm(end)
    ? `${hm(start)}–${hm(end)} ${ampm(end)}`
    : `${formatTime(start)}–${formatTime(end)}`;

/** "1 pm", for hourly forecast labels */
export const formatHour = (d: Date) => `${d.getHours() % 12 || 12} ${ampm(d)}`;

/** "Thu, Oct 15" */
export const formatShortDate = (d: Date) =>
  `${DAYS[d.getDay()]!.slice(0, 3)}, ${MONTHS[d.getMonth()]} ${d.getDate()}`;

/** "Oct 16" */
export const formatMonthDay = (d: Date) => `${MONTHS[d.getMonth()]!} ${d.getDate()}`;

const startOfDay = (d: Date) => new Date(d.getFullYear(), d.getMonth(), d.getDate());

/** Whole days between now and d (0 = today). */
export const dayDiff = (d: Date, now = new Date()) =>
  Math.round((startOfDay(d).getTime() - startOfDay(now).getTime()) / 86_400_000);

/**
 * "Today", "Tomorrow", the weekday within the coming week, then "Wed, Oct 28".
 * A bare weekday stops being an answer once the date is a week out, because
 * "Wednesday" then names two different days and the reader picks the wrong one.
 */
export const dayLabel = (d: Date, now = new Date()) => {
  const k = dayDiff(d, now);
  if (k === 0) return "Today";
  if (k === 1) return "Tomorrow";
  return k < 7 ? DAYS[d.getDay()]! : formatShortDate(d);
};

/** "Today", then "Fri", "Sat" ... for week columns and the daily forecast */
export const shortDayLabel = (d: Date, now = new Date()) =>
  dayDiff(d, now) === 0 ? "Today" : DAYS[d.getDay()]!.slice(0, 3);

/** "in 40 min", "in 2 hr 17 min" */
export const until = (ms: number) => {
  const mins = Math.max(1, Math.round(ms / 60_000));
  if (mins < 60) return `in ${mins} min`;
  const h = Math.floor(mins / 60);
  const m = mins % 60;
  return `in ${h} hr${m ? ` ${m} min` : ""}`;
};

/** "Fri, Aug 14 · 8:12 pm", adding the year only when it isn't this year. */
export const formatTakenAt = (d: Date, now = new Date()) =>
  `${formatShortDate(d)}${d.getFullYear() !== now.getFullYear() ? `, ${d.getFullYear()}` : ""} · ${formatTime(d)}`;
